package vectorstore

import (
	"context"
	"math"
	"testing"
)

type mockVectorStore struct {
	vectors   map[string][]float32
	metadatas map[string]map[string]string
	documents map[string]string
	closed    bool
}

func newMockVectorStore() *mockVectorStore {
	return &mockVectorStore{
		vectors:   make(map[string][]float32),
		metadatas: make(map[string]map[string]string),
		documents: make(map[string]string),
	}
}

func (m *mockVectorStore) Add(ctx context.Context, ids []string, vectors [][]float32, metadatas []map[string]string, documents []string) error {
	if m.closed {
		return ErrClosed
	}
	if len(ids) != len(documents) {
		return ErrLengthMismatch
	}
	for i, id := range ids {
		if vectors != nil && i < len(vectors) {
			m.vectors[id] = vectors[i]
		} else {
			m.vectors[id] = make([]float32, 768)
		}
		if metadatas != nil && i < len(metadatas) {
			m.metadatas[id] = metadatas[i]
		}
		if i < len(documents) {
			m.documents[id] = documents[i]
		}
	}
	return nil
}

func (m *mockVectorStore) Search(ctx context.Context, queryVector []float32, k int, filter map[string]string) ([]SearchResult, error) {
	if m.closed {
		return nil, ErrClosed
	}

	var results []SearchResult
	for id, vec := range m.vectors {
		if filter != nil {
			meta := m.metadatas[id]
			match := true
			for k, v := range filter {
				if meta[k] != v {
					match = false
					break
				}
			}
			if !match {
				continue
			}
		}

		score := testCosine(queryVector, vec)
		if score >= 0 {
			results = append(results, SearchResult{
				ID:       id,
				Score:    score,
				Content:  m.documents[id],
				Metadata: m.metadatas[id],
			})
		}
	}

	// Sort by score descending
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Score < results[j].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	if k > 0 && len(results) > k {
		results = results[:k]
	}

	return results, nil
}

func (m *mockVectorStore) Delete(ctx context.Context, ids []string) error {
	if m.closed {
		return ErrClosed
	}
	for _, id := range ids {
		delete(m.vectors, id)
		delete(m.metadatas, id)
		delete(m.documents, id)
	}
	return nil
}

func (m *mockVectorStore) Count() (int, error) {
	if m.closed {
		return 0, ErrClosed
	}
	return len(m.vectors), nil
}

func (m *mockVectorStore) Close() error {
	m.closed = true
	return nil
}

func (m *mockVectorStore) Name() string {
	return "mock"
}

var (
	ErrClosed         = &VectorStoreError{Code: "CLOSED", Message: "store is closed"}
	ErrLengthMismatch = &VectorStoreError{Code: "LENGTH_MISMATCH", Message: "ids and documents length mismatch"}
)

type VectorStoreError struct {
	Code    string
	Message string
}

func (e *VectorStoreError) Error() string {
	return e.Message
}

func TestMockVectorStore_Add(t *testing.T) {
	store := newMockVectorStore()
	ctx := context.Background()

	ids := []string{"1", "2", "3"}
	vectors := [][]float32{
		{1, 0, 0},
		{0, 1, 0},
		{0, 0, 1},
	}
	metadatas := []map[string]string{
		{"source": "test", "category": "a"},
		{"source": "test", "category": "b"},
		{"source": "test", "category": "a"},
	}
	documents := []string{"doc 1", "doc 2", "doc 3"}

	err := store.Add(ctx, ids, vectors, metadatas, documents)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	count, err := store.Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 3 {
		t.Errorf("Count = %d, expected 3", count)
	}
}

func TestMockVectorStore_Add_LengthMismatch(t *testing.T) {
	store := newMockVectorStore()
	ctx := context.Background()

	ids := []string{"1", "2"}
	vectors := [][]float32{{1, 0}, {0, 1}}
	documents := []string{"doc 1"} // Missing one

	err := store.Add(ctx, ids, vectors, nil, documents)
	if err == nil {
		t.Error("Expected error for length mismatch")
	}
	if err != ErrLengthMismatch {
		t.Errorf("Expected ErrLengthMismatch, got: %v", err)
	}
}

