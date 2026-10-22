package indexer

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/plaesy/spec-kit/internal/config"
	"github.com/plaesy/spec-kit/internal/search/chunker"
	"github.com/plaesy/spec-kit/internal/search/embedder"
	"github.com/plaesy/spec-kit/internal/search/parser"
	"github.com/plaesy/spec-kit/internal/search/store"
	"github.com/plaesy/spec-kit/internal/search/textindex"
	"github.com/plaesy/spec-kit/internal/search/vectorstore"
)

// wordCounter stands in for the model tokenizer: one token per word.
type wordCounter struct{}

func (wordCounter) Count(text string) int { return len(strings.Fields(text)) }

// failingEmbedder fails any batch that contains a passage with the word BOOM,
// so one bad document can be told apart from the rest.
type failingEmbedder struct {
	embedder.Embedder
	dimension int
}

func (f *failingEmbedder) EmbedPassages(ctx context.Context, texts []string) ([][]float32, error) {
	for _, t := range texts {
		if strings.Contains(t, "BOOM") {
			return nil, errors.New("model refused the passage")
		}
	}
	return (&mockEmbedderForIndexer{dimension: f.dimension}).EmbedPassages(ctx, texts)
}

// closeErrTextIndex reports an error from Close, to reach the error branch.
type closeErrTextIndex struct {
	mockTextIndexForIndexer
}

func (closeErrTextIndex) Close() error { return errors.New("bleve busy") }

