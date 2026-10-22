// Package expected judges a check's output against the value it must reach. An expected value is either a
// comparison with one number (">= 9", "<= 2 s") or a binary condition ("pass", "zero",
// "all"). The rules are the same wherever a checklist is run.
package expected

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Results a check can end with.
const (
	ResultMet           = "met"
	ResultNotMet        = "not_met"
	ResultNotMeasurable = "not_measurable"
)

var (
	reNumExpected = regexp.MustCompile(`^\s*(>=|≥|<=|≤|==|=|>|<)\s*([0-9]+(?:\.[0-9]+)?)\s*([A-Za-z%]*)\s*$`)
	reNumActual   = regexp.MustCompile(`^\s*((?:[0-9]{1,3}(?:,[0-9]{3})+)|[0-9]+)(\.[0-9]+)?\s*([A-Za-z%]*)\s*$`)
	reAllWord     = regexp.MustCompile(`(?i)^(all|every)(\s+\S.*)?$`)
	reOfCount     = regexp.MustCompile(`^\s*([0-9]+)\s+of\s+([0-9]+)\s*$`)
)

var timeUnits = map[string]float64{"ms": 1, "s": 1000, "sec": 1000, "secs": 1000, "min": 60000, "mins": 60000, "h": 3600000, "hr": 3600000, "hrs": 3600000}

// Judge returns the result, the actual value and the reason for a check that ran and
// exited 0. A pass expected is met by the exit code alone. Any other expected is judged by
// the output it was given, which must be the last line the command printed.
func Judge(expected, output string) (result, actual, reason string) {
	if binaryKind(expected) == "pass" {
		return ResultMet, "pass", ""
	}
	if output == "" {
		return ResultNotMeasurable, "", "command printed nothing to compare with the expected"
	}
	if binaryKind(expected) != "" {
		if binarySatisfies(expected, output) {
			return ResultMet, output, ""
		}
		return ResultNotMet, output, ""
	}
	r := numericSatisfies(expected, output)
	switch {
	case r.problem != "":
		return ResultNotMeasurable, output, r.problem
	case r.ok:
		return ResultMet, output, ""
	}
	return ResultNotMet, output, ""
}

// binaryKind returns the binary family of a expected, or "" if it is not binary.
func binaryKind(expected string) string {
	t := strings.ToLower(strings.TrimSpace(expected))
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

// binarySatisfies reports whether the actual value matches a binary expected.
func binarySatisfies(expected, actual string) bool {
	a := strings.ToLower(strings.TrimSpace(actual))
	switch binaryKind(expected) {
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

type numericResult struct {
	known   bool
	ok      bool
	problem string
}

// numericSatisfies compares a numeric actual with a numeric expected. Units must match,
// except that time units convert (ms, s, min, h).
func numericSatisfies(expected, actual string) numericResult {
	t := reNumExpected.FindStringSubmatch(strings.ReplaceAll(expected, "≥", ">="))
	if t == nil {
		return numericResult{}
	}
	op := t[1]
	want, _ := strconv.ParseFloat(t[2], 64)
	tUnit := t[3]

	a := reNumActual.FindStringSubmatch(actual)
	if a == nil {
		return numericResult{known: true, problem: "actual must be a number with the expected value's unit, for example 1.5 s for an expected value in s"}
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
			return numericResult{known: true, problem: fmt.Sprintf("units differ: expected %q, actual %q", tUnit, aUnit)}
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
	default:
		return got == want
	}
}
