package docindex

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/plaesy/spec-kit/internal/config"
	"github.com/plaesy/spec-kit/internal/search/chunker"
	"github.com/plaesy/spec-kit/internal/search/embedder"
	"github.com/plaesy/spec-kit/internal/search/retriever"
	"github.com/plaesy/spec-kit/internal/search/store"
	"github.com/plaesy/spec-kit/internal/search/textindex"
	"github.com/plaesy/spec-kit/internal/search/vectorstore"
)

// The fakes below replace the embedding model and the stores, so no test loads
// the real model or touches the network. Each fake records nothing beyond what a
// test asserts on.

type fakeEmbedder struct{}

func (fakeEmbedder) EmbedQuery(_ context.Context, _ string) ([]float32, error) {
	return []float32{1, 0, 0}, nil
}

func (fakeEmbedder) EmbedPassages(_ context.Context, passages []string) ([][]float32, error) {
	out := make([][]float32, len(passages))
	for i := range out {
		out[i] = []float32{1, 0, 0}
	}
	return out, nil
}

// fakeVectorStore returns its hits in the order given, after the exact metadata
// filter and the k cut, as a real store does.
type fakeVectorStore struct {
	hits []vectorstore.SearchResult
	err  error
}

func (f *fakeVectorStore) Add(context.Context, []string, [][]float32, []map[string]string, []string) error {
	return nil
}

func (f *fakeVectorStore) Search(_ context.Context, _ []float32, k int, filter map[string]string) ([]vectorstore.SearchResult, error) {
	if f.err != nil {
		return nil, f.err
	}
	var out []vectorstore.SearchResult
	for _, h := range f.hits {
		if matchesFilter(h.Metadata, filter) {
			out = append(out, h)
		}
	}
	if k > 0 && len(out) > k {
		out = out[:k]
	}
	return out, nil
}

func (f *fakeVectorStore) Delete(context.Context, []string) error { return nil }
func (f *fakeVectorStore) Count() (int, error)                    { return len(f.hits), nil }
func (f *fakeVectorStore) Close() error                           { return nil }

// fakeTextIndex returns its hits in the order given.
type fakeTextIndex struct {
	hits []textindex.SearchResult
}

func (f *fakeTextIndex) Add(context.Context, []string, []string, []map[string]string) error {
	return nil
}

func (f *fakeTextIndex) Search(_ context.Context, _ string, k int, filter map[string]string) ([]textindex.SearchResult, error) {
	var out []textindex.SearchResult
	for _, h := range f.hits {
		if matchesFilter(h.Metadata, filter) {
			out = append(out, h)
		}
	}
	if k > 0 && len(out) > k {
		out = out[:k]
	}
	return out, nil
}

func (f *fakeTextIndex) Delete(context.Context, []string) error { return nil }
func (f *fakeTextIndex) Count() (int, error)                    { return len(f.hits), nil }
func (f *fakeTextIndex) Close() error                           { return nil }

// fakeDocStore serves processed documents by ID.
type fakeDocStore struct {
	docs map[string]*store.ProcessedDocument
}

func (f *fakeDocStore) GetProcessed(_ context.Context, id string) (*store.ProcessedDocument, error) {
	return f.docs[id], nil
}

func matchesFilter(meta, filter map[string]string) bool {
	for k, v := range filter {
		if meta[k] != v {
			return false
		}
	}
	return true
}

// fakeStack is everything QueryWith builds, with an optional open error per part.
type fakeStack struct {
	vs     *fakeVectorStore
	ti     *fakeTextIndex
	ds     *fakeDocStore
	embErr error
	vsErr  error
	tiErr  error
	dsErr  error
}

