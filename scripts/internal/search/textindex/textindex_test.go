package textindex

import (
	"context"
	"testing"
)

type mockTextIndex struct {
	documents map[string]string
	metadatas map[string]map[string]string
	closed    bool
}

func newMockTextIndex() *mockTextIndex {
	return &mockTextIndex{
		documents: make(map[string]string),
		metadatas: make(map[string]map[string]string),
	}
}

func (m *mockTextIndex) Add(ctx context.Context, ids []string, documents []string, metadatas []map[string]string) error {
	if m.closed {
		return ErrClosed
	}
	if len(ids) != len(documents) {
		return ErrLengthMismatch
	}
	for i, id := range ids {
		m.documents[id] = documents[i]
		if metadatas != nil && i < len(metadatas) {
			m.metadatas[id] = metadatas[i]
		}
	}
	return nil
}

func (m *mockTextIndex) Search(ctx context.Context, query string, k int, filter map[string]string) ([]SearchResult, error) {
	if m.closed {
		return nil, ErrClosed
	}

	var results []SearchResult
	queryLower := toLower(query)

	for id, doc := range m.documents {
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

		// Simple BM25-like scoring: count query term occurrences
		docLower := toLower(doc)
		score := float64(countOccurrences(docLower, queryLower))
		if score > 0 {
			results = append(results, SearchResult{
				ID:       id,
				Score:    score,
				Content:  doc,
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

func (m *mockTextIndex) Delete(ctx context.Context, ids []string) error {
	if m.closed {
		return ErrClosed
	}
	for _, id := range ids {
		delete(m.documents, id)
		delete(m.metadatas, id)
	}
	return nil
}

func (m *mockTextIndex) Count() (int, error) {
	if m.closed {
		return 0, ErrClosed
	}
	return len(m.documents), nil
}

func (m *mockTextIndex) Close() error {
	m.closed = true
	return nil
}

func (m *mockTextIndex) Name() string {
	return "mock"
}

var (
	ErrClosed         = &TextIndexError{Code: "CLOSED", Message: "index is closed"}
	ErrLengthMismatch = &TextIndexError{Code: "LENGTH_MISMATCH", Message: "ids and documents length mismatch"}
)

type TextIndexError struct {
	Code    string
	Message string
}

func (e *TextIndexError) Error() string {
	return e.Message
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

func countOccurrences(text, substr string) int {
	if substr == "" {
		return 0
	}
	count := 0
	for i := 0; i <= len(text)-len(substr); i++ {
		if text[i:i+len(substr)] == substr {
			count++
		}
	}
	return count
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Type != "bleve" {
		t.Errorf("Type = %s, expected 'bleve'", cfg.Type)
	}
}

func TestMockTextIndex_Add(t *testing.T) {
	index := newMockTextIndex()
	ctx := context.Background()

	ids := []string{"1", "2", "3"}
	documents := []string{
		"the quick brown fox jumps over the lazy dog",
		"hello world hello",
		"golang is great for programming",
	}
	metadatas := []map[string]string{
		{"source": "test", "category": "animals"},
		{"source": "test", "category": "greeting"},
		{"source": "test", "category": "programming"},
	}

	err := index.Add(ctx, ids, documents, metadatas)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	count, err := index.Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 3 {
		t.Errorf("Count = %d, expected 3", count)
	}
}

func TestMockTextIndex_Add_LengthMismatch(t *testing.T) {
	index := newMockTextIndex()
	ctx := context.Background()

	ids := []string{"1", "2"}
	documents := []string{"doc 1"} // Missing one

	err := index.Add(ctx, ids, documents, nil)
	if err == nil {
		t.Error("Expected error for length mismatch")
	}
	if err != ErrLengthMismatch {
		t.Errorf("Expected ErrLengthMismatch, got: %v", err)
	}
}

func TestMockTextIndex_Search(t *testing.T) {
	index := newMockTextIndex()
	ctx := context.Background()

	ids := []string{"1", "2", "3"}
	documents := []string{
		"the quick brown fox jumps over the lazy dog",
		"hello world hello",
		"golang is great for programming",
	}
	metadatas := []map[string]string{
		{"category": "animals"},
		{"category": "greeting"},
		{"category": "programming"},
	}

	err := index.Add(ctx, ids, documents, metadatas)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Search for "hello" - should match document 2
	results, err := index.Search(ctx, "hello", 10, nil)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'hello', got %d", len(results))
	}
	if results[0].ID != "2" {
		t.Errorf("Result ID = %s, expected '2'", results[0].ID)
	}
	if results[0].Score <= 0 {
		t.Errorf("Score should be positive: %f", results[0].Score)
	}
}

func TestMockTextIndex_Search_MultipleMatches(t *testing.T) {
	index := newMockTextIndex()
	ctx := context.Background()

	ids := []string{"1", "2", "3"}
	documents := []string{
		"the quick brown fox",
		"quick brown fox jumps",
		"lazy dog sleeps",
	}

	err := index.Add(ctx, ids, documents, nil)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Search for "quick" - should match documents 1 and 2
	results, err := index.Search(ctx, "quick", 10, nil)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results for 'quick', got %d", len(results))
	}
}

func TestMockTextIndex_Search_WithFilter(t *testing.T) {
	index := newMockTextIndex()
	ctx := context.Background()

	ids := []string{"1", "2", "3"}
	documents := []string{
		"fox jumps",
		"dog runs",
		"cat sleeps",
	}
	metadatas := []map[string]string{
		{"animal": "fox"},
		{"animal": "dog"},
		{"animal": "cat"},
	}

	err := index.Add(ctx, ids, documents, metadatas)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Search for "jumps" with filter animal=fox
	results, err := index.Search(ctx, "jumps", 10, map[string]string{"animal": "fox"})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 filtered result, got %d", len(results))
	}
	if results[0].ID != "1" {
		t.Errorf("Filtered result ID = %s, expected '1'", results[0].ID)
	}
}

func TestMockTextIndex_Search_NoMatches(t *testing.T) {
	index := newMockTextIndex()
	ctx := context.Background()

	ids := []string{"1"}
	documents := []string{"hello world"}

	err := index.Add(ctx, ids, documents, nil)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	results, err := index.Search(ctx, "nonexistent", 10, nil)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results for non-existent term, got %d", len(results))
	}
}

func TestMockTextIndex_Delete(t *testing.T) {
	index := newMockTextIndex()
	ctx := context.Background()

	ids := []string{"1", "2", "3"}
	documents := []string{"doc 1", "doc 2", "doc 3"}

	err := index.Add(ctx, ids, documents, nil)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	count, _ := index.Count()
	if count != 3 {
		t.Errorf("Initial count = %d, expected 3", count)
	}

	err = index.Delete(ctx, []string{"2"})
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	count, _ = index.Count()
	if count != 2 {
		t.Errorf("Count after delete = %d, expected 2", count)
	}

	// Verify deleted item is gone
	results, _ := index.Search(ctx, "doc", 10, nil)
	for _, r := range results {
		if r.ID == "2" {
			t.Error("Deleted item still in results")
		}
	}
}

func TestMockTextIndex_Close(t *testing.T) {
	index := newMockTextIndex()
	ctx := context.Background()

	err := index.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Operations after close should fail
	err = index.Add(ctx, []string{"1"}, []string{"doc"}, nil)
	if err != ErrClosed {
		t.Errorf("Expected ErrClosed after close, got: %v", err)
	}

	_, err = index.Search(ctx, "test", 10, nil)
	if err != ErrClosed {
		t.Errorf("Expected ErrClosed after close, got: %v", err)
	}

	_, err = index.Count()
	if err != ErrClosed {
		t.Errorf("Expected ErrClosed after close, got: %v", err)
	}
}

func TestSearchResult_Fields(t *testing.T) {
	result := SearchResult{
		ID:       "test-id",
		Score:    1.5,
		Content:  "test content",
		Metadata: map[string]string{"key": "value"},
	}

	if result.ID != "test-id" {
		t.Errorf("ID = %s", result.ID)
	}
	if result.Score != 1.5 {
		t.Errorf("Score = %f", result.Score)
	}
	if result.Content != "test content" {
		t.Errorf("Content = %s", result.Content)
	}
	if result.Metadata["key"] != "value" {
		t.Errorf("Metadata = %v", result.Metadata)
	}
}

func TestConfig(t *testing.T) {
	cfg := Config{Type: "sqlite"}
	if cfg.Type != "sqlite" {
		t.Errorf("Type = %s", cfg.Type)
	}
}

func TestTextIndexInterface_Implementation(t *testing.T) {
	var _ TextIndex = newMockTextIndex()
	t.Log("Mock text index implements TextIndex interface")
}
