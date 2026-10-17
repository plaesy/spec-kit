package contract

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func bp(v bool) *bool { return &v }

// goodMethod is a method that passes the procedure check.
const goodMethod = "go test ./internal/validate/"

// base returns a valid, frozen, open contract with two items. Each test changes one thing.
func base() Contract {
	items := []Item{
		{ID: "C1", KPI: "Regression fixed", Metric: "failing tests in internal/validate", Target: "zero",
			Method: goodMethod, Blocking: bp(true)},
		{ID: "C2", KPI: "Changed function covered", Metric: "statement coverage of the changed function",
			Target: ">= 90", Method: "go test -cover ./internal/validate/", Blocking: bp(false)},
	}
	return Contract{ID: "fix-empty-slug", Phase: "fix", Dimension: "technical", Status: "open",
		Frozen: freezeOf(items), Items: items}
}

// freezeOf takes the frozen snapshot of items, as the phase does before work starts.
func freezeOf(items []Item) *Frozen {
	f := &Frozen{}
	for _, it := range items {
		f.Items = append(f.Items, FrozenItem{ID: it.ID, Metric: it.Metric, Target: it.Target, Method: it.Method, Blocking: it.Blocking})
	}
	return f
}

func errs(issues []Issue) []string {
	var out []string
	for _, i := range issues {
		if i.Severity == SevError {
			out = append(out, i.Item+": "+i.Message)
		}
	}
	return out
}

func warns(issues []Issue) []string {
	var out []string
	for _, i := range issues {
		if i.Severity == SevWarning {
			out = append(out, i.Item+": "+i.Message)
		}
	}
	return out
}

