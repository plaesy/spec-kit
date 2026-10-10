package vectorstore

import (
	"context"
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

		score := cosineSimilarity(queryVector, vec)
		if score >= 0 {
			results = append(results, SearchResult{
				ID:       id,
				Score:    score,
				Content:  m.documents[id],
				Metadata: m.metadatas[id],
				Vector:   vec,
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

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Type != "chromem" {
		t.Errorf("Type = %s, expected 'chromem'", cfg.Type)
	}
	if cfg.HNSW.M != 16 {
		t.Errorf("HNSW.M = %d, expected 16", cfg.HNSW.M)
	}
	if cfg.HNSW.EFConstruction != 200 {
		t.Errorf("HNSW.EFConstruction = %d, expected 200", cfg.HNSW.EFConstruction)
	}
	if cfg.HNSW.EFSearch != 100 {
		t.Errorf("HNSW.EFSearch = %d, expected 100", cfg.HNSW.EFSearch)
	}
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

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		a        []float32
		b        []float32
		expected float32
	}{
		{[]float32{1, 0, 0}, []float32{1, 0, 0}, 1.0},
		{[]float32{1, 0, 0}, []float32{0, 1, 0}, 0.0},
		{[]float32{1, 0, 0}, []float32{-1, 0, 0}, -1.0},
		{[]float32{3, 4}, []float32{3, 4}, 1.0},
		{[]float32{1, 2, 3}, []float32{4, 5, 6}, 0.974}, // approximate
	}

	for _, tt := range tests {
		result := cosineSimilarity(tt.a, tt.b)
		// Allow small floating point difference
		if result < tt.expected-0.01 || result > tt.expected+0.01 {
			t.Errorf("cosineSimilarity(%v, %v) = %f, expected ~%f", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestCosineSimilarity_DifferentLengths(t *testing.T) {
	result := cosineSimilarity([]float32{1, 0}, []float32{1, 0, 0})
	if result != -1 {
		t.Errorf("Expected -1 for different lengths, got %f", result)
	}
}

func TestCosineSimilarity_ZeroVectors(t *testing.T) {
	result := cosineSimilarity([]float32{0, 0}, []float32{1, 1})
	if result != 0 {
		t.Errorf("Expected 0 for zero vector, got %f", result)
	}

	result = cosineSimilarity([]float32{1, 1}, []float32{0, 0})
	if result != 0 {
		t.Errorf("Expected 0 for zero vector, got %f", result)
	}
}

func TestSearchResult_Fields(t *testing.T) {
	result := SearchResult{
		ID:       "test-id",
		Score:    0.95,
		Content:  "test content",
		Metadata: map[string]string{"key": "value"},
		Vector:   []float32{0.1, 0.2, 0.3},
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
	if len(result.Vector) != 3 {
		t.Errorf("Vector length = %d", len(result.Vector))
	}
}

func TestConfig_HNSW(t *testing.T) {
	cfg := Config{
		Type: "chromem",
		HNSW: HNSWConfig{
			M:              32,
			EFConstruction: 400,
			EFSearch:       200,
		},
	}

	if cfg.HNSW.M != 32 {
		t.Errorf("HNSW.M = %d", cfg.HNSW.M)
	}
	if cfg.HNSW.EFConstruction != 400 {
		t.Errorf("HNSW.EFConstruction = %d", cfg.HNSW.EFConstruction)
	}
	if cfg.HNSW.EFSearch != 200 {
		t.Errorf("HNSW.EFSearch = %d", cfg.HNSW.EFSearch)
	}
}

func TestVectorStoreInterface_Implementation(t *testing.T) {
	var _ VectorStore = newMockVectorStore()
	t.Log("Mock vector store implements VectorStore interface")
}
