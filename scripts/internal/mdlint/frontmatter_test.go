package mdlint

import "testing"

// goldmark was configured with extension.GFM only, which does not include
// front matter. The YAML block was therefore parsed as ordinary body markdown
// and every `#` comment inside it became a real ast.Heading. checkMD025 counts
// headings by walking the AST, so those comments were reported as H1s -- 9
// MD025 violations in templates/design.template.md, all of them YAML comments.
// There is no fix for that from inside a template: `#` is the only comment
// sigil YAML has, and removing it corrupts the front matter.
//
// Each case below is written to fail if the rule is *disabled* as well as if it
// misfires. A linter that silently reports nothing looks identical to a correct
// one in a green run, so "no violations" alone would not have caught this.
func TestMD025IgnoresFrontMatterComments(t *testing.T) {
	cases := []struct {
		name   string
		body   string
		byRule int
	}{
		{
			// The original bug: three YAML comments, one real heading.
			name: "front matter comments are not h1",
			body: "---\n# Design System\n# Author: someone\npalette: 1\n---\n\n# Real Title\n\ntext\n",
		},
		{
			// Proves the rule is still live: two real h1s in the body are
			// still reported, so the fix did not just silence MD025.
			name:   "two real h1s in body are still reported",
			body:   "---\n# comment one\n# comment two\n---\n\n# First\n\ntext\n\n# Second\n\ntext\n",
			byRule: 1,
		},
		{
			// Proves the front matter is skipped rather than consuming the
			// document: with the real heading removed, the body genuinely has
			// no h1 and must be reported.
			name:   "missing h1 in body is still reported",
			body:   "---\n# comment\n---\n\n## Only An h2\n\ntext\n",
			byRule: 1,
		},
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
			if got := res.ByRule["MD025"]; got != tc.byRule {
				t.Errorf("MD025 = %d, want %d (by file: %+v)", got, tc.byRule, res.ByFile)
			}
		})
	}
}
