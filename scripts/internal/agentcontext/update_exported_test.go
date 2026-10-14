package agentcontext

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// This file pins the contract added when the agent-file bootstrap was removed
// (2026-09-25): update-agent-context syncs EXISTING context files only, driven
// by the feature's plan.md, and every public seam now routes through Update.
// The package was at 34.9% because Update/extractPlanFields/targets were
// untested — they all resolve the project from the process cwd via the common
// git helpers, so each test builds a real throwaway git repo and chdir's into
// it (Go 1.20 has no t.Chdir). Tests are therefore deliberately serial: no
// t.Parallel anywhere in this package, because a chdir in a parallel test would
// race the working directory.

// --- git + cwd helpers (mirrors internal/common's unexported ones) ---

func reqGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skipf("git not available: %v", err)
	}
}

// initFeatureRepo makes a throwaway git repo on `branch` with one commit (a
// branch with no commit has no HEAD for GetCurrentBranch to read), nested one
// level under t.TempDir() so the short %TEMP% here can't overflow MAX_PATH.
func initFeatureRepo(t *testing.T, branch string) string {
	t.Helper()
	reqGit(t)
	dir := filepath.Join(t.TempDir(), "plaesy-test", "project")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	for _, args := range [][]string{
		{"init", "-q"},
		{"config", "user.email", "test@example.invalid"},
		{"config", "user.name", "Test"},
		{"checkout", "-q", "-b", branch},
		{"commit", "-q", "--allow-empty", "-m", "init"},
	} {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	return dir
}

// cd switches the process working directory for the duration of the test and
// restores it on cleanup. Used only by tests that drive Update (the git helpers
// Read the cwd with no -C flag), so they must not run in parallel.
func cd(t *testing.T, dir string) {
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

// --- plan + context fixtures ---

// writeTempFile writes content to <t.TempDir()>/plaesy-test/<name>, nesting one
// level so the short %TEMP% here cannot overflow MAX_PATH on Windows.
func writeTempFile(t *testing.T, name, content string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "plaesy-test")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}

// writePlan drops a plan.md with the four labeled fields the regexes read.
func writePlan(t *testing.T, path, lang, framework, db, projType string) {
	t.Helper()
	content := fmt.Sprintf(
		"# Implementation Plan\n\n"+
			"**Language/Version**: %s\n"+
			"**Primary Dependencies**: %s\n"+
			"**Storage**: %s\n"+
			"**Project Type**: %s\n", lang, framework, db, projType)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write plan %s: %v", path, err)
	}
}

// writeContext seeds a context file with the standard fixture (Active
// Technologies + Recent Changes + Last updated + a manual block) at path.
func writeContext(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contextFixture), 0o644); err != nil {
		t.Fatalf("write context %s: %v", path, err)
	}
}

// setupRepoWithPlan builds a feature repo, writes plan.md for `branch`, builds
// the target map, and chdir's into the repo so Update sees it as the project.
// Returns the repo root and the target map so callers can address files by key.
func setupRepoWithPlan(t *testing.T, branch, lang, framework, db, projType string) (string, map[string]agentTarget) {
	t.Helper()
	repoRoot := initFeatureRepo(t, branch)
	featureDir := filepath.Join(repoRoot, ".plaesy", "specs", branch)
	if err := os.MkdirAll(featureDir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", featureDir, err)
	}
	writePlan(t, filepath.Join(featureDir, "plan.md"), lang, framework, db, projType)
	tg := targets(repoRoot)
	cd(t, repoRoot)
	return repoRoot, tg
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("readall: %v", err)
	}
	return string(b)
}

// --- extractPlanFields (currently 0% covered) ---

