package quality

// The constitution states the markdown ceiling inline in two places, and the
// number was wrong twice: it read 1,854 while `.markdownlint-baseline.json` held
// 1,601, and it kept reading 1,854 through three rule corrections that dropped
// the tree to 1,417. The gate itself was fine — `TestRepoMarkdownStaysWithinRatchet`
// reads the JSON — so nothing failed. The prose was simply describing a number
// nothing was bound to.
//
// The constitution already argues this exact point about its own coverage
// figures ("A number in this file is not a policy: the policy is that the JSON
// file is measured from the tree"), and the fix applied there is the fix that
// belongs here: the inline number is gone, replaced by a pointer to the
// measured file.
//
// This test is the backstop for that removal. It is deliberately permissive —
// restating the ceiling is not forbidden, restating it *wrong* is — so the
// bar moves with the file instead of becoming a number nobody dares update.

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/plaesy/spec-kit/internal/mdlint"
)

// ceilingClaim matches a markdown violation ceiling written as prose:
// "ceiling 1,854", "ceiling of 1,854 violations".
var ceilingClaim = regexp.MustCompile(`(?i)ceiling\s+(?:of\s+|is\s+|was\s+|stands at\s+)?([0-9][0-9,]*)`)

// amendLogHeading matches the amendment log's own heading, including the
// section number the constitution numbers its sections with. Matching on the
// bare words silently fails to find it when the number is there, and a check
// that quietly examines the whole file reports the historical rows as live
// claims — which is the failure this test exists to prevent, one level down.
var amendLogHeading = regexp.MustCompile(`(?m)^#+\s*(?:\d+[.)]\s*)?Amendment Log\b`)

// amendLogRows are the historical rows of the constitution's amendment log.
// They record what a ceiling was when a version was ratified; rewriting those
// would make the log claim a version moved a ceiling it never moved. Only the
// live statement outside the log is checked.
func amendLogRows(body string) string {
	loc := amendLogHeading.FindStringIndex(body)
	if loc == nil {
		return body
	}
	return body[:loc[0]]
}

// TestConstitutionDoesNotRestateTheMarkdownCeiling keeps the measured number in
// exactly one place.
//
// The rule is "do not restate it" rather than "restate it correctly" on
// purpose. A number in prose that is checked against another number is a number
// that has to be updated in two files every time a rule correction moves it,
// and the one time it was checked here it was already wrong — 1,854 in the
// constitution against 1,601 in the file, with the gate green the whole time
// because the gate reads the file. The constitution's own coverage entry
// rejects the pattern for its own figures, so this holds markdown to the same
// standard.
func TestConstitutionDoesNotRestateTheMarkdownCeiling(t *testing.T) {
	// moduleDir is scripts/; both files being compared sit one level above it.
	repoRoot := filepath.Dir(moduleDir(t))

	baselinePath := filepath.Join(repoRoot, ".markdownlint-baseline.json")
	baseline, err := mdlint.LoadBaseline(baselinePath)
	if err != nil {
		t.Fatalf("load %s: %v", baselinePath, err)
	}

	constitutionPath := filepath.Join(repoRoot, ".plaesy", "memory", "constitution.md")
	raw, err := os.ReadFile(constitutionPath)
	if err != nil {
		t.Skipf("constitution not present (%v)", err)
	}

	matches := ceilingClaim.FindAllStringSubmatch(amendLogRows(string(raw)), -1)
	for _, m := range matches {
		t.Errorf("the constitution states a markdown ceiling of %s outside its "+
			"amendment log. The enforced ceiling is %d, recorded in %s and measured "+
			"from the tree. Point at the file instead of copying the number; the "+
			"amendment log is the right place to record what a number used to be",
			m[1], baseline.MaxViolations, filepath.Base(baselinePath))
	}
}
