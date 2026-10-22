package eval

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/plaesy/spec-kit/internal/search/retriever"
	"github.com/plaesy/spec-kit/internal/search/store"
	"github.com/plaesy/spec-kit/internal/search/textindex"
	"github.com/plaesy/spec-kit/internal/search/vectorstore"
)

type fakeEmbedder struct{ err error }

func (f fakeEmbedder) EmbedQuery(_ context.Context, _ string) ([]float32, error) {
	if f.err != nil {
		return nil, f.err
	}
	return []float32{1, 2}, nil
}

type fakeVec struct{ hits []vectorstore.SearchResult }

func (f fakeVec) Search(_ context.Context, _ []float32, _ int, _ map[string]string) ([]vectorstore.SearchResult, error) {
	return f.hits, nil
}

type fakeText struct{ hits []textindex.SearchResult }

func (f fakeText) Search(_ context.Context, _ string, _ int, _ map[string]string) ([]textindex.SearchResult, error) {
	return f.hits, nil
}

type fakeHybrid struct{ hits []retriever.Result }

func (f fakeHybrid) Retrieve(_ context.Context, _ string, _ map[string]string) ([]retriever.Result, error) {
	return f.hits, nil
}

// fakeDocs counts lookups so the cache can be checked.
type fakeDocs struct {
	paths map[string]string
	err   error
	calls map[string]int
}

func (f *fakeDocs) GetRaw(_ context.Context, id string) (*store.RawDocument, error) {
	if f.calls != nil {
		f.calls[id]++
	}
	if f.err != nil {
		return nil, f.err
	}
	p, ok := f.paths[id]
	if !ok {
		return nil, errors.New("not found: " + id)
	}
	return &store.RawDocument{ID: id, Path: p}, nil
}

func newTestDocs(root string) *fakeDocs {
	return &fakeDocs{
		paths: map[string]string{
			"d1": filepath.Join(root, "docs", "a.md"), // absolute, under root
			"d2": filepath.Join(root, "b.go"),
			"d3": filepath.Join(root, "c.go"),
		},
		calls: map[string]int{},
	}
}

func TestIndexEngine_VectorMode_DedupesAndKeepsRankOrder(t *testing.T) {
	root := t.TempDir()
	// Scores are exact in float32 so the float64 comparison is exact.
	vec := fakeVec{hits: []vectorstore.SearchResult{
		{Score: 0.5, Metadata: map[string]string{"doc_id": "d1"}},
		{Score: 0.25, Metadata: map[string]string{"doc_id": "d1"}}, // second chunk of d1
		{Score: 0.125, Metadata: map[string]string{"doc_id": "d2"}},
	}}
	e := NewIndexEngine(root, fakeEmbedder{}, vec, fakeText{}, fakeHybrid{}, newTestDocs(root))

	paths, scores, err := e.Search("q", 10, ModeVector)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []string{"docs/a.md", "b.go"}; !reflect.DeepEqual(paths, want) {
		t.Errorf("paths = %v, want %v", paths, want)
	}
	if want := []float64{0.5, 0.125}; !reflect.DeepEqual(scores, want) {
		t.Errorf("scores = %v, want %v (first occurrence keeps its score)", scores, want)
	}
}

func TestIndexEngine_BM25Mode(t *testing.T) {
	root := t.TempDir()
	text := fakeText{hits: []textindex.SearchResult{
		{Score: 3.5, Metadata: map[string]string{"doc_id": "d3"}},
	}}
	e := NewIndexEngine(root, fakeEmbedder{}, fakeVec{}, text, fakeHybrid{}, newTestDocs(root))

	paths, scores, err := e.Search("q", 10, ModeBM25)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(paths, []string{"c.go"}) || !reflect.DeepEqual(scores, []float64{3.5}) {
		t.Errorf("got %v %v, want [c.go] [3.5]", paths, scores)
	}
}

