package eval

import (
	"fmt"
	"path/filepath"

	"github.com/plaesy/spec-kit/internal/config"
	"github.com/plaesy/spec-kit/internal/search/embedder"
	"github.com/plaesy/spec-kit/internal/search/retriever"
	"github.com/plaesy/spec-kit/internal/search/store"
	"github.com/plaesy/spec-kit/internal/search/textindex"
	"github.com/plaesy/spec-kit/internal/search/vectorstore"
)

// OpenIndexEngine opens an index built by `plaesy analyze` and returns
// an engine over it. It builds the same components `plaesy search`
// builds, so the evaluation measures the production pipeline.
//
// The returned closer releases the embedder and both indexes.
func OpenIndexEngine(projectRoot, indexDir string) (*IndexEngine, func(), error) {
	root, err := filepath.Abs(projectRoot)
	if err != nil {
		return nil, nil, fmt.Errorf("resolving project root: %w", err)
	}

	unified, err := config.NewConfigManager[config.UnifiedConfig]("", config.DefaultUnifiedConfig()).Load()
	if err != nil {
		return nil, nil, fmt.Errorf("loading unified config: %w", err)
	}
	if unified.Search == nil {
		unified.Search = config.DefaultSearchConfig()
	}
	searchCfg := unified.Search

	emb, err := embedder.NewEmbedder(&searchCfg.Embedder)
	if err != nil {
		return nil, nil, fmt.Errorf("creating embedder: %w", err)
	}

	vs, err := vectorstore.NewChromemStore(vectorstore.Config{Type: searchCfg.VectorStore.Type}, indexDir)
	if err != nil {
		closeEmbedder(emb)
		return nil, nil, fmt.Errorf("opening vector store in %s: %w", indexDir, err)
	}

	ti, err := textindex.NewBleveIndex(textindex.Config{Type: searchCfg.TextIndex.Type}, indexDir)
	if err != nil {
		_ = vs.Close()
		closeEmbedder(emb)
		return nil, nil, fmt.Errorf("opening text index in %s: %w", indexDir, err)
	}

	docs, err := store.NewDocumentStore(indexDir)
	if err != nil {
		_ = ti.Close()
		_ = vs.Close()
		closeEmbedder(emb)
		return nil, nil, fmt.Errorf("opening document store in %s: %w", indexDir, err)
	}

	hybrid := retriever.NewHybridRetriever(searchCfg.Retriever, emb, vs, ti, docs)

	engine := NewIndexEngine(root, emb, vs, ti, hybrid, docs)
	closer := func() {
		_ = ti.Close()
		_ = vs.Close()
		closeEmbedder(emb)
	}
	return engine, closer, nil
}

// closeEmbedder releases embedders that hold resources, such as a loaded model.
func closeEmbedder(emb embedder.Embedder) {
	if c, ok := emb.(interface{ Close() error }); ok {
		_ = c.Close()
	}
}
