// Package docindex builds and queries the document index that plaesy analyze
// writes under the analysis directory. It covers Markdown, HTML and JSON files
// in the project. The index lives in <analysisDir>/docs and is refused by
// `plaesy search` when a document has changed since it was built.
package docindex

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/plaesy/spec-kit/internal/config"
	"github.com/plaesy/spec-kit/internal/search/embedder"
	"github.com/plaesy/spec-kit/internal/search/indexer"
	"github.com/plaesy/spec-kit/internal/search/retriever"
	"github.com/plaesy/spec-kit/internal/search/store"
	"github.com/plaesy/spec-kit/internal/search/textindex"
	"github.com/plaesy/spec-kit/internal/search/vectorstore"
	"github.com/plaesy/spec-kit/internal/exclude"
)

// docExts are the extensions of files the document index covers.
var docExts = map[string]bool{
	".md": true, ".markdown": true, ".mdx": true,
	".html": true, ".htm": true,
	".json": true,
}

// IsDocExt reports whether a file with this extension is part of the document index.
func IsDocExt(path string) bool {
	return docExts[strings.ToLower(filepath.Ext(path))]
}

const (
	dirName         = "docs"
	fingerprintName = "fingerprint"
)

// The constructors Build and QueryWith call. Tests replace them with fakes so the
// embedding model is never loaded; production always uses the defaults below.
var (
	newEmbedder = embedder.NewEmbedder
	newIndexer  = func(cfg *config.SearchConfig, dir string) (docIndexer, error) {
		idx, err := indexer.NewIndexer(cfg, dir)
		if err != nil {
			return nil, err
		}
		return idx, nil
	}
	newVectorStore = func(cfg vectorstore.Config, dir string) (vectorstore.VectorStore, error) {
		vs, err := vectorstore.NewChromemStore(cfg, dir)
		if err != nil {
			return nil, err
		}
		return vs, nil
	}
	newTextIndex = func(cfg textindex.Config, dir string) (textindex.TextIndex, error) {
		ti, err := textindex.NewBleveIndex(cfg, dir)
		if err != nil {
			return nil, err
		}
		return ti, nil
	}
	newDocStore = func(dir string) (retriever.DocumentStoreInterface, error) {
		ds, err := store.NewDocumentStore(dir)
		if err != nil {
			return nil, err
		}
		return ds, nil
	}
)

// docIndexer is the part of the indexer that Build uses.
type docIndexer interface {
	IndexDirectory(ctx context.Context, rootPath string, incremental bool) error
	NeedsFullRebuild() bool
	Close() error
}

// Dir returns the directory the document index is written to.
func Dir(analysisDir string) string {
	return filepath.Join(analysisDir, dirName)
}

