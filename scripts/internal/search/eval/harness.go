package eval

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"time"
)

// SearchMode represents a retrieval mode to evaluate.
type SearchMode string

const (
	ModeVector   SearchMode = "vector"
	ModeBM25     SearchMode = "bm25"
	ModeHybrid   SearchMode = "hybrid"
	ModeReranked SearchMode = "reranked"
)

// SearchEngine defines the interface for a search engine to evaluate.
type SearchEngine interface {
	Search(query string, topK int, mode SearchMode) ([]string, []float64, error)
	Name() string
}

// HarnessConfig holds configuration for the evaluation harness.
type HarnessConfig struct {
	GoldenSetPath string        // Path to golden set JSONL
	OutputPath    string        // Path to output results (empty = stdout)
	OutputFormat  OutputFormat  // json, markdown, csv
	Modes         []SearchMode  // Modes to evaluate
	TopK          int           // Number of results to retrieve
	Timeout       time.Duration // Per-query timeout
	MetricsConfig MetricsConfig // Metrics computation config
	WarmupQueries int           // Number of warmup queries (not counted in metrics)
	Parallel      bool          // Run queries in parallel
	MaxWorkers    int           // Max parallel workers
	Seed          int64         // Random seed for reproducibility
}

// DefaultHarnessConfig returns default harness configuration.
func DefaultHarnessConfig() HarnessConfig {
	return HarnessConfig{
		OutputFormat:  FormatJSON,
		Modes:         []SearchMode{ModeVector, ModeBM25, ModeHybrid},
		TopK:          50,
		Timeout:       30 * time.Second,
		MetricsConfig: DefaultMetricsConfig(),
		WarmupQueries: 0,
		Parallel:      false,
		MaxWorkers:    4,
		Seed:          42,
	}
}

// OutputFormat represents the output format for results.
type OutputFormat string

const (
	FormatJSON     OutputFormat = "json"
	FormatMarkdown OutputFormat = "markdown"
	FormatCSV      OutputFormat = "csv"
)

// Harness runs evaluation of search engines against a golden set.
type Harness struct {
	config    HarnessConfig
	goldenSet *GoldenSet
	engines   map[SearchMode]SearchEngine
	results   []EvaluationResult
}

