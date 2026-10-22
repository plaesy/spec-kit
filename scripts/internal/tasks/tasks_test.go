package tasks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newSpecTasks returns the tasks folder of a fresh spec with its status folders.
func newSpecTasks(t *testing.T) string {
	t.Helper()
	tasksDir := filepath.Join(t.TempDir(), "spec", "tasks")
	if err := CreateStatusDirs(tasksDir); err != nil {
		t.Fatal(err)
	}
	return tasksDir
}

func TestAddTaskCreatesFolderWithTaskAndContract(t *testing.T) {
	tasksDir := newSpecTasks(t)
	name, err := AddTask(tasksDir, "main", "high", "login-form", "Build the login form")
	if err != nil {
		t.Fatal(err)
	}
	if name != "high_login-form" {
		t.Fatalf("name = %q, want high_login-form", name)
	}
	dir := filepath.Join(tasksDir, "backlog", name)
	task, err := os.ReadFile(filepath.Join(dir, TaskFileName))
	if err != nil {
		t.Fatalf("spec.md missing: %v", err)
	}
	for _, section := range []string{"title: Build the login form", "status: backlog", "## Goal", "## KPI", "## Verification"} {
		if !strings.Contains(string(task), section) {
			t.Errorf("spec.md is missing %q", section)
		}
	}
	contract, err := os.ReadFile(filepath.Join(dir, ContractFileName))
	if err != nil {
		t.Fatalf("contracts.json missing: %v", err)
	}
	if !strings.Contains(string(contract), `"checklist": []`) {
		t.Errorf("contracts.json must start with an empty checklist: %s", contract)
	}
}

func TestAddTaskRejectsDuplicatesAndBadInput(t *testing.T) {
	tasksDir := newSpecTasks(t)
	if _, err := AddTask(tasksDir, "main", "high", "login-form", "A"); err != nil {
		t.Fatal(err)
	}
	if _, err := AddTask(tasksDir, "main", "high", "login-form", "B"); err == nil {
		t.Error("a second task with the same name must be refused")
	}
	if _, err := AddTask(tasksDir, "main", "urgent", "x", "C"); err == nil {
		t.Error("an unknown priority must be refused")
	}
	if _, err := AddTask(tasksDir, "main", "backlog", "x", "C"); err == nil {
		t.Error("backlog is a status, not a priority")
	}
	if _, err := AddTask(tasksDir, "main", "low", "a/b", "C"); err == nil {
		t.Error("a slug with a slash must be refused")
	}
}

func TestGetNextTaskPicksHighestPriority(t *testing.T) {
	tasksDir := newSpecTasks(t)
	for _, p := range []string{"low", "critical", "high"} {
		if _, err := AddTask(tasksDir, "main", p, "t", "T"); err != nil {
			t.Fatal(err)
		}
	}
	name, found, err := GetNextTask(tasksDir, "")
	if err != nil || !found {
		t.Fatalf("GetNextTask: %v, found %v", err, found)
	}
	if name != "critical_t" {
		t.Errorf("next = %q, want critical_t", name)
	}
}

func TestMoveTaskMovesWholeFolder(t *testing.T) {
	tasksDir := newSpecTasks(t)
	name, _ := AddTask(tasksDir, "main", "high", "a", "A")
	if err := MoveTask(tasksDir, name, "backlog", "todo"); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{TaskFileName, ContractFileName} {
		if _, err := os.Stat(filepath.Join(tasksDir, "todo", name, f)); err != nil {
			t.Errorf("%s did not move with the task: %v", f, err)
		}
	}
	if _, err := os.Stat(filepath.Join(tasksDir, "backlog", name)); !os.IsNotExist(err) {
		t.Error("the task must be gone from backlog")
	}
	if err := MoveTask(tasksDir, name, "backlog", "todo"); err == nil {
		t.Error("moving a task that is no longer in backlog must fail")
	}
}

func TestTaskLifecycle(t *testing.T) {
	tasksDir := newSpecTasks(t)
	name, _ := AddTask(tasksDir, "main", "high", "a", "A")
	if _, ok, err := PopNextTask(tasksDir, ""); err != nil || !ok {
		t.Fatalf("PopNextTask: %v, %v", ok, err)
	}
	if err := TaskStart(tasksDir, name); err != nil {
		t.Fatal(err)
	}
	if err := TaskBlock(tasksDir, name, "waiting for the API key"); err != nil {
		t.Fatal(err)
	}
	content, status, err := ReadTask(tasksDir, name, "")
	if err != nil || status != "blocked" {
		t.Fatalf("ReadTask: status %q, err %v", status, err)
	}
	if !strings.Contains(content, "## Blocked Reason\nwaiting for the API key") {
		t.Errorf("block reason not recorded: %s", content)
	}
	if err := TaskUnblock(tasksDir, name); err != nil {
		t.Fatal(err)
	}
	if err := TaskStart(tasksDir, name); err != nil {
		t.Fatal(err)
	}
	if err := TaskComplete(tasksDir, name); err != nil {
		t.Fatal(err)
	}
	if s, _, found := FindTask(tasksDir, name); !found || s != "done" {
		t.Errorf("task must end in done, got %q (found %v)", s, found)
	}
}

func TestListTasksReturnsFoldersOnly(t *testing.T) {
	tasksDir := newSpecTasks(t)
	AddTask(tasksDir, "main", "high", "a", "A")
	AddTask(tasksDir, "main", "low", "b", "B")
	if err := os.WriteFile(filepath.Join(tasksDir, "backlog", "stray.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	names, found, err := ListTasks(tasksDir, "backlog")
	if err != nil || !found {
		t.Fatalf("ListTasks: %v, %v", found, err)
	}
	if strings.Join(names, ",") != "high_a,low_b" {
		t.Errorf("names = %v, want high_a and low_b only", names)
	}
}

func TestListTasksMissingStatusIsNotFound(t *testing.T) {
	tasksDir := t.TempDir()
	if _, found, err := ListTasks(tasksDir, "doing"); err != nil || found {
		t.Errorf("a missing status folder is not found, got found=%v err=%v", found, err)
	}
}

func TestValidateStatusAndPriority(t *testing.T) {
	if err := ValidateStatus("blocked"); err != nil {
		t.Errorf("blocked is a status: %v", err)
	}
	if err := ValidateStatus("hold"); err == nil {
		t.Error("hold is not a status")
	}
	if err := ValidatePriority("medium"); err != nil {
		t.Errorf("medium is a priority: %v", err)
	}
	if err := ValidatePriority("backlog"); err == nil {
		t.Error("backlog is not a priority")
	}
}

func TestCreateStatusDirs(t *testing.T) {
	tasksDir := newSpecTasks(t)
	for _, s := range Statuses {
		if info, err := os.Stat(filepath.Join(tasksDir, s)); err != nil || !info.IsDir() {
			t.Errorf("status folder %s missing", s)
		}
	}
}
