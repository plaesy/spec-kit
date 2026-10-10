package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/plaesy/spec-kit/internal/taskmanage"
)

// Tests for the `plaesy task-manage` command surface: the wiring between cobra
// and internal/taskmanage, and the output the user actually sees. The package
// itself is covered at 92% (taskmanage_test.go), so what is untested here is
// exactly the glue — the per-subcommand RunE, the printTaskList format, and the
// tasksDirOrErr resolution from cwd. These exercises run in-process via
// newTasksCmd() and chdir into a temp fixture so FindProjectRoot resolves
// the fixture rather than the real repository. Callers must not use t.Parallel:
// every command resolves the project root from os.Getwd(), and only one test
// owns the process cwd at a time.

// taskFile is one task placed in a status directory by the fixture. The ".md"
// suffix is part of the name on disk but ListTasks strips it, so the visible
// task name is shorter than the file name — tests assert on the stripped form.
type taskFile struct {
	name    string
	content string
}

// makeTasksTree builds a project that looks like one a user ran `plaesy init`
// into: a real .plaesy directory with a tasks/ tree holding every status
// subdirectory. The seed places individual task files; statuses with no seed get
// an empty directory. Returns the project root — chdir into it (via
// withWorkingDirectory, which runTaskManage does) so tasksDirOrErr resolves it.
func makeTasksTree(t *testing.T, seed map[string][]taskFile) string {
	t.Helper()
	dir := tempProject(t)
	tasksDir := filepath.Join(dir, ".plaesy", "tasks")
	for _, s := range taskmanage.Statuses {
		if err := os.MkdirAll(filepath.Join(tasksDir, s), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", s, err)
		}
	}
	for status, files := range seed {
		for _, tf := range files {
			path := filepath.Join(tasksDir, status, tf.name)
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
			}
			if err := os.WriteFile(path, []byte(tf.content), 0o644); err != nil {
				t.Fatalf("write %s: %v", tf.name, err)
			}
		}
	}
	return dir
}

// tasksDirOf returns PROJECT_ROOT/.plaesy/tasks for a fixture built by
// makeTasksTree, so filesystem assertions can address the tree directly
// without re-deriving the path inside (and re-chdir'ing for) each command.
func tasksDirOf(projectRoot string) string {
	return filepath.Join(projectRoot, ".plaesy", "tasks")
}

// runTaskManage chdirs into dir so tasksDirOrErr resolves the fixture from cwd,
// then runs `plaesy tasks <args...>` in-process and returns stdout plus
// the error. The error is what the process exit code is built from, so it is
// the contract these glue tests assert on — not a printed banner.
func runTasks(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	withWorkingDirectory(t, dir)
	cmd := newTasksCmd()
	cmd.SetArgs(args)
	return captureStdout(t, func() error {
		return cmd.Execute()
	})
}

// ancestorHasPlaesy reports whether any ancestor of dir (inclusive) holds a
// .plaesy directory. The "project root not found" branch of tasksDirOrErr can
// only be asserted from a tree that genuinely has no project root above it;
// otherwise the environment makes the negative assertion meaningless.
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

// assertSameFile fails the test unless a and b resolve to the same file. On
// Windows the temp tree can surface through 8.3 short names, so a literal
// string compare of the path tasksDirOrErr returns against the path the fixture
// built would be a comparison of spellings, not of locations.
func assertSameFile(t *testing.T, a, b string) {
	t.Helper()
	ai, err1 := os.Stat(a)
	bi, err2 := os.Stat(b)
	if err1 != nil || err2 != nil {
		t.Errorf("cannot compare %q and %q: %v / %v", a, b, err1, err2)
		return
	}
	if !os.SameFile(ai, bi) {
		t.Errorf("got %q, want the same file as %q", a, b)
	}
}

