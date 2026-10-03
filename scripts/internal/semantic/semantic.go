// Package semantic adds embedding-based semantic search on top of the
// symbols plaesy graph already extracts (scripts/internal/graph), so the
// anti-duplication check in instructions/plaesy.instructions.md can find
// "same purpose, different wording" files — something literal grep/find
// cannot do.
//
// v1 scope is file-level, not function-level: each graph node's extracted
// symbol names are joined into one text blob and embedded as a single
// document (enriched with doc comments for the languages that have an
// extractor: Go, JS/TS, Python — see docsFor). Per-symbol/per-language
// embedding is a possible future step (see
// .plaesy/decisions/embedding-based-semantic-duplicate-search.md) but needs
// doc-comment extraction for the remaining languages symbols.go supports;
// file-level is the smallest slice that's still useful today.
package semantic

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	chromem "github.com/philippgille/chromem-go"
	"github.com/rostamlabs/rembed"

	"github.com/plaesy/spec-kit/internal/graph"
)

const collectionName = "plaesy-nodes"

// defaultModel is a small, widely-used sentence-embedding model. Downloaded
// once from the Hugging Face Hub on first use and cached under
// $REMBED_CACHE (or rembed's default cache dir) thereafter — no network
// access is needed on subsequent runs.
const defaultModel = "sentence-transformers/all-MiniLM-L6-v2"

// Embedder is the minimal interface semantic needs from an embedding
// backend. rembed.Embedder satisfies it; tests use a fake instead of
// downloading a real model.
type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
}

// LoadDefaultEmbedder loads the default local embedding model via rembed.
// This is the only place that touches rembed directly, so callers (and
// tests) can depend on the Embedder interface instead.
func LoadDefaultEmbedder() (Embedder, error) {
	emb, err := rembed.Load(defaultModel)
	if err != nil {
		return nil, fmt.Errorf("semantic: loading embedding model %q: %w", defaultModel, err)
	}
	return emb, nil
}

// embeddingFunc adapts an Embedder to chromem-go's single-text
// EmbeddingFunc signature.
func embeddingFunc(e Embedder) chromem.EmbeddingFunc {
	return func(ctx context.Context, text string) ([]float32, error) {
		vecs, err := e.Embed(ctx, []string{text})
		if err != nil {
			return nil, err
		}
		if len(vecs) != 1 {
			return nil, fmt.Errorf("semantic: expected 1 embedding, got %d", len(vecs))
		}
		return vecs[0], nil
	}
}

// nodeText is the text embedded for one graph node: its id plus its
// extracted symbol names, one per line — "<symbol>: <doc comment>" when a
// doc comment is available, bare "<symbol>" otherwise. Symbol names alone
// already carry real semantic signal ("ValidateUser" and "CheckUserIsValid"
// land close together in embedding space even without a doc comment); the
// doc comment, when present, sharpens that signal further.
//
// Doc-comment extraction covers the languages with the highest share of
// real-world code — Go (godocs.go), JS/TS (jsdocs.go) and Python
// (pydocs.go) — keyed off the file extension by docsFor. The remaining
// languages symbols.go supports (Rust, Dart, Java, Kotlin, C#, C-family,
// Swift, Ruby, PHP, Shell, PowerShell, Markdown) each have their own
// doc-comment grammar (rustdoc, KDoc, XML doc comments, ...) and are
// out-of-scope here: bare symbol names are still embedded for those files,
// just without the doc-comment boost.
func nodeText(repoRoot string, n graph.Node) string {
	docs := docsFor(repoRoot, n)

	var b strings.Builder
	b.WriteString(n.ID)
	for _, s := range n.Symbols {
		b.WriteByte('\n')
		b.WriteString(s)
		if doc, ok := docs[s]; ok {
			b.WriteString(": ")
			b.WriteString(firstLine(doc))
		}
	}
	return b.String()
}

// contentExtractor turns one source file's text into a symbol-name -> doc
// comment map. Every language extractor in this package has this shape, which
// is what lets docsFor dispatch through a table instead of a switch full of
// near-identical wrappers.
type contentExtractor func(content string) map[string]string

// goDocCommentsOrNil adapts goDocComments, the one extractor that can fail
// (go/parser rejects a file it cannot parse), to the contentExtractor shape. A
// file Go cannot parse yields no docs rather than an error: doc comments are a
// best-effort enrichment, never a reason to fail indexing.
func goDocCommentsOrNil(content string) map[string]string {
	docs, err := goDocComments(content)
	if err != nil {
		return nil
	}
	return docs
}

// docsByExt maps a file extension to its doc-comment extractor.
//
// The key set is the set of extensions symbols.go extracts symbols for, minus
// the ones with no doc-comment convention worth extracting (Markdown, whose
// "symbols" are headings and whose prose is the content itself). A doc-comment
// map for a file whose symbols were never extracted would be dead weight, so
// the two lists must be kept in step: adding a language to symbols.go without
// an entry here is not an error, it just means bare symbol names get embedded
// for that language.
//
// TestDocsFor_WiredExtensions in docsfor_test.go covers this table end to end —
// a row per extension, asserting a real file with a real doc comment comes back
// enriched. Add a row there when adding an entry here.
//
// Note ".hh" and ".cxx" are unreachable through the real pipeline today:
// internal/graph's includeExtRE (graph.go) never collects those two extensions,
// so no graph node can carry them even though symbolsOf handles them. They are
// listed here because symbolsOf clearly intends to support them, and the fix
// belongs in internal/graph rather than here. Tracked as
// .plaesy/tasks/backlog/medium_graph-include-ext-missing-hh-cxx.md.
var docsByExt = map[string]contentExtractor{
	".go":  goDocCommentsOrNil,
	".js":  jsDocComments,
	".jsx": jsDocComments,
	".ts":  jsDocComments,
	".tsx": jsDocComments,
	".py":  pyDocComments,

	".rs":  rustDocComments,
	".rb":  rubyDocComments,
	".sh":  shDocComments,
	".ps1": psDocComments,

	".dart": dartDocComments,
	".java": javaDocComments,
	".kt":   kotlinDocComments,
	".kts":  kotlinDocComments,
	".cs":   csharpDocComments,

	".c":   cFamilyDocComments,
	".h":   cFamilyDocComments,
	".cpp": cFamilyDocComments,
	".hpp": cFamilyDocComments,
	".cc":  cFamilyDocComments,
	".hh":  cFamilyDocComments,
	".cxx": cFamilyDocComments,

	".swift": swiftDocComments,
	".php":   phpDocComments,
}