// NewHarness creates a new evaluation harness.
func NewHarness(cfg HarnessConfig) (*Harness, error) {
	gs, err := LoadGoldenSet(GoldenSetConfig{
		Path:          cfg.GoldenSetPath,
		MinEntries:    1,
		RequireScores: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load golden set: %w", err)
	}

	return &Harness{
		config:    cfg,
		goldenSet: gs,
		engines:   make(map[SearchMode]SearchEngine),
		results:   make([]EvaluationResult, 0),
	}, nil
}

// RegisterEngine registers a search engine for a specific mode.
func (h *Harness) RegisterEngine(mode SearchMode, engine SearchEngine) {
	h.engines[mode] = engine
}

// Run executes the evaluation harness.
func (h *Harness) Run() (*AggregateMetrics, error) {
	if len(h.engines) == 0 {
		return nil, fmt.Errorf("no search engines registered")
	}

	if len(h.goldenSet.Entries) == 0 {
		return nil, fmt.Errorf("golden set is empty")
	}

	// Warmup
	if h.config.WarmupQueries > 0 {
		h.runWarmup()
	}

	// Run evaluation for each mode
	for _, mode := range h.config.Modes {
		engine, ok := h.engines[mode]
		if !ok {
			return nil, fmt.Errorf("engine not registered for mode: %s", mode)
		}

		modeResults, err := h.runMode(engine, mode)
		if err != nil {
			return nil, fmt.Errorf("mode %s failed: %w", mode, err)
		}
		h.results = append(h.results, modeResults...)
	}

	// Aggregate results
	agg := AggregateResults(h.results, h.config.MetricsConfig)

	return &agg, nil
}

// runWarmup runs warmup queries to prime caches.
func (h *Harness) runWarmup() {
	warmupCount := min(h.config.WarmupQueries, len(h.goldenSet.Entries))
	for i := 0; i < warmupCount; i++ {
		entry := h.goldenSet.Entries[i]
		for _, engine := range h.engines {
			_, _, _ = engine.Search(entry.Query, h.config.TopK, SearchMode(engine.Name()))
		}
	}
}

// runMode runs evaluation for a single mode.
func (h *Harness) runMode(engine SearchEngine, mode SearchMode) ([]EvaluationResult, error) {
	entries := h.goldenSet.Entries
	results := make([]EvaluationResult, len(entries))

	if h.config.Parallel {
		return h.runModeParallel(engine, mode, entries)
	}

	for i, entry := range entries {
		result, err := h.runSingleQuery(engine, mode, entry)
		if err != nil {
			// Record error but continue
			results[i] = EvaluationResult{
				QueryID:   entry.ID,
				Query:     entry.Query,
				Mode:      string(mode),
				Error:     err.Error(),
				LatencyMs: 0,
			}
			continue
		}
		results[i] = result
	}

	return results, nil
}

// runModeParallel runs evaluation in parallel.
func (h *Harness) runModeParallel(engine SearchEngine, mode SearchMode, entries []GoldenEntry) ([]EvaluationResult, error) {
	results := make([]EvaluationResult, len(entries))
	sem := make(chan struct{}, h.config.MaxWorkers)
	errChan := make(chan error, 1)

	for i, entry := range entries {
		sem <- struct{}{}
		go func(idx int, e GoldenEntry) {
			defer func() { <-sem }()

			result, err := h.runSingleQuery(engine, mode, e)
			if err != nil {
				results[idx] = EvaluationResult{
					QueryID:   e.ID,
					Query:     e.Query,
					Mode:      string(mode),
					Error:     err.Error(),
					LatencyMs: 0,
				}
				return
			}
			results[idx] = result
		}(i, entry)
	}

	// Wait for all goroutines
	for i := 0; i < h.config.MaxWorkers; i++ {
		sem <- struct{}{}
	}

	select {
	case err := <-errChan:
		return nil, err
	default:
	}

	return results, nil
}

// runSingleQuery runs a single query and evaluates it.
func (h *Harness) runSingleQuery(engine SearchEngine, mode SearchMode, entry GoldenEntry) (EvaluationResult, error) {
	start := time.Now()

	retrievedIDs, scores, err := engine.Search(entry.Query, h.config.TopK, mode)
	latency := time.Since(start)

	if err != nil {
		return EvaluationResult{}, err
	}

	retResult := RetrievalResult{
		QueryID:      entry.ID,
		Query:        entry.Query,
		RetrievedIDs: retrievedIDs,
		Scores:       scores,
		Latency:      latency,
		Mode:         string(mode),
	}

	return EvaluateQuery(retResult, entry, h.config.MetricsConfig), nil
}

// WriteResults writes evaluation results to the configured output.
func (h *Harness) WriteResults(agg *AggregateMetrics) error {
	var w io.Writer
	if h.config.OutputPath == "" {
		w = os.Stdout
	} else {
		file, err := os.Create(h.config.OutputPath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		w = file
	}

	switch h.config.OutputFormat {
	case FormatJSON:
		return h.writeJSON(w, agg)
	case FormatMarkdown:
		return h.writeMarkdown(w, agg)
	case FormatCSV:
		return h.writeCSV(w, agg)
	default:
		return fmt.Errorf("unsupported output format: %s", h.config.OutputFormat)
	}
}

// writeJSON writes results as JSON.
func (h *Harness) writeJSON(w io.Writer, agg *AggregateMetrics) error {
	output := map[string]interface{}{
		"config":     h.config,
		"golden_set": h.goldenSet.Stats(),
		"aggregate":  agg,
		"details":    h.results,
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// writeMarkdown writes results as Markdown.
func (h *Harness) writeMarkdown(w io.Writer, agg *AggregateMetrics) error {
	fmt.Fprintf(w, "# Evaluation Results\n\n")
	fmt.Fprintf(w, "**Golden Set:** %s (%d queries)\n\n", h.config.GoldenSetPath, h.goldenSet.Stats().TotalQueries)
	fmt.Fprintf(w, "**Modes Evaluated:** %v\n\n", h.config.Modes)
	fmt.Fprintf(w, "**Top-K:** %d\n\n", h.config.TopK)

	// Aggregate table
	fmt.Fprintf(w, "## Aggregate Metrics\n\n")
	fmt.Fprintf(w, "| Metric | Value |\n")
	fmt.Fprintf(w, "|--------|-------|\n")
	fmt.Fprintf(w, "| Total Queries | %d |\n", agg.TotalQueries)
	fmt.Fprintf(w, "| Successful | %d |\n", agg.SuccessfulQueries)
	fmt.Fprintf(w, "| Failed | %d |\n", agg.FailedQueries)
	fmt.Fprintf(w, "| MRR | %.4f |\n", agg.MRR)

	for _, k := range h.config.MetricsConfig.KValues {
		fmt.Fprintf(w, "| Recall@%d | %.4f |\n", k, agg.RecallAtK[k])
		fmt.Fprintf(w, "| nDCG@%d | %.4f |\n", k, agg.NDCGAtK[k])
	}

	fmt.Fprintf(w, "\n## Latency Statistics\n\n")
	fmt.Fprintf(w, "| Percentile | Latency (ms) |\n")
	fmt.Fprintf(w, "|------------|--------------|\n")
	fmt.Fprintf(w, "| Mean | %.2f |\n", agg.LatencyStats.Mean)
	fmt.Fprintf(w, "| Median | %.2f |\n", agg.LatencyStats.Median)
	fmt.Fprintf(w, "| P50 | %.2f |\n", agg.LatencyStats.P50)
	fmt.Fprintf(w, "| P90 | %.2f |\n", agg.LatencyStats.P90)
	fmt.Fprintf(w, "| P95 | %.2f |\n", agg.LatencyStats.P95)
	fmt.Fprintf(w, "| P99 | %.2f |\n", agg.LatencyStats.P99)
	fmt.Fprintf(w, "| Min | %.2f |\n", agg.LatencyStats.Min)
	fmt.Fprintf(w, "| Max | %.2f |\n", agg.LatencyStats.Max)

	// Per-mode tables
	if len(agg.ByMode) > 0 {
		fmt.Fprintf(w, "\n## Per-Mode Breakdown\n\n")

		// Sort modes for consistent output
		modes := make([]string, 0, len(agg.ByMode))
		for mode := range agg.ByMode {
			modes = append(modes, mode)
		}
		sort.Strings(modes)

		for _, mode := range modes {
			m := agg.ByMode[mode]
			fmt.Fprintf(w, "### Mode: %s\n\n", mode)
			fmt.Fprintf(w, "| Metric | Value |\n")
			fmt.Fprintf(w, "|--------|-------|\n")
			fmt.Fprintf(w, "| Queries | %d |\n", m.TotalQueries)
			fmt.Fprintf(w, "| MRR | %.4f |\n", m.MRR)
			for _, k := range h.config.MetricsConfig.KValues {
				fmt.Fprintf(w, "| Recall@%d | %.4f |\n", k, m.RecallAtK[k])
				fmt.Fprintf(w, "| nDCG@%d | %.4f |\n", k, m.NDCGAtK[k])
			}
			fmt.Fprintf(w, "| Latency P50 | %.2f ms |\n", m.LatencyStats.P50)
			fmt.Fprintf(w, "| Latency P95 | %.2f ms |\n", m.LatencyStats.P95)
			fmt.Fprintf(w, "\n")
		}

		// Mode comparison
		if len(modes) >= 2 {
			fmt.Fprintf(w, "## Mode Comparisons (vs first mode)\n\n")
			baseMode := modes[0]
			baseAgg := agg.ByMode[baseMode]

			for _, mode := range modes[1:] {
				m := agg.ByMode[mode]
				fmt.Fprintf(w, "### %s vs %s\n\n", mode, baseMode)
				fmt.Fprintf(w, "| Metric | %s | %s | Diff | %% Change |\n", baseMode, mode)
				fmt.Fprintf(w, "|--------|------|------|------|----------|\n")

				comparisons := []struct {
					name string
					valA float64
					valB float64
				}{
					{"MRR", baseAgg.MRR, m.MRR},
					{"Recall@10", baseAgg.RecallAtK[10], m.RecallAtK[10]},
					{"nDCG@10", baseAgg.NDCGAtK[10], m.NDCGAtK[10]},
					{"Latency P50 (ms)", baseAgg.LatencyStats.P50, m.LatencyStats.P50},
					{"Latency P95 (ms)", baseAgg.LatencyStats.P95, m.LatencyStats.P95},
				}

				for _, c := range comparisons {
					diff := c.valB - c.valA
					pct := 0.0
					if c.valA != 0 {
						pct = (diff / c.valA) * 100
					}
					fmt.Fprintf(w, "| %s | %.4f | %.4f | %.4f | %.1f%% |\n", c.name, c.valA, c.valB, diff, pct)
				}
				fmt.Fprintf(w, "\n")
			}
		}
	}

	return nil
}

// writeCSV writes results as CSV.
func (h *Harness) writeCSV(w io.Writer, agg *AggregateMetrics) error {
	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	// Header
	header := []string{"query_id", "query", "mode", "mrr", "latency_ms"}
	for _, k := range h.config.MetricsConfig.KValues {
		header = append(header, fmt.Sprintf("recall@%d", k))
		header = append(header, fmt.Sprintf("ndcg@%d", k))
	}
	if err := csvWriter.Write(header); err != nil {
		return err
	}

	// Data rows
	for _, r := range h.results {
		row := []string{
			r.QueryID,
			r.Query,
			r.Mode,
			fmt.Sprintf("%.6f", r.MRR),
			fmt.Sprintf("%.2f", r.LatencyMs),
		}
		for _, k := range h.config.MetricsConfig.KValues {
			row = append(row, fmt.Sprintf("%.6f", r.RecallAtK[k]))
			row = append(row, fmt.Sprintf("%.6f", r.NDCGAtK[k]))
		}
		if err := csvWriter.Write(row); err != nil {
			return err
		}
	}

	// Summary row
	summaryRow := []string{"AGGREGATE", "", "", fmt.Sprintf("%.6f", agg.MRR), fmt.Sprintf("%.2f", agg.LatencyStats.Mean)}
	for _, k := range h.config.MetricsConfig.KValues {
		summaryRow = append(summaryRow, fmt.Sprintf("%.6f", agg.RecallAtK[k]))
		summaryRow = append(summaryRow, fmt.Sprintf("%.6f", agg.NDCGAtK[k]))
	}
	return csvWriter.Write(summaryRow)
}

// GetResults returns the raw evaluation results.
func (h *Harness) GetResults() []EvaluationResult {
	return h.results
}

// GetGoldenSet returns the golden set used.
func (h *Harness) GetGoldenSet() *GoldenSet {
	return h.goldenSet
}
