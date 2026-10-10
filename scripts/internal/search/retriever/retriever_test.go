package retriever

import (
	"context"
	"testing"

	"github.com/plaesy/spec-kit/internal/config"
	"github.com/plaesy/spec-kit/internal/search/chunker"
	"github.com/plaesy/spec-kit/internal/search/embedder"
	"github.com/plaesy/spec-kit/internal/search/store"
	"github.com/plaesy/spec-kit/internal/search/textindex"
	"github.com/plaesy/spec-kit/internal/search/vectorstore"
)

type mockEmbedderForRetriever struct {
	embedder.Embedder
	dimension int
}

func (m *mockEmbedderForRetriever) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for i := range embeddings {
		embeddings[i] = make([]float32, m.dimension)
		for j := range embeddings[i] {
			embeddings[i][j] = float32(i+1) / 100.0
		}
	}
	return embeddings, nil
}

func (m *mockEmbedderForRetriever) Dimension() int {
	return m.dimension
}

func (m *mockEmbedderForRetriever) ModelName() string {
	return "mock-embedder"
}

func (m *mockEmbedderForRetriever) MaxBatchSize() int {
	return 32
}

func (m *mockEmbedderForRetriever) MaxSeqLen() int {
	return 512
}

type mockVectorStoreForRetriever struct {
	vectorstore.VectorStore
	data map[string]vectorstore.SearchResult
}