// newPipelineIndexer builds an Indexer around the fakes, in a temp directory.
// It does not call NewIndexer, which would load the real embedding model and
// open chromem and Bleve. Workers is 1 because the fake stores are not safe for
// concurrent writes.
func newPipelineIndexer(t *testing.T, emb embedder.Embedder) (*Indexer, *mockVectorStoreForIndexer, *mockTextIndexForIndexer) {
	t.Helper()
	base := t.TempDir()
	docStore, err := store.NewDocumentStore(base)
	if err != nil {
		t.Fatalf("NewDocumentStore: %v", err)
	}
	vs := &mockVectorStoreForIndexer{data: map[string]vectorstore.SearchResult{}}
	ti := &mockTextIndexForIndexer{data: map[string]textindex.SearchResult{}}
	cfg := DefaultIndexerConfig()
	cfg.Workers = 1
	cfg.CheckpointInterval = 1
	idx := &Indexer{
		config:       cfg,
		searchConfig: &config.SearchConfig{},
		parserReg:    parser.NewDefaultRegistry(&config.SearchConfig{}),
		chunker: chunker.NewStructureAwareChunker(
			chunker.Config{MinChunkTokens: 1, MaxChunkTokens: 50}, wordCounter{}),
		emb:          emb,
		vectorStore:  vs,
		textIndex:    ti,
		docStore:     docStore,
		progressPath: filepath.Join(base, "indexer_progress.json"),
		basePath:     base,
		progress:     Progress{FileHashes: map[string]string{}},
	}
	return idx, vs, ti
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

const guideMD = `# Guide

## Refunds

Refunds are issued within thirty days of purchase.

## Shipping

Orders ship within two business days.
`

// chunkContaining returns the stored metadata of the chunk whose text contains s.
func chunkContaining(t *testing.T, vs *mockVectorStoreForIndexer, s string) map[string]string {
	t.Helper()
	for _, r := range vs.data {
		if strings.Contains(r.Content, s) {
			return r.Metadata
		}
	}
	t.Fatalf("no stored chunk contains %q", s)
	return nil
}

// The metadata written with each chunk: the path, the document id, the source
// line range, and the modified date. A chunk must carry the lines of the text it
// holds, not the lines of the whole file.
func TestIndexDirectory_WritesMetadataPerChunk(t *testing.T) {
	idx, vs, ti := newPipelineIndexer(t, &mockEmbedderForIndexer{dimension: 3})
	dir := t.TempDir()
	guide := filepath.Join(dir, "docs", "guide.md")
	writeFile(t, guide, guideMD)
	modTime := time.Date(2026, 3, 4, 12, 0, 0, 0, time.UTC)
	if err := os.Chtimes(guide, modTime, modTime); err != nil {
		t.Fatal(err)
	}

	if err := idx.IndexDirectory(context.Background(), dir, false); err != nil {
		t.Fatalf("IndexDirectory: %v", err)
	}
	if len(vs.data) == 0 {
		t.Fatal("no chunks were stored in the vector store")
	}
	if len(ti.data) != len(vs.data) {
		t.Errorf("text index has %d entries, vector store %d; both must get the same chunks", len(ti.data), len(vs.data))
	}

	meta := chunkContaining(t, vs, "Refunds are issued")
	if got := meta["path"]; got != guide {
		t.Errorf("path = %q, want %q", got, guide)
	}
	if meta["doc_id"] == "" {
		t.Error("doc_id is empty")
	}
	if got := meta["modified"]; got != "2026-03-04" {
		t.Errorf("modified = %q, want 2026-03-04", got)
	}
	start, err := strconv.Atoi(meta["start_line"])
	if err != nil {
		t.Fatalf("start_line %q is not a number", meta["start_line"])
	}
	end, err := strconv.Atoi(meta["end_line"])
	if err != nil {
		t.Fatalf("end_line %q is not a number", meta["end_line"])
	}
	// "Refunds are issued" is on line 5 of guideMD.
	if start > 5 || end < 5 || start < 1 || end < start {
		t.Errorf("line range %d-%d does not cover line 5 of the source", start, end)
	}
	if _, ok := meta["source"]; !ok {
		t.Error(`"source" key missing (written even when empty)`)
	}
	if _, ok := meta["section_path"]; !ok {
		t.Error(`"section_path" key missing`)
	}
}

// A file with no modification time must not get a "modified" key, so the date
// filter does not match it by accident.
func TestProcessDocument_NoModTimeOmitsModifiedKey(t *testing.T) {
	idx, vs, _ := newPipelineIndexer(t, &mockEmbedderForIndexer{dimension: 3})
	doc := parser.Document{
		ID:      "d-nomtime",
		Path:    notesPath(t, "# Notes\n\nA paragraph with enough words to become a chunk."),
		Content: "# Notes\n\nA paragraph with enough words to become a chunk.",
		Hash:    "h1",
	}
	chunks, err := idx.processDocument(context.Background(), doc)
	if err != nil {
		t.Fatalf("processDocument: %v", err)
	}
	if len(chunks) == 0 {
		t.Fatal("expected at least one chunk")
	}
	for id, r := range vs.data {
		if _, ok := r.Metadata["modified"]; ok {
			t.Errorf("chunk %s has a modified key despite a zero ModTime", id)
		}
	}
}

// An empty document produces no chunks and no stored vectors, and is not an error.
func TestProcessDocument_EmptyDocumentStoresNothing(t *testing.T) {
	idx, vs, ti := newPipelineIndexer(t, &mockEmbedderForIndexer{dimension: 3})
	chunks, err := idx.processDocument(context.Background(), parser.Document{ID: "doc-empty", Path: notesPath(t, ""), Hash: "h"})
	if err != nil {
		t.Fatalf("processDocument: %v", err)
	}
	if chunks != nil {
		t.Errorf("chunks = %v, want none", chunks)
	}
	if len(vs.data) != 0 || len(ti.data) != 0 {
		t.Errorf("stores got entries for an empty document: vectors %d, text %d", len(vs.data), len(ti.data))
	}
}

// The second run with nothing changed must index nothing. The embedder call
// count is the proof: unchanged files are not embedded again.
func TestIndexDirectory_IncrementalSkipsUnchangedFiles(t *testing.T) {
	emb := &mockEmbedderForIndexer{dimension: 3}
	idx, _, _ := newPipelineIndexer(t, emb)
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.md"), guideMD)
	writeFile(t, filepath.Join(dir, "b.md"), "# Other\n\nSomething else entirely.\n")
	ctx := context.Background()

	if err := idx.IndexDirectory(ctx, dir, false); err != nil {
		t.Fatalf("first run: %v", err)
	}
	afterFirst := emb.callCount
	if afterFirst == 0 {
		t.Fatal("first run embedded nothing")
	}

	if err := idx.IndexDirectory(ctx, dir, true); err != nil {
		t.Fatalf("second run: %v", err)
	}
	if emb.callCount != afterFirst {
		t.Errorf("second run embedded %d more batches; nothing changed, want 0", emb.callCount-afterFirst)
	}
}

// Changing one file re-embeds only that file. Then the next incremental run
// must still skip the other file. Before the fix, the progress was reset to the
// changed file only, so the unchanged file lost its hash and was embedded again.
func TestIndexDirectory_IncrementalKeepsHashesOfUnchangedFiles(t *testing.T) {
	emb := &mockEmbedderForIndexer{dimension: 3}
	idx, _, _ := newPipelineIndexer(t, emb)
	dir := t.TempDir()
	a := filepath.Join(dir, "a.md")
	writeFile(t, a, guideMD)
	writeFile(t, filepath.Join(dir, "b.md"), "# Other\n\nSomething else entirely.\n")
	ctx := context.Background()

	if err := idx.IndexDirectory(ctx, dir, false); err != nil {
		t.Fatalf("first run: %v", err)
	}

	writeFile(t, a, guideMD+"\nA new paragraph appended after the first run.\n")
	before := emb.callCount
	if err := idx.IndexDirectory(ctx, dir, true); err != nil {
		t.Fatalf("second run: %v", err)
	}
	if emb.callCount == before {
		t.Fatal("second run did not re-embed the changed file")
	}

	beforeThird := emb.callCount
	if err := idx.IndexDirectory(ctx, dir, true); err != nil {
		t.Fatalf("third run: %v", err)
	}
	if emb.callCount != beforeThird {
		t.Errorf("third run re-embedded %d batches; both files are unchanged since the second run, want 0", emb.callCount-beforeThird)
	}
}

// A deleted or changed file must not leave its old chunks behind. Otherwise a
// search would return text the file no longer contains.
func TestIndexDirectory_IncrementalRemovesChunksOfDeletedAndChangedFiles(t *testing.T) {
	emb := &mockEmbedderForIndexer{dimension: 3}
	idx, vs, _ := newPipelineIndexer(t, emb)
	dir := t.TempDir()
	gone := filepath.Join(dir, "gone.md")
	edited := filepath.Join(dir, "edited.md")
	writeFile(t, gone, "# Gone\n\nRETIRED-TEXT appears only in this file.\n")
	writeFile(t, edited, "# Edited\n\nOLD-TEXT appears before the edit.\n")
	ctx := context.Background()

	if err := idx.IndexDirectory(ctx, dir, false); err != nil {
		t.Fatalf("first run: %v", err)
	}

	if err := os.Remove(gone); err != nil {
		t.Fatal(err)
	}
	writeFile(t, edited, "# Edited\n\nNEW-TEXT replaces it.\n")
	if err := idx.IndexDirectory(ctx, dir, true); err != nil {
		t.Fatalf("incremental run: %v", err)
	}

	for _, r := range vs.data {
		if strings.Contains(r.Content, "RETIRED-TEXT") {
			t.Errorf("chunk of the deleted file survived: %q", r.Content)
		}
		if strings.Contains(r.Content, "OLD-TEXT") {
			t.Errorf("chunk of the file before its edit survived: %q", r.Content)
		}
	}
	chunkContaining(t, vs, "NEW-TEXT")
}

// One document the model refuses must not stop the others. The rest are indexed,
// the failure is returned, and the failed document gets no hash.
func TestIndexDirectory_FailedDocumentDoesNotStopTheRest(t *testing.T) {
	emb := &failingEmbedder{dimension: 3}
	idx, vs, _ := newPipelineIndexer(t, emb)
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "good.md"), "# Good\n\nThis one indexes fine.\n")
	writeFile(t, filepath.Join(dir, "bad.md"), "# Bad\n\nThis one has BOOM in it.\n")

	// The good document is indexed, and the run still reports the bad one. A
	// silent success here would let analyze mark a partial index as fresh.
	err := idx.IndexDirectory(context.Background(), dir, false)
	if err == nil || !strings.Contains(err.Error(), "bad.md") {
		t.Fatalf("IndexDirectory returned %v; want an error naming bad.md", err)
	}

	for _, r := range vs.data {
		if strings.Contains(r.Content, "BOOM") {
			t.Errorf("the failed document was stored: %q", r.Content)
		}
	}
	chunkContaining(t, vs, "This one indexes fine")

	p := idx.GetProgress()
	goodPath := filepath.Join(dir, "good.md")
	badPath := filepath.Join(dir, "bad.md")
	if _, ok := p.FileHashes[goodPath]; !ok {
		t.Error("good.md has no hash in progress")
	}
	if _, ok := p.FileHashes[badPath]; ok {
		t.Error("bad.md got a hash although it failed; the next incremental run would skip it")
	}
}