func TestExtractPlanFieldsReadsAllFields(t *testing.T) {
	pf, err := extractPlanFields(writeTempFile(t, "plan.md",
		"**Language/Version**: Rust\n"+
			"**Primary Dependencies**: axum\n"+
			"**Storage**: Postgres\n"+
			"**Project Type**: web\n"))
	if err != nil {
		t.Fatalf("extractPlanFields: %v", err)
	}
	if pf.Lang != "Rust" {
		t.Errorf("Lang = %q, want Rust", pf.Lang)
	}
	if pf.Framework != "axum" {
		t.Errorf("Framework = %q, want axum", pf.Framework)
	}
	if pf.DB != "Postgres" {
		t.Errorf("DB = %q, want Postgres", pf.DB)
	}
	if pf.ProjectType != "web" {
		t.Errorf("ProjectType = %q, want web", pf.ProjectType)
	}
}

// Guards the NEEDS-CLARIFICATION / N/A sentinels: a plan that hasn't resolved a
// field must leave that field blank rather than seeding a garbage technology
// line into an agent file.
func TestExtractPlanFieldsSkipsClarifications(t *testing.T) {
	pf, err := extractPlanFields(writeTempFile(t, "plan.md",
		"**Language/Version**: NEEDS CLARIFICATION\n"+
			"**Primary Dependencies**: NEEDS CLARIFICATION\n"+
			"**Storage**: N/A\n"+
			"**Project Type**: mobile\n"))
	if err != nil {
		t.Fatalf("extractPlanFields: %v", err)
	}
	if pf.Lang != "" {
		t.Errorf("Lang = %q, want empty when NEEDS CLARIFICATION", pf.Lang)
	}
	if pf.Framework != "" {
		t.Errorf("Framework = %q, want empty when NEEDS CLARIFICATION", pf.Framework)
	}
	if pf.DB != "" {
		t.Errorf("DB = %q, want empty when N/A", pf.DB)
	}
	// ProjectType has no clarification guard and is always taken verbatim — that
	// is current behavior; pinning it so a future "guard" is a deliberate change.
	if pf.ProjectType != "mobile" {
		t.Errorf("ProjectType = %q, want mobile", pf.ProjectType)
	}
}

// A plan missing the labeled lines yields empty fields and no error: the parser
// is lenient about absent data, it only fails when the file itself can't be read.
func TestExtractPlanFieldsMissingLabelsIsEmpty(t *testing.T) {
	pf, err := extractPlanFields(writeTempFile(t, "plan.md",
		"# Implementation Plan\n\nNo labeled fields here.\n"))
	if err != nil {
		t.Fatalf("extractPlanFields: %v", err)
	}
	if pf.Lang != "" || pf.Framework != "" || pf.DB != "" || pf.ProjectType != "" {
		t.Errorf("expected all empty, got %+v", pf)
	}
}

func TestExtractPlanFieldsMissingFileErrors(t *testing.T) {
	_, err := extractPlanFields(filepath.Join(t.TempDir(), "does", "not", "exist.md"))
	if err == nil {
		t.Fatal("expected an error when the plan file is missing")
	}
}

// Pins the scanner.Err() propagation on the only non-IO failure path of
// extractPlanFields: a single line longer than bufio's 64KiB token limit makes
// Scan return false with ErrTooLong, and Update wraps that as "reading plan.md".
// Without this, an oversized plan.md fails silently past the os.Stat check.
func TestExtractPlanFieldsOversizedLineReturnsScannerError(t *testing.T) {
	giant := strings.Repeat("x", 70000) // well past bufio.MaxScanTokenSize
	path := writeTempFile(t, "plan.md", giant+"\n")
	if _, err := extractPlanFields(path); err == nil {
		t.Fatal("expected a scanner error for an oversized line")
	} else if !strings.Contains(err.Error(), "bufio.Scanner") {
		t.Errorf("expected bufio.Scanner error, got %q", err)
	}
}

