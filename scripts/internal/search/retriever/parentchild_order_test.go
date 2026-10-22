package retriever

import (
	"context"
	"fmt"
	"testing"

	"github.com/plaesy/spec-kit/internal/config"
	"github.com/plaesy/spec-kit/internal/search/chunker"
	"github.com/plaesy/spec-kit/internal/search/store"
)

// Matches must keep their rank and stay at the front. Sibling chunks of a
// matched document come after all matches. Before the fix, siblings were
// interleaved with the matches, and a chunk that was both a match and a
// sibling could be listed twice.
func TestExpandParentChild_MatchesStayFirstInRankOrder(t *testing.T) {
	var bigDoc []chunker.Chunk
	for i := 0; i < 30; i++ {
		bigDoc = append(bigDoc, chunker.Chunk{ID: fmt.Sprintf("a%02d", i), DocID: "docA", Content: "x"})
	}
	smallDoc := []chunker.Chunk{
		{ID: "b00", DocID: "docB", Content: "y"},
		{ID: "b01", DocID: "docB", Content: "y"},
	}
	ds := &mockDocStoreForRetriever{docs: map[string]*store.ProcessedDocument{
		"docA": {DocID: "docA", Chunks: bigDoc},
		"docB": {DocID: "docB", Chunks: smallDoc},
	}}

	cfg := config.RetrieverConfig{TopK: 3, ParentChild: config.ParentChildConfig{Enabled: true}}
	r := NewHybridRetriever(cfg, &mockEmbedderForRetriever{dimension: 3}, &mockVectorStoreForRetriever{}, &mockTextIndexForRetriever{}, ds)

	matched := []Result{
		{ID: "a00", DocID: "docA", Score: 0.9},
		{ID: "b00", DocID: "docB", Score: 0.8},
		{ID: "a01", DocID: "docA", Score: 0.7},
	}
	out := r.expandParentChild(context.Background(), matched)

	wantFront := []string{"a00", "b00", "a01"}
	for i, id := range wantFront {
		if out[i].ID != id {
			t.Fatalf("position %d = %s, want match %s; matches must keep rank order at the front", i, out[i].ID, id)
		}
	}

	seen := map[string]int{}
	for _, res := range out {
		seen[res.ID]++
	}
	for id, n := range seen {
		if n > 1 {
			t.Errorf("chunk %s listed %d times", id, n)
		}
	}

	// Every sibling is a chunk of a matched document and is not itself a match.
	wantTotal := 3 + (30 - 2) + (2 - 1) // matches, rest of docA, rest of docB
	if len(out) != wantTotal {
		t.Errorf("got %d results, want %d", len(out), wantTotal)
	}
}