func TestTasksDirOrErr(t *testing.T) {
	t.Run("found returns tasks dir when .plaesy is present", func(t *testing.T) {
		dir := makeTasksTree(t, nil)
		withWorkingDirectory(t, dir)
		got, err := tasksDirOrErr()
		if err != nil {
			t.Fatalf("tasksDirOrErr: %v", err)
		}
		// The error path must not fire when the project root is in fact present,
		// and the returned path must be PROJECT_ROOT/.plaesy/tasks — never the
		// project root itself, which would let every subcommand read the wrong
		// directory.
		assertSameFile(t, got, tasksDirOf(dir))
		if filepath.Base(got) != "tasks" {
			t.Errorf("tasksDirOrErr base = %q, want \"tasks\"", filepath.Base(got))
		}
	})

	t.Run("not found errors and names the directory it looked for", func(t *testing.T) {
		dir := tempProject(t) // no .plaesy created
		if ancestorHasPlaesy(dir) {
			t.Skip("an ancestor of this temp dir holds a .plaesy directory")
		}
		withWorkingDirectory(t, dir)
		got, err := tasksDirOrErr()
		if err == nil {
			t.Fatalf("tasksDirOrErr = %q, want an error", got)
		}
		if got != "" {
			t.Errorf("tasksDirOrErr returned %q alongside the error, want empty", got)
		}
		// The message must name ".plaeasy" so a user running from the wrong
		// directory knows which marker to create, not just that something failed.
		if !strings.Contains(err.Error(), ".plaesy") {
			t.Errorf("error %q does not name the .plaesy directory it looked for", err)
		}
	})
}

func TestPrintTaskList(t *testing.T) {
	t.Run("lists tasks sorted with a count header", func(t *testing.T) {
		dir := makeTasksTree(t, map[string][]taskFile{
			"todo": {{"b.md", ""}, {"a.md", ""}, {"c.md", ""}},
		})
		out, err := captureStdout(t, func() error {
			return printTaskList(tasksDirOf(dir), "todo")
		})
		if err != nil {
			t.Fatalf("printTaskList: %v", err)
		}
		// The format is "status (N)\n  - name\n..." with the .md suffix
		// stripped and the names sorted — the exact column width the `list`
		// command composes from these lines.
		want := "todo (3)\n  - a\n  - b\n  - c\n"
		if out != want {
			t.Errorf("output =\n%q\nwant\n%q", out, want)
		}
	})

	t.Run("empty but existing dir prints a zero count", func(t *testing.T) {
		dir := makeTasksTree(t, nil)
		out, err := captureStdout(t, func() error {
			return printTaskList(tasksDirOf(dir), "doing")
		})
		if err != nil {
			t.Fatalf("printTaskList: %v", err)
		}
		// found=true with zero entries is distinguishable from a missing
		// directory: it prints the header rather than the "[!]" banner, which
		// is the difference between "no work here" and "not a status".
		if out != "doing (0)\n" {
			t.Errorf("output = %q, want %q", out, "doing (0)\n")
		}
	})

	t.Run("missing dir prints the no-tasks banner", func(t *testing.T) {
		dir := makeTasksTree(t, nil)
		tasksDir := tasksDirOf(dir)
		if err := os.RemoveAll(filepath.Join(tasksDir, "blocked")); err != nil {
			t.Fatal(err)
		}
		out, err := captureStdout(t, func() error {
			return printTaskList(tasksDir, "blocked")
		})
		if err != nil {
			t.Fatalf("printTaskList: %v", err)
		}
		if out != "[!] No tasks in blocked\n" {
			t.Errorf("output = %q, want %q", out, "[!] No tasks in blocked\n")
		}
	})

	t.Run("invalid status errors before printing", func(t *testing.T) {
		dir := makeTasksTree(t, nil)
		out, err := captureStdout(t, func() error {
			return printTaskList(tasksDirOf(dir), "archive")
		})
		if err == nil {
			t.Fatal("expected an error for an invalid status")
		}
		if !strings.Contains(err.Error(), "invalid status: archive") {
			t.Errorf("error = %q, want it to contain \"invalid status: archive\"", err)
		}
		// Nothing is printed on the error path: a half-printed header would
		// look like a successful list of nothing.
		if out != "" {
			t.Errorf("expected no stdout on error, got %q", out)
		}
	})

	t.Run("extension stripped only once", func(*testing.T) {
		dir := makeTasksTree(t, map[string][]taskFile{
			"todo": {{"a.md.md", ""}},
		})
		out, err := captureStdout(t, func() error {
			return printTaskList(tasksDirOf(dir), "todo")
		})
		// ListTasks uses TrimSuffix(".md"), which strips one suffix: "a.md.md"
		// becomes "a.md", not "a". A filepath.Ext split would lose it.
		if err != nil {
			t.Fatalf("printTaskList: %v", err)
		}
		if !strings.Contains(out, "  - a.md") {
			t.Errorf("output = %q, want the .md stripped only once (a.md)", out)
		}
	})
}

