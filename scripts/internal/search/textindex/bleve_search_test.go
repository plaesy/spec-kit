package textindex

import (
	"context"
	"testing"
)

// The real BleveIndex has no search test with hits; the existing tests use a
// mock. This one exercises the production path that the pipeline and the eval
// adapter both rely on.
func TestBleveIndex_SearchReturnsMetadata(t *testing.T) {
	ctx := context.Background()
	idx, err := NewBleveIndex(Config{Type: "bleve"}, t.TempDir())
	if err != nil {
		t.Fatalf("NewBleveIndex: %v", err)
	}
	defer idx.Close()

	err = idx.Add(ctx,
		[]string{"c1", "c2"},
		[]string{"binary search tree insertion", "kubernetes deployment manifests"},
		[]map[string]string{
			{"doc_id": "d1", "source": "markdown"},
			{"doc_id": "d2", "source": "yaml"},
		},
	)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	results, err := idx.Search(ctx, "binary", 10, nil)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].ID != "c1" {
		t.Errorf("ID = %q, want c1", results[0].ID)
	}
	if got := results[0].Metadata["doc_id"]; got != "d1" {
		t.Errorf("Metadata[doc_id] = %q, want d1 (metadata must be rebuilt from bleve's flattened fields)", got)
	}
	if got := results[0].Metadata["source"]; got != "markdown" {
		t.Errorf("Metadata[source] = %q, want markdown", got)
	}
}

func TestBleveIndex_SearchFilterByMetadata(t *testing.T) {
	ctx := context.Background()
	idx, err := NewBleveIndex(Config{Type: "bleve"}, t.TempDir())
	if err != nil {
		t.Fatalf("NewBleveIndex: %v", err)
	}
	defer idx.Close()

	if err := idx.Add(ctx,
		[]string{"c1"},
		[]string{"shared words about deployment"},
		[]map[string]string{{"doc_id": "d1", "source": "markdown"}},
	); err != nil {
		t.Fatalf("Add: %v", err)
	}

	hit, err := idx.Search(ctx, "deployment", 10, map[string]string{"source": "markdown"})
	if err != nil {
		t.Fatalf("Search with matching filter: %v", err)
	}
	if len(hit) != 1 {
		t.Errorf("matching filter returned %d results, want 1", len(hit))
	}

	miss, err := idx.Search(ctx, "deployment", 10, map[string]string{"source": "yaml"})
	if err != nil {
		t.Fatalf("Search with non-matching filter: %v", err)
	}
	if len(miss) != 0 {
		t.Errorf("non-matching filter returned %d results, want 0", len(miss))
	}
}
