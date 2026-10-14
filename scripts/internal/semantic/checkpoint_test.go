package semantic

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/plaesy/spec-kit/internal/graph"
)

// interruptingEmbedder counts the texts it embeds and fails on the call numbered
// failAt (1-based; 0 never fails), to simulate an interrupted build.
type interruptingEmbedder struct {
	fakeEmbedder
	calls  int
	texts  int
	failAt int
}

func (c *interruptingEmbedder) EmbedPassages(ctx context.Context, texts []string) ([][]float32, error) {
	c.calls++
	if c.calls == c.failAt {
		return nil, errors.New("simulated interruption")
	}
	c.texts += len(texts)
	return c.fakeEmbedder.EmbedPassages(ctx, texts)
}

// checkpointGraph returns n file nodes with distinct symbols. Non-Go files have
// no source to read, so each is indexed as one file-level document.
func checkpointGraph(n int) *graph.Graph {
	g := &graph.Graph{}
	for i := range n {
		g.Nodes = append(g.Nodes, graph.Node{
			ID:      fmt.Sprintf("doc%03d.md", i),
			Symbols: []string{fmt.Sprintf("Section%03d", i)},
		})
	}
	return g
}

func TestBuildIndex_ResumesFromCheckpointAfterInterruption(t *testing.T) {
	const n = 150 // three batches of up to 64
	g := checkpointGraph(n)
	repoRoot := t.TempDir()
	outDir := t.TempDir()
	ctx := context.Background()
	ckPath := checkpointPath(outDir)

	// The second batch fails, so the first batch (64 documents) is checkpointed.
	interrupted := &interruptingEmbedder{failAt: 2}
	if _, err := buildIndex(ctx, g, outDir, repoRoot, interrupted); err == nil {
		t.Fatal("expected the interrupted build to fail")
	}
	if _, err := os.Stat(ckPath); err != nil {
		t.Fatalf("checkpoint was not written before the interruption: %v", err)
	}

	resumed := &interruptingEmbedder{}
	embedded, err := buildIndex(ctx, g, outDir, repoRoot, resumed)
	if err != nil {
		t.Fatalf("resumed build: %v", err)
	}
	if want := n - checkpointBatchSize; embedded != want || resumed.texts != want {
		t.Errorf("resumed build embedded %d (texts %d), want only the remaining %d", embedded, resumed.texts, want)
	}
	if _, err := os.Stat(ckPath); !os.IsNotExist(err) {
		t.Errorf("checkpoint should be removed after a finished build, stat err = %v", err)
	}

	matches, err := Query(ctx, outDir, "anything", n, &fakeEmbedder{})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(matches) != n {
		t.Errorf("index holds %d documents after resume, want %d", len(matches), n)
	}
}

func TestBuildIndex_DiscardsCheckpointWhenDocumentsChange(t *testing.T) {
	const n = 100
	repoRoot := t.TempDir()
	outDir := t.TempDir()
	ctx := context.Background()
	ckPath := checkpointPath(outDir)

	if _, err := buildIndex(ctx, checkpointGraph(n), outDir, repoRoot, &interruptingEmbedder{failAt: 2}); err == nil {
		t.Fatal("expected the interrupted build to fail")
	}

	// A changed document changes the digest, so the checkpoint is discarded and
	// every document is embedded again.
	changed := checkpointGraph(n)
	changed.Nodes[0].Symbols = []string{"Renamed"}
	resumed := &interruptingEmbedder{}
	embedded, err := buildIndex(ctx, changed, outDir, repoRoot, resumed)
	if err != nil {
		t.Fatalf("rebuild after change: %v", err)
	}
	if embedded != n {
		t.Errorf("rebuild after change embedded %d, want all %d", embedded, n)
	}
	if _, err := os.Stat(ckPath); !os.IsNotExist(err) {
		t.Errorf("checkpoint should be removed after the rebuild, stat err = %v", err)
	}
}

func TestLoadCheckpoint_IgnoresTruncatedLastLine(t *testing.T) {
	docs := indexDocs(t.TempDir(), []graph.Node{{ID: "a.md", Symbols: []string{"A"}}})
	digest := inputDigest(docs)
	path := filepath.Join(t.TempDir(), checkpointFile)

	entry := checkpointEntry{ID: docs[0].id, Hash: docs[0].meta[metaHash], Vector: []float32{1, 0, 0}}
	if err := writeCheckpointBatch(path, digest, []checkpointEntry{entry}, false); err != nil {
		t.Fatalf("write batch: %v", err)
	}
	// Simulate a write cut off mid-line by a kill.
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(`{"id":"b.md","hash":"x","vec`); err != nil {
		t.Fatal(err)
	}
	f.Close()

	vectors, valid := loadCheckpoint(path, digest, docs)
	if !valid {
		t.Fatal("header should match, so the checkpoint is valid")
	}
	if len(vectors) != 1 || vectors[docs[0].id] == nil {
		t.Errorf("loaded %v, want only the complete entry", vectors)
	}
}
