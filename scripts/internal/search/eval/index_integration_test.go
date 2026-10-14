package eval

import (
	"context"
	"hash/fnv"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/plaesy/spec-kit/internal/config"
	"github.com/plaesy/spec-kit/internal/search/parser"
	"github.com/plaesy/spec-kit/internal/search/retriever"
	"github.com/plaesy/spec-kit/internal/search/store"
	"github.com/plaesy/spec-kit/internal/search/textindex"
	"github.com/plaesy/spec-kit/internal/search/vectorstore"
)

// bagEmbedder maps each word to one of 64 buckets and L2-normalises the
// counts. It is deterministic and needs no model, so the ranking is
// predictable while the real stores still run end to end.
type bagEmbedder struct{}

const bagDims = 64

func (bagEmbedder) vector(text string) []float32 {
	v := make([]float32, bagDims)
	for _, w := range strings.Fields(strings.ToLower(text)) {
		h := fnv.New32a()
		_, _ = h.Write([]byte(w))
		v[h.Sum32()%bagDims]++
	}
	var norm float64
	for _, x := range v {
		norm += float64(x) * float64(x)
	}
	if norm == 0 {
		return v
	}
	n := float32(math.Sqrt(norm))
	for i := range v {
		v[i] /= n
	}
	return v
}

func (b bagEmbedder) EmbedQuery(_ context.Context, query string) ([]float32, error) {
	return b.vector(query), nil
}

func (b bagEmbedder) EmbedPassages(_ context.Context, passages []string) ([][]float32, error) {
	out := make([][]float32, len(passages))
	for i, p := range passages {
		out[i] = b.vector(p)
	}
	return out, nil
}

// TestIndexEngine_EndToEnd runs the adapter over the real DocumentStore,
// ChromemStore, BleveIndex and HybridRetriever. Only the embedder is fake.
// It checks that hits come back as project-relative file paths.
func TestIndexEngine_EndToEnd(t *testing.T) {
	ctx := context.Background()

	root := t.TempDir()
	files := map[string]string{
		"docs/bst.md":   "binary search tree insertion and tree traversal",
		"docs/rnn.md":   "recurrent neural network with LSTM gating",
		"docs/yaml.md":  "kubernetes deployment yaml manifests",
		"notes/misc.md": "weekly team notes about the release",
	}
	for rel, body := range files {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	indexDir := t.TempDir()
	docStore, err := store.NewDocumentStore(indexDir)
	if err != nil {
		t.Fatalf("document store: %v", err)
	}
	vs, err := vectorstore.NewChromemStore(vectorstore.Config{Type: "chromem"}, indexDir)
	if err != nil {
		t.Fatalf("vector store: %v", err)
	}
	defer vs.Close()
	ti, err := textindex.NewBleveIndex(textindex.Config{Type: "bleve"}, indexDir)
	if err != nil {
		t.Fatalf("text index: %v", err)
	}
	defer ti.Close()

	// Index each file the way the pipeline does: store the raw document,
	// then add one chunk per document with doc_id metadata.
	var ids, docs []string
	var metas []map[string]string
	var vecs [][]float32
	for rel, body := range files {
		abs := filepath.Join(root, filepath.FromSlash(rel))
		raw, err := docStore.StoreRaw(ctx, parser.Document{Path: abs})
		if err != nil {
			t.Fatalf("store raw %s: %v", rel, err)
		}
		chunkID := raw.ID + "-c0"
		ids = append(ids, chunkID)
		docs = append(docs, body)
		metas = append(metas, map[string]string{"doc_id": raw.ID, "source": "markdown"})
		vecs = append(vecs, bagEmbedder{}.vector(body))
	}
	if err := vs.Add(ctx, ids, vecs, metas, docs); err != nil {
		t.Fatalf("vector add: %v", err)
	}
	if err := ti.Add(ctx, ids, docs, metas); err != nil {
		t.Fatalf("text add: %v", err)
	}

	hybrid := retriever.NewHybridRetriever(config.DefaultSearchConfig().Retriever, bagEmbedder{}, vs, ti, docStore)
	engine := NewIndexEngine(root, bagEmbedder{}, vs, ti, hybrid, docStore)

	const query = "binary search tree"
	for _, mode := range []SearchMode{ModeBM25, ModeVector, ModeHybrid} {
		paths, scores, err := engine.Search(query, 10, mode)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", mode, err)
		}
		if len(paths) == 0 {
			t.Fatalf("%s: no results", mode)
		}
		if paths[0] != "docs/bst.md" {
			t.Errorf("%s: top result = %q, want docs/bst.md (all: %v)", mode, paths[0], paths)
		}
		for _, p := range paths {
			if strings.Contains(p, `\`) || filepath.IsAbs(p) {
				t.Errorf("%s: path %q is not project-relative with forward slashes", mode, p)
			}
		}
		if len(scores) != len(paths) {
			t.Errorf("%s: %d scores for %d paths", mode, len(scores), len(paths))
		}
	}
}