// installSeams points the package constructors at the fakes for one test. The
// originals come back when the test ends.
func installSeams(t *testing.T, s *fakeStack) {
	t.Helper()
	prevEmb, prevVS, prevTI, prevDS := newEmbedder, newVectorStore, newTextIndex, newDocStore
	t.Cleanup(func() {
		newEmbedder, newVectorStore, newTextIndex, newDocStore = prevEmb, prevVS, prevTI, prevDS
	})
	newEmbedder = func(*config.EmbedderConfig) (embedder.Embedder, error) {
		if s.embErr != nil {
			return nil, s.embErr
		}
		return fakeEmbedder{}, nil
	}
	newVectorStore = func(vectorstore.Config, string) (vectorstore.VectorStore, error) {
		if s.vsErr != nil {
			return nil, s.vsErr
		}
		if s.vs == nil {
			s.vs = &fakeVectorStore{}
		}
		return s.vs, nil
	}
	newTextIndex = func(textindex.Config, string) (textindex.TextIndex, error) {
		if s.tiErr != nil {
			return nil, s.tiErr
		}
		if s.ti == nil {
			s.ti = &fakeTextIndex{}
		}
		return s.ti, nil
	}
	newDocStore = func(string) (retriever.DocumentStoreInterface, error) {
		if s.dsErr != nil {
			return nil, s.dsErr
		}
		if s.ds == nil {
			s.ds = &fakeDocStore{docs: map[string]*store.ProcessedDocument{}}
		}
		return s.ds, nil
	}
}

// fakeIndexer records the roots it was asked to index.
type fakeIndexer struct {
	roots     []string
	err       error
	closed    bool
	needsFull bool // reports a saved index that predates chunk tracking
}

func (f *fakeIndexer) IndexDirectory(_ context.Context, root string, _ bool) error {
	f.roots = append(f.roots, root)
	return f.err
}

func (f *fakeIndexer) NeedsFullRebuild() bool { return f.needsFull }

func (f *fakeIndexer) Close() error {
	f.closed = true
	return nil
}

// installIndexer points newIndexer at idx, or makes it fail with newErr.
func installIndexer(t *testing.T, idx *fakeIndexer, newErr error) {
	t.Helper()
	prev := newIndexer
	t.Cleanup(func() { newIndexer = prev })
	newIndexer = func(*config.SearchConfig, string) (docIndexer, error) {
		if newErr != nil {
			return nil, newErr
		}
		return idx, nil
	}
}

// docMeta is the metadata a document chunk carries in the stores.
func docMeta(path, docID string, extra map[string]string) map[string]string {
	m := map[string]string{"path": path, "doc_id": docID}
	for k, v := range extra {
		m[k] = v
	}
	return m
}

func TestQueryWith_ReturnsPassagesWithLinesAndRelativePaths(t *testing.T) {
	root := t.TempDir()
	guide := filepath.Join(root, "docs", "guide.md")
	page := filepath.Join(root, "site", "index.html")
	s := &fakeStack{vs: &fakeVectorStore{hits: []vectorstore.SearchResult{
		{ID: "g1", Score: 0.9, Content: "Refunds are issued within thirty days.",
			Metadata: docMeta(guide, "d1", map[string]string{"start_line": "5", "end_line": "9", "section_path": "Policy > Refunds"})},
		{ID: "p1", Score: 0.8, Content: "home page",
			Metadata: docMeta(page, "d2", nil)},
	}}}
	installSeams(t, s)

	hits, err := QueryWith(context.Background(), root, t.TempDir(), "refund", 10, config.DefaultSearchConfig(), retriever.QueryOptions{}, false)
	if err != nil {
		t.Fatalf("QueryWith: %v", err)
	}
	if len(hits) != 2 {
		t.Fatalf("got %d hits, want 2: %+v", len(hits), hits)
	}
	md := hits[0]
	if md.Path != "docs/guide.md" || md.StartLine != 5 || md.EndLine != 9 {
		t.Errorf("markdown hit = %+v, want docs/guide.md lines 5-9", md)
	}
	if md.Section != "Policy > Refunds" {
		t.Errorf("section = %q, want the section_path from metadata", md.Section)
	}
	html := hits[1]
	if html.Path != "site/index.html" {
		t.Errorf("html path = %q, want site/index.html", html.Path)
	}
	if html.StartLine != 0 || html.EndLine != 0 {
		t.Errorf("html hit has lines %d-%d, want none (text is extracted, not source lines)", html.StartLine, html.EndLine)
	}
}

