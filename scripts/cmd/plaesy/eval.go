package main

import (
	"fmt"
	"os"
	"time"

	"github.com/plaesy/spec-kit/internal/search/eval"
	"github.com/spf13/cobra"
)

func init() { register(newEvalCmd()) }

func newEvalCmd() *cobra.Command {
	var (
		goldenSetPath string
		outputPath    string
		outputFormat  string
		modesStr      string
		topK          int
		timeout       string
		warmup        int
		parallel      bool
		workers       int
		kValuesStr    string
		seed          int64
		projectPath   string
		indexDir      string
	)

	cmd := &cobra.Command{
		Use:   "eval",
		Short: "Run search evaluation harness against a golden set",
		Long: `Run search evaluation harness to measure retrieval quality (Recall@k, MRR, nDCG)
and latency across different search modes (vector, bm25, hybrid, reranked).

The golden set is a JSONL file where each line contains a query with expected results:

  {"query": "How to implement binary search", "expected_ids": ["doc1", "doc2"], "relevance_scores": [1.0, 0.8], "id": "q1"}

Examples:
  plaesy eval --golden-set golden.jsonl --format markdown --output results.md
  plaesy eval --golden-set golden.jsonl --modes vector,hybrid --parallel --workers 8
  plaesy eval --golden-set golden.jsonl --k-values 1,5,10,20 --top-k 50`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if goldenSetPath == "" {
				return fmt.Errorf("--golden-set is required")
			}

			// Parse modes
			modes := []eval.SearchMode{}
			for _, m := range splitAndTrim(modesStr, ",") {
				switch m {
				case "vector":
					modes = append(modes, eval.ModeVector)
				case "bm25":
					modes = append(modes, eval.ModeBM25)
				case "hybrid":
					modes = append(modes, eval.ModeHybrid)
				case "reranked":
					modes = append(modes, eval.ModeReranked)
				default:
					return fmt.Errorf("unknown mode: %s (valid: vector, bm25, hybrid, reranked)", m)
				}
			}

			// Parse k-values
			kValues := []int{}
			for _, kv := range splitAndTrim(kValuesStr, ",") {
				var k int
				if _, err := fmt.Sscanf(kv, "%d", &k); err == nil && k > 0 {
					kValues = append(kValues, k)
				}
			}
			if len(kValues) == 0 {
				kValues = []int{1, 3, 5, 10, 20, 50}
			}

			// Parse timeout
			timeoutDur, err := time.ParseDuration(timeout)
			if err != nil {
				return fmt.Errorf("invalid timeout: %w", err)
			}

			cfg := eval.HarnessConfig{
				GoldenSetPath: goldenSetPath,
				OutputPath:    outputPath,
				OutputFormat:  eval.OutputFormat(outputFormat),
				Modes:         modes,
				TopK:          topK,
				Timeout:       timeoutDur,
				WarmupQueries: warmup,
				Parallel:      parallel,
				MaxWorkers:    workers,
				Seed:          seed,
				MetricsConfig: eval.MetricsConfig{
					KValues: kValues,
				},
			}

			harness, err := eval.NewHarness(cfg)
			if err != nil {
				return fmt.Errorf("failed to create harness: %w", err)
			}

			// Measure the real index built by `plaesy analyze`. One engine
			// dispatches on mode, so it serves every registered mode.
			engine, closeIndex, err := eval.OpenIndexEngine(projectPath, indexDir)
			if err != nil {
				return fmt.Errorf("opening index (build it first with `plaesy analyze`): %w", err)
			}
			defer closeIndex()
			for _, mode := range modes {
				harness.RegisterEngine(mode, engine)
			}

			fmt.Fprintf(os.Stderr, "Running evaluation with %d queries, modes: %v\n", len(harness.GetGoldenSet().Entries), modes)

			agg, err := harness.Run()
			if err != nil {
				return fmt.Errorf("evaluation failed: %w", err)
			}

			if err := harness.WriteResults(agg); err != nil {
				return fmt.Errorf("failed to write results: %w", err)
			}

			fmt.Fprintf(os.Stderr, "Evaluation complete. MRR: %.4f, Recall@10: %.4f\n", agg.MRR, agg.RecallAtK[10])

			return nil
		},
	}

	cmd.Flags().StringVar(&goldenSetPath, "golden-set", "", "Path to golden set JSONL file (required)")
	cmd.Flags().StringVar(&outputPath, "output", "", "Output file path (default: stdout)")
	cmd.Flags().StringVar(&outputFormat, "format", "json", "Output format: json, markdown, csv")
	cmd.Flags().StringVar(&modesStr, "modes", "vector,bm25,hybrid", "Comma-separated modes to evaluate")
	cmd.Flags().IntVar(&topK, "top-k", 50, "Number of results to retrieve per query")
	cmd.Flags().StringVar(&timeout, "timeout", "30s", "Per-query timeout (e.g. 30s, 1m)")
	cmd.Flags().IntVar(&warmup, "warmup", 0, "Number of warmup queries")
	cmd.Flags().BoolVar(&parallel, "parallel", false, "Run queries in parallel")
	cmd.Flags().IntVar(&workers, "workers", 4, "Max parallel workers")
	cmd.Flags().StringVar(&kValuesStr, "k-values", "1,3,5,10,20,50", "Comma-separated k values for recall@k")
	cmd.Flags().Int64Var(&seed, "seed", 42, "Random seed")
	cmd.Flags().StringVar(&projectPath, "path", ".", "Project directory the index was built from")
	cmd.Flags().StringVar(&indexDir, "index", ".plaesy/search", "Index directory built by `plaesy analyze`")

	return cmd
}

func splitAndTrim(s, sep string) []string {
	parts := []string{}
	for _, p := range split(s, sep) {
		trimmed := trim(p)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func split(s, sep string) []string {
	result := []string{}
	start := 0
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

func trim(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
