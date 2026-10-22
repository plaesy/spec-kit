package trimmer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetNodeDegreeMatchesEitherSeparator(t *testing.T) {
	graph := filepath.Join(t.TempDir(), "project.graph.json")
	writeFile(t, graph, `{"nodes":[{"id":"docs/a.md","degree":7},{"id":"src\\b.go","degree":2}]}`)

	for _, tc := range []struct {
		name     string
		relative string
		want     int
		wantOK   bool
	}{
		{"forward slash id", "docs/a.md", 7, true},
		{"backslash id given forward", "src/b.go", 2, true},
		{"forward id given with a backslash", `src\b.go`, 2, true},
		{"absent from the graph", "docs/zzz.md", 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := getNodeDegree(graph, tc.relative)
			if got != tc.want || ok != tc.wantOK {
				t.Errorf("getNodeDegree(%q) = %d, %v; want %d, %v", tc.relative, got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

func TestGetNodeDegreeOnAnUnusableGraph(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "broken.json"), "not json")
	for _, tc := range []struct{ name, path string }{
		{"missing file", filepath.Join(dir, "absent.json")},
		{"unparseable file", filepath.Join(dir, "broken.json")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, ok := getNodeDegree(tc.path, "a.md"); ok {
				t.Error("a graph that cannot be read must not claim a file was found")
			}
		})
	}
}

func TestResolveLevel(t *testing.T) {
	dir := t.TempDir()
	graph := filepath.Join(dir, "project.graph.json")
	writeFile(t, graph, `{"nodes":[{"id":"hub.md","degree":9},{"id":"mid.md","degree":3},{"id":"leaf.md","degree":0}]}`)
	rules := defaultRules()

	for _, tc := range []struct {
		name      string
		relative  string
		requested string
		want      string
	}{
		{"an explicit level always wins", "hub.md", "ultra", "ultra"},
		{"a well-connected file is compressed least", "hub.md", "", "lite"},
		{"a middling file is compressed fully", "mid.md", "", "full"},
		{"a file with no edges is compressed hardest", "leaf.md", "", "ultra"},
		{"no graph data means full", "unknown.md", "", "full"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ResolveLevel(rules, graph, tc.relative, tc.requested); got != tc.want {
				t.Errorf("ResolveLevel(%q, %q) = %q, want %q", tc.relative, tc.requested, got, tc.want)
			}
		})
	}
}

// nil rules means the built-in defaults, the same contract matchLayer1Rule has:
// a caller that wired its own config should lose the customisation, not crash.
func TestResolveLevelSurvivesNilRules(t *testing.T) {
	dir := t.TempDir()
	graph := filepath.Join(dir, "project.graph.json")
	writeFile(t, graph, `{"nodes":[{"id":"hub.md","degree":9}]}`)
	if got := ResolveLevel(nil, graph, "hub.md", ""); got != "lite" {
		t.Errorf("ResolveLevel(nil rules) = %q, want lite from the built-in thresholds", got)
	}
	if got := ResolveLevel(nil, graph, "hub.md", "full"); got != "full" {
		t.Errorf("an explicit level must still win with nil rules, got %q", got)
	}
}

func TestApplyLevelSubstitutions(t *testing.T) {
	rules := defaultRules()
	for _, tc := range []struct {
		name  string
		level string
		in    string
		want  string
	}{
		// Removing a filler phrase leaves the space that followed it, exactly as
		// the bash sed pass did. Pinned so a future "cleanup" is a deliberate
		// change to the output rather than an accident.
		{"lite replaces filler phrases", "lite", "You should note that in order to ship", " note that to ship"},
		{"lite leaves whitespace alone", "lite", "a    b", "a    b"},
		{"lite leaves stopwords alone", "lite", "that is basically fine", "that is basically fine"},
		{"full collapses runs of spaces", "full", "a    b", "a b"},
		{"full strips leading indentation", "full", "    indented line", "indented line"},
		{"full still does not strip stopwords", "full", "that is fine", "that is fine"},
		{"ultra strips stopwords", "ultra", "that is fine", "is fine"},
		{"an unknown level only normalizes blank lines", "none", "a\n\n\n\nb", "a\n\nb"},
		{"three blank lines collapse to one at every level", "lite", "a\n\n\n\nb", "a\n\nb"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := applyLevel(tc.in, tc.level, rules); got != tc.want {
				t.Errorf("applyLevel(%q, %q) = %q, want %q", tc.in, tc.level, got, tc.want)
			}
		})
	}
}

// An invalid pattern in a project rules file must not stop the whole pass: the
// bash twin's `sed ... || printf original` was a no-op for that one pattern.
func TestApplyLevelSkipsAnUncompilablePattern(t *testing.T) {
	rules := &Rules{
		Layer2FillerPatterns: []FillerPattern{
			{Find: `([unclosed`, Replace: "X"},
			{Find: `\bin order to\b`, Replace: "to"},
		},
	}
	if got := applyLevel("in order to ship", "lite", rules); got != "to ship" {
		t.Errorf("got %q, want %q: the valid pattern must still apply", got, "to ship")
	}
}

