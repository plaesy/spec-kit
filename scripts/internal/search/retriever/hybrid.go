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

// QueryOptions narrows one query. The zero value is the plain hybrid search.
type QueryOptions struct {
	// Filters are exact-match metadata filters, applied by the stores.
	Filters map[string]string
	// Where holds predicate filters (date range, path prefix, file type),
	// applied to the candidates before fusion. See ParseFilter.
	Where Filter
}

// Retrieve performs hybrid search with RRF.
func (h *HybridRetriever) Retrieve(ctx context.Context, query string, filters map[string]string) ([]Result, error) {
	return h.RetrieveWith(ctx, query, QueryOptions{Filters: filters})
}

// RetrieveWith is Retrieve with predicate filters. It also expands siblings
// when parent-child expansion is enabled in the config.
func (h *HybridRetriever) RetrieveWith(ctx context.Context, query string, opts QueryOptions) ([]Result, error) {
	fused, err := h.search(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	if h.parentChildEnabled && h.docStore != nil {
		fused = h.expandParentChild(ctx, fused)
	}
	return fused, nil
}

// search is the hybrid search without sibling expansion: the top-K fused chunks,
// after the stores' exact filters and the predicate filters have run.
func (h *HybridRetriever) search(ctx context.Context, query string, opts QueryOptions) ([]Result, error) {
	// Generate query embedding
	queryVector, err := h.embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	// Search vector store
	vectorResults, err := h.vectorStore.Search(ctx, queryVector, h.config.TopK, opts.Filters)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// Search text index
	textResults, err := h.textIndex.Search(ctx, query, h.config.TopK, opts.Filters)
	if err != nil {
		return nil, fmt.Errorf("text search failed: %w", err)
	}

	// Predicate filters run before fusion, so a dropped chunk never takes a rank.
	vectorResults = keepBy(vectorResults, func(r vectorstore.SearchResult) map[string]string { return r.Metadata }, opts.Where)
	textResults = keepBy(textResults, func(r textindex.SearchResult) map[string]string { return r.Metadata }, opts.Where)

	// Fuse results with RRF
	fused := h.fuseRRF(vectorResults, textResults)

	// Limit the matches to top K before expanding. Expanding first let the
	// siblings of a few documents fill the whole list.
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
	// Ties are broken by ID: allIDs is a map, so without this the order of equal
	// scores changes from run to run.
	sort.Slice(results, func(i, j int) bool {
		if results[i].rrfScore != results[j].rrfScore {
			return results[i].rrfScore > results[j].rrfScore
		}
		return results[i].id < results[j].id
	})

	// Convert to Result
	final := make([]Result, len(results))
	for i, r := range results {
		final[i] = Result{
			ID:          r.id,
			DocID:       r.metadata["doc_id"], // parent-child expansion keys off DocID
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

// expandParentChild returns the matched results, in rank order, followed by the
// sibling chunks of their documents. Siblings come after every match, so they
// never change the order of the matches or push a match out of the list. A
// sibling scores half of the match it belongs to, and no chunk is listed twice.
func (h *HybridRetriever) expandParentChild(ctx context.Context, results []Result) []Result {
	seen := make(map[string]bool, len(results))
	expanded := make([]Result, 0, len(results))
	for _, r := range results {
		if seen[r.ID] {
			continue
		}
		seen[r.ID] = true
		expanded = append(expanded, r)
	}

	var siblings []Result
	for _, r := range expanded {
		if h.docStore == nil || r.DocID == "" {
			continue
		}
		procDoc, err := h.docStore.GetProcessed(ctx, r.DocID)
		if err != nil || procDoc == nil {
			continue
		}
		for _, chunk := range procDoc.Chunks {
			if seen[chunk.ID] {
				continue
			}
			seen[chunk.ID] = true
			siblings = append(siblings, Result{
				ID:          chunk.ID,
				DocID:       chunk.DocID,
				Score:       r.Score * 0.5, // Lower score for siblings
				VectorScore: 0,
				BM25Score:   0,
				Content:     chunk.Content,
				SectionPath: chunk.SectionPath,
				Metadata:    chunk.Metadata,
			})
		}
	}

	return append(expanded, siblings...)
}

// RetrieveParallel performs parallel vector and text search.
func (h *HybridRetriever) RetrieveParallel(ctx context.Context, query string, filters map[string]string) ([]Result, error) {
	// Generate query embedding
	queryVector, err := h.embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

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

	// Limit the matches before expanding; see Retrieve.
	if len(fused) > h.config.TopK {
		fused = fused[:h.config.TopK]
	}

	if h.parentChildEnabled && h.docStore != nil {
		fused = h.expandParentChild(ctx, fused)
	}

	return fused, nil
}
