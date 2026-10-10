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
	cmd := &cobra.Command{
		Use:   "search [command] [query text]",
		Short: "Semantic search over code symbols and local documents",
		Long: `Unified search command for both code symbols and local documents.

Code Symbol Search (existing):
  plaesy search --index                         # Build symbol index
  plaesy search "validates user credentials"    # Search code symbols

Document Search (new):
  plaesy search index ./docs                    # Build document index
  plaesy search query "golang tutorial"         # Search documents
  plaesy search query "golang" --json           # JSON output

Use "plaesy stats" for search index statistics.

Use "plaesy search [command] --help" for more information.`,
		Args: cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if --index flag is set (legacy mode)
			index, _ := cmd.Flags().GetBool("index")
			if index {
				return runLegacySearch(cmd, args)
			}

			// If no subcommand and no args, show help
			if len(args) == 0 {
				return cmd.Help()
			}

			// Check if first arg is a subcommand
			if len(args) >= 1 && (args[0] == "index" || args[0] == "query") {
				// Let cobra handle subcommands
				return cmd.Help()
			}

			// Legacy mode: semantic code search (query without --index)
			return runLegacySearch(cmd, args)
		},
	}

	// Add flags for legacy search
	var (
		repoPath string
		outDir   string
		index    bool
		top      int
	)

	cmd.Flags().StringVar(&repoPath, "path", ".", "project directory to scan")
	cmd.Flags().StringVar(&outDir, "outdir", ".plaesy/analysis", "output directory (relative to --path; shared with plaesy graph)")
	cmd.Flags().BoolVar(&index, "index", false, "(re)build the semantic-search index from the current graph")
	cmd.Flags().IntVar(&top, "top", 10, "max results")

	// Add subcommands
	cmd.AddCommand(newSearchIndexCmd())
	cmd.AddCommand(newSearchQueryCmd())

	return cmd
}

func runLegacySearch(cmd *cobra.Command, args []string) error {
	// Flags are already defined on the parent command
	repoPath, _ := cmd.Flags().GetString("path")
	outDir, _ := cmd.Flags().GetString("outdir")
	index, _ := cmd.Flags().GetBool("index")
	top, _ := cmd.Flags().GetInt("top")

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
}
