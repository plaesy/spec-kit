package main

// Tests for `plaesy trim`, the last large uncovered block in this package.
//
// The subcommands here are thin, but the thinness is the risk: each one resolves
// its own paths and validates its own flags, and a thin wrapper that validates
// nothing still returns nil, which reads as "the trim ran" when nothing was
// trimmed. The argument handling is therefore the behaviour under test — in
// particular `trim run`'s DisableFlagParsing, which moves the `--` handling out
// of cobra and into the command body where nothing else checks it.

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// inTrimRepo points the process at a throwaway repo root. resolveTrimPaths
// reads the cwd, so every trim subcommand resolves its inputs from here rather
// than from the real repository — a test must never compact or rewrite the
// repo it is running in.
func inTrimRepo(t *testing.T) string {
	t.Helper()
	// Nested one level: %TEMP% here is a short path (C:\msys64\tmp), so a deep
	// fixture under a long test name can overflow the legacy MAX_PATH limit.
	dir := filepath.Join(t.TempDir(), "plaesy-test", "project")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })
	// Return what os.Getwd() now reports, not the pre-Chdir `dir` string:
	// resolveTrimPaths derives repoRoot from os.Getwd(), and on macOS a
	// t.TempDir() under /var/folders/... is reached through /var, a symlink
	// to /private/var — Getwd returns the kernel's resolved, symlink-free
	// view ("/private/var/folders/...") where `dir` does not. Callers that
	// compare against this return value are comparing against the same
	// thing resolveTrimPaths actually sees.
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return wd
}

// ---- resolveTrimPaths ----------------------------------------------------

// The three derived paths are the whole contract: they must hang off the
// directory the command was run in, and the rules must always be usable. A nil
// rules value here would mean every later call silently used defaults, so the
// non-nil assertion is the load-bearing one.
func TestResolveTrimPathsDerivesEverythingFromTheWorkingDirectory(t *testing.T) {
	dir := inTrimRepo(t)

	paths := resolveTrimPaths()

	if paths.repoRoot != dir {
		t.Errorf("repoRoot = %q, want the working directory %q", paths.repoRoot, dir)
	}
	if paths.rules == nil {
		t.Fatal("rules must never be nil: every caller dereferences them without a nil check")
	}
	wantGraph := filepath.Join(dir, ".plaesy", "analysis", "project.graph.json")
	if paths.graph != wantGraph {
		t.Errorf("graph = %q, want %q", paths.graph, wantGraph)
	}
	wantStats := filepath.Join(dir, ".plaesy", "memory", "token-stats.json")
	if paths.stats != wantStats {
		t.Errorf("stats = %q, want %q", paths.stats, wantStats)
	}
}

// A repo with no rules file must still get usable rules from the built-in
// default. A compiled binary can be installed anywhere, so "the file is missing"
// is the normal case, not an error — returning nil here would crash every
// caller instead of degrading.
func TestResolveTrimPathsFallsBackToDefaultRules(t *testing.T) {
	dir := inTrimRepo(t)
	if _, err := os.Stat(filepath.Join(dir, "scripts", "configs", "plaesy-trim-rules.json")); err == nil {
		t.Skip("fixture unexpectedly contains a rules file")
	}

	if rules := resolveTrimPaths().rules; rules == nil {
		t.Fatal("missing rules file must fall back to defaults, not nil")
	}
}

