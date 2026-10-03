package mdlint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fixableRules are the rules Fix() is allowed to repair. A document that still
// reports one of them after a fix pass is an unfixed violation, not a style
// choice.
var fixableRules = []string{"MD009", "MD010", "MD012", "MD022", "MD031", "MD032", "MD047", "MD058"}

// fixAndLint runs Fix over a.md and then Lints the result, returning the fixed
// text and the remaining violations per rule.
func fixAndLint(t *testing.T, content string) (string, map[string]int) {
	t.Helper()
	root := writeFiles(t, map[string]string{
		".markdownlint.json": minimalConfig,
		"a.md":               content,
	})
	path := filepath.Join(root, "a.md")
	if _, err := Fix(root, []string{path}, nil, nil); err != nil {
		t.Fatalf("Fix: %v", err)
	}
	out, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixed file: %v", err)
	}
	res, err := Lint(root, nil, nil, nil)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	return string(out), res.ByRule
}

func assertNoFixableViolations(t *testing.T, byRule map[string]int, doc string) {
	t.Helper()
	for _, rule := range fixableRules {
		if n := byRule[rule]; n != 0 {
			t.Errorf("%s still reports %d violation(s) after --fix, doc:\n%s", rule, n, doc)
		}
	}
}

// A list directly below a heading is the shape MD032 reports most in this repo
// (1,769 occurrences). continuesBlock used to ask whether the *current* line
// looked like a construct start, so a line beginning with "- " always counted as
// continuing something and the blank line was never inserted — the fix pass
// reported "nothing to fix" on a file the linter had just listed 20 fixable
// violations for.
func TestFixSeparatesListFromPrecedingHeading(t *testing.T) {
	got, byRule := fixAndLint(t, "# Title\n## Section\n- one\n- two\n\nafter\n")
	want := "# Title\n\n## Section\n\n- one\n- two\n\nafter\n"
	if got != want {
		t.Errorf("fix did not separate the list from the heading\n got: %q\nwant: %q", got, want)
	}
	assertNoFixableViolations(t, byRule, got)
}

// The same suppression hid the MD022 blank line below a heading, because the
// blank line MD022 wants after a heading is the same line MD032 wants before the
// list that follows it.
func TestFixAddsBlankLineBelowHeadingBeforeList(t *testing.T) {
	_, byRule := fixAndLint(t, "# Title\ntext\n## Section\n- one\n")
	assertNoFixableViolations(t, byRule, "## Section\n- one\n")
}

// A list item owns the line that follows it, so no blank line may be inserted
// between an item and its nested item, a blockquote, or a sibling item.
func TestFixKeepsListItemsAttached(t *testing.T) {
	cases := map[string]string{
		"nested bullet":  "- one\n  - nested\n- two\n",
		"nested ordered": "1. one\n   1. nested\n2. two\n",
		"blockquote":     "- one\n> quoted\n",
	}
	for name, doc := range cases {
		t.Run(name, func(t *testing.T) {
			got, byRule := fixAndLint(t, doc)
			if got != doc {
				t.Errorf("fix split a construct that belongs together\n got: %q\nwant: %q", got, doc)
			}
			assertNoFixableViolations(t, byRule, got)
		})
	}
}

// A lazy continuation line looks like ordinary prose, so the fixer cannot tell it
// from a paragraph that ended the list. MD032 reports it (the item is not
// followed by a blank line) and the fixer follows the rule: separating them
// leaves the list intact, it only makes it loose. The alternative — leaving the
// document unchanged — is a fix pass that ends with the violation still there.
func TestFixSeparatesLazyContinuationBecauseTheRuleDoes(t *testing.T) {
	got, byRule := fixAndLint(t, "- one\nstill part of one\n- two\n")
	want := "- one\n\nstill part of one\n\n- two\n"
	if got != want {
		t.Errorf("fix did not separate the lazy continuation\n got: %q\nwant: %q", got, want)
	}
	assertNoFixableViolations(t, byRule, got)
}

// A fenced block owns its body; a list starting after the closing fence is a new
// list and does need the blank line.
func TestFixSeparatesFenceFromFollowingList(t *testing.T) {
	got, byRule := fixAndLint(t, "```go\nx := 1\n```\n- one\n")
	want := "```go\nx := 1\n```\n\n- one\n"
	if got != want {
		t.Errorf("fix did not separate the closing fence from the list\n got: %q\nwant: %q", got, want)
	}
	if strings.Contains(got, "x := 1\n\n") {
		t.Errorf("fix inserted a blank line inside the code block: %q", got)
	}
	assertNoFixableViolations(t, byRule, got)
}

