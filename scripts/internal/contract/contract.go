// Package contract validates Success Contracts (instructions/success-contract.instructions.md).
//
// A contract is a markdown file with exactly one fenced json block. This package reads
// that block, checks the schema and the measurability rules, checks the freeze and the
// amendment log, and, for the gate, checks the recorded results. The freeze is anchored
// to git history when the file is committed (see anchor.go).
//
// What is verified, and what is not. The package checks form and arithmetic:
//
//   - a numeric claim compares its number and unit with the target; a binary claim
//     uses a fixed vocabulary; an evidence entry must be a path that exists (or a
//     recorded command); blocking rules and the freeze are enforced.
//
// It does not check that the metric is the right one, that a command was really run,
// that the evidence file says what the claim says, or that work did not start before
// the first commit of the contract. Those remain judgements for the phase and the human.
package contract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Severity of an issue. Errors fail the command. Warnings are reported only.
const (
	SevError   = "error"
	SevWarning = "warning"
)

// Results a gate may record for an item.
const (
	ResultMet           = "met"
	ResultNotMet        = "not_met"
	ResultNotMeasurable = "not_measurable"
)

// MaxLines is the size cap for one contract. A contract over it is too large to read at the gate.
const MaxLines = 150

// vagueWords are words that cannot be checked. The spec's section 3 lists the same words;
// a test keeps the two in step.
var vagueWords = []string{
	"good", "nice", "clean", "fast", "intuitive", "user-friendly", "user friendly", "robust",
	"high quality", "appropriate", "as needed", "improve", "better", "acceptable", "reasonable",
	"sufficient", "tbd", "todo", "approx", "approximately", "about", "roughly", "around",
	"several", "accessible", "secure", "adequate", "enough", "significant", "usable", "simple",
	"easy", "quick", "modern", "stable", "scalable",
}

// VagueWords returns the words the validator rejects in kpi, metric and target.
func VagueWords() []string {
	return append([]string(nil), vagueWords...)
}

var dimensions = map[string]bool{
	"technical": true, "design": true, "business": true, "legal": true, "marketing": true,
	"financial": true, "management": true, "product": true, "operations": true, "general": true,
}

// procedureWords mark a method as a procedure rather than a bare phrase. A method with
// none of them is rejected: "look at it" is not a check.
var procedureWords = []string{
	"test", "count", "measure", "run", "inspect", "review", "audit", "scan", "render", "query",
	"survey", "log", "script", "check", "compare", "sample", "walk", "record", "grep", "lint",
	"build", "benchmark", "load", "read", "diff", "calculate", "compute", "verify", "validate",
	"lighthouse", "axe", "plaesy", "go", "npm", "make", "pytest", "curl", "git", "find", "wc",
	"jq", "report", "observe", "interview", "walk", "reproduce",
}

var (
	reFence     = regexp.MustCompile("(?s)```json[ \t]*\r?\n(.*?)\r?\n```")
	reSlug      = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
	reItemID    = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]*$`)
	rePhase     = regexp.MustCompile(`^[a-z][a-z0-9:-]*$`)
	rePlacehold = regexp.MustCompile(`\{\{|\}\}`)
	reVague     = regexp.MustCompile(`(?i)(?:^|[^-\w])(` + vagueAlternation() + `)(?:[^\w]|$)`)
	reCodeSpan  = regexp.MustCompile("`[^`]*`")
	reNumTarget = regexp.MustCompile(`^\s*(>=|≥|<=|≤|==|=|>|<)\s*([0-9]+(?:\.[0-9]+)?)\s*([A-Za-z%]*)\s*$`)
	reNumActual = regexp.MustCompile(`^\s*((?:[0-9]{1,3}(?:,[0-9]{3})+)|[0-9]+)(\.[0-9]+)?\s*([A-Za-z%]*)\s*$`)
	reEvidence  = regexp.MustCompile(`^([A-Za-z0-9_][A-Za-z0-9_./\\-]*)(?::([0-9]+)(?:-([0-9]+))?)?$`)
	reAllWord   = regexp.MustCompile(`(?i)^(all|every)\s+\S`)
	reOfCount   = regexp.MustCompile(`^\s*([0-9]+)\s+of\s+([0-9]+)\s*$`)
)

