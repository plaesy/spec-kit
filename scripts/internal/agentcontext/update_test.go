package agentcontext

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const contextFixture = `# Project

## Active Technologies

- Go 1.20

## Recent Changes

- 001-old: Added something

Last updated: 2026-01-01

<!-- MANUAL ADDITIONS START -->
hand-written rule
<!-- MANUAL ADDITIONS END -->
`

// seedContext writes the fixture to a temp file and returns its path.
func seedContext(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(contextFixture), 0o644); err != nil {
		t.Fatalf("seed fixture: %v", err)
	}
	return path
}

// TestUpdateAgentFileRequiresExistingFile locks in that update-agent-context
// never scaffolds a context file. The agent-file template was removed
// (2026-09-25), so creating a file is no longer this command's job; a missing
// target must fail with actionable guidance rather than a bare read error.
func TestUpdateAgentFileRequiresExistingFile(t *testing.T) {
	dir := t.TempDir()
	target := agentTarget{key: "claude", file: filepath.Join(dir, "CLAUDE.md"), name: "Claude Code"}

	_, err := updateAgentFile(target, "001-feature", planFields{Lang: "Go 1.20"})
	if err == nil {
		t.Fatal("expected an error when the context file does not exist")
	}
	msg := err.Error()
	if !strings.Contains(msg, "CLAUDE.md") {
		t.Errorf("error should name the missing file, got %q", msg)
	}
	if !strings.Contains(msg, "create") {
		t.Errorf("error should tell the user to create the file, got %q", msg)
	}
	if _, statErr := os.Stat(target.file); statErr == nil {
		t.Error("updateAgentFile must not create the context file")
	}
}

// TestUpdateAgentFilePatchesExistingFile covers the remaining behavior: a new
// language/framework pair and the storage line are appended, the change is
// prepended to Recent Changes, the date is refreshed, and the manual block
// survives verbatim.
func TestUpdateAgentFilePatchesExistingFile(t *testing.T) {
	path := seedContext(t, "AGENTS.md")
	target := agentTarget{key: "opencode", file: path, name: "opencode"}

	if _, err := updateAgentFile(target, "002-new", planFields{Lang: "Rust", Framework: "axum", DB: "Postgres"}); err != nil {
		t.Fatalf("updateAgentFile: %v", err)
	}

	out, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read patched file: %v", err)
	}
	got := string(out)

	for _, want := range []string{
		"- Rust + axum (002-new)",
		"- Postgres (002-new)",
		"- 002-new: Added Rust + axum",
		"hand-written rule", // manual block preserved
	} {
		if !strings.Contains(got, want) {
			t.Errorf("patched file missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "Last updated: 2026-01-01") {
		t.Errorf("last-updated date not refreshed:\n%s", got)
	}
}

// TestUpdateAgentFileDoesNotDuplicateKnownTechnology guards the idempotency the
// command depends on when it runs repeatedly on the same plan.
func TestUpdateAgentFileDoesNotDuplicateKnownTechnology(t *testing.T) {
	path := seedContext(t, "CLAUDE.md")
	target := agentTarget{key: "claude", file: path, name: "Claude Code"}
	pf := planFields{Lang: "Rust", Framework: "axum", DB: "Postgres"}

	for i := 0; i < 2; i++ {
		if _, err := updateAgentFile(target, "003-a", pf); err != nil {
			t.Fatalf("update %d: %v", i+1, err)
		}
	}
	out, _ := os.ReadFile(path)

	if n := strings.Count(string(out), "Rust + axum (003-a)"); n != 1 {
		t.Errorf("language/framework line duplicated on re-run (%d occurrences):\n%s", n, out)
	}
	if n := strings.Count(string(out), "Postgres (003-a)"); n != 1 {
		t.Errorf("storage line duplicated on re-run (%d occurrences):\n%s", n, out)
	}
}

// TestUpdateAgentFileKeepsKnownLanguageOnce documents current behavior: a plan
// whose language is already listed adds nothing to Active Technologies, even if
// its framework is new. The language acts as the marker for the whole pair. This
// is a known gap, not a contract the command intends to keep forever — the test
// exists so a future change to framework-level tracking is a deliberate, visible
// decision rather than a silent behavior shift.
func TestUpdateAgentFileKeepsKnownLanguageOnce(t *testing.T) {
	path := seedContext(t, "CLAUDE.md")
	target := agentTarget{key: "claude", file: path, name: "Claude Code"}

	if _, err := updateAgentFile(target, "004-b", planFields{Lang: "Go 1.20", Framework: "cobra"}); err != nil {
		t.Fatalf("updateAgentFile: %v", err)
	}
	out, _ := os.ReadFile(path)

	tech := activeTechRE.FindStringSubmatch(string(out))
	if tech == nil {
		t.Fatalf("no Active Technologies section after patch:\n%s", out)
	}
	if strings.Contains(tech[1], "cobra") {
		t.Errorf("known language currently suppresses the framework line; this test exists to make a change to that visible:\n%s", tech[1])
	}
	if n := strings.Count(tech[1], "- Go 1.20"); n != 1 {
		t.Errorf("known language duplicated in Active Technologies (%d occurrences):\n%s", n, tech[1])
	}
}

// All three of patchAgentFile's patterns are optional matches, and the file is
// written back regardless of whether any of them matched — so a context file
// with none of the tracked sections was rewritten byte-identically and then
// reported as "updated successfully". Nothing in the framework creates those
// sections, so a user's own CLAUDE.md/AGENTS.md is exactly this case, and the
// whole run claimed work it did not do. The rule updateAgentFile documents —
// "a run that changed nothing must not report success" — has to survive being
// applied to the file it is written about.
func TestUpdateAgentFileFailsOnAFileWithNoTrackedSections(t *testing.T) {
	path := filepath.Join(t.TempDir(), "CLAUDE.md")
	body := "# My project\n\nJust a README-ish agent file.\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	target := agentTarget{key: "claude", file: path, name: "Claude Code"}

	changed, err := updateAgentFile(target, "001-x", planFields{Lang: "Go 1.20"})
	if err == nil {
		t.Fatal("expected an error when the file has no patchable sections")
	}
	if changed {
		t.Error("a failed patch must not report a change")
	}
	if !strings.Contains(err.Error(), "CLAUDE.md") {
		t.Errorf("error should name the file, got %q", err)
	}
	got, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("read: %v", readErr)
	}
	if string(got) != body {
		t.Errorf("the file was modified on a failed run:\n%s", got)
	}
}

// The mirror case: a file that IS already in sync changed nothing, and must
// say so rather than claim an update — but must not be an error either, or
// re-running the command on an unchanged plan becomes a failure.
func TestUpdateAgentFileReportsAnUnchangedFileWithoutClaimingAnUpdate(t *testing.T) {
	path := seedContext(t, "CLAUDE.md")
	target := agentTarget{key: "claude", file: path, name: "Claude Code"}
	pf := planFields{Lang: "Rust", Framework: "axum", DB: "Postgres"}

	if changed, err := updateAgentFile(target, "003-a", pf); err != nil || !changed {
		t.Fatalf("first run: changed=%v err=%v, want a real change", changed, err)
	}
	changed, err := updateAgentFile(target, "003-a", pf)
	if err != nil {
		t.Fatalf("idempotent re-run must not be an error: %v", err)
	}
	if changed {
		t.Error("an already-in-sync file reported a change")
	}
}
