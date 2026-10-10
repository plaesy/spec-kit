package retriever

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/plaesy/spec-kit/internal/config"
	"github.com/plaesy/spec-kit/internal/search/chunker"
	"github.com/plaesy/spec-kit/internal/search/embedder"
	"github.com/plaesy/spec-kit/internal/search/store"
	"github.com/plaesy/spec-kit/internal/search/textindex"
	"github.com/plaesy/spec-kit/internal/search/vectorstore"
)

// HybridRetriever combines vector search and full-text search with RRF.
type HybridRetriever struct {
	config      config.RetrieverConfig
	embedder    embedder.Embedder
	vectorStore vectorstore.VectorStore
	textIndex   textindex.TextIndex

	// For parent-child retrieval
	parentChildEnabled bool
	docStore           DocumentStoreInterface
}

// DocumentStoreInterface defines the interface for accessing parent documents.
type DocumentStoreInterface interface {
	GetProcessed(ctx context.Context, docID string) (*store.ProcessedDocument, error)
}

// Result represents a search result.
type Result struct {
	ID          string            // Chunk ID
	DocID       string            // Parent document ID
	Score       float64           // Combined score
	VectorScore float32           // Vector similarity score
	BM25Score   float64           // BM25 score
	Content     string            // Chunk content
	SectionPath string            // Section path
	Metadata    map[string]string // Metadata
	Chunks      []chunker.Chunk   // Child chunks (for parent-child)
}

// NewHybridRetriever creates a new hybrid retriever.
func NewHybridRetriever(
	retrieverConfig config.RetrieverConfig,
	embedder embedder.Embedder,
	vectorStore vectorstore.VectorStore,
	textIndex textindex.TextIndex,
	docStore DocumentStoreInterface,
) *HybridRetriever {
	return &HybridRetriever{
		config:             retrieverConfig,
		embedder:           embedder,
		vectorStore:        vectorStore,
		textIndex:          textIndex,
		parentChildEnabled: retrieverConfig.ParentChild.Enabled,
		docStore:           docStore,
	}
}

// Retrieve performs hybrid search with RRF.
func (h *HybridRetriever) Retrieve(ctx context.Context, query string, filters map[string]string) ([]Result, error) {
	// Generate query embedding
	queryEmbeddings, err := h.embedder.Embed(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}
	queryVector := queryEmbeddings[0]

	// Search vector store
	vectorResults, err := h.vectorStore.Search(ctx, queryVector, h.config.TopK, filters)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// Search text index
	textResults, err := h.textIndex.Search(ctx, query, h.config.TopK, filters)
	if err != nil {
		return nil, fmt.Errorf("text search failed: %w", err)
	}

	// Fuse results with RRF
	fused := h.fuseRRF(vectorResults, textResults)

	// Apply parent-child expansion if enabled
	if h.parentChildEnabled && h.docStore != nil {
		fused = h.expandParentChild(ctx, fused)
	}

	// Limit to top K
	if len(fused) > h.config.TopK {
		fused = fused[:h.config.TopK]
	}

	return fused, nil
}