func vagueAlternation() string {
	parts := make([]string, 0, len(vagueWords))
	for _, w := range vagueWords {
		parts = append(parts, strings.ReplaceAll(regexp.QuoteMeta(w), " ", `\s+`))
	}
	return strings.Join(parts, "|")
}

// Contract is the json block of a contract file.
type Contract struct {
	ID         string      `json:"id"`
	Phase      string      `json:"phase"`
	Dimension  string      `json:"dimension"`
	Status     string      `json:"status"`
	Frozen     *Frozen     `json:"frozen"`
	Items      []Item      `json:"items"`
	Amendments []Amendment `json:"amendments"`
	Waivers    []Waiver    `json:"waivers"`
}

// Frozen is the snapshot of targets taken before work started.
type Frozen struct {
	Items []FrozenItem `json:"items"`
}

// FrozenItem is one item's targets at freeze.
type FrozenItem struct {
	ID       string `json:"id"`
	Metric   string `json:"metric"`
	Target   string `json:"target"`
	Method   string `json:"method"`
	Blocking *bool  `json:"blocking"`
}

// Item is one measurable item.
type Item struct {
	ID       string `json:"id"`
	KPI      string `json:"kpi"`
	Metric   string `json:"metric"`
	Baseline string `json:"baseline,omitempty"`
	Target   string `json:"target"`
	Method   string `json:"method"`
	Blocking *bool  `json:"blocking"`
	Evidence string `json:"evidence,omitempty"`
	Result   string `json:"result,omitempty"`
	Actual   string `json:"actual,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

// Amendment records one change to a target after freeze.
type Amendment struct {
	Item   string `json:"item"`
	Field  string `json:"field"`
	From   string `json:"from,omitempty"`
	To     string `json:"to,omitempty"`
	Reason string `json:"reason"`
}

// Waiver is an explicit, logged exception. The only kind is "no-blocking": a contract
// with no blocking item, which needs a reason.
type Waiver struct {
	Kind   string `json:"kind"`
	Reason string `json:"reason"`
}

// Issue is one finding.
type Issue struct {
	Severity string
	Item     string
	Message  string
}

// Options configures a check.
type Options struct {
	// Gate runs the end-of-phase check: results must be recorded and blocking items met.
	Gate bool
	// RequireGit makes an unverifiable freeze (no git history) an error, not a warning.
	RequireGit bool
	// RepoRoot is where evidence paths are resolved. Empty means the contract's repo root,
	// found by walking up from the file.
	RepoRoot string
}

// Report is the outcome of checking one contract.
type Report struct {
	Path   string
	Lines  int
	Anchor Anchor
	Issues []Issue
}

// Errors returns the number of error-severity issues.
func (r Report) Errors() int {
	n := 0
	for _, i := range r.Issues {
		if i.Severity == SevError {
			n++
		}
	}
	return n
}

// Warnings returns the number of warning-severity issues.
func (r Report) Warnings() int {
	return len(r.Issues) - r.Errors()
}

// Check reads the contract at path, validates it, and verifies its freeze against git.
func Check(path string, opts Options) (Report, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Report{Path: path}, err
	}
	rep := Report{Path: path, Lines: countLines(raw)}
	if opts.RepoRoot == "" {
		opts.RepoRoot = findRoot(path)
	}
	c, err := Parse(raw)
	if err != nil {
		rep.Issues = append(rep.Issues, Issue{Severity: SevError, Message: err.Error()})
		return rep, nil
	}
	rep.Issues = Validate(c, opts)
	anchor, anchorIssues := VerifyAnchor(path, c, opts)
	rep.Anchor = anchor
	rep.Issues = append(rep.Issues, anchorIssues...)
	if rep.Lines > MaxLines {
		rep.Issues = append(rep.Issues, Issue{Severity: SevError,
			Message: fmt.Sprintf("contract is %d lines; the cap is %d: cut items or split the phase", rep.Lines, MaxLines)})
	}
	return rep, nil
}

// countLines counts lines as wc -l would, plus a final unterminated line.
func countLines(raw []byte) int {
	if len(raw) == 0 {
		return 0
	}
	n := bytes.Count(raw, []byte("\n"))
	if raw[len(raw)-1] != '\n' {
		n++
	}
	return n
}

// findRoot walks up from the contract to the repository root (the directory holding
// .git or .plaesy). It falls back to the current directory.
func findRoot(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "."
	}
	dir := filepath.Dir(abs)
	for {
		for _, marker := range []string{".git", ".plaesy"} {
			if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			wd, _ := os.Getwd()
			return wd
		}
		dir = parent
	}
}

// Parse reads the single fenced json block of a contract file.
func Parse(raw []byte) (Contract, error) {
	var c Contract
	blocks := reFence.FindAllSubmatch(raw, -1)
	switch {
	case len(blocks) == 0:
		return c, fmt.Errorf("no fenced ```json block: the contract's machine-readable part is missing")
	case len(blocks) > 1:
		return c, fmt.Errorf("found %d fenced ```json blocks: a contract has exactly one; put examples in prose, not in a second block", len(blocks))
	}
	dec := json.NewDecoder(bytes.NewReader(blocks[0][1]))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&c); err != nil {
		return c, fmt.Errorf("json block does not match the contract schema: %w", err)
	}
	var extra json.RawMessage
	if err := dec.Decode(&extra); err != io.EOF {
		return c, fmt.Errorf("json block has content after the contract object")
	}
	return c, nil
}

