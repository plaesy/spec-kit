package vectorstore

import (
	"context"
	"testing"
)

func addNearAndFar(t *testing.T, ctx context.Context, store *ChromemStore) {
	t.Helper()
	err := store.Add(ctx,
		[]string{"near", "far"},
		[][]float32{{1, 0, 0}, {0, 1, 0}},
		[]map[string]string{{"doc_id": "d-near"}, {"doc_id": "d-far"}},
		[]string{"near doc", "far doc"},
	)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
}

func TestChromemStore_SearchOnPopulatedStore(t *testing.T) {
	ctx := context.Background()
	store, err := NewChromemStore(Config{Type: "chromem"}, t.TempDir())
	if err != nil {
		t.Fatalf("NewChromemStore: %v", err)
	}
	defer store.Close()
	addNearAndFar(t, ctx, store)

	results, err := store.Search(ctx, []float32{1, 0, 0}, 10, nil)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if results[0].ID != "near" {
		t.Errorf("top result = %q, want near (cosine order)", results[0].ID)
	}
	if got := results[0].Metadata["doc_id"]; got != "d-near" {
		t.Errorf("Metadata[doc_id] = %q, want d-near", got)
	}
}

// The CLI opens the index in a new process for each query. A search must work
// on a store written by an earlier process, using the vectors chromem persisted.
func TestChromemStore_SearchAfterReopen(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	writer, err := NewChromemStore(Config{Type: "chromem"}, dir)
	if err != nil {
		t.Fatalf("NewChromemStore (writer): %v", err)
	}
	addNearAndFar(t, ctx, writer)
	writer.Close()

	reader, err := NewChromemStore(Config{Type: "chromem"}, dir)
	if err != nil {
		t.Fatalf("NewChromemStore (reader): %v", err)
	}
	defer reader.Close()

	results, err := reader.Search(ctx, []float32{1, 0, 0}, 10, nil)
	if err != nil {
		t.Fatalf("Search after reopen: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("reopened store returned no vector results")
	}
	if results[0].ID != "near" {
		t.Errorf("top result = %q, want near", results[0].ID)
	}
	if results[0].Score < 0.99 {
		t.Errorf("score = %.3f, want ~1.0 for an identical vector", results[0].Score)
	}
}

func TestChromemStore_DeleteRemovesDocuments(t *testing.T) {
	ctx := context.Background()
	store, err := NewChromemStore(Config{Type: "chromem"}, t.TempDir())
	if err != nil {
		t.Fatalf("NewChromemStore: %v", err)
	}
	defer store.Close()
	addNearAndFar(t, ctx, store)

	if err := store.Delete(ctx, []string{"near"}); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	results, err := store.Search(ctx, []float32{1, 0, 0}, 10, nil)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	for _, r := range results {
		if r.ID == "near" {
			t.Fatal("deleted document is still returned by Search")
		}
	}
	if n, _ := store.Count(); n != 1 {
		t.Errorf("Count = %d after delete, want 1", n)
	}
}

func TestChromemStore_AddRejectsMissingVectors(t *testing.T) {
	store, err := NewChromemStore(Config{Type: "chromem"}, t.TempDir())
	if err != nil {
		t.Fatalf("NewChromemStore: %v", err)
	}
	defer store.Close()

	err = store.Add(context.Background(), []string{"a", "b"}, [][]float32{{1, 0}}, nil, []string{"x", "y"})
	if err == nil {
		t.Fatal("expected an error when vectors do not match ids")
	}
}
