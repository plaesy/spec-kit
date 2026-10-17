// Package contract validates Success Contracts (instructions/success-contract.instructions.md).
//
// A contract is a markdown file with one fenced json block. This package reads that
// block, checks the schema and the measurability rules, checks the freeze and the
// amendment log against the frozen snapshot, and, for the gate, checks the recorded
// results. Every check is deterministic and offline.
//
// The package judges only what can be judged mechanically: that a target is numeric
// or binary, that a numeric claim matches its actual value, that a frozen target did
// not change silently. It cannot tell whether a metric is the right one. That remains
// a judgement for the phase and for the human.
package contract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
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

// MaxLinesWarn is the length above which a contract is too large to read at the gate.
const MaxLinesWarn = 150

var dimensions = map[string]bool{
	"technical": true, "design": true, "business": true, "legal": true, "marketing": true,
	"financial": true, "management": true, "product": true, "operations": true, "general": true,
}

var (
	reFence     = regexp.MustCompile("(?s)```json[ \t]*\r?\n(.*?)\r?\n```")
	reSlug      = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
	reItemID    = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]*$`)
	rePhase     = regexp.MustCompile(`^[a-z][a-z0-9:-]*$`)
	reVague     = regexp.MustCompile(`(?i)\b(good|nice|clean|fast|intuitive|user[- ]friendly|robust|high quality|appropriate|as needed|improve|better|acceptable|reasonable|sufficient|tbd|todo)\b`)
	reBinary    = regexp.MustCompile(`(?i)\b(zero|none|all|every|pass|present|absent|yes|no|true|false|blocked)\b`)
	reDigit     = regexp.MustCompile(`[0-9]`)
	reNumTarget = regexp.MustCompile(`^\s*(>=|<=|==|=|>|<)\s*([0-9]+(?:\.[0-9]+)?)`)
	reNumActual = regexp.MustCompile(`^\s*([0-9]+(?:\.[0-9]+)?)`)
)

// Contract is the json block of a contract file.
type Contract struct {
	ID         string      `json:"id"`
	Phase      string      `json:"phase"`
	Dimension  string      `json:"dimension"`
	Status     string      `json:"status"`
	Frozen     *Frozen     `json:"frozen,omitempty"`
	Items      []Item      `json:"items"`
	Amendments []Amendment `json:"amendments,omitempty"`
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

// Issue is one finding.
type Issue struct {
	Severity string
	Item     string
	Message  string
}

// Report is the outcome of checking one contract.
type Report struct {
	Path   string
	Lines  int
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

// Check reads the contract at path and validates it. With gate true it also requires
// every result to be recorded and every blocking item to be met.
func Check(path string, gate bool) (Report, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Report{Path: path}, err
	}
	rep := Report{Path: path, Lines: bytes.Count(raw, []byte("\n")) + 1}
	c, err := Parse(raw)
	if err != nil {
		rep.Issues = append(rep.Issues, Issue{Severity: SevError, Message: err.Error()})
		return rep, nil
	}
	rep.Issues = Validate(c, gate)
	if rep.Lines > MaxLinesWarn {
		rep.Issues = append(rep.Issues, Issue{Severity: SevWarning,
			Message: fmt.Sprintf("contract is %d lines; over %d it is too large to read at the gate: cut items", rep.Lines, MaxLinesWarn)})
	}
	return rep, nil
}

// Parse reads the first fenced json block of a contract file.
func Parse(raw []byte) (Contract, error) {
	var c Contract
	m := reFence.FindSubmatch(raw)
	if m == nil {
		return c, fmt.Errorf("no fenced ```json block: the contract's machine-readable part is missing")
	}
	dec := json.NewDecoder(bytes.NewReader(m[1]))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&c); err != nil {
		return c, fmt.Errorf("json block does not match the contract schema: %w", err)
	}
	return c, nil
}