// Collect lists the document files under root as slash-separated paths relative
// to root, sorted. Directories the exclude package leaves out for indexing are
// not entered; see exclude.SkipIndexDir.
func Collect(root string) ([]string, error) {
	var out []string
	ig := exclude.LoadGitignore(root)
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == root {
			return nil
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if ig.SkipIndexDir(rel, d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !IsDocExt(p) || ig.SkipIndexFile(rel) {
			return nil
		}
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(out)
	return out, nil
}

// Fingerprint identifies the content of every document under root. It hashes
// each file's relative path together with its content hash, so an edit, an
// added file, or a deleted file all change it. Editing a document does not
// touch the symbol index's fingerprint, which is why this one exists.
func Fingerprint(root string) (string, error) {
	files, err := Collect(root)
	if err != nil {
		return "", fmt.Errorf("listing documents: %w", err)
	}
	h := sha256.New()
	for _, rel := range files {
		b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			return "", fmt.Errorf("reading %s: %w", rel, err)
		}
		content := sha256.Sum256(b)
		fmt.Fprintf(h, "%s\x00%s\n", rel, hex.EncodeToString(content[:]))
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// LoadConfig returns the search configuration, using defaults when the unified
// config does not set one.
func LoadConfig() (*config.SearchConfig, error) {
	mgr := config.NewConfigManager[config.UnifiedConfig]("", config.DefaultUnifiedConfig())
	unified, err := mgr.Load()
	if err != nil {
		return nil, fmt.Errorf("loading unified config: %w", err)
	}
	if unified.Search == nil {
		return config.DefaultSearchConfig(), nil
	}
	return unified.Search, nil
}

// buildLockStale is how old a build lock must be before it is taken as left
// behind by a build that died without cleaning up.
const buildLockStale = 2 * time.Hour

// Build brings the document index for the project at root up to date.
//
// With force, the index is cleared and every document is embedded again. Without
// it, only documents whose content changed since the last build are embedded, and
// the chunks of changed or deleted documents are removed. An index saved by an
// older build, which has no per-file chunk records, is rebuilt in full, because
// its stale chunks cannot be found.
//
// The fingerprint is written last, so a failed build leaves the previous claim
// standing and the next `plaesy search` refuses the index. A lock stops two builds
// from writing the same index at once.
func Build(ctx context.Context, root, analysisDir string, cfg *config.SearchConfig, force bool) error {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	unlock, err := acquireBuildLock(analysisDir)
	if err != nil {
		return err
	}
	defer unlock()

	dir := Dir(analysisDir)
	if force {
		if err := clearIndexDir(dir); err != nil {
			return err
		}
	}

	idx, err := openIndexer(cfg, dir)
	if err != nil && !force {
		// An index that cannot be opened cannot be updated either. Rebuild it
		// rather than leave the project without a document index.
		fmt.Printf("Document index cannot be opened (%v); rebuilding it from scratch\n", err)
		force = true
		if err := clearIndexDir(dir); err != nil {
			return err
		}
		idx, err = openIndexer(cfg, dir)
	}
	if err != nil {
		return err
	}
	if !force && idx.NeedsFullRebuild() {
		fmt.Println("Document index predates chunk tracking; rebuilding it from scratch")
		idx.Close()
		force = true
		if err := clearIndexDir(dir); err != nil {
			return err
		}
		if idx, err = openIndexer(cfg, dir); err != nil {
			return err
		}
	}
	defer idx.Close()

	if err := idx.IndexDirectory(ctx, absRoot, !force); err != nil {
		return fmt.Errorf("indexing documents: %w", err)
	}

	fp, err := Fingerprint(absRoot)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, fingerprintName), []byte(fp+"\n"), 0o644)
}

// clearIndexDir removes the index directory so the next build starts empty. The
// build lock sits beside this directory, not inside it, so clearing it is safe.
func clearIndexDir(dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("clearing document index: %w", err)
	}
	return nil
}

// openIndexer creates the index directory if needed and opens an indexer over
// it. An existing index is opened as it is, so an incremental build can update it.
func openIndexer(cfg *config.SearchConfig, dir string) (docIndexer, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("creating document index dir: %w", err)
	}
	idx, err := newIndexer(cfg, dir)
	if err != nil {
		return nil, fmt.Errorf("creating indexer: %w", err)
	}
	return idx, nil
}

// acquireBuildLock takes the lock that keeps two builds from writing the same
// index. A second build fails at once, instead of deleting or overwriting files
// the first is still using. A lock older than buildLockStale is reclaimed.
func acquireBuildLock(analysisDir string) (func(), error) {
	if err := os.MkdirAll(analysisDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating analysis dir: %w", err)
	}
	path := filepath.Join(analysisDir, dirName+".lock")
	create := func() error {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		return f.Close()
	}

	err := create()
	if os.IsExist(err) {
		if info, statErr := os.Stat(path); statErr == nil && time.Since(info.ModTime()) < buildLockStale {
			return nil, fmt.Errorf("another document index build is running (lock %s); wait for it to finish", path)
		}
		_ = os.Remove(path)
		err = create()
	}
	if err != nil {
		return nil, fmt.Errorf("taking build lock: %w", err)
	}
	return func() { _ = os.Remove(path) }, nil
}

// IsStale reports whether the document index is missing, or was built from
// documents that have since changed. A missing index is stale.
func IsStale(root, analysisDir string) (bool, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return true, err
	}
	stored, err := os.ReadFile(filepath.Join(Dir(analysisDir), fingerprintName))
	if os.IsNotExist(err) {
		return true, nil
	}
	if err != nil {
		return true, fmt.Errorf("reading document index fingerprint: %w", err)
	}
	current, err := Fingerprint(absRoot)
	if err != nil {
		return true, err
	}
	return strings.TrimSpace(string(stored)) != current, nil
}

// Hit is one document passage returned by Query.
type Hit struct {
	Path      string  // slash-separated, relative to the project root
	Section   string  // heading path of the passage, "" when it has none
	Snippet   string  // start of the passage text
	Score     float64 // hybrid retrieval score
	StartLine int     // 1-based source line of the passage, 0 when unknown
	EndLine   int     // 1-based, end-inclusive; 0 when unknown
}

