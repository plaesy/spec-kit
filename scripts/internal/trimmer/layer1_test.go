package trimmer

import (
	"strings"
	"testing"
)

func TestCompressCommandOutputTruncatesWithAHeadTailSplit(t *testing.T) {
	lines := make([]string, 0, 60)
	for i := 0; i < 60; i++ {
		lines = append(lines, "line")
	}
	got := CompressCommandOutput(lines, 10, false, "")
	if len(got) != 11 {
		t.Fatalf("got %d lines, want 10 plus the omission marker: %v", len(got), got)
	}
	// head = (10+1)/2 = 5, tail = 5, so 50 lines are omitted and the marker
	// sits at index 5.
	if got[5] != "... 50 lines omitted ..." {
		t.Errorf("marker = %q, want %q", got[5], "... 50 lines omitted ...")
	}
}

func TestCompressCommandOutputDedupeCollapsesRuns(t *testing.T) {
	got := CompressCommandOutput([]string{"a", "a", "a", "b", "b", "c"}, 100, true, "")
	want := []string{"a (x3)", "b (x2)", "c"}
	if !equalLines(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestCompressCommandOutputDedupeLeavesDistinctLinesAlone(t *testing.T) {
	got := CompressCommandOutput([]string{"x"}, 100, true, "")
	if !equalLines(got, []string{"x"}) {
		t.Errorf("got %v, want [x]", got)
	}
}

func TestCompressCommandOutputEmptyInput(t *testing.T) {
	if got := CompressCommandOutput(nil, 10, true, ""); got != nil {
		t.Errorf("got %v, want nil", got)
	}
}

// A run of identical lines is collapsed *before* the length is measured, so
// fifty copies of one line fit inside a five-line budget as a single
// "z (x50)". Truncating first would throw away exactly the repetition the
// dedupe exists to remove.
func TestCompressCommandOutputDedupeHappensBeforeTruncation(t *testing.T) {
	lines := make([]string, 0, 50)
	for i := 0; i < 50; i++ {
		lines = append(lines, "z")
	}
	got := CompressCommandOutput(lines, 5, true, "")
	if !equalLines(got, []string{"z (x50)"}) {
		t.Errorf("got %v, want [z (x50)]", got)
	}
}

func TestCompressTestResultsKeepsFailuresAndTheSummary(t *testing.T) {
	// The summary matcher is "passed|failed|tests? run|ok." — a bare "PASS" or
	// an "ok    pkg/test  1s" line without a dot matches neither, so the only
	// lines kept here are the two failures. That is the whole point of the mode:
	// a failing run's summary is "--- FAIL:" or "FAIL", both of which the
	// failure matcher already catches.
	for _, tc := range []struct {
		name  string
		lines []string
		want  []string
	}{
		{
			name:  "failures only",
			lines: []string{"--- FAIL: TestFoo", "ok    pkg/test  1s", "PASS", "--- FAIL: TestBar", "random line"},
			want:  []string{"--- FAIL: TestFoo", "--- FAIL: TestBar"},
		},
		{
			name:  "duplicates collapse",
			lines: []string{"FAIL", "FAIL", "FAIL"},
			want:  []string{"FAIL"},
		},
		{
			name:  "a passing run says so instead of printing nothing",
			lines: []string{"ok", "PASS", "everything is fine"},
			want:  []string{"(no failures - all tests passed, output suppressed)"},
		},
		{
			name:  "a real summary line in the tail is kept",
			lines: []string{"running 3 tests", "3 tests run, 3 passed"},
			want:  []string{"3 tests run, 3 passed"},
		},
		{
			name:  "a summary before the tail window is not resurrected",
			lines: []string{"12 tests run, 12 passed", "a", "b", "c", "d", "e", "f"},
			want:  []string{"(no failures - all tests passed, output suppressed)"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := CompressCommandOutput(tc.lines, 100, true, "test-results")
			if !equalLines(got, tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCollapseLine(t *testing.T) {
	for _, tc := range []struct {
		name  string
		line  string
		count int
		want  string
	}{
		{"single", "x", 1, "x"},
		{"multi", "x", 5, "x (x5)"},
		{"empty line repeated", "", 3, " (x3)"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := collapseLine(tc.line, tc.count); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSplitLines(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want []string
	}{
		{"empty", "", nil},
		{"no trailing newline", "abc", []string{"abc"}},
		{"one trailing newline is dropped", "abc\n", []string{"abc"}},
		{"two trailing newlines keep one blank", "abc\n\n", []string{"abc", ""}},
		{"multi", "a\nb\nc", []string{"a", "b", "c"}},
		{"blank line in the middle is kept", "a\n\nb", []string{"a", "", "b"}},
		{"crlf keeps the carriage return", "a\r\nb", []string{"a\r", "b"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := splitLines(tc.in)
			if !equalLines(got, tc.want) {
				t.Errorf("splitLines(%q) = %#v, want %#v", tc.in, got, tc.want)
			}
		})
	}
}

func TestMatchLayer1Rule(t *testing.T) {
	rules := defaultRules()
	for _, tc := range []struct {
		name     string
		cmd      string
		maxLines int
		dedupe   bool
		mode     string
	}{
		{"longest prefix wins", "git status --short", 25, true, ""},
		{"a shorter prefix does not shadow a longer one", "npm test", 50, true, "test-results"},
		{"no match falls through to _default", "unknown command", defaultMaxLines, true, ""},
		{"a command that merely contains a key is not a match", "sudo git push", defaultMaxLines, true, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := matchLayer1Rule(rules, tc.cmd)
			if got.MaxLines != tc.maxLines || got.DedupeConsecutive != tc.dedupe || got.Mode != tc.mode {
				t.Errorf("matchLayer1Rule(%q) = %+v, want maxLines=%d dedupe=%v mode=%q",
					tc.cmd, got, tc.maxLines, tc.dedupe, tc.mode)
			}
		})
	}
}

// A nil *Rules used to panic on the first map read. LoadRules never returns nil,
// so the only way here is a caller that wired its own config — the kind of
// wiring mistake that should cost you the defaults, not a stack trace.
func TestMatchLayer1RuleSurvivesNilRules(t *testing.T) {
	got := matchLayer1Rule(nil, "anything")
	if got.MaxLines != defaultMaxLines || !got.DedupeConsecutive {
		t.Errorf("matchLayer1Rule(nil) = %+v, want the built-in default", got)
	}
}

func TestRunWithoutACommand(t *testing.T) {
	if _, err := Run(defaultRules(), "", nil); err == nil {
		t.Error("Run(nil args) = nil error, want one")
	}
}

func TestRunCompressesAndRecordsStats(t *testing.T) {
	stats := t.TempDir() + "/nested/token-stats.json"
	res, err := Run(defaultRules(), stats, goCommand("version"))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(res.Output, "go version") {
		t.Errorf("output = %q, want the command's own output", res.Output)
	}
	if res.Before <= 0 || res.After <= 0 {
		t.Errorf("token counts = %d -> %d, want both positive", res.Before, res.After)
	}
	records := readStats(t, stats)
	if len(records) != 1 {
		t.Fatalf("stats file holds %d records, want 1", len(records))
	}
	if records[0].Layer != "layer1" || records[0].Before != res.Before || records[0].After != res.After {
		t.Errorf("recorded %+v, does not match the run %+v", records[0], res)
	}
}

// A command that exits non-zero still produced output worth compressing, and
// that output is the only part of a failing build anyone reads — so Run must
// return it *together with* the error, never instead of it.
//
// This used to assert the opposite (exit status swallowed, nil error). The
// justification was parity with the deleted bash cmd_run.sh, which never checked
// "$?". That is no longer a reason: this is the only implementation, and
// `trim run` is a transparent wrapper. Swallowing the status made
// `plaesy trim run go build ./... && deploy` deploy after a failed build.
func TestRunKeepsOutputAndErrorFromAFailingCommand(t *testing.T) {
	res, err := Run(defaultRules(), "", goCommand("definitely-not-a-subcommand"))
	if err == nil {
		t.Fatal("a failing command must surface an error, or the wrapper reports success for a failed build")
	}
	if res == nil {
		t.Fatal("the result must be returned alongside the error, or the failure text is lost")
	}
	if !strings.Contains(res.Output, "definitely-not-a-subcommand") {
		t.Errorf("output = %q, want the failing command's own message", res.Output)
	}
	if !strings.Contains(err.Error(), "command failed") {
		t.Errorf("error = %v, want it to say the command failed", err)
	}
}

func TestRunWithoutStatsPathRecordsNothing(t *testing.T) {
	dir := t.TempDir()
	before := readDirEntries(t, dir)
	if _, err := Run(defaultRules(), "", goCommand("version")); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if after := readDirEntries(t, dir); len(after) != len(before) {
		t.Errorf("Run wrote %v with an empty stats path", after)
	}
}

func equalLines(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