// Pins the Recent Changes "keep last 3" trim in patchAgentFile: once the
// prepend pushes the list past three entries, the oldest must drop off. Without
// this, a long-lived branch grows the section unbounded.
func TestPatchAgentFileTrimsRecentChangesToThree(t *testing.T) {
	content := "## Active Technologies\n\n- Go 1.20\n\n" +
		"## Recent Changes\n\n" +
		"- 003-last: Added alpha\n" +
		"- 002-mid: Added beta\n" +
		"- 001-old: Added something\n\n" +
		"Last updated: 2026-01-01\n"
	path := writeTempFile(t, "CLAUDE.md", content)
	target := agentTarget{key: "claude", file: path, name: "Claude Code"}

	changed, err := patchAgentFile(target, "001-trim", planFields{Lang: "Rust", Framework: "axum"})
	if err != nil {
		t.Fatalf("patchAgentFile: %v", err)
	}
	if !changed {
		t.Error("patchAgentFile reported no change on a file whose Recent Changes section gained an entry")
	}
	got := readAll(t, path)
	if !strings.Contains(got, "- 001-trim: Added Rust + axum") {
		t.Errorf("new change not prepended:\n%s", got)
	}
	if strings.Contains(got, "- 001-old: Added something") {
		t.Errorf("old change should have been trimmed away:\n%s", got)
	}
}

// --- Update, the exported entry point (currently 0% covered) ---