// Validate checks a parsed contract. Issues are returned in a stable order.
func Validate(c Contract, opts Options) []Issue {
	var out []Issue
	add := func(sev, item, format string, args ...any) {
		out = append(out, Issue{Severity: sev, Item: item, Message: fmt.Sprintf(format, args...)})
	}

	header := map[string]string{"id": c.ID, "phase": c.Phase, "dimension": c.Dimension, "status": c.Status}
	for _, k := range []string{"id", "phase", "dimension", "status"} {
		if rePlacehold.MatchString(header[k]) {
			add(SevError, "", "%s still holds a template placeholder: fill it in", k)
		}
	}
	if !reSlug.MatchString(c.ID) {
		add(SevError, "", "id %q must be lower-case letters, digits and hyphens", c.ID)
	}
	if !rePhase.MatchString(c.Phase) {
		add(SevError, "", "phase %q must be a command name such as implement or assess:technical", c.Phase)
	}
	if !dimensions[c.Dimension] {
		add(SevError, "", "dimension %q must be one of technical, design, business, legal, marketing, financial, management, product, operations, general", c.Dimension)
	}
	if c.Status != "open" && c.Status != "closed" {
		add(SevError, "", "status %q must be open or closed", c.Status)
	}
	if len(c.Items) == 0 {
		add(SevError, "", "a contract needs at least one item")
	}

	seen := map[string]bool{}
	blocking := 0
	for _, it := range c.Items {
		id := it.ID
		if !reItemID.MatchString(id) {
			add(SevError, id, "item id %q must be letters, digits and hyphens", id)
		}
		if seen[id] {
			add(SevError, id, "duplicate item id %q", id)
		}
		seen[id] = true
		checkItemFields(it, add)
		if it.Blocking != nil && *it.Blocking {
			blocking++
		}
	}
	if blocking == 0 {
		if waived(c) {
			add(SevWarning, "", "no blocking item: waived (see waivers)")
		} else {
			add(SevError, "", "at least one item must be blocking; a contract with none cannot fail its gate (or record a no-blocking waiver with a reason)")
		}
	}

	checkFreeze(c, opts, add)
	checkAmendments(c, add)
	checkWaivers(c, add)

	if opts.Gate || c.Status == "closed" {
		checkGate(c, opts, add)
	} else {
		for _, it := range c.Items {
			if it.Result != "" && !validResult(it.Result) {
				add(SevError, it.ID, "result %q must be met, not_met or not_measurable", it.Result)
			}
		}
	}
	if opts.Gate && c.Status == "open" {
		add(SevWarning, "", "status is still open: set it to closed once the gate passes")
	}
	return out
}