// docsFor returns n's doc comments keyed by symbol name, dispatching on the
// file extension to the language-specific extractor.
//
// nil for an unmapped extension, and for a file that is unreadable or fails
// to parse: doc comments are a best-effort enrichment, never a reason to fail
// indexing.
func docsFor(repoRoot string, n graph.Node) map[string]string {
	extract, ok := docsByExt[strings.ToLower(filepath.Ext(n.ID))]
	if !ok {
		return nil
	}
	content, ok := readSource(repoRoot, n)
	if !ok {
		return nil
	}
	return extract(content)
}

// readSource reads n's source file under repoRoot. ok is false when it
// can't be read, which every caller treats as "no docs for this node".
func readSource(repoRoot string, n graph.Node) (string, bool) {
	content, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(n.ID)))
	if err != nil {
		return "", false
	}
	return string(content), true
}

// indexableNodes filters out nodes with no symbols: an empty text embeds to
// a meaningless vector and only pollutes results.
func indexableNodes(g *graph.Graph) []graph.Node {
	out := make([]graph.Node, 0, len(g.Nodes))
	for _, n := range g.Nodes {
		if len(n.Symbols) == 0 {
			continue
		}
		out = append(out, n)
	}
	return out
}

// BuildIndex embeds every node in g that has extracted symbols and persists
// the resulting vector collection under outDir (outDir/embeddings, one gob
// file per document — see chromem-go's NewPersistentDB). Safe to call again
// later: it recreates the collection each time rather than incrementally
// updating it, which keeps staleness handling simple (callers decide when
// to rebuild, e.g. plaesy analyze's existing fingerprint check).
//
// repoRoot locates each node's source file on disk, used to pull in doc
// comments (see nodeText/docsFor) — pass the same root plaesy graph
// scanned to build g.
func BuildIndex(ctx context.Context, g *graph.Graph, outDir, repoRoot string, embedder Embedder) error {
	dbPath := filepath.Join(outDir, "embeddings")
	db, err := chromem.NewPersistentDB(dbPath, true)
	if err != nil {
		return fmt.Errorf("semantic: opening embedding store at %s: %w", dbPath, err)
	}
	_ = db.DeleteCollection(collectionName)
	coll, err := db.CreateCollection(collectionName, nil, embeddingFunc(embedder))
	if err != nil {
		return fmt.Errorf("semantic: creating collection: %w", err)
	}

	nodes := indexableNodes(g)
	if len(nodes) == 0 {
		return nil
	}
	ids := make([]string, len(nodes))
	contents := make([]string, len(nodes))
	for i, n := range nodes {
		ids[i] = n.ID
		contents[i] = nodeText(repoRoot, n)
	}
	if err := coll.Add(ctx, ids, nil, nil, contents); err != nil {
		return fmt.Errorf("semantic: embedding %d nodes: %w", len(nodes), err)
	}
	return nil
}

// Match is one semantic-query result: a node id, the symbols that made it
// match (already extracted by plaesy graph — no reason to report only the
// file name when project.graph.json has the finer-grained detail), and how
// similar it is to the query text (cosine similarity, [-1, 1], higher =
// more similar).
type Match struct {
	NodeID     string
	Symbols    []string
	Similarity float32
}

// symbolsFromContent recovers the symbol list from a stored document's
// content, which nodeText built as "<nodeID>\n<symbol>\n<symbol>...".
func symbolsFromContent(content string) []string {
	lines := strings.Split(content, "\n")
	if len(lines) <= 1 {
		return nil
	}
	return lines[1:]
}

// Query loads the persisted index built by BuildIndex and returns the
// topN nodes most similar in meaning to queryText. Returns an error if
// BuildIndex was never called for outDir (no collection to query).
func Query(ctx context.Context, outDir, queryText string, topN int, embedder Embedder) ([]Match, error) {
	dbPath := filepath.Join(outDir, "embeddings")
	db, err := chromem.NewPersistentDB(dbPath, true)
	if err != nil {
		return nil, fmt.Errorf("semantic: opening embedding store at %s: %w", dbPath, err)
	}
	coll := db.GetCollection(collectionName, embeddingFunc(embedder))
	if coll == nil {
		return nil, fmt.Errorf("semantic: no embedding index at %s — run with --index first", dbPath)
	}

	n := topN
	if have := coll.Count(); n > have {
		n = have
	}
	if n <= 0 {
		return nil, nil
	}

	results, err := coll.Query(ctx, queryText, n, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("semantic: querying: %w", err)
	}
	out := make([]Match, len(results))
	for i, r := range results {
		out[i] = Match{NodeID: r.ID, Symbols: symbolsFromContent(r.Content), Similarity: r.Similarity}
	}
	return out, nil
}