func TestTaskManageList(t *testing.T) {
	t.Run("prints a heading per status with its tasks", func(t *testing.T) {
		dir := makeTasksTree(t, map[string][]taskFile{
			"backlog": {{"a.md", "backlog task\n"}},
			"todo":    {{"b.md", "todo task\n"}},
			"doing":   {{"c.md", "doing task\n"}},
			"done":    {{"d.md", "done task\n"}},
			"blocked": {{"e.md", "blocked task\n"}},
		})
		out, err := runTasks(t, dir, "list")
		if err != nil {
			t.Fatalf("list failed: %v\n%s", err, out)
		}
		if !strings.HasPrefix(out, "📋 Task Summary\n\n") {
			t.Errorf("output does not begin with the summary banner:\n%s", out)
		}
		// Each task must surface under its own status heading — the whole point
		// of `list` is the per-status grouping, so a task drifting to the wrong
		// heading is a regression even if the count is right.
		for _, tc := range []struct {
			status, task string
		}{
			{"backlog", "a"}, {"todo", "b"}, {"doing", "c"}, {"done", "d"}, {"blocked", "e"},
		} {
			heading := tc.status + " (1)\n  - " + tc.task + "\n"
			if !strings.Contains(out, heading) {
				t.Errorf("output does not list task %q under %q:\n%s", tc.task, tc.status, out)
			}
		}
	})

	t.Run("reports missing project root", func(t *testing.T) {
		dir := tempProject(t)
		if ancestorHasPlaesy(dir) {
			t.Skip("an ancestor of this temp dir holds a .plaesy directory")
		}
		out, err := runTasks(t, dir, "list")
		if err == nil {
			t.Fatalf("list with no .plaesy must fail\n%s", out)
		}
		if !strings.Contains(err.Error(), ".plaesy") {
			t.Errorf("error %q does not name the .plaesy directory it looked for", err)
		}
		// No banner may have been printed before the failure.
		if out != "" {
			t.Errorf("expected no stdout before the error, got %q", out)
		}
	})

	t.Run("all-empty tree still succeeds with zero counts", func(t *testing.T) {
		dir := makeTasksTree(t, nil)
		out, err := runTasks(t, dir, "list")
		if err != nil {
			t.Fatalf("list of an empty tree failed: %v\n%s", err, out)
		}
		if !strings.HasPrefix(out, "📋 Task Summary\n\n") {
			t.Errorf("missing banner:\n%s", out)
		}
		// Every status directory exists, so each reports a zero count rather
		// than the "[!] No tasks" banner that a missing directory would get.
		for _, s := range taskmanage.Statuses {
			if !strings.Contains(out, s+" (0)") {
				t.Errorf("output does not show %q with 0 tasks:\n%s", s, out)
			}
		}
	})
}

func TestTaskManageListStatus(t *testing.T) {
	t.Run("lists one status", func(t *testing.T) {
		dir := makeTasksTree(t, map[string][]taskFile{
			"todo": {{"a.md", ""}, {"b.md", ""}},
		})
		out, err := runTasks(t, dir, "list-status", "todo")
		if err != nil {
			t.Fatalf("list-status failed: %v\n%s", err, out)
		}
		if out != "todo (2)\n  - a\n  - b\n" {
			t.Errorf("output =\n%q\nwant\n%q", out, "todo (2)\n  - a\n  - b\n")
		}
	})

	t.Run("missing status dir says no tasks", func(t *testing.T) {
		dir := makeTasksTree(t, nil)
		if err := os.RemoveAll(filepath.Join(tasksDirOf(dir), "done")); err != nil {
			t.Fatal(err)
		}
		out, err := runTasks(t, dir, "list-status", "done")
		if err != nil {
			t.Fatalf("list-status done failed: %v\n%s", err, out)
		}
		if out != "[!] No tasks in done\n" {
			t.Errorf("output = %q, want %q", out, "[!] No tasks in done\n")
		}
	})

	t.Run("bad status errors and names the bad status and the valid set", func(t *testing.T) {
		dir := makeTasksTree(t, nil)
		out, err := runTasks(t, dir, "list-status", "archive")
		if err == nil {
			t.Fatalf("expected an error for a bad status\n%s", out)
		}
		if !strings.Contains(err.Error(), "invalid status: archive") {
			t.Errorf("error = %q, want it to name the bad status", err)
		}
		// The valid set is advertised so the user can recover without reading
		// the source — a bare "invalid" is a dead end.
		for _, s := range taskmanage.Statuses {
			if !strings.Contains(err.Error(), s) {
				t.Errorf("error %q does not list valid status %q", err, s)
			}
		}
		if out != "" {
			t.Errorf("expected no stdout on error, got %q", out)
		}
	})

	t.Run("rejects a missing argument", func(t *testing.T) {
		dir := makeTasksTree(t, nil)
		_, err := runTasks(t, dir, "list-status")
		if err == nil {
			t.Fatal("list-status with no args must error")
		}
	})
}

