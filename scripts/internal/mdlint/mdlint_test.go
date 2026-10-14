package mdlint

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// writeFiles materializes a map of relative path -> content in a temp dir.
func writeFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, content := range files {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", rel, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
	return root
}

const minimalConfig = `{"default": true}`

func TestLintFlagsMissingBlankLineAroundHeading(t *testing.T) {
	root := writeFiles(t, map[string]string{
		".markdownlint.json": minimalConfig,
		"a.md":               "text\n## Heading\n",
	})
	res, err := Lint(root, nil, nil, nil)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if res.ByRule["MD022"] == 0 {
		t.Errorf("expected MD022, got %+v", res.ByRule)
	}
	if res.Files != 1 {
		t.Errorf("Files = %d, want 1 (config file is not markdown)", res.Files)
	}
}

// MD040 fired on the closing fence of every block in the repository, because
// the check treated any ``` line as an opening fence — and a closing fence has
// no info string, so it read as "no language". 1,092 of the baseline's
// violations were this. The rule has no test at all, which is how it survived.
// MD031 asked for a blank line after the *opening* fence — that is, inside the
// code block — and `--fix` obliged by inserting one. The post-fix lint was then
// clean, so the wrong finding never showed up: the bogus report was satisfied by
// a real edit to the document's content. Testing only "lint after fix" is what
// hid it, so this checks the rule against the text as written.
// MD029's "ordered" style compared each item's number with its *line number*
// (`num != i+1`), so every correctly numbered list that did not start on line 1
// was reported — 987 of the repository's violations, and most of the ordered
// lists in it. It also read "one_or_ordered" as "all ones", rejecting the
// incrementing half of its own name.
func TestMD029OrderedStyleCountsItemsNotLines(t *testing.T) {
	root := writeFiles(t, map[string]string{
		".markdownlint.json": `{"default": true, "MD029": {"style": "ordered"}}`,
		"a.md": "# Title\n\nSome prose first, so the list does not begin on line 1.\n\n" +
			"1. First\n   wrapped onto a second line\n2. Second\n3. Third\n",
	})
	res, err := Lint(root, nil, nil, nil)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if n := res.ByRule["MD029"]; n != 0 {
		t.Errorf("MD029 = %d, want 0 for a 1,2,3 list: the number is compared with the item's "+
			"position in the list, not its line in the file", n)
	}
}

func TestMD029OrderedStyleFlagsASkippedNumber(t *testing.T) {
	root := writeFiles(t, map[string]string{
		".markdownlint.json": `{"default": true, "MD029": {"style": "ordered"}}`,
		"a.md":               "# Title\n\n1. First\n2. Second\n4. Fourth\n",
	})
	res, err := Lint(root, nil, nil, nil)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if n := res.ByRule["MD029"]; n != 1 {
		t.Fatalf("MD029 = %d, want 1: only the item numbered 4 skips a number", n)
	}
	if line := res.Violations[0].Detail[0].Line; line != 5 {
		t.Errorf("the finding must point at the item numbered 4 (line 5), got line %d", line)
	}
}

func TestMD029OrderedStyleRestartsAfterAnInterruption(t *testing.T) {
	root := writeFiles(t, map[string]string{
		".markdownlint.json": `{"default": true, "MD029": {"style": "ordered"}}`,
		"a.md":               "# Title\n\n1. First\n2. Second\n\nA paragraph between them.\n\n1. New list\n2. New list\n",
	})
	res, err := Lint(root, nil, nil, nil)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if n := res.ByRule["MD029"]; n != 0 {
		t.Errorf("MD029 = %d, want 0: a list interrupted by a paragraph starts a new sequence", n)
	}
}

func TestMD029ToleratesBlankLinesAndNestedLists(t *testing.T) {
	root := writeFiles(t, map[string]string{
		".markdownlint.json": `{"default": true, "MD029": {"style": "ordered"}}`,
		"a.md": "# Title\n\n1. First\n\n   More about the first item.\n\n2. Second\n" +
			"   1. Nested one\n   2. Nested two\n3. Third\n",
	})
	res, err := Lint(root, nil, nil, nil)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if n := res.ByRule["MD029"]; n != 0 {
		t.Errorf("MD029 = %d, want 0: blank lines keep a list together and a nested list has its "+
			"own sequence", n)
	}
}

