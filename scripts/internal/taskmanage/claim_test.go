package taskmanage

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func seedTask(t *testing.T, tasks, status, file string) {
	t.Helper()
	dir := filepath.Join(tasks, status)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, file), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestClaimTakesHighestPriorityTodoFirst(t *testing.T) {
	tasks := newTasksDir(t)
	seedTask(t, tasks, "todo", "low_zeta.md")
	seedTask(t, tasks, "todo", "high_beta.md")
	seedTask(t, tasks, "todo", "critical_gamma.md")

	got, ok, err := ClaimTask(tasks)
	if err != nil || !ok {
		t.Fatalf("ClaimTask = (%q, %v, %v), want a claim", got, ok, err)
	}
	if got != "critical_gamma.md" {
		t.Errorf("claimed %q, want the critical task", got)
	}
	if _, err := os.Stat(filepath.Join(tasks, "doing", got)); err != nil {
		t.Errorf("claimed task is not in doing: %v", err)
	}
}

func TestClaimPromotesFromBacklogWhenTodoIsEmpty(t *testing.T) {
	tasks := newTasksDir(t)
	seedTask(t, tasks, "backlog", "medium_idea.md")

	got, ok, err := ClaimTask(tasks)
	if err != nil || !ok {
		t.Fatalf("ClaimTask = (%q, %v, %v), want a claim from backlog", got, ok, err)
	}
	if got != "medium_idea.md" {
		t.Errorf("claimed %q, want medium_idea.md", got)
	}
}

func TestClaimReportsNothingLeft(t *testing.T) {
	tasks := newTasksDir(t)
	if _, ok, err := ClaimTask(tasks); err != nil || ok {
		t.Errorf("empty spec: ok=%v err=%v, want nothing and no error", ok, err)
	}
}

// TestClaimIsExclusiveAcrossWorkers races several claimers for the same tasks:
// every task must be claimed exactly once.
func TestClaimIsExclusiveAcrossWorkers(t *testing.T) {
	tasks := newTasksDir(t)
	const n = 12
	for i := 0; i < n; i++ {
		seedTask(t, tasks, "todo", "high_task"+string(rune('a'+i))+".md")
	}

	var mu sync.Mutex
	claimed := map[string]int{}
	var wg sync.WaitGroup
	for w := 0; w < 6; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				file, ok, err := ClaimTask(tasks)
				if err != nil {
					t.Errorf("ClaimTask: %v", err)
					return
				}
				if !ok {
					return
				}
				mu.Lock()
				claimed[file]++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if len(claimed) != n {
		t.Errorf("claimed %d distinct tasks, want %d", len(claimed), n)
	}
	for file, count := range claimed {
		if count != 1 {
			t.Errorf("%s claimed %d times, want once", file, count)
		}
	}
}

func TestProjectRootOfTasksDir(t *testing.T) {
	root := filepath.Join("/repo")
	tasks := TasksDir(root, "20261015-143022-login")
	if got := ProjectRootOf(tasks); got != root {
		t.Errorf("ProjectRootOf = %q, want %q", got, root)
	}
	if got := SpecDirOf(tasks); got != filepath.Join(root, ".plaesy", "specs", "20261015-143022-login") {
		t.Errorf("SpecDirOf = %q", got)
	}
}

// A fresh lock belongs to a live worker, so the task is not claimed.
func TestFreshClaimLockBlocksTheTask(t *testing.T) {
	tasks := newTasksDir(t)
	seedTask(t, tasks, "todo", "high_held.md")
	if err := os.MkdirAll(filepath.Join(tasks, ".claims", "high_held.md.lock"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got, ok, err := ClaimTask(tasks); err != nil || ok {
		t.Errorf("claimed %q (ok=%v err=%v) while a live lock was held", got, ok, err)
	}
}

// A lock older than claimStale was left by a dead worker and is taken over.
func TestStaleClaimLockIsTakenOver(t *testing.T) {
	tasks := newTasksDir(t)
	seedTask(t, tasks, "todo", "high_orphan.md")
	lock := filepath.Join(tasks, ".claims", "high_orphan.md.lock")
	if err := os.MkdirAll(lock, 0o755); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-2 * claimStale)
	if err := os.Chtimes(lock, old, old); err != nil {
		t.Fatal(err)
	}
	got, ok, err := ClaimTask(tasks)
	if err != nil || !ok || got != "high_orphan.md" {
		t.Errorf("ClaimTask = (%q, %v, %v), want the orphaned task claimed", got, ok, err)
	}
	if _, err := os.Stat(lock); !os.IsNotExist(err) {
		t.Errorf("lock was not released after the claim: %v", err)
	}
}

// A spec with no backlog folder and an empty todo has nothing to claim, and
// that is not an error.
func TestClaimWithoutBacklogFolderReportsNothing(t *testing.T) {
	tasks := filepath.Join(t.TempDir(), "tasks")
	if err := os.MkdirAll(filepath.Join(tasks, "todo"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got, ok, err := ClaimTask(tasks); err != nil || ok {
		t.Errorf("ClaimTask = (%q, %v, %v), want nothing and no error", got, ok, err)
	}
}