func contains(list []string, sub string) bool {
	for _, s := range list {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// root is a repo root for evidence checks. Tests that need an existing file create one here.
func root(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "report.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func mustErr(t *testing.T, c Contract, o Options, sub string) {
	t.Helper()
	got := errs(Validate(c, o))
	if !contains(got, sub) {
		t.Fatalf("want an error containing %q, got errors %q", sub, got)
	}
}

func mustPass(t *testing.T, c Contract, o Options) {
	t.Helper()
	if got := errs(Validate(c, o)); len(got) != 0 {
		t.Fatalf("want no errors, got %q", got)
	}
}

// gateReady returns base with every item recorded as met, with real evidence in dir.
func gateReady(t *testing.T) (Contract, Options) {
	t.Helper()
	c := base()
	c.Status = "closed"
	dir := root(t)
	c.Items[0].Result, c.Items[0].Actual, c.Items[0].Evidence = ResultMet, "0", "cmd:go test ./internal/validate/"
	c.Items[1].Result, c.Items[1].Actual, c.Items[1].Evidence = ResultMet, "92.4", "report.txt"
	return c, Options{Gate: true, RepoRoot: dir}
}

// ---- schema and placeholders ----

func TestValidFrozenOpenContractHasNoErrors(t *testing.T) {
	mustPass(t, base(), Options{RepoRoot: root(t)})
}

func TestSchemaRequiredFields(t *testing.T) {
	o := Options{}
	c := base()
	c.ID = "Bad ID"
	mustErr(t, c, o, "lower-case")

	c = base()
	c.Phase = "Implement!"
	mustErr(t, c, o, "phase")

	c = base()
	c.Dimension = "vibes"
	mustErr(t, c, o, "dimension")

	c = base()
	c.Status = "maybe"
	mustErr(t, c, o, "status")

	c = base()
	c.Items = nil
	mustErr(t, c, o, "at least one item")

	c = base()
	c.Items[0].KPI = ""
	mustErr(t, c, o, "kpi is required")

	c = base()
	c.Items[0].Method = ""
	mustErr(t, c, o, "method is required")

	c = base()
	c.Items[0].Blocking = nil
	mustErr(t, c, o, "blocking is required")
}

func TestDuplicateItemIDIsAnError(t *testing.T) {
	c := base()
	c.Items[1].ID = "C1"
	mustErr(t, c, Options{}, "duplicate item id")
}

// M-6: template placeholders are not accepted in any field.
func TestPlaceholdersAreRejectedInEveryField(t *testing.T) {
	c := base()
	c.Items[0].KPI = "{{KPI}}"
	c.Items[0].Metric = "{{METRIC}}"
	c.Items[0].Method = "{{METHOD}}"
	got := errs(Validate(c, Options{}))
	for _, field := range []string{"kpi still holds", "metric still holds"} {
		if !contains(got, field) {
			t.Errorf("want placeholder error for %q, got %q", field, got)
		}
	}
	c = base()
	c.ID = "{{SLUG}}"
	mustErr(t, c, Options{}, "id still holds a template placeholder")
}

// ---- vague words, targets, methods ----

func TestVagueWordsAreRejected(t *testing.T) {
	o := Options{}
	c := base()
	c.Items[0].KPI = "Good UX"
	mustErr(t, c, o, "vague word \"Good\"")

	c = base()
	c.Items[1].Target = "fast"
	mustErr(t, c, o, "vague word \"fast\"")

	c = base()
	c.Items[0].Metric = "reasonable quality"
	mustErr(t, c, o, "vague word")

	c = base()
	c.Items[1].Target = ">= 90 approx"
	mustErr(t, c, o, "vague word \"approx\"")
}

// m-6: a flag and a code span are not vague words.
func TestVagueWordsIgnoreFlagsAndCodeSpans(t *testing.T) {
	if got := vagueHit("run the suite with --fast"); got != "" {
		t.Errorf("a flag must not count as vague, got %q", got)
	}
	if got := vagueHit("count `TODO` markers"); got != "" {
		t.Errorf("a word in a code span must not count as vague, got %q", got)
	}
	if got := vagueHit("count TODO markers"); got == "" {
		t.Errorf("a bare TODO must count as vague")
	}
}

// M-3: the target grammar.
func TestTargetGrammar(t *testing.T) {
	accepted := []string{">= 90", "≥ 90", "<= 2 s", "<= 200 ms", ">= 80%", "zero", "none", "pass", "all pages", "every route", "all 5 pages"}
	for _, tg := range accepted {
		if !validTarget(tg) {
			t.Errorf("target %q must be accepted", tg)
		}
	}
	rejected := []string{"90", "about 90", "12 files", "approx 95", "between 80 and 90", "no more than 3", ">= 90 and <= 100", "bigger than before", "~90"}
	for _, tg := range rejected {
		if validTarget(tg) {
			t.Errorf("target %q must be rejected", tg)
		}
	}
}

func TestMethodMustBeACheck(t *testing.T) {
	c := base()
	c.Items[0].Method = "look at it"
	mustErr(t, c, Options{}, "is not a check")

	c = base()
	c.Items[0].Method = goodMethod
	mustPass(t, c, Options{RepoRoot: root(t)})
}

// m-7: the spec's list of vague words and the validator's list must match.
func TestSpecVagueListMatchesTheValidator(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "instructions", "success-contract.instructions.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	start := strings.Index(text, "**Rejected by the validator.**")
	if start < 0 {
		t.Fatal("the spec no longer has the 'Rejected by the validator' paragraph")
	}
	para := text[start:]
	if end := strings.Index(para, "\n\n"); end >= 0 {
		para = para[:end]
	}
	var fromSpec []string
	for _, m := range regexp.MustCompile("`([^`]+)`").FindAllStringSubmatch(para, -1) {
		fromSpec = append(fromSpec, strings.Join(strings.Fields(m[1]), " ")) // a wrapped line must not split a phrase
	}
	want := VagueWords()
	sort.Strings(fromSpec)
	sort.Strings(want)
	if !reflect.DeepEqual(fromSpec, want) {
		t.Fatalf("spec vague list and validator list differ:\nspec: %q\ncode: %q", fromSpec, want)
	}
}

// ---- blocking rules (B-2) ----

func TestAtLeastOneBlockingItemIsRequired(t *testing.T) {
	c := base()
	c.Items[0].Blocking = bp(false)
	c.Frozen = freezeOf(c.Items) // frozen with the new value, so only the rule is tested
	mustErr(t, c, Options{}, "at least one item must be blocking")
}

func TestNoBlockingWaiverNeedsAReason(t *testing.T) {
	c := base()
	c.Items[0].Blocking = bp(false)
	c.Frozen = freezeOf(c.Items)
	c.Waivers = []Waiver{{Kind: "no-blocking", Reason: "exploratory spike, no gate"}}
	got := Validate(c, Options{RepoRoot: root(t)})
	if len(errs(got)) != 0 || !contains(warns(got), "waived") {
		t.Fatalf("a logged no-blocking waiver must be accepted as a warning, got %+v", got)
	}

	c.Waivers = []Waiver{{Kind: "no-blocking", Reason: ""}}
	mustErr(t, c, Options{}, "waiver 1 needs a reason")

	c.Waivers = []Waiver{{Kind: "anything", Reason: "x"}}
	mustErr(t, c, Options{}, "kind \"anything\" is not supported")
}

// ---- freeze (B-1, B-3 content checks, M-8, m-3, m-4, m-8) ----

func TestUnfrozenWarnsThenErrorsOnGate(t *testing.T) {
	c := base()
	c.Frozen = nil
	if !contains(warns(Validate(c, Options{})), "not frozen yet") {
		t.Fatalf("an open, unfrozen contract should warn")
	}
	mustErr(t, c, Options{Gate: true}, "no frozen snapshot")
}

func TestChangedTargetNeedsAmendment(t *testing.T) {
	c := base()
	c.Items[1].Target = ">= 50" // weakened after freeze, no amendment
	mustErr(t, c, Options{}, "target changed after freeze")

	c.Amendments = []Amendment{{Item: "C2", Field: "target", From: ">= 90", To: ">= 50", Reason: "coverage floor for generated code"}}
	mustPass(t, c, Options{RepoRoot: root(t)})
}

func TestAmendmentFromMustMatchFrozen(t *testing.T) {
	c := base()
	c.Items[1].Target = ">= 50"
	c.Amendments = []Amendment{{Item: "C2", Field: "target", From: ">= 80", To: ">= 50", Reason: "reason"}}
	mustErr(t, c, Options{}, "does not match the frozen value")
}

func TestAmendmentWithoutReasonIsRejected(t *testing.T) {
	c := base()
	c.Items[1].Target = ">= 50"
	c.Amendments = []Amendment{{Item: "C2", Field: "target", From: ">= 90", To: ">= 50", Reason: ""}}
	mustErr(t, c, Options{}, "needs a reason")
	mustErr(t, c, Options{}, "target changed after freeze")
}

func TestBlockingCannotBeLoosened(t *testing.T) {
	c := base()
	c.Items[0].Blocking = bp(false)
	c.Amendments = []Amendment{{Item: "C1", Field: "blocking", From: "true", To: "false", Reason: "not important"}}
	mustErr(t, c, Options{}, "blocking was loosened from true to false")
}

// m-3: a loosened blocking item is reported once, not twice.
func TestLoosenedBlockingIsReportedOnce(t *testing.T) {
	c := base()
	c.Items[0].Blocking = bp(false)
	n := 0
	for _, e := range errs(Validate(c, Options{})) {
		if strings.HasPrefix(e, "C1: ") && strings.Contains(e, "loosened") {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("want the loosening reported once, got %d", n)
	}
}

// B-1, unit level: a blocking item removed after freeze is an error, with or without an amendment.
func TestBlockingItemRemovedAfterFreezeIsRejectedEvenWithAmendment(t *testing.T) {
	c := base()
	c.Items = c.Items[1:] // C1 (blocking) removed
	c.Amendments = []Amendment{{Item: "C1", Field: "removed", Reason: "scope cut"}}
	mustErr(t, c, Options{}, "blocking item removed after freeze: not allowed")
}

func TestNonBlockingRemovalNeedsAmendment(t *testing.T) {
	c := base()
	c.Items = c.Items[:1] // C2 (non-blocking) removed, no amendment
	mustErr(t, c, Options{}, "item removed after freeze with no amendment")

	c.Amendments = []Amendment{{Item: "C2", Field: "removed", Reason: "measured elsewhere"}}
	mustPass(t, c, Options{RepoRoot: root(t)})
}

func TestAddedItemNeedsAmendment(t *testing.T) {
	c := base()
	c.Items = append(c.Items, Item{ID: "C3", KPI: "k", Metric: "m", Target: "zero", Method: goodMethod, Blocking: bp(false)})
	mustErr(t, c, Options{}, "added after freeze")
	c.Amendments = []Amendment{{Item: "C3", Field: "added", Reason: "discovered in review"}}
	mustPass(t, c, Options{RepoRoot: root(t)})
}

// M-8: frozen blocking must be recorded, so the loosening check cannot be disabled by omitting it.
func TestFrozenItemMustRecordBlocking(t *testing.T) {
	c := base()
	c.Frozen.Items[0].Blocking = nil
	mustErr(t, c, Options{}, "frozen item needs metric, target, method and blocking")
}

func TestFrozenDuplicateIDIsAnError(t *testing.T) {
	c := base()
	c.Frozen.Items = append(c.Frozen.Items, c.Frozen.Items[0])
	mustErr(t, c, Options{}, "duplicate frozen item id")
}

// ---- gate: results, numbers, units, binary values, evidence ----

func TestGateMetNeedsActualAndEvidence(t *testing.T) {
	c := base()
	c.Items[0].Result = ResultMet
	c.Items[1].Result = ResultMet
	mustErr(t, c, Options{Gate: true}, "met needs the actual measured value")
	mustErr(t, c, Options{Gate: true}, "met needs evidence")
}

func TestGateNumericClaimsAreChecked(t *testing.T) {
	cases := []struct {
		target, actual string
		wantErr        string // empty means the claim must pass
	}{
		{">= 90", "92.4", ""},
		{">= 90", "85", "does not satisfy target"},
		{">= 90", "1,250", ""},               // thousands separator is read, not truncated
		{">= 90", "~95", "must be a number"}, // not readable: an error, not a pass
		{"<= 2 s", "1.5 s", ""},
		{"<= 2 s", "2500 ms", "does not satisfy"}, // time units convert
		{"<= 200 ms", "1.5 s", "does not satisfy"},
		{"<= 200 ms", "200", "units differ"}, // unit missing on the actual
		{">= 80%", "92", "units differ"},     // unit on the target, none on the actual
		{">= 90", "95%", "units differ"},
	}
	for _, c := range cases {
		r := numericSatisfies(c.target, c.actual)
		switch {
		case c.wantErr == "" && (r.problem != "" || !r.ok):
			t.Errorf("target %q actual %q: want pass, got %+v", c.target, c.actual, r)
		case c.wantErr != "" && r.problem == "" && r.ok:
			t.Errorf("target %q actual %q: want %q, got a pass", c.target, c.actual, c.wantErr)
		case c.wantErr != "" && r.problem == "" && !r.ok && !strings.Contains(c.wantErr, "does not satisfy"):
			t.Errorf("target %q actual %q: want %q, got a plain failure", c.target, c.actual, c.wantErr)
		}
	}
}

// M-1: binary targets are checked, not accepted on any value.
func TestGateBinaryClaimsAreChecked(t *testing.T) {
	cases := []struct {
		target, actual string
		ok             bool
	}{
		{"zero", "0", true},
		{"zero", "none", true},
		{"zero", "3 failing", false},
		{"pass", "pass", true},
		{"pass", "3 failing", false},
		{"all pages", "5 of 5", true},
		{"all pages", "3 of 5", false},
		{"all pages", "pass", true},
		{"absent", "none", true},
		{"present", "absent", false},
	}
	for _, c := range cases {
		if got := binarySatisfies(c.target, c.actual); got != c.ok {
			t.Errorf("binarySatisfies(%q, %q) = %v, want %v", c.target, c.actual, got, c.ok)
		}
	}
}

func TestGateRefusesMetThatContradictsItsTarget(t *testing.T) {
	c, o := gateReady(t)
	c.Items[0].Actual = "3 failing"
	mustErr(t, c, o, "does not satisfy the binary target")

	c, o = gateReady(t)
	c.Items[1].Actual = "85"
	mustErr(t, c, o, "recorded met, but actual \"85\" does not satisfy target \">= 90\"")

	c, o = gateReady(t)
	mustPass(t, c, o)
}

func TestGateBlockingNotMetFails(t *testing.T) {
	c, o := gateReady(t)
	c.Items[0].Result, c.Items[0].Evidence = ResultNotMet, "report.txt"
	mustErr(t, c, o, "blocking item not met")
}

// Spec section 5: a non-blocking not_met warns, provided the contract has a blocking item that can fail the phase.
func TestGateNonBlockingNotMetOnlyWarnsWhenABlockingItemExists(t *testing.T) {
	c, o := gateReady(t)
	c.Items[1].Result, c.Items[1].Actual = ResultNotMet, "70"
	c.Items[1].Evidence = "report.txt"
	got := Validate(c, o)
	if len(errs(got)) != 0 {
		t.Fatalf("non-blocking not_met must not fail when a blocking item passes, got %q", errs(got))
	}
}

// B-2 (gate side): all items non-blocking and one not met cannot pass the phase.
func TestGateAllNonBlockingWithANotMetFails(t *testing.T) {
	c := base()
	c.Status = "closed"
	c.Items[0].Blocking = bp(false)
	c.Frozen = freezeOf(c.Items)
	c.Waivers = nil
	c.Items[0].Result, c.Items[0].Evidence = ResultNotMet, "report.txt"
	c.Items[1].Result, c.Items[1].Actual, c.Items[1].Evidence = ResultMet, "92", "report.txt"
	mustErr(t, c, Options{Gate: true, RepoRoot: root(t)}, "no item is blocking")
}

func TestGateNotMeasurableNeedsReasonAndFailsWhenBlocking(t *testing.T) {
	c, o := gateReady(t)
	c.Items[0].Result, c.Items[0].Evidence = ResultNotMeasurable, ""
	mustErr(t, c, o, "not_measurable needs a reason")

	c.Items[0].Reason = "no test runner in this environment"
	mustErr(t, c, o, "blocking item not measurable")
}

func TestGateMissingResultIsAnError(t *testing.T) {
	c := base()
	mustErr(t, c, Options{Gate: true}, "result must be recorded")
}

func TestPlainModeRejectsUnknownResultValue(t *testing.T) {
	c := base()
	c.Items[0].Result = "probably"
	mustErr(t, c, Options{}, "must be met, not_met or not_measurable")
}

// M-7: evidence is a pointer that exists, or a recorded command.
func TestEvidenceMustBeAPointer(t *testing.T) {
	c, o := gateReady(t)
	c.Items[0].Evidence = "go test output"
	mustErr(t, c, o, "met needs evidence")

	c, o = gateReady(t)
	c.Items[1].Evidence = "does-not-exist.txt"
	mustErr(t, c, o, "evidence path \"does-not-exist.txt\" does not exist")

	c, o = gateReady(t)
	c.Items[1].Evidence = "report.txt:12"
	mustPass(t, c, o)

	c, o = gateReady(t)
	c.Items[1].Evidence = "cmd:go test -cover ./internal/validate/"
	mustPass(t, c, o)
}

// M-4: status closed is tied to the gate, in both modes.
func TestClosedStatusRunsTheGate(t *testing.T) {
	c := base()
	c.Status = "closed"
	c.Items[0].Result, c.Items[0].Evidence = ResultNotMet, "report.txt"
	c.Items[1].Result, c.Items[1].Actual, c.Items[1].Evidence = ResultMet, "92", "report.txt"
	// Plain mode, no --gate flag: a closed contract must still fail on its blocking not_met.
	mustErr(t, c, Options{RepoRoot: root(t)}, "blocking item not met")
}

func TestOpenStatusWarnsAtTheGate(t *testing.T) {
	c, o := gateReady(t)
	c.Status = "open"
	if !contains(warns(Validate(c, o)), "status is still open") {
		t.Fatalf("a gate run on an open contract must warn")
	}
}

// ---- parsing (M-5, m-2) ----

func TestParseRejectsTwoJSONBlocks(t *testing.T) {
	raw := []byte("```json\n{\"id\":\"x\",\"phase\":\"fix\",\"dimension\":\"general\",\"status\":\"open\",\"items\":[]}\n```\n\nexample:\n\n```json\n{\"item\":\"C1\"}\n```\n")
	if _, err := Parse(raw); err == nil || !strings.Contains(err.Error(), "exactly one") {
		t.Fatalf("a second json block must be an error, got %v", err)
	}
}

func TestParseRejectsTrailingContentInsideTheBlock(t *testing.T) {
	raw := []byte("```json\n{\"id\":\"x\",\"phase\":\"fix\",\"dimension\":\"general\",\"status\":\"open\",\"items\":[]} trailing\n```")
	if _, err := Parse(raw); err == nil {
		t.Fatal("content after the contract object must be rejected")
	}
}

func TestParseRequiresFencedBlockAndRejectsUnknownKeys(t *testing.T) {
	if _, err := Parse([]byte("no json here")); err == nil || !strings.Contains(err.Error(), "no fenced") {
		t.Fatalf("missing fence must be reported, got %v", err)
	}
	raw := []byte("```json\n{\"id\":\"x\",\"phase\":\"fix\",\"dimension\":\"general\",\"status\":\"open\",\"items\":[],\"itmes\":[]}\n```")
	if _, err := Parse(raw); err == nil {
		t.Fatal("a misspelt key must be rejected, not silently dropped")
	}
}

// M-5: the shipped template must parse, and its own example amendment must be a path the
// validator accepts. The amendment example is prose, so the block the validator reads is the
// only JSON in the file.
func TestTemplateFilledInValidates(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "templates", "success-contract.template.md"))
	if err != nil {
		t.Fatal(err)
	}
	filled := strings.NewReplacer(
		"{{PHASE}}", "fix", "{{SLUG}}", "fix-empty-slug", "{{DIMENSION}}", "technical",
		"{{METRIC}}", "failing tests", "{{TARGET}}", "zero", "{{METHOD}}", "go test ./internal/validate/",
		"{{KPI}}", "Regression fixed", "{{DATE}}", "2026-10-17",
	).Replace(string(raw))
	c, err := Parse([]byte(filled))
	if err != nil {
		t.Fatalf("filled template must parse: %v", err)
	}
	if got := errs(Validate(c, Options{})); len(got) != 0 {
		t.Fatalf("filled template must validate, got %q", got)
	}
}

// ---- size (m-1) ----

func TestLineCountMatchesWc(t *testing.T) {
	cases := map[string]int{"": 0, "a": 1, "a\n": 1, "a\nb": 2, "a\nb\n": 2}
	for in, want := range cases {
		if got := countLines([]byte(in)); got != want {
			t.Errorf("countLines(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestLengthCapIsAnError(t *testing.T) {
	dir := t.TempDir()
	var b strings.Builder
	b.WriteString("```json\n")
	b.WriteString(`{"id":"big","phase":"fix","dimension":"general","status":"open","frozen":{"items":[{"id":"C1","metric":"m","target":"zero","method":"` + goodMethod + `","blocking":true}]},"items":[{"id":"C1","kpi":"k","metric":"m","target":"zero","method":"` + goodMethod + `","blocking":true}]}`)
	b.WriteString("\n```\n")
	for i := 0; i < MaxLines+5; i++ {
		b.WriteString("note\n")
	}
	path := filepath.Join(dir, "big.contract.md")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	rep, err := Check(path, Options{RepoRoot: dir})
	if err != nil {
		t.Fatal(err)
	}
	if !contains(errs(rep.Issues), "the cap is 150") {
		t.Fatalf("a contract over the cap must be an error: %+v", rep.Issues)
	}
}

// ---- git anchor (B-3) ----

// gitRepo creates a temp repository with user identity set, for anchor tests.
func gitRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not on PATH: the freeze anchor cannot be exercised here")
	}
	dir := t.TempDir()
	run(t, dir, "init", "-q")
	run(t, dir, "config", "user.email", "test@example.invalid")
	run(t, dir, "config", "user.name", "test")
	run(t, dir, "config", "commit.gpgsign", "false")
	return dir
}

func run(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// writeContract writes c as a contract file (markdown with one json block) and returns its path.
func writeContract(t *testing.T, dir string, c Contract) string {
	t.Helper()
	body, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "fix-empty-slug.contract.md")
	text := "# Success Contract\n\n```json\n" + string(body) + "\n```\n"
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func commit(t *testing.T, dir, msg string) {
	t.Helper()
	run(t, dir, "add", "-A")
	run(t, dir, "commit", "-q", "-m", msg)
}

// freshFrozen returns an open contract frozen before any result, ready to commit as the anchor.
func freshFrozen() Contract {
	c := base()
	c.Amendments = nil
	return c
}

// anchored sets up a repo whose first commit holds the frozen contract. It returns the repo and file path.
func anchored(t *testing.T) (string, string) {
	t.Helper()
	dir := gitRepo(t)
	path := writeContract(t, dir, freshFrozen())
	commit(t, dir, "freeze the contract before work")
	return dir, path
}

// workDone records results and closes the contract, as the gate does.
func workDone(c Contract) Contract {
	c.Status = "closed"
	c.Items[0].Result, c.Items[0].Actual, c.Items[0].Evidence = ResultMet, "0", "cmd:go test ./internal/validate/"
	c.Items[1].Result, c.Items[1].Actual, c.Items[1].Evidence = ResultMet, "92.4", "cmd:go test -cover ./internal/validate/"
	return c
}

func TestAnchorPassesWhenUnchangedAndCommitted(t *testing.T) {
	dir, path := anchored(t)
	rep, err := Check(path, Options{RepoRoot: dir})
	if err != nil {
		t.Fatal(err)
	}
	if !rep.Anchor.Verified || rep.Errors() != 0 {
		t.Fatalf("an unchanged, committed freeze must verify: anchor=%+v issues=%+v", rep.Anchor, rep.Issues)
	}
}

// B-3 core: rewriting the frozen block to match an edited target is caught by git.
func TestAnchorCatchesRewrittenFreeze(t *testing.T) {
	dir, path := anchored(t)
	c := freshFrozen()
	c.Items[1].Target = "< 5"
	c.Frozen = freezeOf(c.Items) // the phase rewrites frozen to match its own edit
	writeContract(t, dir, c)
	commit(t, dir, "rewrite the freeze")
	rep, _ := Check(path, Options{Gate: true, RepoRoot: dir})
	if rep.Anchor.Verified || !contains(errs(rep.Issues), "frozen snapshot differs from the first committed version") {
		t.Fatalf("a rewritten freeze must fail the anchor: %+v", rep.Issues)
	}
}

// A target changed with a recorded amendment is allowed; the frozen snapshot stays as committed.
func TestAnchorAcceptsAnAmendmentAppendedToTheLog(t *testing.T) {
	dir, path := anchored(t)
	c := freshFrozen()
	c.Items[1].Target = ">= 80"
	c.Amendments = []Amendment{{Item: "C2", Field: "target", From: ">= 90", To: ">= 80", Reason: "coverage floor lowered for generated code"}}
	writeContract(t, dir, c)
	commit(t, dir, "amend target with a reason")
	rep, _ := Check(path, Options{RepoRoot: dir})
	if !rep.Anchor.Verified || rep.Errors() != 0 {
		t.Fatalf("an appended amendment must verify: %+v", rep.Issues)
	}
}

// B-1 through the anchor: removing a committed amendment is caught.
func TestAnchorCatchesAmendmentRemovedFromHistory(t *testing.T) {
	dir := gitRepo(t)
	c := freshFrozen()
	c.Items[1].Target = ">= 80"
	c.Frozen = freezeOf(base().Items) // frozen at the original target
	c.Amendments = []Amendment{{Item: "C2", Field: "target", From: ">= 90", To: ">= 80", Reason: "reason"}}
	path := writeContract(t, dir, c)
	commit(t, dir, "freeze and amend")
	c.Amendments = nil
	c.Items[1].Target = ">= 90"
	c.Frozen = freezeOf(base().Items)
	writeContract(t, dir, c)
	commit(t, dir, "drop the amendment")
	rep, _ := Check(path, Options{RepoRoot: dir})
	if rep.Anchor.Verified || !contains(errs(rep.Issues), "was removed or changed: the amendment log is append-only") {
		t.Fatalf("a dropped amendment must fail the anchor: %+v", rep.Issues)
	}
}

// A contract whose first frozen version already holds results was filled after the work.
func TestAnchorRejectsResultsInTheFirstFrozenVersion(t *testing.T) {
	dir := gitRepo(t)
	path := writeContract(t, dir, workDone(freshFrozen()))
	commit(t, dir, "freeze with results already in it")
	rep, _ := Check(path, Options{Gate: true, RepoRoot: dir})
	if !contains(errs(rep.Issues), "already records results") {
		t.Fatalf("results in the first frozen version must be an error: %+v", rep.Issues)
	}
}

// B-1 reproduction through the gate: a blocking item removed after freeze, with a reason,
// must not make the gate pass. This is the CLI case from the review.
func TestGateFailsWhenABlockingItemIsRemovedWithAReason(t *testing.T) {
	dir, path := anchored(t)
	c := freshFrozen()
	c.Items = c.Items[1:] // C1 (blocking) removed
	c.Amendments = []Amendment{{Item: "C1", Field: "removed", Reason: "scope cut"}}
	c.Status = "closed"
	c.Items[0].Result, c.Items[0].Actual, c.Items[0].Evidence = ResultMet, "92.4", "cmd:go test -cover ./internal/validate/"
	writeContract(t, dir, c)
	commit(t, dir, "remove the blocking item")
	rep, err := Check(path, Options{Gate: true, RepoRoot: dir})
	if err != nil {
		t.Fatal(err)
	}
	if !contains(errs(rep.Issues), "blocking item removed after freeze: not allowed") {
		t.Fatalf("removing a blocking item with a reason must fail the gate with the B-1 rule: %+v", rep.Issues)
	}
}

// An uncommitted contract cannot prove its freeze. The gate refuses; a plain check warns.
func TestUncommittedFreezeIsAnErrorAtTheGateAndAWarningOtherwise(t *testing.T) {
	dir := gitRepo(t)
	path := writeContract(t, dir, freshFrozen())
	rep, _ := Check(path, Options{RepoRoot: dir})
	if rep.Anchor.Verified || rep.Errors() != 0 || !contains(warns(rep.Issues), "freeze NOT VERIFIED") {
		t.Fatalf("an uncommitted freeze must warn in plain mode: %+v", rep.Issues)
	}
	rep, _ = Check(path, Options{Gate: true, RepoRoot: dir})
	if !contains(errs(rep.Issues), "freeze NOT VERIFIED") {
		t.Fatalf("an uncommitted freeze must fail the gate: %+v", rep.Issues)
	}
}

// Outside git the freeze cannot be verified. That is said plainly, and --require-git makes it fail.
func TestOutsideGitTheFreezeIsNotVerified(t *testing.T) {
	dir := t.TempDir() // not a repository
	path := writeContract(t, dir, freshFrozen())
	rep, _ := Check(path, Options{RepoRoot: dir})
	if rep.Anchor.Verified || !contains(warns(rep.Issues), "not in a git repository") {
		t.Fatalf("outside git the freeze must be reported as not verified: %+v", rep.Issues)
	}
	rep, _ = Check(path, Options{RepoRoot: dir, RequireGit: true})
	if !contains(errs(rep.Issues), "not in a git repository") {
		t.Fatalf("--require-git must make an unverifiable freeze an error: %+v", rep.Issues)
	}
}

// M-10: every starter target and method in the spec must pass the validator's own rules.
// Each starter line states its target and method in code spans:
//   - KPI. Target `X`. Method `Y`.
func TestSpecStarterSetsPassTheValidator(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "instructions", "success-contract.instructions.md"))
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile("Target `([^`]+)`\\. Method `([^`]+)`")
	matches := re.FindAllStringSubmatch(string(raw), -1)
	if len(matches) < 20 {
		t.Fatalf("expected the starter sets in section 8, found %d starter lines", len(matches))
	}
	for _, m := range matches {
		target, method := m[1], m[2]
		if !validTarget(target) {
			t.Errorf("starter target %q is rejected by the validator", target)
		}
		if !validMethod(method) {
			t.Errorf("starter method %q is rejected by the validator", method)
		}
		if vagueHit(target) != "" {
			t.Errorf("starter target %q uses a vague word", target)
		}
	}
}

// The worked example in section 9 must itself be a valid open contract.
func TestSpecWorkedExampleValidates(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "instructions", "success-contract.instructions.md"))
	if err != nil {
		t.Fatal(err)
	}
	m := reFence.FindSubmatch(raw)
	if m == nil {
		t.Fatal("the spec's worked example json block is missing")
	}
	var c Contract
	if err := json.Unmarshal(m[1], &c); err != nil {
		t.Fatalf("worked example does not parse: %v", err)
	}
	if got := errs(Validate(c, Options{RepoRoot: root(t)})); len(got) != 0 {
		t.Fatalf("worked example must validate, got %q", got)
	}
}

// contractText renders c as a contract file: prose, then one json block.
func contractText(t *testing.T, c Contract) string {
	t.Helper()
	body, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return "# Success Contract\n\n```json\n" + string(body) + "\n```\n"
}