// Tables own their rows; a list after a table is a new block.
func TestFixSeparatesTableFromSurroundingBlocks(t *testing.T) {
	got, byRule := fixAndLint(t, "# Title\n| a | b |\n| --- | --- |\n| 1 | 2 |\n- one\n")
	want := "# Title\n\n| a | b |\n| --- | --- |\n| 1 | 2 |\n\n- one\n"
	if got != want {
		t.Errorf("fix did not separate the table\n got: %q\nwant: %q", got, want)
	}
	assertNoFixableViolations(t, byRule, got)
}

// A setext heading is one construct spanning two lines; splitting it with a
// blank line turns the text into a paragraph.
func TestFixKeepsSetextHeadingIntact(t *testing.T) {
	doc := "Title\n=====\n\ntext\n"
	got, byRule := fixAndLint(t, doc)
	if got != doc {
		t.Errorf("fix split a setext heading\n got: %q\nwant: %q", got, doc)
	}
	assertNoFixableViolations(t, byRule, got)
}

// The real proof: after a fix pass, no fixable rule may still report. "Lint
// after fix is clean" was the check that let a no-op fix ship, because the same
// document was clean before the fix too.
func TestFixLeavesNoFixableViolationBehind(t *testing.T) {
	doc := strings.Join([]string{
		"# Title",
		"## Section",
		"- one",
		"- two",
		"  - nested",
		"",
		"```go",
		"x := 1",
		"- not a list",
		"```",
		"| a | b |",
		"| --- | --- |",
		"| 1 | 2 |",
		"1. first",
		"2. second",
		"### Deep",
		"text with trailing spaces   ",
		"",
		"",
		"",
		"end",
	}, "\n")

	got, byRule := fixAndLint(t, doc)
	assertNoFixableViolations(t, byRule, got)
	if strings.Contains(got, "x := 1\n\n") {
		t.Errorf("a blank line appeared inside the code block:\n%s", got)
	}
	if !strings.Contains(got, "  - nested") {
		t.Errorf("the nested list item lost its indent:\n%s", got)
	}

	// A second pass has nothing left to do: fixing is idempotent.
	root := writeFiles(t, map[string]string{
		".markdownlint.json": minimalConfig,
		"a.md":               got,
	})
	changed, err := Fix(root, []string{filepath.Join(root, "a.md")}, nil, nil)
	if err != nil {
		t.Fatalf("second Fix: %v", err)
	}
	if changed {
		after, _ := os.ReadFile(filepath.Join(root, "a.md"))
		t.Errorf("fix is not idempotent; second pass rewrote the file:\nbefore:\n%s\nafter:\n%s", got, string(after))
	}
}

// Fix's per-file read branch returned collectFiles' `err`, which is
// necessarily nil at that point (the caller's `if err != nil` already
// returned). An unreadable file therefore reported `changed=false, err=nil`,
// and the command printed "[OK] nothing to fix" and exited 0 — abandoning
// every remaining file in the batch, not just the unreadable one. A file
// held open by an editor or a scanner returns EACCES on Windows, and a
// broken symlink is unreadable everywhere.
func TestFixReportsAReadErrorInsteadOfNothingToFix(t *testing.T) {
	root := writeFiles(t, map[string]string{
		".markdownlint.json": minimalConfig,
		"a.md":               "text\n## Heading\n",
		"b.md":               "text\n## Heading\n",
	})
	// A directory that collectFiles walks into but that cannot be read as a
	// file. Windows denies this; elsewhere fall back to a dangling symlink,
	// which is unreadable on every platform.
	unreadable := filepath.Join(root, "locked.md")
	if err := os.Mkdir(unreadable, 0o000); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.Symlink(filepath.Join(root, "does-not-exist"), filepath.Join(root, "z-broken.md")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	changed, err := Fix(root, nil, nil, nil)
	if err == nil {
		t.Fatalf("Fix reported no error for an unreadable file (changed=%v), so the caller prints \"nothing to fix\" and exits 0", changed)
	}
	if !strings.Contains(err.Error(), "z-broken.md") && !strings.Contains(err.Error(), "locked.md") {
		t.Errorf("error %q should name the unreadable file", err)
	}
}
