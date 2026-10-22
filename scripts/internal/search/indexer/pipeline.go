package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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
	FileHashes      map[string]string     `json:"file_hashes"` // path -> content hash
	Indexed         map[string]IndexedDoc `json:"indexed"`     // path -> what was written for it
	StartedAt       time.Time             `json:"started_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
	Completed       bool                  `json:"completed"`
}

// IndexedDoc records what indexing one file wrote, so an update can remove
// exactly those chunks when the file changes or is deleted.
type IndexedDoc struct {
	DocID    string   `json:"doc_id"`
	ChunkIDs []string `json:"chunk_ids"`
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

// tokenizerCounter adapts the embedder's tokenizer to the chunker's counter.
type tokenizerCounter struct{ t embedder.Tokenizer }

// Count returns the model token count of text.
func (c tokenizerCounter) Count(text string) int { return c.t.CountTokens(text) }

// componentFactories builds the components an Indexer needs. NewIndexer uses the
// real constructors. Tests substitute fakes, so the wiring and its error paths
// can be covered without the embedding model, chromem, or Bleve.
type componentFactories struct {
	embedder    func(cfg *config.EmbedderConfig) (embedder.Embedder, error)
	vectorStore func(cfg vectorstore.Config, basePath string) (vectorstore.VectorStore, error)
	textIndex   func(cfg textindex.Config, basePath string) (textindex.TextIndex, error)
	docStore    func(basePath string) (*store.DocumentStore, error)
}

// realComponents returns the factories that create the real components. The
// wrappers return a nil interface on error, so a typed nil never leaks out.
func realComponents() componentFactories {
	return componentFactories{
		embedder: embedder.NewEmbedder,
		vectorStore: func(cfg vectorstore.Config, basePath string) (vectorstore.VectorStore, error) {
			s, err := vectorstore.NewChromemStore(cfg, basePath)
			if err != nil {
				return nil, err
			}
			return s, nil
		},
		textIndex: func(cfg textindex.Config, basePath string) (textindex.TextIndex, error) {
			ix, err := textindex.NewBleveIndex(cfg, basePath)
			if err != nil {
				return nil, err
			}
			return ix, nil
		},
		docStore: store.NewDocumentStore,
	}
}

// NewIndexer creates a new indexer.
func NewIndexer(searchConfig *config.SearchConfig, basePath string) (*Indexer, error) {
	return newIndexerWith(searchConfig, basePath, realComponents())
}

// newIndexerWith creates an indexer from the given component factories.
func newIndexerWith(searchConfig *config.SearchConfig, basePath string, f componentFactories) (*Indexer, error) {
	// Create components
	parserReg := parser.NewDefaultRegistry(searchConfig)

	emb, err := f.embedder(&searchConfig.Embedder)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedder: %w", err)
	}

	// Chunk sizes are counted with the embedding model's tokenizer, so the cap
	// matches what the model actually receives.
	tokenizer, ok := emb.(embedder.Tokenizer)
	if !ok {
		return nil, fmt.Errorf("embedder %T cannot count tokens; chunking needs the model tokenizer", emb)
	}
	chunkerCfg := chunker.Config{
		OverlapTokens:  searchConfig.Chunker.OverlapTokens,
		MinChunkTokens: searchConfig.Chunker.MinChunkTokens,
		MaxChunkTokens: searchConfig.Chunker.MaxChunkTokens,
	}
	ch := chunker.NewStructureAwareChunker(chunkerCfg, tokenizerCounter{tokenizer})

	// Create vector store
	vectorStore, err := f.vectorStore(vectorstore.Config{Type: searchConfig.VectorStore.Type}, basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create vector store: %w", err)
	}

	// Create text index
	textIndex, err := f.textIndex(textindex.Config{Type: searchConfig.TextIndex.Type}, basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create text index: %w", err)
	}

	// Create document store
	docStore, err := f.docStore(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create document store: %w", err)
	}

	return assembleIndexer(searchConfig, basePath, parserReg, ch, emb, vectorStore, textIndex, docStore), nil
}

// assembleIndexer wires the components into an Indexer and, when resuming, loads
// the saved progress. The components are parameters so this wiring can be tested
// without the real model, chromem, or Bleve.
func assembleIndexer(
	searchConfig *config.SearchConfig,
	basePath string,
	parserReg *parser.ParserRegistry,
	ch chunker.Chunker,
	emb embedder.Embedder,
	vectorStore vectorstore.VectorStore,
	textIndex textindex.TextIndex,
	docStore *store.DocumentStore,
) *Indexer {
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

	return idx
}

// IndexDirectory indexes all supported files in a directory. With incremental
// set, a file whose content hash has not changed is left alone. A changed or
// deleted file first has its old chunks removed from every store, so a search
// cannot return text the file no longer contains.
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
		if err := i.removeStale(ctx, docs); err != nil {
			return err
		}
		docs = i.filterChangedFiles(docs)
		fmt.Printf("After incremental filtering: %d documents to process\n", len(docs))
	}

	if len(docs) == 0 {
		fmt.Println("No documents to index")
		return nil
	}

	// Initialize progress. Carry over the hashes of files this run does not
	// touch: a fresh map would drop them, and the next incremental run would then
	// embed those unchanged files again.
	i.mu.Lock()
	hashes := make(map[string]string, len(i.progress.FileHashes))
	for path, hash := range i.progress.FileHashes {
		hashes[path] = hash
	}
	indexed := make(map[string]IndexedDoc, len(i.progress.Indexed))
	for path, entry := range i.progress.Indexed {
		indexed[path] = entry
	}
	i.progress = Progress{
		TotalFiles: len(docs),
		FileHashes: hashes,
		Indexed:    indexed,
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

	// Collect results. A failed file does not stop the run; it is reported at the
	// end, and its hash is not recorded, so the next run retries it.
	processed := 0
	var failures []error
	for result := range resultChan {
		if result.err != nil {
			fmt.Printf("Error processing %s: %v\n", result.doc.Path, result.err)
			failures = append(failures, fmt.Errorf("%s: %w", result.doc.Path, result.err))
			continue
		}

		processed++
		ids := make([]string, len(result.chunks))
		for n, chunk := range result.chunks {
			ids[n] = chunk.ID
		}

		// Update progress
		i.mu.Lock()
		i.progress.ProcessedFiles = processed
		i.progress.TotalChunks += len(result.chunks)
		i.progress.ProcessedChunks += len(result.chunks)
		i.progress.CurrentFile = result.doc.Path
		i.progress.FileHashes[result.doc.Path] = result.doc.Hash
		i.progress.Indexed[result.doc.Path] = IndexedDoc{DocID: result.doc.ID, ChunkIDs: ids}
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
	i.progress.Completed = len(failures) == 0
	i.progress.UpdatedAt = time.Now()
	i.mu.Unlock()
	i.saveProgress()

	fmt.Printf("Indexing complete: %d files, %d chunks\n", processed, i.progress.TotalChunks)
	if len(failures) > 0 {
		return fmt.Errorf("%d of %d documents failed to index; first: %w", len(failures), len(docs), failures[0])
	}
	return nil
}

// removeStale deletes what the index holds for every file that is gone or whose
// content has changed, so the next pass embeds the changed ones again.
func (i *Indexer) removeStale(ctx context.Context, docs []parser.Document) error {
	current := make(map[string]string, len(docs))
	for _, doc := range docs {
		current[doc.Path] = doc.Hash
	}

	i.mu.Lock()
	var stale []string
	for path, hash := range i.progress.FileHashes {
		if now, ok := current[path]; !ok || now != hash {
			stale = append(stale, path)
		}
	}
	i.mu.Unlock()

	for _, path := range stale {
		if err := i.removeDocument(ctx, path); err != nil {
			return fmt.Errorf("removing old chunks of %s: %w", path, err)
		}
	}
	return nil
}

// removeDocument deletes one file's chunks from the vector store and the text
// index, and its processed record, then forgets the file.
func (i *Indexer) removeDocument(ctx context.Context, path string) error {
	i.mu.Lock()
	entry, known := i.progress.Indexed[path]
	i.mu.Unlock()

	if known {
		if len(entry.ChunkIDs) > 0 {
			if err := i.vectorStore.Delete(ctx, entry.ChunkIDs); err != nil {
				return fmt.Errorf("vector store: %w", err)
			}
			if err := i.textIndex.Delete(ctx, entry.ChunkIDs); err != nil {
				return fmt.Errorf("text index: %w", err)
			}
		}
		if err := i.docStore.Delete(ctx, entry.DocID); err != nil {
			return fmt.Errorf("document store: %w", err)
		}
	}

	i.mu.Lock()
	delete(i.progress.Indexed, path)
	delete(i.progress.FileHashes, path)
	i.mu.Unlock()
	return nil
}

// NeedsFullRebuild reports whether the saved progress predates per-file chunk
// tracking. Such an index cannot be updated safely: the chunks of a changed file
// cannot be found, so they would stay in the index.
func (i *Indexer) NeedsFullRebuild() bool {
	i.mu.Lock()
	defer i.mu.Unlock()
	for path := range i.progress.FileHashes {
		if _, ok := i.progress.Indexed[path]; !ok {
			return true
		}
	}
	return false
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

	embeddings, err := i.emb.EmbedPassages(ctx, chunkTexts)
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
			"path":         doc.Path,
			// 0 means the chunk has no source line numbers (see chunker.Chunk).
			"start_line": strconv.Itoa(chunk.StartLine),
			"end_line":   strconv.Itoa(chunk.EndLine),
		}
		// The date filter reads this. A file with no modification time has no
		// date, so the key is left out and the date filter does not match it.
		if !doc.ModTime.IsZero() {
			metadatas[i]["modified"] = doc.ModTime.Format("2006-01-02")
		}
	}

	// Add to vector store
	if err := i.vectorStore.Add(ctx, ids, embeddings, metadatas, documents); err != nil {
		return nil, fmt.Errorf("failed to add to vector store: %w", err)
	}

	// Add to text index
	if err := i.textIndex.Add(ctx, ids, documents, metadatas); err != nil {
		i.dropChunks(ctx, ids)
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
		i.dropChunks(ctx, ids)
		return nil, fmt.Errorf("failed to store processed doc: %w", err)
	}

	return chunks, nil
}

// dropChunks removes chunks a failed document had already written, so a retry
// does not leave them behind. Errors are ignored: the document is retried anyway.
func (i *Indexer) dropChunks(ctx context.Context, ids []string) {
	_ = i.vectorStore.Delete(ctx, ids)
	_ = i.textIndex.Delete(ctx, ids)
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

// GetProgress returns a copy of the current indexing progress. The hash map is
// copied too, so a caller cannot change the indexer's state or race with it.
func (i *Indexer) GetProgress() Progress {
	i.mu.Lock()
	defer i.mu.Unlock()
	p := i.progress
	p.FileHashes = make(map[string]string, len(i.progress.FileHashes))
	for path, hash := range i.progress.FileHashes {
		p.FileHashes[path] = hash
	}
	return p
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
