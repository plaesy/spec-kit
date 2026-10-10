package semantic

import (
	"path/filepath"
	"testing"

	"github.com/plaesy/spec-kit/internal/graph"
)

// TestDocsFor_WiredExtensions is the end-to-end guard on the extension
// dispatch: for every extension docsFor claims to handle, a real file with a
// real doc comment must come back enriched.
//
// The per-language extractor tests all call their extractor directly, and the
// per-language drift tests build a real graph — but neither proves the
// extension is actually routed. An extractor that is written, correct, fully
// tested and never wired looks identical to a working one from inside its own
// test file. This table is what catches that.
//
// Assertions are on the doc *text*, not on symbol names: nodeText is the only
// consumer that cares about key format, and it does its own lookup. What
// matters here is that the right file reached the right extractor.
//
// When a new language is added, add a row. A row for an extension that is
// implemented but not wired fails here, which is the point.
func TestDocsFor_WiredExtensions(t *testing.T) {
	cases := []struct {
		ext     string
		content string
		want    string // doc text that must appear among the returned values
	}{
		// A .go file needs a package clause or go/parser rejects it outright,
		// which is the one language here that parses rather than regexes.
		{ext: ".go", content: "package lib\n\n// Target docs.\nfunc Target() {}\n", want: "Target docs."},
		{ext: ".js", content: "/** Target docs. */\nfunction target() {}\n", want: "Target docs."},
		{ext: ".jsx", content: "/** Target docs. */\nfunction target() {}\n", want: "Target docs."},
		{ext: ".ts", content: "/** Target docs. */\nfunction target() {}\n", want: "Target docs."},
		{ext: ".tsx", content: "/** Target docs. */\nfunction target() {}\n", want: "Target docs."},
		{ext: ".py", content: "def target():\n    \"\"\"Target docs.\"\"\"\n    return 1\n", want: "Target docs."},
		{ext: ".rs", content: "/// Target docs.\nfn target() {}\n", want: "Target docs."},
		{ext: ".rb", content: "# Target docs.\ndef target\nend\n", want: "Target docs."},
		{ext: ".sh", content: "# Target docs.\ntarget() {\n  :\n}\n", want: "Target docs."},
		{ext: ".ps1", content: "# Target docs.\nfunction target {\n}\n", want: "Target docs."},

		// Wave 2: the remaining languages.
		{ext: ".dart", content: "/// Target docs.\nclass Target {}\n", want: "Target docs."},
		{ext: ".java", content: "/** Target docs. */\nclass Target {}\n", want: "Target docs."},
		{ext: ".kt", content: "/** Target docs. */\nclass Target\n", want: "Target docs."},
		{ext: ".kts", content: "/** Target docs. */\nclass Target\n", want: "Target docs."},
		{ext: ".cs", content: "/// Target docs.\nclass Target { }\n", want: "Target docs."},
		{ext: ".swift", content: "/// Target docs.\nclass Target {}\n", want: "Target docs."},
		{ext: ".php", content: "<?php\n/** Target docs. */\nfunction target() {}\n", want: "Target docs."},
		// C-family: every extension symbols.go routes to cFamilySymbols. All seven use
		// the same definition-with-a-body, because symbols.go extracts nothing from a
		// bare prototype (`int add(int a, int b);`) and nothing from `(void)`, so
		// neither gives any extractor a symbol to key a doc comment to. Headers
		// included — a .h file with a body parses the same as a .c file. Verified
		// against graph.Build rather than assumed.
		{ext: ".c", content: "/* Target docs. */\nint add(int a, int b) {\n  return a + b;\n}\n", want: "Target docs."},
		{ext: ".h", content: "/* Target docs. */\nint add(int a, int b) {\n  return a + b;\n}\n", want: "Target docs."},
		{ext: ".cpp", content: "/* Target docs. */\nint add(int a, int b) {\n  return a + b;\n}\n", want: "Target docs."},
		{ext: ".hpp", content: "/* Target docs. */\nint add(int a, int b) {\n  return a + b;\n}\n", want: "Target docs."},
		{ext: ".cc", content: "/* Target docs. */\nint add(int a, int b) {\n  return a + b;\n}\n", want: "Target docs."},
		{ext: ".hh", content: "/* Target docs. */\nint add(int a, int b) {\n  return a + b;\n}\n", want: "Target docs."},
		{ext: ".cxx", content: "/* Target docs. */\nint add(int a, int b) {\n  return a + b;\n}\n", want: "Target docs."},
	}

	for _, tc := range cases {
		t.Run("ext="+tc.ext, func(t *testing.T) {
			repoRoot := t.TempDir()
			writeFile(t, filepath.Join(repoRoot, "src", "lib"+tc.ext), tc.content)

			docs := docsFor(repoRoot, graph.Node{ID: "src/lib" + tc.ext})
			if docs == nil {
				t.Fatalf("docsFor(.%s) = nil; the extension is either not wired into docsFor or the "+
					"extractor found nothing", tc.ext)
			}
			for _, got := range docs {
				if got == tc.want {
					return
				}
			}
			t.Errorf("docsFor(.%s) = %v, want some value to be %q — the extension is wired but the "+
				"extractor did not pick up the doc comment", tc.ext, docs, tc.want)
		})
	}
}

// TestDocsFor_MissingFileReturnsNil: a wired extension whose file cannot be
// read must return nil, not panic and not a partial map. Doc comments are
// best-effort enrichment and must never fail indexing.
func TestDocsFor_MissingFileReturnsNil(t *testing.T) {
	repoRoot := t.TempDir()
	for _, ext := range []string{".go", ".js", ".py", ".rs", ".rb", ".sh", ".ps1"} {
		if docs := docsFor(repoRoot, graph.Node{ID: "src/absent" + ext}); docs != nil {
			t.Errorf("docsFor(missing .%s) = %v, want nil", ext, docs)
		}
	}
}