func TestMD029OneOrOrderedAcceptsBothShapes(t *testing.T) {
	cfg := `{"default": true, "MD029": {"style": "one_or_ordered"}}`
	for name, body := range map[string]string{
		"all ones":       "# Title\n\n1. a\n1. b\n1. c\n",
		"incrementing":   "# Title\n\n1. a\n2. b\n3. c\n",
		"loose all ones": "# Title\n\n1. a\n\n1. b\n\n1. c\n",
	} {
		root := writeFiles(t, map[string]string{".markdownlint.json": cfg, "a.md": body})
		res, err := Lint(root, nil, nil, nil)
		if err != nil {
			t.Fatalf("%s: Lint: %v", name, err)
		}
		if n := res.ByRule["MD029"]; n != 0 {
			t.Errorf("%s: MD029 = %d, want 0: the style is named one_or_ordered", name, n)
		}
	}
	root := writeFiles(t, map[string]string{".markdownlint.json": cfg, "a.md": "# Title\n\n1. a\n3. c\n"})
	res, err := Lint(root, nil, nil, nil)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if n := res.ByRule["MD029"]; n == 0 {
		t.Error("a list that is neither all-ones nor incrementing must still be reported")
	}
}

func TestMD031DoesNotAskForABlankLineInsideACodeBlock(t *testing.T) {
	root := writeFiles(t, map[string]string{
		".markdownlint.json": minimalConfig,
		"a.md":               "# Title\n\n```go\nfmt.Println()\n```\ntext right after\n",
	})
	res, err := Lint(root, nil, nil, nil)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if n := res.ByRule["MD031"]; n != 1 {
		t.Errorf("MD031 = %d, want 1: only the missing blank line after the closing fence. "+
			"A blank line inside the block would be an edit to the code", n)
	}
}

func TestMD031AcceptsAnEmptyCodeBlock(t *testing.T) {
	root := writeFiles(t, map[string]string{
		".markdownlint.json": minimalConfig,
		"a.md":               "# Title\n\n```\n```\n\nprose\n",
	})
	res, err := Lint(root, nil, nil, nil)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if n := res.ByRule["MD031"]; n != 0 {
		t.Errorf("MD031 = %d, want 0 for an empty block with its surrounding blank lines", n)
	}
}

// Four or more leading spaces mark an indented code block, which cannot open
// or close a fence — but content indented that deeply *inside* a fence is
// still that block's content. markFences skipped the code marking for such
// lines, so the line before a closing fence stopped looking like code and
// MD031 reported the fence itself as missing a blank line above it. The only
// way to have silenced that finding was to edit the code the report was
// about, which is the failure mode the MD040 and MD031 comments above this
// package describe.
func TestIndentedContentInsideAFenceIsStillCode(t *testing.T) {
	root := writeFiles(t, map[string]string{
		".markdownlint.json": minimalConfig,
		"a.md":               "# Title\n\n```text\nBAD:  one\nGOOD: two\n      three\n```\n\nprose\n",
	})
	res, err := Lint(root, nil, nil, nil)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if n := res.ByRule["MD031"]; n != 0 {
		t.Errorf("MD031 = %d, want 0: the line above the closing fence is code, "+
			"however deeply it is indented", n)
	}
}

// The other half of that branch: a 4-space block that is *not* inside a fence
// is an indented code block. Marking it as code is what keeps MD031 and MD040
// from reading its content as prose, so both halves of the indent>3 branch
// need to be exercised.
func TestIndentedCodeBlockOutsideAFenceIsCode(t *testing.T) {
	root := writeFiles(t, map[string]string{
		".markdownlint.json": minimalConfig,
		"a.md":               "# Title\n\nprose\n\n    still prose to a rule, but\n    indented code to CommonMark\n\nmore\n",
	})
	res, err := Lint(root, nil, nil, nil)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	for _, rule := range []string{"MD031", "MD040", "MD010"} {
		if n := res.ByRule[rule]; n != 0 {
			t.Errorf("%s = %d, want 0: an indented code block is code, not a fence", rule, n)
		}
	}
}