type adder func(sev, item, format string, args ...any)

func checkItemFields(it Item, add adder) {
	need := func(field, value string) {
		if strings.TrimSpace(value) == "" {
			add(SevError, it.ID, "%s is required", field)
		}
	}
	need("kpi", it.KPI)
	need("metric", it.Metric)
	need("target", it.Target)
	need("method", it.Method)
	if it.Blocking == nil {
		add(SevError, it.ID, "blocking is required (true or false)")
	}
	strs := []struct{ name, value string }{
		{"kpi", it.KPI}, {"metric", it.Metric}, {"target", it.Target}, {"method", it.Method},
		{"evidence", it.Evidence}, {"actual", it.Actual}, {"reason", it.Reason}, {"baseline", it.Baseline},
	}
	for _, s := range strs {
		if rePlacehold.MatchString(s.value) {
			add(SevError, it.ID, "%s still holds a template placeholder ({{...}}): fill it in", s.name)
		}
	}
	for _, f := range []struct{ name, value string }{{"kpi", it.KPI}, {"metric", it.Metric}, {"target", it.Target}} {
		if w := vagueHit(f.value); w != "" {
			add(SevError, it.ID, "%s uses the vague word %q: state a measurable value", f.name, w)
		}
	}
	if strings.TrimSpace(it.Target) != "" && !validTarget(it.Target) {
		add(SevError, it.ID, "target %q must be a comparison with a number (>= 90, <= 2 s, >= 80%%) or a binary condition (zero, none, pass, yes, no, present, absent, all X, every X)", it.Target)
	}
	if strings.TrimSpace(it.Method) != "" && !validMethod(it.Method) {
		add(SevError, it.ID, "method %q is not a check: name the command, count or procedure (at least 8 characters, with a procedure word such as test, count, measure, run, inspect)", it.Method)
	}
}

// vagueHit returns the first vague word in s, ignoring text inside code spans and
// flags (a word preceded by "-" such as --fast).
func vagueHit(s string) string {
	s = reCodeSpan.ReplaceAllStringFunc(s, func(m string) string { return strings.Repeat(" ", len(m)) })
	if m := reVague.FindStringSubmatch(s); m != nil {
		return m[1]
	}
	return ""
}

func validMethod(s string) bool {
	trimmed := strings.TrimSpace(s)
	if len(trimmed) < 8 {
		return false
	}
	lower := strings.ToLower(reCodeSpan.ReplaceAllString(trimmed, " "))
	words := regexp.MustCompile(`[a-z0-9]+`).FindAllString(lower, -1)
	for _, w := range words {
		for _, p := range procedureWords {
			if w == p {
				return true
			}
		}
	}
	return false
}

// binaryKind returns the binary family of a target, or "" if it is not binary.
func binaryKind(target string) string {
	t := strings.ToLower(strings.TrimSpace(target))
	switch t {
	case "zero", "none":
		return "zero"
	case "pass", "yes", "true", "present":
		return "pass"
	case "no", "false", "absent":
		return "absent"
	}
	if reAllWord.MatchString(t) {
		return "all"
	}
	return ""
}

// validTarget accepts a numeric comparison with a single clause, or a binary condition.
func validTarget(t string) bool {
	if reNumTarget.MatchString(t) {
		return true
	}
	return binaryKind(t) != ""
}

// binarySatisfies reports whether the recorded actual value matches a binary target.
// A value outside the vocabulary never satisfies a binary target.
func binarySatisfies(target, actual string) bool {
	a := strings.ToLower(strings.TrimSpace(actual))
	switch binaryKind(target) {
	case "zero":
		return a == "0" || a == "none"
	case "pass":
		return a == "pass" || a == "yes" || a == "true" || a == "present"
	case "absent":
		return a == "0" || a == "none" || a == "no" || a == "false" || a == "absent"
	case "all":
		if a == "pass" {
			return true
		}
		m := reOfCount.FindStringSubmatch(a)
		return m != nil && m[1] == m[2] && m[1] != "0"
	}
	return false
}