// Progress is written to disk with Completed set, and a fresh indexer on the
// same path reads back the hashes.
func TestIndexDirectory_ProgressIsPersisted(t *testing.T) {
	idx, _, _ := newPipelineIndexer(t, &mockEmbedderForIndexer{dimension: 3})
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.md"), guideMD)
	if err := idx.IndexDirectory(context.Background(), dir, false); err != nil {
		t.Fatalf("IndexDirectory: %v", err)
	}

	reloaded := &Indexer{progressPath: idx.progressPath}
	if err := reloaded.loadProgress(); err != nil {
		t.Fatalf("loadProgress: %v", err)
	}
	p := reloaded.GetProgress()
	if !p.Completed {
		t.Error("saved progress is not marked Completed")
	}
	if _, ok := p.FileHashes[filepath.Join(dir, "a.md")]; !ok {
		t.Error("saved progress lost the file hash")
	}
}

// An empty directory indexes nothing and is not an error.
func TestIndexDirectory_EmptyDirectory(t *testing.T) {
	emb := &mockEmbedderForIndexer{dimension: 3}
	idx, vs, _ := newPipelineIndexer(t, emb)
	if err := idx.IndexDirectory(context.Background(), t.TempDir(), false); err != nil {
		t.Fatalf("IndexDirectory on an empty directory: %v", err)
	}
	if emb.callCount != 0 || len(vs.data) != 0 {
		t.Errorf("embedded %d batches and stored %d vectors for an empty directory", emb.callCount, len(vs.data))
	}
}