// siblings_only means "compare headings that share a parent", not "compare
// headings of the same level". The key used to be level plus text, so every
// `### Subcommands` in the document collided with every other one — the rule
// fired on docs/reference.md, which documents a dozen commands that each have
// their own subcommand table.
func TestMD024SiblingsOnlyComparesHeadingsUnderTheSameParent(t *testing.T) {
	body := "# Root\n\n## alpha\n\n### Subcommands\n\nx\n\n## beta\n\n### Subcommands\n\ny\n\n## alpha\n\n### Subcommands\n\nz\n"
	root := writeFiles(t, map[string]string{".markdownlint.json": minimalConfig, "a.md": body})
	res, err := Lint(root, nil, nil, nil)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	// Exactly two: the repeated `## alpha`, and the repeated `### Subcommands`
	// underneath that second alpha. The `### Subcommands` under beta is a
	// different sibling set and must not be reported.
	if n := res.ByRule["MD024"]; n != 2 {
		t.Errorf("MD024 = %d, want 2: only the repeated heading and its child", n)
	}
}

func TestMD036ReportsOnlyAParagraphThatIsNothingButEmphasis(t *testing.T) {
	// The rule asks whether a *paragraph* is only emphasis, not whether a line
	// happens to hold nothing else. Prose above the emphasis, or a list item it
	// lazily continues, makes it ordinary text — and the `--fix` passes insert
	// exactly those adjacencies, so a line-based reading manufactures reports
	// out of its own output.
	cases := []struct {
		name string
		body string
		want int
	}{
		{"standalone paragraph", "# T\n\n*Emphasis*\n\nprose\n", 1},
		{"second line of a paragraph", "Intro line.\n*Emphasis*\n\nprose\n", 0},
		{"lazy continuation of a list item", "# T\n\n- item\n*Emphasis*\n\nprose\n", 0},
		{"directly under a heading", "# T\n*Emphasis*\n\nprose\n", 1},
		{"directly under a blockquote", "# T\n\n> quote\n*Emphasis*\n\nprose\n", 1},
		{"no blank line after it", "# T\n\n*Emphasis*\nprose\n", 0},
		{"emphasis wrapping a longer sentence", "# T\n\n**FR-001: User authentication**\n\nprose\n", 1},
		{"a list item's label with text after it", "# T\n\n- **Priority:** high\n\nprose\n", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := writeFiles(t, map[string]string{
				".markdownlint.json": minimalConfig,
				"a.md":               tc.body,
			})
			res, err := Lint(root, nil, nil, nil)
			if err != nil {
				t.Fatalf("Lint: %v", err)
			}
			if n := res.ByRule["MD036"]; n != tc.want {
				t.Errorf("MD036 = %d, want %d for:\n%s", n, tc.want, tc.body)
			}
		})
	}
}

func TestMD040IgnoresClosingFences(t *testing.T) {
	root := writeFiles(t, map[string]string{
		".markdownlint.json": minimalConfig,
		"a.md":               "# Title\n\n```go\nfmt.Println()\n```\n\n~~~\nplain\n~~~\n",
	})
	res, err := Lint(root, nil, nil, nil)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if n := res.ByRule["MD040"]; n != 1 {
		t.Errorf("MD040 = %d, want 1: the ```go block declares a language and the ~~~ block does not; "+
			"closing fences are not openings", n)
	}
}

