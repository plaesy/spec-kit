package trimmer

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// loadSample reads a real captured command output from testdata and splits it
// into lines the same way Run does.
func loadSample(t *testing.T, name string) []string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return splitLines(string(data))
}

func TestDetectFilter(t *testing.T) {
	cases := []struct {
		args []string
		want string
	}{
		{[]string{"git", "diff"}, "git-diff"},
		{[]string{"git", "diff", "--stat"}, "git-diff"},
		{[]string{"git", "status"}, "git-status"},
		{[]string{"git", "log", "-n", "15"}, "git-log"},
		{[]string{"git", "commit", "-m", "x"}, ""},
		{[]string{"git"}, ""},
		{[]string{"grep", "-rn", "func"}, "grep"},
		{[]string{"rg", "func"}, "grep"},
		{[]string{"ls", "-la"}, "listing"},
		{[]string{"find", "."}, "listing"},
		{[]string{"tree"}, "listing"},
		{[]string{"go", "test", "./..."}, ""},
		{[]string{"echo", "hi"}, ""},
		{nil, ""},
	}
	for _, c := range cases {
		name, fn := detectFilter(c.args)
		if name != c.want {
			t.Errorf("detectFilter(%v) = %q, want %q", c.args, name, c.want)
		}
		if (fn == nil) != (c.want == "") {
			t.Errorf("detectFilter(%v) filter nil = %v, want nil = %v", c.args, fn == nil, c.want == "")
		}
	}
}

// TestFiltersOnCapturedOutput runs each filter on real output and checks the
// structural promise it makes: the lines an agent needs are still there, and
// the lines it does not need are gone.
func TestFiltersOnCapturedOutput(t *testing.T) {
	cases := []struct {
		name   string
		sample string
		args   []string
		check  func(t *testing.T, in, out []string)
	}{
		{
			name:   "git-diff keeps headers, hunks and changes, drops index and context",
			sample: "git_diff.txt",
			args:   []string{"git", "diff"},
			check: func(t *testing.T, in, out []string) {
				if !containsPrefix(out, "diff --git") || !containsPrefix(out, "@@") {
					t.Error("diff headers or hunk markers were dropped")
				}
				if containsPrefix(out, "index ") {
					t.Error("index lines should be dropped")
				}
				for _, l := range out {
					if strings.HasPrefix(l, " ") {
						t.Errorf("context line kept: %q", l)
						break
					}
				}
				if len(out) >= len(in) {
					t.Errorf("no reduction: %d -> %d lines", len(in), len(out))
				}
			},
		},
		{
			name:   "git-status drops the usage hints",
			sample: "git_status.txt",
			args:   []string{"git", "status"},
			check: func(t *testing.T, in, out []string) {
				if !containsPrefix(out, "On branch") {
					t.Error("branch line was dropped")
				}
				for _, l := range out {
					if strings.Contains(l, `(use "git`) {
						t.Errorf("hint kept: %q", l)
						break
					}
				}
				if len(out) >= len(in) {
					t.Errorf("no reduction: %d -> %d lines", len(in), len(out))
				}
			},
		},
		{
			name:   "git-log keeps one short line per commit with its subject",
			sample: "git_log.txt",
			args:   []string{"git", "log"},
			check: func(t *testing.T, in, out []string) {
				commits := 0
				for _, l := range out {
					if strings.HasPrefix(l, "Author:") || strings.HasPrefix(l, "Date:") {
						t.Errorf("metadata kept: %q", l)
					}
					if strings.HasPrefix(l, "commit ") {
						commits++
						if sha := strings.Fields(l)[1]; len(sha) != 7 {
							t.Errorf("sha not shortened to 7: %q", l)
						}
					}
				}
				if commits == 0 {
					t.Error("no commit lines kept")
				}
				if len(out) >= len(in) {
					t.Errorf("no reduction: %d -> %d lines", len(in), len(out))
				}
			},
		},
		{
			name:   "grep drops blank lines and collapses separators",
			sample: "grep_out.txt",
			args:   []string{"grep", "-rn", "func"},
			check: func(t *testing.T, in, out []string) {
				if len(out) == 0 {
					t.Fatal("grep output emptied")
				}
				for _, l := range out {
					if strings.TrimSpace(l) == "" {
						t.Error("blank line kept")
						break
					}
				}
			},
		},
		{
			name:   "listing drops the total line and blank lines",
			sample: "ls_out.txt",
			args:   []string{"ls", "-la"},
			check: func(t *testing.T, in, out []string) {
				if containsPrefix(out, "total ") {
					t.Error("ls total line kept")
				}
				if len(out) == 0 {
					t.Fatal("listing emptied")
				}
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			in := loadSample(t, c.sample)
			name, fn := detectFilter(c.args)
			if fn == nil {
				t.Fatalf("no filter for %v (name %q)", c.args, name)
			}
			out, err := fn(in)
			if err != nil {
				t.Fatalf("filter returned error on real output: %v", err)
			}
			c.check(t, in, out)
		})
	}
}

