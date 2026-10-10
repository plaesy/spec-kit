package main

// Tests for the `plaesy clean` command surface: the flag wiring between cobra
// and internal/cleaner, and the platform name the user types versus the id the
// config declares. The package itself is covered at 95.5% (cleaner_test.go),
// so what is untested here is exactly the glue — and the glue is where both
// defects this file pins lived.

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/plaesy/spec-kit/internal/config"
)

// cleanFixture builds a project that looks like one a user has actually run
// `plaesy init --ai <platform>` in, then returns its path. Only the named
// platforms' files are created, so auto-detection sees a realistic project
// rather than one that has every AI tool installed at once.
func cleanFixture(t *testing.T, platforms ...string) string {
	t.Helper()
	if len(platforms) == 0 {
		platforms = []string{"claude"}
	}
	dir := tempProject(t)
	files := map[string]string{
		".plaesy/context.md":        "# Context\n",
		".plaesy/tasks/todo/a.md":   "task\n",
		"specs/001-feature/spec.md": "spec\n",
		"src/main.go":               "package main\n",
	}
	// The mapped targets come from the config rather than from a hard-coded
	// list, so the fixture still means something if a mapping changes.
	for target, content := range mappedTargets(t, platforms) {
		files[filepath.ToSlash(target)] = content
	}
	// Some mappings point a directory at the same path another mapping nests a
	// file inside (cursor_ai maps prompts -> .cursor/rules and core ->
	// .cursor/rules/plaesy.mdc), so the directories are created first. Writing
	// the file first would leave a file where a directory has to go.
	paths := make([]string, 0, len(files))
	for rel := range files {
		paths = append(paths, filepath.Join(dir, filepath.FromSlash(rel)))
	}
	for _, path := range paths {
		parent := filepath.Dir(path)
		if err := os.MkdirAll(parent, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", parent, err)
		}
	}
	for rel, content := range files {
		path := filepath.Join(dir, filepath.FromSlash(rel))
		info, err := os.Stat(path)
		if err == nil && info.IsDir() {
			continue
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
	return dir
}

// mappedTargets reads the shipped platform.json and returns the
// core/instructions/prompts/agents paths the named platforms map to. Reading the
// config is what keeps the fixture honest: a name that resolves to the wrong
// platform resolves to paths this fixture did not create, so the test fails
// rather than quietly passing on an empty plan.
func mappedTargets(t *testing.T, platforms []string) map[string]string {
	t.Helper()
	cfg, err := config.Load(repoConfigPath())
	if err != nil {
		t.Fatalf("load platform.json: %v", err)
	}
	out := map[string]string{}
	for _, id := range platforms {
		p, ok := cfg.Platforms[id]
		if !ok {
			t.Fatalf("platform.json does not declare %q", id)
		}
		for kind, target := range p.Mapping {
			switch kind {
			case "core", "instructions", "prompts", "agents":
				if target != "" && target != "null" {
					out[target] = "plaesy " + id + " " + kind + "\n"
				}
			}
		}
	}
	return out
}

func repoConfigPath() string { return filepath.Join("..", "..", "configs", "platform.json") }

// tempProject returns a fresh directory under a deliberately long path.
// t.TempDir() is short on this machine (TMP is C:\msys64	mp), and a fixture
// that lays down .claude/commands/ or .cursor/rules/ under a long test name
// overflows the legacy MAX_PATH limit. Nesting one level fixes it.
func tempProject(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "plaesy-test", "project")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

// runClean executes `plaesy clean` in-process and returns everything it printed.
// cobra's own output is captured separately: the command writes its plan with
// fmt.Println, so only stdout-as-os-stdout is a real test seam here.
func runClean(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	full := append([]string{dir, "--yes", "--config", repoConfigPath()}, args...)
	return captureStdout(t, func() error {
		cmd := newCleanCmd()
		cmd.SetArgs(full)
		return cmd.Execute()
	})
}

// captureStdout redirects os.Stdout for the duration of fn. Every command in
// this package prints with fmt.Println rather than through cmd.OutOrStdout, so
// asserting on behaviour means asserting on the process's real stdout.
func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		done <- buf.String()
	}()

	fnErr := fn()

	os.Stdout = orig
	w.Close()
	out := <-done
	r.Close()
	return out, fnErr
}