// A cancelled context must not hang the pipeline. Whether the call returns an
// error or nil depends on where the cancel is seen, so only the return matters.
func TestIndexDirectory_CancelledContextReturns(t *testing.T) {
	idx, _, _ := newPipelineIndexer(t, &mockEmbedderForIndexer{dimension: 3})
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.md"), guideMD)
	writeFile(t, filepath.Join(dir, "b.md"), "# B\n\nAnother file.\n")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan error, 1)
	go func() { done <- idx.IndexDirectory(ctx, dir, false) }()
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("IndexDirectory did not return after the context was cancelled")
	}
}

// saveProgress reports a write error instead of hiding it.
func TestSaveProgress_ErrorWhenDirectoryMissing(t *testing.T) {
	idx, _, _ := newPipelineIndexer(t, &mockEmbedderForIndexer{dimension: 3})
	idx.progressPath = filepath.Join(t.TempDir(), "no", "such", "dir", "progress.json")
	if err := idx.saveProgress(); err == nil {
		t.Error("saveProgress returned nil for an unwritable path")
	}
}

// loadProgress rejects a checkpoint file that is not valid JSON.
func TestLoadProgress_RejectsCorruptFile(t *testing.T) {
	idx, _, _ := newPipelineIndexer(t, &mockEmbedderForIndexer{dimension: 3})
	writeFile(t, idx.progressPath, "{not json")
	if err := idx.loadProgress(); err == nil {
		t.Error("loadProgress accepted corrupt JSON")
	}
}

// A missing checkpoint is a fresh start, not an error.
func TestLoadProgress_MissingFileIsFreshStart(t *testing.T) {
	idx, _, _ := newPipelineIndexer(t, &mockEmbedderForIndexer{dimension: 3})
	if err := idx.loadProgress(); err != nil {
		t.Errorf("loadProgress with no checkpoint: %v", err)
	}
}

// Close reports an error from a component instead of dropping it.
func TestClose_ReportsComponentError(t *testing.T) {
	idx, _, _ := newPipelineIndexer(t, &mockEmbedderForIndexer{dimension: 3})
	idx.textIndex = &closeErrTextIndex{mockTextIndexForIndexer{data: map[string]textindex.SearchResult{}}}
	err := idx.Close()
	if err == nil || !strings.Contains(err.Error(), "bleve busy") {
		t.Errorf("Close error = %v, want the text index error", err)
	}
}

