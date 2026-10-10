package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/plaesy/spec-kit/internal/config"
	"github.com/plaesy/spec-kit/internal/search/embedder"
	"github.com/plaesy/spec-kit/internal/search/indexer"
	"github.com/plaesy/spec-kit/internal/search/retriever"
	"github.com/plaesy/spec-kit/internal/search/store"
	"github.com/plaesy/spec-kit/internal/search/textindex"
	"github.com/plaesy/spec-kit/internal/search/vectorstore"
	"github.com/spf13/cobra"
)

func init() {
	register(newSearchIndexCmd())
	register(newSearchQueryCmd())
}

func newSearchIndexCmd() *cobra.Command {
	var (
		repoPath    string
		outDir      string
		incremental bool
		force       bool
		resume      bool
		workers     int
	)

	cmd := &cobra.Command{
		Use:   "index [path]",
		Short: "Build or update the search index",
		Long: `Build a semantic search index from documents in the specified path.
Supports Markdown (.md, .markdown, .mdx), HTML (.html, .htm), and JSON (.json) files.

The index includes:
- Vector embeddings (ONNX, multilingual-e5-base)
- Full-text search (BM25 via bleve)
- Structure-aware chunking (200-500 tokens, 50 token overlap)

Examples:
  plaesy search index ./docs                    # Full index
  plaesy search index ./docs --incremental      # Only changed files
  plaesy search index --force                   # Rebuild from scratch
  plaesy search index --resume                  # Resume interrupted indexing`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rootPath := repoPath
			if len(args) > 0 {
				rootPath = args[0]
			}

			// Load search config
			unifiedConfigMgr := config.NewConfigManager[config.UnifiedConfig](
				"", // Use default unified config path
				config.DefaultUnifiedConfig(),
			)
			unifiedConfig, err := unifiedConfigMgr.Load()
			if err != nil {
				return fmt.Errorf("loading unified config: %w", err)
			}

			if unifiedConfig.Search == nil {
				unifiedConfig.Search = config.DefaultSearchConfig()
			}
			searchConfig := unifiedConfig.Search

			if workers > 0 {
				searchConfig.Indexer.Workers = workers
			}
			searchConfig.Indexer.Resume = resume

			// Create indexer
			idx, err := indexer.NewIndexer(searchConfig, outDir)
			if err != nil {
				return fmt.Errorf("creating indexer: %w", err)
			}
			defer idx.Close()

			// Run indexing
			fmt.Fprintf(os.Stderr, "[plaesy search index] Indexing %s\n", rootPath)
			fmt.Fprintf(os.Stderr, "[plaesy search index] Output directory: %s\n", outDir)
			fmt.Fprintf(os.Stderr, "[plaesy search index] Incremental: %v, Force: %v, Resume: %v\n", incremental, force, resume)

			if force {
				// Clean existing index
				if err := cleanIndex(outDir); err != nil {
					return fmt.Errorf("cleaning index: %w", err)
				}
				fmt.Fprintln(os.Stderr, "[plaesy search index] Cleaned existing index")
			}

			if err := idx.IndexDirectory(cmd.Context(), rootPath, incremental); err != nil {
				return fmt.Errorf("indexing failed: %w", err)
			}

			fmt.Fprintln(os.Stderr, "[plaesy search index] Indexing complete")
			return nil
		},
	}

	cmd.Flags().StringVar(&repoPath, "path", ".", "directory to index")
	cmd.Flags().StringVar(&outDir, "outdir", ".plaesy/search", "output directory for index")
	cmd.Flags().BoolVar(&incremental, "incremental", true, "only index changed files")
	cmd.Flags().BoolVar(&force, "force", false, "rebuild index from scratch")
	cmd.Flags().BoolVar(&resume, "resume", true, "resume from checkpoint")
	cmd.Flags().IntVar(&workers, "workers", 4, "number of worker goroutines")

	return cmd
}

