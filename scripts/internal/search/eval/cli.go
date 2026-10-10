package eval

import (
	"flag"
	"fmt"
	"os"
	"time"
)

// Duration is a wrapper for time.Duration that can be parsed from string.
type Duration time.Duration

// ParseDuration parses a duration string.
func ParseDuration(s string) (Duration, error) {
	d, err := time.ParseDuration(s)
	return Duration(d), err
}

// RunCLI runs the evaluation harness from command line.
func RunCLI(args []string) error {
	fs := flag.NewFlagSet("plaesy eval", flag.ExitOnError)

	goldenSetPath := fs.String("golden-set", "", "Path to golden set JSONL file (required)")
	outputPath := fs.String("output", "", "Output file path (default: stdout)")
	outputFormat := fs.String("format", "json", "Output format: json, markdown, csv")
	modesStr := fs.String("modes", "vector,bm25,hybrid", "Comma-separated modes to evaluate")
	topK := fs.Int("top-k", 50, "Number of results to retrieve per query")
	timeout := fs.Duration("timeout", 30*time.Second, "Per-query timeout")
	warmup := fs.Int("warmup", 0, "Number of warmup queries")
	parallel := fs.Bool("parallel", false, "Run queries in parallel")
	workers := fs.Int("workers", 4, "Max parallel workers")
	kValuesStr := fs.String("k-values", "1,3,5,10,20,50", "Comma-separated k values for recall@k")
	seed := fs.Int64("seed", 42, "Random seed")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `plaesy eval - Run search evaluation harness

Usage:
  plaesy eval [flags]

Flags:
`)
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Golden Set Format (JSONL):
  Each line is a JSON object with:
  {
    "query": "search query",
    "expected_ids": ["doc1", "doc2"],
    "relevance_scores": [1.0, 0.8],  // optional, for nDCG
    "id": "q1",                      // optional
    "metadata": {"category": "tech"} // optional
  }

Example:
  plaesy eval --golden-set golden.jsonl --format markdown --output results.md
  plaesy eval --golden-set golden.jsonl --modes vector,hybrid --parallel --workers 8
`)
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *goldenSetPath == "" {
		fmt.Fprintln(os.Stderr, "Error: --golden-set is required")
		fs.Usage()
		os.Exit(1)
	}

	// Parse modes
	modes := []SearchMode{}
	for _, m := range splitAndTrim(*modesStr, ",") {
		switch m {
		case "vector":
			modes = append(modes, ModeVector)
		case "bm25":
			modes = append(modes, ModeBM25)
		case "hybrid":
			modes = append(modes, ModeHybrid)
		case "reranked":
			modes = append(modes, ModeReranked)
		default:
			fmt.Fprintf(os.Stderr, "Unknown mode: %s\n", m)
			os.Exit(1)
		}
	}

	// Parse k-values
	kValues := []int{}
	for _, kv := range splitAndTrim(*kValuesStr, ",") {
		var k int
		if _, err := fmt.Sscanf(kv, "%d", &k); err == nil && k > 0 {
			kValues = append(kValues, k)
		}
	}
	if len(kValues) == 0 {
		kValues = []int{1, 3, 5, 10, 20, 50}
	}

	cfg := HarnessConfig{
		GoldenSetPath: *goldenSetPath,
		OutputPath:    *outputPath,
		OutputFormat:  OutputFormat(*outputFormat),
		Modes:         modes,
		TopK:          *topK,
		Timeout:       *timeout,
		WarmupQueries: *warmup,
		Parallel:      *parallel,
		MaxWorkers:    *workers,
		Seed:          *seed,
		MetricsConfig: MetricsConfig{
			KValues: kValues,
		},
	}

	harness, err := NewHarness(cfg)
	if err != nil {
		return fmt.Errorf("failed to create harness: %w", err)
	}

	// Register mock engines for now (replace with real retriever adapter)
	for _, mode := range modes {
		mock := NewMockSearchEngine(fmt.Sprintf("mock-%s", mode), mode)
		// Add some default mock data for testing
		mock.SetFixedResults("test query", []string{"doc1", "doc2", "doc3"})
		harness.RegisterEngine(mode, mock)
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
	// Simple split implementation
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
	// Trim whitespace
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
