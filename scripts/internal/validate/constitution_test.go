package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// constitutionFixture is a fully-valid, already-ratified constitution. Tests
// mutate single fields of it so each failure mode is isolated to one cause.
const constitutionFixture = `---
title: "Example Constitution"
version: "1.2.0"
ratified: "2026-01-15"
last_amended: "2026-09-25"
active_dimensions: [technical, business]
---

# Example Constitution

## 0. How To Read This File

## 1. Active Dimensions

| Dimension | Active | Deliverable | Binding threshold(s) | Mandatory audit (hard stop) |
|---|---|---|---|---|
| Technical | yes | code, service, API | 90% coverage, <200ms p95 | Full quality-gates.md run |
| Design | no | UI/UX | WCAG 2.1 AA | WCAG pass |
| Business | yes | business plan | unit economics reviewed | Viability review |
| Product | no | roadmap | prioritization criteria | Criteria applied |
| Marketing | no | campaign | every claim sourced | 100% claims sourced |
| Legal | no | contracts | traceable to source | Compliance pass |
| Financial | no | pricing | costs sourced | Unit economics hold |
| Management | no | team | roles named | No single point of failure |

## 4. Universal Rules (all dimensions)

| ID | Rule | Level | Detection anchor |
|---|---|---|---|
| EV-01 | Every claim carries a source | MUST | citation present |
| QG-01 | A check that cannot run must not report as passing | MUST | skip is distinct |

## 5. Non-Negotiables (Hard Stops)

| ID | Non-negotiable | Level | Detection anchor |
|---|---|---|---|
| NN-01 | Never store secrets in source control | MUST | secret scanner |
| NN-02 | Never disable a failing test | MUST | CI gate |

## 9. Amendment Log

| Date | Version | Change | Reason |
|------|---------|--------|--------|
| 2026-01-15 | 1.0.0 | Initial ratification | Project start |
| 2026-09-25 | 1.2.0 | Added business dimension | Scope change |
`

func writeFixture(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "constitution.md")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return path
}

// mutate replaces old with new exactly once, failing the test if old is absent
// so a fixture edit that silently stops exercising a check cannot pass green.
func mutate(t *testing.T, old, new string) string {
	t.Helper()
	if strings.Count(constitutionFixture, old) != 1 {
		t.Fatalf("fixture no longer contains exactly one %q", old)
	}
	return strings.Replace(constitutionFixture, old, new, 1)
}

func TestSplitRowTreatsEscapedPipeAsLiteral(t *testing.T) {
	got := splitRow(`| Technical | {{ACTIVE_TECHNICAL\|no}} | a \| b |`)
	want := []string{"Technical", "{{ACTIVE_TECHNICAL|no}}", "a | b"}
	if !equalStrings(got, want) {
		t.Errorf("splitRow = %q, want %q", got, want)
	}
}

func TestConstitutionAcceptsValidFile(t *testing.T) {
	res, err := Constitution(writeFixture(t, constitutionFixture))
	if err != nil {
		t.Fatalf("Constitution: unexpected error %v", err)
	}
	// ActiveDimensions is sorted so the result is stable across runs.
	if got, want := res.ActiveDimensions, []string{"business", "technical"}; !equalStrings(got, want) {
		t.Errorf("ActiveDimensions = %v, want %v", got, want)
	}
	if res.Rules != 2 {
		t.Errorf("Rules = %d, want 2", res.Rules)
	}
	if res.NonNegotiables != 2 {
		t.Errorf("NonNegotiables = %d, want 2", res.NonNegotiables)
	}
}

func TestConstitutionRejectsUnfilledPlaceholders(t *testing.T) {
	path := writeFixture(t, mutate(t, "90% coverage, <200ms p95", "{{TEST_COVERAGE_MIN|90}}% coverage, <200ms p95"))
	_, err := Constitution(path)
	if err == nil || !strings.Contains(err.Error(), "unfilled") {
		t.Fatalf("expected unfilled-placeholder error, got %v", err)
	}
}

func TestConstitutionRequiresFrontmatter(t *testing.T) {
	body := strings.SplitN(constitutionFixture, "---\n", 3)[2]
	if _, err := Constitution(writeFixture(t, body)); err == nil || !strings.Contains(err.Error(), "frontmatter") {
		t.Fatalf("expected frontmatter error, got %v", err)
	}
}

