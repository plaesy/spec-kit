package taskmanage

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// newTasksDir creates a tasks tree containing every status directory and
// returns its path. The bash twin's layout is created up front by the
// scaffolding step, so an empty-but-present tree is the normal case.
func newTasksDir(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, s := range Statuses {
		if err := os.MkdirAll(filepath.Join(root, s), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", s, err)
		}
	}
	return root
}

// writeTask creates <tasksDir>/<status>/<name> with the given content.
func writeTask(t *testing.T, tasksDir, status, name, content string) string {
	t.Helper()
	p := filepath.Join(tasksDir, status, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
	return p
}

// chdir switches the process working directory for the duration of the test.
// The package's cwd-sensitive functions (FindProjectRoot) have no injection
// seam, so the process cwd is the only control available; Go 1.20 has no
// t.Chdir. Callers must not use t.Parallel.
func chdir(t *testing.T, dir string) {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(prev); err != nil {
			t.Errorf("restore cwd to %s: %v", prev, err)
		}
	})
}

// sameDir compares two directory paths after resolving symlinks and
// case differences introduced by the OS (Windows temp dirs are 8.3-cased).
func sameDir(a, b string) bool {
	ra, err := filepath.EvalSymlinks(a)
	if err != nil {
		ra = a
	}
	rb, err := filepath.EvalSymlinks(b)
	if err != nil {
		rb = b
	}
	if ra == rb {
		return true
	}
	if strings.EqualFold(ra, rb) {
		return true
	}
	ia, err1 := os.Stat(ra)
	ib, err2 := os.Stat(rb)
	return err1 == nil && err2 == nil && os.SameFile(ia, ib)
}

// ancestorHasPlaesy reports whether any ancestor of dir (inclusive) holds a
// .plaesy directory, which would make the "project root not found" assertion
// environment-dependent.
func ancestorHasPlaesy(t *testing.T, dir string) bool {
	t.Helper()
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

func TestValidateStatus(t *testing.T) {
	tests := []struct {
		name    string
		status  string
		wantErr bool
	}{
		{"backlog accepted", "backlog", false},
		{"todo accepted", "todo", false},
		{"doing accepted", "doing", false},
		{"done accepted", "done", false},
		{"blocked accepted", "blocked", false},
		{"wrong case rejected", "Todo", true},
		{"empty rejected", "", true},
		{"unknown rejected", "archive", true},
		{"path traversal rejected", "../done", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStatus(tt.status)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateStatus(%q) error = %v, wantErr %v", tt.status, err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), "backlog, todo, doing, done, blocked") {
				t.Errorf("error should list the valid statuses, got %q", err)
			}
		})
	}
}

func TestTasksDir(t *testing.T) {
	got := TasksDir("/repo")
	want := filepath.Join("/repo", ".plaesy", "tasks")
	if got != want {
		t.Errorf("TasksDir = %q, want %q", got, want)
	}
}

func TestGetNextTask(t *testing.T) {
	tests := []struct {
		name    string
		files   []string
		noBackl bool // remove the backlog directory entirely
		want    string
		wantOK  bool
		wantErr string
	}{
		{name: "empty backlog", want: "", wantOK: false},
		{name: "single task", files: []string{"a.md"}, want: "a.md", wantOK: true},
		{
			name:  "reverse sorted first entry",
			files: []string{"001-alpha.md", "003-gamma.md", "002-beta.md"},
			want:  "003-gamma.md", wantOK: true,
		},
		{
			name:  "plain lexicographic not numeric",
			files: []string{"10-ten.md", "9-nine.md"},
			want:  "9-nine.md", wantOK: true,
		},
		{
			name:  "non markdown ignored",
			files: []string{"b.md", "notes.txt", "README.md"},
			want:  "b.md", wantOK: true,
		},
		{
			name:  "only non markdown is no task",
			files: []string{"notes.txt"},
			want:  "", wantOK: false,
		},
		{name: "missing backlog dir errors", noBackl: true, wantErr: "backlog directory not found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasksDir := newTasksDir(t)
			if tt.noBackl {
				if err := os.RemoveAll(filepath.Join(tasksDir, "backlog")); err != nil {
					t.Fatalf("remove backlog: %v", err)
				}
			}
			for _, f := range tt.files {
				writeTask(t, tasksDir, "backlog", f, "body\n")
			}

			got, ok, err := GetNextTask(tasksDir)
			switch {
			case tt.wantErr != "":
				if err == nil {
					t.Fatalf("expected error containing %q, got (%q, %v)", tt.wantErr, got, ok)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error = %q, want it to contain %q", err, tt.wantErr)
				}
				if ok {
					t.Error("found must be false alongside an error")
				}
			default:
				if err != nil {
					t.Fatalf("GetNextTask: %v", err)
				}
				if ok != tt.wantOK {
					t.Errorf("found = %v, want %v", ok, tt.wantOK)
				}
				if got != tt.want {
					t.Errorf("task = %q, want %q", got, tt.want)
				}
			}
		})
	}
}

