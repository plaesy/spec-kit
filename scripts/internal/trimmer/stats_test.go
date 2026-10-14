package trimmer

import (
	"path/filepath"
	"testing"
)

func TestEstimateTokens(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want int
	}{
		{"empty", "", 0},
		{"one char rounds up", "a", 1},
		{"exactly four", "abcd", 1},
		{"five rounds up", "abcde", 2},
		{"a thousand characters", string(make([]byte, 1000)), 250},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := EstimateTokens(tc.in); got != tc.want {
				t.Errorf("EstimateTokens(%d chars) = %d, want %d", len(tc.in), got, tc.want)
			}
		})
	}
}

func TestPctSaved(t *testing.T) {
	for _, tc := range []struct {
		name          string
		before, after int
		want          float64
	}{
		{"halved", 1000, 500, 50.0},
		{"nothing saved", 1000, 1000, 0.0},
		{"empty input is not infinite saving", 0, 0, 0.0},
		{"all of it", 100, 0, 100.0},
		{"rounded to one decimal", 3, 1, 66.7},
		{"negative when the output grew", 100, 150, -50.0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := PctSaved(tc.before, tc.after); got != tc.want {
				t.Errorf("PctSaved(%d, %d) = %v, want %v", tc.before, tc.after, got, tc.want)
			}
		})
	}
}

func TestAddStatRecordAppends(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "token-stats.json")
	if err := AddStatRecord(path, "layer1", "git status", 1000, 400); err != nil {
		t.Fatalf("AddStatRecord: %v", err)
	}
	if err := AddStatRecord(path, "layer2", "README.md", 200, 100); err != nil {
		t.Fatalf("AddStatRecord: %v", err)
	}
	records := readStats(t, path)
	if len(records) != 2 {
		t.Fatalf("got %d records, want 2 (the second call must append, not overwrite)", len(records))
	}
	if records[0].Path != "git status" || records[1].Path != "README.md" {
		t.Errorf("records out of order: %+v", records)
	}
	if records[0].SavedPct != 60.0 || records[1].SavedPct != 50.0 {
		t.Errorf("saved_pct = %v, %v; want 60, 50", records[0].SavedPct, records[1].SavedPct)
	}
	if records[0].Timestamp == "" {
		t.Error("a record with no timestamp cannot be placed in a report")
	}
}

// A corrupt stats file is replaced rather than appended to: appending to
// unparseable JSON would leave a file that is corrupt in a different way, and
// the bash twin did the same best-effort reset.
func TestAddStatRecordResetsACorruptFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "token-stats.json")
	writeFile(t, path, "{ this is not json")
	if err := AddStatRecord(path, "layer1", "x", 10, 5); err != nil {
		t.Fatalf("AddStatRecord: %v", err)
	}
	records := readStats(t, path)
	if len(records) != 1 {
		t.Errorf("got %d records, want the corrupt file to be reset to 1", len(records))
	}
}

func TestReportWithoutStats(t *testing.T) {
	got, err := Report(filepath.Join(t.TempDir(), "absent.json"))
	if err != nil {
		t.Fatalf("Report: %v", err)
	}
	want := "No token stats yet. Run 'compress' or 'run' first.\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestReportSumsPerLayerAndTotals(t *testing.T) {
	path := filepath.Join(t.TempDir(), "token-stats.json")
	for _, r := range []struct {
		layer, target string
		before, after int
	}{
		{"layer2", "README.md", 1000, 500},
		{"layer1", "git status", 400, 100},
		{"layer1", "npm test", 600, 300},
	} {
		if err := AddStatRecord(path, r.layer, r.target, r.before, r.after); err != nil {
			t.Fatalf("AddStatRecord: %v", err)
		}
	}
	got, err := Report(path)
	if err != nil {
		t.Fatalf("Report: %v", err)
	}
	for _, want := range []string{
		"plaesy-trim report",
		"layer1: 2 runs, 1000 -> 400 tokens (60.0% avg saved)",
		"layer2: 1 runs, 1000 -> 500 tokens (50.0% avg saved)",
		"total: 3 runs", // the totals line reports tokens, not runs — checked below
	} {
		if want == "total: 3 runs" {
			continue
		}
		if !contains(got, want) {
			t.Errorf("report is missing %q:\n%s", want, got)
		}
	}
	if !contains(got, "total: 2000 -> 900 tokens (55.0% saved)") {
		t.Errorf("grand total wrong:\n%s", got)
	}
}

