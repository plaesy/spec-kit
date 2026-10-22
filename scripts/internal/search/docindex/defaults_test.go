package docindex

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/plaesy/spec-kit/internal/config"
	"github.com/plaesy/spec-kit/internal/search/retriever"
	"github.com/plaesy/spec-kit/internal/search/store"
	"github.com/plaesy/spec-kit/internal/search/textindex"
	"github.com/plaesy/spec-kit/internal/search/vectorstore"
)

// The default constructors are what production runs. These tests exercise them
// only where they never load the embedding model: an unknown type is refused
// before any model work, and the vector store, text index, and document store
// are real but local, on a temporary directory.

func TestDefaultVectorStore_RefusesUnknownTypeAndOpensChromem(t *testing.T) {
	if _, err := newVectorStore(vectorstore.Config{Type: "nope"}, t.TempDir()); err == nil {
		t.Error("an unknown vector store type was accepted")
	}
	vs, err := newVectorStore(vectorstore.Config{Type: "chromem"}, t.TempDir())
	if err != nil {
		t.Fatalf("chromem store: %v", err)
	}
	if err := vs.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}

func TestDefaultTextIndex_RefusesUnknownTypeAndOpensBleve(t *testing.T) {
	if _, err := newTextIndex(textindex.Config{Type: "nope"}, t.TempDir()); err == nil {
		t.Error("an unknown text index type was accepted")
	}
	ti, err := newTextIndex(textindex.Config{Type: "bleve"}, t.TempDir())
	if err != nil {
		t.Fatalf("bleve index: %v", err)
	}
	if err := ti.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}

func TestDefaultDocStore_OpensAndRefusesAFileAsBase(t *testing.T) {
	ds, err := newDocStore(t.TempDir())
	if err != nil {
		t.Fatalf("document store: %v", err)
	}
	if ds == nil {
		t.Fatal("document store is nil")
	}

	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := newDocStore(blocker); err == nil {
		t.Error("a file used as the store base was accepted")
	}
}

func TestDefaultIndexer_RefusesUnknownEmbedder(t *testing.T) {
	cfg := config.DefaultSearchConfig()
	cfg.Embedder.Type = "bogus" // refused before any model is loaded
	if _, err := newIndexer(cfg, t.TempDir()); err == nil {
		t.Error("an unknown embedder type built an indexer")
	}
}

func TestDefaultEmbedder_RefusesUnknownType(t *testing.T) {
	if _, err := newEmbedder(&config.EmbedderConfig{Type: "bogus"}); err == nil {
		t.Error("an unknown embedder type was accepted")
	}
}

// Parent-child results are capped at top, just like plain results.
func TestQueryWith_ParentChildTopLimitsSections(t *testing.T) {
	root := t.TempDir()
	guide := filepath.Join(root, "docs", "guide.md")
	installSeams(t, &fakeStack{
		vs: &fakeVectorStore{hits: []vectorstore.SearchResult{
			{ID: "c1", Score: 0.9, Content: "intro one", Metadata: docMeta(guide, "d1", nil)},
			{ID: "c3", Score: 0.8, Content: "usage text", Metadata: docMeta(guide, "d1", nil)},
		}},
		ds: &fakeDocStore{docs: map[string]*store.ProcessedDocument{"d1": parentFixture()}},
	})

	got, err := QueryWith(context.Background(), root, t.TempDir(), "q", 1, config.DefaultSearchConfig(), retriever.QueryOptions{}, true)
	if err != nil {
		t.Fatalf("QueryWith: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("top 1 returned %d sections", len(got))
	}
}

// A file the walk lists but cannot be read makes the fingerprint fail rather
// than be silently skipped. A dangling symlink is the portable way to get one.
func TestFingerprint_UnreadableDocumentIsReported(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "docs", "broken.md")
	if err := os.Symlink(filepath.Join(root, "missing-target.md"), link); err != nil {
		t.Skipf("cannot create a symlink on this system: %v", err)
	}
	if _, err := Fingerprint(root); err == nil || !strings.Contains(err.Error(), "broken.md") {
		t.Errorf("err = %v, want a read error naming broken.md", err)
	}
}