func TestMD040AcceptsTildeFencesAndInfoStrings(t *testing.T) {
	root := writeFiles(t, map[string]string{
		".markdownlint.json": `{"default": true, "MD040": {"language_only": false}}`,
		"a.md":               "# Title\n\n```go title=\"main.go\"\nfmt.Println()\n```\n\n~~~sh\nls\n~~~\n",
	})
	res, err := Lint(root, nil, nil, nil)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if n := res.ByRule["MD040"]; n != 0 {
		t.Errorf("MD040 = %d, want 0: an info string after the language is allowed with "+
			"language_only:false", n)
	}
}

func TestLintRespectsRuleOptions(t *testing.T) {
	// 100-char line: violates the 80-char default, passes with line_length 120.
	long := "# Title\n\n" + strings.Repeat("x", 100) + "\n"
	root := writeFiles(t, map[string]string{
		".markdownlint.json": `{"default": true, "MD013": {"line_length": 120}}`,
		"a.md":               long,
	})
	res, err := Lint(root, nil, nil, nil)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if res.ByRule["MD013"] != 0 {
		t.Errorf("MD013 should be satisfied at 120 chars, got %d", res.ByRule["MD013"])
	}

	// A tighter limit must flag it.
	root2 := writeFiles(t, map[string]string{
		".markdownlint.json": `{"default": true, "MD013": {"line_length": 40}}`,
		"a.md":               long,
	})
	res2, err := Lint(root2, nil, nil, nil)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if res2.ByRule["MD013"] == 0 {
		t.Errorf("MD013 at 40 chars should flag the 100-char line, got %+v", res2.ByRule)
	}
}

func TestLintDisabledRuleIsOff(t *testing.T) {
	root := writeFiles(t, map[string]string{
		".markdownlint.json": `{"default": true, "MD022": false}`,
		"a.md":               "text\n## Heading\n",
	})
	res, err := Lint(root, nil, nil, nil)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if res.ByRule["MD022"] != 0 {
		t.Errorf("MD022 disabled but reported %d", res.ByRule["MD022"])
	}
}

func TestLintAcceptsAliasKeys(t *testing.T) {
	// "ul-style" is markdownlint's alias for MD004; both spellings must work.
	root := writeFiles(t, map[string]string{
		".markdownlint.json": `{"default": true, "ul-style": {"style": "dash"}}`,
		"a.md":               "- one\n- two\n",
	})
	res, err := Lint(root, nil, nil, nil)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if res.ByRule["MD004"] != 0 {
		t.Errorf("dash list should satisfy MD004, got %d", res.ByRule["MD004"])
	}
}

func TestLintRejectsUnknownRuleOption(t *testing.T) {
	// A config key the linter does not implement must fail loudly rather than
	// be ignored (QG-01: a check that cannot run must not report as passing).
	root := writeFiles(t, map[string]string{
		".markdownlint.json": `{"default": true, "MD013": {"line_lenght": 120}}`,
		"a.md":               "# Title\n",
	})
	if _, err := Lint(root, nil, nil, nil); err == nil || !strings.Contains(err.Error(), "line_lenght") {
		t.Fatalf("expected unknown-option error naming line_lenght, got %v", err)
	}
}

func TestLintRejectsNpmExtends(t *testing.T) {
	root := writeFiles(t, map[string]string{
		".markdownlint.json": `{"extends": "@github/markdownlint-github"}`,
		"a.md":               "# Title\n",
	})
	_, err := Lint(root, nil, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "extends") {
		t.Fatalf("expected a clear error for npm 'extends', got %v", err)
	}
}

func TestLintSkipsIgnoredDirs(t *testing.T) {
	root := writeFiles(t, map[string]string{
		".markdownlint.json":    minimalConfig,
		"a.md":                  "text\n## Heading\n",
		"node_modules/pkg/x.md": "text\n## Heading\n",
		".plaesy/memory/y.md":   "text\n## Heading\n",
		".claude/commands/a.md": "text\n## Heading\n",
		".kilo/commands/a.md":   "text\n## Heading\n",
	})
	res, err := Lint(root, nil, nil, nil)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if res.Scanned != 1 {
		t.Errorf("Scanned = %d, want 1 (every ignored dir skipped)", res.Scanned)
	}
	for _, f := range res.WorstFiles(0) {
		for _, ignored := range DefaultIgnoreDirs {
			if strings.Contains(f, ignored) {
				t.Errorf("ignored dir %q leaked into results via %q", ignored, f)
			}
		}
	}
}

