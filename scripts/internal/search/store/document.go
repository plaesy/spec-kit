package store

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/plaesy/spec-kit/internal/search/chunker"
	"github.com/plaesy/spec-kit/internal/search/parser"
)

// DocumentStore manages raw and processed document storage.
type DocumentStore struct {
	rawPath       string // Path for raw document storage
	processedPath string // Path for processed documents (chunks + metadata)
	mu            sync.RWMutex
}

// RawDocument represents an immutable raw document.
type RawDocument struct {
	ID       string            `json:"id"`
	Path     string            `json:"path"`
	Hash     string            `json:"hash"`
	Size     int64             `json:"size"`
	ModTime  time.Time         `json:"mod_time"`
	Metadata map[string]string `json:"metadata"`
	StoredAt time.Time         `json:"stored_at"`
}

// ProcessedDocument represents a chunked and processed document.
type ProcessedDocument struct {
	DocID     string          `json:"doc_id"`
	Chunks    []chunker.Chunk `json:"chunks"`
	ChunkedAt time.Time       `json:"chunked_at"`
	Chunker   string          `json:"chunker"`
	Config    chunker.Config  `json:"config"`
}

// NewDocumentStore creates a new document store.
func NewDocumentStore(basePath string) (*DocumentStore, error) {
	rawPath := filepath.Join(basePath, "store", "raw")
	processedPath := filepath.Join(basePath, "store", "processed")

	for _, p := range []string{rawPath, processedPath} {
		if err := os.MkdirAll(p, 0755); err != nil {
			return nil, fmt.Errorf("failed to create store directory %s: %w", p, err)
		}
	}

	return &DocumentStore{
		rawPath:       rawPath,
		processedPath: processedPath,
	}, nil
}

// StoreRaw stores a raw document (immutable, content-addressable).
func (s *DocumentStore) StoreRaw(ctx context.Context, doc parser.Document) (*RawDocument, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Read file content
	content, err := os.ReadFile(doc.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Compute hash
	hash := sha256.Sum256(content)
	hashStr := fmt.Sprintf("%x", hash)

	// Get file info
	info, err := os.Stat(doc.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	rawDoc := RawDocument{
		ID:       hashStr[:16],
		Path:     doc.Path,
		Hash:     hashStr,
		Size:     info.Size(),
		ModTime:  info.ModTime(),
		Metadata: doc.Metadata,
		StoredAt: time.Now(),
	}

	// Store raw document metadata
	rawDocPath := filepath.Join(s.rawPath, rawDoc.ID+".json")
	data, err := json.MarshalIndent(rawDoc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal raw doc: %w", err)
	}

	if err := os.WriteFile(rawDocPath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write raw doc: %w", err)
	}

	// Also store raw content (optional - for deduplication)
	contentPath := filepath.Join(s.rawPath, rawDoc.ID+".content")
	if err := os.WriteFile(contentPath, content, 0644); err != nil {
		return nil, fmt.Errorf("failed to write raw content: %w", err)
	}

	return &rawDoc, nil
}

// GetRaw retrieves a raw document by ID.
func (s *DocumentStore) GetRaw(ctx context.Context, id string) (*RawDocument, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rawDocPath := filepath.Join(s.rawPath, id+".json")
	data, err := os.ReadFile(rawDocPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("raw document not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read raw doc: %w", err)
	}

	var rawDoc RawDocument
	if err := json.Unmarshal(data, &rawDoc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal raw doc: %w", err)
	}

	return &rawDoc, nil
}

// GetRawContent retrieves the raw content of a document.
func (s *DocumentStore) GetRawContent(ctx context.Context, id string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	contentPath := filepath.Join(s.rawPath, id+".content")
	content, err := os.ReadFile(contentPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("raw content not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read raw content: %w", err)
	}

	return content, nil
}

// StoreProcessed stores a processed document with chunks.
func (s *DocumentStore) StoreProcessed(ctx context.Context, procDoc *ProcessedDocument) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	procPath := filepath.Join(s.processedPath, procDoc.DocID+".json")
	data, err := json.MarshalIndent(procDoc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal processed doc: %w", err)
	}

	if err := os.WriteFile(procPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write processed doc: %w", err)
	}

	return nil
}

// GetProcessed retrieves a processed document by ID.
func (s *DocumentStore) GetProcessed(ctx context.Context, docID string) (*ProcessedDocument, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	procPath := filepath.Join(s.processedPath, docID+".json")
	data, err := os.ReadFile(procPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("processed document not found: %s", docID)
		}
		return nil, fmt.Errorf("failed to read processed doc: %w", err)
	}

	var procDoc ProcessedDocument
	if err := json.Unmarshal(data, &procDoc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal processed doc: %w", err)
	}

	return &procDoc, nil
}

// Exists checks if a raw document exists.
func (s *DocumentStore) Exists(ctx context.Context, hash string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Use first 16 chars of hash, or full hash if shorter
	id := hash
	if len(hash) >= 16 {
		id = hash[:16]
	}
	rawDocPath := filepath.Join(s.rawPath, id+".json")
	_, err := os.Stat(rawDocPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// Delete removes a document and its processed version.
func (s *DocumentStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Delete raw
	rawDocPath := filepath.Join(s.rawPath, id+".json")
	os.Remove(rawDocPath)
	rawContentPath := filepath.Join(s.rawPath, id+".content")
	os.Remove(rawContentPath)

	// Delete processed
	procPath := filepath.Join(s.processedPath, id+".json")
	os.Remove(procPath)

	return nil
}

// ListRaw lists all raw documents.
func (s *DocumentStore) ListRaw(ctx context.Context) ([]RawDocument, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.rawPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read raw directory: %w", err)
	}

	var docs []RawDocument
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		id := entry.Name()[:len(entry.Name())-5] // Remove .json
		rawDoc, err := s.GetRaw(ctx, id)
		if err != nil {
			continue
		}
		docs = append(docs, *rawDoc)
	}

	return docs, nil
}

// ComputeHash computes SHA256 hash of a file.
func ComputeHash(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(content)
	return fmt.Sprintf("%x", hash), nil
}

// ComputeContentHash computes SHA256 hash of content.
func ComputeContentHash(content []byte) string {
	hash := sha256.Sum256(content)
	return fmt.Sprintf("%x", hash)
}
