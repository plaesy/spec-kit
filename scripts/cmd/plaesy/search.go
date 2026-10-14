package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/plaesy/spec-kit/internal/graph"
	"github.com/plaesy/spec-kit/internal/search/docindex"
	"github.com/plaesy/spec-kit/internal/search/retriever"
	"github.com/plaesy/spec-kit/internal/semantic"
	"github.com/spf13/cobra"
)

func init() { register(newSearchCmd()) }

// rrfK is the damping constant of Reciprocal Rank Fusion. 60 is the value from
// Cormack, Clarke and Büttcher (2009); it is a convention, not tuned for this
// project.
const rrfK = 60

// newSearchCmd is the single search command. The indexes it reads are built by
// `plaesy analyze`, not by search itself.
func newSearchCmd() *cobra.Command {
	var repoPath, outDir, kind, filterSpec, extSpec string
	var parentChild bool
	var top int

	cmd := &cobra.Command{
		Use:   "search <text>",
		Short: "Search the project by meaning (indexes are built by plaesy analyze)",
		Long: `Search the project by meaning. "plaesy analyze" builds the indexes it reads;
a query refuses an index that is missing or older than the sources.

Symbols (code declarations, and headings in Markdown) and document passages
(Markdown, HTML and JSON files) are ranked in one list. Each row is labelled
symbol or doc. The two lists are merged with Reciprocal Rank Fusion, because
their similarity scores are not on the same scale.

  plaesy search "validates user credentials"
  plaesy search "refund policy" --kind doc`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return searchQuery(cmd, args[0], repoPath, outDir, top, kind, filterSpec, extSpec, parentChild)
		},
	}

	cmd.Flags().StringVar(&repoPath, "path", ".", "project directory")
	cmd.Flags().StringVar(&outDir, "outdir", ".plaesy/analysis", "analysis directory (relative to --path; written by plaesy analyze)")
	cmd.Flags().IntVar(&top, "top", 10, "max results")
	cmd.Flags().StringVar(&kind, "kind", "", "only symbol or only doc results (default: both)")
	cmd.Flags().StringVar(&filterSpec, "filter", "", "narrow doc results, e.g. source:markdown,path:docs,file_type:md,date:>2024-01-01 (doc results only)")
	cmd.Flags().StringVar(&extSpec, "ext", "", "keep only these extensions, comma-separated, e.g. go,ts or md,html (matches either; applies to symbol and doc results)")
	cmd.Flags().BoolVar(&parentChild, "parent-child", false, "return whole sections with the matched passages as children (doc results only)")
	return cmd
}

// row is one search result before printing. Score is the fused RRF score.
type row struct {
	score    float64
	kind     string
	ext      string
	location string
	name     string
}

// searchQuery answers from both indexes. It refuses an index that is missing or
// older than the sources it was built from, for each kind it would return.
func searchQuery(cmd *cobra.Command, text, repoPath, outDir string, top int, kind, filterSpec, extSpec string, parentChild bool) error {
	switch kind {
	case "", "symbol", "doc":
	default:
		return fmt.Errorf("--kind must be symbol or doc, got %q", kind)
	}
	// Parsed before any index is read, so a bad filter fails at once.
	where, err := retriever.ParseFilter(filterSpec)
	if err != nil {
		return fmt.Errorf("--filter: %w", err)
	}

	opts := graph.Options{RepoPath: repoPath, OutDir: outDir}
	paths, err := graph.ResolvePaths(opts)
	if err != nil {
		return err
	}

	var symbolRows, docRows []row
	if kind != "doc" {
		rows, err := symbolResults(cmd, text, opts, paths, top)
		if err != nil {
			return err
		}
		symbolRows = rows
	}
	if kind != "symbol" {
		rows, err := docResults(cmd, text, repoPath, paths.OutFull, top, where, parentChild)
		if err != nil {
			return err
		}
		docRows = rows
	}

	merged := fuse(symbolRows, docRows)
	// A Markdown heading comes from the symbol index but is labelled doc, so the
	// filter has to look at the label, not at which index the row came from.
	if kind != "" {
		merged = keepKind(merged, kind)
	}
	// --ext applies to both kinds: symbol and doc rows carry their extension.
	merged = keepExt(merged, parseExts(extSpec))
	if top > 0 && len(merged) > top {
		merged = merged[:top]
	}
	// Columns: score, kind, extension, path:start-end, name. The path is relative
	// to the project, so it can be opened directly.
	for _, r := range merged {
		fmt.Fprintf(os.Stdout, "%.4f\t%s\t%s\t%s\t%s\n", r.score, r.kind, r.ext, r.location, r.name)
	}
	return nil
}

