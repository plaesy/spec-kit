package tasks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClaimTakesHighestPriorityTodoFirst(t *testing.T) {
	tasksDir := newSpecTasks(t)
	for _, p := range []string{"low", "high"} {
		name, _ := AddTask(tasksDir, "main", p, "t", "T")
		if err := MoveTask(tasksDir, name, "backlog", "todo"); err != nil {
			t.Fatal(err)
		}
	}
	name, found, err := ClaimTask(tasksDir, "")
	if err != nil || !found || name != "high_t" {
		t.Fatalf("ClaimTask = %q, %v, %v; want high_t", name, found, err)
	}
	if s, _, _ := FindTask(tasksDir, "high_t"); s != "doing" {
		t.Errorf("claimed task must be in doing, is in %q", s)
	}
}

func TestClaimPromotesFromBacklogWhenTodoIsEmpty(t *testing.T) {
	tasksDir := newSpecTasks(t)
	AddTask(tasksDir, "main", "medium", "b", "B")
	name, found, err := ClaimTask(tasksDir, "")
	if err != nil || !found || name != "medium_b" {
		t.Fatalf("ClaimTask = %q, %v, %v; want medium_b", name, found, err)
	}
}

func TestClaimReportsNothingLeft(t *testing.T) {
	tasksDir := newSpecTasks(t)
	if _, found, err := ClaimTask(tasksDir, ""); err != nil || found {
		t.Errorf("an empty spec has nothing to claim: found=%v err=%v", found, err)
	}
}

func TestFreshClaimLockBlocksTheTask(t *testing.T) {
	tasksDir := newSpecTasks(t)
	name, _ := AddTask(tasksDir, "main", "high", "a", "A")
	if err := MoveTask(tasksDir, name, "backlog", "todo"); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tasksDir, ".claims", name+".lock"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, found, _ := ClaimTask(tasksDir, ""); found {
		t.Error("a live lock must keep another worker from taking the task")
	}
}
