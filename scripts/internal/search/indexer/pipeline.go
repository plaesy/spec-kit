package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/plaesy/spec-kit/internal/config"
	"github.com/plaesy/spec-kit/internal/search/chunker"
	"github.com/plaesy/spec-kit/internal/search/embedder"
	"github.com/plaesy/spec-kit/internal/search/parser"
	"github.com/plaesy/spec-kit/internal/search/store"
	"github.com/plaesy/spec-kit/internal/search/textindex"
	"github.com/plaesy/spec-kit/internal/search/vectorstore"
)

// IndexerConfig holds indexer configuration.
type IndexerConfig struct {
	Workers            int  // Number of worker goroutines
	Resume             bool // Resume from checkpoint
	BatchSize          int  // Documents per batch
	CheckpointInterval int  // Checkpoint every N documents
}

// DefaultIndexerConfig returns default indexer configuration.
func DefaultIndexerConfig() IndexerConfig {
	return IndexerConfig{
		Workers:            4,
		Resume:             true,
		BatchSize:          100,
		CheckpointInterval: 50,
	}
}

// Progress tracks indexing progress for resume capability.
type Progress struct {
	TotalFiles      int               `json:"total_files"`
	ProcessedFiles  int               `json:"processed_files"`
	TotalChunks     int               `json:"total_chunks"`
	ProcessedChunks int               `json:"processed_chunks"`
	CurrentFile     string            `json:"current_file"`
	FileHashes      map[string]string `json:"file_hashes"` // path -> content hash
	StartedAt       time.Time         `json:"started_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	Completed       bool              `json:"completed"`
}

// Indexer orchestrates the indexing pipeline.
type Indexer struct {
	config       IndexerConfig
	searchConfig *config.SearchConfig

	parserReg   *parser.ParserRegistry
	chunker     chunker.Chunker
	emb         embedder.Embedder
	vectorStore vectorstore.VectorStore
	textIndex   textindex.TextIndex
	docStore    *store.DocumentStore

	progress     Progress
	progressPath string
	mu           sync.Mutex

	basePath string
}

// NewIndexer creates a new indexer.
func NewIndexer(searchConfig *config.SearchConfig, basePath string) (*Indexer, error) {
	// Create components
	parserReg := parser.NewDefaultRegistry(searchConfig)

	chunkerCfg := chunker.Config{
		TargetTokens:   searchConfig.Chunker.TargetTokens,
		OverlapTokens:  searchConfig.Chunker.OverlapTokens,
		MinChunkTokens: searchConfig.Chunker.MinChunkTokens,
	}
	ch := chunker.NewStructureAwareChunker(chunkerCfg, &chunker.SimpleTokenCounter{})

	emb, err := embedder.NewEmbedder(&searchConfig.Embedder)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedder: %w", err)
	}

	// Create vector store
	vectorStoreConfig := vectorstore.Config{
		Type: searchConfig.VectorStore.Type,
		HNSW: vectorstore.HNSWConfig{
			M:              searchConfig.VectorStore.HNSW.M,
			EFConstruction: searchConfig.VectorStore.HNSW.EFConstruction,
			EFSearch:       searchConfig.VectorStore.HNSW.EFSearch,
		},
	}
	vectorStore, err := vectorstore.NewChromemStore(
		vectorStoreConfig,
		func(ctx context.Context, text string) ([]float32, error) {
			embs, err := emb.Embed(ctx, []string{text})
			if err != nil {
				return nil, err
			}
			if len(embs) == 0 {
				return nil, fmt.Errorf("no embedding returned")
			}
			return embs[0], nil
		},
		basePath,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create vector store: %w", err)
	}

	// Create text index
	textIndexConfig := textindex.Config{
		Type: searchConfig.TextIndex.Type,
	}
	textIndex, err := textindex.NewBleveIndex(textIndexConfig, basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create text index: %w", err)
	}

	// Create document store
	docStore, err := store.NewDocumentStore(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create document store: %w", err)
	}

	progressPath := filepath.Join(basePath, "indexer_progress.json")

	idx := &Indexer{
		config:       DefaultIndexerConfig(),
		searchConfig: searchConfig,
		parserReg:    parserReg,
		chunker:      ch,
		emb:          emb,
		vectorStore:  vectorStore,
		textIndex:    textIndex,
		docStore:     docStore,
		progressPath: progressPath,
		basePath:     basePath,
	}

	// Load progress if resuming
	if idx.config.Resume {
		if err := idx.loadProgress(); err != nil {
			// Non-fatal, just start fresh
			fmt.Printf("Warning: failed to load progress: %v\n", err)
		}
	}

	return idx, nil
}

// IndexDirectory indexes all supported files in a directory.
func (i *Indexer) IndexDirectory(ctx context.Context, rootPath string, incremental bool) error {
	// Scan for files
	extensions := i.getSupportedExtensions()
	docs, err := i.parserReg.ParseDirectory(ctx, rootPath, extensions)
	if err != nil {
		return fmt.Errorf("failed to parse directory: %w", err)
	}

	fmt.Printf("Found %d documents to index\n", len(docs))

	// Filter changed files if incremental
	if incremental {
		docs = i.filterChangedFiles(docs)
		fmt.Printf("After incremental filtering: %d documents to process\n", len(docs))
	}

	if len(docs) == 0 {
		fmt.Println("No documents to index")
		return nil
	}

	// Initialize progress
	i.mu.Lock()
	i.progress = Progress{
		TotalFiles: len(docs),
		FileHashes: make(map[string]string),
		StartedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	i.mu.Unlock()

	// Process documents with worker pool
	return i.processDocuments(ctx, docs)
}

// processDocuments processes documents using worker pool.
func (i *Indexer) processDocuments(ctx context.Context, docs []parser.Document) error {
	// Create channels
	docChan := make(chan parser.Document, len(docs))
	resultChan := make(chan indexResult, len(docs))

	// Start workers
	var wg sync.WaitGroup
	for w := 0; w < i.config.Workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			i.worker(ctx, docChan, resultChan)
		}()
	}

	// Send documents
	go func() {
		for _, doc := range docs {
			select {
			case docChan <- doc:
			case <-ctx.Done():
				close(docChan)
				return
			}
		}
		close(docChan)
	}()

	// Wait for workers
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	processed := 0
	for result := range resultChan {
		if result.err != nil {
			fmt.Printf("Error processing %s: %v\n", result.doc.Path, result.err)
			continue
		}

		processed++

		// Update progress
		i.mu.Lock()
		i.progress.ProcessedFiles = processed
		i.progress.TotalChunks += len(result.chunks)
		i.progress.ProcessedChunks += len(result.chunks)
		i.progress.CurrentFile = result.doc.Path
		i.progress.FileHashes[result.doc.Path] = result.doc.Hash
		i.progress.UpdatedAt = time.Now()
		i.mu.Unlock()

		// Checkpoint
		if processed%i.config.CheckpointInterval == 0 {
			i.saveProgress()
			fmt.Printf("Progress: %d/%d files, %d chunks\n", processed, i.progress.TotalFiles, i.progress.TotalChunks)
		}
	}

	// Final checkpoint
	i.mu.Lock()
	i.progress.Completed = true
	i.progress.UpdatedAt = time.Now()
	i.mu.Unlock()
	i.saveProgress()

	fmt.Printf("Indexing complete: %d files, %d chunks\n", processed, i.progress.TotalChunks)
	return nil
}

type indexResult struct {
	doc    parser.Document
	chunks []chunker.Chunk
	err    error
}

// worker processes documents from the channel.
func (i *Indexer) worker(ctx context.Context, docChan <-chan parser.Document, resultChan chan<- indexResult) {
	for doc := range docChan {
		select {
		case <-ctx.Done():
			resultChan <- indexResult{doc: doc, err: ctx.Err()}
			return
		default:
		}

		chunks, err := i.processDocument(ctx, doc)
		resultChan <- indexResult{doc: doc, chunks: chunks, err: err}
	}
}

// processDocument processes a single document.
func (i *Indexer) processDocument(ctx context.Context, doc parser.Document) ([]chunker.Chunk, error) {
	// Store raw document
	_, err := i.docStore.StoreRaw(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to store raw doc: %w", err)
	}

	// Chunk document
	chunks, err := i.chunker.Chunk(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to chunk doc: %w", err)
	}

	if len(chunks) == 0 {
		return nil, nil
	}

	// Generate embeddings for chunks
	chunkTexts := make([]string, len(chunks))
	for i, chunk := range chunks {
		chunkTexts[i] = chunk.Content
	}

	embeddings, err := i.emb.Embed(ctx, chunkTexts)
	if err != nil {
		return nil, fmt.Errorf("failed to embed chunks: %w", err)
	}

	// Prepare data for vector store
	ids := make([]string, len(chunks))
	metadatas := make([]map[string]string, len(chunks))
	documents := make([]string, len(chunks))

	for i, chunk := range chunks {
		ids[i] = chunk.ID
		documents[i] = chunk.Content
		metadatas[i] = map[string]string{
			"doc_id":       chunk.DocID,
			"section_path": chunk.SectionPath,
			"source":       chunk.Metadata["source"],
		}
	}

	// Add to vector store
	if err := i.vectorStore.Add(ctx, ids, embeddings, metadatas, documents); err != nil {
		return nil, fmt.Errorf("failed to add to vector store: %w", err)
	}

	// Add to text index
	if err := i.textIndex.Add(ctx, ids, documents, metadatas); err != nil {
		return nil, fmt.Errorf("failed to add to text index: %w", err)
	}

	// Store processed document
	procDoc := &store.ProcessedDocument{
		DocID:     doc.ID,
		Chunks:    chunks,
		ChunkedAt: time.Now(),
		Chunker:   i.chunker.Name(),
		Config:    i.chunker.Config(),
	}

	if err := i.docStore.StoreProcessed(ctx, procDoc); err != nil {
		return nil, fmt.Errorf("failed to store processed doc: %w", err)
	}

	return chunks, nil
}

// filterChangedFiles filters out files that haven't changed.
func (i *Indexer) filterChangedFiles(docs []parser.Document) []parser.Document {
	var changed []parser.Document

	for _, doc := range docs {
		// Check if file hash matches previous
		prevHash, exists := i.progress.FileHashes[doc.Path]
		if !exists || prevHash != doc.Hash {
			changed = append(changed, doc)
		}
	}

	return changed
}

// getSupportedExtensions returns all supported file extensions.
func (i *Indexer) getSupportedExtensions() []string {
	return i.parserReg.GetSupportedExtensions()
}

// loadProgress loads progress from checkpoint file.
func (i *Indexer) loadProgress() error {
	i.mu.Lock()
	defer i.mu.Unlock()

	data, err := os.ReadFile(i.progressPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return json.Unmarshal(data, &i.progress)
}

// saveProgress saves progress to checkpoint file.
func (i *Indexer) saveProgress() error {
	i.mu.Lock()
	defer i.mu.Unlock()

	data, err := json.MarshalIndent(i.progress, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := i.progressPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmpPath, i.progressPath)
}

// GetProgress returns current indexing progress.
func (i *Indexer) GetProgress() Progress {
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.progress
}

// Close closes all indexer resources.
func (i *Indexer) Close() error {
	var errs []error

	if err := i.vectorStore.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := i.textIndex.Close(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing indexer: %v", errs)
	}
	return nil
}