func TestQueryWith_TopLimitsResults(t *testing.T) {
	root := t.TempDir()
	var hits []vectorstore.SearchResult
	for _, id := range []string{"a", "b", "c"} {
		hits = append(hits, vectorstore.SearchResult{ID: id, Score: 0.5, Content: id,
			Metadata: docMeta(filepath.Join(root, id+".md"), "d"+id, nil)})
	}
	installSeams(t, &fakeStack{vs: &fakeVectorStore{hits: hits}})

	got, err := QueryWith(context.Background(), root, t.TempDir(), "q", 2, config.DefaultSearchConfig(), retriever.QueryOptions{}, false)
	if err != nil {
		t.Fatalf("QueryWith: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("top 2 returned %d hits", len(got))
	}
}

func TestQueryWith_SnippetIsShortenedAndEllipsised(t *testing.T) {
	root := t.TempDir()
	long := strings.Repeat("word ", 80) // 400 bytes
	installSeams(t, &fakeStack{vs: &fakeVectorStore{hits: []vectorstore.SearchResult{
		{ID: "x", Score: 1, Content: long, Metadata: docMeta(filepath.Join(root, "a.md"), "d", nil)},
	}}})

	got, err := QueryWith(context.Background(), root, t.TempDir(), "q", 10, config.DefaultSearchConfig(), retriever.QueryOptions{}, false)
	if err != nil {
		t.Fatalf("QueryWith: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d hits", len(got))
	}
	if !strings.HasSuffix(got[0].Snippet, "...") || len(got[0].Snippet) > 163 {
		t.Errorf("snippet = %q (len %d), want at most 160 bytes plus \"...\"", got[0].Snippet, len(got[0].Snippet))
	}
}

func TestQueryWith_PathFilterKeepsOnlyThatFolder(t *testing.T) {
	root := t.TempDir()
	installSeams(t, &fakeStack{vs: &fakeVectorStore{hits: []vectorstore.SearchResult{
		{ID: "in", Score: 0.9, Content: "in docs", Metadata: docMeta(filepath.Join(root, "docs", "a.md"), "d1", nil)},
		{ID: "out", Score: 0.8, Content: "outside", Metadata: docMeta(filepath.Join(root, "notes", "b.md"), "d2", nil)},
	}}})

	where, err := retriever.ParseFilter("path:docs")
	if err != nil {
		t.Fatalf("ParseFilter: %v", err)
	}
	got, err := QueryWith(context.Background(), root, t.TempDir(), "q", 10, config.DefaultSearchConfig(),
		retriever.QueryOptions{Where: where}, false)
	if err != nil {
		t.Fatalf("QueryWith: %v", err)
	}
	if len(got) != 1 || got[0].Path != "docs/a.md" {
		t.Errorf("path:docs kept %+v, want only docs/a.md", got)
	}
}

// parentFixture is a guide with two sections. Intro has two chunks on lines 1-4,
// Usage has one chunk on lines 6-7.
func parentFixture() *store.ProcessedDocument {
	return &store.ProcessedDocument{
		DocID: "d1",
		Chunks: []chunker.Chunk{
			{ID: "c1", DocID: "d1", SectionPath: "Intro", Content: "intro one", StartLine: 1, EndLine: 2},
			{ID: "c2", DocID: "d1", SectionPath: "Intro", Content: "intro two", StartLine: 3, EndLine: 4},
			{ID: "c3", DocID: "d1", SectionPath: "Usage", Content: "usage text", StartLine: 6, EndLine: 7},
		},
	}
}

func TestQueryWith_ParentChildReturnsSectionLines(t *testing.T) {
	root := t.TempDir()
	guide := filepath.Join(root, "docs", "guide.md")
	installSeams(t, &fakeStack{
		vs: &fakeVectorStore{hits: []vectorstore.SearchResult{
			{ID: "c1", Score: 0.9, Content: "intro one", Metadata: docMeta(guide, "d1", nil)},
			{ID: "c2", Score: 0.8, Content: "intro two", Metadata: docMeta(guide, "d1", nil)},
		}},
		ds: &fakeDocStore{docs: map[string]*store.ProcessedDocument{"d1": parentFixture()}},
	})

	got, err := QueryWith(context.Background(), root, t.TempDir(), "q", 10, config.DefaultSearchConfig(), retriever.QueryOptions{}, true)
	if err != nil {
		t.Fatalf("QueryWith: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d sections, want 1 (both Intro chunks fold into one): %+v", len(got), got)
	}
	h := got[0]
	if h.Path != "docs/guide.md" || h.Section != "Intro" {
		t.Errorf("section = %q at %q, want Intro in docs/guide.md", h.Section, h.Path)
	}
	if h.StartLine != 1 || h.EndLine != 4 {
		t.Errorf("section lines = %d-%d, want 1-4 (the span of its chunks)", h.StartLine, h.EndLine)
	}
	if !strings.Contains(h.Snippet, "intro one") {
		t.Errorf("snippet = %q, want the section start", h.Snippet)
	}
}

func TestQueryWith_ParentChildSkipsChildrenWithoutDocument(t *testing.T) {
	root := t.TempDir()
	installSeams(t, &fakeStack{
		vs: &fakeVectorStore{hits: []vectorstore.SearchResult{
			{ID: "c1", Score: 0.9, Content: "orphan", Metadata: docMeta(filepath.Join(root, "a.md"), "missing", nil)},
		}},
		ds: &fakeDocStore{docs: map[string]*store.ProcessedDocument{}},
	})

	got, err := QueryWith(context.Background(), root, t.TempDir(), "q", 10, config.DefaultSearchConfig(), retriever.QueryOptions{}, true)
	if err != nil {
		t.Fatalf("QueryWith: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("a child whose document is missing produced sections: %+v", got)
	}
}

func TestQueryWith_OpenErrorsAreReported(t *testing.T) {
	cases := []struct {
		name string
		s    fakeStack
		want string
	}{
		{"embedder", fakeStack{embErr: errors.New("no model")}, "creating embedder"},
		{"vector store", fakeStack{vsErr: errors.New("no store")}, "opening vector store"},
		{"text index", fakeStack{tiErr: errors.New("no index")}, "opening text index"},
		{"document store", fakeStack{dsErr: errors.New("no docs")}, "opening document store"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := c.s
			installSeams(t, &s)
			_, err := QueryWith(context.Background(), t.TempDir(), t.TempDir(), "q", 10, config.DefaultSearchConfig(), retriever.QueryOptions{}, false)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("err = %v, want it to mention %q", err, c.want)
			}
		})
	}
}

