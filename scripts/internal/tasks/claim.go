package tasks

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// claimStale is how long a claim lock may exist before it is treated as left
// behind by a worker that died. A claim holds the lock only for one rename, so
// anything older than this is not a live claim.
const claimStale = 30 * time.Minute

// errClaimed means another worker holds the claim on a task.
var errClaimed = errors.New("task is claimed by another worker")

// ClaimTask takes one task of spec (any spec when empty) for a worker: it moves the highest
// priority todo task to doing and returns its folder name. When todo is empty it
// first promotes the next backlog task. found is false when nothing is left.
//
// Exclusivity comes from a lock directory per task, created with os.Mkdir,
// which fails for everyone but one caller. A rename alone is not enough: on
// Windows, concurrent renames of one folder can all report success. The lock is
// held only while the task moves from todo to doing, so several `tasks run`
// processes can share a spec without taking the same task twice.
func ClaimTask(tasksDir, spec string) (task string, found bool, err error) {
	for attempt := 0; attempt < 2; attempt++ {
		names, _, err := ListTasks(tasksDir, "todo")
		if err != nil {
			return "", false, err
		}
		names = forSpec(tasksDir, "todo", names, spec)
		sort.SliceStable(names, func(i, j int) bool {
			return priorityRank(names[i]) < priorityRank(names[j])
		})
		for _, name := range names {
			if err := claimTask(tasksDir, name); err != nil {
				if errors.Is(err, errClaimed) || os.IsNotExist(err) {
					continue // another worker holds it or already moved it
				}
				return "", false, err
			}
			return name, true, nil
		}
		// Nothing in todo was claimable: promote the next backlog task, then
		// look again. A spec without a backlog folder simply has nothing left.
		if _, statErr := os.Stat(filepath.Join(tasksDir, "backlog")); os.IsNotExist(statErr) {
			return "", false, nil
		}
		_, ok, err := PopNextTask(tasksDir, spec)
		if err != nil {
			return "", false, err
		}
		if !ok {
			return "", false, nil
		}
	}
	return "", false, nil
}

// claimTask takes the lock for one task, then moves it from todo to doing while
// holding the lock. The lock is released on every path.
func claimTask(tasksDir, name string) error {
	release, err := acquireClaim(tasksDir, name)
	if err != nil {
		return err
	}
	defer release()
	return MoveTask(tasksDir, name, "todo", "doing")
}

// acquireClaim creates the lock directory for a task. It returns errClaimed
// when a live lock exists, and takes over a lock older than claimStale.
func acquireClaim(tasksDir, name string) (release func(), err error) {
	locks := filepath.Join(tasksDir, ".claims")
	if err := os.MkdirAll(locks, 0o755); err != nil {
		return nil, err
	}
	lock := filepath.Join(locks, name+".lock")
	for attempt := 0; attempt < 2; attempt++ {
		err := os.Mkdir(lock, 0o755)
		if err == nil {
			return func() { _ = os.Remove(lock) }, nil
		}
		if !os.IsExist(err) {
			return nil, err
		}
		info, statErr := os.Stat(lock)
		if statErr != nil {
			if os.IsNotExist(statErr) {
				continue // released between our Mkdir and Stat: try again
			}
			return nil, statErr
		}
		if time.Since(info.ModTime()) < claimStale {
			return nil, errClaimed
		}
		_ = os.Remove(lock) // left behind by a dead worker
	}
	return nil, errClaimed
}

// priorityRank orders task names by their priority prefix. Lower ranks are
// claimed first. A name without a known prefix ranks last.
func priorityRank(name string) int {
	for rank, p := range priorities {
		if strings.HasPrefix(name, p+"_") {
			return rank
		}
	}
	return len(priorities)
}

// SpecDirOf returns the spec folder that owns a tasks folder: <spec>/tasks -> <spec>.
func SpecDirOf(tasksDir string) string {
	return filepath.Dir(tasksDir)
}

// ProjectRootOf returns the project root that owns a tasks folder:
// <root>/.plaesy/specs/<spec>/tasks -> <root>.
func ProjectRootOf(tasksDir string) string {
	return filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(tasksDir))))
}