// TestGetNextTaskIgnoresSubdirectories locks in that a subdirectory inside
// backlog is not mistaken for a task file (the walk is flat, as in the bash
// `ls | sort -r` twin).
func TestGetNextTaskIgnoresSubdirectories(t *testing.T) {
	tasksDir := newTasksDir(t)
	writeTask(t, tasksDir, "backlog", "a.md", "body\n")
	if err := os.MkdirAll(filepath.Join(tasksDir, "backlog", "z-archive.md"), 0o755); err != nil {
		t.Fatalf("mkdir decoy: %v", err)
	}

	got, ok, err := GetNextTask(tasksDir)
	if err != nil || !ok {
		t.Fatalf("GetNextTask = (%q, %v, %v), want a found task", got, ok, err)
	}
	if got != "a.md" {
		t.Errorf("task = %q, want a.md (directory named like a task must be ignored)", got)
	}
}

func TestMoveTask(t *testing.T) {
	tests := []struct {
		name      string
		from      string
		to        string
		seedFrom  string // status the task is seeded in; empty = seed nowhere
		seedName  string
		wantMoved bool
		wantErr   string
	}{
		{name: "backlog to todo", from: "backlog", to: "todo", seedFrom: "backlog", seedName: "t.md", wantMoved: true},
		{name: "todo to doing", from: "todo", to: "doing", seedFrom: "todo", seedName: "t.md", wantMoved: true},
		{name: "doing to done", from: "doing", to: "done", seedFrom: "doing", seedName: "t.md", wantMoved: true},
		{name: "blocked to todo", from: "blocked", to: "todo", seedFrom: "blocked", seedName: "t.md", wantMoved: true},
		{name: "invalid from status", from: "archive", to: "todo", seedFrom: "backlog", seedName: "t.md", wantErr: "invalid status: archive"},
		{name: "invalid to status", from: "backlog", to: "wip", seedFrom: "backlog", seedName: "t.md", wantErr: "invalid status: wip"},
		{name: "missing source file", from: "backlog", to: "todo", wantErr: "task not found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasksDir := newTasksDir(t)
			if tt.seedFrom != "" {
				writeTask(t, tasksDir, tt.seedFrom, tt.seedName, "body\n")
			}

			err := MoveTask(tasksDir, tt.seedName, tt.from, tt.to)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %q, want it to contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("MoveTask: %v", err)
			}
			if _, err := os.Stat(filepath.Join(tasksDir, tt.from, tt.seedName)); err == nil {
				t.Errorf("source file still present in %s", tt.from)
			}
			if _, err := os.Stat(filepath.Join(tasksDir, tt.to, tt.seedName)); err != nil {
				t.Errorf("task not found in %s after move: %v", tt.to, err)
			}
		})
	}
}

