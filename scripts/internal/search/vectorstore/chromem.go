package vectorstore

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"sort"
	"sync"

	chromem "github.com/philippgille/chromem-go"
)

// ChromemStore wraps chromem-go as a VectorStore with full vector access.
type ChromemStore struct {
	config   Config
	db       *chromem.DB
	coll     *chromem.Collection
	embedder chromem.EmbeddingFunc
	path     string
	mu       sync.RWMutex
	closed   bool

	// Cache for direct vector access (for hybrid search)
	vectors map[string][]float32
}

// NewChromemStore creates a new Chromem vector store.
func NewChromemStore(cfg Config, embedder chromem.EmbeddingFunc, basePath string) (*ChromemStore, error) {
	if cfg.Type != "chromem" {
		return nil, fmt.Errorf("invalid store type: %s", cfg.Type)
	}

	dbPath := filepath.Join(basePath, "vectorstore")

	db, err := chromem.NewPersistentDB(dbPath, true)
	if err != nil {
		return nil, fmt.Errorf("failed to open chromem DB: %w", err)
	}

	// Create or get collection
	coll := db.GetCollection("search_vectors", embedder)
	if coll == nil {
		coll, err = db.CreateCollection("search_vectors", nil, embedder)
		if err != nil {
			return nil, fmt.Errorf("failed to create collection: %w", err)
		}
	}

	return &ChromemStore{
		config:   cfg,
		db:       db,
		coll:     coll,
		embedder: embedder,
		path:     dbPath,
		vectors:  make(map[string][]float32),
	}, nil
}

// Add adds documents to the store (embeddings generated via embedder).
func (s *ChromemStore) Add(ctx context.Context, ids []string, vectors [][]float32, metadatas []map[string]string, documents []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return fmt.Errorf("store is closed")
	}

	if len(ids) != len(documents) {
		return fmt.Errorf("ids and documents length mismatch")
	}

	// Prepare metadatas if nil
	if metadatas == nil {
		metadatas = make([]map[string]string, len(ids))
		for i := range metadatas {
			metadatas[i] = make(map[string]string)
		}
	}

	// Use chromem-go's Add with embedder function
	if err := s.coll.Add(ctx, ids, nil, metadatas, documents); err != nil {
		return fmt.Errorf("failed to add to chromem: %w", err)
	}

	// Also store vectors for hybrid search
	if vectors != nil && len(vectors) == len(ids) {
		for i, id := range ids {
			s.vectors[id] = vectors[i]
		}
	} else if s.embedder != nil {
		// Generate embeddings ourselves for hybrid search
		for i, doc := range documents {
			emb, err := s.embedder(ctx, doc)
			if err != nil {
				return fmt.Errorf("failed to embed document %s: %w", ids[i], err)
			}
			s.vectors[ids[i]] = emb
		}
	}

	return nil
}

// SearchByVector searches using a pre-computed query vector.
func (s *ChromemStore) SearchByVector(ctx context.Context, queryVector []float32, k int, filter map[string]string) ([]SearchResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, fmt.Errorf("store is closed")
	}

	if len(s.vectors) == 0 {
		return []SearchResult{}, nil
	}

	type scoredResult struct {
		id       string
		score    float32
		content  string
		metadata map[string]string
		vector   []float32
	}

	var results []scoredResult

	allCount := s.coll.Count()
	if allCount == 0 {
		return []SearchResult{}, nil
	}

	allResults, err := s.coll.Query(ctx, "", allCount, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query chromem: %w", err)
	}

	for _, r := range allResults {
		if filter != nil && !matchFilter(r.Metadata, filter) {
			continue
		}

		vec, ok := s.vectors[r.ID]
		if !ok {
			continue
		}

		score := cosineSimilarity(queryVector, vec)
		if score < 0 {
			continue
		}

		results = append(results, scoredResult{
			id:       r.ID,
			score:    score,
			content:  r.Content,
			metadata: r.Metadata,
			vector:   vec,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	if k > 0 && len(results) > k {
		results = results[:k]
	}

	searchResults := make([]SearchResult, len(results))
	for i, r := range results {
		searchResults[i] = SearchResult{
			ID:       r.id,
			Score:    r.score,
			Content:  r.content,
			Metadata: r.metadata,
			Vector:   r.vector,
		}
	}

	return searchResults, nil
}

// Search searches using query vector.
func (s *ChromemStore) Search(ctx context.Context, queryVector []float32, k int, filter map[string]string) ([]SearchResult, error) {
	return s.SearchByVector(ctx, queryVector, k, filter)
}

// Delete removes vectors by IDs.
func (s *ChromemStore) Delete(ctx context.Context, ids []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return fmt.Errorf("store is closed")
	}

	for _, id := range ids {
		delete(s.vectors, id)
	}

	return fmt.Errorf("delete not implemented; use reindex")
}

// Count returns the number of vectors in the store.
func (s *ChromemStore) Count() (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return 0, fmt.Errorf("store is closed")
	}

	return s.coll.Count(), nil
}

// Close closes the store.
func (s *ChromemStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true
	// chromem-go DB doesn't have a Close method, it's managed by GC
	return nil
}

// Name returns the store name.
func (s *ChromemStore) Name() string {
	return "chromem"
}

func matchFilter(metadata, filter map[string]string) bool {
	for k, v := range filter {
		if metadata[k] != v {
			return false
		}
	}
	return true
}

func cosineSimilarity(a, b []float32) float32 {
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