func TestMockVectorStore_Search(t *testing.T) {
	store := newMockVectorStore()
	ctx := context.Background()

	// Add orthogonal vectors
	ids := []string{"1", "2", "3"}
	vectors := [][]float32{
		{1, 0, 0},
		{0, 1, 0},
		{0, 0, 1},
	}
	documents := []string{"doc 1", "doc 2", "doc 3"}

	err := store.Add(ctx, ids, vectors, nil, documents)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Search with query matching first vector
	queryVector := []float32{1, 0, 0}
	results, err := store.Search(ctx, queryVector, 2, nil)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// First result should be "1" with highest score
	if results[0].ID != "1" {
		t.Errorf("First result ID = %s, expected '1'", results[0].ID)
	}
	if results[0].Score <= 0 {
		t.Errorf("First result score should be positive: %f", results[0].Score)
	}
	if results[1].ID != "2" && results[1].ID != "3" {
		t.Errorf("Second result ID = %s, expected '2' or '3'", results[1].ID)
	}
}

func TestMockVectorStore_Search_WithFilter(t *testing.T) {
	store := newMockVectorStore()
	ctx := context.Background()

	ids := []string{"1", "2", "3"}
	vectors := [][]float32{
		{1, 0, 0},
		{0, 1, 0},
		{0, 0, 1},
	}
	metadatas := []map[string]string{
		{"category": "a"},
		{"category": "b"},
		{"category": "a"},
	}
	documents := []string{"doc 1", "doc 2", "doc 3"}

	err := store.Add(ctx, ids, vectors, metadatas, documents)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Filter for category "a"
	queryVector := []float32{1, 0, 0}
	results, err := store.Search(ctx, queryVector, 10, map[string]string{"category": "a"})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Should only return IDs 1 and 3
	if len(results) != 2 {
		t.Errorf("Expected 2 filtered results, got %d", len(results))
	}
	for _, r := range results {
		if r.ID != "1" && r.ID != "3" {
			t.Errorf("Filtered result ID = %s, expected '1' or '3'", r.ID)
		}
	}
}

func TestMockVectorStore_Delete(t *testing.T) {
	store := newMockVectorStore()
	ctx := context.Background()

	ids := []string{"1", "2", "3"}
	vectors := [][]float32{{1}, {1}, {1}}
	documents := []string{"doc 1", "doc 2", "doc 3"}

	err := store.Add(ctx, ids, vectors, nil, documents)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	count, _ := store.Count()
	if count != 3 {
		t.Errorf("Initial count = %d, expected 3", count)
	}

	err = store.Delete(ctx, []string{"2"})
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	count, _ = store.Count()
	if count != 2 {
		t.Errorf("Count after delete = %d, expected 2", count)
	}

	// Verify deleted item is gone
	results, _ := store.Search(ctx, []float32{1}, 10, nil)
	for _, r := range results {
		if r.ID == "2" {
			t.Error("Deleted item still in results")
		}
	}
}

func TestMockVectorStore_Close(t *testing.T) {
	store := newMockVectorStore()
	ctx := context.Background()

	err := store.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Operations after close should fail
	err = store.Add(ctx, []string{"1"}, [][]float32{{1}}, nil, []string{"doc"})
	if err != ErrClosed {
		t.Errorf("Expected ErrClosed after close, got: %v", err)
	}

	_, err = store.Search(ctx, []float32{1}, 10, nil)
	if err != ErrClosed {
		t.Errorf("Expected ErrClosed after close, got: %v", err)
	}

	_, err = store.Count()
	if err != ErrClosed {
		t.Errorf("Expected ErrClosed after close, got: %v", err)
	}
}

// testCosine is the similarity the mock store ranks by. Production search now
// uses the similarity chromem computes, so this helper exists only for the mock.
func testCosine(a, b []float32) float32 {
	if len(a) != len(b) {
		return -1
	}
	var dot, normA, normB float32
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

func TestSearchResult_Fields(t *testing.T) {
	result := SearchResult{
		ID:       "test-id",
		Score:    0.95,
		Content:  "test content",
		Metadata: map[string]string{"key": "value"},
	}

	if result.ID != "test-id" {
		t.Errorf("ID = %s", result.ID)
	}
	if result.Score != 0.95 {
		t.Errorf("Score = %f", result.Score)
	}
	if result.Content != "test content" {
		t.Errorf("Content = %s", result.Content)
	}
	if result.Metadata["key"] != "value" {
		t.Errorf("Metadata = %v", result.Metadata)
	}
}

func TestVectorStoreInterface_Implementation(t *testing.T) {
	var _ VectorStore = newMockVectorStore()
	t.Log("Mock vector store implements VectorStore interface")
}
