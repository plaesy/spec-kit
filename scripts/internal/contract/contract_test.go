package contract

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func bp(v bool) *bool { return &v }

// base returns a valid, frozen, open contract with two items. Each test changes one thing.
func base() Contract {
	items := []Item{
		{ID: "C1", KPI: "Regression fixed", Metric: "failing tests in internal/validate", Target: "zero",
			Method: "go test ./internal/validate/", Blocking: bp(true)},
		{ID: "C2", KPI: "Changed function covered", Metric: "statement coverage of the changed function",
			Target: ">= 90", Method: "go test -cover ./internal/validate/", Blocking: bp(false)},
	}
	frozen := &Frozen{}
	for _, it := range items {
		frozen.Items = append(frozen.Items, FrozenItem{ID: it.ID, Metric: it.Metric, Target: it.Target, Method: it.Method, Blocking: it.Blocking})
	}
	return Contract{ID: "fix-empty-slug", Phase: "fix", Dimension: "technical", Status: "open", Frozen: frozen, Items: items}
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

func mustErr(t *testing.T, c Contract, gate bool, sub string) {
	t.Helper()
	got := errs(Validate(c, gate))
	if !contains(got, sub) {
		t.Fatalf("want an error containing %q, got errors %q", sub, got)
	}
}

func TestValidFrozenOpenContractHasNoErrors(t *testing.T) {
	if got := errs(Validate(base(), false)); len(got) != 0 {
		t.Fatalf("valid contract must have no errors, got %q", got)
	}
}

func TestSchemaRequiredFields(t *testing.T) {
	c := base()
	c.ID = "Bad ID"
	mustErr(t, c, false, "lower-case")

	c = base()
	c.Phase = "Implement!"
	mustErr(t, c, false, "phase")

	c = base()
	c.Dimension = "vibes"
	mustErr(t, c, false, "dimension")

	c = base()
	c.Status = "maybe"
	mustErr(t, c, false, "status")

	c = base()
	c.Items = nil
	mustErr(t, c, false, "at least one item")

	c = base()
	c.Items[0].KPI = ""
	mustErr(t, c, false, "kpi is required")

	c = base()
	c.Items[0].Method = ""
	mustErr(t, c, false, "method is required")

	c = base()
	c.Items[0].Blocking = nil
	mustErr(t, c, false, "blocking is required")
}

func TestDuplicateItemIDIsAnError(t *testing.T) {
	c := base()
	c.Items[1].ID = "C1"
	mustErr(t, c, false, "duplicate item id")
}

func TestVagueWordsAreRejected(t *testing.T) {
	c := base()
	c.Items[0].KPI = "Good UX"
	mustErr(t, c, false, "vague word \"Good\"")

	c = base()
	c.Items[1].Target = "fast"
	mustErr(t, c, false, "vague word \"fast\"")

	c = base()
	c.Items[0].Metric = "reasonable quality"
	mustErr(t, c, false, "vague word")
}

func TestTargetMustBeNumericOrBinary(t *testing.T) {
	c := base()
	c.Items[1].Target = "bigger than before"
	mustErr(t, c, false, "neither a number with a comparison nor a binary condition")

	c = base()
	c.Items[1].Target = ">= 80%"
	c.Frozen.Items[1].Target = ">= 80%" // frozen with the value, so only validity is tested
	if got := errs(Validate(c, false)); len(got) != 0 {
		t.Fatalf("a percentage target with a comparison is valid, got %q", got)
	}
}

func TestMethodMustBeACheck(t *testing.T) {
	c := base()
	c.Items[0].Method = "look"
	mustErr(t, c, false, "too short to be a check")
}

func TestUnfrozenIsWarningBeforeGateAndErrorAfter(t *testing.T) {
	c := base()
	c.Frozen = nil
	if !contains(warns(Validate(c, false)), "not frozen yet") {
		t.Fatalf("an open, unfrozen contract should warn")
	}
	mustErr(t, c, true, "not frozen before work started")
}

func TestChangedTargetNeedsAmendment(t *testing.T) {
	c := base()
	c.Items[1].Target = ">= 50" // weakened after freeze, no amendment
	mustErr(t, c, false, "target changed after freeze")

	c.Amendments = []Amendment{{Item: "C2", Field: "target", From: ">= 90", To: ">= 50", Reason: "coverage floor for generated code"}}
	if got := errs(Validate(c, false)); len(got) != 0 {
		t.Fatalf("a recorded amendment with a reason must be accepted, got %q", got)
	}
}

func TestAmendmentWithoutReasonIsRejected(t *testing.T) {
	c := base()
	c.Items[1].Target = ">= 50"
	c.Amendments = []Amendment{{Item: "C2", Field: "target", From: ">= 90", To: ">= 50", Reason: ""}}
	mustErr(t, c, false, "needs a reason")
	mustErr(t, c, false, "target changed after freeze")
}

func TestBlockingCannotBeLoosened(t *testing.T) {
	c := base()
	c.Items[0].Blocking = bp(false)
	c.Amendments = []Amendment{{Item: "C1", Field: "blocking", From: "true", To: "false", Reason: "not important"}}
	mustErr(t, c, false, "blocking was loosened from true to false")
}

func TestAddedAndRemovedItemsNeedAmendments(t *testing.T) {
	c := base()
	c.Items = append(c.Items, Item{ID: "C3", KPI: "k", Metric: "m", Target: "zero", Method: "a real command", Blocking: bp(false)})
	mustErr(t, c, false, "added after freeze")
	c.Amendments = []Amendment{{Item: "C3", Field: "added", Reason: "discovered in review"}}
	if got := errs(Validate(c, false)); len(got) != 0 {
		t.Fatalf("recorded addition must pass, got %q", got)
	}

	c = base()
	c.Items = c.Items[:1]
	mustErr(t, c, false, "removed after freeze")
}

func TestGateMetNeedsActualAndEvidence(t *testing.T) {
	c := base()
	c.Items[0].Result = ResultMet
	c.Items[1].Result = ResultMet
	mustErr(t, c, true, "met needs the actual measured value")
	mustErr(t, c, true, "met needs evidence")
}

func TestGateRefusesMetThatContradictsNumericTarget(t *testing.T) {
	c := base()
	c.Items[0].Result, c.Items[0].Actual, c.Items[0].Evidence = ResultMet, "0", "go test output"
	c.Items[1].Result, c.Items[1].Actual, c.Items[1].Evidence = ResultMet, "85", "coverage.out"
	mustErr(t, c, true, "recorded met, but actual \"85\" does not satisfy target \">= 90\"")

	c.Items[1].Actual = "92.4"
	if got := errs(Validate(c, true)); len(got) != 0 {
		t.Fatalf("met with a satisfying actual must pass, got %q", got)
	}
}

func TestGateBlockingNotMetFails(t *testing.T) {
	c := base()
	c.Items[0].Result, c.Items[0].Evidence = ResultNotMet, "failing test name"
	c.Items[1].Result, c.Items[1].Actual, c.Items[1].Evidence = ResultMet, "95", "coverage.out"
	mustErr(t, c, true, "blocking item not met")
}

func TestGateNonBlockingNotMetOnlyWarns(t *testing.T) {
	c := base()
	c.Items[0].Result, c.Items[0].Actual, c.Items[0].Evidence = ResultMet, "0", "go test output"
	c.Items[1].Result, c.Items[1].Actual, c.Items[1].Evidence = ResultNotMet, "70", "coverage.out"
	got := Validate(c, true)
	if len(errs(got)) != 0 {
		t.Fatalf("non-blocking not_met must not fail the gate, got errors %q", errs(got))
	}
}

func TestGateNotMeasurableNeedsReasonAndFailsWhenBlocking(t *testing.T) {
	c := base()
	c.Items[0].Result = ResultNotMeasurable
	mustErr(t, c, true, "not_measurable needs a reason")

	c.Items[0].Reason = "no test runner in this environment"
	mustErr(t, c, true, "blocking item not measurable")
}

func TestGateMissingResultIsAnError(t *testing.T) {
	c := base()
	mustErr(t, c, true, "result must be recorded")
}

func TestPlainModeRejectsUnknownResultValue(t *testing.T) {
	c := base()
	c.Items[0].Result = "probably"
	mustErr(t, c, false, "must be met, not_met or not_measurable")
}

func TestNumericSatisfies(t *testing.T) {
	cases := []struct {
		target, actual string
		ok, known      bool
	}{
		{">= 90", "92.4", true, true},
		{">= 90", "85", false, true},
		{"<= 200", "150", true, true},
		{"< 200", "200", false, true},
		{"> 0", "0.1", true, true},
		{"= 0", "0", true, true},
		{"zero", "0", false, false},
		{">= 90", "unknown", false, false},
	}
	for _, c := range cases {
		ok, known := numericSatisfies(c.target, c.actual)
		if ok != c.ok || known != c.known {
			t.Errorf("numericSatisfies(%q, %q) = %v, %v; want %v, %v", c.target, c.actual, ok, known, c.ok, c.known)
		}
	}
}

func TestParseReadsFirstJSONFence(t *testing.T) {
	raw := []byte("# Contract\n\nprose here\n\n```json\n{\"id\":\"x\",\"phase\":\"fix\",\"dimension\":\"general\",\"status\":\"open\",\"items\":[]}\n```\n")
	c, err := Parse(raw)
	if err != nil || c.ID != "x" {
		t.Fatalf("Parse = %+v, %v", c, err)
	}
}

func TestParseRejectsMissingFenceAndUnknownKeys(t *testing.T) {
	if _, err := Parse([]byte("no json here")); err == nil || !strings.Contains(err.Error(), "no fenced") {
		t.Fatalf("missing fence must be reported, got %v", err)
	}
	raw := []byte("```json\n{\"id\":\"x\",\"phase\":\"fix\",\"dimension\":\"general\",\"status\":\"open\",\"items\":[],\"itmes\":[]}\n```")
	if _, err := Parse(raw); err == nil {
		t.Fatal("a misspelt key must be rejected, not silently dropped")
	}
}

// The same contract must produce the same issues, in the same order, every run.
func TestIssueOrderIsStable(t *testing.T) {
	c := base()
	c.Items[0].KPI = "Good UX"
	c.Items[0].Target = "fast enough"
	c.Items[0].Method = "look"
	c.Items[1].Blocking = nil
	c.Items[1].Target = "bigger"
	want := Validate(c, true)
	for i := 0; i < 30; i++ {
		if got := Validate(c, true); !reflect.DeepEqual(got, want) {
			t.Fatalf("run %d differs:\n%v\n%v", i, want, got)
		}
	}
}

func TestCheckReportsLongContractAsWarning(t *testing.T) {
	dir := t.TempDir()
	var b strings.Builder
	b.WriteString("```json\n")
	b.WriteString(`{"id":"big","phase":"fix","dimension":"general","status":"open","items":[{"id":"C1","kpi":"k","metric":"m","target":"zero","method":"a real command","blocking":false}]}`)
	b.WriteString("\n```\n")
	for i := 0; i < MaxLinesWarn+5; i++ {
		b.WriteString("note\n")
	}
	path := filepath.Join(dir, "big.contract.md")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	rep, err := Check(path, false)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Errors() != 0 || rep.Warnings() == 0 || !contains(warns(rep.Issues), "too large") {
		t.Fatalf("a long but valid contract must warn about size only: %+v", rep)
	}
}
