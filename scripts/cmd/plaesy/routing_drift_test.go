package main

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// instructions/dimension-mapping.instructions.md is the single canonical
// routing table for the corpus. Every other file is supposed to point at it
// rather than restate the mapping, which is exactly what makes it load-bearing:
// a row that is wrong here is wrong everywhere, and nothing else notices,
// because each caller is correctly deferring to it.
//
// It is currently complete — 54 of 54 (family x dimension) prompt files present,
// 43 of 43 {family}:{dimension} references resolving, 9 of 9 assess-<dimension>
// instruction files present. This test is here so that stays true, since a
// matrix this size drifts by deletion long before anyone re-reads it.

var (
	canonicalDimensions = []string{
		"technical", "design", "business", "product", "marketing",
		"legal", "financial", "management", "operations",
	}
	// The scope-taking command families. /doc, /save, /start, /continue,
	// /create and bare /loop take no dimension, and /improve's adoption, spec
	// and artifact forms are explicitly "not a dimension" — the table says so
	// itself, so they are excluded here rather than treated as gaps.
	scopeFamilies = []string{"assess", "implement", "fix", "optimize", "loop", "improve"}
	// Trailing arguments are allowed and expected: the table routes
	// `/optimize:financial --focus cost`, so anchoring the pattern to the
	// closing backtick would skip every reference that carries a --focus or a
	// --priority hint. Those are the majority of the routing cells.
	dimSelector = regexp.MustCompile(`` + "`" + `/([a-z]+):([a-z-]+)[^` + "`" + `]*` + "`")
	mdTableRow  = regexp.MustCompile(`(?m)^\|\s*\*\*([A-Za-z]+)\*\*\s*\|`)
)

func corpusRoot(t *testing.T) string {
	t.Helper()
	docs := repoDocs(t)
	if docs == "" {
		t.Skip("not a repository checkout")
	}
	return filepath.Dir(filepath.Dir(docs))
}

// TestRoutingMatrixIsComplete checks the 9 dimensions against the 6 scope
// families from the other direction: not "does every reference resolve" but
// "does every cell have a prompt file". A reference check cannot see a cell
// nobody referenced.
func TestRoutingMatrixIsComplete(t *testing.T) {
	root := corpusRoot(t)

	var gaps []string
	cells := 0
	for _, fam := range scopeFamilies {
		for _, dim := range canonicalDimensions {
			cells++
			rel := filepath.Join("prompts", fam, dim+".md")
			if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
				gaps = append(gaps, filepath.ToSlash(rel))
			}
		}
	}
	sort.Strings(gaps)
	for _, g := range gaps {
		t.Errorf("%s is missing. dimension-mapping.instructions.md is the canonical "+
			"routing table and this cell has no prompt, so an agent that routes there has "+
			"nothing to load. Add the file or remove the dimension from the table.", g)
	}
	t.Logf("%d routing cells checked, %d missing", cells, len(gaps))
}

// TestRoutingTableReferencesResolve checks the forward direction: every
// {family}:{dimension} the canonical table names must exist.
func TestRoutingTableReferencesResolve(t *testing.T) {
	root := corpusRoot(t)
	mapping := filepath.Join(root, "instructions", "dimension-mapping.instructions.md")
	raw, err := os.ReadFile(mapping)
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)

	fams := map[string]bool{}
	for _, f := range scopeFamilies {
		fams[f] = true
	}
	dims := map[string]bool{}
	for _, d := range canonicalDimensions {
		dims[d] = true
	}

	checked, skipped := 0, map[string]bool{}
	for _, m := range dimSelector.FindAllStringSubmatch(body, -1) {
		fam, dim := m[1], m[2]
		if !fams[fam] || !dims[dim] {
			skipped[fam+":"+dim] = true
			continue
		}
		checked++
		rel := filepath.Join("prompts", fam, dim+".md")
		if _, statErr := os.Stat(filepath.Join(root, rel)); statErr != nil {
			t.Errorf("dimension-mapping.instructions.md routes /%s:%s, but prompts/%s/%s.md "+
				"does not exist. Every other file defers to this table, so this one bad row "+
				"propagates to the whole corpus.", fam, dim, fam, dim)
		}
	}
	if checked == 0 {
		t.Fatal("no dimension selectors found; the pattern is wrong, which would make this " +
			"test pass without checking anything")
	}
	t.Logf("%d {family}:{dimension} reference(s) resolved, %d non-dimension selector(s) ignored: %v",
		checked, len(skipped), keysOf(skipped))
}

// TestEveryDimensionHasAnAssessmentInstructionFile is the third leg: the table
// routes /assess:<dimension> to a Mode 1/2/3 process that lives in
// assess-<dimension>.instructions.md. A dimension with a prompt but no
// instruction file routes an agent to a procedure that does not exist.
func TestEveryDimensionHasAnAssessmentInstructionFile(t *testing.T) {
	root := corpusRoot(t)
	for _, dim := range canonicalDimensions {
		rel := filepath.Join("instructions", "assess-"+dim+".instructions.md")
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			t.Errorf("instructions/assess-%s.instructions.md is missing. The routing table sends "+
				"/assess:%s to a process this file defines; without it the dimension is routable "+
				"but not assessable.", dim, dim)
		}
	}
	t.Logf("%d dimension(s) checked", len(canonicalDimensions))
}

