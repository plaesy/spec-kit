package eval

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/plaesy/spec-kit/internal/search/retriever"
	"github.com/plaesy/spec-kit/internal/search/store"
	"github.com/plaesy/spec-kit/internal/search/textindex"
	"github.com/plaesy/spec-kit/internal/search/vectorstore"
)

// ErrModeUnsupported is returned for modes the search pipeline does not have.
// The harness records it as a per-query error instead of aborting the run.
var ErrModeUnsupported = errors.New("search mode not implemented in the pipeline")

// Narrow interfaces so the adapter can be exercised without a real index.
// The concrete types in the search packages satisfy them.

type queryEmbedder interface {
	EmbedQuery(ctx context.Context, query string) ([]float32, error)
}

type vectorSearcher interface {
	Search(ctx context.Context, queryVector []float32, k int, filter map[string]string) ([]vectorstore.SearchResult, error)
}

type textSearcher interface {
	Search(ctx context.Context, query string, k int, filter map[string]string) ([]textindex.SearchResult, error)
}

type hybridSearcher interface {
	Retrieve(ctx context.Context, query string, filters map[string]string) ([]retriever.Result, error)
}

type rawLookup interface {
	GetRaw(ctx context.Context, id string) (*store.RawDocument, error)
}

// IndexEngine answers evaluation queries with the components that
// `plaesy search` uses for documents. It returns project-relative file paths rather
// than document hashes, because golden sets name files and content hashes
// change on every edit.
//
// Mode mapping:
//   - vector: embed the query, search the vector store only
//   - bm25:   search the text index only
//   - hybrid: the production HybridRetriever (RRF fusion)
//   - other:  ErrModeUnsupported
type IndexEngine struct {
	root   string
	emb    queryEmbedder
	vec    vectorSearcher
	text   textSearcher
	hybrid hybridSearcher
	docs   rawLookup

	mu        sync.Mutex
	pathCache map[string]string
}

// NewIndexEngine wires the adapter. root is the project directory the index
// was built from; paths are reported relative to it.
func NewIndexEngine(
	root string,
	emb queryEmbedder,
	vec vectorSearcher,
	text textSearcher,
	hybrid hybridSearcher,
	docs rawLookup,
) *IndexEngine {
	return &IndexEngine{
		root:      root,
		emb:       emb,
		vec:       vec,
		text:      text,
		hybrid:    hybrid,
		docs:      docs,
		pathCache: make(map[string]string),
	}
}

// Name identifies the engine in harness output.
func (e *IndexEngine) Name() string { return "index" }

// Search implements SearchEngine. It returns one entry per document, in rank
// order, so several matching chunks from the same file count once.
func (e *IndexEngine) Search(query string, topK int, mode SearchMode) ([]string, []float64, error) {
	ctx := context.Background()

	var docIDs []string
	var scores []float64
	switch mode {
	case ModeVector:
		vec, err := e.emb.EmbedQuery(ctx, query)
		if err != nil {
			return nil, nil, fmt.Errorf("embed query: %w", err)
		}
		hits, err := e.vec.Search(ctx, vec, topK, nil)
		if err != nil {
			return nil, nil, fmt.Errorf("vector search: %w", err)
		}
		for _, h := range hits {
			docIDs = append(docIDs, h.Metadata["doc_id"])
			scores = append(scores, float64(h.Score))
		}

	case ModeBM25:
		hits, err := e.text.Search(ctx, query, topK, nil)
		if err != nil {
			return nil, nil, fmt.Errorf("text search: %w", err)
		}
		for _, h := range hits {
			docIDs = append(docIDs, h.Metadata["doc_id"])
			scores = append(scores, h.Score)
		}

	case ModeHybrid:
		hits, err := e.hybrid.Retrieve(ctx, query, nil)
		if err != nil {
			return nil, nil, fmt.Errorf("hybrid search: %w", err)
		}
		for _, h := range hits {
			docIDs = append(docIDs, h.DocID)
			scores = append(scores, h.Score)
		}

	default:
		return nil, nil, fmt.Errorf("%w: %q", ErrModeUnsupported, mode)
	}

	return e.toPaths(ctx, docIDs, scores, topK)
}

// toPaths maps document IDs to project-relative paths, keeping the first
// (best-ranked) occurrence of each file and stopping at topK distinct files.
func (e *IndexEngine) toPaths(ctx context.Context, docIDs []string, scores []float64, topK int) ([]string, []float64, error) {
	seen := make(map[string]bool)
	var paths []string
	var outScores []float64
	for i, id := range docIDs {
		if id == "" {
			return nil, nil, errors.New("search hit has no doc_id metadata")
		}
		p, err := e.pathOf(ctx, id)
		if err != nil {
			return nil, nil, err
		}
		if seen[p] {
			continue
		}
		seen[p] = true
		paths = append(paths, p)
		outScores = append(outScores, scores[i])
		if topK > 0 && len(paths) >= topK {
			break
		}
	}
	return paths, outScores, nil
}

// pathOf resolves a document ID to its project-relative, slash-separated path.
func (e *IndexEngine) pathOf(ctx context.Context, docID string) (string, error) {
	e.mu.Lock()
	if p, ok := e.pathCache[docID]; ok {
		e.mu.Unlock()
		return p, nil
	}
	e.mu.Unlock()

	raw, err := e.docs.GetRaw(ctx, docID)
	if err != nil {
		return "", fmt.Errorf("resolve doc %s: %w", docID, err)
	}

	p := raw.Path
	if filepath.IsAbs(p) {
		rel, err := filepath.Rel(e.root, p)
		if err != nil {
			return "", fmt.Errorf("make %s relative to %s: %w", p, e.root, err)
		}
		p = rel
	}
	p = filepath.ToSlash(filepath.Clean(p))

	e.mu.Lock()
	e.pathCache[docID] = p
	e.mu.Unlock()
	return p, nil
}
