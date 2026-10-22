package eval

import (
	"math"
	"sort"
	"time"
)

// MetricsConfig holds configuration for metrics computation.
type MetricsConfig struct {
	KValues        []int     // k values for recall@k (e.g., [1, 5, 10, 20, 50])
	LatencyBuckets []float64 // Latency percentile buckets (e.g., [0.5, 0.9, 0.95, 0.99])
}

// DefaultMetricsConfig returns default metrics configuration.
func DefaultMetricsConfig() MetricsConfig {
	return MetricsConfig{
		KValues:        []int{1, 3, 5, 10, 20, 50},
		LatencyBuckets: []float64{0.5, 0.9, 0.95, 0.99},
	}
}

// RetrievalResult represents the result of a single query retrieval.
type RetrievalResult struct {
	QueryID      string        `json:"query_id"`
	Query        string        `json:"query"`
	RetrievedIDs []string      `json:"retrieved_ids"`
	Scores       []float64     `json:"scores,omitempty"` // Retrieval scores
	Latency      time.Duration `json:"latency_ms"`       // Latency in milliseconds
	Mode         string        `json:"mode"`             // "vector", "bm25", "hybrid", "reranked"
	Error        string        `json:"error,omitempty"`  // Error if any
}

// EvaluationResult holds the evaluation results for a single query.
type EvaluationResult struct {
	QueryID         string          `json:"query_id"`
	Query           string          `json:"query"`
	ExpectedIDs     []string        `json:"expected_ids"`
	RetrievedIDs    []string        `json:"retrieved_ids"`
	RecallAtK       map[int]float64 `json:"recall_at_k"`
	MRR             float64         `json:"mrr"`
	NDCGAtK         map[int]float64 `json:"ndcg_at_k"`
	LatencyMs       float64         `json:"latency_ms"`
	Mode            string          `json:"mode"`
	RelevanceScores []float64       `json:"relevance_scores,omitempty"`
	Error           string          `json:"error,omitempty"`
}

// AggregateMetrics holds aggregated metrics across all queries.
type AggregateMetrics struct {
	RecallAtK         map[int]float64              `json:"recall_at_k"`
	MRR               float64                      `json:"mrr"`
	NDCGAtK           map[int]float64              `json:"ndcg_at_k"`
	LatencyStats      LatencyStats                 `json:"latency_stats"`
	TotalQueries      int                          `json:"total_queries"`
	SuccessfulQueries int                          `json:"successful_queries"`
	FailedQueries     int                          `json:"failed_queries"`
	ByMode            map[string]*AggregateMetrics `json:"by_mode,omitempty"`
}

// LatencyStats holds latency percentile statistics.
type LatencyStats struct {
	Mean   float64 `json:"mean_ms"`
	Median float64 `json:"median_ms"`
	P50    float64 `json:"p50_ms"`
	P90    float64 `json:"p90_ms"`
	P95    float64 `json:"p95_ms"`
	P99    float64 `json:"p99_ms"`
	Min    float64 `json:"min_ms"`
	Max    float64 `json:"max_ms"`
	Count  int     `json:"count"`
}

// ComputeRecallAtK computes recall@k for a single query.
func ComputeRecallAtK(retrievedIDs, expectedIDs []string, k int) float64 {
	if k <= 0 || len(expectedIDs) == 0 {
		return 0.0
	}

	// Limit retrieved to top-k
	limit := min(k, len(retrievedIDs))
	topK := retrievedIDs[:limit]

	// Create set of expected IDs for O(1) lookup
	expectedSet := make(map[string]bool)
	for _, id := range expectedIDs {
		expectedSet[id] = true
	}

	// Count hits
	hits := 0
	for _, id := range topK {
		if expectedSet[id] {
			hits++
		}
	}

	return float64(hits) / float64(len(expectedIDs))
}

// ComputeMRR computes Mean Reciprocal Rank for a single query.
func ComputeMRR(retrievedIDs, expectedIDs []string) float64 {
	if len(expectedIDs) == 0 {
		return 0.0
	}

	expectedSet := make(map[string]bool)
	for _, id := range expectedIDs {
		expectedSet[id] = true
	}

	for rank, id := range retrievedIDs {
		if expectedSet[id] {
			return 1.0 / float64(rank+1)
		}
	}

	return 0.0
}