// The report lists layers in sorted order, so two runs of the same report read
// the same way regardless of the order the records were appended in.
func TestReportOrdersLayersDeterministically(t *testing.T) {
	path := filepath.Join(t.TempDir(), "token-stats.json")
	for _, layer := range []string{"zeta", "alpha", "mu"} {
		if err := AddStatRecord(path, layer, "t", 100, 50); err != nil {
			t.Fatalf("AddStatRecord: %v", err)
		}
	}
	got, err := Report(path)
	if err != nil {
		t.Fatalf("Report: %v", err)
	}
	alpha, mu, zeta := indexOf(got, "alpha:"), indexOf(got, "mu:"), indexOf(got, "zeta:")
	if !(alpha < mu && mu < zeta) {
		t.Errorf("layers are not in sorted order (alpha=%d mu=%d zeta=%d):\n%s", alpha, mu, zeta, got)
	}
}

func TestReportOnACorruptFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "token-stats.json")
	writeFile(t, path, "not json at all")
	got, err := Report(path)
	if err != nil {
		t.Fatalf("a corrupt stats file must not be an error: %v", err)
	}
	if !contains(got, "No token stats yet") {
		t.Errorf("got %q, want the empty-report message", got)
	}
}

func TestReportOnAnEmptyArray(t *testing.T) {
	path := filepath.Join(t.TempDir(), "token-stats.json")
	writeFile(t, path, "[]\n")
	got, err := Report(path)
	if err != nil {
		t.Fatalf("Report: %v", err)
	}
	if !contains(got, "No token stats yet") {
		t.Errorf("got %q, want the empty-report message", got)
	}
}

// The resolution order is: a project override wins, then the framework's own
// copy, and when neither exists the *project* path is returned anyway so the
// caller can report which file it looked for. An empty frameworkRoot disables
// the middle step rather than searching a relative path.
func TestResolveRulesPathResolutionOrder(t *testing.T) {
	repo := t.TempDir()
	frameworkRoot := t.TempDir()
	projectPath := filepath.Join(repo, ".plaesy", "scripts", "configs", "plaesy-trim-rules.json")
	frameworkPath := filepath.Join(frameworkRoot, "scripts", "configs", "plaesy-trim-rules.json")

	if got := ResolveRulesPath(repo, frameworkRoot); got != projectPath {
		t.Errorf("with nothing on disk, got %q, want the project path %q", got, projectPath)
	}

	writeFile(t, frameworkPath, "{}")
	if got := ResolveRulesPath(repo, frameworkRoot); got != frameworkPath {
		t.Errorf("with only the framework copy, got %q, want %q", got, frameworkPath)
	}
	if got := ResolveRulesPath(repo, ""); got != projectPath {
		t.Errorf("an empty frameworkRoot must skip the framework copy, got %q", got)
	}

	writeFile(t, projectPath, "{}")
	if got := ResolveRulesPath(repo, frameworkRoot); got != projectPath {
		t.Errorf("the project override must win, got %q, want %q", got, projectPath)
	}
}

func TestLoadRulesFallsBackToTheBuiltInDefaults(t *testing.T) {
	builtIn := defaultRules()
	for _, tc := range []struct {
		name    string
		content string
		write   bool
	}{
		{name: "file is absent"},
		{name: "file is not json", content: "{{{", write: true},
		{name: "layer1_commands absent", content: `{"version":"9.9.9"}`, write: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "rules.json")
			if tc.write {
				writeFile(t, path, tc.content)
			}
			got := LoadRules(path)
			if got == nil {
				t.Fatal("LoadRules returned nil")
			}
			if len(got.Layer1Commands) == 0 {
				t.Error("LoadRules must always return usable rules")
			}
			if _, ok := got.Layer1Commands["_default"]; !ok && tc.name == "file is absent" {
				t.Error("the built-in fallback must carry the default rule")
			}
		})
	}
	if builtIn.Version == "" {
		t.Error("the built-in rules are missing a version")
	}
}

func TestLoadRulesKeepsAProjectOverride(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rules.json")
	writeFile(t, path, `{"version":"2.0.0","layer1_commands":{"_default":{"max_lines":3}}}`)
	got := LoadRules(path)
	if got.Version != "2.0.0" {
		t.Errorf("version = %q, want 2.0.0", got.Version)
	}
	if got.Layer1Commands["_default"].MaxLines != 3 {
		t.Errorf("project max_lines = %d, want 3", got.Layer1Commands["_default"].MaxLines)
	}
}

func TestFileExists(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "f.json")
	writeFile(t, file, "{}")
	if !fileExists(file) {
		t.Error("fileExists on a regular file = false")
	}
	if fileExists(dir) {
		t.Error("fileExists on a directory = true, want false: a directory cannot be read as a rules file")
	}
	if fileExists(filepath.Join(dir, "absent")) {
		t.Error("fileExists on a missing path = true")
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}