// TestMoveTaskCreatesMissingTargetDir covers the MkdirAll seam: a status
// directory that does not exist yet is created by the move.
func TestMoveTaskCreatesMissingTargetDir(t *testing.T) {
	tasksDir := newTasksDir(t)
	writeTask(t, tasksDir, "backlog", "t.md", "body\n")
	if err := os.RemoveAll(filepath.Join(tasksDir, "todo")); err != nil {
		t.Fatalf("remove todo: %v", err)
	}

	if err := MoveTask(tasksDir, "t.md", "backlog", "todo"); err != nil {
		t.Fatalf("MoveTask: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tasksDir, "todo", "t.md")); err != nil {
		t.Errorf("move into recreated directory failed: %v", err)
	}
}

// TestMoveTaskRejectsDirectory documents that a directory sitting where the
// task file is expected is reported as "task not found" rather than moved.
func TestMoveTaskRejectsDirectory(t *testing.T) {
	tasksDir := newTasksDir(t)
	if err := os.MkdirAll(filepath.Join(tasksDir, "backlog", "t.md"), 0o755); err != nil {
		t.Fatalf("mkdir decoy: %v", err)
	}
	err := MoveTask(tasksDir, "t.md", "backlog", "todo")
	if err == nil || !strings.Contains(err.Error(), "task not found") {
		t.Errorf("error = %v, want a task-not-found error", err)
	}
}

func TestPopNextTask(t *testing.T) {
	tests := []struct {
		name    string
		files   []string
		noBackl bool
		want    string
		wantOK  bool
		wantErr string
	}{
		{name: "empty backlog", want: "", wantOK: false},
		{
			name: "moves highest backlog entry to todo", files: []string{"a.md", "c.md", "b.md"},
			want: "c.md", wantOK: true,
		},
		{name: "missing backlog dir errors", noBackl: true, wantErr: "backlog directory not found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasksDir := newTasksDir(t)
			if tt.noBackl {
				if err := os.RemoveAll(filepath.Join(tasksDir, "backlog")); err != nil {
					t.Fatalf("remove backlog: %v", err)
				}
			}
			for _, f := range tt.files {
				writeTask(t, tasksDir, "backlog", f, "body\n")
			}

			got, ok, err := PopNextTask(tasksDir)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want it to contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("PopNextTask: %v", err)
			}
			if ok != tt.wantOK || got != tt.want {
				t.Errorf("PopNextTask = (%q, %v), want (%q, %v)", got, ok, tt.want, tt.wantOK)
			}
			if !ok {
				return
			}
			if _, err := os.Stat(filepath.Join(tasksDir, "todo", got)); err != nil {
				t.Errorf("task not moved to todo: %v", err)
			}
			if _, err := os.Stat(filepath.Join(tasksDir, "backlog", got)); err == nil {
				t.Error("task still present in backlog after pop")
			}
		})
	}
}

// TestPopNextTaskKeepsRemainingBacklog asserts the queue is a queue: only one
// task is promoted per call and the rest survive.
func TestPopNextTaskKeepsRemainingBacklog(t *testing.T) {
	tasksDir := newTasksDir(t)
	for _, f := range []string{"a.md", "b.md", "c.md"} {
		writeTask(t, tasksDir, "backlog", f, "body\n")
	}

	first, ok, err := PopNextTask(tasksDir)
	if err != nil || !ok {
		t.Fatalf("first PopNextTask = (%q, %v, %v)", first, ok, err)
	}
	if first != "c.md" {
		t.Errorf("first popped = %q, want c.md", first)
	}
	rest, ok, err := GetNextTask(tasksDir)
	if err != nil || !ok {
		t.Fatalf("GetNextTask after pop = (%q, %v, %v)", rest, ok, err)
	}
	if rest != "b.md" {
		t.Errorf("next after pop = %q, want b.md", rest)
	}
}