// numericResult is the outcome of comparing a numeric claim with its target.
type numericResult struct {
	known   bool   // the target is numeric
	ok      bool   // the claim satisfies the target
	problem string // why the claim cannot be judged, when it cannot
}

var timeUnits = map[string]float64{"ms": 1, "s": 1000, "sec": 1000, "secs": 1000, "min": 60000, "mins": 60000, "h": 3600000, "hr": 3600000, "hrs": 3600000}

// numericSatisfies compares a numeric actual with a numeric target. Units must match,
// with one exception: time units convert (ms, s, min, h).
func numericSatisfies(target, actual string) numericResult {
	t := reNumTarget.FindStringSubmatch(strings.ReplaceAll(target, "≥", ">="))
	if t == nil {
		return numericResult{}
	}
	op := t[1]
	want, _ := strconv.ParseFloat(t[2], 64)
	tUnit := t[3]

	a := reNumActual.FindStringSubmatch(actual)
	if a == nil {
		return numericResult{known: true, problem: "actual must be a number with the target's unit, for example 1.5 s for a target in s"}
	}
	got, err := strconv.ParseFloat(strings.ReplaceAll(a[1], ",", "")+a[2], 64)
	if err != nil {
		return numericResult{known: true, problem: "actual is not a readable number"}
	}
	aUnit := a[3]
	if tUnit != aUnit {
		tf, tTime := timeUnits[strings.ToLower(tUnit)]
		af, aTime := timeUnits[strings.ToLower(aUnit)]
		if !(tTime && aTime) {
			return numericResult{known: true, problem: fmt.Sprintf("units differ: target %q, actual %q", tUnit, aUnit)}
		}
		got = got * af
		want = want * tf
	}
	return numericResult{known: true, ok: compare(op, got, want)}
}

func compare(op string, got, want float64) bool {
	switch op {
	case ">=", "≥":
		return got >= want
	case "<=", "≤":
		return got <= want
	case ">":
		return got > want
	case "<":
		return got < want
	default: // "=" or "=="
		return got == want
	}
}

// checkFreeze compares the current items with the frozen snapshot.
func checkFreeze(c Contract, opts Options, add adder) {
	if c.Frozen == nil {
		if opts.Gate || c.Status == "closed" {
			add(SevError, "", "the contract has no frozen snapshot: freeze it before the first action")
		} else {
			add(SevWarning, "", "contract is not frozen yet: freeze it before the first action")
		}
		return
	}
	current := map[string]Item{}
	for _, it := range c.Items {
		current[it.ID] = it
	}
	frozen := map[string]FrozenItem{}
	for i, f := range c.Frozen.Items {
		if !reItemID.MatchString(f.ID) {
			add(SevError, f.ID, "frozen item %d needs a valid id", i+1)
			continue
		}
		if _, dup := frozen[f.ID]; dup {
			add(SevError, f.ID, "duplicate frozen item id %q", f.ID)
		}
		if f.Metric == "" || f.Target == "" || f.Method == "" || f.Blocking == nil {
			add(SevError, f.ID, "frozen item needs metric, target, method and blocking")
		}
		frozen[f.ID] = f
	}
	for _, f := range c.Frozen.Items {
		cur, ok := current[f.ID]
		if !ok {
			if f.Blocking != nil && *f.Blocking {
				add(SevError, f.ID, "blocking item removed after freeze: not allowed; stop and ask the human")
			} else if !hasAmendment(c, f.ID, "removed") {
				add(SevError, f.ID, "item removed after freeze with no amendment (field \"removed\")")
			}
			continue
		}
		if f.Blocking != nil && *f.Blocking && cur.Blocking != nil && !*cur.Blocking {
			add(SevError, f.ID, "blocking was loosened from true to false after freeze: not allowed; stop and ask the human")
		} else if f.Blocking != nil && cur.Blocking != nil && *f.Blocking != *cur.Blocking &&
			!hasChangeAmendment(c, f.ID, "blocking", boolText(*cur.Blocking)) {
			add(SevError, f.ID, "blocking changed after freeze with no amendment recorded")
		}
		changed := []struct{ field, from, to string }{
			{"metric", f.Metric, cur.Metric},
			{"target", f.Target, cur.Target},
			{"method", f.Method, cur.Method},
		}
		for _, ch := range changed {
			if ch.from != ch.to && !hasChangeAmendment(c, f.ID, ch.field, ch.to) {
				add(SevError, f.ID, "%s changed after freeze (%q -> %q) with no amendment recorded", ch.field, ch.from, ch.to)
			}
		}
	}
	for _, it := range c.Items {
		if _, ok := frozen[it.ID]; !ok && !hasAmendment(c, it.ID, "added") {
			add(SevError, it.ID, "item was added after freeze with no amendment (field \"added\")")
		}
	}
}