func TestTaskManageNext(t *testing.T) {
	t.Run("pops the highest backlog task into todo", func(t *testing.T) {
		dir := makeTasksTree(t, map[string][]taskFile{
			"backlog": {{"a.md", ""}, {"c.md", ""}, {"b.md", ""}},
		})
		out, err := runTasks(t, dir, "next")
		if err != nil {
			t.Fatalf("next failed: %v\n%s", err, out)
		}
		// GetNextTask is reverse-sorted, so "c.md" is the one promoted — the
		// bash twin uses `ls | sort -r | head -1`.
		if !strings.Contains(out, "[✓] Moved: c.md (backlog → todo)") {
			t.Errorf("output does not report the moved task:\n%s", out)
		}
		if _, err := os.Stat(filepath.Join(tasksDirOf(dir), "backlog", "c.md")); err == nil {
			t.Error("task still in backlog after next")
		}
		if _, err := os.Stat(filepath.Join(tasksDirOf(dir), "todo", "c.md")); err != nil {
			t.Errorf("task not moved to todo: %v", err)
		}
	})

	t.Run("empty backlog is not an error", func(t *testing.T) {
		dir := makeTasksTree(t, nil)
		out, err := runTasks(t, dir, "next")
		if err != nil {
			t.Fatalf("next on an empty backlog failed: %v\n%s", err, out)
		}
		if !strings.Contains(out, "[!] No tasks in backlog") {
			t.Errorf("output does not report an empty backlog:\n%s", out)
		}
	})

	t.Run("missing backlog dir errors and names it", func(t *testing.T) {
		dir := makeTasksTree(t, nil)
		if err := os.RemoveAll(filepath.Join(tasksDirOf(dir), "backlog")); err != nil {
			t.Fatal(err)
		}
		out, err := runTasks(t, dir, "next")
		if err == nil {
			t.Fatalf("next with no backlog dir must error\n%s", out)
		}
		// PopNextTask checks for the backlog directory itself, so a missing
		// tasks/ tree fails here rather than silently printing "no tasks".
		if !strings.Contains(err.Error(), "backlog directory not found") {
			t.Errorf("error = %q, want it to name the backlog directory", err)
		}
	})
}