func TestConstitutionRequiresActiveDimensionsKey(t *testing.T) {
	path := writeFixture(t, mutate(t, "active_dimensions: [technical, business]\n", ""))
	_, err := Constitution(path)
	if err == nil || !strings.Contains(err.Error(), "active_dimensions") {
		t.Fatalf("expected active_dimensions error, got %v", err)
	}
}

func TestConstitutionRejectsFrontmatterTableMismatch(t *testing.T) {
	// Table says technical+business are active; frontmatter claims design too.
	path := writeFixture(t, mutate(t, "active_dimensions: [technical, business]", "active_dimensions: [technical, business, design]"))
	_, err := Constitution(path)
	if err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("expected mismatch error, got %v", err)
	}
}

func TestConstitutionRejectsNoActiveDimension(t *testing.T) {
	body := mutate(t, "active_dimensions: [technical, business]", "active_dimensions: []")
	body = mutateIn(t, body, "| Technical | yes |", "| Technical | no |")
	body = mutateIn(t, body, "| Business | yes |", "| Business | no |")
	_, err := Constitution(writeFixture(t, body))
	if err == nil || !strings.Contains(err.Error(), "no active dimension") {
		t.Fatalf("expected no-active-dimension error, got %v", err)
	}
}

func TestConstitutionRejectsNonYesNoActiveCell(t *testing.T) {
	path := writeFixture(t, mutate(t, "| Business | yes |", "| Business | maybe |"))
	_, err := Constitution(path)
	if err == nil || !strings.Contains(err.Error(), "must be yes or no") {
		t.Fatalf("expected yes/no error, got %v", err)
	}
}

func TestConstitutionRejectsUnknownDimensionRow(t *testing.T) {
	path := writeFixture(t, mutate(t, "| Financial | no | pricing |", "| Cosmic | no | pricing |"))
	_, err := Constitution(path)
	if err == nil || !strings.Contains(err.Error(), "unknown dimension") {
		t.Fatalf("expected unknown-dimension error, got %v", err)
	}
}

func TestConstitutionRejectsVersionDrift(t *testing.T) {
	path := writeFixture(t, mutate(t, `| 2026-09-25 | 1.2.0 | Added business dimension | Scope change |`, `| 2026-09-25 | 1.3.0 | Added business dimension | Scope change |`))
	_, err := Constitution(path)
	if err == nil || !strings.Contains(err.Error(), "version") {
		t.Fatalf("expected version-drift error, got %v", err)
	}
}

func TestConstitutionRejectsDuplicateRuleIDs(t *testing.T) {
	path := writeFixture(t, mutate(t, "| QG-01 | A check that cannot run must not report as passing | MUST | skip is distinct |", "| EV-01 | A check that cannot run must not report as passing | MUST | skip is distinct |"))
	_, err := Constitution(path)
	if err == nil || !strings.Contains(err.Error(), "duplicate rule ID") {
		t.Fatalf("expected duplicate-rule error, got %v", err)
	}
}

func TestConstitutionRejectsNonISODate(t *testing.T) {
	path := writeFixture(t, mutate(t, "ratified: \"2026-01-15\"", "ratified: \"15/01/2026\""))
	_, err := Constitution(path)
	if err == nil || !strings.Contains(err.Error(), "ISO") {
		t.Fatalf("expected ISO-date error, got %v", err)
	}
}

func TestConstitutionReportsMissingFileAsUnratified(t *testing.T) {
	_, err := Constitution(filepath.Join(t.TempDir(), "absent.md"))
	if err == nil || !strings.Contains(err.Error(), "/start") {
		t.Fatalf("expected unratified-guidance error, got %v", err)
	}
}

func TestConstitutionAcceptsSingleDimensionSoftwareProject(t *testing.T) {
	body := strings.ReplaceAll(constitutionFixture, "active_dimensions: [technical, business]", "active_dimensions: [technical]")
	body = strings.Replace(body, "| Business | yes | business plan | unit economics reviewed | Viability review |", "| Business | no | business plan | unit economics reviewed | Viability review |", 1)
	res, err := Constitution(writeFixture(t, body))
	if err != nil {
		t.Fatalf("Constitution: unexpected error %v", err)
	}
	if got, want := res.ActiveDimensions, []string{"technical"}; !equalStrings(got, want) {
		t.Errorf("ActiveDimensions = %v, want %v", got, want)
	}
}

// mutateIn is mutate for an already-mutated body.
func mutateIn(t *testing.T, body, old, new string) string {
	t.Helper()
	if strings.Count(body, old) != 1 {
		t.Fatalf("body no longer contains exactly one %q", old)
	}
	return strings.Replace(body, old, new, 1)
}