func boolText(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func hasAmendment(c Contract, item, field string) bool {
	for _, a := range c.Amendments {
		if a.Item == item && a.Field == field && strings.TrimSpace(a.Reason) != "" {
			return true
		}
	}
	return false
}

// hasChangeAmendment reports whether a change of item.field to value is recorded, with a
// reason, and with a from value that matches the frozen snapshot.
func hasChangeAmendment(c Contract, item, field, to string) bool {
	for _, a := range c.Amendments {
		if a.Item == item && a.Field == field && a.To == to && strings.TrimSpace(a.Reason) != "" {
			return true
		}
	}
	return false
}

func checkAmendments(c Contract, add adder) {
	frozen := map[string]FrozenItem{}
	if c.Frozen != nil {
		for _, f := range c.Frozen.Items {
			frozen[f.ID] = f
		}
	}
	for i, a := range c.Amendments {
		label := fmt.Sprintf("amendment %d (%s.%s)", i+1, a.Item, a.Field)
		if strings.TrimSpace(a.Item) == "" || strings.TrimSpace(a.Field) == "" {
			add(SevError, a.Item, "amendment %d needs item and field", i+1)
			continue
		}
		if strings.TrimSpace(a.Reason) == "" {
			add(SevError, a.Item, "%s needs a reason", label)
		}
		if f, ok := frozen[a.Item]; ok {
			var fromWant string
			switch a.Field {
			case "metric":
				fromWant = f.Metric
			case "target":
				fromWant = f.Target
			case "method":
				fromWant = f.Method
			}
			if fromWant != "" && a.From != fromWant {
				add(SevError, a.Item, "%s: from %q does not match the frozen value %q", label, a.From, fromWant)
			}
		}
	}
}

func checkWaivers(c Contract, add adder) {
	for i, w := range c.Waivers {
		if w.Kind != "no-blocking" {
			add(SevError, "", "waiver %d: kind %q is not supported (only no-blocking)", i+1, w.Kind)
		}
		if strings.TrimSpace(w.Reason) == "" {
			add(SevError, "", "waiver %d needs a reason", i+1)
		}
	}
}

func waived(c Contract) bool {
	for _, w := range c.Waivers {
		if w.Kind == "no-blocking" && strings.TrimSpace(w.Reason) != "" {
			return true
		}
	}
	return false
}

func validResult(r string) bool {
	return r == ResultMet || r == ResultNotMet || r == ResultNotMeasurable
}

// checkGate verifies the recorded results. A blocking item that is not met, or not
// measurable, fails the gate. So does a contract whose items are all non-blocking and
// one of them is not met: nothing in it was able to pass or fail the phase.
func checkGate(c Contract, opts Options, add adder) {
	if c.Frozen == nil {
		add(SevError, "", "gate: no frozen snapshot, so the targets cannot be compared with the results")
	}
	anyNotMet := false
	for _, it := range c.Items {
		if !validResult(it.Result) {
			add(SevError, it.ID, "gate: result must be recorded as met, not_met or not_measurable (got %q)", it.Result)
			continue
		}
		blocking := it.Blocking != nil && *it.Blocking
		switch it.Result {
		case ResultMet:
			checkMet(it, opts, add)
		case ResultNotMet:
			anyNotMet = true
			if !evidenceShaped(it.Evidence) {
				add(SevError, it.ID, "gate: not_met needs evidence (a path or cmd:<command>)")
			} else if !evidenceExists(it.Evidence, opts.RepoRoot) {
				add(SevError, it.ID, "gate: evidence path %q does not exist under %s", it.Evidence, opts.RepoRoot)
			}
			if blocking {
				add(SevError, it.ID, "gate: blocking item not met: the phase must not report completion")
			}
		case ResultNotMeasurable:
			if strings.TrimSpace(it.Reason) == "" {
				add(SevError, it.ID, "gate: not_measurable needs a reason")
			}
			if blocking {
				add(SevError, it.ID, "gate: blocking item not measurable: the phase must not report completion")
			}
		}
	}
	if anyNotMet && !anyBlocking(c) && !waived(c) {
		add(SevError, "", "gate: an item is not met and no item is blocking: the contract cannot fail the phase; make the failing item blocking or record a waiver")
	}
	if anyNotMet && !anyBlocking(c) && waived(c) {
		add(SevWarning, "", "gate: an item is not met; the contract is waived as having no blocking item")
	}
}

func anyBlocking(c Contract) bool {
	for _, it := range c.Items {
		if it.Blocking != nil && *it.Blocking {
			return true
		}
	}
	return false
}

// checkMet checks a claim of met against its target, and its evidence.
func checkMet(it Item, opts Options, add adder) {
	if strings.TrimSpace(it.Actual) == "" {
		add(SevError, it.ID, "gate: met needs the actual measured value")
	}
	if !evidenceShaped(it.Evidence) {
		add(SevError, it.ID, "gate: met needs evidence: a repo-relative path (optionally :line) that exists, or cmd:<the command that was run>")
	} else if !evidenceExists(it.Evidence, opts.RepoRoot) {
		add(SevError, it.ID, "gate: evidence path %q does not exist under %s", it.Evidence, opts.RepoRoot)
	}
	switch {
	case binaryKind(it.Target) != "":
		if strings.TrimSpace(it.Actual) != "" && !binarySatisfies(it.Target, it.Actual) {
			add(SevError, it.ID, "gate: recorded met, but actual %q does not satisfy the binary target %q (use 0 or none for zero; pass, yes, present; or n of n for all)", it.Actual, it.Target)
		}
	default:
		r := numericSatisfies(it.Target, it.Actual)
		switch {
		case r.problem != "":
			add(SevError, it.ID, "gate: %s", r.problem)
		case r.known && !r.ok:
			add(SevError, it.ID, "gate: recorded met, but actual %q does not satisfy target %q", it.Actual, it.Target)
		}
	}
}

// evidenceShaped reports whether evidence is a pointer: a repo-relative path with an
// optional :line or :start-end, or cmd: followed by the command that was run. It does
// not check that the file exists; see evidenceExists for that.
func evidenceShaped(ev string) bool {
	ev = strings.TrimSpace(ev)
	if strings.HasPrefix(ev, "cmd:") {
		return len(strings.TrimSpace(strings.TrimPrefix(ev, "cmd:"))) >= 3
	}
	return reEvidence.MatchString(ev)
}

// evidenceExists reports whether a path-shaped evidence entry names a file under root.
// A cmd: entry and a non-path entry cannot be checked, and it returns true for them.
func evidenceExists(ev, root string) bool {
	ev = strings.TrimSpace(ev)
	if strings.HasPrefix(ev, "cmd:") || !reEvidence.MatchString(ev) {
		return true
	}
	m := reEvidence.FindStringSubmatch(ev)
	p := filepath.Join(root, filepath.FromSlash(strings.ReplaceAll(m[1], `\`, "/")))
	_, err := os.Stat(p)
	return err == nil
}
