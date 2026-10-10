package main

// Tests for the `plaesy validate <target>` shape. The grouping is the point of
// the refactor — six top-level commands that all did the same verb — so the
// tests pin the shape, not just the behaviour behind it.

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func rootForTest() *cobra.Command {
	root := &cobra.Command{Use: "plaesy", SilenceUsage: true, SilenceErrors: true}
	root.AddCommand(registry...)
	return root
}

// The help text is the discoverability surface. Six near-identical top-level
// entries used to crowd out commands that do different things; the deprecated
// aliases must stay out of it entirely.
func TestValidateHelpListsOneVerbAndNoLegacyNames(t *testing.T) {
	var out bytes.Buffer
	root := rootForTest()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("root --help: %v", err)
	}
	help := out.String()

	if !strings.Contains(help, "validate ") {
		t.Errorf("help does not mention validate:\n%s", help)
	}
	for _, old := range []string{"validate-constitution", "validate-memory", "validate-markdown",
		"validate-docx", "validate-pptx", "validate-xlsx"} {
		if strings.Contains(help, old) {
			t.Errorf("help still lists the deprecated %q", old)
		}
	}
}

func TestValidateParentHasEveryTargetAsASubcommand(t *testing.T) {
	cmd := newValidateCmd()
	have := map[string]bool{}
	for _, sub := range cmd.Commands() {
		have[sub.Name()] = true
	}
	for _, target := range validateTargets {
		if !have[target.name] {
			t.Errorf("no subcommand for target %q", target.name)
		}
		if target.build().Short == "" {
			t.Errorf("target %q has no Short, so --list prints a blank column", target.name)
		}
	}
}

func TestValidateListNamesEveryTarget(t *testing.T) {
	var out bytes.Buffer
	cmd := newValidateCmd()
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--list"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("validate --list: %v", err)
	}
	for _, target := range validateTargets {
		if !strings.Contains(out.String(), target.name) {
			t.Errorf("--list omits %q:\n%s", target.name, out.String())
		}
	}
}

// A file argument the binary cannot classify must say what it accepts. Guessing
// would either validate nothing or validate the wrong thing, and both report
// success.
func TestValidateUnknownArgumentNamesTheValidTargets(t *testing.T) {
	err := runValidateFiles([]string{"notes.txt"})
	if err == nil {
		t.Fatal("expected an error for an unclassifiable argument")
	}
	msg := err.Error()
	for _, target := range validateTargets {
		if !strings.Contains(msg, target.name) {
			t.Errorf("error does not mention %q: %s", target.name, msg)
		}
	}
	if !strings.Contains(msg, "notes.txt") {
		t.Errorf("error does not name the offending argument: %s", msg)
	}
}

// Old spellings must keep working — prompts, docs, and muscle memory all use
// them — and must announce where they moved.
func TestLegacyValidateAliasIsDeprecatedAndRepointed(t *testing.T) {
	for _, target := range validateTargets {
		alias := legacyValidateAlias(target)
		if want := "validate-" + target.name; !strings.HasPrefix(alias.Use, want) {
			t.Errorf("alias for %q has Use %q, want prefix %q", target.name, alias.Use, want)
		}
		if !strings.Contains(alias.Deprecated, "plaesy validate "+target.name) {
			t.Errorf("alias for %q does not point at the new spelling: %q", target.name, alias.Deprecated)
		}
		// Deprecated commands are hidden by cobra, which is what keeps the help
		// text to one verb.
		if alias.Deprecated == "" {
			t.Errorf("alias for %q is not deprecated", target.name)
		}
	}
}

func TestMarkdownSkipReasonExplainsAnUnconfiguredProject(t *testing.T) {
	dir := t.TempDir()
	reason := markdownSkipReason(dir)
	if reason == "" {
		t.Fatal("a project with no Markdown config must skip the lint, not require zero violations repo-wide")
	}
	if !strings.Contains(reason, "validate markdown") {
		t.Errorf("skip reason does not say how to opt in: %q", reason)
	}
}

func TestMarkdownSkipReasonIsEmptyOnceConfigured(t *testing.T) {
	dir := t.TempDir()
	if err := writeFileForTest(dir, ".markdownlint-baseline.json", `{"max_violations": 1}`); err != nil {
		t.Fatal(err)
	}
	if reason := markdownSkipReason(dir); reason != "" {
		t.Errorf("a configured project must not skip the lint, got %q", reason)
	}
}

func TestSkipIfMissingNamesTheFileThatIsAbsent(t *testing.T) {
	dir := t.TempDir()
	if reason := skipIfMissing(dir+"/constitution.md", "none yet"); reason != "none yet" {
		t.Errorf("missing file should report the reason, got %q", reason)
	}
	if err := writeFileForTest(dir, "constitution.md", "# c\n"); err != nil {
		t.Fatal(err)
	}
	if reason := skipIfMissing(dir+"/constitution.md", "none yet"); reason != "" {
		t.Errorf("present file must not be skipped, got %q", reason)
	}
}

// TestNoBaselineRequiresZeroViolations proves the flag does what its help text
// says. It used to print its own violations and exit 0, because "no baseline"
// was read downstream as "no ceiling" — the worst kind of bug in a check, since
// the output looks like a pass with detail attached.
func TestNoBaselineRequiresZeroViolations(t *testing.T) {
	dir := t.TempDir()
	dirty := filepath.Join(dir, "dirty.md")
	// Trailing whitespace (MD009) with no final newline (MD047).
	if err := writeFileForTest(dir, "dirty.md", "# Title\n\ntrailing spaces   \n\n# Second\n\ntext   "); err != nil {
		t.Fatal(err)
	}
	if err := runMarkdownCmd(t, dirty, "--no-baseline"); err == nil {
		t.Error("--no-baseline must fail on a file with violations, got success")
	}

	clean := filepath.Join(dir, "clean.md")
	if err := writeFileForTest(dir, "clean.md", "# Title\n\nA line of prose.\n\n## Section\n\nMore prose.\n"); err != nil {
		t.Fatal(err)
	}
	if err := runMarkdownCmd(t, clean, "--no-baseline"); err != nil {
		t.Errorf("--no-baseline must pass a clean file, got %v", err)
	}
}

// runMarkdownCmd executes the markdown validator in-process with the given
// arguments and returns the error, which is what the process exit code is built
// from. Testing the error rather than the printed text is the point: the old
// failure printed a tidy [OK] line while the error was nil.
func runMarkdownCmd(t *testing.T, args ...string) error {
	t.Helper()
	cmd := newMarkdownCmd()
	cmd.SetArgs(args)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	return cmd.Execute()
}

func writeFileForTest(dir, name, content string) error {
	return os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)
}

// A fresh project has no .plaesy/memory yet, and the bare run must not turn
// that into a failure: "check my project" is a question with an answer, not an
// error. The explicit `validate memory` keeps reporting the missing directory,
// because there the user asked for something that is not there.
func TestBareValidateSkipsMemoryInAFreshProject(t *testing.T) {
	dir := t.TempDir()
	if reason := skipIfMissing(filepath.Join(dir, ".plaesy", "memory"), "no memory yet"); reason != "no memory yet" {
		t.Fatalf("a project without .plaesy/memory must skip the check, got %q", reason)
	}
	if err := writeFileForTest(dir, "placeholder.md", "x\n"); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".plaesy", "memory"), 0o755); err != nil {
		t.Fatal(err)
	}
	if reason := skipIfMissing(filepath.Join(dir, ".plaesy", "memory"), "no memory yet"); reason != "" {
		t.Errorf("an existing memory directory must not be skipped, got %q", reason)
	}
}