// `--backup` defaults to true, so the flag's own value cannot decide it: only
// an explicit opt-out should. The old `if noBackup ... else if backup` chain read
// `--backup=false` as "not mentioned", fell through to the default, and took the
// backup anyway — the opposite of what the command line said, and the backup
// copies the user's specs.
func TestCleanBackupFlagCanActuallyBeTurnedOff(t *testing.T) {
	cases := []struct {
		name      string
		args      []string
		wantSaved bool
	}{
		{"default", nil, true},
		{"--backup is the default and says so", []string{"--backup"}, true},
		{"--no-backup", []string{"--no-backup"}, false},
		{"--backup=false", []string{"--backup=false"}, false},
		{"--no-backup wins over --backup", []string{"--no-backup", "--backup"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := cleanFixture(t)
			out, err := runClean(t, dir, append([]string{"--level", "thorough"}, tc.args...)...)
			if err != nil {
				t.Fatalf("clean failed: %v\n%s", err, out)
			}
			backups, err := filepath.Glob(filepath.Join(dir, ".plaesy-backup-*"))
			if err != nil {
				t.Fatal(err)
			}
			gotBackup := len(backups) > 0
			if gotBackup != tc.wantSaved {
				t.Errorf("backup present = %v, want %v (args %v)", gotBackup, tc.wantSaved, tc.args)
			}
			// The banner reports the effective setting, so the two can be
			// compared; a mismatch would mean the flag and the plan disagree.
			if tc.wantSaved && !strings.Contains(out, "backup") {
				t.Errorf("output does not mention the backup:\n%s", out)
			}
		})
	}
}

// The alias the CLI advertises in its own error message has to reach a config
// key. "--ai claude" was accepted, echoed back as the chosen platform, and then
// matched no mapping, so the plan found nothing and the run removed nothing.
func TestCleanResolvesTheShorthandPlatformNames(t *testing.T) {
	cases := []struct{ shorthand, canonical string }{
		{"claude", "claude"},
		{"claude_code", "claude"},
		{"copilot", "github_copilot"},
		{"github_copilot", "github_copilot"},
		{"cursor", "cursor_ai"},
		{"windsurf", "windsurf_ai"},
	}
	for _, tc := range cases {
		t.Run(tc.shorthand, func(t *testing.T) {
			// The fixture holds the canonical platform's files, so the plan is
			// only non-empty if the shorthand reached that same platform.
			targets := mappedTargets(t, []string{tc.canonical})
			dir := cleanFixture(t, tc.canonical)
			out, err := runClean(t, dir, "--ai", tc.shorthand, "--dry-run")
			if err != nil {
				t.Fatalf("clean --ai %s failed: %v\n%s", tc.shorthand, err, out)
			}
			for target := range targets {
				leaf := filepath.ToSlash(target)
				if !strings.Contains(out, leaf) {
					t.Errorf("--ai %s resolved to the wrong platform: the plan never lists %s\n%s",
						tc.shorthand, leaf, out)
				}
			}
			// Dry run, so nothing is removed; the point is what the plan lists.
			if _, err := os.Stat(filepath.Join(dir, "src", "main.go")); err != nil {
				t.Errorf("--dry-run removed a file: %v", err)
			}
		})
	}
}

func TestCleanRejectsAnUnknownPlatform(t *testing.T) {
	dir := cleanFixture(t)
	out, err := runClean(t, dir, "--ai", "mystery")
	if err == nil {
		t.Fatalf("an unknown platform must be rejected\n%s", out)
	}
	if !strings.Contains(err.Error(), "mystery") {
		t.Errorf("error %q does not name the platform the user typed", err)
	}
	// The message has to list what is available, or the user cannot recover.
	if !strings.Contains(err.Error(), "claude") {
		t.Errorf("error %q does not list the configured platforms", err)
	}
	// A rejected name must not have removed anything.
	if _, statErr := os.Stat(filepath.Join(dir, ".plaesy")); statErr != nil {
		t.Error("a rejected run still cleaned the project")
	}
}

func TestCleanRejectsAnUnknownLevel(t *testing.T) {
	dir := cleanFixture(t)
	out, err := runClean(t, dir, "--level", "aggressive")
	if err == nil || !strings.Contains(err.Error(), "invalid cleanup level") {
		t.Fatalf("err = %v, want an invalid-level error\n%s", err, out)
	}
	for _, level := range []string{"safe", "thorough", "complete"} {
		if !strings.Contains(err.Error(), level) {
			t.Errorf("error %q does not list %q", err, level)
		}
	}
	if _, statErr := os.Stat(filepath.Join(dir, ".plaesy")); statErr != nil {
		t.Error("a rejected level still cleaned the project")
	}
}

func TestCleanRejectsAMissingTarget(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nope")
	cmd := newCleanCmd()
	cmd.SetArgs([]string{missing, "--yes", "--config", repoConfigPath()})
	out, err := captureStdout(t, cmd.Execute)
	if err == nil {
		t.Fatalf("a missing target directory must be an error\n%s", out)
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("err = %v, want a missing-target error", err)
	}
}