// ComputeNDCGAtK computes nDCG@k for a single query.
// If relevanceScores are provided, uses graded relevance; otherwise binary relevance.
func ComputeNDCGAtK(retrievedIDs, expectedIDs []string, relevanceScores []float64, k int) float64 {
	if k <= 0 || len(expectedIDs) == 0 {
		return 0.0
	}

	limit := min(k, len(retrievedIDs))

	// Build relevance map
	relevance := make(map[string]float64)
	if len(relevanceScores) == len(expectedIDs) {
		// Graded relevance
		for i, id := range expectedIDs {
			relevance[id] = relevanceScores[i]
		}
	} else {
		// Binary relevance
		for _, id := range expectedIDs {
			relevance[id] = 1.0
		}
	}

	// Compute DCG
	dcg := 0.0
	for i := 0; i < limit; i++ {
		if rel, ok := relevance[retrievedIDs[i]]; ok && rel > 0 {
			dcg += (math.Pow(2, rel) - 1) / math.Log2(float64(i+2))
		}
	}

	// Compute IDCG (ideal DCG)
	idealScores := make([]float64, 0, len(relevance))
	for _, score := range relevance {
		if score > 0 {
			idealScores = append(idealScores, score)
		}
	}
	sort.Sort(sort.Reverse(sort.Float64Slice(idealScores)))

	idcg := 0.0
	idealLimit := min(k, len(idealScores))
	for i := 0; i < idealLimit; i++ {
		idcg += (math.Pow(2, idealScores[i]) - 1) / math.Log2(float64(i+2))
	}

	if idcg == 0 {
		return 0.0
	}

	return dcg / idcg
}

// ComputeLatencyStats computes latency statistics from a slice of latencies.
func ComputeLatencyStats(latencies []time.Duration) LatencyStats {
	if len(latencies) == 0 {
		return LatencyStats{}
	}

	// Convert to milliseconds
	ms := make([]float64, len(latencies))
	for i, d := range latencies {
		ms[i] = float64(d.Milliseconds())
	}

	sort.Float64s(ms)

	sum := 0.0
	for _, v := range ms {
		sum += v
	}

	stats := LatencyStats{
		Count:  len(ms),
		Mean:   sum / float64(len(ms)),
		Min:    ms[0],
		Max:    ms[len(ms)-1],
		Median: percentile(ms, 0.5),
		P50:    percentile(ms, 0.5),
		P90:    percentile(ms, 0.9),
		P95:    percentile(ms, 0.95),
		P99:    percentile(ms, 0.99),
	}

	return stats
}

// percentile computes the p-th percentile (0 <= p <= 1) of sorted data.
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0.0
	}
	if p <= 0 {
		return sorted[0]
	}
	if p >= 1 {
		return sorted[len(sorted)-1]
	}

	// Linear interpolation
	index := p * float64(len(sorted)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sorted[lower]
	}

	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

// EvaluateQuery evaluates a single query's retrieval results.
func EvaluateQuery(result RetrievalResult, golden GoldenEntry, cfg MetricsConfig) EvaluationResult {
	recallAtK := make(map[int]float64)
	for _, k := range cfg.KValues {
		recallAtK[k] = ComputeRecallAtK(result.RetrievedIDs, golden.ExpectedIDs, k)
	}

	mrr := ComputeMRR(result.RetrievedIDs, golden.ExpectedIDs)

	ndcgAtK := make(map[int]float64)
	for _, k := range cfg.KValues {
		ndcgAtK[k] = ComputeNDCGAtK(result.RetrievedIDs, golden.ExpectedIDs, golden.RelevanceScores, k)
	}

	return EvaluationResult{
		QueryID:         result.QueryID,
		Query:           result.Query,
		ExpectedIDs:     golden.ExpectedIDs,
		RetrievedIDs:    result.RetrievedIDs,
		RecallAtK:       recallAtK,
		MRR:             mrr,
		NDCGAtK:         ndcgAtK,
		LatencyMs:       float64(result.Latency.Milliseconds()),
		Mode:            result.Mode,
		RelevanceScores: golden.RelevanceScores,
	}
}

