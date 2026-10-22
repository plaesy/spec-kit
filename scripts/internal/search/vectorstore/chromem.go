package vectorstore

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"sync"

	chromem "github.com/philippgille/chromem-go"
)

// ChromemStore wraps chromem-go as a VectorStore. chromem persists each
// document's vector on disk, so a store opened by a later process searches the
// same vectors as the process that wrote them.
type ChromemStore struct {
	coll   *chromem.Collection
	mu     sync.RWMutex
	closed bool
}

// NewChromemStore opens or creates the vector store under basePath. Vectors are
// supplied to Add by the caller, so the store needs no embedding function.
func NewChromemStore(cfg Config, basePath string) (*ChromemStore, error) {
	if cfg.Type != "chromem" {
		return nil, fmt.Errorf("invalid store type: %s", cfg.Type)
	}

	db, err := chromem.NewPersistentDB(filepath.Join(basePath, "vectorstore"), true)
	if err != nil {
		return nil, fmt.Errorf("failed to open chromem DB: %w", err)
	}

	coll := db.GetCollection("search_vectors", nil)
	if coll == nil {
		coll, err = db.CreateCollection("search_vectors", nil, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create collection: %w", err)
		}
	}

	return &ChromemStore{coll: coll}, nil
}

// Add stores one vector per document. The vectors come from the caller, so the
// text is not embedded a second time.
func (s *ChromemStore) Add(ctx context.Context, ids []string, vectors [][]float32, metadatas []map[string]string, documents []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return fmt.Errorf("store is closed")
	}

	if len(ids) != len(documents) {
		return fmt.Errorf("ids and documents length mismatch")
	}
	if len(vectors) != len(ids) {
		return fmt.Errorf("ids and vectors length mismatch: %d vs %d", len(ids), len(vectors))
	}

	if metadatas == nil {
		metadatas = make([]map[string]string, len(ids))
		for i := range metadatas {
			metadatas[i] = make(map[string]string)
		}
	}

	if err := s.coll.Add(ctx, ids, vectors, metadatas, documents); err != nil {
		return fmt.Errorf("failed to add to chromem: %w", err)
	}
	return nil
}

// Search returns the k most similar documents to the query vector, best first.
func (s *ChromemStore) Search(ctx context.Context, queryVector []float32, k int, filter map[string]string) ([]SearchResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, fmt.Errorf("store is closed")
	}

	allCount := s.coll.Count()
	if allCount == 0 {
		return []SearchResult{}, nil
	}

	// chromem's Query rejects an empty text, so search by the vector itself.
	// It returns every document ranked by the similarity chromem stored at Add.
	allResults, err := s.coll.QueryEmbedding(ctx, queryVector, allCount, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query chromem: %w", err)
	}

	var results []SearchResult
	for _, r := range allResults {
		if filter != nil && !matchFilter(r.Metadata, filter) {
			continue
		}
		if r.Similarity < 0 {
			continue
		}
		results = append(results, SearchResult{
			ID:       r.ID,
			Score:    r.Similarity,
			Content:  r.Content,
			Metadata: r.Metadata,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if k > 0 && len(results) > k {
		results = results[:k]
	}
	return results, nil
}

// Delete removes the documents with the given IDs.
func (s *ChromemStore) Delete(ctx context.Context, ids []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return fmt.Errorf("store is closed")
	}
	if len(ids) == 0 {
		return nil
	}
	if err := s.coll.Delete(ctx, nil, nil, ids...); err != nil {
		return fmt.Errorf("failed to delete from chromem: %w", err)
	}
	return nil
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

// Close marks the store closed. The chromem DB has no close call; it is
// managed by the garbage collector.
func (s *ChromemStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.closed = true
	return nil
}

// Name returns the store name.
func (s *ChromemStore) Name() string {
	return "chromem"
}

func matchFilter(metadata, filter map[string]string) bool {
	for k, v := range filter {
		metaVal, ok := metadata[k]
		if !ok || metaVal != v {
			return false
		}
	}
	return true
}