// Update resolves the repo from the process cwd via git, so a non-repo working
// directory surfaces as GetRepoRoot's error before anything else runs.
func TestUpdateNotInAGitRepo(t *testing.T) {
	plain := filepath.Join(t.TempDir(), "plaesy-test", "notgit")
	if err := os.MkdirAll(plain, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cd(t, plain)

	err := Update("")
	if err == nil {
		t.Fatal("expected an error outside a git repository")
	}
	if !strings.Contains(err.Error(), "not a git repository") {
		t.Errorf("error %q does not explain the cause", err)
	}
}

// The bootstrap was deleted on purpose, but a missing plan.md is distinct from a
// missing context file: the former is a prerequisite failure, the latter is a
// "create it yourself" failure. This pins the prerequisite error and its path.
func TestUpdateNoPlanFound(t *testing.T) {
	branch := "001-noplan"
	repo := initFeatureRepo(t, branch)
	cd(t, repo)

	err := Update("claude")
	if err == nil {
		t.Fatal("expected an error when plan.md is missing")
	}
	// Only the suffix is checked, not an absolute root: on some Windows CI
	// runners, two independent os.Getwd() calls after the same os.Chdir
	// returned differently-spelled, equally valid paths to the same
	// directory within the same test run (observed: the production error
	// named the long form, e.g. "...\runneradmin\...", while a Getwd() call
	// made here for comparison returned the short 8.3 form,
	// "...\RUNNER~1\..."). Which one Windows hands back is apparently not
	// stable even within one process, so asserting on the OS-stable part --
	// the relative path under the repo -- is what the test actually needs:
	// that the error names the missing plan.md's location, not which of
	// Windows' two spellings of the temp dir happened to be reported.
	wantSuffix := filepath.Join(".plaesy", "specs", branch, "plan.md")
	if !strings.Contains(err.Error(), "no plan.md found") {
		t.Errorf("error %q should mention no plan.md found", err)
	}
	if !strings.HasSuffix(err.Error(), wantSuffix) {
		t.Errorf("error %q should name a path ending in %s", err, wantSuffix)
	}
}

// A plan.md that exists but contains a runaway line must still fail Update, this
// time through the "reading plan.md" wrap around extractPlanFields' scanner error.
func TestUpdatePlanOversizedLineErrors(t *testing.T) {
	branch := "001-hugeplan"
	repo, tg := setupRepoWithPlan(t, branch, "Rust", "axum", "Postgres", "web")
	// Overwrite plan.md with a single oversized line, then re-seed a context file
	// so the failure is squarely in plan parsing, not the missing-file path.
	featureDir := filepath.Join(repo, ".plaesy", "specs", branch)
	if err := os.WriteFile(filepath.Join(featureDir, "plan.md"),
		[]byte(strings.Repeat("x", 70000)+"\n"), 0o644); err != nil {
		t.Fatalf("write plan: %v", err)
	}
	writeContext(t, tg["claude"].file)

	err := Update("claude")
	if err == nil {
		t.Fatal("expected an error for an oversized plan.md")
	}
	if !strings.Contains(err.Error(), "reading plan.md") {
		t.Errorf("error %q should wrap reading plan.md", err)
	}
}

// An unknown agent type must fail at the run() lookup, not fall through to a
// missing-file error that would read as the user's fault for not creating a file.
func TestUpdateUnknownAgentType(t *testing.T) {
	branch := "001-unknowntype"
	setupRepoWithPlan(t, branch, "Rust", "axum", "Postgres", "web")

	err := Update("nonsense")
	if err == nil {
		t.Fatal("expected an error for an unknown agent type")
	}
	if !strings.Contains(err.Error(), "unknown agent type") {
		t.Errorf("error %q should mention unknown agent type", err)
	}
	if !strings.Contains(err.Error(), "nonsense") {
		t.Errorf("error %q should name the bad type", err)
	}
}

// Explicit agent type on a missing file reuses the updateAgentFile error verbatim
// — it must NOT silently succeed (a run that changed nothing must not report
// success), and it must not create the file.
func TestUpdateExplicitAgentTypeMissingFile(t *testing.T) {
	branch := "001-explicitmissing"
	_, tg := setupRepoWithPlan(t, branch, "Rust", "axum", "Postgres", "web")

	err := Update("claude")
	if err == nil {
		t.Fatal("expected an error when the explicit context file is missing")
	}
	if !strings.Contains(err.Error(), "CLAUDE.md") {
		t.Errorf("error %q should name CLAUDE.md", err)
	}
	if !strings.Contains(err.Error(), "create") {
		t.Errorf("error %q should tell the user to create the file", err)
	}
	if _, statErr := os.Stat(tg["claude"].file); statErr == nil {
		t.Error("Update must not create the context file")
	}
}

// Explicit agent type on an existing file rewrites the Active Technologies /
// Recent Changes / date block and preserves the manual additions verbatim.
func TestUpdateExplicitAgentTypePatchesExisting(t *testing.T) {
	branch := "001-explicitpatch"
	_, tg := setupRepoWithPlan(t, branch, "Rust", "axum", "Postgres", "web")
	writeContext(t, tg["claude"].file)

	if err := Update("claude"); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got := readAll(t, tg["claude"].file)
	for _, want := range []string{
		"- Rust + axum (001-explicitpatch)",
		"- Postgres (001-explicitpatch)",
		"- 001-explicitpatch: Added Rust + axum",
		"hand-written rule",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("patched file missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "Last updated: 2026-01-01") {
		t.Errorf("last-updated date not refreshed:\n%s", got)
	}
}

// No-argument form with NO context files present must error and list every
// path it looked for — this is the contract that replaced the old createAgentFile
// bootstrap. The list in the error is what the user creates from.
func TestUpdateNoArgListsPathsWhenNoContextFile(t *testing.T) {
	branch := "001-noctx"
	setupRepoWithPlan(t, branch, "Rust", "axum", "Postgres", "web")
	// Intentionally no context files written into the repo root.

	err := Update("")
	if err == nil {
		t.Fatal("expected an error when no context file exists")
	}
	msg := err.Error()
	if !strings.Contains(msg, "no agent context file found") {
		t.Errorf("error %q should say no context file found", msg)
	}
	// The error must enumerate what to create, naming each searched path.
	for _, name := range []string{"CLAUDE.md", "GEMINI.md", "AGENTS.md"} {
		if !strings.Contains(msg, name) {
			t.Errorf("error %q should list %s as a candidate to create", msg, name)
		}
	}
	if !strings.Contains(msg, "create one of") {
		t.Errorf("error %q should instruct creating one of the files", msg)
	}
}

// No-argument form with multiple context files updates every one that exists and
// skips the rest without erroring — the success path of the no-arg sweep.
func TestUpdateNoArgUpdatesAllExistingContextFiles(t *testing.T) {
	branch := "001-noargall"
	_, tg := setupRepoWithPlan(t, branch, "Rust", "axum", "Postgres", "web")
	// Two of the six targets exist; the other four are skipped (not errors).
	writeContext(t, tg["claude"].file)
	writeContext(t, tg["opencode"].file)

	if err := Update(""); err != nil {
		t.Fatalf("Update: %v", err)
	}
	for _, key := range []string{"claude", "opencode"} {
		got := readAll(t, tg[key].file)
		if !strings.Contains(got, "- Rust + axum (001-noargall)") {
			t.Errorf("%s not patched with the new tech line", key)
		}
		if !strings.Contains(got, "hand-written rule") {
			t.Errorf("manual block not preserved in %s", key)
		}
	}
	// The four missing targets must remain absent (no scaffold).
	for _, key := range []string{"gemini", "copilot", "cursor", "qwen"} {
		if _, err := os.Stat(tg[key].file); !os.IsNotExist(err) {
			t.Errorf("%s should not have been created (got err=%v)", key, err)
		}
	}
}

// Idempotency through the public API: a second no-arg run must not double the
// Active Technologies entry (the language guard), even though Recent Changes
// keeps prepending — the contract the command leans on is "don't grow tech
// lines", not "rewrite nothing".
func TestUpdateIsIdempotentOnTechLines(t *testing.T) {
	branch := "001-idempotent"
	_, tg := setupRepoWithPlan(t, branch, "Rust", "axum", "Postgres", "web")
	writeContext(t, tg["claude"].file)

	for i := 0; i < 2; i++ {
		if err := Update(""); err != nil {
			t.Fatalf("Update iteration %d: %v", i+1, err)
		}
	}
	got := readAll(t, tg["claude"].file)
	if n := strings.Count(got, "- Rust + axum (001-idempotent)"); n != 1 {
		t.Errorf("tech line duplicated on re-run (%d occurrences)", n)
	}
}

// Recent Changes must hold one line per branch, whatever the command is run
// how many times. This used to prepend unconditionally, so each re-run on a
// branch added another identical line and the keep-last-3 cap pushed out
// *other* branches' history to make room — the file's contents depended on how
// many times the user happened to run the command. Running the command is not
// supposed to be how you lose a branch's entry.
func TestUpdateRecentChangesIsIdempotentPerBranch(t *testing.T) {
	branch := "001-dup"
	_, tg := setupRepoWithPlan(t, branch, "Rust", "axum", "Postgres", "web")
	writeContext(t, tg["claude"].file)

	for i := 0; i < 3; i++ {
		if err := Update(""); err != nil {
			t.Fatalf("Update iteration %d: %v", i+1, err)
		}
	}
	got := readAll(t, tg["claude"].file)
	if n := strings.Count(got, "- 001-dup: Added Rust + axum"); n != 1 {
		t.Errorf("one line per branch, got %d occurrences across 3 runs:%s", n, got)
	}
}

// Re-running on a branch whose plan changed must REFRESH that branch's line, not
// add a second one and not leave the stale technology recorded. A file that
// claims the branch added something it no longer uses is worse than no entry.
func TestUpdateRecentChangesRefreshesAChangedPlan(t *testing.T) {
	branch := "001-refresh"
	dir, tg := setupRepoWithPlan(t, branch, "Rust", "axum", "Postgres", "web")
	writeContext(t, tg["claude"].file)

	if err := Update(""); err != nil {
		t.Fatalf("first run: %v", err)
	}
	// The plan now names a different framework.
	plan := filepath.Join(dir, ".plaesy", "specs", branch, "plan.md")
	body, err := os.ReadFile(plan)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(plan, []byte(strings.ReplaceAll(string(body), "axum", "actix")), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Update(""); err != nil {
		t.Fatalf("second run: %v", err)
	}
	got := readAll(t, tg["claude"].file)
	if strings.Contains(got, "- 001-refresh: Added Rust + axum") {
		t.Errorf("the superseded framework must not survive a re-run:%s", got)
	}
	if !strings.Contains(got, "- 001-refresh: Added Rust + actix") {
		t.Errorf("the line must be refreshed to the new framework:%s", got)
	}
}

// A plan that never resolved its language (NEEDS CLARIFICATION) previously
// produced "Added  + Gin" — a double space exactly where the missing value
// should have been read as missing, instead looking like a typo in the tool.
func TestUpdateRecentChangesOmitsAnUnresolvedLanguage(t *testing.T) {
	branch := "001-nolang"
	dir, tg := setupRepoWithPlan(t, branch, "Rust", "axum", "Postgres", "web")
	plan := filepath.Join(dir, ".plaesy", "specs", branch, "plan.md")
	body, err := os.ReadFile(plan)
	if err != nil {
		t.Fatal(err)
	}
	blurred := strings.Replace(string(body), "**Language**: Rust", "**Language**: v1.0 | NEEDS CLARIFICATION", 1)
	if blurred == string(body) {
		t.Skip("plan fixture no longer has the expected Language line")
	}
	if err := os.WriteFile(plan, []byte(blurred), 0o644); err != nil {
		t.Fatal(err)
	}
	writeContext(t, tg["claude"].file)
	if err := Update(""); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got := readAll(t, tg["claude"].file)
	if strings.Contains(got, "Added  +") {
		t.Errorf("an unresolved language must not leave a double space:%s", got)
	}
	if !strings.Contains(got, "- 001-nolang: Added axum") {
		t.Errorf("the framework that IS known must still be recorded:%s", got)
	}
}

// A context file that exists but cannot be rewritten (read-only) must make the
// no-arg sweep fail at that file rather than exit 0 having changed nothing. A
// run that could not do the work must not report success — the trimmed bootstrap
// exists so a silently-no-op run stays impossible; this guards the rewrite path
// of that same rule.
func TestUpdateNoArgFailsWhenContextFileUnwritable(t *testing.T) {
	branch := "001-ro"
	_, tg := setupRepoWithPlan(t, branch, "Rust", "axum", "Postgres", "web")
	claudePath := tg["claude"].file
	writeContext(t, claudePath)
	if err := os.Chmod(claudePath, 0o444); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(claudePath, 0o644) })

	err := Update("")
	if err == nil {
		t.Fatal("expected an error when the context file cannot be rewritten")
	}
	if !strings.Contains(err.Error(), "CLAUDE.md") {
		t.Errorf("error %q should name the failing file", err)
	}
}

// The console summary must reflect what was extracted from plan.md, including
// the N/A and empty-field skips — a summary that prints a field Update never
// acted on would mislead the user about what changed.
func TestUpdateSummaryReflectsPlanFields(t *testing.T) {
	cases := []struct {
		name                            string
		lang, framework, db, projType   string
		wantLang, wantFramework, wantDB bool
	}{
		{"with database", "Ruby", "Rails", "Postgres", "web", true, true, true},
		{"database N/A", "Ruby", "Rails", "N/A", "web", true, true, false},
		{"language missing", "NEEDS CLARIFICATION", "cobra", "SQLite", "mobile", false, true, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Each subtest gets its own repo; a single valid branch name is shared.
			_, tg := setupRepoWithPlan(t, "001-summary", tc.lang, tc.framework, tc.db, tc.projType)
			writeContext(t, tg["claude"].file)
			// NEEDS CLARIFICATION makes Lang empty, which suppresses the language
			// summary line AND skips adding a tech line — still a successful run.
			// The summary is captured from the run that *changed* the file. It
			// used to be captured from the second run, which is idempotent and
			// therefore changed nothing — and Update rendered that summary from
			// plan.md rather than from what it wrote, so a run that touched
			// nothing announced every field it had extracted. That is exactly
			// what this test exists to catch, and reading it from the second run
			// was the one way it could not.
			var first string
			first = captureStdout(t, func() {
				if err := Update("claude"); err != nil {
					t.Fatalf("Update: %v", err)
				}
			})
			// The second run must be a clean no-op, not a second announcement.
			second := captureStdout(t, func() {
				if err := Update("claude"); err != nil {
					t.Fatalf("Update (re-run): %v", err)
				}
			})
			if strings.Contains(second, "Summary of changes:") {
				t.Errorf("an idempotent re-run must not print a change summary:\n%s", second)
			}
			out := first
			check := func(sub string, want bool) {
				if want == strings.Contains(out, sub) {
					return
				}
				t.Errorf("summary contains %q = %v, want %v:\n%s", sub, strings.Contains(out, sub), want, out)
			}
			check("Added language:", tc.wantLang)
			check("Added framework:", tc.wantFramework)
			check("Added database:", tc.wantDB)
		})
	}
}

