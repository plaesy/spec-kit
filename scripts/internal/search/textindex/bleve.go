package textindex

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
)

// BleveIndex wraps bleve as a TextIndex with BM25 similarity.
type BleveIndex struct {
	config Config
	index  bleve.Index
	path   string
	mu     sync.RWMutex
	closed bool
}

// NewBleveIndex creates a new Bleve text index.
func NewBleveIndex(cfg Config, basePath string) (*BleveIndex, error) {
	if cfg.Type != "bleve" {
		return nil, fmt.Errorf("invalid index type: %s", cfg.Type)
	}

	indexPath := filepath.Join(basePath, "textindex")

	// Create custom mapping for BM25
	indexMapping := createBM25Mapping()

	index, err := bleve.New(indexPath, indexMapping)
	if err != nil {
		// Try to open existing
		index, err = bleve.Open(indexPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create/open bleve index: %w", err)
		}
	}

	return &BleveIndex{
		config: cfg,
		index:  index,
		path:   indexPath,
	}, nil
}

// createBM25Mapping creates a field mapping with BM25 similarity.
func createBM25Mapping() *mapping.IndexMappingImpl {
	// Document mapping
	docMapping := bleve.NewDocumentMapping()

	// ID field (not indexed, just stored)
	idField := bleve.NewKeywordFieldMapping()
	idField.Index = false
	idField.Store = true
	docMapping.AddFieldMappingsAt("id", idField)

	// Content field (indexed with BM25)
	contentField := bleve.NewTextFieldMapping()
	contentField.Index = true
	contentField.Store = true
	docMapping.AddFieldMappingsAt("content", contentField)

	// Metadata field (stored, not indexed)
	metaField := bleve.NewKeywordFieldMapping()
	metaField.Index = false
	metaField.Store = true
	docMapping.AddFieldMappingsAt("metadata", metaField)

	indexMapping := bleve.NewIndexMapping()
	indexMapping.DefaultMapping = docMapping

	return indexMapping
}

// Add adds documents to the index.
func (b *BleveIndex) Add(ctx context.Context, ids []string, documents []string, metadatas []map[string]string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return fmt.Errorf("index is closed")
	}

	if len(ids) != len(documents) {
		return fmt.Errorf("ids and documents length mismatch")
	}

	batch := b.index.NewBatch()

	for i, id := range ids {
		doc := map[string]interface{}{
			"id":      id,
			"content": documents[i],
		}

		if metadatas != nil && i < len(metadatas) {
			doc["metadata"] = metadatas[i]
		}

		if err := batch.Index(id, doc); err != nil {
			return fmt.Errorf("failed to add document %s to batch: %w", id, err)
		}
	}

	if err := b.index.Batch(batch); err != nil {
		return fmt.Errorf("failed to execute batch: %w", err)
	}

	return nil
}

// Search searches for documents matching the query.
func (b *BleveIndex) Search(ctx context.Context, query string, k int, filter map[string]string) ([]SearchResult, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return nil, fmt.Errorf("index is closed")
	}

	// Build query
	// Use match query for BM25 scoring
	matchQuery := bleve.NewMatchQuery(query)
	matchQuery.SetField("content")

	// Build search request
	searchReq := bleve.NewSearchRequest(matchQuery)
	searchReq.Size = k
	searchReq.Fields = []string{"id", "content", "metadata"}

	// Add filter if provided
	if filter != nil && len(filter) > 0 {
		conj := bleve.NewConjunctionQuery()
		for k, v := range filter {
			termQuery := bleve.NewTermQuery(v)
			termQuery.SetField("metadata." + k)
			conj.AddQuery(termQuery)
		}
		searchReq.Query = conj
	}

	// Execute search
	results, err := b.index.Search(searchReq)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Convert results
	searchResults := make([]SearchResult, 0, len(results.Hits))
	for _, hit := range results.Hits {
		// Verify filter manually (bleve filter may not work as expected)
		if filter != nil {
			meta, _ := hit.Fields["metadata"].(map[string]interface{})
			if !matchFilter(meta, filter) {
				continue
			}
		}

		content, _ := hit.Fields["content"].(string)

		searchResults = append(searchResults, SearchResult{
			ID:       hit.ID,
			Score:    hit.Score,
			Content:  content,
			Metadata: hit.Fields["metadata"].(map[string]string),
		})
	}

	return searchResults, nil
}

// Delete removes documents by IDs.
func (b *BleveIndex) Delete(ctx context.Context, ids []string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return fmt.Errorf("index is closed")
	}

	batch := b.index.NewBatch()
	for _, id := range ids {
		batch.Delete(id)
	}

	if err := b.index.Batch(batch); err != nil {
		return fmt.Errorf("failed to delete documents: %w", err)
	}

	return nil
}

// Count returns the number of documents in the index.
func (b *BleveIndex) Count() (int, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return 0, fmt.Errorf("index is closed")
	}

	count, err := b.index.DocCount()
	if err != nil {
		return 0, fmt.Errorf("failed to get doc count: %w", err)
	}

	return int(count), nil
}

// Close closes the index.
func (b *BleveIndex) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}

	b.closed = true
	return b.index.Close()
}

// Name returns the index name.
func (b *BleveIndex) Name() string {
	return "bleve"
}

func matchFilter(metadata map[string]interface{}, filter map[string]string) bool {
	for k, v := range filter {
		metaVal, ok := metadata[k].(string)
		if !ok || metaVal != v {
			return false
		}
	}
	return true
}
