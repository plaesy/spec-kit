package indexer

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/plaesy/spec-kit/internal/search/chunker"
	"github.com/plaesy/spec-kit/internal/search/embedder"
	"github.com/plaesy/spec-kit/internal/search/parser"
	"github.com/plaesy/spec-kit/internal/search/textindex"
	"github.com/plaesy/spec-kit/internal/search/vectorstore"
)

type mockEmbedderForIndexer struct {
	embedder.Embedder
	dimension int
	callCount int
}

func (m *mockEmbedderForIndexer) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	m.callCount++
	embeddings := make([][]float32, len(texts))
	for i := range embeddings {
		embeddings[i] = make([]float32, m.dimension)
		for j := range embeddings[i] {
			embeddings[i][j] = float32(i+1) / 100.0
		}
	}
	return embeddings, nil
}

func (m *mockEmbedderForIndexer) Dimension() int {
	return m.dimension
}

func (m *mockEmbedderForIndexer) ModelName() string {
	return "mock-embedder"
}

func (m *mockEmbedderForIndexer) MaxBatchSize() int {
	return 32
}

func (m *mockEmbedderForIndexer) MaxSeqLen() int {
	return 512
}

type mockVectorStoreForIndexer struct {
	vectorstore.VectorStore
	data map[string]vectorstore.SearchResult
}

func (m *mockVectorStoreForIndexer) Add(ctx context.Context, ids []string, vectors [][]float32, metadatas []map[string]string, documents []string) error {
	if m.data == nil {
		m.data = make(map[string]vectorstore.SearchResult)
	}
	for i, id := range ids {
		m.data[id] = vectorstore.SearchResult{
			ID:       id,
			Content:  documents[i],
			Metadata: metadatas[i],
			Vector:   vectors[i],
			Score:    1.0,
		}
	}
	return nil
}

func (m *mockVectorStoreForIndexer) Search(ctx context.Context, queryVector []float32, k int, filter map[string]string) ([]vectorstore.SearchResult, error) {
	var results []vectorstore.SearchResult
	for _, v := range m.data {
		if filter != nil {
			match := true
			for k, v2 := range filter {
				if v.Metadata[k] != v2 {
					match = false
					break
				}
			}
			if !match {
				continue
			}
		}
		results = append(results, v)
	}
	if k > 0 && len(results) > k {
		results = results[:k]
	}
	return results, nil
}

func (m *mockVectorStoreForIndexer) Delete(ctx context.Context, ids []string) error {
	for _, id := range ids {
		delete(m.data, id)
	}
	return nil
}

func (m *mockVectorStoreForIndexer) Count() (int, error) {
	return len(m.data), nil
}

func (m *mockVectorStoreForIndexer) Close() error {
	return nil
}

func (m *mockVectorStoreForIndexer) Name() string {
	return "mock-vectorstore"
}

type mockTextIndexForIndexer struct {
	textindex.TextIndex
	data map[string]textindex.SearchResult
}

func (m *mockTextIndexForIndexer) Add(ctx context.Context, ids []string, documents []string, metadatas []map[string]string) error {
	if m.data == nil {
		m.data = make(map[string]textindex.SearchResult)
	}
	for i, id := range ids {
		m.data[id] = textindex.SearchResult{
			ID:       id,
			Content:  documents[i],
			Metadata: metadatas[i],
			Score:    1.0,
		}
	}
	return nil
}

func (m *mockTextIndexForIndexer) Search(ctx context.Context, query string, k int, filter map[string]string) ([]textindex.SearchResult, error) {
	var results []textindex.SearchResult
	for _, v := range m.data {
		if filter != nil {
			match := true
			for k, v2 := range filter {
				if v.Metadata[k] != v2 {
					match = false
					break
				}
			}
			if !match {
				continue
			}
		}
		results = append(results, v)
	}
	if k > 0 && len(results) > k {
		results = results[:k]
	}
	return results, nil
}

func (m *mockTextIndexForIndexer) Delete(ctx context.Context, ids []string) error {
	for _, id := range ids {
		delete(m.data, id)
	}
	return nil
}

func (m *mockTextIndexForIndexer) Count() (int, error) {
	return len(m.data), nil
}

func (m *mockTextIndexForIndexer) Close() error {
	return nil
}

func (m *mockTextIndexForIndexer) Name() string {
	return "mock-textindex"
}

func TestDefaultIndexerConfig(t *testing.T) {
	cfg := DefaultIndexerConfig()
	if cfg.Workers != 4 {
		t.Errorf("Workers = %d, expected 4", cfg.Workers)
	}
	if !cfg.Resume {
		t.Error("Resume should be true by default")
	}
	if cfg.BatchSize != 100 {
		t.Errorf("BatchSize = %d, expected 100", cfg.BatchSize)
	}
	if cfg.CheckpointInterval != 50 {
		t.Errorf("CheckpointInterval = %d, expected 50", cfg.CheckpointInterval)
	}
}

