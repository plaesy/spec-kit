package analyze

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/plaesy/spec-kit/internal/graph"
)

// writeGraphFixture builds a minimal project.graph.json at the path
// generateSymbolsIndex reads from, so these tests don't depend on running a
// full graph build.
func writeGraphFixture(t *testing.T, projectPath string, nodes []graph.Node) {
	t.Helper()
	paths, err := graph.ResolvePaths(graph.Options{RepoPath: projectPath})
	if err != nil {
		t.Fatalf("ResolvePaths: %v", err)
	}
	if err := os.MkdirAll(paths.OutFull, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	g := &graph.Graph{Nodes: nodes}
	if err := graph.SaveJSON(g, paths.ProjectJSON); err != nil {
		t.Fatalf("SaveJSON: %v", err)
	}
}

func readSymbolsIndex(t *testing.T, analysisDir string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(analysisDir, "project.symbols.md"))
	if err != nil {
		t.Fatalf("reading project.symbols.md: %v", err)
	}
	return string(b)
}

func TestGenerateSymbolsIndexListsSymbolsPerFile(t *testing.T) {
	dir := t.TempDir()
	writeGraphFixture(t, dir, []graph.Node{
		{ID: "a.go", Symbols: []string{"DoThing", "helper"}},
		{ID: "b.go", Symbols: []string{"OtherThing"}},
	})

	if err := generateSymbolsIndex(dir, dir); err != nil {
		t.Fatalf("generateSymbolsIndex: %v", err)
	}
	got := readSymbolsIndex(t, dir)

	for _, want := range []string{"## a.go", "`DoThing`", "`helper`", "## b.go", "`OtherThing`"} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q\n---\n%s", want, got)
		}
	}
	// No line-number suffixes on the listed symbols themselves (e.g. "a.go:12").
	if strings.Contains(got, "a.go:") || strings.Contains(got, "b.go:") {
		t.Errorf("output should not contain line numbers:\n%s", got)
	}
}

func TestGenerateSymbolsIndexFlagsCrossFileDuplicates(t *testing.T) {
	dir := t.TempDir()
	writeGraphFixture(t, dir, []graph.Node{
		{ID: "a.go", Symbols: []string{"New"}},
		{ID: "b.go", Symbols: []string{"New"}},
		{ID: "c.go", Symbols: []string{"Unique"}},
	})

	if err := generateSymbolsIndex(dir, dir); err != nil {
		t.Fatalf("generateSymbolsIndex: %v", err)
	}
	got := readSymbolsIndex(t, dir)

	if !strings.Contains(got, "## Duplicate Names") {
		t.Fatalf("missing duplicate names section:\n%s", got)
	}
	if !strings.Contains(got, "`New` — a.go, b.go") {
		t.Errorf("expected New flagged as a cross-file duplicate:\n%s", got)
	}
	if strings.Contains(got, "`Unique`") && strings.Contains(got, "Unique —") {
		t.Errorf("Unique should not appear in the duplicates section:\n%s", got)
	}
}

func TestGenerateSymbolsIndexCollapsesSameFileRepeats(t *testing.T) {
	dir := t.TempDir()
	writeGraphFixture(t, dir, []graph.Node{
		{ID: "a_test.go", Symbols: []string{"RoundTrip", "RoundTrip"}},
	})

	if err := generateSymbolsIndex(dir, dir); err != nil {
		t.Fatalf("generateSymbolsIndex: %v", err)
	}
	got := readSymbolsIndex(t, dir)

	if !strings.Contains(got, "`RoundTrip` — a_test.go (x2)") {
		t.Errorf("expected collapsed same-file repeat as 'a_test.go (x2)':\n%s", got)
	}
	if strings.Contains(got, "a_test.go, a_test.go") {
		t.Errorf("same file should not be listed twice:\n%s", got)
	}
}

func TestGenerateSymbolsIndexExcludesMarkdownFromDuplicates(t *testing.T) {
	dir := t.TempDir()
	writeGraphFixture(t, dir, []graph.Node{
		{ID: "a.md", Symbols: []string{"Overview"}},
		{ID: "b.md", Symbols: []string{"Overview"}},
	})

	if err := generateSymbolsIndex(dir, dir); err != nil {
		t.Fatalf("generateSymbolsIndex: %v", err)
	}
	got := readSymbolsIndex(t, dir)

	if !strings.Contains(got, "## Duplicate Names (appear more than once)\n\n_None found._") {
		t.Errorf("repeated markdown heading must not be flagged as a duplicate:\n%s", got)
	}
}

func TestGenerateSymbolsIndexFlagsSimilarNamesAcrossConventions(t *testing.T) {
	dir := t.TempDir()
	writeGraphFixture(t, dir, []graph.Node{
		{ID: "a.go", Symbols: []string{"getUserById"}},
		{ID: "b.py", Symbols: []string{"get_user_by_id"}},
		{ID: "c.java", Symbols: []string{"GetUserById"}},
	})

	if err := generateSymbolsIndex(dir, dir); err != nil {
		t.Fatalf("generateSymbolsIndex: %v", err)
	}
	got := readSymbolsIndex(t, dir)

	if !strings.Contains(got, "## Similar Names") {
		t.Fatalf("missing similar names section:\n%s", got)
	}
	for _, want := range []string{"`getUserById` (a.go)", "`get_user_by_id` (b.py)", "`GetUserById` (c.java)"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in similar-names line:\n%s", want, got)
		}
	}
}

func TestGenerateSymbolsIndexNoSymbols(t *testing.T) {
	dir := t.TempDir()
	writeGraphFixture(t, dir, []graph.Node{{ID: "a.go", Symbols: nil}})

	if err := generateSymbolsIndex(dir, dir); err != nil {
		t.Fatalf("generateSymbolsIndex: %v", err)
	}
	got := readSymbolsIndex(t, dir)

	if !strings.Contains(got, "_No symbols extracted for this project's current file set._") {
		t.Errorf("expected the empty-project message:\n%s", got)
	}
}