// Close on healthy components returns nil.
func TestClose_HealthyComponents(t *testing.T) {
	idx, _, _ := newPipelineIndexer(t, &mockEmbedderForIndexer{dimension: 3})
	if err := idx.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}

// GetProgress returns a copy, so a caller cannot change the indexer's state.
func TestGetProgress_ReturnsCopy(t *testing.T) {
	idx, _, _ := newPipelineIndexer(t, &mockEmbedderForIndexer{dimension: 3})
	idx.progress.FileHashes["x"] = "1"
	p := idx.GetProgress()
	p.FileHashes["x"] = "changed"
	if idx.progress.FileHashes["x"] != "1" {
		t.Error("GetProgress shares the hash map with the indexer")
	}
}

// NewIndexer rejects an unknown embedder type before it touches any store.
func TestNewIndexer_RejectsUnknownEmbedder(t *testing.T) {
	cfg := &config.SearchConfig{Embedder: config.EmbedderConfig{Type: "no-such-embedder"}}
	if _, err := NewIndexer(cfg, t.TempDir()); err == nil {
		t.Error("NewIndexer accepted an unknown embedder type")
	}
}

// The supported extensions include Markdown, the format the pipeline is built for.
func TestGetSupportedExtensions_IncludesMarkdown(t *testing.T) {
	idx, _, _ := newPipelineIndexer(t, &mockEmbedderForIndexer{dimension: 3})
	found := false
	for _, ext := range idx.getSupportedExtensions() {
		if ext == ".md" {
			found = true
		}
	}
	if !found {
		t.Errorf("supported extensions %v do not include .md", idx.getSupportedExtensions())
	}
}

// notesPath writes content to a real file, because processDocument reads the
// file from disk when it stores the raw document.
func notesPath(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "notes.md")
	writeFile(t, path, content)
	return path
}

// fakeTokenizer counts words, like wordCounter, but through embedder.Tokenizer.
type fakeTokenizer struct{}

func (fakeTokenizer) CountTokens(text string) int { return len(strings.Fields(text)) }

// The chunker counter adapts the embedder's tokenizer.
func TestTokenizerCounter_Count(t *testing.T) {
	c := tokenizerCounter{t: fakeTokenizer{}}
	if got := c.Count("one two three"); got != 3 {
		t.Errorf("Count = %d, want 3", got)
	}
}

// assembleIndexer loads saved progress when resuming, so an interrupted run
// continues from its checkpoint.
func TestAssembleIndexer_LoadsProgressWhenResuming(t *testing.T) {
	base := t.TempDir()
	saved := Progress{FileHashes: map[string]string{"/docs/a.md": "h-a"}, Completed: true}
	data, err := json.Marshal(saved)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(base, "indexer_progress.json"), string(data))

	idx := assembleIndexer(&config.SearchConfig{}, base, nil, nil, nil, nil, nil, nil)
	p := idx.GetProgress()
	if p.FileHashes["/docs/a.md"] != "h-a" || !p.Completed {
		t.Errorf("progress not loaded on resume: %+v", p)
	}
	if idx.progressPath != filepath.Join(base, "indexer_progress.json") {
		t.Errorf("progressPath = %q", idx.progressPath)
	}
}

// A corrupt checkpoint is not fatal: the indexer starts fresh.
func TestAssembleIndexer_CorruptProgressStartsFresh(t *testing.T) {
	base := t.TempDir()
	writeFile(t, filepath.Join(base, "indexer_progress.json"), "{broken")
	idx := assembleIndexer(&config.SearchConfig{}, base, nil, nil, nil, nil, nil, nil)
	if p := idx.GetProgress(); len(p.FileHashes) != 0 || p.Completed {
		t.Errorf("expected a fresh start, got %+v", p)
	}
}

// failingChunker makes the chunking step fail.
type failingChunker struct{}

func (failingChunker) Chunk(ctx context.Context, doc parser.Document) ([]chunker.Chunk, error) {
	return nil, errors.New("chunker broke")
}
func (failingChunker) Name() string           { return "failing" }
func (failingChunker) Config() chunker.Config { return chunker.Config{} }

