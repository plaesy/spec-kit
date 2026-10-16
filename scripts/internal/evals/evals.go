// Package evals runs behavioural cases for instruction files.
//
// A case file (evals/cases/<name>.json) names one instruction file and holds
// three kinds of case:
//
//   - static: the instruction text must contain some phrases (requires) and must
//     not contain others (forbids). Offline and deterministic, so it runs in
//     go test and in CI.
//   - outputs: a fixture directory of produced frontend files is linted with
//     designlint. expect is "pass" (no FAIL) or "fail"; rules lists the rule
//     identifiers that must appear in the findings.
//   - briefs: a model writes an output for each brief. This needs an LLM, so it
//     is optional. The command in PLAESY_EVAL_RUNNER is run as
//     `<command> <brief-id> <output-dir>` with the brief text in PLAESY_EVAL_BRIEF,
//     and the output is linted. Without a runner, every brief is reported SKIP,
//     never PASS.
package evals

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/plaesy/spec-kit/internal/designlint"
)

// CasesDir and FixturesDir are relative to the repository root.
const (
	CasesDir    = "evals/cases"
	FixturesDir = "evals/fixtures"
)

// Status of one outcome.
const (
	StatusPass = "PASS"
	StatusFail = "FAIL"
	StatusSkip = "SKIP"
)

// File is one case file. Unknown keys are rejected, so a misspelt key cannot
// silently drop a check.
type File struct {
	Instruction string   `json:"instruction"`
	Static      []Static `json:"static"`
	Outputs     []Output `json:"outputs"`
	Briefs      []Brief  `json:"briefs"`
}

// Static is a phrase check on the instruction text. Matching is case-insensitive.
type Static struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	Requires    []string `json:"requires,omitempty"`
	Forbids     []string `json:"forbids,omitempty"`
}

// Output is a lint expectation on a fixture directory.
type Output struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	Fixture     string   `json:"fixture"`
	Expect      string   `json:"expect"`
	Rules       []string `json:"rules,omitempty"`
}

// Brief is one prompt for the optional runner.
type Brief struct {
	ID    string `json:"id"`
	Brief string `json:"brief"`
}

// Outcome is the result of one case.
type Outcome struct {
	File   string
	Kind   string
	ID     string
	Status string
	Detail string
}

// Options configures a run.
type Options struct {
	// Runner is the command for brief cases. Empty means briefs are skipped.
	Runner string
}

// Load reads and validates every case file under root/CasesDir.
func Load(root string) (map[string]File, error) {
	dir := filepath.Join(root, filepath.FromSlash(CasesDir))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	out := map[string]File{}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		path := filepath.Join(dir, e.Name())
		raw, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil, readErr
		}
		var f File
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.DisallowUnknownFields()
		if decErr := dec.Decode(&f); decErr != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), decErr)
		}
		if err := validate(f); err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		out[e.Name()] = f
	}
	return out, nil
}

func validate(f File) error {
	if f.Instruction == "" {
		return fmt.Errorf("instruction is required")
	}
	for _, s := range f.Static {
		if s.ID == "" {
			return fmt.Errorf("static case without id")
		}
		if len(s.Requires) == 0 && len(s.Forbids) == 0 {
			return fmt.Errorf("static case %q has neither requires nor forbids", s.ID)
		}
	}
	for _, o := range f.Outputs {
		if o.ID == "" || o.Fixture == "" {
			return fmt.Errorf("output case needs id and fixture")
		}
		if o.Expect != "pass" && o.Expect != "fail" {
			return fmt.Errorf("output case %q: expect must be pass or fail, got %q", o.ID, o.Expect)
		}
	}
	for _, b := range f.Briefs {
		if b.ID == "" || b.Brief == "" {
			return fmt.Errorf("brief needs id and brief text")
		}
	}
	return nil
}

// Run evaluates every case file under root and returns one outcome per case.
// The order is stable: file name, then kind, then case order.
func Run(root string, opts Options) ([]Outcome, error) {
	files, err := Load(root)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(files))
	for n := range files {
		names = append(names, n)
	}
	sort.Strings(names)

	var out []Outcome
	for _, name := range names {
		f := files[name]
		out = append(out, runStatic(root, name, f)...)
		out = append(out, runOutputs(root, name, f)...)
		out = append(out, runBriefs(root, name, f, opts)...)
	}
	return out, nil
}

