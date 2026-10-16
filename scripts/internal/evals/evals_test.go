package evals

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot is the repository root, from scripts/internal/evals.
const repoRoot = "../../.."

func TestRepositoryEvalsPass(t *testing.T) {
	outcomes, err := Run(repoRoot, Options{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(outcomes) == 0 {
		t.Fatal("no eval cases found under evals/cases")
	}
	for _, o := range outcomes {
		if o.Status == StatusFail {
			t.Errorf("%s %s/%s: %s", o.File, o.Kind, o.ID, o.Detail)
		}
	}
}

func TestEveryCaseFileHasStaticAndOutputCoverage(t *testing.T) {
	files, err := Load(repoRoot)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	for _, name := range []string{"frontend-design.json", "redesign.json"} {
		f, ok := files[name]
		if !ok {
			t.Fatalf("missing case file %s", name)
		}
		if len(f.Static) == 0 || len(f.Outputs) == 0 {
			t.Errorf("%s needs static and output cases, has %d and %d", name, len(f.Static), len(f.Outputs))
		}
	}
}

func TestBriefsSkipWithoutRunner(t *testing.T) {
	outcomes, err := Run(repoRoot, Options{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	briefs := 0
	for _, o := range outcomes {
		if o.Kind != "brief" {
			continue
		}
		briefs++
		if o.Status != StatusSkip {
			t.Errorf("brief %s without a runner must be SKIP, got %s", o.ID, o.Status)
		}
	}
	if briefs == 0 {
		t.Fatal("expected at least one brief case to be declared")
	}
}

func TestStaticDetectsMissingAndForbiddenText(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "instructions/x.md"), "Use picsum.photos for pictures.\n")
	mustWrite(t, filepath.Join(root, CasesDir, "x.json"), `{
  "instruction": "instructions/x.md",
  "static": [
    {"id": "needs-widths", "description": "d", "requires": ["1440", "390"]},
    {"id": "no-picsum", "description": "d", "forbids": ["picsum"]}
  ]
}`)
	outcomes, err := Run(root, Options{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	status := map[string]string{}
	for _, o := range outcomes {
		status[o.ID] = o.Status
	}
	if status["needs-widths"] != StatusFail {
		t.Errorf("missing required phrase must FAIL, got %s", status["needs-widths"])
	}
	if status["no-picsum"] != StatusFail {
		t.Errorf("forbidden phrase must FAIL, got %s", status["no-picsum"])
	}
}

func TestLoadRejectsUnknownKeysAndBadExpect(t *testing.T) {
	cases := map[string]string{
		"typo key":     `{"instruction": "a.md", "statik": []}`,
		"bad expect":   `{"instruction": "a.md", "outputs": [{"id": "o", "fixture": "f", "expect": "maybe"}]}`,
		"empty static": `{"instruction": "a.md", "static": [{"id": "s", "description": "d"}]}`,
	}
	for name, body := range cases {
		root := t.TempDir()
		mustWrite(t, filepath.Join(root, CasesDir, "bad.json"), body)
		if _, err := Load(root); err == nil {
			t.Errorf("%s: Load should reject this case file", name)
		}
	}
}

func TestOutputExpectationsFollowLintResult(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "fixtures/bad/index.html"), `<img src="https://picsum.photos/x.jpg">`)
	mustWrite(t, filepath.Join(root, "fixtures/ok/index.html"), `<p>Hello</p>`)
	mustWrite(t, filepath.Join(root, CasesDir, "x.json"), `{
  "instruction": "instructions/none.md",
  "outputs": [
    {"id": "bad-fails-with-rule", "description": "d", "fixture": "fixtures/bad", "expect": "fail", "rules": ["DL-IMG-HOST"]},
    {"id": "ok-passes", "description": "d", "fixture": "fixtures/ok", "expect": "pass"},
    {"id": "ok-wrongly-expected-fail", "description": "d", "fixture": "fixtures/ok", "expect": "fail"}
  ]
}`)
	outcomes, err := Run(root, Options{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	status := map[string]string{}
	for _, o := range outcomes {
		status[o.ID] = o.Status
	}
	if status["bad-fails-with-rule"] != StatusPass {
		t.Errorf("expected FAIL with DL-IMG-HOST should PASS as an expectation, got %s", status["bad-fails-with-rule"])
	}
	if status["ok-passes"] != StatusPass {
		t.Errorf("clean fixture expect pass, got %s", status["ok-passes"])
	}
	if status["ok-wrongly-expected-fail"] != StatusFail {
		t.Errorf("clean fixture expect fail must FAIL the case, got %s", status["ok-wrongly-expected-fail"])
	}
}

// TestEvalRunnerHelper is not a test. It is the runner the brief tests call in a
// child process: it writes an output into the directory it is given.
func TestEvalRunnerHelper(t *testing.T) {
	mode := os.Getenv("PLAESY_EVAL_HELPER")
	if mode == "" {
		return
	}
	dir := os.Args[len(os.Args)-1]
	body := `<main><h1>Clean</h1></main>`
	if mode == "bad" {
		body = `<img src="https://placehold.co/600x400">`
	}
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(body), 0o644); err != nil {
		os.Exit(3)
	}
	os.Exit(0)
}

func TestRunnerProducesPassOrFailPerBrief(t *testing.T) {
	self := os.Args[0]
	runner := self + " -test.run=^TestEvalRunnerHelper$ --"
	cases := map[string]string{"clean": StatusPass, "bad": StatusFail}
	for mode, want := range cases {
		t.Setenv("PLAESY_EVAL_HELPER", mode)
		got, detail := runOne(runner, Brief{ID: "b", Brief: "a page"}, t.TempDir())
		if got != want {
			t.Errorf("mode %s: status %s (%s), want %s", mode, got, detail, want)
		}
	}
}

func TestRunnerFailureIsReported(t *testing.T) {
	status, detail := runOne(os.Args[0]+" -test.run=^$ -nonexistent-flag", Brief{ID: "b", Brief: "x"}, t.TempDir())
	if status != StatusFail || !strings.Contains(detail, "runner failed") {
		t.Fatalf("a failing runner must FAIL with its error, got %s %q", status, detail)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