// Validate checks a parsed contract. Issues are returned in a stable order.
func Validate(c Contract, gate bool) []Issue {
	var out []Issue
	add := func(sev, item, format string, args ...any) {
		out = append(out, Issue{Severity: sev, Item: item, Message: fmt.Sprintf(format, args...)})
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
	}

	checkFreeze(c, gate, add)
	checkAmendments(c, add)
	if gate {
		checkGate(c, add)
	} else {
		for _, it := range c.Items {
			if it.Result != "" && !validResult(it.Result) {
				add(SevError, it.ID, "result %q must be met, not_met or not_measurable", it.Result)
			}
		}
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
	// A slice, not a map: issues must come out in the same order on every run.
	fields := []struct{ name, value string }{
		{"kpi", it.KPI}, {"metric", it.Metric}, {"target", it.Target}, {"method", it.Method},
	}
	for _, f := range fields {
		if m := reVague.FindString(f.value); m != "" {
			add(SevError, it.ID, "%s uses the vague word %q: state a measurable value", f.name, m)
		}
	}
	if strings.TrimSpace(it.Target) != "" && !reDigit.MatchString(it.Target) && !reBinary.MatchString(it.Target) {
		add(SevError, it.ID, "target %q is neither a number with a comparison nor a binary condition (zero, none, all, pass...)", it.Target)
	}
	if strings.TrimSpace(it.Method) != "" && len(strings.TrimSpace(it.Method)) < 8 {
		add(SevError, it.ID, "method %q is too short to be a check: name the command, count or procedure", it.Method)
	}
}

// checkFreeze compares the current items with the frozen snapshot.
func checkFreeze(c Contract, gate bool, add adder) {
	if c.Frozen == nil {
		if gate || c.Status == "closed" {
			add(SevError, "", "the contract was not frozen before work started: a frozen snapshot is required")
		} else {
			add(SevWarning, "", "contract is not frozen yet: freeze it before the first action")
		}
		return
	}
	current := map[string]Item{}
	for _, it := range c.Items {
		current[it.ID] = it
	}
	frozen := map[string]bool{}
	for _, f := range c.Frozen.Items {
		frozen[f.ID] = true
		cur, ok := current[f.ID]
		if !ok {
			if !hasAmendment(c, f.ID, "removed") {
				add(SevError, f.ID, "item was removed after freeze with no amendment (field \"removed\")")
			}
			continue
		}
		if f.Blocking != nil && *f.Blocking && cur.Blocking != nil && !*cur.Blocking {
			add(SevError, f.ID, "blocking was loosened from true to false after freeze: not allowed, stop and ask the human")
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
		if f.Blocking != nil && cur.Blocking != nil && *f.Blocking != *cur.Blocking && !hasChangeAmendment(c, f.ID, "blocking", boolText(*cur.Blocking)) {
			add(SevError, f.ID, "blocking changed after freeze with no amendment recorded")
		}
	}
	for _, it := range c.Items {
		if !frozen[it.ID] && !hasAmendment(c, it.ID, "added") {
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

// hasChangeAmendment reports whether a change to item.field to value is recorded with a reason.
func hasChangeAmendment(c Contract, item, field, to string) bool {
	for _, a := range c.Amendments {
		if a.Item == item && a.Field == field && a.To == to && strings.TrimSpace(a.Reason) != "" {
			return true
		}
	}
	return false
}

func checkAmendments(c Contract, add adder) {
	for i, a := range c.Amendments {
		if strings.TrimSpace(a.Item) == "" || strings.TrimSpace(a.Field) == "" {
			add(SevError, a.Item, "amendment %d needs item and field", i+1)
		}
		if strings.TrimSpace(a.Reason) == "" {
			add(SevError, a.Item, "amendment %d (%s.%s) needs a reason", i+1, a.Item, a.Field)
		}
	}
}

func validResult(r string) bool {
	return r == ResultMet || r == ResultNotMet || r == ResultNotMeasurable
}

// checkGate verifies the recorded results. A blocking item that is not met, or not
// measurable, fails the gate.
func checkGate(c Contract, add adder) {
	for _, it := range c.Items {
		if !validResult(it.Result) {
			add(SevError, it.ID, "gate: result must be recorded as met, not_met or not_measurable (got %q)", it.Result)
			continue
		}
		blocking := it.Blocking != nil && *it.Blocking
		switch it.Result {
		case ResultMet:
			if strings.TrimSpace(it.Actual) == "" {
				add(SevError, it.ID, "gate: met needs the actual measured value")
			}
			if strings.TrimSpace(it.Evidence) == "" {
				add(SevError, it.ID, "gate: met needs evidence (where the proof lives)")
			}
			if ok, known := numericSatisfies(it.Target, it.Actual); known && !ok {
				add(SevError, it.ID, "gate: recorded met, but actual %q does not satisfy target %q", it.Actual, it.Target)
			}
		case ResultNotMet:
			if strings.TrimSpace(it.Evidence) == "" {
				add(SevError, it.ID, "gate: not_met needs evidence")
			}
			if ok, known := numericSatisfies(it.Target, it.Actual); known && ok {
				add(SevWarning, it.ID, "gate: recorded not_met, but actual %q satisfies target %q: check the record", it.Actual, it.Target)
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
}

// numericSatisfies reports whether actual satisfies a numeric target such as ">= 90".
// known is false when the target or the actual value is not numeric, in which case
// the validator cannot judge the claim and does not guess.
func numericSatisfies(target, actual string) (ok, known bool) {
	t := reNumTarget.FindStringSubmatch(target)
	a := reNumActual.FindStringSubmatch(actual)
	if t == nil || a == nil {
		return false, false
	}
	want, err1 := strconv.ParseFloat(t[2], 64)
	got, err2 := strconv.ParseFloat(a[1], 64)
	if err1 != nil || err2 != nil {
		return false, false
	}
	switch t[1] {
	case ">=":
		return got >= want, true
	case "<=":
		return got <= want, true
	case ">":
		return got > want, true
	case "<":
		return got < want, true
	default: // "=" or "=="
		return got == want, true
	}
}
