package semantic

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/plaesy/spec-kit/internal/graph"
)

// countingEmbedder counts how many texts were actually embedded, so a test can
// assert that an unchanged rebuild embeds nothing.
type countingEmbedder struct {
	inner *fakeEmbedder
	count int
}

func (c *countingEmbedder) EmbedPassages(ctx context.Context, texts []string) ([][]float32, error) {
	c.count += len(texts)
	return c.inner.EmbedPassages(ctx, texts)
}

func (c *countingEmbedder) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	return c.inner.EmbedQuery(ctx, text)
}

func TestIncrementalBuild_UnchangedRebuildEmbedsNothing(t *testing.T) {
	repo := t.TempDir()
	out := t.TempDir()
	g := testGraph()
	emb := &countingEmbedder{inner: &fakeEmbedder{}}

	first, err := buildIndex(context.Background(), g, out, repo, emb)
	if err != nil {
		t.Fatalf("first build: %v", err)
	}
	if first != 4 { // testGraph has four nodes with symbols; empty/noop.go is skipped
		t.Fatalf("first build embedded %d, want 4", first)
	}

	emb.count = 0
	second, err := buildIndex(context.Background(), g, out, repo, emb)
	if err != nil {
		t.Fatalf("second build: %v", err)
	}
	if second != 0 || emb.count != 0 {
		t.Errorf("unchanged rebuild embedded %d (counter %d), want 0", second, emb.count)
	}
}

func TestIncrementalBuild_ChangedNodeIsReembeddedAlone(t *testing.T) {
	repo := t.TempDir()
	out := t.TempDir()
	emb := &countingEmbedder{inner: &fakeEmbedder{}}
	if _, err := buildIndex(context.Background(), testGraph(), out, repo, emb); err != nil {
		t.Fatalf("first build: %v", err)
	}

	changed := testGraph()
	changed.Nodes[2].Symbols = []string{"FormatDate", "newHelper"} // ui/date.go only
	emb.count = 0
	n, err := buildIndex(context.Background(), changed, out, repo, emb)
	if err != nil {
		t.Fatalf("second build: %v", err)
	}
	if n != 1 {
		t.Errorf("changed rebuild embedded %d documents, want 1 (only ui/date.go)", n)
	}
}

func TestIncrementalBuild_DeletedNodeIsOrphanRemoved(t *testing.T) {
	repo := t.TempDir()
	out := t.TempDir()
	emb := &countingEmbedder{inner: &fakeEmbedder{}}
	if _, err := buildIndex(context.Background(), testGraph(), out, repo, emb); err != nil {
		t.Fatalf("first build: %v", err)
	}

	smaller := &graph.Graph{Nodes: testGraph().Nodes[:2]} // keep auth/* only
	if _, err := buildIndex(context.Background(), smaller, out, repo, emb); err != nil {
		t.Fatalf("second build: %v", err)
	}
	matches, err := Query(context.Background(), out, "anything", 10, emb)
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(matches) != 2 {
		t.Errorf("after deleting two nodes, index holds %d documents, want 2", len(matches))
	}
	for _, m := range matches {
		if m.NodeID == "ui/date.go" || m.NodeID == "ui/display.go" {
			t.Errorf("orphan %q still indexed", m.NodeID)
		}
	}
}

// TestIncrementalBuild_BodyOnlyEditEmbedsNothing confirms the hash covers the
// embedded text, not the file bytes: a body-only edit keeps the same unit text,
// so nothing is re-embedded.
func TestIncrementalBuild_BodyOnlyEditEmbedsNothing(t *testing.T) {
	repo := t.TempDir()
	out := t.TempDir()
	src := "package p\n\n// Add adds.\nfunc Add(a, b int) int {\n\treturn a + b\n}\n"
	if err := os.WriteFile(filepath.Join(repo, "p.go"), []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}
	g := &graph.Graph{Nodes: []graph.Node{{ID: "p.go", Symbols: []string{"Add"}}}}
	emb := &countingEmbedder{inner: &fakeEmbedder{}}
	if _, err := buildIndex(context.Background(), g, out, repo, emb); err != nil {
		t.Fatalf("first build: %v", err)
	}

	edited := "package p\n\n// Add adds.\nfunc Add(a, b int) int {\n\treturn b + a\n}\n"
	if err := os.WriteFile(filepath.Join(repo, "p.go"), []byte(edited), 0o600); err != nil {
		t.Fatal(err)
	}
	emb.count = 0
	n, err := buildIndex(context.Background(), g, out, repo, emb)
	if err != nil {
		t.Fatalf("second build: %v", err)
	}
	if n != 0 {
		t.Errorf("body-only edit embedded %d documents, want 0", n)
	}
}
