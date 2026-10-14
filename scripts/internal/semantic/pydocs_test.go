package semantic

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/plaesy/spec-kit/internal/graph"
)

func TestPyDocComments_Func(t *testing.T) {
	src := `def validate_user(user, pass):
    """Checks the given credentials against the user store.

    Returns false on any mismatch.
    """
    return True


def undocumented():
    return None
`
	docs := pyDocComments(src)
	want := "Checks the given credentials against the user store.\nReturns false on any mismatch."
	if docs["validate_user"] != want {
		t.Errorf("docs[validate_user] = %q, want %q", docs["validate_user"], want)
	}
	if _, ok := docs["undocumented"]; ok {
		t.Errorf("undocumented should have no docstring, got %q", docs["undocumented"])
	}
}

func TestPyDocComments_SingleQuoteDelimiter(t *testing.T) {
	src := `def render_date(ts):
    '''Formats a timestamp for display.'''
    return ts
`
	docs := pyDocComments(src)
	if docs["render_date"] != "Formats a timestamp for display." {
		t.Errorf("docs[render_date] = %q, want %q", docs["render_date"], "Formats a timestamp for display.")
	}
}

func TestPyDocComments_Class(t *testing.T) {
	src := `class Session:
    """A login session."""

    def refresh(self):
        """Rotates the session token."""

    def undocumented_method(self):
        pass


class Undocumented:
    pass
`
	docs := pyDocComments(src)
	if docs["Session"] != "A login session." {
		t.Errorf("docs[Session] = %q, want %q", docs["Session"], "A login session.")
	}
	if docs["refresh"] != "Rotates the session token." {
		t.Errorf("docs[refresh] = %q, want %q", docs["refresh"], "Rotates the session token.")
	}
	for _, name := range []string{"undocumented_method", "Undocumented"} {
		if _, ok := docs[name]; ok {
			t.Errorf("%s should have no docstring, got %q", name, docs[name])
		}
	}
}

// A leading "#" comment is allowed before the docstring, and a trailing
// "#" comment on the def line must not hide the body that follows it.
func TestPyDocComments_CommentsAroundDocstring(t *testing.T) {
	src := `def hash_password(pw):  # noqa: E501
    # keep this next to the definition
    """Hashes a password for storage."""
    return pw


def no_docstring_after_comment():
    # just a comment
    return 1
`
	docs := pyDocComments(src)
	if docs["hash_password"] != "Hashes a password for storage." {
		t.Errorf("docs[hash_password] = %q, want %q", docs["hash_password"], "Hashes a password for storage.")
	}
	if _, ok := docs["no_docstring_after_comment"]; ok {
		t.Errorf("a comment-only body has no docstring, got %q", docs["no_docstring_after_comment"])
	}
}

// An unterminated triple quote would otherwise swallow the rest of the file.
func TestPyDocComments_UnterminatedDocstring(t *testing.T) {
	src := `def broken():
    """never closed
    return None
`
	if docs := pyDocComments(src); docs != nil {
		t.Errorf("expected nil for an unterminated docstring, got %v", docs)
	}
}

func TestPyDocComments_NoDocstrings_ReturnsNil(t *testing.T) {
	if docs := pyDocComments("x = 1\n\ndef f(a):\n    return a\n"); docs != nil {
		t.Errorf("expected nil docs for a file with no docstrings, got %v", docs)
	}
}

// The keys must be exactly the names internal/graph's pySymbols puts into
// graph.Node.Symbols, or nodeText's lookup would never match. The expected
// keys below are the ones asserted by graph's own TestSymbolsOfPython
// (including the indented method form).
func TestPyDocComments_KeysMatchGraphSymbols(t *testing.T) {
	src := `class A:
    """a docs"""

    def m(self):
        """m docs"""
`
	docs := pyDocComments(src)
	for _, name := range []string{"A", "m"} {
		if docs[name] == "" {
			t.Errorf("keys = %v, want an entry for %q (pySymbols yields that name)", keysOf(docs), name)
		}
	}
	if len(docs) != 2 {
		t.Errorf("keys = %v, want exactly the 2 declared symbols", keysOf(docs))
	}
}

// nodeText is the only caller, so this is the end-to-end check that the .py
// extension dispatch actually reaches pyDocComments — a doc map keyed
// correctly but never wired up would fail here and nowhere else.
func TestNodeText_PyIncludesDocstring(t *testing.T) {
	repoRoot := t.TempDir()
	rel := "svc/auth.py"
	writeFile(t, filepath.Join(repoRoot, filepath.FromSlash(rel)), `def validate_user(user, pass):
    """Checks the given credentials against the user store.

    Returns false on any mismatch.
    """
    return True
`)
	n := graph.Node{ID: rel, Symbols: []string{"validate_user"}}
	want := "validate_user: Checks the given credentials against the user store."
	if text := nodeText(repoRoot, n); !strings.Contains(text, want) {
		t.Errorf("nodeText = %q, want it to contain %q", text, want)
	}
}

func TestNodeText_PyFileWithNoDocstring_FallsBackToBareSymbol(t *testing.T) {
	repoRoot := t.TempDir()
	writeFile(t, filepath.Join(repoRoot, "svc", "util.py"), "def noop():\n    return None\n")
	n := graph.Node{ID: "svc/util.py", Symbols: []string{"noop"}}

	text := nodeText(repoRoot, n)
	for _, line := range strings.Split(text, "\n") {
		if line == "noop" {
			return
		}
	}
	t.Errorf("nodeText = %q, want a bare \"noop\" line", text)
}