func (m *mockVectorStoreForRetriever) Add(ctx context.Context, ids []string, vectors [][]float32, metadatas []map[string]string, documents []string) error {
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

func (m *mockVectorStoreForRetriever) Search(ctx context.Context, queryVector []float32, k int, filter map[string]string) ([]vectorstore.SearchResult, error) {
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

func (m *mockVectorStoreForRetriever) Delete(ctx context.Context, ids []string) error {
	for _, id := range ids {
		delete(m.data, id)
	}
	return nil
}

func (m *mockVectorStoreForRetriever) Count() (int, error) {
	return len(m.data), nil
}

func (m *mockVectorStoreForRetriever) Close() error {
	return nil
}

func (m *mockVectorStoreForRetriever) Name() string {
	return "mock-vectorstore"
}

type mockTextIndexForRetriever struct {
	textindex.TextIndex
	data map[string]textindex.SearchResult
}

func (m *mockTextIndexForRetriever) Add(ctx context.Context, ids []string, documents []string, metadatas []map[string]string) error {
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

func (m *mockTextIndexForRetriever) Search(ctx context.Context, query string, k int, filter map[string]string) ([]textindex.SearchResult, error) {
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

func (m *mockTextIndexForRetriever) Delete(ctx context.Context, ids []string) error {
	for _, id := range ids {
		delete(m.data, id)
	}
	return nil
}

func (m *mockTextIndexForRetriever) Count() (int, error) {
	return len(m.data), nil
}

func (m *mockTextIndexForRetriever) Close() error {
	return nil
}

func (m *mockTextIndexForRetriever) Name() string {
	return "mock-textindex"
}

type mockDocStoreForRetriever struct {
	DocumentStoreInterface
	docs map[string]*store.ProcessedDocument
}

func (m *mockDocStoreForRetriever) GetProcessed(ctx context.Context, docID string) (*store.ProcessedDocument, error) {
	if m.docs == nil {
		return nil, nil
	}
	return m.docs[docID], nil
}

func TestNewHybridRetriever(t *testing.T) {
	retrieverConfig := config.RetrieverConfig{
		TopK: 10,
		Hybrid: config.HybridConfig{
			RRFK:         60,
			VectorWeight: 1.0,
			BM25Weight:   1.0,
		},
		ParentChild: config.ParentChildConfig{
			Enabled: false,
		},
	}

	mockEmb := &mockEmbedderForRetriever{dimension: 768}
	mockVS := &mockVectorStoreForRetriever{}
	mockTI := &mockTextIndexForRetriever{}
	mockDS := &mockDocStoreForRetriever{}

	retriever := NewHybridRetriever(retrieverConfig, mockEmb, mockVS, mockTI, mockDS)

	if retriever == nil {
		t.Fatal("NewHybridRetriever returned nil")
	}
	if retriever.config.TopK != 10 {
		t.Errorf("TopK = %d", retriever.config.TopK)
	}
	if retriever.config.Hybrid.RRFK != 60 {
		t.Errorf("RRFK = %d", retriever.config.Hybrid.RRFK)
	}
	if !retriever.parentChildEnabled {
		t.Log("Parent-child disabled as expected")
	}
}

func TestHybridRetriever_Retrieve(t *testing.T) {
	retrieverConfig := config.RetrieverConfig{
		TopK: 5,
		Hybrid: config.HybridConfig{
			RRFK:         60,
			VectorWeight: 1.0,
			BM25Weight:   1.0,
		},
		ParentChild: config.ParentChildConfig{
			Enabled: false,
		},
	}

	mockEmb := &mockEmbedderForRetriever{dimension: 3} // Small dimension for test
	mockVS := &mockVectorStoreForRetriever{}
	mockTI := &mockTextIndexForRetriever{}
	mockDS := &mockDocStoreForRetriever{}

	retriever := NewHybridRetriever(retrieverConfig, mockEmb, mockVS, mockTI, mockDS)

	// Add test data to vector store
	ctx := context.Background()
	vsData := map[string]vectorstore.SearchResult{
		"chunk1": {
			ID:       "chunk1",
			Content:  "Vector search result 1",
			Metadata: map[string]string{"doc_id": "doc1", "source": "test"},
			Vector:   []float32{1, 0, 0},
			Score:    0.9,
		},
		"chunk2": {
			ID:       "chunk2",
			Content:  "Vector search result 2",
			Metadata: map[string]string{"doc_id": "doc1", "source": "test"},
			Vector:   []float32{0, 1, 0},
			Score:    0.8,
		},
		"chunk3": {
			ID:       "chunk3",
			Content:  "Vector search result 3",
			Metadata: map[string]string{"doc_id": "doc2", "source": "test"},
			Vector:   []float32{0, 0, 1},
			Score:    0.7,
		},
	}
	mockVS.data = vsData

	// Add test data to text index
	tiData := map[string]textindex.SearchResult{
		"chunk1": {
			ID:       "chunk1",
			Content:  "Text search result 1",
			Metadata: map[string]string{"doc_id": "doc1", "source": "test"},
			Score:    1.5,
		},
		"chunk2": {
			ID:       "chunk2",
			Content:  "Text search result 2",
			Metadata: map[string]string{"doc_id": "doc1", "source": "test"},
			Score:    1.2,
		},
		"chunk4": { // Only in text index
			ID:       "chunk4",
			Content:  "Text search result 4",
			Metadata: map[string]string{"doc_id": "doc2", "source": "test"},
			Score:    1.0,
		},
	}
	mockTI.data = tiData

	results, err := retriever.Retrieve(ctx, "test query", nil)
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("Expected results")
	}

	t.Logf("Retrieved %d results", len(results))
	for i, r := range results {
		t.Logf("  %d: ID=%s, Score=%.4f, VectorScore=%.4f, BM25Score=%.4f", i, r.ID, r.Score, r.VectorScore, r.BM25Score)
	}

	// Should have results from both sources (RRF fusion)
	// chunk1 and chunk2 in both, chunk3 only in vector, chunk4 only in text
	expectedIDs := map[string]bool{"chunk1": true, "chunk2": true, "chunk3": true, "chunk4": true}
	for _, r := range results {
		if !expectedIDs[r.ID] {
			t.Errorf("Unexpected result ID: %s", r.ID)
		}
		delete(expectedIDs, r.ID)
	}
	if len(expectedIDs) > 0 {
		t.Logf("Missing expected IDs: %v", expectedIDs)
	}
}

func TestHybridRetriever_Retrieve_WithFilter(t *testing.T) {
	retrieverConfig := config.RetrieverConfig{
		TopK: 10,
		Hybrid: config.HybridConfig{
			RRFK:         60,
			VectorWeight: 1.0,
			BM25Weight:   1.0,
		},
		ParentChild: config.ParentChildConfig{
			Enabled: false,
		},
	}

	mockEmb := &mockEmbedderForRetriever{dimension: 3}
	mockVS := &mockVectorStoreForRetriever{}
	mockTI := &mockTextIndexForRetriever{}
	mockDS := &mockDocStoreForRetriever{}

	retriever := NewHybridRetriever(retrieverConfig, mockEmb, mockVS, mockTI, mockDS)

	ctx := context.Background()
	vsData := map[string]vectorstore.SearchResult{
		"chunk1": {ID: "chunk1", Content: "c1", Metadata: map[string]string{"category": "a"}, Vector: []float32{1, 0, 0}},
		"chunk2": {ID: "chunk2", Content: "c2", Metadata: map[string]string{"category": "b"}, Vector: []float32{0, 1, 0}},
	}
	mockVS.data = vsData

	tiData := map[string]textindex.SearchResult{
		"chunk1": {ID: "chunk1", Content: "c1", Metadata: map[string]string{"category": "a"}, Score: 1.0},
		"chunk2": {ID: "chunk2", Content: "c2", Metadata: map[string]string{"category": "b"}, Score: 1.0},
	}
	mockTI.data = tiData

	// Filter for category "a"
	results, err := retriever.Retrieve(ctx, "test", map[string]string{"category": "a"})
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}

	for _, r := range results {
		if r.Metadata["category"] != "a" {
			t.Errorf("Result %s has wrong category: %s", r.ID, r.Metadata["category"])
		}
	}
}

func TestHybridRetriever_FuseRRF(t *testing.T) {
	retrieverConfig := config.RetrieverConfig{
		TopK: 10,
		Hybrid: config.HybridConfig{
			RRFK:         60,
			VectorWeight: 1.0,
			BM25Weight:   1.0,
		},
		ParentChild: config.ParentChildConfig{
			Enabled: false,
		},
	}

	mockEmb := &mockEmbedderForRetriever{dimension: 3}
	mockVS := &mockVectorStoreForRetriever{}
	mockTI := &mockTextIndexForRetriever{}
	mockDS := &mockDocStoreForRetriever{}

	retriever := NewHybridRetriever(retrieverConfig, mockEmb, mockVS, mockTI, mockDS)

	vectorResults := []vectorstore.SearchResult{
		{ID: "a", Score: 0.9, Content: "vec a", Metadata: map[string]string{}},
		{ID: "b", Score: 0.8, Content: "vec b", Metadata: map[string]string{}},
		{ID: "c", Score: 0.7, Content: "vec c", Metadata: map[string]string{}},
	}

	textResults := []textindex.SearchResult{
		{ID: "b", Score: 1.5, Content: "text b", Metadata: map[string]string{}},
		{ID: "c", Score: 1.2, Content: "text c", Metadata: map[string]string{}},
		{ID: "d", Score: 1.0, Content: "text d", Metadata: map[string]string{}},
	}

	fused := retriever.fuseRRF(vectorResults, textResults)

	if len(fused) != 4 {
		t.Errorf("Expected 4 fused results, got %d", len(fused))
	}

	// Check all IDs present
	idSet := make(map[string]bool)
	for _, r := range fused {
		idSet[r.ID] = true
	}
	expected := []string{"a", "b", "c", "d"}
	for _, id := range expected {
		if !idSet[id] {
			t.Errorf("Missing ID in fused results: %s", id)
		}
	}

	// Results should be sorted by RRF score (descending)
	for i := 0; i < len(fused)-1; i++ {
		if fused[i].Score < fused[i+1].Score {
			t.Errorf("Results not sorted by score: %f < %f", fused[i].Score, fused[i+1].Score)
		}
	}

	// "b" and "c" appear in both, should have higher RRF scores
	// "a" only in vector, "d" only in text
	var scoreB, scoreC, scoreA, scoreD float64
	for _, r := range fused {
		switch r.ID {
		case "a":
			scoreA = r.Score
		case "b":
			scoreB = r.Score
		case "c":
			scoreC = r.Score
		case "d":
			scoreD = r.Score
		}
	}

	// Both "b" and "c" should rank higher than "a" and "d"
	if scoreB <= scoreA || scoreB <= scoreD {
		t.Errorf("Expected b (in both) to rank higher: b=%.4f, a=%.4f, d=%.4f", scoreB, scoreA, scoreD)
	}
	if scoreC <= scoreA || scoreC <= scoreD {
		t.Errorf("Expected c (in both) to rank higher: c=%.4f, a=%.4f, d=%.4f", scoreC, scoreA, scoreD)
	}
}

func TestHybridRetriever_FuseRRF_CustomWeights(t *testing.T) {
	retrieverConfig := config.RetrieverConfig{
		TopK: 10,
		Hybrid: config.HybridConfig{
			RRFK:         60,
			VectorWeight: 2.0, // Vector weighted higher
			BM25Weight:   0.5,
		},
		ParentChild: config.ParentChildConfig{
			Enabled: false,
		},
	}

	mockEmb := &mockEmbedderForRetriever{dimension: 3}
	mockVS := &mockVectorStoreForRetriever{}
	mockTI := &mockTextIndexForRetriever{}
	mockDS := &mockDocStoreForRetriever{}

	retriever := NewHybridRetriever(retrieverConfig, mockEmb, mockVS, mockTI, mockDS)

	vectorResults := []vectorstore.SearchResult{
		{ID: "vec_only", Score: 0.9, Content: "v", Metadata: map[string]string{}},
	}

	textResults := []textindex.SearchResult{
		{ID: "text_only", Score: 1.5, Content: "t", Metadata: map[string]string{}},
	}

	fused := retriever.fuseRRF(vectorResults, textResults)

	if len(fused) != 2 {
		t.Errorf("Expected 2 results, got %d", len(fused))
	}

	// With VectorWeight=2.0, BM25Weight=0.5, vec_only should rank higher
	if fused[0].ID != "vec_only" {
		t.Errorf("Expected vec_only to rank first with higher vector weight, got %s", fused[0].ID)
	}
}

func TestHybridRetriever_ExpandParentChild(t *testing.T) {
	retrieverConfig := config.RetrieverConfig{
		TopK: 10,
		Hybrid: config.HybridConfig{
			RRFK:         60,
			VectorWeight: 1.0,
			BM25Weight:   1.0,
		},
		ParentChild: config.ParentChildConfig{
			Enabled: true,
		},
	}

	mockEmb := &mockEmbedderForRetriever{dimension: 3}
	mockVS := &mockVectorStoreForRetriever{}
	mockTI := &mockTextIndexForRetriever{}

	// Mock doc store with parent document
	parentDoc := &store.ProcessedDocument{
		DocID: "doc1",
		Chunks: []chunker.Chunk{
			{ID: "chunk1", DocID: "doc1", Content: "Chunk 1", TokenCount: 100},
			{ID: "chunk2", DocID: "doc1", Content: "Chunk 2", TokenCount: 150},
			{ID: "chunk3", DocID: "doc1", Content: "Chunk 3", TokenCount: 120},
		},
	}
	mockDS := &mockDocStoreForRetriever{
		docs: map[string]*store.ProcessedDocument{
			"doc1": parentDoc,
		},
	}

	retriever := NewHybridRetriever(retrieverConfig, mockEmb, mockVS, mockTI, mockDS)

	ctx := context.Background()
	// Initial results only have chunk1
	initialResults := []Result{
		{ID: "chunk1", DocID: "doc1", Score: 0.9, Content: "Chunk 1"},
	}

	expanded := retriever.expandParentChild(ctx, initialResults)

	// Should have original + siblings (chunk2, chunk3)
	if len(expanded) != 3 {
		t.Errorf("Expected 3 expanded results (1 original + 2 siblings), got %d", len(expanded))
	}

	// Check IDs
	idSet := make(map[string]bool)
	for _, r := range expanded {
		idSet[r.ID] = true
	}
	for _, id := range []string{"chunk1", "chunk2", "chunk3"} {
		if !idSet[id] {
			t.Errorf("Missing chunk in expanded: %s", id)
		}
	}

	// Sibling chunks should have lower scores
	for _, r := range expanded {
		if r.ID == "chunk1" {
			if r.Score != 0.9 {
				t.Errorf("Original chunk score changed: %f", r.Score)
			}
		} else {
			if r.Score >= 0.9 {
				t.Errorf("Sibling chunk score should be lower: %f", r.Score)
			}
		}
	}
}

func TestHybridRetriever_ExpandParentChild_Disabled(t *testing.T) {
	retrieverConfig := config.RetrieverConfig{
		TopK: 10,
		Hybrid: config.HybridConfig{
			RRFK:         60,
			VectorWeight: 1.0,
			BM25Weight:   1.0,
		},
		ParentChild: config.ParentChildConfig{
			Enabled: false, // Disabled
		},
	}

	mockEmb := &mockEmbedderForRetriever{dimension: 3}
	mockVS := &mockVectorStoreForRetriever{}
	mockTI := &mockTextIndexForRetriever{}
	mockDS := &mockDocStoreForRetriever{}

	retriever := NewHybridRetriever(retrieverConfig, mockEmb, mockVS, mockTI, mockDS)

	ctx := context.Background()
	initialResults := []Result{
		{ID: "chunk1", DocID: "doc1", Score: 0.9, Content: "Chunk 1"},
	}

	expanded := retriever.expandParentChild(ctx, initialResults)

	// Should return unchanged when disabled
	if len(expanded) != 1 {
		t.Errorf("Expected 1 result when parent-child disabled, got %d", len(expanded))
	}
	if expanded[0].ID != "chunk1" {
		t.Errorf("Result ID changed: %s", expanded[0].ID)
	}
}

func TestHybridRetriever_RetrieveParallel(t *testing.T) {
	retrieverConfig := config.RetrieverConfig{
		TopK: 5,
		Hybrid: config.HybridConfig{
			RRFK:         60,
			VectorWeight: 1.0,
			BM25Weight:   1.0,
		},
		ParentChild: config.ParentChildConfig{
			Enabled: false,
		},
	}

	mockEmb := &mockEmbedderForRetriever{dimension: 3}
	mockVS := &mockVectorStoreForRetriever{}
	mockTI := &mockTextIndexForRetriever{}
	mockDS := &mockDocStoreForRetriever{}

	retriever := NewHybridRetriever(retrieverConfig, mockEmb, mockVS, mockTI, mockDS)

	ctx := context.Background()
	vsData := map[string]vectorstore.SearchResult{
		"chunk1": {ID: "chunk1", Content: "v1", Metadata: map[string]string{}, Vector: []float32{1, 0, 0}},
	}
	mockVS.data = vsData

	tiData := map[string]textindex.SearchResult{
		"chunk1": {ID: "chunk1", Content: "t1", Metadata: map[string]string{}, Score: 1.0},
	}
	mockTI.data = tiData

	results, err := retriever.RetrieveParallel(ctx, "test", nil)
	if err != nil {
		t.Fatalf("RetrieveParallel failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
	if results[0].ID != "chunk1" {
		t.Errorf("Result ID = %s", results[0].ID)
	}
}

func TestResult_Fields(t *testing.T) {
	result := Result{
		ID:          "chunk1",
		DocID:       "doc1",
		Score:       0.85,
		VectorScore: 0.9,
		BM25Score:   1.2,
		Content:     "Test content",
		SectionPath: "Section 1",
		Metadata:    map[string]string{"source": "test"},
		Chunks: []chunker.Chunk{
			{ID: "chunk2", DocID: "doc1", Content: "Sibling"},
		},
	}

	if result.ID != "chunk1" {
		t.Errorf("ID = %s", result.ID)
	}
	if result.DocID != "doc1" {
		t.Errorf("DocID = %s", result.DocID)
	}
	if result.Score != 0.85 {
		t.Errorf("Score = %f", result.Score)
	}
	if result.VectorScore != 0.9 {
		t.Errorf("VectorScore = %f", result.VectorScore)
	}
	if result.BM25Score != 1.2 {
		t.Errorf("BM25Score = %f", result.BM25Score)
	}
	if result.Content != "Test content" {
		t.Errorf("Content = %s", result.Content)
	}
	if result.SectionPath != "Section 1" {
		t.Errorf("SectionPath = %s", result.SectionPath)
	}
	if result.Metadata["source"] != "test" {
		t.Errorf("Metadata = %v", result.Metadata)
	}
	if len(result.Chunks) != 1 {
		t.Errorf("Chunks count = %d", len(result.Chunks))
	}
}