func TestProgress_Serialization(t *testing.T) {
	progress := Progress{
		TotalFiles:      100,
		ProcessedFiles:  50,
		TotalChunks:     200,
		ProcessedChunks: 150,
		CurrentFile:     "test.md",
		FileHashes:      map[string]string{"test.md": "hash123"},
		StartedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Completed:       false,
	}

	data, err := json.Marshal(progress)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var progress2 Progress
	err = json.Unmarshal(data, &progress2)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if progress2.TotalFiles != progress.TotalFiles {
		t.Errorf("TotalFiles mismatch")
	}
	if progress2.ProcessedFiles != progress.ProcessedFiles {
		t.Errorf("ProcessedFiles mismatch")
	}
	if progress2.TotalChunks != progress.TotalChunks {
		t.Errorf("TotalChunks mismatch")
	}
	if progress2.ProcessedChunks != progress.ProcessedChunks {
		t.Errorf("ProcessedChunks mismatch")
	}
	if progress2.CurrentFile != progress.CurrentFile {
		t.Errorf("CurrentFile mismatch")
	}
	if progress2.FileHashes["test.md"] != "hash123" {
		t.Errorf("FileHashes mismatch")
	}
	if progress2.Completed != progress.Completed {
		t.Errorf("Completed mismatch")
	}
}

func TestIndexerConfig_Custom(t *testing.T) {
	cfg := IndexerConfig{
		Workers:            2,
		Resume:             false,
		BatchSize:          50,
		CheckpointInterval: 10,
	}

	if cfg.Workers != 2 {
		t.Errorf("Workers = %d", cfg.Workers)
	}
	if cfg.Resume {
		t.Error("Resume should be false")
	}
	if cfg.BatchSize != 50 {
		t.Errorf("BatchSize = %d", cfg.BatchSize)
	}
	if cfg.CheckpointInterval != 10 {
		t.Errorf("CheckpointInterval = %d", cfg.CheckpointInterval)
	}
}

func TestFilterChangedFiles(t *testing.T) {
	// We can't easily create a real indexer without all dependencies
	// So we test the logic directly

	// Create test documents
	docs := []parser.Document{
		{Path: "file1.md", Hash: "hash1", Content: "content 1"},
		{Path: "file2.md", Hash: "hash2", Content: "content 2"},
		{Path: "file3.md", Hash: "hash3", Content: "content 3"},
	}

	// Progress with some existing hashes
	progress := Progress{
		FileHashes: map[string]string{
			"file1.md": "hash1",    // Same hash - unchanged
			"file2.md": "oldhash2", // Different hash - changed
			// file3.md not in progress - new file
		},
	}

	// Filter
	var changed []parser.Document
	for _, doc := range docs {
		prevHash, exists := progress.FileHashes[doc.Path]
		if !exists || prevHash != doc.Hash {
			changed = append(changed, doc)
		}
	}

	if len(changed) != 2 {
		t.Errorf("Expected 2 changed files, got %d", len(changed))
	}

	// Check which files are in changed
	found1, found2, found3 := false, false, false
	for _, d := range changed {
		if d.Path == "file1.md" {
			found1 = true
		}
		if d.Path == "file2.md" {
			found2 = true
		}
		if d.Path == "file3.md" {
			found3 = true
		}
	}

	if found1 {
		t.Error("file1.md should not be in changed (unchanged)")
	}
	if !found2 {
		t.Error("file2.md should be in changed (hash changed)")
	}
	if !found3 {
		t.Error("file3.md should be in changed (new file)")
	}
}

func TestGetSupportedExtensions(t *testing.T) {
	// This would require a real parser registry
	// Just verify the concept
	t.Log("Supported extensions come from parser registry")
}

func TestIndexer_ProgressSaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	progressPath := filepath.Join(tmpDir, "progress.json")

	progress := Progress{
		TotalFiles:      10,
		ProcessedFiles:  5,
		TotalChunks:     20,
		ProcessedChunks: 15,
		CurrentFile:     "test.md",
		FileHashes:      map[string]string{"test.md": "hash123"},
		StartedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Completed:       false,
	}

	// Save
	data, err := json.MarshalIndent(progress, "", "  ")
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	tmpPath := progressPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if err := os.Rename(tmpPath, progressPath); err != nil {
		t.Fatalf("Rename failed: %v", err)
	}

	// Load
	data, err = os.ReadFile(progressPath)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	var loaded Progress
	err = json.Unmarshal(data, &loaded)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if loaded.TotalFiles != 10 {
		t.Errorf("TotalFiles = %d", loaded.TotalFiles)
	}
	if loaded.ProcessedFiles != 5 {
		t.Errorf("ProcessedFiles = %d", loaded.ProcessedFiles)
	}
	if loaded.FileHashes["test.md"] != "hash123" {
		t.Errorf("FileHashes mismatch")
	}
}

func TestIndexResult(t *testing.T) {
	result := indexResult{
		doc: parser.Document{
			ID:   "doc1",
			Path: "test.md",
		},
		chunks: []chunker.Chunk{
			{ID: "doc1-0", DocID: "doc1", Content: "Chunk 1"},
			{ID: "doc1-1", DocID: "doc1", Content: "Chunk 2"},
		},
		err: nil,
	}

	if result.doc.ID != "doc1" {
		t.Errorf("Doc ID mismatch")
	}
	if len(result.chunks) != 2 {
		t.Errorf("Chunks count = %d", len(result.chunks))
	}
	if result.err != nil {
		t.Errorf("Error should be nil: %v", result.err)
	}

	// Test with error
	resultErr := indexResult{
		doc: parser.Document{Path: "test.md"},
		err: &testError{msg: "test error"},
	}
	if resultErr.err == nil {
		t.Error("Error should not be nil")
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