// --- registry vs docs drift guard (cf. scaffold/copy_test.go) ---

// docRoot walks up from this test's source file to the repository root, the dir
// that holds docs/scripts/update-agent-context.md. runtime.Caller(0) returns the
// source path (scaffold/copy_test.go relies on the same), so it is valid at test
// time even though the binary is compiled to a temp dir.
func docRoot(t *testing.T) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "docs", "scripts", "README.md")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not locate docs/scripts/README.md from %s", file)
		}
		dir = parent
	}
}

// agentContextDoc locates the doc page describing `plaesy context update`. The
// command has been renamed twice — `update-agent-context`, then `context
// update` — and the file that documents it has been renamed to match both
// times. Globbing keeps these tests about the agent-type table they actually
// check rather than about where the page happens to live.
func agentContextDoc(t *testing.T) string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(docRoot(t), "docs", "scripts", "*context*.md"))
	if err != nil {
		t.Fatalf("glob for the context-update doc: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected exactly one docs/scripts/*context*.md page, found %v", matches)
	}
	return matches[0]
}

func stripBackticks(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '`' {
			return -1
		}
		return r
	}, s)
}

// parseAgentTypeTable pulls the agent-type rows out of the doc's Supported Agent
// Types table, returning key -> context-file-as-printed-in-docs.
func parseAgentTypeTable(markdown string) map[string]string {
	out := map[string]string{}
	valid := regexp.MustCompile(`^(claude|gemini|copilot|cursor|qwen|opencode)$`)
	for _, line := range strings.Split(markdown, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") || !strings.HasSuffix(line, "|") {
			continue
		}
		cells := strings.Split(strings.Trim(line, "|"), "|")
		if len(cells) < 2 {
			continue
		}
		key := stripBackticks(strings.TrimSpace(cells[0]))
		if !valid.MatchString(key) {
			continue // header, separator, and the "*(none)*" row
		}
		out[key] = stripBackticks(strings.TrimSpace(cells[1]))
	}
	return out
}

