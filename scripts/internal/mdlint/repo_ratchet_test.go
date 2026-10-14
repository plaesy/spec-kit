package mdlint_test

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/plaesy/spec-kit/internal/mdlint"
)

// repoRoot walks up from the package directory to the directory holding
// .markdownlint.json. The test is a repo check, not a library check, so it
// silently opts out when the config is not there — vendoring the Go module
// elsewhere should not drag a Markdown policy along with it.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 8; i++ {
		if _, err := os.Stat(filepath.Join(dir, mdlint.DefaultConfigName)); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Skipf("no %s above %s: not a repository checkout", mdlint.DefaultConfigName, t.Name())
	return ""
}

// TestRepoMarkdownStaysWithinRatchet is the CI gate. `go test ./...` is the only
// job in CI, so this test is what keeps Markdown linting alive after the npm
// toolchain was removed: no Node, no npm, no second CI job to forget to update.
//
// It fails only when the violation total rises above the recorded ceiling.
// Re-measuring is an explicit, reviewable act (validate markdown --write-baseline),
// not a side effect of a green run.
func TestRepoMarkdownStaysWithinRatchet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping repository-wide Markdown scan in -short mode")
	}
	root := repoRoot(t)

	cfg, err := mdlint.LoadConfig(filepath.Join(root, mdlint.DefaultConfigName))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	res, err := mdlint.Lint(root, []string{root}, cfg, nil)
	if err != nil {
		t.Fatalf("lint: %v", err)
	}
	if res.Scanned == 0 {
		t.Fatal("scanned zero Markdown files: the walker is broken, not the tree clean")
	}

	baseline, err := mdlint.LoadBaseline(filepath.Join(root, mdlint.DefaultBaselineName))
	if err != nil {
		t.Fatalf("load baseline: %v", err)
	}
	if baseline == nil {
		// LoadBaseline returns (nil, nil) for an absent file. That is how a
		// zero-tolerance corpus is encoded: with no ceiling recorded there is
		// nothing to breach, so any violation at all is a regression. Treating
		// absence as "markdown would be unchecked" would mean the ratchet could
		// be switched off by deleting the file it lives in.
		if res.Total > 0 {
			t.Errorf("no %s and %d markdown violation(s): this tree is zero-tolerance.\n\n"+
				"Fix them (plaesy validate markdown --fix, then by hand). Record a positive\n"+
				"ceiling with 'plaesy validate markdown --write-baseline' only once a rise is\n"+
				"understood and accepted.",
				mdlint.DefaultBaselineName, res.Total)
		}
		t.Logf("%d file(s) scanned, %d violation(s), no ceiling recorded (zero tolerance)",
			res.Scanned, res.Total)
		return
	}
	if res.Exceeds(baseline) {
		t.Errorf("markdown ratchet broken: %d violations, ceiling %d\n\nNew violations by rule:\n%s\n\n"+
			"Fix them (plaesy validate markdown --fix, then by hand). Only re-measure with\n"+
			"'plaesy validate markdown --write-baseline' once the rise is understood.",
			res.Total, baseline.MaxViolations, histogram(res, baseline))
	}
	t.Logf("%d file(s) scanned, %d violation(s), ceiling %d", res.Scanned, res.Total, baseline.MaxViolations)
}

// histogram shows the rules that grew relative to the recorded per-rule counts,
// so the failure message names the cause instead of only the total.
func histogram(res *mdlint.Result, baseline *mdlint.Baseline) string {
	var grew []string
	for rule, n := range res.ByRule {
		if n > baseline.ByRule[rule] {
			grew = append(grew, "  "+rule+": "+strconv.Itoa(n)+" (was "+strconv.Itoa(baseline.ByRule[rule])+")")
		}
	}
	for rule, n := range baseline.ByRule {
		if res.ByRule[rule] == 0 && n > 0 {
			grew = append(grew, "  "+rule+": 0 (was "+strconv.Itoa(n)+", improved)")
		}
	}
	if len(grew) == 0 {
		return "  (per-rule counts match the baseline; the totals disagree — check the config or the walker)"
	}
	return strings.Join(grew, "\n")
}
