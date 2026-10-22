package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Anthropic's prompting guidance says to tell the model what to DO rather than
// what NOT to: positive framing is the reliable lever for output control, and a
// section that is *only* prohibitions gives the model a list of things to avoid
// with nothing to aim at. It also warns against anti-laziness caps, because
// "CRITICAL: you MUST" causes overtriggering.
//
// The corpus had institutionalised the wrong shape. Twenty-two files carried a
// section headed `## Anti-Patterns (NEVER Do These)`, carrying 251 bullets of
// the form `- ❌ **Never X** - Y`, where Y was already the positive rule. All 22
// were rewritten to lead with the action and keep the boundary. Pure
// prohibitions fell 778 -> 641; what remains is overwhelmingly the endorsed
// shape ("Version the contract, don't mutate it silently" -- action, then
// boundary), which is why the count is deliberately not ratcheted further.
//
// So this guard pins the specific defect -- prohibition-titled sections and
// bullets that open with a marker rather than an action -- rather than a ratio.
// A ratio would have been the fourth bad metric of this project: the two before
// it each reported a number an order of magnitude off.
var (
	prohibitionHeading = regexp.MustCompile(`(?i)^#{1,6}\s+.*(` +
		`anti-pattern|things?\s+(not|never)\s+to\s+do|what\s+not\s+to\s+do|never\s+do\s+these` +
		`)`)
	// A bullet that opens with a marker instead of naming the action. Wrapped
	// continuation lines are excluded, and so is the `❌ old -> ✅ new` migration
	// shape, which does name its positive on the same line.
	markerLedBullet = regexp.MustCompile(`(?m)^\s*-\s*(❌|⛔|🚫)\s+`)
	migrationPair   = regexp.MustCompile(`❌.*(→|=>)\s*✅`)
	// A `❌ **BAD**:` example is a demonstration pair; the `✅ **GOOD**` line
	// beside it is the instruction, so the pair is positive in effect.
	badExample = regexp.MustCompile(`❌\s*\*\*BAD\*\*`)
	// Descriptive headings that mention an anti-pattern without being one:
	// `Anti-Patterns -> Solutions` names the solution, and `Watch for the ...`
	// is a thing to notice rather than a rule against something.
	descriptiveHeading = regexp.MustCompile(`(?i)(->|→|watch for)`)
	// A stop-condition block, where ❌ means "do not keep going". That is the
	// marker's actual meaning there, not a prohibition on an action.
	stopConditionBlock = regexp.MustCompile(`(?i)(stop loop if|continue loop if|stop .* if\b)`)
)

// behaviouralCorpus is prompts/ and instructions/ only. templates/ holds
// documents written for people to fill in, where a ❌ checkbox is the correct
// notation and stripping it would make the template worse.
var behaviouralDirs = []string{"prompts", "instructions"}

func walkMarkdown(t *testing.T, root string) []string {
	t.Helper()
	var out []string
	for _, d := range behaviouralDirs {
		err := filepath.Walk(filepath.Join(root, d), func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				if strings.HasPrefix(info.Name(), ".") {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.HasSuffix(p, ".md") {
				out = append(out, p)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return out
}

// TestNoProhibitionTitledSections is the structural half of the finding: a
// section whose title is itself a prohibition tells the model what to avoid
// without telling it what to produce.
func TestNoProhibitionTitledSections(t *testing.T) {
	root := corpusRoot(t)
	bad := 0
	for _, p := range walkMarkdown(t, root) {
		raw, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		rel, _ := filepath.Rel(root, p)
		lines := strings.Split(string(raw), "\n")
		for i, line := range lines {
			if !prohibitionHeading.MatchString(line) {
				continue
			}
			if descriptiveHeading.MatchString(line) {
				continue
			}
			rel := rel
			t.Errorf("%s:%d: section is titled as a prohibition: %q\n"+
				"  Lead with the action and keep the boundary instead, e.g.\n"+
				"  '## What a Good Spec Does' rather than '## Anti-Patterns (NEVER Do These)'.",
				rel, i+1, strings.TrimSpace(line))
			bad++
		}
	}
	if bad == 0 {
		t.Logf("no prohibition-titled section in %d files", len(walkMarkdown(t, root)))
	}
}

// TestNoMarkerLedBullets is the content half. A bullet opening with ❌ names the
// thing to stop doing; the same bullet rewritten to name the thing to do is
// strictly more useful, and the reason can be carried along as a clause.
func TestNoMarkerLedBullets(t *testing.T) {
	root := corpusRoot(t)
	files := walkMarkdown(t, root)
	checked := 0
	for _, p := range files {
		raw, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		rel, _ := filepath.Rel(root, p)
		lines := strings.Split(string(raw), "\n")
		for i, line := range lines {
			if !markerLedBullet.MatchString(line) {
				continue
			}
			checked++
			if migrationPair.MatchString(line) || badExample.MatchString(line) {
				// `❌ old -> ✅ new` and `❌ **BAD**` next to `✅ **GOOD**` both
				// name the positive in the same breath, so they steer rather
				// than merely forbid.
				continue
			}
			// Look back for a stop-condition marker. The window has to cover
			// the whole list, not just its first entries: a six-item block with
			// one marker above it would otherwise fail on items five and six.
			inStopBlock := false
			for j := i - 1; j >= 0 && j >= i-12; j-- {
				if stopConditionBlock.MatchString(lines[j]) {
					inStopBlock = true
					break
				}
			}
			if inStopBlock {
				continue
			}
			t.Errorf("%s:%d: bullet opens with a prohibition marker rather than an action:\n  %s\n"+
				"  Rewrite it to name what to produce, and carry the reason as a clause.",
				rel, i+1, strings.TrimSpace(line))
		}
	}
	t.Logf("%d marker-led bullets examined across %d files", checked, len(files))
}
