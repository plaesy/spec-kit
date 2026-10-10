package textindex

import (
	"context"
)

// TextIndex defines the interface for a full-text search index.
type TextIndex interface {
	// Add adds documents to the index.
	Add(ctx context.Context, ids []string, documents []string, metadatas []map[string]string) error

	// Search searches for documents matching the query.
	Search(ctx context.Context, query string, k int, filter map[string]string) ([]SearchResult, error)

	// Delete removes documents by IDs.
	Delete(ctx context.Context, ids []string) error

	// Count returns the number of documents in the index.
	Count() (int, error)

	// Close closes the index.
	Close() error

	// Name returns the index name.
	Name() string
}

// SearchResult represents a text search result.
type SearchResult struct {
	ID       string            // Document ID
	Score    float64           // BM25 score
	Content  string            // Document content
	Metadata map[string]string // Metadata
}

// Config holds text index configuration.
type Config struct {
	Type string // "bleve" | "sqlite" | "meilisearch"
}

func DefaultConfig() Config {
	return Config{
		Type: "bleve",
	}
}