// The set of agent types and their context files lives in three places that have
// drifted before: the targets() registry, the docs table, and the cobra Use
// arg list. This pins all three to one another so a rename in one place cannot
// silently desync the others.
func TestTargetRegistryMatchesDocs(t *testing.T) {
	tg := targets("")

	if len(tg) != len(orderedKeys) {
		t.Fatalf("targets() has %d keys, orderedKeys has %d", len(tg), len(orderedKeys))
	}
	for _, k := range orderedKeys {
		if _, ok := tg[k]; !ok {
			t.Errorf("orderedKeys contains %q not present in targets()", k)
		}
	}
	for k := range tg {
		if !containsKey(orderedKeys, k) {
			t.Errorf("targets() has key %q not present in orderedKeys", k)
		}
	}

	docPath := agentContextDoc(t)
	doc, err := os.ReadFile(docPath)
	if err != nil {
		t.Fatalf("read doc %s: %v", docPath, err)
	}
	rows := parseAgentTypeTable(string(doc))
	if len(rows) != len(orderedKeys) {
		t.Fatalf("docs table lists %d agent types, registry has %d", len(rows), len(orderedKeys))
	}
	for _, k := range orderedKeys {
		f, ok := rows[k]
		if !ok {
			t.Errorf("docs table missing agent type %q", k)
			continue
		}
		want := filepath.Base(tg[k].file)
		got := filepath.Base(f)
		if want != got {
			t.Errorf("docs say %q -> %s but registry -> %s", k, got, want)
		}
	}
}

