package taskmanage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makeSpecs creates one folder per name under <root>/.plaesy/specs.
func makeSpecs(t *testing.T, root string, names ...string) {
	t.Helper()
	for _, n := range names {
		if err := os.MkdirAll(filepath.Join(root, ".plaesy", "specs", n), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", n, err)
		}
	}
}

func TestResolveSpec(t *testing.T) {
	tests := []struct {
		name    string
		specs   []string
		branch  string
		want    string
		wantErr string
	}{
		{"timestamp spec matches its branch", []string{"20261015-143022-auth-login"}, "auth-login", "20261015-143022-auth-login", ""},
		{"legacy folder named exactly after branch", []string{"001-add-login"}, "001-add-login", "001-add-login", ""},
		{"no spec for branch", []string{"20261015-143022-auth-login"}, "billing", "", "no spec for branch"},
		{"two specs for one branch", []string{"20261015-143022-auth-login", "20261016-090000-auth-login"}, "auth-login", "", "matches several specs"},
		{"branch that is a suffix of another is not a match", []string{"20261015-143022-login"}, "auth-login", "", "no spec for branch"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			makeSpecs(t, root, tt.specs...)
			got, err := ResolveSpec(root, tt.branch)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("ResolveSpec = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCreateStatusDirs(t *testing.T) {
	tasks := filepath.Join(t.TempDir(), "tasks")
	if err := CreateStatusDirs(tasks); err != nil {
		t.Fatalf("CreateStatusDirs: %v", err)
	}
	for _, s := range Statuses {
		if info, err := os.Stat(filepath.Join(tasks, s)); err != nil || !info.IsDir() {
			t.Errorf("status dir %s missing: %v", s, err)
		}
	}
}

func TestAddTask(t *testing.T) {
	tasks := newTasksDir(t)

	name, err := AddTask(tasks, "high", "login-form", "Build the login form")
	if err != nil {
		t.Fatalf("AddTask: %v", err)
	}
	if name != "high_login-form.md" {
		t.Errorf("file name = %q, want high_login-form.md", name)
	}
	b, err := os.ReadFile(filepath.Join(tasks, "backlog", name))
	if err != nil {
		t.Fatalf("read added task: %v", err)
	}
	if !strings.Contains(string(b), "title: Build the login form") || !strings.Contains(string(b), "status: backlog") {
		t.Errorf("task frontmatter missing title or status:\n%s", b)
	}

	if _, err := AddTask(tasks, "high", "login-form", "again"); err == nil {
		t.Error("adding an existing task succeeded, want an error")
	}
	if _, err := AddTask(tasks, "urgent", "x", "bad priority"); err == nil {
		t.Error("invalid priority accepted")
	}
	if _, err := AddTask(tasks, "low", "../escape", "bad slug"); err == nil {
		t.Error("slug with a path separator accepted")
	}
}