// CURRENT BEHAVIOUR — suspected bug: newTaskManageNextCmd prints the task file
// name twice on success — once inside the "[✓] Moved: ... (backlog → todo)"
// line and again as a bare filename on the next line. The bare line carries no
// label and is not a documented part of the command contract; it looks like a
// leftover from an earlier draft. This test pins the current output so a fix is
// visible as a deliberate change, not a silent one.
// `next` reported the move and then printed the bare file name again on its own
// line, so a run that moved one task looked like it had done two things. Every
// sibling subcommand prints one labeled line for a move, and this now matches.
func TestTaskManageNextReportsTheMoveOnce(t *testing.T) {
	dir := makeTasksTree(t, map[string][]taskFile{
		"backlog": {{"c.md", ""}},
	})
	out, err := runTasks(t, dir, "next")
	if err != nil {
		t.Fatalf("next failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "[✓] Moved: c.md (backlog → todo)") {
		t.Errorf("output does not report the move:\n%s", out)
	}
	if n := strings.Count(out, "c.md"); n != 1 {
		t.Errorf("the task file appears %d times, want once:\n%s", n, out)
	}
}

func TestTaskManageStart(t *testing.T) {
	t.Run("promotes todo to doing", func(t *testing.T) {
		dir := makeTasksTree(t, map[string][]taskFile{
			"todo": {{"t.md", "# task\n"}},
		})
		out, err := runTasks(t, dir, "start", "t.md")
		if err != nil {
			t.Fatalf("start failed: %v\n%s", err, out)
		}
		if !strings.Contains(out, "[✓] Moved: t.md (todo → doing)") {
			t.Errorf("output does not report the move:\n%s", out)
		}
		if _, err := os.Stat(filepath.Join(tasksDirOf(dir), "todo", "t.md")); err == nil {
			t.Error("task still in todo after start")
		}
		if _, err := os.Stat(filepath.Join(tasksDirOf(dir), "doing", "t.md")); err != nil {
			t.Errorf("task not moved to doing: %v", err)
		}
	})

	t.Run("task not in todo errors and leaves the file", func(t *testing.T) {
		dir := makeTasksTree(t, map[string][]taskFile{
			"backlog": {{"t.md", ""}},
		})
		out, err := runTasks(t, dir, "start", "t.md")
		if err == nil {
			t.Fatalf("starting a backlog task must fail\n%s", out)
		}
		if !strings.Contains(err.Error(), "task not found") {
			t.Errorf("error = %q, want \"task not found\"", err)
		}
		// A failed transition must not move the file: the command layer reports
		// the error but the filesystem must agree with the error.
		if _, err := os.Stat(filepath.Join(tasksDirOf(dir), "backlog", "t.md")); err != nil {
			t.Errorf("failed start still moved the file: %v", err)
		}
	})

	t.Run("rejects a missing argument", func(t *testing.T) {
		dir := makeTasksTree(t, nil)
		_, err := runTasks(t, dir, "start")
		if err == nil {
			t.Fatal("start with no task file must error")
		}
	})
}

func TestTaskManageComplete(t *testing.T) {
	t.Run("promotes doing to done", func(t *testing.T) {
		dir := makeTasksTree(t, map[string][]taskFile{
			"doing": {{"t.md", "# task\n"}},
		})
		out, err := runTasks(t, dir, "complete", "t.md")
		if err != nil {
			t.Fatalf("complete failed: %v\n%s", err, out)
		}
		if !strings.Contains(out, "[✓] Moved: t.md (doing → done)") {
			t.Errorf("output does not report the move:\n%s", out)
		}
		if !strings.Contains(out, "[✓] Task completed: t.md") {
			t.Errorf("output does not report completion:\n%s", out)
		}
		if _, err := os.Stat(filepath.Join(tasksDirOf(dir), "doing", "t.md")); err == nil {
			t.Error("task still in doing after complete")
		}
		if _, err := os.Stat(filepath.Join(tasksDirOf(dir), "done", "t.md")); err != nil {
			t.Errorf("task not moved to done: %v", err)
		}
	})

	t.Run("task not in doing errors", func(t *testing.T) {
		dir := makeTasksTree(t, map[string][]taskFile{
			"todo": {{"t.md", ""}},
		})
		out, err := runTasks(t, dir, "complete", "t.md")
		if err == nil {
			t.Fatalf("completing a todo task must fail\n%s", out)
		}
		if !strings.Contains(err.Error(), "task not found") {
			t.Errorf("error = %q, want \"task not found\"", err)
		}
	})
}

func TestTaskManageBlock(t *testing.T) {
	t.Run("with reason appends a Blocked Reason section", func(t *testing.T) {
		dir := makeTasksTree(t, map[string][]taskFile{
			"doing": {{"t.md", "# task\n\nbody\n"}},
		})
		out, err := runTasks(t, dir, "block", "t.md", "waiting on API key")
		if err != nil {
			t.Fatalf("block failed: %v\n%s", err, out)
		}
		if !strings.Contains(out, "[✓] Moved: t.md (doing → blocked)") {
			t.Errorf("output does not report the move:\n%s", out)
		}
		got, err := os.ReadFile(filepath.Join(tasksDirOf(dir), "blocked", "t.md"))
		if err != nil {
			t.Fatalf("read blocked task: %v", err)
		}
		// The reason is appended verbatim as a "## Blocked Reason" section,
		// matching the bash twin's `cat >>` append rather than a rewrite.
		want := "# task\n\nbody\n\n## Blocked Reason\nwaiting on API key\n"
		if string(got) != want {
			t.Errorf("blocked file content = %q, want %q", got, want)
		}
	})

	t.Run("without a reason uses the default", func(t *testing.T) {
		dir := makeTasksTree(t, map[string][]taskFile{
			"doing": {{"t.md", "# task\n"}},
		})
		out, err := runTasks(t, dir, "block", "t.md")
		if err != nil {
			t.Fatalf("block failed: %v\n%s", err, out)
		}
		got, err := os.ReadFile(filepath.Join(tasksDirOf(dir), "blocked", "t.md"))
		if err != nil {
			t.Fatalf("read blocked task: %v", err)
		}
		if !strings.Contains(string(got), "## Blocked Reason\nNo reason provided\n") {
			t.Errorf("expected the default reason, got %q", got)
		}
	})

	t.Run("task not in doing errors", func(t *testing.T) {
		dir := makeTasksTree(t, map[string][]taskFile{
			"todo": {{"t.md", ""}},
		})
		out, err := runTasks(t, dir, "block", "t.md", "r")
		if err == nil {
			t.Fatalf("blocking a todo task must fail\n%s", out)
		}
		if !strings.Contains(err.Error(), "task not found") {
			t.Errorf("error = %q, want \"task not found\"", err)
		}
	})

	t.Run("rejects too many arguments", func(t *testing.T) {
		dir := makeTasksTree(t, nil)
		_, err := runTasks(t, dir, "block", "t.md", "r", "extra")
		if err == nil {
			t.Fatal("block with three args must error (RangeArgs(1,2))")
		}
	})
}

func TestTaskManageUnblock(t *testing.T) {
	t.Run("returns blocked to todo", func(t *testing.T) {
		dir := makeTasksTree(t, map[string][]taskFile{
			"blocked": {{"t.md", "# task\n"}},
		})
		out, err := runTasks(t, dir, "unblock", "t.md")
		if err != nil {
			t.Fatalf("unblock failed: %v\n%s", err, out)
		}
		if !strings.Contains(out, "[✓] Moved: t.md (blocked → todo)") {
			t.Errorf("output does not report the move:\n%s", out)
		}
		if !strings.Contains(out, "[✓] Task unblocked: t.md") {
			t.Errorf("output does not report unblocking:\n%s", out)
		}
		if _, err := os.Stat(filepath.Join(tasksDirOf(dir), "blocked", "t.md")); err == nil {
			t.Error("task still in blocked after unblock")
		}
		if _, err := os.Stat(filepath.Join(tasksDirOf(dir), "todo", "t.md")); err != nil {
			t.Errorf("task not moved to todo: %v", err)
		}
	})

	t.Run("task not in blocked errors", func(t *testing.T) {
		dir := makeTasksTree(t, map[string][]taskFile{
			"doing": {{"t.md", ""}},
		})
		out, err := runTasks(t, dir, "unblock", "t.md")
		if err == nil {
			t.Fatalf("unblocking a doing task must fail\n%s", out)
		}
		if !strings.Contains(err.Error(), "task not found") {
			t.Errorf("error = %q, want \"task not found\"", err)
		}
	})
}

func TestTaskManageMove(t *testing.T) {
	t.Run("moves between arbitrary statuses", func(t *testing.T) {
		dir := makeTasksTree(t, map[string][]taskFile{
			"backlog": {{"t.md", "# task\n"}},
		})
		out, err := runTasks(t, dir, "move", "t.md", "backlog", "doing")
		if err != nil {
			t.Fatalf("move failed: %v\n%s", err, out)
		}
		if !strings.Contains(out, "[✓] Moved: t.md (backlog → doing)") {
			t.Errorf("output does not report the move:\n%s", out)
		}
		if _, err := os.Stat(filepath.Join(tasksDirOf(dir), "backlog", "t.md")); err == nil {
			t.Error("task still in backlog after move")
		}
		if _, err := os.Stat(filepath.Join(tasksDirOf(dir), "doing", "t.md")); err != nil {
			t.Errorf("task not moved to doing: %v", err)
		}
	})

	t.Run("invalid from status errors", func(t *testing.T) {
		dir := makeTasksTree(t, map[string][]taskFile{
			"backlog": {{"t.md", ""}},
		})
		_, err := runTasks(t, dir, "move", "t.md", "archive", "todo")
		if err == nil || !strings.Contains(err.Error(), "invalid status: archive") {
			t.Errorf("error = %v, want \"invalid status: archive\"", err)
		}
	})

	t.Run("invalid to status errors", func(t *testing.T) {
		dir := makeTasksTree(t, map[string][]taskFile{
			"backlog": {{"t.md", ""}},
		})
		_, err := runTasks(t, dir, "move", "t.md", "backlog", "wip")
		if err == nil || !strings.Contains(err.Error(), "invalid status: wip") {
			t.Errorf("error = %v, want \"invalid status: wip\"", err)
		}
	})

	t.Run("missing source file errors", func(t *testing.T) {
		dir := makeTasksTree(t, nil)
		out, err := runTasks(t, dir, "move", "missing.md", "backlog", "todo")
		if err == nil {
			t.Fatalf("move of a missing task must fail\n%s", out)
		}
		if !strings.Contains(err.Error(), "task not found") {
			t.Errorf("error = %q, want \"task not found\"", err)
		}
	})

	// Guards against os.Rename trying to rename over a directory, which would be
	// a different class of failure than "task not found". A directory planted at
	// the task path must be reported as not-found, never silently clobbered.
	t.Run("directory where a task file is expected errors", func(t *testing.T) {
		dir := makeTasksTree(t, nil)
		if err := os.MkdirAll(filepath.Join(tasksDirOf(dir), "backlog", "t.md"), 0o755); err != nil {
			t.Fatal(err)
		}
		out, err := runTasks(t, dir, "move", "t.md", "backlog", "todo")
		if err == nil {
			t.Fatalf("move of a directory must fail\n%s", out)
		}
		if !strings.Contains(err.Error(), "task not found") {
			t.Errorf("error = %q, want \"task not found\"", err)
		}
	})

	t.Run("rejects the wrong argument count", func(t *testing.T) {
		dir := makeTasksTree(t, nil)
		_, err := runTasks(t, dir, "move", "t.md", "backlog")
		if err == nil {
			t.Fatal("move with two args must error (ExactArgs(3))")
		}
	})
}

func TestTaskManageShow(t *testing.T) {
	const body = "# task\n\nsome content\n"

	t.Run("shows the task in the named status", func(t *testing.T) {
		dir := makeTasksTree(t, map[string][]taskFile{
			"doing": {{"t.md", body}},
		})
		out, err := runTasks(t, dir, "show", "t.md", "doing")
		if err != nil {
			t.Fatalf("show failed: %v\n%s", err, out)
		}
		// The header is a fixed two-line preamble followed by a blank line,
		// then the raw file content verbatim — no trailing newline is added.
		if !strings.HasPrefix(out, "📄 Task: t.md\nStatus: doing\n\n") {
			t.Errorf("output does not begin with the task header + blank line:\n%s", out)
		}
		if !strings.HasSuffix(out, body) {
			t.Errorf("output does not end with the raw task content:\n%s", out)
		}
	})

	t.Run("searches every status when status is omitted", func(t *testing.T) {
		dir := makeTasksTree(t, map[string][]taskFile{
			"blocked": {{"t.md", body}},
		})
		out, err := runTasks(t, dir, "show", "t.md")
		if err != nil {
			t.Fatalf("show failed: %v\n%s", err, out)
		}
		if !strings.Contains(out, "Status: blocked") {
			t.Errorf("show without status did not resolve to blocked:\n%s", out)
		}
	})

	t.Run("missing task errors", func(t *testing.T) {
		dir := makeTasksTree(t, nil)
		out, err := runTasks(t, dir, "show", "nope.md")
		if err == nil {
			t.Fatalf("show of a missing task must fail\n%s", out)
		}
		if !strings.Contains(err.Error(), "task not found: nope.md") {
			t.Errorf("error = %q, want \"task not found: nope.md\"", err)
		}
	})

	t.Run("invalid status errors", func(t *testing.T) {
		dir := makeTasksTree(t, nil)
		out, err := runTasks(t, dir, "show", "t.md", "archive")
		if err == nil {
			t.Fatalf("show with an invalid status must fail\n%s", out)
		}
		if !strings.Contains(err.Error(), "invalid status: archive") {
			t.Errorf("error = %q, want \"invalid status: archive\"", err)
		}
	})

	t.Run("explicit status that does not hold the task errors", func(t *testing.T) {
		dir := makeTasksTree(t, map[string][]taskFile{
			"todo": {{"t.md", body}},
		})
		out, err := runTasks(t, dir, "show", "t.md", "done")
		if err == nil {
			t.Fatalf("show done for a todo task must fail\n%s", out)
		}
		if !strings.Contains(err.Error(), "task not found: t.md") {
			t.Errorf("error = %q, want \"task not found: t.md\"", err)
		}
	})

	// A directory planted where a task file is expected must read as
	// "task not found", not as a successful (and empty) read of a directory.
	t.Run("task that is a directory errors", func(t *testing.T) {
		dir := makeTasksTree(t, nil)
		if err := os.MkdirAll(filepath.Join(tasksDirOf(dir), "todo", "t.md"), 0o755); err != nil {
			t.Fatal(err)
		}
		out, err := runTasks(t, dir, "show", "t.md", "todo")
		if err == nil {
			t.Fatalf("show of a directory must fail\n%s", out)
		}
		if !strings.Contains(err.Error(), "task not found") {
			t.Errorf("error = %q, want \"task not found\"", err)
		}
	})
}

// TestTaskManageSubcommandsReportMissingProjectRoot pins that every subcommand,
// before it does any work, bails out with the tasksDirOrErr error when there is
// no .plaesy directory above the cwd. Without this guard a command could print
// a success banner from a partial run and exit 0 — the worst kind of silent
// success. The dummy task-file arguments never reach the underlying package,
// because the root check runs first.
func TestTaskManageSubcommandsReportMissingProjectRoot(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"list", []string{"list"}},
		{"list-status", []string{"list-status", "todo"}},
		{"next", []string{"next"}},
		{"start", []string{"start", "t.md"}},
		{"complete", []string{"complete", "t.md"}},
		{"block", []string{"block", "t.md"}},
		{"unblock", []string{"unblock", "t.md"}},
		{"move", []string{"move", "t.md", "todo", "doing"}},
		{"show", []string{"show", "t.md"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := tempProject(t) // no .plaesy created
			if ancestorHasPlaesy(dir) {
				t.Skip("an ancestor of this temp dir holds a .plaesy directory")
			}
			out, err := runTasks(t, dir, tc.args...)
			if err == nil {
				t.Fatalf("%s with no project root must error\n%s", tc.name, out)
			}
			if !strings.Contains(err.Error(), ".plaesy") {
				t.Errorf("%s error %q does not name the .plaesy directory it looked for", tc.name, err)
			}
			// No banner may have been printed before the check failed.
			if out != "" {
				t.Errorf("%s printed stdout before failing: %q", tc.name, out)
			}
		})
	}
}