func TestCobraUseArgsMatchRegistry(t *testing.T) {
	// The command was renamed twice — `update-agent-context`, then
	// `create`/`context update` — so the file that holds its Use string moved
	// too. Resolving it by glob rather than by the old path keeps this test
	// about the argument list it actually checks instead of about where the
	// command happens to live.
	//
	// The pattern is anchored on `Use:` because the parent `context` command's
	// Long help also shows `plaesy context update [agent]` as an example. An
	// unanchored `update \[...\]` matched that placeholder first and reported a
	// one-argument list against a six-entry registry.
	src, err := os.ReadFile(filepath.Join(docRoot(t), "scripts", "cmd", "plaesy", "context.go"))
	if err != nil {
		t.Fatalf("read context.go: %v", err)
	}
	m := regexp.MustCompile(`(?m)^\s*Use:\s+"update \[([^\]]+)\]"`).FindStringSubmatch(string(src))
	if m == nil {
		t.Fatalf("could not parse Use arg list from context.go")
	}
	args := strings.Split(m[1], "|")
	if len(args) != len(orderedKeys) {
		t.Fatalf("cobra Use lists %d args, registry has %d", len(args), len(orderedKeys))
	}
	for i, k := range orderedKeys {
		if args[i] != k {
			t.Errorf("cobra arg[%d] = %q, want %q", i, args[i], k)
		}
	}
}