func newSearchQueryCmd() *cobra.Command {
	var (
		repoPath    string
		outDir      string
		topK        int
		jsonOutput  bool
		source      string
		filter      string
		rerank      bool
		parentChild bool
	)

	cmd := &cobra.Command{
		Use:   "query [query text]",
		Short: "Search the local index",
		Long: `Search the local semantic search index using hybrid retrieval (vector + BM25).
Returns ranked results with citations.

Examples:
  plaesy search query "golang tutorial" --top=10
  plaesy search query "golang tutorial" --json
  plaesy search query "golang tutorial" --filter "source:markdown"`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("query text required: plaesy search query \"<text>\"")
			}
			query := args[0]

			// Load search config
			unifiedConfigMgr := config.NewConfigManager[config.UnifiedConfig](
				"", // Use default unified config path
				config.DefaultUnifiedConfig(),
			)
			unifiedConfig, err := unifiedConfigMgr.Load()
			if err != nil {
				return fmt.Errorf("loading unified config: %w", err)
			}

			if unifiedConfig.Search == nil {
				unifiedConfig.Search = config.DefaultSearchConfig()
			}
			searchConfig := unifiedConfig.Search

			// Parse filter
			filters := parseFilter(filter)

			// Create components
			emb, err := embedder.NewEmbedder(&searchConfig.Embedder)
			if err != nil {
				return fmt.Errorf("creating embedder: %w", err)
			}

			// Convert vector store config
			vectorStoreConfig := vectorstore.Config{
				Type: searchConfig.VectorStore.Type,
				HNSW: vectorstore.HNSWConfig{
					M:              searchConfig.VectorStore.HNSW.M,
					EFConstruction: searchConfig.VectorStore.HNSW.EFConstruction,
					EFSearch:       searchConfig.VectorStore.HNSW.EFSearch,
				},
			}

			vectorStore, err := vectorstore.NewChromemStore(
				vectorStoreConfig,
				func(ctx context.Context, text string) ([]float32, error) {
					embs, err := emb.Embed(ctx, []string{text})
					if err != nil {
						return nil, err
					}
					if len(embs) == 0 {
						return nil, fmt.Errorf("no embedding returned")
					}
					return embs[0], nil
				},
				outDir,
			)
			if err != nil {
				return fmt.Errorf("creating vector store: %w", err)
			}
			defer vectorStore.Close()

			// Convert text index config
			textIndexConfig := textindex.Config{
				Type: searchConfig.TextIndex.Type,
			}

			textIndex, err := textindex.NewBleveIndex(textIndexConfig, outDir)
			if err != nil {
				return fmt.Errorf("creating text index: %w", err)
			}
			defer textIndex.Close()

			docStore, err := store.NewDocumentStore(outDir)
			if err != nil {
				return fmt.Errorf("creating document store: %w", err)
			}

			// Create retriever
			retr := retriever.NewHybridRetriever(
				searchConfig.Retriever,
				emb,
				vectorStore,
				textIndex,
				docStore,
			)

			// Perform search
			results, err := retr.Retrieve(cmd.Context(), query, filters)
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}

			// Output results
			if jsonOutput {
				return outputJSON(results)
			}

			return outputTable(results)
		},
	}

	cmd.Flags().StringVar(&repoPath, "path", ".", "project directory")
	cmd.Flags().StringVar(&outDir, "outdir", ".plaesy/search", "index directory")
	cmd.Flags().IntVar(&topK, "top", 10, "max results")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	cmd.Flags().StringVar(&source, "source", "local", "search source (local)")
	cmd.Flags().StringVar(&filter, "filter", "", "filter (e.g., source:markdown,path:docs)")
	cmd.Flags().BoolVar(&rerank, "rerank", false, "enable reranker")
	cmd.Flags().BoolVar(&parentChild, "parent-child", true, "enable parent-child retrieval")

	return cmd
}

func cleanIndex(outDir string) error {
	dirs := []string{
		filepath.Join(outDir, "vectorstore"),
		filepath.Join(outDir, "textindex"),
		filepath.Join(outDir, "store"),
	}

	for _, dir := range dirs {
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
	}
	return nil
}

func parseFilter(filter string) map[string]string {
	filters := make(map[string]string)
	if filter == "" {
		return filters
	}

	parts := strings.Split(filter, ",")
	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), ":", 2)
		if len(kv) == 2 {
			filters[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return filters
}

func outputJSON(results []retriever.Result) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func outputTable(results []retriever.Result) error {
	if len(results) == 0 {
		fmt.Println("No results found")
		return nil
	}

	fmt.Printf("Found %d results:\n\n", len(results))
	for i, r := range results {
		fmt.Printf("%d. Score: %.4f (vec: %.4f, bm25: %.4f)\n", i+1, r.Score, r.VectorScore, r.BM25Score)
		fmt.Printf("   ID: %s\n", r.ID)
		if r.DocID != "" {
			fmt.Printf("   Doc: %s\n", r.DocID)
		}
		if r.SectionPath != "" {
			fmt.Printf("   Section: %s\n", r.SectionPath)
		}
		fmt.Printf("   Content: %s\n", truncate(r.Content, 200))
		fmt.Println()
	}
	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
