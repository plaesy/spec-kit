package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/plaesy/spec-kit/internal/graph"
	"github.com/plaesy/spec-kit/internal/semantic"
	"github.com/spf13/cobra"
)

func init() { register(newSearchCmd()) }

func newSearchCmd() *cobra.Command {
	var (
		repoPath string
		outDir   string
		index    bool
		top      int
	)

	cmd := &cobra.Command{
		Use:   "search [query text]",
		Short: "Find files whose purpose is similar in meaning to a query, not just by keyword",
		Long: `Semantic search over the symbols "plaesy graph"/"plaesy analyze" already
extract (function/class/heading names) — catches "same purpose, different
name" matches that "grep -r" cannot, which is what the Anti-Duplication
Protocol (instructions/plaesy.instructions.md) needs before adding new code.

Build the index once (or after symbols change), then search:

  plaesy search --index
  plaesy search "validates a user's credentials"

Prints "similarity<TAB>file<TAB>matched symbols" per line, highest similarity
first.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := graph.Options{RepoPath: repoPath, OutDir: outDir}
			paths, err := graph.ResolvePaths(opts)
			if err != nil {
				return err
			}

			if index {
				g, err := loadOrError(paths)
				if err != nil {
					return err
				}
				fmt.Fprintln(os.Stderr, "[plaesy-search] Loading local embedding model (first run downloads it; cached afterwards)...")
				emb, err := semantic.LoadDefaultEmbedder()
				if err != nil {
					return err
				}
				fmt.Fprintln(os.Stderr, "[plaesy-search] Embedding node symbols...")
				if err := semantic.BuildIndex(cmd.Context(), g, paths.OutFull, paths.RepoRoot, emb); err != nil {
					return err
				}
				fmt.Fprintf(os.Stderr, "[plaesy-search] Index written under %s/embeddings\n", paths.OutFull)
				if err := semantic.RecordFingerprint(opts, paths); err != nil {
					return fmt.Errorf("recording fingerprint: %w", err)
				}
				fmt.Fprintln(os.Stderr, "[plaesy-search] Source fingerprint recorded for staleness detection")
				return nil
			}

			if len(args) != 1 {
				return fmt.Errorf("search needs either --index or a query: plaesy search \"<text>\"")
			}
			// A missing fingerprint is a different condition from a changed
			// source, and conflating them produces a message that is simply
			// false ("source files changed") for every index built before
			// staleness detection existed. Both refuse to query — the index's
			// freshness cannot be proven — but each says what to actually do.
			stored, err := semantic.StoredFingerprint(paths.OutFull)
			if err != nil {
				return fmt.Errorf("checking index staleness: %w", err)
			}
			if stored == "" {
				return fmt.Errorf("no source fingerprint recorded for the semantic index, so its freshness cannot be verified. Run 'plaesy search --index' once to enable staleness detection")
			}
			stale, err := semantic.IsIndexStale(opts, paths, paths.OutFull)
			if err != nil {
				return fmt.Errorf("checking index staleness: %w", err)
			}
			if stale {
				return fmt.Errorf("semantic index is stale (source files changed since last --index). Run 'plaesy search --index' to rebuild")
			}
			emb, err := semantic.LoadDefaultEmbedder()
			if err != nil {
				return err
			}
			matches, err := semantic.Query(cmd.Context(), paths.OutFull, args[0], top, emb)
			if err != nil {
				return err
			}
			for _, m := range matches {
				fmt.Fprintf(os.Stdout, "%.4f\t%s\t%s\n", m.Similarity, m.NodeID, strings.Join(m.Symbols, ", "))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&repoPath, "path", ".", "project directory to scan")
	cmd.Flags().StringVar(&outDir, "outdir", ".plaesy/analysis", "output directory (relative to --path; shared with plaesy graph)")
	cmd.Flags().BoolVar(&index, "index", false, "(re)build the semantic-search index from the current graph")
	cmd.Flags().IntVar(&top, "top", 10, "max results")

	return cmd
}