// AggregateResults aggregates evaluation results across queries.
func AggregateResults(results []EvaluationResult, cfg MetricsConfig) AggregateMetrics {
	return aggregateResultsInternal(results, cfg, true)
}

// aggregateResultsInternal does the actual aggregation work.
// If includeByMode is false, it skips the per-mode breakdown to avoid recursion.
func aggregateResultsInternal(results []EvaluationResult, cfg MetricsConfig, includeByMode bool) AggregateMetrics {
	if len(results) == 0 {
		return AggregateMetrics{}
	}

	// Group by mode
	byMode := make(map[string][]EvaluationResult)
	successful := 0
	failed := 0
	var allLatencies []time.Duration

	for _, r := range results {
		if r.Mode != "" {
			byMode[r.Mode] = append(byMode[r.Mode], r)
		}
		if len(r.RetrievedIDs) > 0 || r.Error == "" {
			successful++
			allLatencies = append(allLatencies, time.Duration(r.LatencyMs)*time.Millisecond)
		} else {
			failed++
		}
	}

	// Compute aggregate metrics
	agg := AggregateMetrics{
		TotalQueries:      len(results),
		SuccessfulQueries: successful,
		FailedQueries:     failed,
		RecallAtK:         make(map[int]float64),
		NDCGAtK:           make(map[int]float64),
		LatencyStats:      ComputeLatencyStats(allLatencies),
		ByMode:            make(map[string]*AggregateMetrics),
	}

	// Sum up recall@k and MRR, NDCG
	mrrSum := 0.0
	for _, k := range cfg.KValues {
		sum := 0.0
		for _, r := range results {
			sum += r.RecallAtK[k]
		}
		agg.RecallAtK[k] = sum / float64(len(results))
	}

	for _, r := range results {
		mrrSum += r.MRR
	}
	agg.MRR = mrrSum / float64(len(results))

	for _, k := range cfg.KValues {
		sum := 0.0
		for _, r := range results {
			sum += r.NDCGAtK[k]
		}
		agg.NDCGAtK[k] = sum / float64(len(results))
	}

	// Per-mode aggregates
	if includeByMode {
		for mode, modeResults := range byMode {
			modeAgg := aggregateResultsInternal(modeResults, cfg, false)
			agg.ByMode[mode] = &modeAgg
		}
	}

	return agg
}

// CompareModes compares two modes and returns statistical significance.
func CompareModes(modeA, modeB *AggregateMetrics, metric string) ModeComparison {
	comp := ModeComparison{
		ModeA:  modeA,
		ModeB:  modeB,
		Metric: metric,
	}

	switch metric {
	case "recall@10":
		comp.ValueA = modeA.RecallAtK[10]
		comp.ValueB = modeB.RecallAtK[10]
	case "mrr":
		comp.ValueA = modeA.MRR
		comp.ValueB = modeB.MRR
	case "ndcg@10":
		comp.ValueA = modeA.NDCGAtK[10]
		comp.ValueB = modeB.NDCGAtK[10]
	case "latency_p50":
		comp.ValueA = modeA.LatencyStats.P50
		comp.ValueB = modeB.LatencyStats.P50
	case "latency_p95":
		comp.ValueA = modeA.LatencyStats.P95
		comp.ValueB = modeB.LatencyStats.P95
	}

	comp.Diff = comp.ValueB - comp.ValueA
	comp.PctChange = 0.0
	if comp.ValueA != 0 {
		comp.PctChange = (comp.Diff / comp.ValueA) * 100
	}

	return comp
}

// ModeComparison holds comparison results between two modes.
type ModeComparison struct {
	ModeA       *AggregateMetrics `json:"mode_a"`
	ModeB       *AggregateMetrics `json:"mode_b"`
	Metric      string            `json:"metric"`
	ValueA      float64           `json:"value_a"`
	ValueB      float64           `json:"value_b"`
	Diff        float64           `json:"diff"`
	PctChange   float64           `json:"pct_change"`
	Significant bool              `json:"significant"` // Placeholder for future statistical test
}