// fuseRRF combines vector and text search results using Reciprocal Rank Fusion.
func (h *HybridRetriever) fuseRRF(vectorResults []vectorstore.SearchResult, textResults []textindex.SearchResult) []Result {
	// RRF parameters
	k := h.config.Hybrid.RRFK
	if k <= 0 {
		k = 60
	}

	vectorWeight := h.config.Hybrid.VectorWeight
	bm25Weight := h.config.Hybrid.BM25Weight

	// Build rank maps
	vectorRanks := make(map[string]int)
	for i, r := range vectorResults {
		vectorRanks[r.ID] = i + 1 // 1-based rank
	}

	textRanks := make(map[string]int)
	for i, r := range textResults {
		textRanks[r.ID] = i + 1
	}

	// Collect all unique IDs
	allIDs := make(map[string]bool)
	for id := range vectorRanks {
		allIDs[id] = true
	}
	for id := range textRanks {
		allIDs[id] = true
	}

	// Compute RRF scores
	type scoredResult struct {
		id          string
		rrfScore    float64
		vectorScore float32
		bm25Score   float64
		content     string
		metadata    map[string]string
	}

	var results []scoredResult

	for id := range allIDs {
		var rrfScore float64

		// Vector component
		if rank, ok := vectorRanks[id]; ok {
			rrfScore += vectorWeight * (1.0 / float64(k+rank))
		}

		// BM25 component
		if rank, ok := textRanks[id]; ok {
			rrfScore += bm25Weight * (1.0 / float64(k+rank))
		}

		// Get content and metadata from either result
		var content string
		var metadata map[string]string
		var vectorScore float32
		var bm25Score float64

		if vr, ok := findVectorResult(vectorResults, id); ok {
			content = vr.Content
			metadata = vr.Metadata
			vectorScore = vr.Score
		}
		if tr, ok := findTextResult(textResults, id); ok {
			if content == "" {
				content = tr.Content
			}
			if metadata == nil {
				metadata = tr.Metadata
			}
			bm25Score = tr.Score
		}

		results = append(results, scoredResult{
			id:          id,
			rrfScore:    rrfScore,
			vectorScore: vectorScore,
			bm25Score:   bm25Score,
			content:     content,
			metadata:    metadata,
		})
	}

	// Sort by RRF score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].rrfScore > results[j].rrfScore
	})

	// Convert to Result
	final := make([]Result, len(results))
	for i, r := range results {
		final[i] = Result{
			ID:          r.id,
			Score:       r.rrfScore,
			VectorScore: r.vectorScore,
			BM25Score:   r.bm25Score,
			Content:     r.content,
			Metadata:    r.metadata,
		}
	}

	return final
}

func findVectorResult(results []vectorstore.SearchResult, id string) (vectorstore.SearchResult, bool) {
	for _, r := range results {
		if r.ID == id {
			return r, true
		}
	}
	return vectorstore.SearchResult{}, false
}

func findTextResult(results []textindex.SearchResult, id string) (textindex.SearchResult, bool) {
	for _, r := range results {
		if r.ID == id {
			return r, true
		}
	}
	return textindex.SearchResult{}, false
}

// expandParentChild expands child chunks to include parent context.
func (h *HybridRetriever) expandParentChild(ctx context.Context, results []Result) []Result {
	var expanded []Result
	seen := make(map[string]bool)

	for _, r := range results {
		// Add child chunk
		expanded = append(expanded, r)
		seen[r.ID] = true

		// Get parent document
		if h.docStore != nil && r.DocID != "" {
			procDoc, err := h.docStore.GetProcessed(ctx, r.DocID)
			if err == nil && procDoc != nil {
				// Find sibling chunks (other chunks from same document)
				for _, chunk := range procDoc.Chunks {
					if chunk.ID != r.ID && !seen[chunk.ID] {
						expanded = append(expanded, Result{
							ID:          chunk.ID,
							DocID:       chunk.DocID,
							Score:       r.Score * 0.5, // Lower score for siblings
							VectorScore: 0,
							BM25Score:   0,
							Content:     chunk.Content,
							SectionPath: chunk.SectionPath,
							Metadata:    chunk.Metadata,
						})
						seen[chunk.ID] = true
					}
				}
			}
		}
	}

	return expanded
}

// RetrieveParallel performs parallel vector and text search.
func (h *HybridRetriever) RetrieveParallel(ctx context.Context, query string, filters map[string]string) ([]Result, error) {
	// Generate query embedding
	queryEmbeddings, err := h.embedder.Embed(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}
	queryVector := queryEmbeddings[0]

	var vectorResults []vectorstore.SearchResult
	var textResults []textindex.SearchResult
	var vecErr, txtErr error

	var wg sync.WaitGroup
	wg.Add(2)

	// Vector search
	go func() {
		defer wg.Done()
		vectorResults, vecErr = h.vectorStore.Search(ctx, queryVector, h.config.TopK, filters)
	}()

	// Text search
	go func() {
		defer wg.Done()
		textResults, txtErr = h.textIndex.Search(ctx, query, h.config.TopK, filters)
	}()

	wg.Wait()

	if vecErr != nil {
		return nil, fmt.Errorf("vector search failed: %w", vecErr)
	}
	if txtErr != nil {
		return nil, fmt.Errorf("text search failed: %w", txtErr)
	}

	// Fuse results
	fused := h.fuseRRF(vectorResults, textResults)

	// Apply parent-child expansion if enabled
	if h.parentChildEnabled && h.docStore != nil {
		fused = h.expandParentChild(ctx, fused)
	}

	if len(fused) > h.config.TopK {
		fused = fused[:h.config.TopK]
	}

	return fused, nil
}