// A project-level rules file must win over the framework default: that is the
// documented way to tune trimming for one project without editing the install.
func TestResolveTrimPathsPrefersTheProjectRulesFile(t *testing.T) {
	dir := inTrimRepo(t)
	rulesPath := filepath.Join(dir, ".plaesy", "scripts", "configs", "plaesy-trim-rules.json")
	if err := os.MkdirAll(filepath.Dir(rulesPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(rulesPath, []byte(`{"layer1_commands":{"go test":{"max_lines":7,"dedupe_consecutive":true}}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	rules := resolveTrimPaths().rules
	if rules == nil {
		t.Fatal("rules must not be nil")
	}
	// The project file named exactly one command, so that key must be the one
	// that loaded. A rules file that silently lost its own entry would leave
	// the default set in place and trim everything the built-in way.
	if _, ok := rules.Layer1Commands["go test"]; !ok {
		t.Errorf("the project rules file must be the one loaded, got %v", rules.Layer1Commands)
	}
}

// A malformed rules file must also degrade to defaults rather than propagating a
// parse error: trimming is an optimization, and a broken config should cost
// quality, not the run.
func TestResolveTrimPathsSurvivesAMalformedRulesFile(t *testing.T) {
	dir := inTrimRepo(t)
	rulesPath := filepath.Join(dir, ".plaesy", "scripts", "configs", "plaesy-trim-rules.json")
	if err := os.MkdirAll(filepath.Dir(rulesPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(rulesPath, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if rules := resolveTrimPaths().rules; rules == nil {
		t.Fatal("a malformed rules file must fall back to defaults, not nil")
	}
}

// runLeafCmdNoArgs executes a command with genuinely no arguments.
// runLeafCmd(t, cmd) passes a nil slice to SetArgs, which cobra reads as
// "arguments were never set" and falls back to os.Args[1:] — the test binary's
// own flags. The empty-but-non-nil slice is the only way to say "no arguments".
func runLeafCmdNoArgs(t *testing.T, cmd *cobra.Command) error {
	t.Helper()
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	return cmd.Execute()
}

// ---- trim run ------------------------------------------------------------

// `trim run` is how you compress a command whose own flags start with "-",
// which is why DisableFlagParsing is on and the "--" is stripped by hand.
// Cobra only strips that separator when it is doing the parsing, so the body
// owns it; if that line were dropped, the trimmer would try to execute "--" as
// the command.
func TestTrimRunStripsTheDoubleDashSeparator(t *testing.T) {
	inTrimRepo(t)

	// "--" alone leaves nothing to run, which is only detectable if the
	// separator was stripped: with it left in, the command would be "--".
	_, err := captureStdout(t, func() error { return runLeafCmd(t, newTrimRunCmd(), "--") })
	if err == nil {
		t.Fatal("`trim run --` with no command must be a usage error")
	}
	if !strings.Contains(err.Error(), "usage: plaesy trim run") {
		t.Errorf("the error should show the usage line, got %v", err)
	}
}

// No command at all is the same mistake written differently, and must not be
// treated as "trim nothing successfully".
func TestTrimRunRequiresACommand(t *testing.T) {
	inTrimRepo(t)
	if err := runLeafCmdNoArgs(t, newTrimRunCmd()); err == nil {
		t.Fatal("`trim run` with no arguments must be a usage error")
	}
}

// A command that does not exist must surface the failure. trim run wraps a real
// subprocess, so a silent nil here would report a successful trim of a command
// that never ran.
func TestTrimRunReportsAFailingCommand(t *testing.T) {
	inTrimRepo(t)
	out, err := captureStdout(t, func() error {
		return runLeafCmd(t, newTrimRunCmd(), "plaesy-no-such-binary-exists")
	})
	if err == nil {
		t.Fatalf("running a nonexistent command must fail, not report a clean trim:\n%s", out)
	}
}

// ---- trim compress -------------------------------------------------------

// Every trim subcommand that needs a path must say so. The check is the whole
// body of a thin wrapper, and without it the command would resolve paths and
// compact the repository root.
func TestTrimSubcommandsRequireTheirPathFlag(t *testing.T) {
	inTrimRepo(t)
	for _, tc := range []struct {
		name string
		cmd  *cobra.Command
		want string
	}{
		{"compress", newTrimCompressCmd(), "usage: plaesy trim compress"},
		{"llm-queue", newTrimLLMQueueCmd(), "usage: plaesy trim llm-queue"},
	} {
		_, err := captureStdout(t, func() error { return runLeafCmd(t, tc.cmd) })
		if err == nil {
			t.Errorf("%s without --path must be a usage error", tc.name)
			continue
		}
		if !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s error = %v, want it to contain %q", tc.name, err, tc.want)
		}
	}
}

// apply-llm needs two inputs; either one alone must be rejected, because
// applying annotations without a target (or a target without annotations) is
// never a meaningful run.
func TestTrimApplyLLMRequiresBothPathAndAnnotations(t *testing.T) {
	inTrimRepo(t)
	cases := [][]string{
		{},
		{"--path", "doc.md"},
		{"--annotations", "ann.json"},
	}
	for _, args := range cases {
		_, err := captureStdout(t, func() error { return runLeafCmd(t, newTrimApplyLLMCmd(), args...) })
		if err == nil {
			t.Errorf("args %v must be rejected", args)
			continue
		}
		if !strings.Contains(err.Error(), "usage: plaesy trim apply-llm") {
			t.Errorf("args %v: error = %v, want the usage line", args, err)
		}
	}
}

// compress with a path that does not exist must fail. The alternative — an empty
// result list and nil — would report a successful compression of nothing.
func TestTrimCompressFailsOnAMissingPath(t *testing.T) {
	dir := inTrimRepo(t)
	missing := filepath.Join(dir, "no-such-file.md")
	if _, err := captureStdout(t, func() error { return runLeafCmd(t, newTrimCompressCmd(), "--path", missing) }); err == nil {
		t.Fatal("compressing a missing file must fail")
	}
}

// llm-queue must fail on a missing file for the same reason.
func TestTrimLLMQueueFailsOnAMissingFile(t *testing.T) {
	dir := inTrimRepo(t)
	missing := filepath.Join(dir, "no-such-file.md")
	if _, err := captureStdout(t, func() error { return runLeafCmd(t, newTrimLLMQueueCmd(), "--path", missing) }); err == nil {
		t.Fatal("queueing a missing file must fail")
	}
}

// ---- trim report ---------------------------------------------------------

// Report is the cumulative-savings view, and the empty state is the first thing
// every user sees. It must be a clear next-step message and a success, not an
// error about a missing file — a fresh project has no stats and that is normal.
func TestTrimReportExplainsTheEmptyState(t *testing.T) {
	inTrimRepo(t)

	out, err := captureStdout(t, func() error { return runLeafCmd(t, newTrimReportCmd()) })
	if err != nil {
		t.Fatalf("report with no stats must succeed, got %v", err)
	}
	if !strings.Contains(out, "No token stats yet") {
		t.Errorf("expected an empty-state message, got:\n%s", out)
	}
	if !strings.Contains(out, "compress") {
		t.Errorf("the empty state should name the command that creates stats:\n%s", out)
	}
}

// An empty stats file is corrupt-but-present. It must read as "no stats yet"
// rather than an unmarshal error, since the outcome the user cares about is
// identical.
func TestTrimReportTreatsAnEmptyStatsFileAsNoStats(t *testing.T) {
	dir := inTrimRepo(t)
	stats := filepath.Join(dir, ".plaesy", "memory", "token-stats.json")
	if err := os.MkdirAll(filepath.Dir(stats), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(stats, []byte("[]"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := captureStdout(t, func() error { return runLeafCmd(t, newTrimReportCmd()) })
	if err != nil {
		t.Fatalf("an empty stats array must not be an error, got %v", err)
	}
	if !strings.Contains(out, "No token stats yet") {
		t.Errorf("expected the empty-state message, got:\n%s", out)
	}
}

// ---- the trim command group ---------------------------------------------

// The parent exists to route, so a typo at this level has to name the layers
// rather than fail as an unknown flag or, worse, succeed.
func TestTrimParentRoutesToItsLayers(t *testing.T) {
	cmd := newTrimCmd()
	want := map[string]bool{
		"run": false, "compress": false, "llm-queue": false, "apply-llm": false, "report": false,
	}
	for _, sub := range cmd.Commands() {
		if _, ok := want[sub.Name()]; ok {
			want[sub.Name()] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("trim is missing the %q subcommand", name)
		}
	}
}
