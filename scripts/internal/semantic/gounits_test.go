package semantic

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/plaesy/spec-kit/internal/graph"
)

// unitsFixture has known line numbers. The comment on line 3 is separated from
// the func by a blank line, so it is not a doc comment and is not a unit.
const unitsFixture = `package p

// Free-standing comment, separated by a blank line.

// Add returns the sum.
// It adds two numbers.
func Add(a, b int) int {
	return a + b
}

type (
	// Pair holds two values.
	Pair struct {
		A, B int
	}

	Loose int
)

type Single struct{}

func (p Pair) Sum() int {
	return p.A + p.B
}
`

func TestGoUnits_ExactRangesAndDocs(t *testing.T) {
	units, err := goUnits(unitsFixture)
	if err != nil {
		t.Fatalf("goUnits: %v", err)
	}
	want := []Unit{
		{Name: "Add", Kind: "func", Doc: "Add returns the sum.", Start: 5, End: 9},
		{Name: "Pair", Kind: "type", Doc: "Pair holds two values.", Start: 12, End: 15},
		{Name: "Loose", Kind: "type", Doc: "", Start: 17, End: 17},
		{Name: "Single", Kind: "type", Doc: "", Start: 20, End: 20},
		{Name: "Sum", Kind: "method", Doc: "", Start: 22, End: 24},
	}
	if len(units) != len(want) {
		t.Fatalf("got %d units %+v, want %d", len(units), units, len(want))
	}
	for i := range want {
		if units[i] != want[i] {
			t.Errorf("unit %d = %+v, want %+v", i, units[i], want[i])
		}
	}
}

func TestGoUnits_FreeStandingCommentIsNotAUnit(t *testing.T) {
	units, err := goUnits("package p\n\n// TODO: later\n\nvar X = 1\n")
	if err != nil {
		t.Fatalf("goUnits: %v", err)
	}
	if len(units) != 0 {
		t.Errorf("got units %+v, want none: a comment and a var are not declarations here", units)
	}
}

func TestGoUnits_ParseErrorIsReported(t *testing.T) {
	if _, err := goUnits("this is not valid go {{{"); err == nil {
		t.Error("expected a parse error for invalid Go source")
	}
}

// TestBuildIndex_SameLineDeclarationsKeepBothUnits: two declarations on one line
// must both reach the index. An ID built from the line range alone would make
// the second overwrite the first.
func TestBuildIndex_SameLineDeclarationsKeepBothUnits(t *testing.T) {
	dir := t.TempDir()
	src := "package p\n\nfunc A() {}; func B() {}\n\nvar _ = 0\n"
	if err := os.WriteFile(filepath.Join(dir, "same.go"), []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}
	g := &graph.Graph{Nodes: []graph.Node{{ID: "same.go", Symbols: []string{"A", "B"}}}}
	emb := &fakeEmbedder{}
	outDir := t.TempDir()
	if err := BuildIndex(context.Background(), g, outDir, dir, emb); err != nil {
		t.Fatalf("BuildIndex: %v", err)
	}
	matches, err := Query(context.Background(), outDir, "anything", 10, emb)
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	names := map[string]bool{}
	for _, m := range matches {
		if m.Start == 3 {
			names[m.Name] = true
		}
	}
	if !names["A"] || !names["B"] {
		t.Errorf("same-line units found %v, want both A and B", names)
	}
}

// TestBuildIndex_GoDeclarationUnits checks the whole path: a real Go file on
// disk becomes one document per declaration, Query returns the range as
// metadata, and an unparseable Go file falls back to one file document.
func TestBuildIndex_GoDeclarationUnits(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "p.go"), []byte(unitsFixture), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bad.go"), []byte("this is not valid go {{{"), 0o600); err != nil {
		t.Fatal(err)
	}
	g := &graph.Graph{Nodes: []graph.Node{
		{ID: "p.go", Symbols: []string{"Add", "Pair", "Loose", "Single", "Sum"}},
		{ID: "bad.go", Symbols: []string{"Broken"}},
	}}
	emb := &fakeEmbedder{}
	outDir := t.TempDir()
	if err := BuildIndex(context.Background(), g, outDir, dir, emb); err != nil {
		t.Fatalf("BuildIndex: %v", err)
	}

	matches, err := Query(context.Background(), outDir, "anything", 20, emb)
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	got := map[string]Match{}
	for _, m := range matches {
		got[m.Kind+"|"+m.Name] = m
	}

	add, ok := got["func|Add"]
	if !ok || add.NodeID != "p.go" || add.Start != 5 || add.End != 9 {
		t.Errorf("func Add match = %+v (found %v), want p.go lines 5-9", add, ok)
	}
	sum, ok := got["method|Sum"]
	if !ok || sum.Start != 22 || sum.End != 24 {
		t.Errorf("method Sum match = %+v (found %v), want lines 22-24", sum, ok)
	}
	if file, ok := got["file|"]; !ok || file.NodeID != "bad.go" || file.Start != 0 {
		t.Errorf("unparseable bad.go = %+v (found %v), want one file-level document", file, ok)
	}
	for _, m := range matches {
		if m.NodeID == "p.go" && m.Kind == "file" {
			t.Errorf("p.go was also indexed as a whole file: %+v", m)
		}
	}
}