// TestFilterErrorFallsBackToRawOutput: a filter that cannot recognise its input
// returns an error, so the raw lines pass through, with a warning on stderr.
func TestFilterErrorFallsBackToRawOutput(t *testing.T) {
	raw := []string{"not a diff", "at all"}
	var warn bytes.Buffer
	withWarnWriter(t, &warn)

	out := applyFilter("git-diff", func([]string) ([]string, error) {
		return nil, errors.New("unrecognised input")
	}, raw)

	if !equalLines(out, raw) {
		t.Errorf("out = %q, want raw %q", out, raw)
	}
	if !strings.Contains(warn.String(), "git-diff") || !strings.Contains(warn.String(), "unrecognised input") {
		t.Errorf("warning = %q, want the filter name and error", warn.String())
	}
}

// TestFilterPanicFallsBackToRawOutput: a panicking filter is contained the same
// way, so one bad parser never turns command output into nothing.
func TestFilterPanicFallsBackToRawOutput(t *testing.T) {
	raw := []string{"line one", "line two"}
	var warn bytes.Buffer
	withWarnWriter(t, &warn)

	out := applyFilter("grep", func([]string) ([]string, error) {
		panic("boom")
	}, raw)

	if !equalLines(out, raw) {
		t.Errorf("out = %q, want raw %q", out, raw)
	}
	if !strings.Contains(warn.String(), "boom") {
		t.Errorf("warning = %q, want the panic value", warn.String())
	}
}

// TestGitDiffNonDiffInputFallsBack: real filter path. Output that is not a diff
// makes git-diff return an error, so the caller gets the raw text back.
func TestGitDiffNonDiffInputFallsBack(t *testing.T) {
	var warn bytes.Buffer
	withWarnWriter(t, &warn)
	raw := []string{"fatal: not a git repository"}
	lines, layer := prepareLines([]string{"git", "diff"}, raw)
	if !equalLines(lines, raw) {
		t.Errorf("lines = %q, want raw %q", lines, raw)
	}
	if layer != "layer1:git-diff" {
		t.Errorf("layer = %q, want layer1:git-diff", layer)
	}
	if warn.Len() == 0 {
		t.Error("expected a stderr warning when the filter fails open")
	}
}

// TestPrepareLinesUnknownCommandIsUnchanged is the regression guard: an
// unrecognised command passes through untouched and records under the old layer.
func TestPrepareLinesUnknownCommandIsUnchanged(t *testing.T) {
	raw := []string{"alpha", "", "alpha", "beta"}
	lines, layer := prepareLines([]string{"go", "version"}, raw)
	if !equalLines(lines, raw) {
		t.Errorf("lines = %q, want %q", lines, raw)
	}
	if layer != "layer1" {
		t.Errorf("layer = %q, want layer1", layer)
	}
}

// TestFilteredOutputIsTruncatedWithAnOmittedCount: a long filtered result still
// keeps head and tail with the omitted-lines count, through the shared path.
func TestFilteredOutputIsTruncatedWithAnOmittedCount(t *testing.T) {
	var raw []string
	for i := 0; i < 200; i++ {
		raw = append(raw, "pkg/file.go:1: func F"+strings.Repeat("x", i%3)+"()")
	}
	lines, _ := prepareLines([]string{"grep", "func"}, raw)
	compressed := CompressCommandOutput(lines, 10, true, "")
	if len(compressed) != 11 {
		t.Fatalf("got %d lines, want 10 kept plus one marker", len(compressed))
	}
	if !strings.Contains(strings.Join(compressed, "\n"), "lines omitted") {
		t.Error("omitted-lines marker missing")
	}
}

func withWarnWriter(t *testing.T, buf *bytes.Buffer) {
	t.Helper()
	prev := warnWriter
	warnWriter = buf
	t.Cleanup(func() { warnWriter = prev })
}