func TestQueryWith_SearchErrorIsReported(t *testing.T) {
	for _, parentChild := range []bool{false, true} {
		installSeams(t, &fakeStack{vs: &fakeVectorStore{err: errors.New("disk gone")}})
		_, err := QueryWith(context.Background(), t.TempDir(), t.TempDir(), "q", 10, config.DefaultSearchConfig(), retriever.QueryOptions{}, parentChild)
		if err == nil || !strings.Contains(err.Error(), "document search") {
			t.Errorf("parentChild=%v: err = %v, want a document search error", parentChild, err)
		}
	}
}

func TestQuery_IsPlainQueryWithNoOptions(t *testing.T) {
	root := t.TempDir()
	installSeams(t, &fakeStack{vs: &fakeVectorStore{hits: []vectorstore.SearchResult{
		{ID: "a", Score: 0.7, Content: "alpha", Metadata: docMeta(filepath.Join(root, "a.md"), "d", nil)},
	}}})
	got, err := Query(context.Background(), root, t.TempDir(), "alpha", 5, config.DefaultSearchConfig())
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(got) != 1 || got[0].Path != "a.md" {
		t.Errorf("Query = %+v, want a.md", got)
	}
}

// --- Build ---

func TestBuild_IndexesThenRecordsFingerprint(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "README.md", "# Title")
	analysis := t.TempDir()
	idx := &fakeIndexer{}
	installIndexer(t, idx, nil)

	if err := Build(context.Background(), root, analysis, config.DefaultSearchConfig(), false); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(idx.roots) != 1 || idx.roots[0] != root {
		t.Errorf("indexed roots = %v, want [%s]", idx.roots, root)
	}
	if !idx.closed {
		t.Error("the indexer was not closed")
	}
	stale, err := IsStale(root, analysis)
	if err != nil || stale {
		t.Errorf("right after Build: stale=%v err=%v, want fresh", stale, err)
	}
}

