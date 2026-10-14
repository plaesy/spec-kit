package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// .plaesy/instructions.md is the only description an agent has of what the
// installed instruction set contains. It is hand-maintained — `plaesy init`
// does not write it — which makes it drift by default.
//
// It had: three wrong counts (30 installed / 71 in the repo / 41 not
// installed, against a real 33 / 73 / 40), an Always Loaded heading claiming 9
// against a real 8, and three installed files missing from the list entirely,
// including `uncertainty-surfacing.md` and `context7-protocol.md`. An agent
// reading it concluded those protocols did not exist.
//
// The byte-identical freshness test in internal/scaffold cannot cover this file
// for the same reason it skips it: a hand-maintained index is legitimately not
// a copy. So the claim is checked semantically instead — recomputed from
// mapping.json, which is the source of truth.

var (
	indexEntry = regexp.MustCompile(`(?m)^- \[([a-z0-9.-]+)\.md\]`)
	mdHeading  = regexp.MustCompile(`^##\s`)
	headingNum = regexp.MustCompile(`\((\d+)\)`)
)

type instructionIndex struct {
	body   string
	listed map[string]bool
}

func loadInstructionIndex(t *testing.T) (root string, idx instructionIndex) {
	t.Helper()
	docs := repoDocs(t)
	if docs == "" {
		t.Skip("not a repository checkout")
	}
	root = filepath.Dir(filepath.Dir(docs))

	body, err := os.ReadFile(filepath.Join(root, ".plaesy", "instructions.md"))
	if err != nil {
		t.Skipf(".plaesy/instructions.md not present: %v", err)
	}
	idx = instructionIndex{body: string(body), listed: map[string]bool{}}
	for _, m := range indexEntry.FindAllStringSubmatch(idx.body, -1) {
		idx.listed[m[1]] = true
	}
	return root, idx
}

func loadMapping(t *testing.T, root string) (always, scope []string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "instructions", "mapping.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m struct {
		Mappings struct {
			AlwaysLoad []string `json:"always_load"`
			ScopeLoad  []string `json:"scope_load"`
		} `json:"mappings"`
	}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	trim := func(names []string) []string {
		out := make([]string, 0, len(names))
		for _, n := range names {
			out = append(out, strings.TrimSuffix(n, ".instructions.md"))
		}
		return out
	}
	return trim(m.Mappings.AlwaysLoad), trim(m.Mappings.ScopeLoad)
}

// indexSection returns the "- [x.md]" entries under the heading containing
// name, plus the heading line itself.
func indexSection(body, name string) (entries []string, heading string) {
	inSection := false
	for _, line := range strings.Split(body, "\n") {
		switch {
		case mdHeading.MatchString(line):
			inSection = strings.Contains(line, name)
			if inSection {
				heading = line
				entries = nil
			}
		case inSection:
			if m := indexEntry.FindStringSubmatch(line); m != nil {
				entries = append(entries, m[1])
			}
		}
	}
	return entries, heading
}

func TestInstructionsIndexListsExactlyTheInstalledSet(t *testing.T) {
	root, idx := loadInstructionIndex(t)
	always, scope := loadMapping(t, root)

	installed := map[string]bool{}
	for _, n := range always {
		installed[n] = true
	}
	for _, n := range scope {
		installed[n] = true
	}

	for name := range installed {
		if !idx.listed[name] {
			t.Errorf(".plaesy/instructions.md does not list %s.md, but mapping.json installs it. "+
				"An agent reading the index concludes the protocol does not exist.", name)
		}
	}
	for name := range idx.listed {
		if !installed[name] {
			t.Errorf(".plaesy/instructions.md lists %s.md, but mapping.json does not install it. "+
				"The link will 404 in an initialised project.", name)
		}
	}
}

func TestInstructionsIndexCountsMatchTheTree(t *testing.T) {
	root, idx := loadInstructionIndex(t)
	always, scope := loadMapping(t, root)

	repoFiles, err := filepath.Glob(filepath.Join(root, "instructions", "*.instructions.md"))
	if err != nil {
		t.Fatal(err)
	}
	total := len(repoFiles)
	notInstalled := total - len(always) - len(scope)

	claims := []struct {
		what    string
		pattern string
		want    int
	}{
		{"files installed", `\*\*(\d+) files installed`, len(always) + len(scope)},
		{"instruction files in the repository", `holds (\d+) instruction files`, total},
		{"files not installed", `Not [Ii]nstalled \((\d+)\)`, notInstalled},
	}
	for _, c := range claims {
		m := regexp.MustCompile(c.pattern).FindStringSubmatch(idx.body)
		if m == nil {
			t.Errorf(".plaesy/instructions.md has no %q claim, so this test cannot verify it", c.what)
			continue
		}
		got, err := strconv.Atoi(m[1])
		if err != nil {
			t.Errorf("%s claim %q is not a number: %v", c.what, m[1], err)
			continue
		}
		if got != c.want {
			t.Errorf(".plaesy/instructions.md claims %d %s; the tree has %d. Fix the number, "+
				"not the test — the index is a report, and a wrong one is worse than none.", got, c.what, c.want)
		}
	}
}

