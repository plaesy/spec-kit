package vectorstore

import (
	"context"
)

// VectorStore defines the interface for a vector store.
type VectorStore interface {
	// Add adds vectors to the store.
	Add(ctx context.Context, ids []string, vectors [][]float32, metadatas []map[string]string, documents []string) error

	// Search searches for similar vectors.
	Search(ctx context.Context, queryVector []float32, k int, filter map[string]string) ([]SearchResult, error)

	// Delete removes vectors by IDs.
	Delete(ctx context.Context, ids []string) error

	// Count returns the number of vectors in the store.
	Count() (int, error)

	// Close closes the store.
	Close() error

	// Name returns the store name.
	Name() string
}

// SearchResult represents a search result.
type SearchResult struct {
	ID       string            // Document/chunk ID
	Score    float32           // Similarity score (higher = more similar)
	Content  string            // Document content
	Metadata map[string]string // Metadata
	Vector   []float32         // Optional: the vector itself
}

// Config holds vector store configuration.
type Config struct {
	Type string // "chromem" | "hnswlib" | "lancedb"
	HNSW HNSWConfig
}

type HNSWConfig struct {
	M              int // Number of connections per layer (default: 16)
	EFConstruction int // Size of dynamic candidate list for construction (default: 200)
	EFSearch       int // Size of dynamic candidate list for search (default: 100)
}

func DefaultConfig() Config {
	return Config{
		Type: "chromem",
		HNSW: HNSWConfig{
			M:              16,
			EFConstruction: 200,
			EFSearch:       100,
		},
	}
}