func TestBuild_ForceClearsThePreviousIndex(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "a.md", "x")
	analysis := t.TempDir()
	writeFile(t, analysis, filepath.Join("docs", "old-chunk.txt"), "stale")
	installIndexer(t, &fakeIndexer{}, nil)

	if err := Build(context.Background(), root, analysis, config.DefaultSearchConfig(), true); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if _, err := os.Stat(filepath.Join(Dir(analysis), "old-chunk.txt")); !os.IsNotExist(err) {
		t.Errorf("a file from the previous index survived a forced rebuild: %v", err)
	}
}

func TestBuild_IncrementalKeepsThePreviousIndex(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "a.md", "x")
	analysis := t.TempDir()
	writeFile(t, analysis, filepath.Join("docs", "kept-chunk.txt"), "still indexed")
	installIndexer(t, &fakeIndexer{}, nil)

	if err := Build(context.Background(), root, analysis, config.DefaultSearchConfig(), false); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if _, err := os.Stat(filepath.Join(Dir(analysis), "kept-chunk.txt")); err != nil {
		t.Errorf("an incremental build removed the existing index: %v", err)
	}
}

func TestBuild_PredatingIndexIsRebuiltInFull(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "a.md", "x")
	analysis := t.TempDir()
	writeFile(t, analysis, filepath.Join("docs", "untracked-chunk.txt"), "no record of it")
	installIndexer(t, &fakeIndexer{needsFull: true}, nil)

	if err := Build(context.Background(), root, analysis, config.DefaultSearchConfig(), false); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if _, err := os.Stat(filepath.Join(Dir(analysis), "untracked-chunk.txt")); !os.IsNotExist(err) {
		t.Errorf("an index without chunk records was kept, so its chunks cannot be removed: %v", err)
	}
}

func TestBuild_ConcurrentBuildIsRefused(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "a.md", "x")
	analysis := t.TempDir()
	writeFile(t, analysis, "docs.lock", "held by another build")
	installIndexer(t, &fakeIndexer{}, nil)

	err := Build(context.Background(), root, analysis, config.DefaultSearchConfig(), false)
	if err == nil || !strings.Contains(err.Error(), "another document index build") {
		t.Fatalf("err = %v, want the concurrent-build refusal", err)
	}
	if _, statErr := os.Stat(filepath.Join(analysis, "docs.lock")); statErr != nil {
		t.Errorf("the refused build removed the other build's lock: %v", statErr)
	}
}

func TestBuild_FailuresLeaveTheIndexStale(t *testing.T) {
	cases := []struct {
		name string
		idx  *fakeIndexer
		err  error // indexer constructor error
		want string
	}{
		{"indexer cannot be created", &fakeIndexer{}, errors.New("no embedder"), "creating indexer"},
		{"indexing fails", &fakeIndexer{err: errors.New("parse error")}, nil, "indexing documents"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := t.TempDir()
			writeFile(t, root, "a.md", "x")
			analysis := t.TempDir()
			installIndexer(t, c.idx, c.err)

			err := Build(context.Background(), root, analysis, config.DefaultSearchConfig(), false)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("err = %v, want it to mention %q", err, c.want)
			}
			stale, _ := IsStale(root, analysis)
			if !stale {
				t.Error("a failed build left a fresh-looking index")
			}
		})
	}
}

func TestBuild_FingerprintFailureIsReported(t *testing.T) {
	// The indexer succeeds, but the root cannot be listed to fingerprint it.
	missing := filepath.Join(t.TempDir(), "gone")
	installIndexer(t, &fakeIndexer{}, nil)
	err := Build(context.Background(), missing, t.TempDir(), config.DefaultSearchConfig(), false)
	if err == nil {
		t.Error("Build over a missing root returned nil, want the fingerprint error")
	}
}

func TestBuild_UnwritableAnalysisDirIsReported(t *testing.T) {
	// A file where the analysis directory should be makes the index dir impossible.
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	installIndexer(t, &fakeIndexer{}, nil)
	if err := Build(context.Background(), t.TempDir(), blocker, config.DefaultSearchConfig(), false); err == nil {
		t.Error("Build into a path under a file returned nil")
	}
}