func TestLintGitHubAltTextRules(t *testing.T) {
	root := writeFiles(t, map[string]string{
		".markdownlint.json": `{"default": true, "MD036": false}`,
		"a.md": "# Title\n\n![](/img/a.png)\n\n![](/img/logo.png)\n\n" +
			"![diagram](/img/diagram.png)\n\n" +
			"[click here](/docs)\n",
	})
	res, err := Lint(root, nil, nil, nil)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if res.ByRule["GHA001"] == 0 {
		t.Errorf("empty alt text not flagged: %+v", res.ByRule)
	}
	if res.ByRule["GHA003"] == 0 {
		t.Errorf("generic link text not flagged: %+v", res.ByRule)
	}
	if res.ByRule["GHA002"] == 0 {
		t.Errorf("default alt text (the file name) not flagged: %+v", res.ByRule)
	}
}

func TestLintCleanFileHasNoViolations(t *testing.T) {
	root := writeFiles(t, map[string]string{
		".markdownlint.json": `{"default": true, "MD013": {"line_length": 120, "code_blocks": false}}`,
		"a.md":               "# Title\n\nSome text.\n\n## Section\n\nMore text.\n",
	})
	res, err := Lint(root, nil, nil, nil)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if res.Total != 0 {
		t.Errorf("expected a clean file, got %d violations: %s", res.Total, res.Histogram())
	}
}

func TestResultExceedsBaseline(t *testing.T) {
	res := &Result{Total: 12}
	if res.Exceeds(&Baseline{MaxViolations: 12}) {
		t.Error("12 against a baseline of 12 must pass (ratchet, not zero)")
	}
	if !res.Exceeds(&Baseline{MaxViolations: 11}) {
		t.Error("12 against a baseline of 11 must fail")
	}
	// A nil baseline is `--no-baseline`: a ceiling of zero, not the absence of a
	// ceiling. Reading it as "nothing to compare" made every --no-baseline run
	// report success while listing its own violations.
	if !res.Exceeds(nil) {
		t.Error("12 violations against no baseline must fail — the flag promises zero")
	}
	if (&Result{}).Exceeds(nil) {
		t.Error("a clean run must pass with no baseline")
	}
}

func TestLoadBaselineRequiresPositiveCeiling(t *testing.T) {
	root := writeFiles(t, map[string]string{
		"b.json": `{"max_violations": 0, "measured_at": "2026-09-25"}`,
	})
	if _, err := LoadBaseline(filepath.Join(root, "b.json")); err == nil {
		t.Fatal("a zero ceiling must be rejected: pass no baseline file to require zero")
	}
}

// --fix used to add a blank line after the *opening* fence as well as after the
// closing one, because the fixer matched every fence delimiter while the rule
// only ever requires a blank line before an opening fence and after a closing
// one. The result was a blank line inserted into the middle of the code block —
// an edit to the document's content that no violation asked for.
func TestFixLeavesCodeBlockBodiesAlone(t *testing.T) {
	root := writeFiles(t, map[string]string{
		".markdownlint.json": `{"default": true}`,
		"a.md":               "# Title\n\n```go\nfmt.Println()\n```\ntext right after\n",
	})
	if _, err := Fix(root, []string{filepath.Join(root, "a.md")}, nil, nil); err != nil {
		t.Fatalf("Fix: %v", err)
	}
	out, err := os.ReadFile(filepath.Join(root, "a.md"))
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	if !strings.Contains(got, "```go\nfmt.Println()\n```\n") {
		t.Errorf("the code block body must survive untouched, got:\n%s", got)
	}
	if !strings.Contains(got, "```\n\ntext right after") {
		t.Errorf("a blank line after the closing fence is the fix that was asked for, got:\n%s", got)
	}
}