// The level is the knob Phase 11 found was only a banner; this pins it at the
// command surface too, because that is where a user reaches it.
func TestCleanLevelDecidesWhetherSpecsSurvive(t *testing.T) {
	cases := []struct {
		level      string
		wantSpecs  bool
		wantBackup bool
	}{
		{"safe", true, true},
		{"thorough", false, true},
		{"complete", false, true},
	}
	for _, tc := range cases {
		t.Run(tc.level, func(t *testing.T) {
			dir := cleanFixture(t)
			out, err := runClean(t, dir, "--level", tc.level)
			if err != nil {
				t.Fatalf("clean --level %s failed: %v\n%s", tc.level, err, out)
			}
			_, statErr := os.Stat(filepath.Join(dir, "specs"))
			gotSpecs := statErr == nil
			if gotSpecs != tc.wantSpecs {
				t.Errorf("specs present after --level %s = %v, want %v", tc.level, gotSpecs, tc.wantSpecs)
			}
			// User code is never a cleanup target at any level.
			if _, err := os.Stat(filepath.Join(dir, "src", "main.go")); err != nil {
				t.Errorf("--level %s removed source code: %v", tc.level, err)
			}
		})
	}
}

func TestCleanDryRunChangesNothing(t *testing.T) {
	dir := cleanFixture(t)
	out, err := runClean(t, dir, "--dry-run")
	if err != nil {
		t.Fatalf("dry run failed: %v\n%s", err, out)
	}
	for _, rel := range []string{".plaesy", "specs", "CLAUDE.md", ".claude/commands"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel))); err != nil {
			t.Errorf("--dry-run removed %s: %v", rel, err)
		}
	}
	backups, _ := filepath.Glob(filepath.Join(dir, ".plaesy-backup-*"))
	if len(backups) > 0 {
		t.Error("--dry-run created a backup, which is a copy of everything it would have removed")
	}
	// It must still say what it would do, or the flag buys nothing.
	if !strings.Contains(out, "specs") {
		t.Errorf("the dry-run plan does not list what it would remove:\n%s", out)
	}
}

// With nothing to remove the run is a success, not a failure: "there is no Plaesy
// here" is an answer to a question, and the exit code is what a CI step reads.
func TestCleanOnAPlainDirectorySucceedsAndSaysSo(t *testing.T) {
	dir := tempProject(t)
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := runClean(t, dir)
	if err != nil {
		t.Fatalf("cleaning a directory with nothing to clean failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "No Plaesy") {
		t.Errorf("output does not say there was nothing to do:\n%s", out)
	}
}

// A detection that finds nothing must not fall back to removing the generic
// fallback platform's files at the default level; that is `complete`'s job.
func TestCleanOnAnUnconfiguredProjectKeepsTheGenericFallback(t *testing.T) {
	dir := tempProject(t)
	for rel, content := range map[string]string{
		".plaesy/context.md":    "# Context\n",
		"AI-INSTRUCTIONS.md":    "fallback\n",
		"ai-config/config.yaml": "k: v\n",
	} {
		path := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	out, err := runClean(t, dir)
	if err != nil {
		t.Fatalf("clean failed: %v\n%s", err, out)
	}
	for _, rel := range []string{"AI-INSTRUCTIONS.md", filepath.Join("ai-config", "config.yaml")} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel))); err != nil {
			t.Errorf("the default level removed the generic fallback file %s: %v", rel, err)
		}
	}

	// complete is the level that takes them, and it is the level that says so.
	out, err = runClean(t, dir, "--level", "complete")
	if err != nil {
		t.Fatalf("clean --level complete failed: %v\n%s", err, out)
	}
	if _, err := os.Stat(filepath.Join(dir, "AI-INSTRUCTIONS.md")); err == nil {
		t.Error("--level complete kept the generic fallback file it exists to remove")
	}
}

func TestCleanVerbosePrintsTheEffectiveSettings(t *testing.T) {
	dir := cleanFixture(t)
	out, err := runClean(t, dir, "--verbose", "--level", "thorough", "--no-backup", "--dry-run")
	if err != nil {
		t.Fatalf("clean failed: %v\n%s", err, out)
	}
	for _, want := range []string{"Cleanup Level: thorough", "Dry Run: true", "Backup: false", "Target Directory:"} {
		if !strings.Contains(out, want) {
			t.Errorf("--verbose output is missing %q:\n%s", want, out)
		}
	}
	// The settings it prints are the ones the run used, which is the only
	// reason to print them: --no-backup must not read back as true.
	if strings.Contains(out, "Backup: true") {
		t.Errorf("--verbose reports a backup that will not be taken:\n%s", out)
	}
}