func TestTaskLifecycleTransitions(t *testing.T) {
	tests := []struct {
		name string
		// move runs the transition under test and returns an error, if any.
		move func(tasksDir, task string) error
		// seed is the status the task starts in.
		seed    string
		want    string
		wantErr string
	}{
		{
			name: "start promotes todo to doing", seed: "todo", want: "doing",
			move: func(d, f string) error { return TaskStart(d, f) },
		},
		{
			name: "complete promotes doing to done", seed: "doing", want: "done",
			move: func(d, f string) error { return TaskComplete(d, f) },
		},
		{
			name: "unblock returns blocked to todo", seed: "blocked", want: "todo",
			move: func(d, f string) error { return TaskUnblock(d, f) },
		},
		{
			name: "start fails when task is not in todo", seed: "backlog", wantErr: "task not found",
			move: func(d, f string) error { return TaskStart(d, f) },
		},
		{
			name: "complete fails when task is not in doing", seed: "todo", wantErr: "task not found",
			move: func(d, f string) error { return TaskComplete(d, f) },
		},
		{
			name: "unblock fails when task is not in blocked", seed: "doing", wantErr: "task not found",
			move: func(d, f string) error { return TaskUnblock(d, f) },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasksDir := newTasksDir(t)
			writeTask(t, tasksDir, tt.seed, "t.md", "# task\n")

			err := tt.move(tasksDir, "t.md")
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want it to contain %q", err, tt.wantErr)
				}
				if _, err := os.Stat(filepath.Join(tasksDir, tt.seed, "t.md")); err != nil {
					t.Errorf("failed transition must not move the file: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("transition: %v", err)
			}
			if _, err := os.Stat(filepath.Join(tasksDir, tt.want, "t.md")); err != nil {
				t.Errorf("task not found in %s after transition: %v", tt.want, err)
			}
		})
	}
}

func TestTaskBlockAppendsReason(t *testing.T) {
	tests := []struct {
		name        string
		reason      string
		seedContent string
		wantSuffix  string
	}{
		{
			name:        "explicit reason appended",
			reason:      "waiting on API key",
			seedContent: "# task\n\nbody\n",
			wantSuffix:  "\n## Blocked Reason\nwaiting on API key\n",
		},
		{
			name:        "empty reason gets the bash default",
			reason:      "",
			seedContent: "# task\n",
			wantSuffix:  "\n## Blocked Reason\nNo reason provided\n",
		},
		{
			name:        "multiline reason preserved verbatim",
			reason:      "line one\nline two",
			seedContent: "",
			wantSuffix:  "\n## Blocked Reason\nline one\nline two\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasksDir := newTasksDir(t)
			writeTask(t, tasksDir, "doing", "t.md", tt.seedContent)

			if err := TaskBlock(tasksDir, "t.md", tt.reason); err != nil {
				t.Fatalf("TaskBlock: %v", err)
			}

			blocked := filepath.Join(tasksDir, "blocked", "t.md")
			if _, err := os.Stat(blocked); err != nil {
				t.Fatalf("task not moved to blocked: %v", err)
			}
			if _, err := os.Stat(filepath.Join(tasksDir, "doing", "t.md")); err == nil {
				t.Error("task still present in doing after block")
			}
			got, err := os.ReadFile(blocked)
			if err != nil {
				t.Fatalf("read blocked task: %v", err)
			}
			if want := tt.seedContent + tt.wantSuffix; string(got) != want {
				t.Errorf("blocked file content = %q, want %q", got, want)
			}
		})
	}
}