// TestTaskManageLifecycleEndToEnd walks a single task through the full
// backlog -> todo -> doing -> done pipeline, pinning that each command hands
// off the file correctly and that `show` can read it at the end. A break in any
// link returns the task to the wrong status or loses it, and this is the only
// assertion that covers the handoff as a whole rather than per-command.
func TestTaskManageLifecycleEndToEnd(t *testing.T) {
	dir := makeTasksTree(t, map[string][]taskFile{
		"backlog": {{"t.md", "# task\n\nbody\n"}},
	})
	tasksDir := tasksDirOf(dir)

	out, err := runTasks(t, dir, "next")
	if err != nil {
		t.Fatalf("next: %v\n%s", err, out)
	}

	out, err = runTasks(t, dir, "start", "t.md")
	if err != nil {
		t.Fatalf("start: %v\n%s", err, out)
	}
	if _, err := os.Stat(filepath.Join(tasksDir, "doing", "t.md")); err != nil {
		t.Errorf("task not in doing after start: %v", err)
	}

	out, err = runTasks(t, dir, "complete", "t.md")
	if err != nil {
		t.Fatalf("complete: %v\n%s", err, out)
	}
	if _, err := os.Stat(filepath.Join(tasksDir, "done", "t.md")); err != nil {
		t.Errorf("task not in done after complete: %v", err)
	}

	out, err = runTasks(t, dir, "show", "t.md", "done")
	if err != nil {
		t.Fatalf("show done: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Status: done") {
		t.Errorf("show does not report the done status after the lifecycle:\n%s", out)
	}
	if !strings.Contains(out, "body") {
		t.Errorf("show does not contain the task body after the lifecycle:\n%s", out)
	}
}