func TestFixAppliesMachineFixableViolations(t *testing.T) {
	root := writeFiles(t, map[string]string{
		".markdownlint.json": `{"default": true}`,
		"a.md":               "#   Title\n\ntext\t\n",
	})
	changed, err := Fix(root, []string{filepath.Join(root, "a.md")}, nil, nil)
	if err != nil {
		t.Fatalf("Fix: %v", err)
	}
	out, readErr := os.ReadFile(filepath.Join(root, "a.md"))
	if readErr != nil {
		t.Fatalf("read fixed file: %v", readErr)
	}
	if !changed {
		t.Logf("fix reported no change; file now: %q", string(out))
	}
	if strings.Contains(string(out), "text\t") {
		t.Errorf("trailing tab not fixed: %q", string(out))
	}
}

// A setext heading spans the whole paragraph above its underline, so the
// heading is anchored at the paragraph's first line with the paragraph's full
// text. Anchoring at the line above the underline instead turned the extremely
// common "paragraph, then --- separator" shape into a phantom heading and cost
// 941 false positives repo-wide when it was fixed (MD022 2,840 -> 1,972,
// MD026 122 -> 29).
func TestLintTreatsMultiLineSetextAsOneHeading(t *testing.T) {
	root := writeFiles(t, map[string]string{
		".markdownlint.json": minimalConfig,
		// "---" after prose is a setext H2 whose text is both lines, so only the
		// blank line above line 1 and the blank line below the underline matter.
		"a.md": "# Title\n\nprose line one\nline two continues here\n---\n\nafter\n",
	})
	res, err := Lint(root, nil, nil, nil)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	for _, rule := range []string{"MD022", "MD026", "MD024", "MD025", "MD001"} {
		if n := res.ByRule[rule]; n != 0 {
			for _, f := range res.Violations {
				for _, v := range f.Detail {
					if v.Rule == rule {
						t.Logf("%s at line %d: %s", rule, v.Line, v.Message)
					}
				}
			}
			t.Errorf("%s = %d, want 0: continuation prose must not be read as a heading", rule, n)
		}
	}
}

// A "---" that follows a list, a quote, or a fence is a thematic break, not a
// setext underline: the block it would underline does not exist.
func TestLintThematicBreakAfterListIsNotAHeading(t *testing.T) {
	root := writeFiles(t, map[string]string{
		".markdownlint.json": minimalConfig,
		"a.md":               "# Title\n\n- item\n\n---\n\nafter\n",
	})
	res, err := Lint(root, nil, nil, nil)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if res.ByRule["MD025"] != 0 {
		t.Errorf("MD025 = %d, want 0: --- after a list is a thematic break", res.ByRule["MD025"])
	}
}

// A list item that follows a wrapped item is not a new list and needs no blank
// line above it. Reading the indent from the trimmed line made every wrapped
// item look like the end of a list, which cost 303 false positives repo-wide
// (MD032 2,074 -> 1,771).
func TestLintAcceptsWrappedListItemFollowedBySibling(t *testing.T) {
	root := writeFiles(t, map[string]string{
		".markdownlint.json": minimalConfig,
		"a.md":               "# Title\n\n- first item that is long enough\n  to wrap onto a second line\n- sibling item\n- another sibling\n\nafter\n",
	})
	res, err := Lint(root, nil, nil, nil)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if n := res.ByRule["MD032"]; n != 0 {
		for _, f := range res.Violations {
			for _, v := range f.Detail {
				t.Logf("MD032 at line %d: %s", v.Line, v.Message)
			}
		}
		t.Errorf("MD032 = %d, want 0: a sibling of a wrapped item continues the same list", n)
	}
}

