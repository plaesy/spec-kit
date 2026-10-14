package retriever

import (
	"context"
	"strings"
	"testing"

	"github.com/plaesy/spec-kit/internal/config"
	"github.com/plaesy/spec-kit/internal/search/chunker"
	"github.com/plaesy/spec-kit/internal/search/store"
	"github.com/plaesy/spec-kit/internal/search/textindex"
	"github.com/plaesy/spec-kit/internal/search/vectorstore"
)

// parentFixture is one document with three chunks in two sections. c1 and c2
// share "Intro", and c3 is alone in "Usage".
func parentFixture() *store.ProcessedDocument {
	return &store.ProcessedDocument{
		DocID: "d1",
		Chunks: []chunker.Chunk{
			{ID: "c1", DocID: "d1", SectionPath: "Intro", Content: "intro one", StartOffset: 0, EndOffset: 9},
			{ID: "c2", DocID: "d1", SectionPath: "Intro", Content: "intro two", StartOffset: 10, EndOffset: 19},
			{ID: "c3", DocID: "d1", SectionPath: "Usage", Content: "usage text", StartOffset: 20, EndOffset: 30},
		},
	}
}

func newParentChildFixture(t *testing.T, vectorHits, textHits map[string]float64) *ParentChildRetriever {
	t.Helper()
	meta := map[string]string{"doc_id": "d1", "path": "/r/docs/guide.md"}
	vs := &mockVectorStoreForRetriever{data: map[string]vectorstore.SearchResult{}}
	for id, score := range vectorHits {
		vs.data[id] = vectorstore.SearchResult{ID: id, Score: float32(score), Content: id, Metadata: meta}
	}
	ti := &mockTextIndexForRetriever{data: map[string]textindex.SearchResult{}}
	for id, score := range textHits {
		ti.data[id] = textindex.SearchResult{ID: id, Score: score, Content: id, Metadata: meta}
	}
	ds := &mockDocStoreForRetriever{docs: map[string]*store.ProcessedDocument{"d1": parentFixture()}}
	cfg := config.RetrieverConfig{
		TopK:        5,
		Hybrid:      config.HybridConfig{RRFK: 60, VectorWeight: 1, BM25Weight: 1},
		ParentChild: config.ParentChildConfig{Enabled: true},
	}
	h := NewHybridRetriever(cfg, &mockEmbedderForRetriever{dimension: 3}, vs, ti, ds)
	return NewParentChildRetriever(h)
}

// Two children in one section return one parent section, with both children as
// citations, and the section text covers every chunk of that section.
func TestParentChild_GroupsChildrenIntoOneSection(t *testing.T) {
	p := newParentChildFixture(t,
		map[string]float64{"c1": 0.9, "c2": 0.8},
		map[string]float64{"c2": 4})

	sections, err := p.Retrieve(context.Background(), "q", QueryOptions{})
	if err != nil {
		t.Fatalf("Retrieve: %v", err)
	}
	if len(sections) != 1 {
		t.Fatalf("got %d sections, want 1 (both children are in Intro): %+v", len(sections), sections)
	}
	s := sections[0]
	if s.SectionPath != "Intro" || s.DocID != "d1" {
		t.Errorf("section = %q in %q, want Intro in d1", s.SectionPath, s.DocID)
	}
	if s.Content != "intro one\n\nintro two" {
		t.Errorf("Content = %q, want both Intro chunks in document order", s.Content)
	}
	if len(s.Children) != 2 {
		t.Fatalf("got %d citations, want 2: %+v", len(s.Children), s.Children)
	}
	if s.Children[0].ChunkID != "c2" || s.Children[1].ChunkID != "c1" {
		t.Errorf("citations not best first: %+v", s.Children)
	}
	if s.Children[1].StartOffset != 0 || s.Children[1].EndOffset != 9 {
		t.Errorf("c1 offsets = %d-%d, want 0-9", s.Children[1].StartOffset, s.Children[1].EndOffset)
	}
}

// A child in another section gets its own parent section, and sections come
// back best first.
func TestParentChild_SeparateSectionsStaySeparate(t *testing.T) {
	p := newParentChildFixture(t,
		map[string]float64{"c3": 0.95, "c1": 0.5},
		nil)

	sections, err := p.Retrieve(context.Background(), "q", QueryOptions{})
	if err != nil {
		t.Fatalf("Retrieve: %v", err)
	}
	if len(sections) != 2 {
		t.Fatalf("got %d sections, want 2: %+v", len(sections), sections)
	}
	if sections[0].SectionPath != "Usage" || sections[1].SectionPath != "Intro" {
		t.Errorf("sections = [%q %q], want [Usage Intro] (best first)", sections[0].SectionPath, sections[1].SectionPath)
	}
	if sections[0].Content != "usage text" {
		t.Errorf("Usage content = %q, want only its own chunk", sections[0].Content)
	}
}

// Without a document store the wrapper cannot read section text, so it refuses
// instead of returning the child chunks as if they were sections.
func TestParentChild_RequiresDocStore(t *testing.T) {
	cfg := config.RetrieverConfig{TopK: 5, Hybrid: config.HybridConfig{RRFK: 60, VectorWeight: 1, BM25Weight: 1}}
	h := NewHybridRetriever(cfg, &mockEmbedderForRetriever{dimension: 3},
		&mockVectorStoreForRetriever{}, &mockTextIndexForRetriever{}, nil)
	_, err := NewParentChildRetriever(h).Retrieve(context.Background(), "q", QueryOptions{})
	if err == nil || !strings.Contains(err.Error(), "document store") {
		t.Errorf("err = %v, want a document-store error", err)
	}
}

// The predicate filter also applies to the parent-child path, so a section
// whose children are all filtered out never appears.
func TestParentChild_PredicateRemovesWholeSection(t *testing.T) {
	p := newParentChildFixture(t,
		map[string]float64{"c3": 0.9},
		nil)
	where, err := ParseFilter("path:elsewhere")
	if err != nil {
		t.Fatal(err)
	}
	sections, err := p.Retrieve(context.Background(), "q", QueryOptions{Where: where})
	if err != nil {
		t.Fatalf("Retrieve: %v", err)
	}
	if len(sections) != 0 {
		t.Errorf("got sections %+v, want none", sections)
	}
}
