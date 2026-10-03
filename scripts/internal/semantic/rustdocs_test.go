package semantic

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/plaesy/spec-kit/internal/graph"
)

// rustDriftFixture mixes the shapes rustSymbols recognises with the rustdoc
// forms that document them: a plain fn, a `pub` fn, an `async` fn, and a
// struct, each documented with "///", plus a `//!` crate header that must NOT be
// attributed to the first item, a "////" divider that is an ordinary comment,
// and an undocumented fn.
const rustDriftFixture = `//! Crate-level docs. These are inner doc comments: they
//! document the crate, not the item below them.

use std::collections::HashMap;

/// A parsed configuration.
#[derive(Debug)]
pub struct Config {
    pub name: String,
}

//// ------------------------------------------------------------------------
// Plain function.

/// Adds two numbers together.
fn add(a: i32, b: i32) -> i32 {
    a + b
}

/** Block-style outer doc comment.
 *
 * Doubles a number.
 */
pub fn double(n: i32) -> i32 {
    n * 2
}

async fn fetch_remote(url: &str) -> String {
    url.to_string()
}

fn undocumented_helper() -> i32 {
    0
}
`

// rustWantDocs are the symbols the fixture documents and that rustSymbols
// extracts. Every one must get a doc comment.
var rustWantDocs = []string{"add", "double", "Config"}

// TestRustDocExtractorNamesMatchGraphSymbols is the drift guard for the
// symbol-name regexes rustdocs.go copies from internal/graph/symbols.go.
//
// Both directions are required. Checking only that every keyed doc name is a
// real graph symbol is not enough: breaking a regex makes the extractor produce
// *fewer* docs, and the survivors are still validly keyed, so a subset check
// passes while enrichment has quietly stopped working. So the fixture's
// documented symbols are all required, every keyed name must be a real graph
// symbol, and the undocumented symbol must stay out of the doc set.
func TestRustDocExtractorNamesMatchGraphSymbols(t *testing.T) {
	repoRoot := t.TempDir()
	outDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(repoRoot, "lib.rs"), []byte(rustDriftFixture), 0o644); err != nil {
		t.Fatal(err)
	}

	opts := graph.Options{RepoPath: repoRoot, OutDir: outDir}
	paths, err := graph.ResolvePaths(opts)
	if err != nil {
		t.Fatalf("ResolvePaths: %v", err)
	}
	paths.OutFull = outDir

	g, err := graph.Build(opts, paths)
	if err != nil {
		t.Fatalf("graph.Build: %v", err)
	}

	var symbols []string
	for _, n := range g.Nodes {
		if filepath.ToSlash(n.ID) == "lib.rs" {
			symbols = n.Symbols
			break
		}
	}
	if symbols == nil {
		t.Fatalf("graph.Build produced no node for lib.rs")
	}
	symbolSet := make(map[string]bool, len(symbols))
	for _, s := range symbols {
		symbolSet[s] = true
	}

	docs := rustDocComments(rustDriftFixture)
	for _, want := range rustWantDocs {
		if _, ok := docs[want]; !ok {
			t.Errorf("no doc comment for documented symbol %q; got %v. If graph.Build extracts it (%v) "+
				"but this package does not, the copied regex has drifted from symbols.go",
				want, rustSortedKeys(docs), symbols)
		}
		if !symbolSet[want] {
			t.Errorf("fixture symbol %q is documented but absent from graph.Node.Symbols %v", want, symbols)
		}
	}
	for name := range docs {
		if !symbolSet[name] {
			t.Errorf("keyed a doc under %q, but graph.Build extracted %v — nodeText would never look it up",
				name, symbols)
		}
		if !rustContains(rustWantDocs, name) {
			t.Errorf("produced a doc for undocumented symbol %q; the fixture documents only %v",
				name, rustWantDocs)
		}
	}
}

// TestRustCrateHeaderNotAttributed guards the inner/outer doc distinction. A
// "//! ..." run is an INNER doc comment: it documents the enclosing crate, never
// the item below it. graph.Node.Symbols holds no name for a crate, so keying
// the header to the first function would both be wrong and poison the embedding
// with unrelated text.
func TestRustCrateHeaderNotAttributed(t *testing.T) {
	docs := rustDocComments("//! Crate docs here.\n//! Still crate docs.\n\n/// Real docs.\nfn target() {}\n")
	got, ok := docs["target"]
	if !ok {
		t.Fatalf("expected a doc for target, got %v", rustSortedKeys(docs))
	}
	if got != "Real docs." {
		t.Errorf("target doc = %q, want %q — the //! header leaked into it", got, "Real docs.")
	}
}