// TestInstructionsIndexAlwaysLoadedSectionIsExact is separate from the count
// check because the count alone was not what failed. The heading said 9 while
// always_load held 8, and date-system sat in that section after a slimming pass
// moved it to scope_load. Both the count and the membership are checked, because
// listing a scope-load file as always-loaded tells an agent to rely on
// something that will not be in its context.
func TestInstructionsIndexAlwaysLoadedSectionIsExact(t *testing.T) {
	root, idx := loadInstructionIndex(t)
	always, _ := loadMapping(t, root)

	entries, heading := indexSection(idx.body, "Always Loaded")
	if heading == "" {
		t.Fatal(".plaesy/instructions.md has no \"Always Loaded\" heading to check")
	}
	if m := headingNum.FindStringSubmatch(heading); m == nil {
		t.Errorf("the Always Loaded heading %q carries no count", heading)
	} else if n, _ := strconv.Atoi(m[1]); n != len(entries) {
		t.Errorf("the Always Loaded heading claims %s but the section lists %d file(s)", m[1], len(entries))
	}

	inAlways := map[string]bool{}
	for _, n := range always {
		inAlways[n] = true
	}
	listed := map[string]bool{}
	for _, n := range entries {
		if listed[n] {
			t.Errorf("%s.md is listed twice under Always Loaded", n)
		}
		listed[n] = true
		if !inAlways[n] {
			t.Errorf("%s.md is listed under Always Loaded, but mapping.json puts it in scope_load", n)
		}
	}
	for _, n := range always {
		if !listed[n] {
			t.Errorf("%s.md is in always_load but is not listed under Always Loaded", n)
		}
	}
}

// TestInstructionsReadmeAgreesWithMapping is the same guarantee applied to the
// second hand-maintained file that describes the install set.
//
// `docs/instructions/README.md` is not covered by the index test above, and it
// had drifted further: it claimed 30 installed of 71, against a real 33 of 73;
// an `always_load` of 9 including `date-system`, against a real 8 without it; a
// `scope_load` of 21 against a real 25; and "cannot load the other 41 files"
// against a real 40. Every one of those is a number a reader would act on, and
// the `date-system` one is a behavioural claim — the file was marked
// "Always Load: YES" long after a slimming pass moved it to `scope_load`.
//
// The important part is not the correction, it is that this file now has a
// guard at all. One hand-maintained description of a fact was enforced while a
// second, adjacent description of the same fact was not, and the enforced one
// being correct is exactly what kept the gap invisible.
func TestInstructionsReadmeAgreesWithMapping(t *testing.T) {
	root, _ := loadInstructionIndex(t)
	always, scope := loadMapping(t, root)

	repoFiles, err := filepath.Glob(filepath.Join(root, "instructions", "*.instructions.md"))
	if err != nil {
		t.Fatal(err)
	}
	installed := len(always) + len(scope)
	notInstalled := len(repoFiles) - installed

	readmePath := filepath.Join(root, "docs", "instructions", "README.md")
	raw, err := os.ReadFile(readmePath)
	if err != nil {
		t.Skipf("docs/instructions/README.md not present: %v", err)
	}
	body := string(raw)

	// Each claim is (what it describes, pattern, expected value). A missing
	// claim is itself a failure: silently passing on a file that stopped making
	// the claim would be the blind-guard failure this test exists to prevent.
	claims := []struct {
		what    string
		pattern string
		want    int
	}{
		{"installed count", "installs (\\d+) of", installed},
		{"repository total", `installs \d+ of the (\d+) `, len(repoFiles)},
		{"not-installed count", `cannot load the other (\d+) files`, notInstalled},
		{"index-size claim", `the (\d+) framework/scope files listed`, installed},
		{"always_load count", "\\| `always_load` \\| (\\d+) \\|", len(always)},
		{"scope_load count", "\\| `scope_load` \\| (\\d+) \\|", len(scope)},
	}
	for _, c := range claims {
		m := regexp.MustCompile(c.pattern).FindStringSubmatch(body)
		if m == nil {
			t.Errorf("docs/instructions/README.md makes no %q claim, so this test cannot verify it. "+
				"Either restore the sentence or drop it from this list.", c.what)
			continue
		}
		got, convErr := strconv.Atoi(m[1])
		if convErr != nil {
			t.Errorf("the %s claim %q is not a number: %v", c.what, m[1], convErr)
			continue
		}
		if got != c.want {
			t.Errorf("docs/instructions/README.md claims %d for the %s; mapping.json and the tree say %d. "+
				"Fix the README, not the test — an agent that reads this file plans against these numbers.",
				got, c.what, c.want)
		}
	}

	// The "Always Load" column is a behavioural claim, not a count: it tells a
	// reader which files are in context without being asked. Checked one way
	// deliberately — the table lists a subset of always_load, so demanding every
	// always_load entry appear would fail on files the table never claimed to be
	// exhaustive. But anything the table does call always-loaded must be.
	inAlways := map[string]bool{}
	for _, n := range always {
		inAlways[n] = true
	}
	rowRe := regexp.MustCompile(`(?m)^\|\s*\*\*\[([a-z0-9.-]+)\.instructions\.md\]\([^)]*\)\*\*[^\n]*\|` +
		`\s*(✅ YES|❌[^|]*?)\s*\|`)
	rows := 0
	for _, m := range rowRe.FindAllStringSubmatch(body, -1) {
		rows++
		if m[2] != "✅ YES" {
			continue
		}
		if !inAlways[m[1]] {
			t.Errorf("docs/instructions/README.md marks %s.instructions.md as \"Always Load: YES\", but "+
				"mapping.json puts it in scope_load. An agent reading this will rely on a file that is "+
				"copied but never loaded.", m[1])
		}
	}
	if rows == 0 {
		t.Error("docs/instructions/README.md has no parseable \"Always Load\" rows; this test would " +
			"pass without checking any of them")
	} else {
		t.Logf("%d Always Load row(s) checked against mapping.json", rows)
	}
}
