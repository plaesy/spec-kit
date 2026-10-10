package trimmer

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

var (
	testFailRE = regexp.MustCompile(`(?i)fail|error|✗|✘`)
	testDoneRE = regexp.MustCompile(`(?i)passed|failed|tests? run|ok\.`)
)

// CompressCommandOutput mirrors compress_command_output(): dedupe
// consecutive identical lines (as "line (xN)") and, for "test-results" mode,
// keep only failure lines plus the trailing pass/fail summary; otherwise
// truncate to maxLines by keeping a head/tail split with an
// "... N lines omitted ..." marker in between.
func CompressCommandOutput(lines []string, maxLines int, dedupe bool, mode string) []string {
	if mode == "test-results" {
		return compressTestResults(lines)
	}

	var collapsed []string
	if !dedupe {
		collapsed = append(collapsed, lines...)
	} else {
		var prev string
		count := 0
		first := true
		for _, line := range lines {
			if first {
				prev, count, first = line, 1, false
				continue
			}
			if line == prev {
				count++
				continue
			}
			collapsed = append(collapsed, collapseLine(prev, count))
			prev, count = line, 1
		}
		if !first {
			collapsed = append(collapsed, collapseLine(prev, count))
		}
	}

	total := len(collapsed)
	if total <= maxLines {
		return collapsed
	}
	head := (maxLines + 1) / 2
	tail := maxLines - head
	omitted := total - head - tail
	out := make([]string, 0, maxLines+1)
	out = append(out, collapsed[:head]...)
	out = append(out, fmt.Sprintf("... %d lines omitted ...", omitted))
	out = append(out, collapsed[total-tail:]...)
	return out
}

func collapseLine(line string, count int) string {
	if count > 1 {
		return fmt.Sprintf("%s (x%d)", line, count)
	}
	return line
}

func compressTestResults(lines []string) []string {
	var kept []string
	for _, line := range lines {
		if testFailRE.MatchString(line) {
			kept = append(kept, line)
		}
	}
	total := len(lines)
	tailStart := 0
	if total > 5 {
		tailStart = total - 5
	}
	for i := tailStart; i < total; i++ {
		if testDoneRE.MatchString(lines[i]) {
			kept = append(kept, lines[i])
		}
	}
	if len(kept) == 0 {
		return []string{"(no failures - all tests passed, output suppressed)"}
	}
	seen := map[string]bool{}
	var out []string
	for _, line := range kept {
		if seen[line] {
			continue
		}
		seen[line] = true
		out = append(out, line)
	}
	return out
}

// matchLayer1Rule mirrors get_layer1_keys/get_layer1_rule: the longest
// "_default"-excluded key that is a prefix of cmdString wins; otherwise
// "_default" (or the built-in default if even that is absent).
//
// A nil rules pointer yields the built-in default. LoadRules never returns nil,
// so reaching here means a caller wired its own config; falling back to the
// defaults costs that caller the customisation, while a nil map read costs them
// a stack trace.
func matchLayer1Rule(rules *Rules, cmdString string) Layer1Rule {
	builtin := Layer1Rule{MaxLines: defaultMaxLines, DedupeConsecutive: true}
	if rules == nil {
		return builtin
	}
	var matched string
	for key := range rules.Layer1Commands {
		if key == "_default" {
			continue
		}
		if strings.HasPrefix(cmdString, key) && len(key) > len(matched) {
			matched = key
		}
	}
	if matched != "" {
		return rules.Layer1Commands[matched]
	}
	if rule, ok := rules.Layer1Commands["_default"]; ok {
		return rule
	}
	return builtin
}

// RunResult is the outcome of Run(): the compressed output plus the token
// counts recorded for the report.
type RunResult struct {
	Output string
	Before int
	After  int
}

// Run mirrors cmd_run(): executes args as a subprocess (stdout+stderr
// merged, matching the bash script's `"$@" 2>&1`), compresses the captured
// output per the matched Layer 1 rule, records the before/after token
// counts to statsPath, and returns the compressed text.
func Run(rules *Rules, statsPath string, args []string) (*RunResult, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("no command given")
	}
	cmdString := strings.Join(args, " ")
	rule := matchLayer1Rule(rules, cmdString)
	maxLines := rule.MaxLines
	if maxLines == 0 {
		maxLines = defaultMaxLines
	}

	c := exec.Command(args[0], args[1:]...)
	var buf bytes.Buffer
	c.Stdout = &buf
	c.Stderr = &buf
	runErr := c.Run() // mirrors bash's `set -uo pipefail` without `-e`: a failing command's output is still compressed and reported, not treated as fatal.

	raw := buf.String()
	rawLines := splitLines(raw)
	compressedLines := CompressCommandOutput(rawLines, maxLines, rule.DedupeConsecutive, rule.Mode)
	compressed := strings.Join(compressedLines, "\n")
	if len(compressedLines) > 0 {
		compressed += "\n"
	}

	before := EstimateTokens(raw)
	after := EstimateTokens(compressed)
	if statsPath != "" {
		if err := AddStatRecord(statsPath, "layer1", cmdString, before, after); err != nil {
			return nil, err
		}
	}

	result := &RunResult{Output: compressed, Before: before, After: after}
	if runErr != nil {
		// The compressed output is still returned alongside the failure: a
		// failed build's errors are exactly what the user needs to read. But
		// the failure itself must reach the caller.
		//
		// This used to return (result, nil), justified as parity with the
		// deleted scripts/bash/cmd_run.sh, which never checked "$?". That
		// justification expired when the bash scripts were removed: this Go
		// function is now the only implementation, and `trim run` is a
		// transparent wrapper — a wrapper that reports success for a command
		// that failed turns `plaesy trim run go build ./... && deploy` into a
		// deploy after a failed build. The caller prints result.Output before
		// propagating the error, so nothing is lost.
		return result, fmt.Errorf("command failed: %w", runErr)
	}
	return result, nil
}

// splitLines mirrors bash's `while IFS= read -r line` loop: split on "\n",
// dropping one trailing empty element left by a final newline.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	s = strings.TrimSuffix(s, "\n")
	if s == "" {
		return []string{""}
	}
	return strings.Split(s, "\n")
}
