package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// `docs/reference.md` and `docs/scripts/platforms.md` carry the same
// subcommand table — and the same paragraph explaining why four `plaesy config`
// verbs moved to `plaesy platforms`. That repetition is deliberate and correct:
// a reference index has to be self-contained, and a per-command page has to stand
// on its own.
//
// `docs_drift_test.go` already checks that both files *mention* every
// subcommand. It does not check that the tables agree, so the risk is specific
// and real: someone edits one table and the other silently goes stale. That is
// the same failure the constitution's documentation bar exists to prevent — six
// `validate-*` rows once sat in that table for a release after the commands were
// gone, and nobody noticed because nothing compared the two.
//
// So: the duplication stays, and the drift gets a guard. De-duplicating instead
// would mean gutting the reference's self-containedness or breaking a passing
// check for a 22-line win.
// duplicatedTableDocs lists the two files and the row that identifies the table
// inside each. The marker is needed because docs/reference.md carries more than
// one `| Subcommand |` table: an earlier one covers the `plaesy config` verbs
// that stayed behind, and matching on the header alone compared the wrong pair
// and reported a length mismatch that did not exist.
var duplicatedTableDocs = []struct {
	a, b   string
	marker string
}{
	{"docs/reference.md", "docs/scripts/platforms.md", "`detect`"},
}

// tableContaining returns the contiguous run of markdown table rows that holds
// `marker`. Comparing the whole run rather than one named row means a row added,
// removed or reworded on either side is caught.
func tableContaining(t *testing.T, path, marker string) []string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	lines := strings.Split(string(raw), "\n")

	var current []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "|") {
			if len(current) > 0 {
				if hasMarker(current, marker) {
					return current
				}
				current = nil
			}
			continue
		}
		current = append(current, trimmed)
	}
	if len(current) > 0 && hasMarker(current, marker) {
		return current
	}
	t.Fatalf("%s: no markdown table contains a row starting with %q. The comparison "+
		"below would pass on an empty result, which is exactly the failure it exists "+
		"to catch.", path, marker)
	return nil
}

func hasMarker(table []string, marker string) bool {
	for _, row := range table {
		cells := strings.Split(row, "|")
		if len(cells) > 1 && strings.Contains(strings.TrimSpace(cells[1]), marker) {
			return true
		}
	}
	return false
}

func TestDuplicatedDocTablesStayInSync(t *testing.T) {
	root := corpusRoot(t)
	for _, pair := range duplicatedTableDocs {
		a := tableContaining(t, filepath.Join(root, pair.a), pair.marker)
		b := tableContaining(t, filepath.Join(root, pair.b), pair.marker)

		if len(a) != len(b) {
			t.Errorf("%s and %s carry the same subcommand table but at different lengths "+
				"(%d and %d rows). One was edited without the other.\n  %s:\n%s\n  %s:\n%s",
				pair.a, pair.b, len(a), len(b),
				pair.a, strings.Join(a, "\n    "),
				pair.b, strings.Join(b, "\n    "))
			continue
		}
		for i := range a {
			if a[i] != b[i] {
				t.Errorf("%s and %s must carry an identical subcommand table; row %d differs:\n"+
					"  %s: %s\n  %s: %s\n"+
					"  Edit both, or drop the duplication deliberately -- do not let them drift.",
					pair.a, pair.b, i+1, pair.a, a[i], pair.b, b[i])
			}
		}
		t.Logf("%s <-> %s: %d identical table row(s)", pair.a, pair.b, len(a))
	}
}