func TestCompressProseLeavesCodeBlocksAlone(t *testing.T) {
	input := strings.Join([]string{
		"You should note that the   answer",
		"",
		"```go",
		"You should note that   this is code",
		"```",
		"",
		"trailing prose You should note that",
		"",
	}, "\n")
	got := CompressProse(input, "full", defaultRules())

	if !strings.Contains(got, "You should note that   this is code") {
		t.Errorf("the code block was rewritten:\n%s", got)
	}
	if !strings.Contains(got, "the answer") {
		t.Errorf("prose before the fence was not compressed:\n%s", got)
	}
	if !strings.Contains(got, "trailing prose") {
		t.Errorf("prose after the fence was lost:\n%s", got)
	}
}

func TestCompressProseReproducesTheTrailingNewline(t *testing.T) {
	for _, tc := range []struct{ name, in, want string }{
		{"one trailing newline", "in order to a\n", "to a\n"},
		{"no trailing newline", "in order to a", "to a\n"},
		{"empty input", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := CompressProse(tc.in, "lite", defaultRules()); got != tc.want {
				t.Errorf("CompressProse(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestSplitLinesKeepTrailing(t *testing.T) {
	if got := splitLinesKeepTrailing(""); got != nil {
		t.Errorf("got %v, want nil", got)
	}
	if got := splitLinesKeepTrailing("a\n\n"); len(got) != 2 || got[1] != "" {
		t.Errorf("got %#v, want the trailing blank kept", got)
	}
}

func TestFileCompressResultString(t *testing.T) {
	got := FileCompressResult{Relative: "docs/a.md", Level: "full", Before: 1000, After: 400}.String()
	for _, want := range []string{"docs/a.md [full]", "1000 tokens -> 400 tokens", "60.0% saved"} {
		if !contains(got, want) {
			t.Errorf("result line is missing %q: %q", want, got)
		}
	}
}

func TestCompressOneFileWritesABackupAndStats(t *testing.T) {
	repo := t.TempDir()
	stats := filepath.Join(repo, "stats", "token-stats.json")
	doc := filepath.Join(repo, "a.md")
	writeFile(t, doc, "in order to ship\n")

	res, err := CompressOneFile(defaultRules(), repo, "", stats, doc, "lite", false)
	if err != nil {
		t.Fatalf("CompressOneFile: %v", err)
	}
	if res.Level != "lite" || res.Relative != "a.md" {
		t.Errorf("result = %+v, want level lite and a repo-relative name", res)
	}
	backup, err := os.ReadFile(doc + ".bak")
	if err != nil {
		t.Fatalf("no backup was written: %v", err)
	}
	if string(backup) != "in order to ship\n" {
		t.Errorf("backup = %q, want the original bytes", backup)
	}
	now, err := os.ReadFile(doc)
	if err != nil {
		t.Fatal(err)
	}
	if string(now) != "to ship\n" {
		t.Errorf("file = %q, want it compressed", now)
	}
	records := readStats(t, stats)
	if len(records) != 1 || records[0].Layer != "layer2" || records[0].Path != "a.md" {
		t.Errorf("stats = %+v", records)
	}
}

func TestCompressOneFileDryRunTouchesNothing(t *testing.T) {
	repo := t.TempDir()
	doc := filepath.Join(repo, "a.md")
	writeFile(t, doc, "in order to ship\n")

	res, err := CompressOneFile(defaultRules(), repo, "", "", doc, "lite", true)
	if err != nil {
		t.Fatalf("CompressOneFile: %v", err)
	}
	if res.After >= res.Before {
		t.Errorf("a dry run must still report what it would save: %+v", res)
	}
	got, err := os.ReadFile(doc)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "in order to ship\n" {
		t.Errorf("a dry run modified the file: %q", got)
	}
	if _, err := os.Stat(doc + ".bak"); !os.IsNotExist(err) {
		t.Error("a dry run wrote a backup")
	}
}

func TestCompressOneFileOnAMissingFile(t *testing.T) {
	_, err := CompressOneFile(defaultRules(), t.TempDir(), "", "", filepath.Join(t.TempDir(), "absent.md"), "lite", true)
	if err == nil {
		t.Error("compressing a file that does not exist must be an error, not an empty result")
	}
}

func TestCompressPathDirectoryDispatch(t *testing.T) {
	repo := t.TempDir()
	docs := filepath.Join(repo, "docs")
	writeFile(t, filepath.Join(docs, "a.md"), "in order to a\n")
	writeFile(t, filepath.Join(docs, "b.md"), "in order to b\n")
	writeFile(t, filepath.Join(docs, "notes.txt"), "in order to c\n")
	writeFile(t, filepath.Join(docs, "deep", "c.md"), "in order to d\n")

	t.Run("direct children only", func(t *testing.T) {
		results, err := CompressPath(defaultRules(), repo, "", "", filepath.Join(repo, "docs"), "lite", false, true)
		if err != nil {
			t.Fatalf("CompressPath: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("compressed %d files, want the 2 .md children (a .txt and a subdirectory are not prose)", len(results))
		}
	})

	t.Run("recursing reaches nested markdown", func(t *testing.T) {
		results, err := CompressPath(defaultRules(), repo, "", "", filepath.Join(repo, "docs"), "lite", true, true)
		if err != nil {
			t.Fatalf("CompressPath: %v", err)
		}
		if len(results) != 3 {
			t.Errorf("compressed %d files, want 3", len(results))
		}
	})

	t.Run("a single file is compressed whatever its extension", func(t *testing.T) {
		results, err := CompressPath(defaultRules(), repo, "", "", filepath.Join(docs, "notes.txt"), "lite", false, true)
		if err != nil {
			t.Fatalf("CompressPath: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("compressed %d files, want 1", len(results))
		}
	})

	t.Run("a path relative to the repo root is resolved", func(t *testing.T) {
		results, err := CompressPath(defaultRules(), repo, "", "", "docs", "lite", false, true)
		if err != nil {
			t.Fatalf("CompressPath: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("compressed %d files, want 2", len(results))
		}
	})

	t.Run("a path that does not exist is an error", func(t *testing.T) {
		if _, err := CompressPath(defaultRules(), repo, "", "", "nowhere", "lite", false, true); err == nil {
			t.Error("a missing path must be an error")
		}
	})
}

// A rules file that parses but omits a section used to leave that section at
// Go's zero value, because a missing key and an explicit 0 decode the same.
// Omitting layer2_degree_thresholds set both thresholds to 0, so ResolveLevel
// took the "lite" branch for every file; omitting layer2_filler_patterns left
// no patterns, so compress performed no passes at all. trim then reported
// 0.0% saved and exited 0. Layer1Commands already fell back per section —
// these two did not.
func TestLoadRulesFillsAbsentSectionsFromTheDefaults(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"empty object", `{}`},
		{"only layer 1", `{"layer1_commands":{"_default":{"max_lines":5}}}`},
		{"only filler patterns", `{"layer2_filler_patterns":[{"find":"\bfoo\b","replace":"bar"}]}`},
		{"only thresholds", `{"layer2_degree_thresholds":{"lite_min_degree":9,"full_min_degree":4}}`},
		{"one threshold key", `{"layer2_degree_thresholds":{"lite_min_degree":9}}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "plaesy-trim-rules.json")
			if err := os.WriteFile(path, []byte(tc.body), 0o644); err != nil {
				t.Fatalf("write: %v", err)
			}
			got := LoadRules(path)
			def := defaultRules()
			if len(got.Layer2FillerPatterns) == 0 {
				t.Errorf("no filler patterns: a partial override disabled every Layer 2 pass")
			}
			if got.Layer2DegreeThreshold.LiteMinDegree <= 0 {
				t.Errorf("LiteMinDegree = %d, want the built-in default %d", got.Layer2DegreeThreshold.LiteMinDegree, def.Layer2DegreeThreshold.LiteMinDegree)
			}
			if got.Layer2DegreeThreshold.FullMinDegree <= 0 {
				t.Errorf("FullMinDegree = %d, want the built-in default %d", got.Layer2DegreeThreshold.FullMinDegree, def.Layer2DegreeThreshold.FullMinDegree)
			}
			if got.Layer1Commands == nil {
				t.Error("Layer1Commands must fall back to the built-in defaults")
			}
		})
	}
}

// A value that IS present must win over the default — the per-section fallback
// must not overwrite a project's own configuration.
func TestLoadRulesKeepsValuesTheFileDoesSupply(t *testing.T) {
	path := filepath.Join(t.TempDir(), "plaesy-trim-rules.json")
	body := `{"layer2_degree_thresholds":{"lite_min_degree":11,"full_min_degree":7}}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	got := LoadRules(path)
	if got.Layer2DegreeThreshold.LiteMinDegree != 11 || got.Layer2DegreeThreshold.FullMinDegree != 7 {
		t.Errorf("thresholds = %+v, want the values the file supplied", got.Layer2DegreeThreshold)
	}
	// The section the file did not supply still comes from the defaults.
	if len(got.Layer2FillerPatterns) != len(defaultRules().Layer2FillerPatterns) {
		t.Errorf("filler patterns = %d, want the %d defaults", len(got.Layer2FillerPatterns), len(defaultRules().Layer2FillerPatterns))
	}
}