// symbolResults ranks the symbol index. Its rank order is the order semantic
// returns, and it is the only input to fusion for symbols.
func symbolResults(cmd *cobra.Command, text string, opts graph.Options, paths graph.Paths, top int) ([]row, error) {
	stored, err := semantic.StoredFingerprint(paths.OutFull)
	if err != nil {
		return nil, fmt.Errorf("checking index staleness: %w", err)
	}
	if stored == "" {
		return nil, fmt.Errorf("no index found in %s. Run 'plaesy analyze' first", paths.OutFull)
	}
	stale, err := semantic.IsIndexStale(opts, paths, paths.OutFull)
	if err != nil {
		return nil, fmt.Errorf("checking index staleness: %w", err)
	}
	if stale {
		return nil, fmt.Errorf("index is stale (source files changed since the last analyze). Run 'plaesy analyze' to refresh it")
	}
	emb, err := semantic.LoadDefaultEmbedder()
	if err != nil {
		return nil, err
	}
	matches, err := semantic.Query(cmd.Context(), paths.OutFull, text, top, emb)
	if err != nil {
		return nil, err
	}

	rows := make([]row, 0, len(matches))
	for _, m := range matches {
		ext := strings.TrimPrefix(filepath.Ext(m.NodeID), ".")
		location := m.NodeID
		name := strings.Join(m.Symbols, ", ")
		if m.Kind != "file" {
			name = m.Name
		}
		if m.Start > 0 {
			location = fmt.Sprintf("%s:%d-%d", m.NodeID, m.Start, m.End)
		}
		kind := "symbol"
		if docindex.IsDocExt(m.NodeID) {
			kind = "doc"
		}
		rows = append(rows, row{kind: kind, ext: ext, location: location, name: name})
	}
	return rows, nil
}

// docResults ranks the document index. It refuses an index that is missing or
// built from documents that have since changed.
func docResults(cmd *cobra.Command, text, repoPath, analysisDir string, top int, where retriever.Filter, parentChild bool) ([]row, error) {
	stale, err := docindex.IsStale(repoPath, analysisDir)
	if err != nil {
		return nil, fmt.Errorf("checking document index staleness: %w", err)
	}
	if stale {
		return nil, fmt.Errorf("document index is missing or stale (a document changed since the last analyze). Run 'plaesy analyze' to refresh it")
	}
	cfg, err := docindex.LoadConfig()
	if err != nil {
		return nil, err
	}
	hits, err := docindex.QueryWith(cmd.Context(), repoPath, analysisDir, text, top, cfg,
		retriever.QueryOptions{Where: where}, parentChild)
	if err != nil {
		return nil, err
	}

	rows := make([]row, 0, len(hits))
	for _, h := range hits {
		name := h.Section
		if name == "" {
			name = h.Snippet
		}
		rows = append(rows, row{
			kind:     "doc",
			ext:      strings.TrimPrefix(filepath.Ext(h.Path), "."),
			location: docLocation(h),
			name:     name,
		})
	}
	return rows, nil
}

// keepKind returns only the rows whose label is kind, keeping their order.
func keepKind(rows []row, kind string) []row {
	out := make([]row, 0, len(rows))
	for _, r := range rows {
		if r.kind == kind {
			out = append(out, r)
		}
	}
	return out
}

// parseExts reads a comma-separated --ext value into lowercase extensions without
// the leading dot. Empty terms are skipped, so "go, .ts," means [go ts].
func parseExts(spec string) []string {
	var exts []string
	for _, term := range strings.Split(spec, ",") {
		ext := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(term), "."))
		if ext != "" {
			exts = append(exts, ext)
		}
	}
	return exts
}

// keepExt returns only the rows whose extension is one of exts, keeping their
// order. An empty exts keeps every row. The extensions are OR-ed together.
func keepExt(rows []row, exts []string) []row {
	if len(exts) == 0 {
		return rows
	}
	want := make(map[string]bool, len(exts))
	for _, e := range exts {
		want[e] = true
	}
	out := make([]row, 0, len(rows))
	for _, r := range rows {
		if want[strings.ToLower(r.ext)] {
			out = append(out, r)
		}
	}
	return out
}

// fuse merges the two ranked lists with Reciprocal Rank Fusion. A row's score is
// 1/(rrfK + rank) in its own list, so the best row of each list ranks first and
// the lists interleave. The result is sorted by score, highest first.
func fuse(symbols, docs []row) []row {
	out := make([]row, 0, len(symbols)+len(docs))
	for _, list := range [][]row{symbols, docs} {
		for i, r := range list {
			r.score = 1 / float64(rrfK+i+1)
			out = append(out, r)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].score > out[j].score })
	return out
}

// docLocation is the location column for a document hit: path:start-end when
// the passage has source lines, and the path alone when it does not.
func docLocation(h docindex.Hit) string {
	if h.StartLine > 0 {
		return fmt.Sprintf("%s:%d-%d", h.Path, h.StartLine, h.EndLine)
	}
	return h.Path
}