// TestRoutingTableCoversExactlyTheCanonicalDimensions guards the table against
// silently gaining or losing a row. The row scan is scoped to the Routing Table
// section, because the Priority Mapping table further down also uses bolded
// first cells (CRITICAL, HIGH, ...) and a whole-file scan counts those as
// dimensions — which is how an earlier check of this file reported 15 rows and
// looked like it had found six extra dimensions.
func TestRoutingTableCoversExactlyTheCanonicalDimensions(t *testing.T) {
	root := corpusRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, "instructions", "dimension-mapping.instructions.md"))
	if err != nil {
		t.Fatal(err)
	}
	section := section(string(raw), "Routing Table")
	if section == "" {
		t.Fatal("no 'Routing Table' section found; this test cannot check the table")
	}

	found := map[string]bool{}
	for _, m := range mdTableRow.FindAllStringSubmatch(section, -1) {
		found[strings.ToLower(m[1])] = true
	}
	for _, d := range canonicalDimensions {
		if !found[d] {
			t.Errorf("the Routing Table has no row for %q. It is the canonical mapping; a "+
				"dimension with no row is one an agent cannot route to.", d)
		}
	}
	for name := range found {
		known := false
		for _, d := range canonicalDimensions {
			if d == name {
				known = true
				break
			}
		}
		if !known {
			t.Errorf("the Routing Table has a row for %q, which is not one of the %d canonical "+
				"dimensions. Either it is a new dimension — in which case every cell of the "+
				"matrix needs a prompt file too — or the row is stale.", name, len(canonicalDimensions))
		}
	}
	t.Logf("%d dimension row(s) in the Routing Table", len(found))
}

// multiSegmentSelector matches a nested sub-command such as
// `/create:templates:api`. The dimension selectors above are two-segment, so
// nothing else in this file covers this shape.
var multiSegmentSelector = regexp.MustCompile(`` + "`" + `/([a-z-]+):([a-z-]+):([a-z-]+)` + "`")

// TestNestedSubCommandNamesResolveToFiles binds a three-segment sub-command to
// the file it names.
//
// `/create:templates:ci` is only a true statement if
// `prompts/create/templates/ci.md` exists, because the segment in the name is
// the directory in the path. `prompts/create/api.md` was documented as
// `/create:templates:api` at three places while sitting outside `templates/`,
// next to the three files that were in the right place — so the command name
// and the file that backs it disagreed, and nothing checked. The mirror in
// `.kilo/commands/` reproduced the same wrong layout, which is what made it
// look deliberate rather than accidental.
func TestNestedSubCommandNamesResolveToFiles(t *testing.T) {
	root := corpusRoot(t)

	// Collect every nested selector documented anywhere in the prompt tree.
	documented := map[string][]string{} // command -> []selector
	err := filepath.Walk(filepath.Join(root, "prompts"), func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(p, ".md") {
			return nil
		}
		raw, readErr := os.ReadFile(p)
		if readErr != nil {
			return readErr
		}
		for _, m := range multiSegmentSelector.FindAllStringSubmatch(string(raw), -1) {
			sel := m[1] + ":" + m[2] + ":" + m[3]
			documented[sel] = append(documented[sel], relTo(root, p))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(documented) == 0 {
		// `/create:templates:{api,ci,infra,project}` were the only three-segment
		// selectors ever in this corpus; they were removed in favor of stack
		// scaffolding living in instructions/*.instructions.md (auto-loaded per
		// file type via mapping.json) instead of a dedicated /create sub-router.
		// A future three-segment selector should still be caught, so this stays
		// a Skip with a reason rather than a silent pass.
		t.Skip("no nested sub-command selectors in the corpus — the three-segment " +
			"shape (/create:templates:*) was removed; if one is reintroduced, this " +
			"test starts checking it automatically")
	}

	keys := make([]string, 0, len(documented))
	for k := range documented {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, sel := range keys {
		parts := strings.Split(sel, ":")
		want := filepath.Join(root, "prompts", parts[0], parts[1], parts[2]+".md")
		if _, statErr := os.Stat(want); statErr != nil {
			t.Errorf("/%s is documented in %v, but %s does not exist.\n"+
				"  The middle segment is a directory in the path, so the name and the file must "+
				"  agree. Either move the file into place or correct the name.",
				sel, documented[sel], relTo(root, want))
		}
	}
	t.Logf("%d nested sub-command selector(s) resolved against the tree", len(keys))
}

func relTo(root, p string) string {
	r, err := filepath.Rel(root, p)
	if err != nil {
		return p
	}
	return filepath.ToSlash(r)
}

func keysOf(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// section returns the body of the "## <heading>" section that contains
// `heading`, up to the next "## " heading of the same or higher level. Used to
// scope row scans to one table when the file contains several.
func section(body, heading string) string {
	lines := strings.Split(body, "\n")
	start := -1
	for i, l := range lines {
		if strings.HasPrefix(l, "## ") && strings.EqualFold(strings.TrimSpace(strings.TrimPrefix(l, "## ")), heading) {
			start = i + 1
			break
		}
	}
	if start < 0 {
		return ""
	}
	end := len(lines)
	for i := start; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			end = i
			break
		}
	}
	return strings.Join(lines[start:end], "\n")
}