// A heading that is entirely inline markup (`### `target“) has no direct text
// segment in goldmark's tree. Reading only direct children made every such
// heading look like the same empty heading, so unrelated headings collided and
// the message named nothing: "duplicate heading """.
func TestLintDistinguishesCodeSpanHeadings(t *testing.T) {
	root := writeFiles(t, map[string]string{
		".markdownlint.json": `{"default": true, "MD013": false}`,
		"a.md":               "# Title\n\n## `alpha`\n\ntext\n\n## `beta`\n\ntext\n",
	})
	res, err := Lint(root, nil, nil, nil)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if n := res.ByRule["MD024"]; n != 0 {
		for _, f := range res.Violations {
			for _, v := range f.Detail {
				t.Logf("MD024 at line %d: %s", v.Line, v.Message)
			}
		}
		t.Errorf("MD024 = %d, want 0: two different code-span headings are not duplicates", n)
	}
}

func TestLintReportsDuplicateCodeSpanHeadingText(t *testing.T) {
	root := writeFiles(t, map[string]string{
		".markdownlint.json": `{"default": true, "MD013": false}`,
		"a.md":               "# Title\n\n## `alpha`\n\ntext\n\n## `alpha`\n\ntext\n",
	})
	res, err := Lint(root, nil, nil, nil)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if res.ByRule["MD024"] != 1 {
		t.Fatalf("MD024 = %d, want 1: the same heading twice is a duplicate", res.ByRule["MD024"])
	}
	found := false
	for _, f := range res.Violations {
		for _, v := range f.Detail {
			if v.Rule == "MD024" && strings.Contains(v.Message, "alpha") {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("the duplicate message must name the heading, not an empty string: %+v", res.Violations[0].Detail)
	}
}

// The fix passes return a new slice instead of mutating in place. If the line
// masks are not rebuilt from that new slice, they keep their original length
// while the loops walk a longer one, and the first direct index into a mask runs
// off the end. `--fix` panicked on the first real file in this repository; the
// unit tests missed it because their fixtures are too short for a rewrite pass
// to grow the document past the stale mask. This one is long enough.
func TestFixSurvivesGrowingTheDocument(t *testing.T) {
	var b strings.Builder
	b.WriteString("# Title\n\n")
	for i := 0; i < 40; i++ {
		b.WriteString("## Section ")
		b.WriteString(strconv.Itoa(i))
		b.WriteString("\ntext under the heading\n\n")
		b.WriteString("```bash\necho hi\n```\n- item one\n- item two\n\nmore prose\n\n")
	}
	root := writeFiles(t, map[string]string{
		".markdownlint.json": minimalConfig,
		"a.md":               b.String(),
	})
	changed, err := Fix(root, nil, nil, nil)
	if err != nil {
		t.Fatalf("Fix: %v", err)
	}
	if !changed {
		t.Fatal("expected the document to be rewritten")
	}
	// A pass that inserts lines must leave a document the checker can still read.
	res, err := Lint(root, nil, nil, nil)
	if err != nil {
		t.Fatalf("Lint after Fix: %v", err)
	}
	for _, rule := range []string{"MD022", "MD031", "MD032", "MD012"} {
		if n := res.ByRule[rule]; n != 0 {
			t.Errorf("%s = %d after --fix: the fix pass should have resolved these", rule, n)
		}
	}
}

// A .markdownlint.json must stay usable as-is by anyone who has one, so the
// option keys are markdownlint's. `include_code_blocks` was a name of our own
// invention and rejected every real config that set MD010's `code_blocks`.
func TestOptionKeysAreMarkdownlints(t *testing.T) {
	root := writeFiles(t, map[string]string{
		".markdownlint.json": `{
  "default": true,
  "MD009": { "br_spaces": 2, "list_item_empty_lines": false },
  "MD010": { "code_blocks": false, "spaces_per_tab": 4 },
  "MD012": { "maximum": 1 },
  "MD013": { "line_length": 120, "code_blocks": false, "tables": false, "headings": false },
  "MD022": { "lines_above": 1, "lines_below": 1 },
  "MD024": { "siblings_only": true },
  "MD025": { "level": 1 },
  "MD026": { "punctuation": ".,;:" },
  "MD029": { "style": "one" },
  "MD031": { "list_items": true },
  "MD032": { "lists": true },
  "MD033": false,
  "MD040": { "allowed_languages": ["bash", "go"], "language_only": false },
  "MD041": { "front_matter_title": "^[ \t]*title[ \t]*[:=]" },
  "MD046": { "style": "fenced" },
  "MD048": { "style": "backtick" },
  "MD058": { "tables": true }
}`,
		"a.md": "# Title\n\n```bash\n\techo indented with a tab\n```\n",
	})
	cfg, err := LoadConfig(filepath.Join(root, ".markdownlint.json"))
	if err != nil {
		t.Fatalf("a markdownlint-shaped config must load: %v", err)
	}
	res, err := Lint(root, nil, cfg, nil)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if n := res.ByRule["MD010"]; n != 0 {
		t.Errorf("MD010 = %d, want 0: code_blocks:false must exempt fenced code", n)
	}
}

// A rule this linter does not implement, set to false, is already off. Loading
// the config anyway is what lets someone keep the .markdownlint.json they
// already have instead of rewriting it to match our gaps. Enabling or
// configuring an unimplemented rule is still an error, because that asks for a
// check that would silently not happen.
func TestUnimplementedRuleIsAllowedOnlyWhenDisabled(t *testing.T) {
	base := `{"default": true, "MD033": %s}`
	dir := t.TempDir()

	// The ruleset is compiled from the loaded config, so the check under test
	// lives on that path rather than in LoadConfig.
	compileFile := func(body string) error {
		path := filepath.Join(dir, "cfg.json")
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		cfg, err := LoadConfig(path)
		if err != nil {
			return err
		}
		_, err = compile(cfg)
		return err
	}
	if err := compileFile(fmt.Sprintf(base, "false")); err != nil {
		t.Errorf("an unimplemented rule switched off must load: %v", err)
	}
	if err := compileFile(fmt.Sprintf(base, "true")); err == nil {
		t.Error("an unimplemented rule switched on must be rejected, not ignored")
	}
	if err := compileFile(`{"default": true, "MD033": {"allowed_elements": ["b"]}}`); err == nil {
		t.Error("a configured unimplemented rule must be rejected, not ignored")
	}
}

// `.claude` and `.kilo` hold a generated copy of every file in `prompts/`, so
// walking them counted each prompt violation once per platform directory that
// happened to exist on the machine. The total then depended on local state, and
// the ratchet recorded at HEAD was already stale on a clean checkout of the very
// commit that wrote it. The copies must not be linted; the sources still are.
func TestLintDoesNotCountGeneratedPlatformCopies(t *testing.T) {
	const body = "text\n## Heading\n"

	// Measure the source on its own, so the expected total comes from the file
	// rather than from a hardcoded count of its violations.
	sourceOnly := writeFiles(t, map[string]string{
		".markdownlint.json": minimalConfig,
		"prompts/a.md":       body,
	})
	want, err := Lint(sourceOnly, nil, nil, nil)
	if err != nil {
		t.Fatalf("Lint source only: %v", err)
	}
	if want.Total == 0 {
		t.Fatal("fixture is wrong: the source has no violations to compare")
	}

	withCopies := writeFiles(t, map[string]string{
		".markdownlint.json":    minimalConfig,
		"prompts/a.md":          body,
		".claude/commands/a.md": body,
		".kilo/commands/a.md":   body,
	})
	got, err := Lint(withCopies, nil, nil, nil)
	if err != nil {
		t.Fatalf("Lint with copies: %v", err)
	}

	if got.Scanned != want.Scanned {
		t.Errorf("Scanned = %d, want %d: the two generated copies were walked", got.Scanned, want.Scanned)
	}
	if got.Total != want.Total {
		t.Errorf("Total = %d, want %d: adding a generated copy per platform changed the count, "+
			"so the ratchet total depends on which platforms a machine has initialised",
			got.Total, want.Total)
	}
	for path := range got.ByFile {
		if strings.Contains(path, ".claude") || strings.Contains(path, ".kilo") {
			t.Errorf("generated copy %q was linted", path)
		}
	}
}