// addErrVectorStore fails Add and Close.
type addErrVectorStore struct {
	mockVectorStoreForIndexer
}

func (a *addErrVectorStore) Add(ctx context.Context, ids []string, vectors [][]float32, metadatas []map[string]string, documents []string) error {
	return errors.New("vector write failed")
}
func (a *addErrVectorStore) Close() error { return errors.New("chroma busy") }

// addErrTextIndex fails Add.
type addErrTextIndex struct {
	mockTextIndexForIndexer
}

func (a *addErrTextIndex) Add(ctx context.Context, ids []string, documents []string, metadatas []map[string]string) error {
	return errors.New("text write failed")
}

// Each step of processDocument reports its own error, so a failure is never
// mistaken for a success.
func TestProcessDocument_ReportsEachStepFailure(t *testing.T) {
	ctx := context.Background()
	good := notesPath(t, "# Notes\n\nA paragraph with enough words to become a chunk.")
	doc := parser.Document{ID: "doc00001", Path: good, Content: "# Notes\n\nSome words here.", Hash: "h"}

	t.Run("missing file", func(t *testing.T) {
		idx, _, _ := newPipelineIndexer(t, &mockEmbedderForIndexer{dimension: 3})
		_, err := idx.processDocument(ctx, parser.Document{ID: "doc-gone00", Path: filepath.Join(t.TempDir(), "gone.md")})
		if err == nil || !strings.Contains(err.Error(), "store raw") {
			t.Errorf("err = %v, want a store-raw error", err)
		}
	})
	t.Run("chunker fails", func(t *testing.T) {
		idx, _, _ := newPipelineIndexer(t, &mockEmbedderForIndexer{dimension: 3})
		idx.chunker = failingChunker{}
		if _, err := idx.processDocument(ctx, doc); err == nil || !strings.Contains(err.Error(), "chunk") {
			t.Errorf("err = %v, want a chunk error", err)
		}
	})
	t.Run("embedder fails", func(t *testing.T) {
		idx, _, _ := newPipelineIndexer(t, &failingEmbedder{dimension: 3})
		content := "# BOOM\n\nthis passage has BOOM in it."
		d := parser.Document{ID: "doc00002", Path: notesPath(t, content), Content: content, Hash: "h"}
		if _, err := idx.processDocument(ctx, d); err == nil || !strings.Contains(err.Error(), "embed") {
			t.Errorf("err = %v, want an embed error", err)
		}
	})
	t.Run("vector store fails", func(t *testing.T) {
		idx, _, _ := newPipelineIndexer(t, &mockEmbedderForIndexer{dimension: 3})
		idx.vectorStore = &addErrVectorStore{}
		if _, err := idx.processDocument(ctx, doc); err == nil || !strings.Contains(err.Error(), "vector store") {
			t.Errorf("err = %v, want a vector-store error", err)
		}
	})
	t.Run("text index fails", func(t *testing.T) {
		idx, _, _ := newPipelineIndexer(t, &mockEmbedderForIndexer{dimension: 3})
		idx.textIndex = &addErrTextIndex{}
		if _, err := idx.processDocument(ctx, doc); err == nil || !strings.Contains(err.Error(), "text index") {
			t.Errorf("err = %v, want a text-index error", err)
		}
	})
}

// Close reports a vector store error too, not only the text index error.
func TestClose_ReportsVectorStoreError(t *testing.T) {
	idx, _, _ := newPipelineIndexer(t, &mockEmbedderForIndexer{dimension: 3})
	idx.vectorStore = &addErrVectorStore{}
	err := idx.Close()
	if err == nil || !strings.Contains(err.Error(), "chroma busy") {
		t.Errorf("Close error = %v, want the vector store error", err)
	}
}

// A checkpoint path that is a directory is a read error, which is reported and
// is not the same as a missing file.
func TestLoadProgress_ReadErrorIsReported(t *testing.T) {
	idx, _, _ := newPipelineIndexer(t, &mockEmbedderForIndexer{dimension: 3})
	idx.progressPath = t.TempDir()
	if err := idx.loadProgress(); err == nil {
		t.Error("loadProgress returned nil when the checkpoint path is a directory")
	}
}

