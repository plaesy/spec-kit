package store

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/plaesy/spec-kit/internal/search/chunker"
	"github.com/plaesy/spec-kit/internal/search/parser"
)

func TestNewDocumentStore(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewDocumentStore(tmpDir)
	if err != nil {
		t.Fatalf("NewDocumentStore failed: %v", err)
	}
	if store == nil {
		t.Fatal("Store is nil")
	}

	// Check directories were created
	rawPath := filepath.Join(tmpDir, "store", "raw")
	processedPath := filepath.Join(tmpDir, "store", "processed")

	if _, err := os.Stat(rawPath); os.IsNotExist(err) {
		t.Error("Raw directory not created")
	}
	if _, err := os.Stat(processedPath); os.IsNotExist(err) {
		t.Error("Processed directory not created")
	}
}

func TestStoreRaw(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewDocumentStore(tmpDir)
	if err != nil {
		t.Fatalf("NewDocumentStore failed: %v", err)
	}

	ctx := context.Background()
	doc := parser.Document{
		ID:       "test-doc-1",
		Path:     "/tmp/test.md",
		Content:  "# Test\n\nContent here",
		Metadata: map[string]string{"source": "markdown"},
		ModTime:  time.Now(),
		Hash:     "abc123",
	}

	// Create the actual file
	testFile := filepath.Join(tmpDir, "test.md")
	err = os.WriteFile(testFile, []byte(doc.Content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	doc.Path = testFile

	rawDoc, err := store.StoreRaw(ctx, doc)
	if err != nil {
		t.Fatalf("StoreRaw failed: %v", err)
	}

	// ID should be first 16 chars of the computed hash
	expectedID := rawDoc.Hash[:16]
	if rawDoc.ID != expectedID {
		t.Errorf("RawDoc ID = %s, expected %s", rawDoc.ID, expectedID)
	}
	if rawDoc.Path != testFile {
		t.Errorf("RawDoc Path = %s, expected %s", rawDoc.Path, testFile)
	}
	if rawDoc.Hash == "" {
		t.Errorf("RawDoc Hash should not be empty")
	}
	if rawDoc.Size == 0 {
		t.Error("RawDoc Size should not be zero")
	}
	if rawDoc.Metadata["source"] != "markdown" {
		t.Errorf("RawDoc Metadata source = %s", rawDoc.Metadata["source"])
	}
	if rawDoc.StoredAt.IsZero() {
		t.Error("RawDoc StoredAt should not be zero")
	}
}

func TestGetRaw(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewDocumentStore(tmpDir)
	if err != nil {
		t.Fatalf("NewDocumentStore failed: %v", err)
	}

	ctx := context.Background()
	doc := parser.Document{
		ID:       "test-doc-1",
		Path:     "/tmp/test.md",
		Content:  "# Test\n\nContent here",
		Metadata: map[string]string{"source": "markdown"},
		ModTime:  time.Now(),
		Hash:     "abc123def456",
	}

	testFile := filepath.Join(tmpDir, "test.md")
	err = os.WriteFile(testFile, []byte(doc.Content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	doc.Path = testFile

	rawDoc, err := store.StoreRaw(ctx, doc)
	if err != nil {
		t.Fatalf("StoreRaw failed: %v", err)
	}

	// Retrieve
	retrieved, err := store.GetRaw(ctx, rawDoc.ID)
	if err != nil {
		t.Fatalf("GetRaw failed: %v", err)
	}

	if retrieved.ID != rawDoc.ID {
		t.Errorf("Retrieved ID = %s, expected %s", retrieved.ID, rawDoc.ID)
	}
	if retrieved.Path != rawDoc.Path {
		t.Errorf("Retrieved Path = %s, expected %s", retrieved.Path, rawDoc.Path)
	}
	if retrieved.Hash != rawDoc.Hash {
		t.Errorf("Retrieved Hash = %s, expected %s", retrieved.Hash, rawDoc.Hash)
	}
	if retrieved.Size != rawDoc.Size {
		t.Errorf("Retrieved Size = %d, expected %d", retrieved.Size, rawDoc.Size)
	}
	if retrieved.Metadata["source"] != "markdown" {
		t.Errorf("Retrieved Metadata source = %s", retrieved.Metadata["source"])
	}
}

func TestGetRaw_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewDocumentStore(tmpDir)
	if err != nil {
		t.Fatalf("NewDocumentStore failed: %v", err)
	}

	ctx := context.Background()
	_, err = store.GetRaw(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent document")
	}
}

func TestGetRawContent(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewDocumentStore(tmpDir)
	if err != nil {
		t.Fatalf("NewDocumentStore failed: %v", err)
	}

	ctx := context.Background()
	doc := parser.Document{
		ID:       "test-doc-1",
		Path:     "/tmp/test.md",
		Content:  "# Test\n\nContent here",
		Metadata: map[string]string{"source": "markdown"},
		ModTime:  time.Now(),
		Hash:     "abc123",
	}

	testFile := filepath.Join(tmpDir, "test.md")
	err = os.WriteFile(testFile, []byte(doc.Content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	doc.Path = testFile

	rawDoc, err := store.StoreRaw(ctx, doc)
	if err != nil {
		t.Fatalf("StoreRaw failed: %v", err)
	}

	content, err := store.GetRawContent(ctx, rawDoc.ID)
	if err != nil {
		t.Fatalf("GetRawContent failed: %v", err)
	}

	if string(content) != doc.Content {
		t.Errorf("Retrieved content mismatch: got %q, expected %q", string(content), doc.Content)
	}
}

func TestStoreProcessed(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewDocumentStore(tmpDir)
	if err != nil {
		t.Fatalf("NewDocumentStore failed: %v", err)
	}

	ctx := context.Background()
	chunks := []chunker.Chunk{
		{
			ID:          "doc1-0",
			DocID:       "doc1",
			Content:     "Chunk 1 content",
			TokenCount:  100,
			SectionPath: "Section 1",
			Metadata:    map[string]string{"source": "markdown"},
		},
		{
			ID:          "doc1-1",
			DocID:       "doc1",
			ParentID:    "doc1-0",
			Content:     "Chunk 2 content",
			TokenCount:  150,
			SectionPath: "Section 2",
			Metadata:    map[string]string{"source": "markdown"},
		},
	}

	procDoc := &ProcessedDocument{
		DocID:     "doc1",
		Chunks:    chunks,
		ChunkedAt: time.Now(),
		Chunker:   "structure-aware",
		Config:    chunker.DefaultConfig(),
	}

	err = store.StoreProcessed(ctx, procDoc)
	if err != nil {
		t.Fatalf("StoreProcessed failed: %v", err)
	}

	// Retrieve
	retrieved, err := store.GetProcessed(ctx, "doc1")
	if err != nil {
		t.Fatalf("GetProcessed failed: %v", err)
	}

	if retrieved.DocID != "doc1" {
		t.Errorf("DocID = %s, expected 'doc1'", retrieved.DocID)
	}
	if len(retrieved.Chunks) != 2 {
		t.Errorf("Chunks count = %d, expected 2", len(retrieved.Chunks))
	}
	if retrieved.Chunks[0].ID != "doc1-0" {
		t.Errorf("Chunk 0 ID = %s", retrieved.Chunks[0].ID)
	}
	if retrieved.Chunks[1].ParentID != "doc1-0" {
		t.Errorf("Chunk 1 ParentID = %s", retrieved.Chunks[1].ParentID)
	}
	if retrieved.Chunker != "structure-aware" {
		t.Errorf("Chunker = %s", retrieved.Chunker)
	}
	if retrieved.Config.TargetTokens != 350 {
		t.Errorf("Config.TargetTokens = %d", retrieved.Config.TargetTokens)
	}
}

func TestGetProcessed_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewDocumentStore(tmpDir)
	if err != nil {
		t.Fatalf("NewDocumentStore failed: %v", err)
	}

	ctx := context.Background()
	_, err = store.GetProcessed(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent processed document")
	}
}

func TestExists(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewDocumentStore(tmpDir)
	if err != nil {
		t.Fatalf("NewDocumentStore failed: %v", err)
	}

	ctx := context.Background()
	doc := parser.Document{
		ID:       "test-doc-1",
		Path:     "/tmp/test.md",
		Content:  "# Test",
		Metadata: map[string]string{"source": "markdown"},
		ModTime:  time.Now(),
		Hash:     "abc123",
	}

	testFile := filepath.Join(tmpDir, "test.md")
	err = os.WriteFile(testFile, []byte(doc.Content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	doc.Path = testFile

	// Compute the actual hash that will be used
	actualHash, err := ComputeHash(testFile)
	if err != nil {
		t.Fatalf("ComputeHash failed: %v", err)
	}

	// Before storing
	exists, err := store.Exists(ctx, actualHash)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("Document should not exist before storing")
	}

	// Store
	_, err = store.StoreRaw(ctx, doc)
	if err != nil {
		t.Fatalf("StoreRaw failed: %v", err)
	}

	// After storing
	exists, err = store.Exists(ctx, actualHash)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("Document should exist after storing")
	}
}

func TestDelete(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewDocumentStore(tmpDir)
	if err != nil {
		t.Fatalf("NewDocumentStore failed: %v", err)
	}

	ctx := context.Background()
	doc := parser.Document{
		ID:       "test-doc-1",
		Path:     "/tmp/test.md",
		Content:  "# Test",
		Metadata: map[string]string{"source": "markdown"},
		ModTime:  time.Now(),
		Hash:     "abc123",
	}

	testFile := filepath.Join(tmpDir, "test.md")
	err = os.WriteFile(testFile, []byte(doc.Content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	doc.Path = testFile

	rawDoc, err := store.StoreRaw(ctx, doc)
	if err != nil {
		t.Fatalf("StoreRaw failed: %v", err)
	}

	chunks := []chunker.Chunk{
		{ID: "doc1-0", DocID: "doc1", Content: "Chunk 1"},
	}
	procDoc := &ProcessedDocument{
		DocID:   "doc1",
		Chunks:  chunks,
		Chunker: "test",
	}
	err = store.StoreProcessed(ctx, procDoc)
	if err != nil {
		t.Fatalf("StoreProcessed failed: %v", err)
	}

	// Verify both exist
	_, err = store.GetRaw(ctx, rawDoc.ID)
	if err != nil {
		t.Fatalf("GetRaw failed before delete: %v", err)
	}
	_, err = store.GetProcessed(ctx, "doc1")
	if err != nil {
		t.Fatalf("GetProcessed failed before delete: %v", err)
	}

	// Delete raw document
	err = store.Delete(ctx, rawDoc.ID)
	if err != nil {
		t.Fatalf("Delete raw failed: %v", err)
	}

	// Delete processed document (uses different ID)
	err = store.Delete(ctx, "doc1")
	if err != nil {
		t.Fatalf("Delete processed failed: %v", err)
	}

	// Verify both are gone
	_, err = store.GetRaw(ctx, rawDoc.ID)
	if err == nil {
		t.Error("Raw document should be deleted")
	}
	_, err = store.GetProcessed(ctx, "doc1")
	if err == nil {
		t.Error("Processed document should be deleted")
	}
}

func TestListRaw(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewDocumentStore(tmpDir)
	if err != nil {
		t.Fatalf("NewDocumentStore failed: %v", err)
	}

	ctx := context.Background()

	// Store multiple documents
	for i := 1; i <= 3; i++ {
		doc := parser.Document{
			ID:       "test-doc-" + string(rune('0'+i)),
			Path:     "/tmp/test" + string(rune('0'+i)) + ".md",
			Content:  "# Test " + string(rune('0'+i)),
			Metadata: map[string]string{"source": "markdown", "index": string(rune('0' + i))},
			ModTime:  time.Now(),
			Hash:     "hash" + string(rune('0'+i)),
		}
		testFile := filepath.Join(tmpDir, "test"+string(rune('0'+i))+".md")
		os.WriteFile(testFile, []byte(doc.Content), 0644)
		doc.Path = testFile
		store.StoreRaw(ctx, doc)
	}

	docs, err := store.ListRaw(ctx)
	if err != nil {
		t.Fatalf("ListRaw failed: %v", err)
	}

	if len(docs) != 3 {
		t.Errorf("ListRaw returned %d docs, expected 3", len(docs))
	}

	for _, doc := range docs {
		if doc.ID == "" {
			t.Error("Doc ID should not be empty")
		}
		if doc.Path == "" {
			t.Error("Doc Path should not be empty")
		}
	}
}

func TestComputeHash(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	content := "test content for hashing"
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	hash, err := ComputeHash(filePath)
	if err != nil {
		t.Fatalf("ComputeHash failed: %v", err)
	}

	if hash == "" {
		t.Error("Hash should not be empty")
	}

	// Hash should be deterministic
	hash2, err := ComputeHash(filePath)
	if err != nil {
		t.Fatalf("ComputeHash failed: %v", err)
	}
	if hash != hash2 {
		t.Error("Hash should be deterministic")
	}

	// Hash should be 64 chars (SHA256 hex)
	if len(hash) != 64 {
		t.Errorf("Hash length = %d, expected 64", len(hash))
	}
}

func TestComputeContentHash(t *testing.T) {
	content := []byte("test content")
	hash := ComputeContentHash(content)

	if hash == "" {
		t.Error("Hash should not be empty")
	}
	if len(hash) != 64 {
		t.Errorf("Hash length = %d, expected 64", len(hash))
	}

	// Deterministic
	hash2 := ComputeContentHash(content)
	if hash != hash2 {
		t.Error("Content hash should be deterministic")
	}

	// Different content = different hash
	hash3 := ComputeContentHash([]byte("different content"))
	if hash == hash3 {
		t.Error("Different content should produce different hash")
	}
}

func TestRawDocument_Serialization(t *testing.T) {
	rawDoc := RawDocument{
		ID:       "test-id",
		Path:     "/path/to/file.md",
		Hash:     "abc123",
		Size:     1024,
		ModTime:  time.Now(),
		Metadata: map[string]string{"source": "markdown"},
		StoredAt: time.Now(),
	}

	// Marshal to JSON
	data, err := json.Marshal(rawDoc)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal
	var rawDoc2 RawDocument
	err = json.Unmarshal(data, &rawDoc2)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if rawDoc2.ID != rawDoc.ID {
		t.Errorf("ID mismatch: %s != %s", rawDoc2.ID, rawDoc.ID)
	}
	if rawDoc2.Path != rawDoc.Path {
		t.Errorf("Path mismatch: %s != %s", rawDoc2.Path, rawDoc.Path)
	}
	if rawDoc2.Hash != rawDoc.Hash {
		t.Errorf("Hash mismatch: %s != %s", rawDoc2.Hash, rawDoc.Hash)
	}
	if rawDoc2.Size != rawDoc.Size {
		t.Errorf("Size mismatch: %d != %d", rawDoc2.Size, rawDoc.Size)
	}
	if rawDoc2.Metadata["source"] != "markdown" {
		t.Errorf("Metadata mismatch: %v", rawDoc2.Metadata)
	}
}

func TestProcessedDocument_Serialization(t *testing.T) {
	procDoc := ProcessedDocument{
		DocID: "doc1",
		Chunks: []chunker.Chunk{
			{ID: "doc1-0", DocID: "doc1", Content: "Chunk 1", TokenCount: 100},
			{ID: "doc1-1", DocID: "doc1", ParentID: "doc1-0", Content: "Chunk 2", TokenCount: 150},
		},
		ChunkedAt: time.Now(),
		Chunker:   "structure-aware",
		Config:    chunker.DefaultConfig(),
	}

	data, err := json.Marshal(procDoc)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var procDoc2 ProcessedDocument
	err = json.Unmarshal(data, &procDoc2)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if procDoc2.DocID != procDoc.DocID {
		t.Errorf("DocID mismatch")
	}
	if len(procDoc2.Chunks) != 2 {
		t.Errorf("Chunks count mismatch: %d", len(procDoc2.Chunks))
	}
	if procDoc2.Chunks[0].ID != "doc1-0" {
		t.Errorf("Chunk 0 ID mismatch")
	}
	if procDoc2.Chunks[1].ParentID != "doc1-0" {
		t.Errorf("Chunk 1 ParentID mismatch")
	}
	if procDoc2.Chunker != "structure-aware" {
		t.Errorf("Chunker mismatch: %s", procDoc2.Chunker)
	}
}
