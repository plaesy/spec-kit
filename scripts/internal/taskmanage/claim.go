package taskmanage

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

// ClaimTask takes one task from a spec for a worker: it moves the highest
// priority todo task to doing and returns its file name. When todo is empty it
// first promotes the next backlog task. found is false when nothing is left.
//
// Exclusivity comes from a lock directory per task, created with os.Mkdir,
// which fails for everyone but one caller. A rename alone is not enough: on
// Windows, concurrent renames of one file can all report success. The lock is
// held only while the task moves from todo to doing, so several `tasks run`
// processes can share a spec without taking the same task twice.
func ClaimTask(tasksDir string) (file string, found bool, err error) {
	for attempt := 0; attempt < 2; attempt++ {
		names, _, err := ListTasks(tasksDir, "todo")
		if err != nil {
			return "", false, err
		}
		sort.SliceStable(names, func(i, j int) bool {
			return priorityRank(names[i]) < priorityRank(names[j])
		})
		for _, n := range names {
			file := n + ".md"
			if err := claimFile(tasksDir, file); err != nil {
				if errors.Is(err, errClaimed) || os.IsNotExist(err) {
					continue // another worker holds it or already moved it
				}
				return "", false, err
			}
			return file, true, nil
		}
		// Nothing in todo was claimable: promote the next backlog task, then
		// look again. A spec without a backlog folder simply has nothing left.
		if _, statErr := os.Stat(filepath.Join(tasksDir, "backlog")); os.IsNotExist(statErr) {
			return "", false, nil
		}
		_, ok, err := PopNextTask(tasksDir)
		if err != nil {
			return "", false, err
		}
		if !ok {
			return "", false, nil
		}
	}
	return "", false, nil
}

// claimFile takes the lock for one task, then moves it from todo to doing
// while holding the lock. The lock is released on every path.
func claimFile(tasksDir, file string) error {
	release, err := acquireClaim(tasksDir, file)
	if err != nil {
		return err
	}
	defer release()

	doing := filepath.Join(tasksDir, "doing")
	if err := os.MkdirAll(doing, 0o755); err != nil {
		return err
	}
	return os.Rename(filepath.Join(tasksDir, "todo", file), filepath.Join(doing, file))
}

// acquireClaim creates the lock directory for a task. It returns errClaimed
// when a live lock exists, and takes over a lock older than claimStale.
func acquireClaim(tasksDir, file string) (release func(), err error) {
	locks := filepath.Join(tasksDir, ".claims")
	if err := os.MkdirAll(locks, 0o755); err != nil {
		return nil, err
	}
	lock := filepath.Join(locks, file+".lock")
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

// priorityRank orders task file names by their priority prefix. Lower ranks
// are claimed first. A name without a known prefix ranks last. Within a rank,
// the order is the one ListTasks already gave (alphabetical).
func priorityRank(file string) int {
	switch {
	case strings.HasPrefix(file, "critical_"):
		return 0
	case strings.HasPrefix(file, "high_"):
		return 1
	case strings.HasPrefix(file, "medium_"):
		return 2
	case strings.HasPrefix(file, "low_"):
		return 3
	}
	return 4
}

// SpecDirOf returns the spec folder that owns a tasks directory:
// <spec>/tasks -> <spec>.
func SpecDirOf(tasksDir string) string {
	return filepath.Dir(tasksDir)
}

// ProjectRootOf returns the project root that owns a tasks directory:
// <root>/.plaesy/specs/<spec>/tasks -> <root>.
func ProjectRootOf(tasksDir string) string {
	return filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(tasksDir))))
}