// A root that does not exist fails at the scan, and IndexDirectory says so.
func TestIndexDirectory_MissingRootIsAnError(t *testing.T) {
	idx, _, _ := newPipelineIndexer(t, &mockEmbedderForIndexer{dimension: 3})
	if err := idx.IndexDirectory(context.Background(), filepath.Join(t.TempDir(), "absent"), false); err == nil {
		t.Error("IndexDirectory on a missing root returned nil")
	}
}

// tokenEmbedder is a fake embedder that can also count tokens, so it passes the
// tokenizer check in newIndexerWith.
type tokenEmbedder struct {
	mockEmbedderForIndexer
}

func (tokenEmbedder) CountTokens(text string) int { return len(strings.Fields(text)) }

// fakeFactories returns factories that hand back the given fakes, or an error
// for the component named in failAt.
func fakeFactories(emb embedder.Embedder, failAt string) componentFactories {
	return componentFactories{
		embedder: func(*config.EmbedderConfig) (embedder.Embedder, error) {
			if failAt == "embedder" {
				return nil, errors.New("no model")
			}
			return emb, nil
		},
		vectorStore: func(vectorstore.Config, string) (vectorstore.VectorStore, error) {
			if failAt == "vectorStore" {
				return nil, errors.New("no chromem")
			}
			return &mockVectorStoreForIndexer{data: map[string]vectorstore.SearchResult{}}, nil
		},
		textIndex: func(textindex.Config, string) (textindex.TextIndex, error) {
			if failAt == "textIndex" {
				return nil, errors.New("no bleve")
			}
			return &mockTextIndexForIndexer{data: map[string]textindex.SearchResult{}}, nil
		},
		docStore: func(base string) (*store.DocumentStore, error) {
			if failAt == "docStore" {
				return nil, errors.New("no store")
			}
			return store.NewDocumentStore(base)
		},
	}
}

// newIndexerWith wires every component when all factories succeed.
func TestNewIndexerWith_WiresEveryComponent(t *testing.T) {
	emb := &tokenEmbedder{mockEmbedderForIndexer{dimension: 3}}
	cfg := &config.SearchConfig{}
	idx, err := newIndexerWith(cfg, t.TempDir(), fakeFactories(emb, ""))
	if err != nil {
		t.Fatalf("newIndexerWith: %v", err)
	}
	if idx.emb != emb || idx.vectorStore == nil || idx.textIndex == nil || idx.docStore == nil || idx.chunker == nil {
		t.Errorf("not every component was wired: %+v", idx)
	}
	if idx.searchConfig != cfg {
		t.Error("search config was not kept")
	}
}

// Each component failure is reported with the step that failed.
func TestNewIndexerWith_ReportsEachFailure(t *testing.T) {
	cases := []struct {
		name   string
		emb    embedder.Embedder
		failAt string
		want   string
	}{
		{"embedder", &tokenEmbedder{}, "embedder", "create embedder"},
		{"no tokenizer", &mockEmbedderForIndexer{dimension: 3}, "", "cannot count tokens"},
		{"vector store", &tokenEmbedder{}, "vectorStore", "create vector store"},
		{"text index", &tokenEmbedder{}, "textIndex", "create text index"},
		{"document store", &tokenEmbedder{}, "docStore", "create document store"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := newIndexerWith(&config.SearchConfig{}, t.TempDir(), fakeFactories(c.emb, c.failAt))
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("err = %v, want it to mention %q", err, c.want)
			}
		})
	}
}

// The real factories, given a temp directory, build a working vector store, text
// index and document store without loading the embedding model.
func TestRealComponents_BuildStoresWithoutModel(t *testing.T) {
	f := realComponents()
	base := t.TempDir()

	vs, err := f.vectorStore(vectorstore.Config{Type: "chromem"}, base)
	if err != nil || vs == nil {
		t.Fatalf("vectorStore: %v", err)
	}
	ti, err := f.textIndex(textindex.Config{Type: "bleve"}, base)
	if err != nil || ti == nil {
		t.Fatalf("textIndex: %v", err)
	}
	t.Cleanup(func() { _ = ti.Close() })
	ds, err := f.docStore(base)
	if err != nil || ds == nil {
		t.Fatalf("docStore: %v", err)
	}
	_ = vs.Close()
}