func containsKey(keys []string, k string) bool {
	for _, s := range keys {
		if s == k {
			return true
		}
	}
	return false
}

func readAll(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

// A plan that resolves neither a language nor a framework must still produce a
// well-formed line. The old positional format wrote "- branch: Added  + " — a
// trailing " + " and nothing after it, which reads as a truncated write rather
// than as "nothing known yet".
func TestUpdateRecentChangesWithNothingResolved(t *testing.T) {
	branch := "001-empty"
	dir, tg := setupRepoWithPlan(t, branch, "Rust", "axum", "Postgres", "web")
	plan := filepath.Join(dir, ".plaesy", "specs", branch, "plan.md")
	body, err := os.ReadFile(plan)
	if err != nil {
		t.Fatal(err)
	}
	blurred := string(body)
	for _, old := range []string{"Rust", "axum"} {
		blurred = strings.Replace(blurred, old, "NEEDS CLARIFICATION", 1)
	}
	if blurred == string(body) {
		t.Skip("plan fixture no longer carries the expected values")
	}
	if err := os.WriteFile(plan, []byte(blurred), 0o644); err != nil {
		t.Fatal(err)
	}
	writeContext(t, tg["claude"].file)
	if err := Update(""); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got := readAll(t, tg["claude"].file)
	if strings.Contains(got, "Added  ") || strings.Contains(got, " + \n") {
		t.Errorf("an unresolved plan must not leave an empty or dangling value:\n%s", got)
	}
	if !strings.Contains(got, "- 001-empty:") {
		t.Errorf("the branch must still be recorded:\n%s", got)
	}
}
