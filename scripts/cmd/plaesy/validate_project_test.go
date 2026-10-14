package main

// Tests for bare `plaesy validate` — the default path, with no arguments and no
// target. It is what a user types to ask "is this project ok?", so the property
// that matters most is that the summary cannot claim success while a check
// failed or was quietly dropped.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// withWorkingDirectory switches the process cwd for one test. runValidateProject
// resolves the project with common.GetRepoRoot(), which shells out to git, so
// the only seam available is the cwd (Go 1.20 has no t.Chdir). Callers must not
// use t.Parallel.
func withWorkingDirectory(t *testing.T, dir string) {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(prev); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})
}

// validateProject builds a throwaway project. The constitution is the committed
// fixture in testdata/, not `.plaesy/memory/constitution.md`: that path holds
// generated, untracked project state (`make clean` deletes it and nothing
// regenerates it in a checkout), so reading it here made the suite fail on any
// clean tree. The fixture is still a real constitution, validated by the same
// production validator, rather than a hand-tuned stub.
func validateProject(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "plaesy-test", "project")
	memory := filepath.Join(dir, ".plaesy", "memory")
	if err := os.MkdirAll(memory, 0o755); err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(filepath.Join("testdata", "constitution.md"))
	if err != nil {
		t.Fatalf("read the constitution fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(memory, "constitution.md"), source, 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// The bare run must report all three checks, and say which ones it skipped and
// why. A check that prints nothing reads as a passing one, which is the whole
// reason [SKIP] exists.
func TestBareValidateNamesEveryCheckItRan(t *testing.T) {
	dir := validateProject(t)
	withWorkingDirectory(t, dir)

	out, err := captureStdout(t, func() error { return runValidateProject() })
	if err != nil {
		t.Fatalf("bare validate failed on a valid project: %v\n%s", err, out)
	}
	for _, check := range []string{"constitution", "memory", "markdown"} {
		if !strings.Contains(out, "--- "+check+" ---") {
			t.Errorf("the summary does not report the %s check:\n%s", check, out)
		}
	}
	if !strings.Contains(out, "2 passed") {
		t.Errorf("summary does not count the two runnable checks as passing:\n%s", out)
	}
	// Markdown is the one that skips here: the fixture opts into nothing, and a
	// bare lint over an unconfigured tree would demand zero violations, which is
	// a different check from the one the user asked for.
	if !strings.Contains(out, "[SKIP] markdown") {
		t.Errorf("an unconfigured project must say the markdown check was skipped:\n%s", out)
	}
	if !strings.Contains(out, "1 skipped") {
		t.Errorf("summary does not count the skipped check:\n%s", out)
	}
	// The skip has to explain itself, or the user cannot tell what to do.
	if !strings.Contains(out, "validate markdown") {
		t.Errorf("the skip reason does not say how to opt in:\n%s", out)
	}
}

// Once a project has opted into Markdown linting, the bare run must include it
// rather than still reporting it as skipped. A skip that never lifts is a check
// that never runs.
func TestBareValidateRunsMarkdownOnceTheProjectOptsIn(t *testing.T) {
	dir := validateProject(t)
	baseline := filepath.Join(dir, ".markdownlint-baseline.json")
	// A real ceiling, not zero: a baseline of 0 is rejected as unable to express
	// a ratchet, which is itself worth knowing but is not what this test is for.
	if err := os.WriteFile(baseline, []byte(`{"max_violations": 5, "files": {}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	withWorkingDirectory(t, dir)

	out, err := captureStdout(t, func() error { return runValidateProject() })
	if err != nil {
		t.Fatalf("bare validate failed on a valid project: %v\n%s", err, out)
	}
	if strings.Contains(out, "[SKIP] markdown") {
		t.Errorf("markdown was skipped even though a baseline exists:\n%s", out)
	}
	if !strings.Contains(out, "3 passed") {
		t.Errorf("summary does not count all three checks as passing:\n%s", out)
	}
	if strings.Contains(out, "skipped") {
		t.Errorf("a fully configured project still reports a skip:\n%s", out)
	}
}

// A failing check must fail the run. The exit code is the whole contract for CI:
// a constitution that does not validate has to stop a pipeline.
func TestBareValidateFailsWhenAProjectWideCheckFails(t *testing.T) {
	dir := validateProject(t)
	// A constitution with an unfilled placeholder: the exact class the validator
	// exists for, and one the project's own file never has.
	broken := strings.Replace(string(mustReadFile(t, filepath.Join(dir, ".plaesy", "memory", "constitution.md"))),
		"version:", "version: TODO", 1)
	if err := os.WriteFile(filepath.Join(dir, ".plaesy", "memory", "constitution.md"), []byte(broken), 0o644); err != nil {
		t.Fatal(err)
	}
	withWorkingDirectory(t, dir)

	out, err := captureStdout(t, func() error { return runValidateProject() })
	if err == nil {
		t.Fatalf("an invalid constitution must fail the run\n%s", out)
	}
	// The message has to count, so a reader can tell one check failed from three.
	if !strings.Contains(err.Error(), "failed") {
		t.Errorf("error %q does not say a check failed", err)
	}
	// The summary is the report, and it has to admit the failure rather than
	// rounding it away: constitution failed, memory passed, markdown skipped.
	if !strings.Contains(out, "1 passed, 1 failed, 1 skipped") {
		t.Errorf("summary does not report the failure alongside the passes:\n%s", out)
	}
	// The checks that did pass still ran: one failure must not hide the rest.
	if !strings.Contains(out, "--- memory ---") {
		t.Errorf("a constitution failure stopped the other checks:\n%s", out)
	}
}

func TestBareValidateSucceedsOnAProjectWithNothingToCheck(t *testing.T) {
	// A fresh project: no constitution, no memory, no Markdown config. Every check
	// is inapplicable, and "check my project" is a question with an answer — not
	// an error. Three skips and a success is the correct report.
	dir := filepath.Join(t.TempDir(), "plaesy-test", "project")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	withWorkingDirectory(t, dir)

	out, err := captureStdout(t, func() error { return runValidateProject() })
	if err != nil {
		t.Fatalf("a fresh project must not fail: %v\n%s", err, out)
	}
	if !strings.Contains(out, "0 passed") || !strings.Contains(out, "3 skipped") {
		t.Errorf("summary does not report three skips and no passes:\n%s", out)
	}
	for _, check := range []string{"constitution", "memory", "markdown"} {
		if !strings.Contains(out, "[SKIP] "+check) {
			t.Errorf("%s was not reported as skipped:\n%s", check, out)
		}
	}
}

// The bare run must not touch anything. `plaesy validate` is the check people
// wire into a pre-commit hook, so a run that rewrote a file would be a surprise
// with no flag to prevent it.
func TestBareValidateWritesNothing(t *testing.T) {
	dir := validateProject(t)
	before := snapshotTree(t, dir)
	withWorkingDirectory(t, dir)

	out, err := captureStdout(t, func() error { return runValidateProject() })
	if err != nil {
		t.Fatalf("bare validate failed: %v\n%s", err, out)
	}
	if after := snapshotTree(t, dir); after != before {
		t.Errorf("bare validate changed the project:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// `plaesy validate --list` writes to the command's output stream, so it can be
// piped or redirected. The bare-run help text is the discoverability surface for
// every target, and it is only useful if the two cannot drift apart silently.
func TestValidateListCoversEveryRegisteredTarget(t *testing.T) {
	var out strings.Builder
	cmd := newValidateCmd()
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--list"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("validate --list failed: %v", err)
	}
	for _, target := range validateTargets {
		if !strings.Contains(out.String(), target.name) {
			t.Errorf("--list does not mention %q:\n%s", target.name, out.String())
		}
	}
	// The project-wide ones are the ones the bare run executes; saying so is what
	// stops a user wondering what bare `validate` covers.
	if !strings.Contains(out.String(), "project-wide") {
		t.Errorf("--list does not mark which targets the bare run covers:\n%s", out.String())
	}
}

// A document argument is dispatched by extension, so `plaesy validate README.md`
// does what the user meant rather than reporting a usage error.
func TestValidateFileArgumentDispatchesByExtension(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "plaesy-test", "project")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Title\n\nSome text.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// The cwd matters: the validator resolves the project root from git, so
	// without this the file argument is linted against the real repository and
	// found to be outside it.
	withWorkingDirectory(t, dir)

	captured, err := captureStdout(t, func() error { return runValidateFiles([]string{"README.md"}) })
	if err != nil {
		t.Fatalf("validate <file.md> failed: %v\n%s", err, captured)
	}
	// Names exactly how many files it scanned, so a file the linter skipped
	// silently cannot pass as one it read.
	if !strings.Contains(captured, "scanned 1 markdown file(s)") {
		t.Errorf("the .md file was not actually linted:\n%s", captured)
	}
}

func TestValidateFileArgumentRejectsAnUnknownExtension(t *testing.T) {
	_, err := captureStdout(t, func() error { return runValidateFiles([]string{"notes.txt"}) })
	if err == nil {
		t.Fatal("an extension with no validator must be an error, not a silent pass")
	}
	// The message lists what it can dispatch to, or the user has to guess.
	for _, name := range validateTargetNames() {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error %q does not list %q", err, name)
		}
	}
}

// A file that does not exist must be an error rather than a check that quietly
// passes: "validate this document" on a typo is not a pass.
func TestValidateFileArgumentRejectsAMissingFile(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nope.md")
	_, err := captureStdout(t, func() error { return runValidateFiles([]string{missing}) })
	if err == nil {
		t.Error("validating a file that does not exist must not succeed")
	}
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

// snapshotTree renders every file under root with its content, so "did this
// command change anything" is answerable without guessing which paths matter.
func snapshotTree(t *testing.T, root string) string {
	t.Helper()
	var b strings.Builder
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		b.WriteString(filepath.ToSlash(rel))
		b.WriteString("\n")
		b.Write(data)
		b.WriteString("\n")
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot %s: %v", root, err)
	}
	return b.String()
}
