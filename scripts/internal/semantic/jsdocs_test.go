package semantic

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/plaesy/spec-kit/internal/graph"
)

func TestJSDocComments_Function(t *testing.T) {
	src := `/**
 * ValidateUser checks the given credentials against the user store.
 * @param user the account name
 */
export function validateUser(user, pass) {
  return true;
}

// LineComment is documented with plain // lines instead.
export function lineComment() {}

export function undocumented() {}
`
	docs := jsDocComments(src)
	want := "ValidateUser checks the given credentials against the user store.\n@param user the account name"
	if docs["validateUser"] != want {
		t.Errorf("docs[validateUser] = %q, want %q", docs["validateUser"], want)
	}
	if docs["lineComment"] != "LineComment is documented with plain // lines instead." {
		t.Errorf("docs[lineComment] = %q", docs["lineComment"])
	}
	if _, ok := docs["undocumented"]; ok {
		t.Errorf("undocumented should have no doc comment, got %q", docs["undocumented"])
	}
}

func TestJSDocComments_Class(t *testing.T) {
	src := `/** Session is a login session. */
export class Session {}

class Undocumented {}
`
	docs := jsDocComments(src)
	if docs["Session"] != "Session is a login session." {
		t.Errorf("docs[Session] = %q, want %q", docs["Session"], "Session is a login session.")
	}
	if _, ok := docs["Undocumented"]; ok {
		t.Errorf("Undocumented should have no doc comment, got %q", docs["Undocumented"])
	}
}

// A blank line detaches a comment from the declaration below it, the same
// rule Go's doc-comment grammar uses — otherwise a comment belonging to some
// earlier statement would be merged into the next declaration's docs.
func TestJSDocComments_BlankLineDetachesComment(t *testing.T) {
	src := `// hashPassword hashes the stored password.

// detachedComment is not a doc comment for the function below it.
export function detached() {}
`
	docs := jsDocComments(src)
	want := "detachedComment is not a doc comment for the function below it."
	if docs["detached"] != want {
		t.Errorf("docs[detached] = %q, want only the contiguous comment %q", docs["detached"], want)
	}

	orphan := jsDocComments("// orphan comment\n\nexport function orphan() {}\n")
	if _, ok := orphan["orphan"]; ok {
		t.Errorf("a comment separated by a blank line must not be attributed, got %q", orphan["orphan"])
	}
}

func TestJSDocComments_NoComments_ReturnsNil(t *testing.T) {
	docs := jsDocComments("export const x = 1;\nexport function f() {}\n")
	if docs != nil {
		t.Errorf("expected nil docs for a file with no doc comments, got %v", docs)
	}
}

func TestJSDocComments_UnterminatedBlockComment_ReturnsNil(t *testing.T) {
	src := "const s = 1;\n*/\nexport function f() {}\n"
	if docs := jsDocComments(src); docs != nil {
		t.Errorf("expected nil for a dangling */ with no opening /*, got %v", docs)
	}
}

// The keys must be exactly the names internal/graph's jsSymbols puts into
// graph.Node.Symbols, or nodeText's lookup would never match. The expected
// keys below are the ones asserted by graph's own
// TestSymbolsOfJS (function declaration / exported function / class /
// exported class).
func TestJSDocComments_KeysMatchGraphSymbols(t *testing.T) {
	src := `/** foo docs */
function foo(a) {}

/** bar docs */
export function bar(b) {}

/** baz docs */
class Baz {}
`
	docs := jsDocComments(src)
	for _, name := range []string{"foo", "bar", "Baz"} {
		if docs[name] == "" {
			t.Errorf("keys = %v, want an entry for %q (jsSymbols yields that name)", keysOf(docs), name)
		}
	}
	if len(docs) != 3 {
		t.Errorf("keys = %v, want exactly the 3 declared symbols", keysOf(docs))
	}
}

func keysOf(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// nodeText is the only caller, so this is the end-to-end check that the
// JS/TS extension dispatch actually reaches jsDocComments — a doc map keyed
// correctly but never wired up would fail here and nowhere else.
func TestNodeText_JSExtensionsIncludeJSDoc(t *testing.T) {
	for _, ext := range []string{".js", ".jsx", ".ts", ".tsx", ".JS"} {
		t.Run(ext, func(t *testing.T) {
			repoRoot := t.TempDir()
			rel := "web/auth" + ext
			writeFile(t, filepath.Join(repoRoot, filepath.FromSlash(rel)), `/**
 * ValidateUser checks the given credentials against the user store.
 * @param user the account name
 */
export function validateUser(user, pass) {
  return true;
}
`)
			n := graph.Node{ID: rel, Symbols: []string{"validateUser"}}
			want := "validateUser: ValidateUser checks the given credentials against the user store."
			if text := nodeText(repoRoot, n); !strings.Contains(text, want) {
				t.Errorf("nodeText = %q, want it to contain %q", text, want)
			}
		})
	}
}

func TestNodeText_JSFileWithNoDocComment_FallsBackToBareSymbol(t *testing.T) {
	repoRoot := t.TempDir()
	writeFile(t, filepath.Join(repoRoot, "web", "util.js"), "export function noop() {}\n")
	n := graph.Node{ID: "web/util.js", Symbols: []string{"noop"}}

	text := nodeText(repoRoot, n)
	for _, line := range strings.Split(text, "\n") {
		if line == "noop" {
			return
		}
	}
	t.Errorf("nodeText = %q, want a bare \"noop\" line", text)
}

// A node whose file is missing (or unreadable) must degrade to bare symbols
// rather than fail indexing — the doc comment is best-effort signal.
func TestNodeText_JSFileMissingOnDisk_FallsBackToBareSymbol(t *testing.T) {
	n := graph.Node{ID: "web/gone.js", Symbols: []string{"missing"}}
	text := nodeText(t.TempDir(), n)
	for _, line := range strings.Split(text, "\n") {
		if line == "missing" {
			return
		}
	}
	t.Errorf("nodeText = %q, want a bare \"missing\" line", text)
}

// An extension with no extractor (see docsFor) contributes bare symbols.
//
// The file has to actually exist: docsFor returns nil for an unreadable file
// regardless of extension, so against an empty directory this test would pass
// even for a wired extension and prove nothing about the dispatch.
func TestDocsFor_UnsupportedExtension_ReturnsNil(t *testing.T) {
	for _, ext := range []string{".md", ".txt", ".json", ""} {
		t.Run("ext="+ext, func(t *testing.T) {
			repoRoot := t.TempDir()
			writeFile(t, filepath.Join(repoRoot, "src", "lib"+ext), "# Docs.\n")
			if docs := docsFor(repoRoot, graph.Node{ID: "src/lib" + ext}); docs != nil {
				t.Errorf("docsFor(%q) = %v, want nil", "src/lib"+ext, docs)
			}
		})
	}
}
