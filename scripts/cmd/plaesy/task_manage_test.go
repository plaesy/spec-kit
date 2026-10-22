package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/plaesy/spec-kit/internal/tasks"
)

// makeTasksTree builds a project with a .plaesy directory and the task folders in
// seed. seed maps a status to task names; each task gets a spec.md and a
// contracts.json. Returns the project root. Tests chdir into it, so the commands
// find it from the working directory.
func makeTasksTree(t *testing.T, seed map[string][]string) string {
	t.Helper()
	dir := tempProject(t)
	tasksDir := tasks.TasksDir(dir)
	if err := tasks.CreateStatusDirs(tasksDir); err != nil {
		t.Fatal(err)
	}
	for status, names := range seed {
		for _, name := range names {
			folder := filepath.Join(tasksDir, status, name)
			if err := os.MkdirAll(folder, 0o755); err != nil {
				t.Fatal(err)
			}
			doc := "---\nspec: main\ntitle: " + name + "\nstatus: " + status + "\n---\n"
			if err := os.WriteFile(filepath.Join(folder, tasks.TaskFileName), []byte(doc), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(folder, tasks.ContractFileName), []byte(`{"id":"x","checklist":[]}`), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	return dir
}

// runTasks runs `plaesy tasks <args...>` in-process from dir and returns stdout and the error.
func runTasks(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	withWorkingDirectory(t, dir)
	cmd := newTasksCmd()
	cmd.SetArgs(args)
	return captureStdout(t, func() error {
		return cmd.Execute()
	})
}

func TestTasksDirIsTheProjectTaskFolder(t *testing.T) {
	dir := makeTasksTree(t, nil)
	withWorkingDirectory(t, dir)
	got, err := tasksDirOrErr()
	if err != nil {
		t.Fatalf("tasksDirOrErr: %v", err)
	}
	if !strings.HasSuffix(filepath.ToSlash(got), ".plaesy/tasks") {
		t.Errorf("tasksDirOrErr = %q, want the project .plaesy/tasks folder", got)
	}
}

func TestTasksListShowsEveryStatus(t *testing.T) {
	dir := makeTasksTree(t, map[string][]string{
		"backlog": {"low_a"},
		"doing":   {"high_b"},
	})
	out, err := runTasks(t, dir, "list")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	for _, want := range []string{"low_a", "high_b"} {
		if !strings.Contains(out, want) {
			t.Errorf("list does not show %q:\n%s", want, out)
		}
	}
}

func TestTasksListStatusOnlyShowsThatStatus(t *testing.T) {
	dir := makeTasksTree(t, map[string][]string{"todo": {"high_a"}, "done": {"high_b"}})
	out, err := runTasks(t, dir, "list-status", "todo")
	if err != nil {
		t.Fatalf("list-status: %v", err)
	}
	if !strings.Contains(out, "high_a") || strings.Contains(out, "high_b") {
		t.Errorf("list-status todo = %q", out)
	}
}

func TestTasksNextPromotesMostUrgent(t *testing.T) {
	dir := makeTasksTree(t, map[string][]string{"backlog": {"low_a", "critical_c"}})
	out, err := runTasks(t, dir, "next")
	if err != nil {
		t.Fatalf("next: %v", err)
	}
	if !strings.Contains(out, "critical_c (backlog → todo)") {
		t.Errorf("next did not move the critical task: %q", out)
	}
	if _, err := os.Stat(filepath.Join(tasks.TasksDir(dir), "todo", "critical_c")); err != nil {
		t.Errorf("critical_c not in todo: %v", err)
	}
}

func TestTasksStartCompleteAndBlock(t *testing.T) {
	dir := makeTasksTree(t, map[string][]string{"todo": {"high_t"}})
	if _, err := runTasks(t, dir, "start", "high_t"); err != nil {
		t.Fatalf("start: %v", err)
	}
	if _, err := runTasks(t, dir, "block", "high_t", "waiting on API key"); err != nil {
		t.Fatalf("block: %v", err)
	}
	doc, err := os.ReadFile(filepath.Join(tasks.TasksDir(dir), "blocked", "high_t", tasks.TaskFileName))
	if err != nil || !strings.Contains(string(doc), "waiting on API key") {
		t.Fatalf("block reason not recorded: %v %q", err, doc)
	}
	if _, err := runTasks(t, dir, "unblock", "high_t"); err != nil {
		t.Fatalf("unblock: %v", err)
	}
	if _, err := runTasks(t, dir, "start", "high_t"); err != nil {
		t.Fatalf("start again: %v", err)
	}
	if _, err := runTasks(t, dir, "complete", "high_t"); err != nil {
		t.Fatalf("complete: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tasks.TasksDir(dir), "done", "high_t")); err != nil {
		t.Errorf("high_t not in done: %v", err)
	}
}

func TestTasksStartRefusesTaskNotInTodo(t *testing.T) {
	dir := makeTasksTree(t, map[string][]string{"backlog": {"high_t"}})
	if _, err := runTasks(t, dir, "start", "high_t"); err == nil {
		t.Error("start must refuse a task that is not in todo")
	}
}

func TestTasksSubcommandsReportMissingProjectRoot(t *testing.T) {
	dir := t.TempDir()
	if ancestorHasPlaesy(dir) {
		t.Skip("an ancestor of this temp dir holds a .plaesy directory")
	}
	withWorkingDirectory(t, dir)
	if _, err := tasksDirOrErr(); err == nil || !strings.Contains(err.Error(), "project root not found") {
		t.Errorf("want project root not found, got %v", err)
	}
}

// ancestorHasPlaesy reports whether dir or any ancestor holds a .plaesy directory.
func ancestorHasPlaesy(dir string) bool {
	for cur := dir; ; {
		if info, err := os.Stat(filepath.Join(cur, ".plaesy")); err == nil && info.IsDir() {
			return true
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return false
		}
		cur = parent
	}
}