// TestRustNoBlankLineAfterCrateHeader covers the case where a crate header runs
// straight into the first item with no separating blank line. That is legal and
// common, and without explicit handling the whole header becomes the item's
// doc.
func TestRustNoBlankLineAfterCrateHeader(t *testing.T) {
	docs := rustDocComments("//! Crate docs.\n/// Item docs.\nfn target() {}\n")
	got, ok := docs["target"]
	if !ok {
		t.Fatalf("expected a doc for target, got %v", rustSortedKeys(docs))
	}
	if got != "Item docs." {
		t.Errorf("target doc = %q, want %q", got, "Item docs.")
	}
}

// TestRustDividerIsNotDocComment: "////" is a plain LINE_COMMENT per the Rust
// Reference (the lexer takes the longer "//" "//" match), not an outer doc
// comment. A section divider must not be embedded as if it documented the item.
func TestRustDividerIsNotDocComment(t *testing.T) {
	docs := rustDocComments("//// Section divider.\nfn target() {}\n")
	if got, ok := docs["target"]; ok {
		t.Errorf("target got doc %q from a //// divider; //// is an ordinary line comment", got)
	}
}

// TestRustBlockDocNotOpened: "/**/" and "/*** ... */" are ordinary block
// comments, not outer block docs (which are exactly "/**" followed by neither
// "*" nor "/").
func TestRustBlockDocNotOpened(t *testing.T) {
	for _, src := range []string{
		"/**/\nfn target() {}\n",
		"/*** banner ***/\nfn target() {}\n",
	} {
		if got, ok := rustDocComments(src)["target"]; ok {
			t.Errorf("%q: target got doc %q from an ordinary block comment", src, got)
		}
	}
}

// TestRustBlockDocAboveLineDocs: a "/** ... */" block doc may be followed by a
// "///" run; the whole contiguous comment belongs to the item.
func TestRustBlockDocAboveLineDocs(t *testing.T) {
	docs := rustDocComments("/** Block docs. */\n/// More docs.\nfn target() {}\n")
	if _, ok := docs["target"]; !ok {
		t.Errorf("expected a doc for target, got %v", rustSortedKeys(docs))
	}
}

// TestRustUndocumentedItemHasNoDoc: an item with no comment above it contributes
// nothing, so nodeText falls back to a bare symbol name.
func TestRustUndocumentedItemHasNoDoc(t *testing.T) {
	docs := rustDocComments("/// Docs for other.\nfn documented() {}\n\nfn bare() {}\n")
	if _, ok := docs["bare"]; ok {
		t.Errorf("bare got doc %q, want none", docs["bare"])
	}
	if _, ok := docs["documented"]; !ok {
		t.Error("documented should have a doc")
	}
}

// TestRustFileWithNoDocs: returns nil, not an empty map, matching
// goDocComments/jsDocComments so callers can treat nil as "nothing to add".
func TestRustFileWithNoDocs(t *testing.T) {
	if got := rustDocComments("fn a() {}\nstruct B;\n"); got != nil {
		t.Errorf("rustDocComments = %v, want nil for a file with no docs", got)
	}
}

// TestRustDocsForExtensionWiring checks the nodeText-level path: a .rs node with
// symbols must get its docs through the docsFor dispatch, and a non-.rs node
// must not.
func TestRustDocsForExtensionWiring(t *testing.T) {
	repoRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(repoRoot, "lib.rs"), []byte("/// Real docs.\nfn target() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	n := graph.Node{ID: "lib.rs", Symbols: []string{"target"}}
	docs := docsFor(repoRoot, n)
	if got := docs["target"]; got != "Real docs." {
		t.Errorf("docsFor(.rs)[target] = %q, want %q", got, "Real docs.")
	}
	other := graph.Node{ID: "main.go", Symbols: []string{"target"}}
	if got := docsFor(repoRoot, other); got != nil {
		t.Errorf("docsFor(.go) = %v, want nil", got)
	}
}

func rustSortedKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func rustContains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
