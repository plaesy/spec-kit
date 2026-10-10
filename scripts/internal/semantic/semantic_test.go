package semantic

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/plaesy/spec-kit/internal/graph"
)

// fakeEmbedder is a deterministic stand-in for rembed in tests: it never
// downloads a model or touches the network. It maps each text to a small
// hand-picked vector so similarity ordering is predictable.
type fakeEmbedder struct {
	vectors map[string][]float32
}

func (f *fakeEmbedder) Embed(_ context.Context, texts []string) ([][]float32, error) {
	out := make([][]float32, len(texts))
	for i, t := range texts {
		out[i] = f.vectorFor(t)
	}
	return out, nil
}

// vectorFor buckets a text into one of three directions based on keywords
// it contains, so "purpose-similar" texts (sharing a keyword) land close
// together and everything else lands orthogonal to them — enough signal to
// test ranking without a real model.
func (f *fakeEmbedder) vectorFor(text string) []float32 {
	lower := strings.ToLower(text)
	switch {
	case strings.Contains(lower, "validateuser") || strings.Contains(lower, "checkuserisvalid"):
		return []float32{1, 0, 0}
	case strings.Contains(lower, "formatdate") || strings.Contains(lower, "renderdate"):
		return []float32{0, 1, 0}
	default:
		return []float32{0, 0, 1}
	}
}

func testGraph() *graph.Graph {
	return &graph.Graph{
		Nodes: []graph.Node{
			{ID: "auth/validate.go", Symbols: []string{"ValidateUser", "hashPassword"}},
			{ID: "auth/check.go", Symbols: []string{"CheckUserIsValid"}},
			{ID: "ui/date.go", Symbols: []string{"FormatDate"}},
			{ID: "ui/display.go", Symbols: []string{"RenderDate"}},
			{ID: "empty/noop.go", Symbols: nil},
		},
	}
}

func TestBuildIndex_SkipsNodesWithNoSymbols(t *testing.T) {
	dir := t.TempDir()
	emb := &fakeEmbedder{}
	if err := BuildIndex(context.Background(), testGraph(), dir, dir, emb); err != nil {
		t.Fatalf("BuildIndex: %v", err)
	}

	matches, err := Query(context.Background(), dir, "ValidateUser", 10, emb)
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	for _, m := range matches {
		if m.NodeID == "empty/noop.go" {
			t.Errorf("expected node with no symbols to be excluded, got it in results: %+v", matches)
		}
	}
}

func TestQuery_RanksSemanticMatchAboveUnrelated(t *testing.T) {
	dir := t.TempDir()
	emb := &fakeEmbedder{}
	if err := BuildIndex(context.Background(), testGraph(), dir, dir, emb); err != nil {
		t.Fatalf("BuildIndex: %v", err)
	}

	matches, err := Query(context.Background(), dir, "ValidateUser", 4, emb)
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(matches) < 2 {
		t.Fatalf("expected at least 2 matches, got %d", len(matches))
	}

	// Both auth/validate.go and auth/check.go share the "same purpose"
	// vector and so are exactly tied — the embedder can't distinguish them
	// further, and which one sorts first isn't meaningful. What matters is
	// that both outrank the unrelated ui/* nodes.
	top2 := map[string]bool{matches[0].NodeID: true, matches[1].NodeID: true}
	if !top2["auth/validate.go"] || !top2["auth/check.go"] {
		t.Errorf("top 2 matches = %v, want {auth/validate.go, auth/check.go}", matches[:2])
	}
	for _, m := range matches[2:] {
		if m.NodeID == "auth/validate.go" || m.NodeID == "auth/check.go" {
			t.Errorf("auth/* node ranked below an unrelated node: %+v", matches)
		}
	}
}

func TestQuery_WithoutPriorBuildIndex_ErrorsInsteadOfPanicking(t *testing.T) {
	dir := t.TempDir()
	emb := &fakeEmbedder{}
	if _, err := Query(context.Background(), dir, "anything", 5, emb); err == nil {
		t.Fatal("expected an error when querying before BuildIndex, got nil")
	}
}

func TestBuildIndex_EmptyGraph_NoError(t *testing.T) {
	dir := t.TempDir()
	emb := &fakeEmbedder{}
	if err := BuildIndex(context.Background(), &graph.Graph{}, dir, dir, emb); err != nil {
		t.Fatalf("BuildIndex on empty graph: %v", err)
	}
}

func TestQuery_ReturnsMatchedSymbols(t *testing.T) {
	dir := t.TempDir()
	emb := &fakeEmbedder{}
	if err := BuildIndex(context.Background(), testGraph(), dir, dir, emb); err != nil {
		t.Fatalf("BuildIndex: %v", err)
	}

	matches, err := Query(context.Background(), dir, "ValidateUser", 1, emb)
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	got := matches[0]
	switch got.NodeID {
	case "auth/validate.go":
		if !contains(got.Symbols, "ValidateUser") || !contains(got.Symbols, "hashPassword") {
			t.Errorf("Symbols = %v, want both ValidateUser and hashPassword", got.Symbols)
		}
	case "auth/check.go":
		if !contains(got.Symbols, "CheckUserIsValid") {
			t.Errorf("Symbols = %v, want CheckUserIsValid", got.Symbols)
		}
	default:
		t.Fatalf("unexpected top match %q", got.NodeID)
	}
}

func contains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

func TestNodeText_IncludesIDAndSymbols(t *testing.T) {
	n := graph.Node{ID: "foo.go", Symbols: []string{"Bar", "Baz"}}
	text := nodeText(t.TempDir(), n)
	for _, want := range []string{"foo.go", "Bar", "Baz"} {
		if !strings.Contains(text, want) {
			t.Errorf("nodeText(%+v) = %q, missing %q", n, text, want)
		}
	}
}

func TestNodeText_IncludesGoDocComment(t *testing.T) {
	repoRoot := t.TempDir()
	writeFile(t, filepath.Join(repoRoot, "pkg", "auth.go"), `package pkg

// ValidateUser checks the given credentials against the user store.
func ValidateUser(user, pass string) bool { return true }
`)
	n := graph.Node{ID: "pkg/auth.go", Symbols: []string{"ValidateUser"}}

	text := nodeText(repoRoot, n)
	want := "ValidateUser: ValidateUser checks the given credentials against the user store."
	if !strings.Contains(text, want) {
		t.Errorf("nodeText = %q, want it to contain %q", text, want)
	}
}

func TestNodeText_GoFileWithNoDocComment_FallsBackToBareSymbol(t *testing.T) {
	repoRoot := t.TempDir()
	writeFile(t, filepath.Join(repoRoot, "pkg", "nodoc.go"), `package pkg

func Undocumented() {}
`)
	n := graph.Node{ID: "pkg/nodoc.go", Symbols: []string{"Undocumented"}}

	text := nodeText(repoRoot, n)
	for _, line := range strings.Split(text, "\n") {
		if line == "Undocumented" {
			return
		}
	}
	t.Errorf("nodeText = %q, want a bare \"Undocumented\" line", text)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
