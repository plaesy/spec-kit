package semantic

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/plaesy/spec-kit/internal/graph"
)

// TestDocExtractorNamesMatchGraphSymbols guards the one coupling this package
// cannot express in the type system.
//
// jsdocs.go and pydocs.go carry private copies of internal/graph/symbols.go's
// symbol-name regexes, because those regexes are unexported in another
// package. The copies are load-bearing: nodeText looks a doc comment up by the
// name in graph.Node.Symbols, so a name the extractor recognises but graph
// does not extract produces a doc keyed under a symbol that is never looked
// up.
//
// If someone edits a regex in symbols.go, the other tests in this package
// still pass — the copied regex just quietly stops recognising names, and
// JS/TS/Python doc-comment enrichment silently stops working. That is the
// worst failure mode for this feature: a false negative inside an
// anti-duplication check, with nothing reporting it.
//
// Two directions are asserted, and both are needed. Checking only that every
// keyed doc name appears in graph.Node.Symbols is not sufficient: breaking a
// regex makes the extractor produce *fewer* docs, and the few that remain are
// still legitimately keyed, so a subset check passes while enrichment has
// quietly rotted. So the fixture names the symbols it documents and the test
// requires all of them:
//
//  1. every documented symbol in the fixture gets a doc  (catches drift),
//  2. every keyed doc name is a real graph symbol         (catches mis-keying),
//  3. the undocumented fixture symbol gets no doc         (catches over-reach).
func TestDocExtractorNamesMatchGraphSymbols(t *testing.T) {
	cases := []struct {
		rel      string
		content  string
		extract  func(string) map[string]string
		wantDocs []string
	}{
		{
			rel:      "web/auth.js",
			content:  jsDriftFixture,
			extract:  jsDocComments,
			wantDocs: []string{"validateUser", "createSession", "AccountPanel"},
		},
		{
			rel:      "web/auth.ts",
			content:  jsDriftFixture,
			extract:  jsDocComments,
			wantDocs: []string{"validateUser", "createSession", "AccountPanel"},
		},
		{
			rel:      "web/widget.tsx",
			content:  jsDriftFixture,
			extract:  jsDocComments,
			wantDocs: []string{"validateUser", "createSession", "AccountPanel"},
		},
		{
			rel:      "svc/accounts.py",
			content:  pyDriftFixture,
			extract:  pyDocComments,
			wantDocs: []string{"validate_user", "Account"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.rel, func(t *testing.T) {
			repoRoot := t.TempDir()
			outDir := t.TempDir()
			path := filepath.Join(repoRoot, filepath.FromSlash(tc.rel))
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(tc.content), 0o644); err != nil {
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
				if filepath.ToSlash(n.ID) == tc.rel {
					symbols = n.Symbols
					break
				}
			}
			if symbols == nil {
				t.Fatalf("graph.Build produced no node for %s (ids: %v)", tc.rel, nodeIDs(g))
			}
			symbolSet := make(map[string]bool, len(symbols))
			for _, s := range symbols {
				symbolSet[s] = true
			}

			docs := tc.extract(tc.content)

			// (1) every documented symbol must have been recognised.
			for _, want := range tc.wantDocs {
				if _, ok := docs[want]; !ok {
					t.Errorf("extractor found no doc comment for documented symbol %q; got %v. "+
						"If graph.Build still extracts it (%v) but this package does not, the copied "+
						"regex has drifted from internal/graph/symbols.go and enrichment is silently dead",
						want, sortedKeys(docs), symbols)
				}
				if !symbolSet[want] {
					t.Errorf("fixture symbol %q is documented but absent from graph.Node.Symbols %v — "+
						"the fixture and internal/graph/symbols.go disagree", want, symbols)
				}
			}

			// (2) nothing may be keyed under a name graph does not extract,
			// (3) and an undocumented symbol must stay out of the doc set.
			for name := range docs {
				if !symbolSet[name] {
					t.Errorf("extractor keyed a doc comment under %q, but graph.Build extracted %v — "+
						"nodeText would never look this comment up", name, symbols)
				}
				if !containsString(tc.wantDocs, name) {
					t.Errorf("extractor produced a doc for undocumented symbol %q; the fixture documents only %v",
						name, tc.wantDocs)
				}
			}
		})
	}
}

// jsDriftFixture deliberately mixes shapes: a plain declaration, an exported
// one, a class, and a non-JSDoc "/* */" block — anything the extractor
// recognises here must also survive graph.Build's own extraction.
const jsDriftFixture = `/**
 * Validates a user's credentials.
 * @param {string} user
 */
function validateUser(user, pass) {
  return !!(user && pass);
}

/** Creates a session token for a user. */
export function createSession(user) {
  return token(user);
}

// Renders the account panel.
export class AccountPanel {
  render() {}
}
`

// pyDriftFixture mixes a documented def, a documented class, and an
// undocumented symbol that must stay out of the extracted doc set.
const pyDriftFixture = `def validate_user(name, password):
    """Checks a user's credentials against the store."""
    return bool(name and password)


class Account:
    """Represents a single account."""

    def balance(self):
        return 0


def undocumented_helper():
    return 1
`

func nodeIDs(g *graph.Graph) []string {
	out := make([]string, 0, len(g.Nodes))
	for _, n := range g.Nodes {
		out = append(out, n.ID)
	}
	sort.Strings(out)
	return out
}

func sortedKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func containsString(haystack []string, needle string) bool {
	for _, s := range haystack {
		if strings.EqualFold(s, needle) {
			return true
		}
	}
	return false
}
