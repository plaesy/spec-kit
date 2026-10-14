// Package semantic adds embedding-based semantic search on top of the
// symbols plaesy analyze already extracts (scripts/internal/graph), so the
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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	chromem "github.com/philippgille/chromem-go"

	"github.com/plaesy/spec-kit/internal/graph"
	"github.com/plaesy/spec-kit/internal/search/embedder"
)

const collectionName = "plaesy-nodes"

// LoadDefaultEmbedder loads the project's embedding model, the same E5 model
// the document search uses, so one model serves every index. The first call
// downloads it; later calls read the cache.
func LoadDefaultEmbedder() (embedder.Embedder, error) {
	emb, err := embedder.NewRembedEmbedder(embedder.DefaultConfig())
	if err != nil {
		return nil, fmt.Errorf("semantic: loading embedding model: %w", err)
	}
	return emb, nil
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
// comments (see nodeText/docsFor) — pass the same root that plaesy analyze
// scanned to build g.
func BuildIndex(ctx context.Context, g *graph.Graph, outDir, repoRoot string, emb embedder.Embedder) error {
	_, err := buildIndex(ctx, g, outDir, repoRoot, emb)
	return err
}

// buildIndex is BuildIndex, also reporting how many documents were embedded.
// A document whose text hash matches the stored one keeps its stored vector, so
// an unchanged rebuild embeds nothing. The collection is still recreated each
// time, which is what drops orphans: a document no longer in the graph is not
// carried over.
func buildIndex(ctx context.Context, g *graph.Graph, outDir, repoRoot string, emb embedder.Embedder) (int, error) {
	dbPath := filepath.Join(outDir, "embeddings")
	db, err := chromem.NewPersistentDB(dbPath, true)
	if err != nil {
		return 0, fmt.Errorf("semantic: opening embedding store at %s: %w", dbPath, err)
	}

	docs := indexDocs(repoRoot, indexableNodes(g))
	ckPath := checkpointPath(outDir)
	if len(docs) == 0 {
		removeCheckpoint(ckPath)
	}
	reused := reusableVectors(ctx, db.GetCollection(collectionName, nil), docs)

	// Vectors an interrupted earlier build already embedded for this same input.
	digest := inputDigest(docs)
	resumed, ckValid := loadCheckpoint(ckPath, digest, docs)
	for id, v := range resumed {
		reused[id] = v
	}

	_ = db.DeleteCollection(collectionName)
	coll, err := db.CreateCollection(collectionName, nil, nil)
	if err != nil {
		return 0, fmt.Errorf("semantic: creating collection: %w", err)
	}
	if len(docs) == 0 {
		return 0, nil
	}

	vecs := make([][]float32, len(docs))
	var toEmbedAt []int
	for i, d := range docs {
		if v, ok := reused[d.id]; ok {
			vecs[i] = v
			continue
		}
		toEmbedAt = append(toEmbedAt, i)
	}

	// Embed in batches. Each batch is written to the checkpoint before the next
	// one starts, so an interrupted build keeps every batch that finished.
	embedded := 0
	for start := 0; start < len(toEmbedAt); start += checkpointBatchSize {
		end := min(start+checkpointBatchSize, len(toEmbedAt))
		batchAt := toEmbedAt[start:end]
		texts := make([]string, len(batchAt))
		for j, i := range batchAt {
			texts[j] = docs[i].text
		}
		// Passages get the passage prefix; the vectors are stored as computed.
		computed, err := emb.EmbedPassages(ctx, texts)
		if err != nil {
			return 0, fmt.Errorf("semantic: embedding %d documents: %w", len(texts), err)
		}
		entries := make([]checkpointEntry, len(batchAt))
		for j, i := range batchAt {
			vecs[i] = computed[j]
			entries[j] = checkpointEntry{ID: docs[i].id, Hash: docs[i].meta[metaHash], Vector: computed[j]}
		}
		if err := writeCheckpointBatch(ckPath, digest, entries, ckValid); err != nil {
			return 0, err
		}
		ckValid = true
		embedded += len(batchAt)
	}

	ids := make([]string, len(docs))
	contents := make([]string, len(docs))
	metas := make([]map[string]string, len(docs))
	for i, d := range docs {
		ids[i] = d.id
		contents[i] = d.text
		metas[i] = d.meta
	}
	if err := coll.Add(ctx, ids, vecs, metas, contents); err != nil {
		return 0, fmt.Errorf("semantic: storing %d documents: %w", len(docs), err)
	}
	// Only a finished build removes its checkpoint: until the collection is
	// stored, the checkpoint is what a rerun resumes from.
	removeCheckpoint(ckPath)
	return embedded, nil
}

// reusableVectors returns the stored vector for each new document whose text
// hash matches the one recorded when it was embedded. A nil or empty old
// collection yields nothing, so every document is embedded.
func reusableVectors(ctx context.Context, old *chromem.Collection, docs []indexDoc) map[string][]float32 {
	reused := map[string][]float32{}
	if old == nil {
		return reused
	}
	for _, d := range docs {
		prev, err := old.GetByID(ctx, d.id)
		if err != nil || prev.Metadata[metaHash] != d.meta[metaHash] {
			continue
		}
		reused[d.id] = prev.Embedding
	}
	return reused
}

// indexDoc is one embedded document: a Go declaration unit, or a whole file.
// meta[metaHash] is the hash of text; the vector depends on text alone.
type indexDoc struct {
	id   string
	text string
	meta map[string]string
}

// Metadata keys stored with every document. A file document has kind "file"
// and start/end "0"; a unit document has the declaration's kind and range.
const (
	metaPath  = "path"
	metaKind  = "kind"
	metaName  = "name"
	metaStart = "start"
	metaEnd   = "end"
	metaHash  = "hash"
)

// textHash is the content hash that decides whether a stored vector is reused.
func textHash(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

// indexDocs turns each node into documents. A Go file becomes one document per
// declaration (goUnits), so a query returns the function that matches, not the
// whole file. Every other file, and a Go file that cannot be read, parsed, or
// has no declarations, becomes one file-level document from nodeText.
func indexDocs(repoRoot string, nodes []graph.Node) []indexDoc {
	var out []indexDoc
	for _, n := range nodes {
		if strings.EqualFold(filepath.Ext(n.ID), ".go") {
			if units := goUnitsFor(repoRoot, n); len(units) > 0 {
				for _, u := range units {
					out = append(out, unitDoc(n.ID, u))
				}
				continue
			}
		}
		out = append(out, indexDoc{
			id:   n.ID,
			text: nodeText(repoRoot, n),
			meta: map[string]string{metaPath: n.ID, metaKind: "file", metaStart: "0", metaEnd: "0"},
		})
	}
	for i := range out {
		out[i].meta[metaHash] = textHash(out[i].text)
	}
	return out
}

// goUnitsFor returns n's declarations, or nil when the source cannot be read
// or parsed. Callers fall back to file-level indexing in that case.
func goUnitsFor(repoRoot string, n graph.Node) []Unit {
	content, ok := readSource(repoRoot, n)
	if !ok {
		return nil
	}
	units, err := goUnits(content)
	if err != nil {
		return nil
	}
	return units
}

// unitDoc is the embedded text for one declaration: its path, then its name and
// the first line of its doc comment. The range is metadata, not embedded text.
func unitDoc(path string, u Unit) indexDoc {
	var b strings.Builder
	b.WriteString(path)
	b.WriteByte('\n')
	b.WriteString(u.Name)
	if u.Doc != "" {
		b.WriteString(": ")
		b.WriteString(u.Doc)
	}
	return indexDoc{
		// The name is part of the ID so two declarations on one line stay distinct.
		id:   fmt.Sprintf("%s:%d-%d#%s", path, u.Start, u.End, u.Name),
		text: b.String(),
		meta: map[string]string{
			metaPath:  path,
			metaKind:  u.Kind,
			metaName:  u.Name,
			metaStart: strconv.Itoa(u.Start),
			metaEnd:   strconv.Itoa(u.End),
		},
	}
}

// Match is one semantic-query result. A unit match has Kind, Name and the
// Start/End line range (1-based, End inclusive) and Symbols holds only Name. A
// file match has Kind "file", Symbols from the file's extracted symbol names,
// and Start/End of 0. Similarity is the cosine similarity to the query text
// ([-1, 1], higher = more similar).
type Match struct {
	NodeID     string
	Kind       string
	Name       string
	Start      int
	End        int
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
func Query(ctx context.Context, outDir, queryText string, topN int, emb embedder.Embedder) ([]Match, error) {
	dbPath := filepath.Join(outDir, "embeddings")
	db, err := chromem.NewPersistentDB(dbPath, true)
	if err != nil {
		return nil, fmt.Errorf("semantic: opening embedding store at %s: %w", dbPath, err)
	}
	coll := db.GetCollection(collectionName, nil)
	if coll == nil {
		return nil, fmt.Errorf("semantic: no embedding index at %s — run plaesy analyze first", dbPath)
	}

	n := topN
	if have := coll.Count(); n > have {
		n = have
	}
	if n <= 0 {
		return nil, nil
	}

	// The query gets the query prefix, which the stored passages do not share.
	qvec, err := emb.EmbedQuery(ctx, queryText)
	if err != nil {
		return nil, fmt.Errorf("semantic: embedding query: %w", err)
	}
	results, err := coll.QueryEmbedding(ctx, qvec, n, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("semantic: querying: %w", err)
	}
	out := make([]Match, len(results))
	for i, r := range results {
		m := Match{NodeID: r.Metadata[metaPath], Kind: r.Metadata[metaKind], Similarity: r.Similarity}
		if m.Kind == "file" {
			m.Symbols = symbolsFromContent(r.Content)
		} else {
			m.Name = r.Metadata[metaName]
			m.Symbols = []string{m.Name}
			m.Start, _ = strconv.Atoi(r.Metadata[metaStart])
			m.End, _ = strconv.Atoi(r.Metadata[metaEnd])
		}
		out[i] = m
	}
	return out, nil
}