func TestBuild_UnopenableIndexIsRebuilt(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "a.md", "x")
	analysis := t.TempDir()
	writeFile(t, analysis, filepath.Join("docs", "broken-chunk.txt"), "unreadable")

	prev := newIndexer
	t.Cleanup(func() { newIndexer = prev })
	opens := 0
	newIndexer = func(*config.SearchConfig, string) (docIndexer, error) {
		opens++
		if opens == 1 {
			return nil, errors.New("collection metadata file not found")
		}
		return &fakeIndexer{}, nil
	}

	if err := Build(context.Background(), root, analysis, config.DefaultSearchConfig(), false); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if opens != 2 {
		t.Errorf("indexer opened %d times, want 2 (failed open, then the rebuild)", opens)
	}
	if _, err := os.Stat(filepath.Join(Dir(analysis), "broken-chunk.txt")); !os.IsNotExist(err) {
		t.Errorf("the unopenable index was kept: %v", err)
	}
}

// --- staleness and collection edges ---

func TestIsStale_UnreadableFingerprintIsStale(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "a.md", "x")
	analysis := t.TempDir()
	// A directory where the fingerprint file should be cannot be read as a file.
	if err := os.MkdirAll(filepath.Join(Dir(analysis), fingerprintName), 0o755); err != nil {
		t.Fatal(err)
	}
	stale, err := IsStale(root, analysis)
	if err == nil || !stale {
		t.Errorf("stale=%v err=%v, want stale with a read error", stale, err)
	}
}

func TestIsStale_MissingRootIsStaleWithError(t *testing.T) {
	analysis := t.TempDir()
	if err := os.MkdirAll(Dir(analysis), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(Dir(analysis), fingerprintName), []byte("abc\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stale, err := IsStale(filepath.Join(t.TempDir(), "gone"), analysis)
	if err == nil || !stale {
		t.Errorf("stale=%v err=%v, want stale with the listing error", stale, err)
	}
}

func TestCollect_MissingRootIsAnError(t *testing.T) {
	if _, err := Collect(filepath.Join(t.TempDir(), "gone")); err == nil {
		t.Error("Collect over a missing root returned nil")
	}
}

func TestIsDocExt_CaseInsensitiveAndExcludesCode(t *testing.T) {
	for _, p := range []string{"README.MD", "page.Html", "data.json", "notes.markdown", "x.mdx"} {
		if !IsDocExt(p) {
			t.Errorf("IsDocExt(%q) = false, want true", p)
		}
	}
	for _, p := range []string{"main.go", "script.py", "noext"} {
		if IsDocExt(p) {
			t.Errorf("IsDocExt(%q) = true, want false", p)
		}
	}
}

func TestDir_IsUnderTheAnalysisDirectory(t *testing.T) {
	if got, want := Dir(filepath.Join("a", "analysis")), filepath.Join("a", "analysis", "docs"); got != want {
		t.Errorf("Dir = %q, want %q", got, want)
	}
}

// --- LoadConfig ---

func TestLoadConfig_DefaultsWhenNoConfigFile(t *testing.T) {
	t.Chdir(t.TempDir())
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg == nil || cfg.Retriever.TopK != config.DefaultSearchConfig().Retriever.TopK {
		t.Errorf("LoadConfig with no file = %+v, want the defaults", cfg)
	}
}

func TestLoadConfig_UsesTheSearchSectionWhenSet(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.MkdirAll(".plaesy", 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"version":1,"search":{"retriever":{"top_k":7}}}`
	if err := os.WriteFile(filepath.Join(".plaesy", "config.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Retriever.TopK != 7 {
		t.Errorf("TopK = %d, want 7 from the file", cfg.Retriever.TopK)
	}
}

func TestLoadConfig_NullSearchSectionUsesDefaults(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(".plaesy", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(".plaesy", "config.json"), []byte(`{"version":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Retriever.TopK != config.DefaultSearchConfig().Retriever.TopK {
		t.Errorf("TopK = %d, want the default", cfg.Retriever.TopK)
	}
}

func TestLoadConfig_InvalidJSONIsReported(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(".plaesy", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(".plaesy", "config.json"), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadConfig(); err == nil || !strings.Contains(err.Error(), "loading unified config") {
		t.Errorf("err = %v, want a loading error", err)
	}
}

func (f *fakeVectorStore) Name() string { return "fake" }
func (f *fakeTextIndex) Name() string   { return "fake" }