func TestIndexEngine_HybridMode_UsesRetrieverDocID(t *testing.T) {
	root := t.TempDir()
	hyb := fakeHybrid{hits: []retriever.Result{
		{ID: "chunk-1", DocID: "d2", Score: 0.031},
	}}
	e := NewIndexEngine(root, fakeEmbedder{}, fakeVec{}, fakeText{}, hyb, newTestDocs(root))

	paths, _, err := e.Search("q", 10, ModeHybrid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(paths, []string{"b.go"}) {
		t.Errorf("paths = %v, want [b.go]", paths)
	}
}

func TestIndexEngine_TopKCountsDistinctFiles(t *testing.T) {
	root := t.TempDir()
	vec := fakeVec{hits: []vectorstore.SearchResult{
		{Score: 0.9, Metadata: map[string]string{"doc_id": "d1"}},
		{Score: 0.8, Metadata: map[string]string{"doc_id": "d1"}},
		{Score: 0.7, Metadata: map[string]string{"doc_id": "d2"}},
	}}
	e := NewIndexEngine(root, fakeEmbedder{}, vec, fakeText{}, fakeHybrid{}, newTestDocs(root))

	// topK=2 must return two distinct files, not two chunks of d1.
	paths, _, err := e.Search("q", 2, ModeVector)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []string{"docs/a.md", "b.go"}; !reflect.DeepEqual(paths, want) {
		t.Errorf("paths = %v, want %v", paths, want)
	}
}

func TestIndexEngine_ReturnsSentinelForUnsupportedMode(t *testing.T) {
	root := t.TempDir()
	e := NewIndexEngine(root, fakeEmbedder{}, fakeVec{}, fakeText{}, fakeHybrid{}, newTestDocs(root))

	_, _, err := e.Search("q", 10, ModeReranked)
	if !errors.Is(err, ErrModeUnsupported) {
		t.Fatalf("err = %v, want ErrModeUnsupported", err)
	}
}

func TestIndexEngine_PropagatesDocLookupError(t *testing.T) {
	root := t.TempDir()
	docs := &fakeDocs{err: errors.New("disk gone")}
	vec := fakeVec{hits: []vectorstore.SearchResult{
		{Score: 0.9, Metadata: map[string]string{"doc_id": "d1"}},
	}}
	e := NewIndexEngine(root, fakeEmbedder{}, vec, fakeText{}, fakeHybrid{}, docs)

	if _, _, err := e.Search("q", 10, ModeVector); err == nil {
		t.Fatal("expected lookup error to propagate, got nil (a silent skip would hide a broken index)")
	}
}

func TestIndexEngine_PropagatesEmbedError(t *testing.T) {
	root := t.TempDir()
	e := NewIndexEngine(root, fakeEmbedder{err: errors.New("model missing")}, fakeVec{}, fakeText{}, fakeHybrid{}, newTestDocs(root))

	if _, _, err := e.Search("q", 10, ModeVector); err == nil {
		t.Fatal("expected embed error to propagate")
	}
}

func TestIndexEngine_RejectsHitWithoutDocID(t *testing.T) {
	root := t.TempDir()
	text := fakeText{hits: []textindex.SearchResult{
		{Score: 1, Metadata: map[string]string{}}, // no doc_id: indexer contract broken
	}}
	e := NewIndexEngine(root, fakeEmbedder{}, fakeVec{}, text, fakeHybrid{}, newTestDocs(root))

	if _, _, err := e.Search("q", 10, ModeBM25); err == nil {
		t.Fatal("expected error for hit without doc_id")
	}
}

func TestIndexEngine_CachesPathLookups(t *testing.T) {
	root := t.TempDir()
	docs := newTestDocs(root)
	vec := fakeVec{hits: []vectorstore.SearchResult{
		{Score: 0.9, Metadata: map[string]string{"doc_id": "d1"}},
		{Score: 0.8, Metadata: map[string]string{"doc_id": "d1"}},
	}}
	e := NewIndexEngine(root, fakeEmbedder{}, vec, fakeText{}, fakeHybrid{}, docs)

	for i := 0; i < 3; i++ {
		if _, _, err := e.Search("q", 10, ModeVector); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if docs.calls["d1"] != 1 {
		t.Errorf("GetRaw(d1) called %d times, want 1 (cache miss on every query)", docs.calls["d1"])
	}
}
