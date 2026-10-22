package tasks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTaskFolder creates a task folder under tasksDir, with a spec.md carrying the spec and status.
func writeTaskFolder(t *testing.T, tasksDir, spec, status, name, taskStatus string) {
	t.Helper()
	dir := filepath.Join(tasksDir, status, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\nspec: " + spec + "\ntitle: " + name + "\nphase: implement\nstatus: " + taskStatus + "\n---\n"
	if err := os.WriteFile(filepath.Join(dir, TaskFileName), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestListAllTasksShowsTheSpecOfEachTask(t *testing.T) {
	tasksDir := t.TempDir()
	writeTaskFolder(t, tasksDir, "auth-login", "doing", "high_login-form", "doing")
	writeTaskFolder(t, tasksDir, "auth-login", "backlog", "low_reset", "backlog")
	writeTaskFolder(t, tasksDir, "billing", "done", "medium_invoice", "done")

	refs, err := ListAllTasks(tasksDir)
	if err != nil || len(refs) != 3 {
		t.Fatalf("ListAllTasks: %d refs, err %v", len(refs), err)
	}
	if refs[0].Name != "low_reset" || refs[0].Spec != "auth-login" {
		t.Errorf("first = %+v, want the backlog task of auth-login", refs[0])
	}
	if refs[1].Name != "high_login-form" || refs[1].Phase != "implement" {
		t.Errorf("second = %+v", refs[1])
	}
}

func TestCheckIntegrityFindsProblems(t *testing.T) {
	tasksDir := t.TempDir()
	// Frontmatter says done, folder says doing.
	writeTaskFolder(t, tasksDir, "auth-login", "doing", "high_a", "done")
	// Same task name in two statuses.
	writeTaskFolder(t, tasksDir, "auth-login", "todo", "high_b", "todo")
	writeTaskFolder(t, tasksDir, "auth-login", "doing", "high_b", "doing")
	// A task folder with no spec.md.
	if err := os.MkdirAll(filepath.Join(tasksDir, "done", "empty"), 0o755); err != nil {
		t.Fatal(err)
	}

	problems, err := CheckIntegrity(tasksDir)
	if err != nil {
		t.Fatal(err)
	}
	var all []string
	for _, p := range problems {
		all = append(all, p.Message)
	}
	joined := strings.Join(all, "\n")
	for _, want := range []string{"frontmatter status done, folder is doing", "in both todo and doing", "no spec.md in the task folder"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing problem %q in:\n%s", want, joined)
		}
	}
	if len(problems) != 3 {
		t.Errorf("want 3 problems, got %d:\n%s", len(problems), joined)
	}
}