func runStatic(root, name string, f File) []Outcome {
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(f.Instruction)))
	if err != nil {
		return []Outcome{{File: name, Kind: "static", ID: "(instruction)", Status: StatusFail, Detail: err.Error()}}
	}
	text := strings.ToLower(string(raw))
	var out []Outcome
	for _, s := range f.Static {
		o := Outcome{File: name, Kind: "static", ID: s.ID, Status: StatusPass, Detail: s.Description}
		var problems []string
		for _, want := range s.Requires {
			if !strings.Contains(text, strings.ToLower(want)) {
				problems = append(problems, fmt.Sprintf("missing %q", want))
			}
		}
		for _, bad := range s.Forbids {
			if strings.Contains(text, strings.ToLower(bad)) {
				problems = append(problems, fmt.Sprintf("contains forbidden %q", bad))
			}
		}
		if len(problems) > 0 {
			o.Status = StatusFail
			o.Detail = strings.Join(problems, "; ")
		}
		out = append(out, o)
	}
	return out
}

func runOutputs(root, name string, f File) []Outcome {
	var out []Outcome
	for _, c := range f.Outputs {
		o := Outcome{File: name, Kind: "output", ID: c.ID, Status: StatusPass, Detail: c.Description}
		res, err := designlint.Lint([]string{filepath.Join(root, filepath.FromSlash(c.Fixture))}, designlint.Options{})
		if err != nil {
			o.Status = StatusFail
			o.Detail = err.Error()
			out = append(out, o)
			continue
		}
		o.Status, o.Detail = judge(c.Expect, c.Rules, res)
		if o.Detail == "" {
			o.Detail = c.Description
		}
		out = append(out, o)
	}
	return out
}

// judge applies one expectation to a lint result. expect "pass" allows WARN but
// not FAIL. expect "fail" needs at least one FAIL. Every rule in want must appear
// in the findings at any severity.
func judge(expect string, want []string, res designlint.Result) (string, string) {
	var problems []string
	if expect == "pass" && res.Failed() {
		problems = append(problems, "expected no FAIL, got "+failRules(res))
	}
	if expect == "fail" && !res.Failed() {
		problems = append(problems, "expected a FAIL, got none")
	}
	for _, rule := range want {
		if !hasRule(res, rule) {
			problems = append(problems, "rule "+rule+" did not appear")
		}
	}
	if len(problems) > 0 {
		return StatusFail, strings.Join(problems, "; ")
	}
	return StatusPass, ""
}

func hasRule(res designlint.Result, rule string) bool {
	for _, f := range res.Findings {
		if f.Rule == rule {
			return true
		}
	}
	return false
}

func failRules(res designlint.Result) string {
	seen := map[string]bool{}
	var rules []string
	for _, f := range res.Findings {
		if f.Severity == designlint.SevFail && !seen[f.Rule] {
			seen[f.Rule] = true
			rules = append(rules, f.Rule)
		}
	}
	return strings.Join(rules, ", ")
}

func runBriefs(root, name string, f File, opts Options) []Outcome {
	var out []Outcome
	for _, b := range f.Briefs {
		o := Outcome{File: name, Kind: "brief", ID: b.ID, Status: StatusSkip, Detail: "no PLAESY_EVAL_RUNNER set; brief not run"}
		if opts.Runner == "" {
			out = append(out, o)
			continue
		}
		dir, err := os.MkdirTemp("", "plaesy-eval-*")
		if err != nil {
			o.Status, o.Detail = StatusFail, err.Error()
			out = append(out, o)
			continue
		}
		o.Status, o.Detail = runOne(opts.Runner, b, dir)
		out = append(out, o)
		_ = os.RemoveAll(dir)
	}
	return out
}

// runOne calls the runner, then lints what it wrote. The runner's stdout and
// stderr are kept in the detail when it fails, so the cause is visible.
func runOne(runner string, b Brief, dir string) (string, string) {
	fields := strings.Fields(runner)
	if len(fields) == 0 {
		return StatusFail, "PLAESY_EVAL_RUNNER is empty"
	}
	args := append(fields[1:], b.ID, dir)
	cmd := exec.Command(fields[0], args...)
	cmd.Env = append(os.Environ(), "PLAESY_EVAL_BRIEF="+b.Brief)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return StatusFail, fmt.Sprintf("runner failed: %v: %s", err, strings.TrimSpace(buf.String()))
	}
	res, err := designlint.Lint([]string{dir}, designlint.Options{})
	if err != nil {
		return StatusFail, err.Error()
	}
	if res.Failed() {
		return StatusFail, "output has FAIL findings: " + failRules(res)
	}
	return StatusPass, "output lints clean"
}

// Summary counts outcomes by status.
func Summary(outcomes []Outcome) (pass, fail, skip int) {
	for _, o := range outcomes {
		switch o.Status {
		case StatusPass:
			pass++
		case StatusFail:
			fail++
		case StatusSkip:
			skip++
		}
	}
	return pass, fail, skip
}