// TestTaskBlockAppendsOnEveryCall documents that the reason section is
// appended, not replaced: unblocking and re-blocking adds a second section.
// The bash twin behaves the same way (cat >> in task_block).
func TestTaskBlockAppendsOnEveryCall(t *testing.T) {
	tasksDir := newTasksDir(t)
	writeTask(t, tasksDir, "doing", "t.md", "# task\n")

	if err := TaskBlock(tasksDir, "t.md", "first"); err != nil {
		t.Fatalf("first block: %v", err)
	}
	if err := TaskUnblock(tasksDir, "t.md"); err != nil {
		t.Fatalf("unblock: %v", err)
	}
	if err := TaskStart(tasksDir, "t.md"); err != nil {
		t.Fatalf("start: %v", err)
	}
	if err := TaskBlock(tasksDir, "t.md", "second"); err != nil {
		t.Fatalf("second block: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(tasksDir, "blocked", "t.md"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if n := strings.Count(string(got), "## Blocked Reason"); n != 2 {
		t.Errorf("Blocked Reason sections = %d, want 2 (append, not replace):\n%s", n, got)
	}
}

// TestTaskBlockFailsWithoutDoingTask proves the reason is never written when
// the move itself fails.
func TestTaskBlockFailsWithoutDoingTask(t *testing.T) {
	tasksDir := newTasksDir(t)
	writeTask(t, tasksDir, "todo", "t.md", "# task\n")

	err := TaskBlock(tasksDir, "t.md", "reason")
	if err == nil || !strings.Contains(err.Error(), "task not found") {
		t.Fatalf("error = %v, want a task-not-found error", err)
	}
	if got, _ := os.ReadFile(filepath.Join(tasksDir, "todo", "t.md")); string(got) != "# task\n" {
		t.Errorf("untouched task was modified: %q", got)
	}
}

func TestListTasks(t *testing.T) {
	tests := []struct {
		name        string
		seed        map[string][]string
		dropDir     bool
		status      string
		want        []string
		wantEmpty   bool
		wantFound   bool
		wantErr     string
		wantNilList bool
	}{
		{
			name: "basenames sorted", status: "todo", wantFound: true,
			seed: map[string][]string{"todo": {"b.md", "a.md", "c.md"}},
			want: []string{"a", "b", "c"},
		},
		{
			name: "extension stripped only once", status: "todo", wantFound: true,
			seed: map[string][]string{"todo": {"a.md.md"}},
			want: []string{"a.md"},
		},
		{
			name: "non markdown ignored", status: "todo", wantFound: true,
			seed: map[string][]string{"todo": {"a.md", "notes.txt"}},
			want: []string{"a"},
		},
		{
			name: "empty existing dir is found but empty", status: "doing", wantFound: true, wantEmpty: true,
			seed: map[string][]string{},
		},
		{
			name: "missing dir is not found", status: "blocked", dropDir: true,
			wantNilList: true,
		},
		{
			name: "invalid status errors", status: "nope", wantErr: "invalid status: nope",
			wantNilList: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasksDir := newTasksDir(t)
			for status, files := range tt.seed {
				for _, f := range files {
					writeTask(t, tasksDir, status, f, "body\n")
				}
			}
			if tt.dropDir {
				if err := os.RemoveAll(filepath.Join(tasksDir, tt.status)); err != nil {
					t.Fatalf("remove %s: %v", tt.status, err)
				}
			}

			got, found, err := ListTasks(tasksDir, tt.status)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want it to contain %q", err, tt.wantErr)
				}
				if found {
					t.Error("found must be false alongside an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ListTasks: %v", err)
			}
			if found != tt.wantFound {
				t.Errorf("found = %v, want %v", found, tt.wantFound)
			}
			if !tt.wantFound && tt.wantNilList && got != nil {
				t.Errorf("list = %v, want nil for a missing directory", got)
			}
			if tt.wantEmpty && len(got) != 0 {
				t.Errorf("list = %v, want empty", got)
			}
			if tt.want != nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("list = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestListTasksIgnoresSubdirectory covers the walk staying flat: a directory
// named like a task file is not listed.
func TestListTasksIgnoresSubdirectory(t *testing.T) {
	tasksDir := newTasksDir(t)
	writeTask(t, tasksDir, "todo", "a.md", "body\n")
	if err := os.MkdirAll(filepath.Join(tasksDir, "todo", "b.md"), 0o755); err != nil {
		t.Fatalf("mkdir decoy: %v", err)
	}

	got, found, err := ListTasks(tasksDir, "todo")
	if err != nil || !found {
		t.Fatalf("ListTasks = (%v, %v, %v)", got, found, err)
	}
	if !reflect.DeepEqual(got, []string{"a"}) {
		t.Errorf("list = %v, want [a] (directory named like a task must be ignored)", got)
	}
}

func TestFindProjectRoot(t *testing.T) {
	t.Run("walks up to the nearest .plaesy", func(t *testing.T) {
		base := t.TempDir()
		root := filepath.Join(base, "repo")
		deep := filepath.Join(root, "a", "b")
		if err := os.MkdirAll(filepath.Join(root, ".plaesy"), 0o755); err != nil {
			t.Fatalf("mkdir .plaesy: %v", err)
		}
		if err := os.MkdirAll(deep, 0o755); err != nil {
			t.Fatalf("mkdir deep: %v", err)
		}
		chdir(t, deep)

		got, err := FindProjectRoot()
		if err != nil {
			t.Fatalf("FindProjectRoot: %v", err)
		}
		if !sameDir(got, root) {
			t.Errorf("root = %q, want %q", got, root)
		}
	})

	t.Run("uses cwd itself when it holds .plaesy", func(t *testing.T) {
		base := t.TempDir()
		if err := os.MkdirAll(filepath.Join(base, ".plaesy"), 0o755); err != nil {
			t.Fatalf("mkdir .plaesy: %v", err)
		}
		chdir(t, base)

		got, err := FindProjectRoot()
		if err != nil {
			t.Fatalf("FindProjectRoot: %v", err)
		}
		if !sameDir(got, base) {
			t.Errorf("root = %q, want %q", got, base)
		}
	})

	t.Run("errors when no ancestor holds .plaesy", func(t *testing.T) {
		base := t.TempDir()
		if ancestorHasPlaesy(t, base) {
			t.Skip("an ancestor of the temp dir holds a .plaesy directory; cannot assert the not-found path")
		}
		chdir(t, base)

		got, err := FindProjectRoot()
		if err == nil {
			t.Fatalf("expected an error, got root %q", got)
		}
		if !strings.Contains(err.Error(), "project root not found") {
			t.Errorf("error = %q, want it to mention the project root", err)
		}
		if got != "" {
			t.Errorf("root = %q, want empty alongside the error", got)
		}
	})
}

// TestFindProjectRootPrefersNearest walks two nested .plaesy directories and
// asserts the closest one wins.
func TestFindProjectRootPrefersNearest(t *testing.T) {
	outer := t.TempDir()
	inner := filepath.Join(outer, "inner")
	deep := filepath.Join(inner, "src")
	for _, p := range []string{outer, inner} {
		if err := os.MkdirAll(filepath.Join(p, ".plaesy"), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", p, err)
		}
	}
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatalf("mkdir deep: %v", err)
	}
	chdir(t, deep)

	got, err := FindProjectRoot()
	if err != nil {
		t.Fatalf("FindProjectRoot: %v", err)
	}
	if !sameDir(got, inner) {
		t.Errorf("root = %q, want the nearest .plaesy (%q)", got, inner)
	}
}

// TestMoveTaskOntoExistingDirectory proves the move fails (rather than
// clobbering) when a directory already occupies the destination name.
func TestMoveTaskOntoExistingDirectory(t *testing.T) {
	tasksDir := newTasksDir(t)
	writeTask(t, tasksDir, "backlog", "t.md", "body\n")
	if err := os.MkdirAll(filepath.Join(tasksDir, "todo", "t.md"), 0o755); err != nil {
		t.Fatalf("mkdir destination decoy: %v", err)
	}

	if err := MoveTask(tasksDir, "t.md", "backlog", "todo"); err == nil {
		t.Fatal("expected the move onto an existing directory to fail")
	}
	if _, err := os.Stat(filepath.Join(tasksDir, "backlog", "t.md")); err != nil {
		t.Errorf("source file must survive a failed move: %v", err)
	}
}

func TestFindTask(t *testing.T) {
	tests := []struct {
		name       string
		seedStatus string
		task       string
		wantStatus string
		wantFound  bool
	}{
		{name: "found in todo", seedStatus: "todo", task: "t.md", wantStatus: "todo", wantFound: true},
		{name: "found in done", seedStatus: "done", task: "t.md", wantStatus: "done", wantFound: true},
		{name: "missing task", task: "nope.md", wantFound: false},
		{name: "directory named like a task is not a task", task: "dir.md", wantFound: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasksDir := newTasksDir(t)
			if tt.seedStatus != "" {
				writeTask(t, tasksDir, tt.seedStatus, tt.task, "body\n")
			}
			if tt.task == "dir.md" {
				if err := os.MkdirAll(filepath.Join(tasksDir, "todo", "dir.md"), 0o755); err != nil {
					t.Fatalf("mkdir decoy: %v", err)
				}
			}

			status, path, found := FindTask(tasksDir, tt.task)
			if found != tt.wantFound {
				t.Fatalf("found = %v, want %v", found, tt.wantFound)
			}
			if status != tt.wantStatus {
				t.Errorf("status = %q, want %q", status, tt.wantStatus)
			}
			if tt.wantFound {
				if want := filepath.Join(tasksDir, tt.wantStatus, tt.task); path != want {
					t.Errorf("path = %q, want %q", path, want)
				}
			} else if path != "" {
				t.Errorf("path = %q, want empty when not found", path)
			}
		})
	}
}

// TestFindTaskPrefersStatusOrder documents the search order: a file present in
// two status directories resolves to the earlier status in Statuses.
func TestFindTaskPrefersStatusOrder(t *testing.T) {
	tasksDir := newTasksDir(t)
	writeTask(t, tasksDir, "todo", "t.md", "in todo\n")
	writeTask(t, tasksDir, "backlog", "t.md", "in backlog\n")

	status, path, found := FindTask(tasksDir, "t.md")
	if !found {
		t.Fatal("expected the task to be found")
	}
	if status != "backlog" {
		t.Errorf("status = %q, want backlog (Statuses order wins)", status)
	}
	if want := filepath.Join(tasksDir, "backlog", "t.md"); path != want {
		t.Errorf("path = %q, want %q", path, want)
	}
}

func TestReadTask(t *testing.T) {
	const body = "# task\n\nsome content\n"
	tests := []struct {
		name       string
		seedStatus string
		status     string // requested status; "" = search all
		task       string
		wantStatus string
		wantBody   string
		wantErr    string
	}{
		{
			name: "explicit status", seedStatus: "doing", status: "doing", task: "t.md",
			wantStatus: "doing", wantBody: body,
		},
		{
			name: "empty status searches every status", seedStatus: "blocked", status: "", task: "t.md",
			wantStatus: "blocked", wantBody: body,
		},
		{
			name: "explicit status wins over search order", seedStatus: "todo", status: "todo", task: "t.md",
			wantStatus: "todo", wantBody: body,
		},
		{
			name: "explicit status that does not hold the task errors", seedStatus: "todo", status: "done", task: "t.md",
			wantErr: "task not found: t.md",
		},
		{
			name: "unknown task with explicit status errors", seedStatus: "todo", status: "todo", task: "nope.md",
			wantErr: "task not found: nope.md",
		},
		{
			name: "unknown task with no status errors", task: "nope.md", wantErr: "task not found: nope.md",
		},
		{
			name: "invalid explicit status errors", seedStatus: "todo", status: "nope", task: "t.md",
			wantErr: "invalid status: nope",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasksDir := newTasksDir(t)
			if tt.seedStatus != "" {
				writeTask(t, tasksDir, tt.seedStatus, "t.md", body)
			}

			got, gotStatus, err := ReadTask(tasksDir, tt.task, tt.status)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want it to contain %q", err, tt.wantErr)
				}
				if got != "" || gotStatus != "" {
					t.Errorf("error path returned (%q, %q), want both empty", got, gotStatus)
				}
				return
			}
			if err != nil {
				t.Fatalf("ReadTask: %v", err)
			}
			if got != tt.wantBody {
				t.Errorf("content = %q, want %q", got, tt.wantBody)
			}
			if gotStatus != tt.wantStatus {
				t.Errorf("status = %q, want %q", gotStatus, tt.wantStatus)
			}
		})
	}
}

// TestReadTaskRejectsDirectory documents that a directory in the place of a
// task file is reported as not found rather than read.
func TestReadTaskRejectsDirectory(t *testing.T) {
	tasksDir := newTasksDir(t)
	if err := os.MkdirAll(filepath.Join(tasksDir, "todo", "t.md"), 0o755); err != nil {
		t.Fatalf("mkdir decoy: %v", err)
	}
	_, _, err := ReadTask(tasksDir, "t.md", "todo")
	if err == nil || !strings.Contains(err.Error(), "task not found") {
		t.Errorf("error = %v, want a task-not-found error", err)
	}
}