// Query returns up to top passages from the document index, best first. It
// does not check staleness; callers do that with IsStale.
func Query(ctx context.Context, root, analysisDir, text string, top int, cfg *config.SearchConfig) ([]Hit, error) {
	return QueryWith(ctx, root, analysisDir, text, top, cfg, retriever.QueryOptions{}, false)
}

// QueryWith is Query with predicate filters (opts.Where) and, when parentChild
// is true, the parent-section view: each hit is a whole section, with its
// matched children folded into it. The section's Snippet is its start text.
func QueryWith(ctx context.Context, root, analysisDir, text string, top int, cfg *config.SearchConfig, opts retriever.QueryOptions, parentChild bool) ([]Hit, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	dir := Dir(analysisDir)

	emb, err := newEmbedder(&cfg.Embedder)
	if err != nil {
		return nil, fmt.Errorf("creating embedder: %w", err)
	}
	vs, err := newVectorStore(vectorstore.Config{Type: cfg.VectorStore.Type}, dir)
	if err != nil {
		return nil, fmt.Errorf("opening vector store: %w", err)
	}
	defer vs.Close()
	ti, err := newTextIndex(textindex.Config{Type: cfg.TextIndex.Type}, dir)
	if err != nil {
		return nil, fmt.Errorf("opening text index: %w", err)
	}
	defer ti.Close()
	ds, err := newDocStore(dir)
	if err != nil {
		return nil, fmt.Errorf("opening document store: %w", err)
	}

	retr := retriever.NewHybridRetriever(cfg.Retriever, emb, vs, ti, ds)
	if parentChild {
		return parentHits(ctx, retr, absRoot, text, top, opts)
	}
	results, err := retr.RetrieveWith(ctx, text, opts)
	if err != nil {
		return nil, fmt.Errorf("document search: %w", err)
	}
	if top > 0 && len(results) > top {
		results = results[:top]
	}

	hits := make([]Hit, 0, len(results))
	for _, r := range results {
		rel := r.Metadata["path"]
		if p, err := filepath.Rel(absRoot, rel); err == nil {
			rel = filepath.ToSlash(p)
		}
		// The line metadata is written by the indexer. A passage stored
		// without it (an older index) parses to 0, which means unknown.
		startLine, _ := strconv.Atoi(r.Metadata["start_line"])
		endLine, _ := strconv.Atoi(r.Metadata["end_line"])
		// The retriever sets SectionPath only on parent-child expansion. Plain hits
		// carry it in metadata, which the indexer writes as "section_path".
		section := r.SectionPath
		if section == "" {
			section = r.Metadata["section_path"]
		}
		hits = append(hits, Hit{
			Path:      rel,
			Section:   section,
			Snippet:   snippet(r.Content, 160),
			Score:     r.Score,
			StartLine: startLine,
			EndLine:   endLine,
		})
	}
	return hits, nil
}

// parentHits answers with whole sections. A section's line range is the span of
// its chunks, so the hit points at the section, not at one matched child.
func parentHits(ctx context.Context, retr *retriever.HybridRetriever, absRoot, text string, top int, opts retriever.QueryOptions) ([]Hit, error) {
	sections, err := retriever.NewParentChildRetriever(retr).Retrieve(ctx, text, opts)
	if err != nil {
		return nil, fmt.Errorf("document search: %w", err)
	}
	if top > 0 && len(sections) > top {
		sections = sections[:top]
	}
	hits := make([]Hit, 0, len(sections))
	for _, s := range sections {
		rel := s.Path
		if p, err := filepath.Rel(absRoot, rel); err == nil {
			rel = filepath.ToSlash(p)
		}
		hits = append(hits, Hit{
			Path:      rel,
			Section:   s.SectionPath,
			Snippet:   snippet(s.Content, 160),
			Score:     s.Score,
			StartLine: s.StartLine,
			EndLine:   s.EndLine,
		})
	}
	return hits, nil
}

// snippet returns the first max bytes of s, cut back to a rune boundary.
func snippet(s string, max int) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) <= max {
		return s
	}
	cut := max
	for cut > 0 && s[cut]&0xC0 == 0x80 {
		cut--
	}
	return s[:cut] + "..."
}
