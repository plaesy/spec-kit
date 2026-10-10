package eval

import (
	"os"
	"testing"
	"time"
)

func TestGoldenSet_Load(t *testing.T) {
	// Create a temporary golden set file
	content := `{"query": "test query 1", "expected_ids": ["doc1", "doc2"], "relevance_scores": [1.0, 0.8], "id": "q1"}
{"query": "test query 2", "expected_ids": ["doc3"], "id": "q2"}
{"query": "test query 3", "expected_ids": ["doc4", "doc5", "doc6"], "relevance_scores": [1.0, 0.9, 0.7]}
`
	tmpFile, err := os.CreateTemp("", "golden_*.jsonl")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	gs, err := LoadGoldenSet(GoldenSetConfig{
		Path:          tmpFile.Name(),
		MinEntries:    1,
		RequireScores: false,
	})
	if err != nil {
		t.Fatalf("LoadGoldenSet failed: %v", err)
	}

	if len(gs.Entries) != 3 {
		t.Fatalf("Expected 3 entries, got %d", len(gs.Entries))
	}

	if gs.Entries[0].Query != "test query 1" {
		t.Errorf("Entry 0 query mismatch: %s", gs.Entries[0].Query)
	}
	if len(gs.Entries[0].ExpectedIDs) != 2 {
		t.Errorf("Entry 0 expected 2 IDs, got %d", len(gs.Entries[0].ExpectedIDs))
	}
	if len(gs.Entries[0].RelevanceScores) != 2 {
		t.Errorf("Entry 0 expected 2 scores, got %d", len(gs.Entries[0].RelevanceScores))
	}
	if gs.Entries[0].ID != "q1" {
		t.Errorf("Entry 0 ID mismatch: %s", gs.Entries[0].ID)
	}

	// Entry 1 has no relevance scores
	if len(gs.Entries[1].RelevanceScores) != 0 {
		t.Errorf("Entry 1 expected 0 scores, got %d", len(gs.Entries[1].RelevanceScores))
	}
	if gs.Entries[1].ID != "q2" {
		t.Errorf("Entry 1 ID mismatch: %s", gs.Entries[1].ID)
	}

	// Entry 2 has no ID, should be auto-generated
	if gs.Entries[2].ID == "" {
		t.Error("Entry 2 ID should be auto-generated")
	}
}

func TestGoldenSet_Load_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
	}{
		{
			name:        "empty query",
			content:     `{"query": "", "expected_ids": ["doc1"]}`,
			expectError: true,
		},
		{
			name:        "no expected ids",
			content:     `{"query": "test", "expected_ids": []}`,
			expectError: true,
		},
		{
			name:        "duplicate expected ids",
			content:     `{"query": "test", "expected_ids": ["doc1", "doc1"]}`,
			expectError: true,
		},
		{
			name:        "score length mismatch",
			content:     `{"query": "test", "expected_ids": ["doc1", "doc2"], "relevance_scores": [1.0]}`,
			expectError: true,
		},
		{
			name:        "score out of range",
			content:     `{"query": "test", "expected_ids": ["doc1"], "relevance_scores": [1.5]}`,
			expectError: true,
		},
		{
			name:        "negative score",
			content:     `{"query": "test", "expected_ids": ["doc1"], "relevance_scores": [-0.1]}`,
			expectError: true,
		},
		{
			name:        "valid entry",
			content:     `{"query": "test", "expected_ids": ["doc1"]}`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "golden_*.jsonl")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.WriteString(tt.content); err != nil {
				t.Fatalf("Failed to write temp file: %v", err)
			}
			tmpFile.Close()

			_, err = LoadGoldenSet(GoldenSetConfig{
				Path:          tmpFile.Name(),
				MinEntries:    1,
				RequireScores: false,
			})

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestGoldenSet_Load_RequireScores(t *testing.T) {
	content := `{"query": "test", "expected_ids": ["doc1"]}
`
	tmpFile, err := os.CreateTemp("", "golden_*.jsonl")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Should fail when RequireScores is true
	_, err = LoadGoldenSet(GoldenSetConfig{
		Path:          tmpFile.Name(),
		MinEntries:    1,
		RequireScores: true,
	})
	if err == nil {
		t.Error("Expected error when RequireScores=true but no scores provided")
	}

	// Should pass when RequireScores is false
	_, err = LoadGoldenSet(GoldenSetConfig{
		Path:             tmpFile.Name(),
		MinEntries:       1,
		RequireScores:    false,
		AllowEmptyScores: true,
	})
	if err != nil {
		t.Errorf("Unexpected error when RequireScores=false: %v", err)
	}
}

func TestGoldenSet_SaveAndLoad(t *testing.T) {
	original := &GoldenSet{
		Version:   "1.0",
		CreatedAt: "2024-01-01T00:00:00Z",
		Entries: []GoldenEntry{
			{Query: "q1", ExpectedIDs: []string{"a", "b"}, RelevanceScores: []float64{1.0, 0.5}, ID: "q1"},
			{Query: "q2", ExpectedIDs: []string{"c"}, ID: "q2"},
		},
	}

	tmpFile, err := os.CreateTemp("", "golden_*.jsonl")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	if err := SaveGoldenSet(original, tmpFile.Name()); err != nil {
		t.Fatalf("SaveGoldenSet failed: %v", err)
	}

	loaded, err := LoadGoldenSet(GoldenSetConfig{
		Path:          tmpFile.Name(),
		MinEntries:    1,
		RequireScores: false,
	})
	if err != nil {
		t.Fatalf("LoadGoldenSet failed: %v", err)
	}

	if len(loaded.Entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(loaded.Entries))
	}
	if loaded.Entries[0].Query != "q1" {
		t.Errorf("Entry 0 query mismatch")
	}
	if loaded.Entries[1].Query != "q2" {
		t.Errorf("Entry 1 query mismatch")
	}
}

func TestGoldenSet_Stats(t *testing.T) {
	gs := &GoldenSet{
		Entries: []GoldenEntry{
			{Query: "q1", ExpectedIDs: []string{"a", "b"}, RelevanceScores: []float64{1.0, 0.5}},
			{Query: "q2", ExpectedIDs: []string{"c"}},
			{Query: "q3", ExpectedIDs: []string{"d", "e", "f"}, RelevanceScores: []float64{1.0, 0.8, 0.6}},
		},
	}

	stats := gs.Stats()
	if stats.TotalQueries != 3 {
		t.Errorf("TotalQueries: expected 3, got %d", stats.TotalQueries)
	}
	if stats.TotalExpectedIDs != 6 {
		t.Errorf("TotalExpectedIDs: expected 6, got %d", stats.TotalExpectedIDs)
	}
	if stats.WithScores != 2 {
		t.Errorf("WithScores: expected 2, got %d", stats.WithScores)
	}
	if stats.AvgExpectedPerQuery != 2.0 {
		t.Errorf("AvgExpectedPerQuery: expected 2.0, got %.2f", stats.AvgExpectedPerQuery)
	}
}

func TestGoldenSet_FilterEntries(t *testing.T) {
	gs := &GoldenSet{
		Entries: []GoldenEntry{
			{Query: "go test", ExpectedIDs: []string{"a"}, ID: "1"},
			{Query: "python test", ExpectedIDs: []string{"b"}, ID: "2"},
			{Query: "go benchmark", ExpectedIDs: []string{"c"}, ID: "3"},
		},
	}

	filtered := gs.FilterEntries(func(e GoldenEntry) bool {
		return len(e.Query) > 0 && e.Query[:2] == "go"
	})

	if len(filtered) != 2 {
		t.Errorf("Expected 2 filtered entries, got %d", len(filtered))
	}
}

func TestGoldenSet_MergeGoldenSets(t *testing.T) {
	gs1 := &GoldenSet{Entries: []GoldenEntry{{Query: "q1", ExpectedIDs: []string{"a"}, ID: "1"}}}
	gs2 := &GoldenSet{Entries: []GoldenEntry{{Query: "q2", ExpectedIDs: []string{"b"}, ID: "2"}}}
	gs3 := &GoldenSet{Entries: []GoldenEntry{{Query: "q1", ExpectedIDs: []string{"a"}, ID: "1"}}} // duplicate

	merged := MergeGoldenSets([]*GoldenSet{gs1, gs2, gs3})
	if len(merged.Entries) != 2 {
		t.Errorf("Expected 2 unique entries after merge, got %d", len(merged.Entries))
	}
}

func TestMetrics_RecallAtK(t *testing.T) {
	tests := []struct {
		name           string
		retrieved      []string
		expected       []string
		k              int
		expectedRecall float64
	}{
		{"perfect match at k=3", []string{"a", "b", "c"}, []string{"a", "b", "c"}, 3, 1.0},
		{"partial match at k=3", []string{"a", "x", "b"}, []string{"a", "b", "c"}, 3, 2.0 / 3.0},
		{"no match at k=3", []string{"x", "y", "z"}, []string{"a", "b", "c"}, 3, 0.0},
		{"k larger than retrieved", []string{"a", "b"}, []string{"a", "b", "c"}, 5, 2.0 / 3.0},
		{"k=1 perfect", []string{"a", "b", "c"}, []string{"a", "b", "c"}, 1, 1.0 / 3.0},
		{"empty expected", []string{"a", "b"}, []string{}, 3, 0.0},
		{"empty retrieved", []string{}, []string{"a", "b"}, 3, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recall := ComputeRecallAtK(tt.retrieved, tt.expected, tt.k)
			if recall != tt.expectedRecall {
				t.Errorf("Recall@%d: expected %.4f, got %.4f", tt.k, tt.expectedRecall, recall)
			}
		})
	}
}

func TestMetrics_MRR(t *testing.T) {
	tests := []struct {
		name        string
		retrieved   []string
		expected    []string
		expectedMRR float64
	}{
		{"first result correct", []string{"a", "b", "c"}, []string{"a", "b"}, 1.0},
		{"second result correct", []string{"x", "a", "b"}, []string{"a", "b"}, 0.5},
		{"third result correct", []string{"x", "y", "a"}, []string{"a", "b"}, 1.0 / 3.0},
		{"no match", []string{"x", "y", "z"}, []string{"a", "b"}, 0.0},
		{"empty expected", []string{"a", "b"}, []string{}, 0.0},
		{"empty retrieved", []string{}, []string{"a", "b"}, 0.0},
		{"multiple expected, first match", []string{"a", "b"}, []string{"a", "c", "d"}, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mrr := ComputeMRR(tt.retrieved, tt.expected)
			if mrr != tt.expectedMRR {
				t.Errorf("MRR: expected %.4f, got %.4f", tt.expectedMRR, mrr)
			}
		})
	}
}

func TestMetrics_NDCG(t *testing.T) {
	tests := []struct {
		name            string
		retrieved       []string
		expected        []string
		relevanceScores []float64
		k               int
		expectedNDCG    float64
	}{
		{
			name:            "perfect ranking with graded relevance",
			retrieved:       []string{"a", "b", "c"},
			expected:        []string{"a", "b", "c"},
			relevanceScores: []float64{1.0, 0.8, 0.6},
			k:               3,
			expectedNDCG:    1.0,
		},
		{
			name:            "swapped top 2 with graded relevance",
			retrieved:       []string{"b", "a", "c"},
			expected:        []string{"a", "b", "c"},
			relevanceScores: []float64{1.0, 0.8, 0.6},
			k:               3,
			expectedNDCG:    0.95, // approximate
		},
		{
			name:            "binary relevance perfect",
			retrieved:       []string{"a", "b", "c"},
			expected:        []string{"a", "b", "c"},
			relevanceScores: []float64{},
			k:               3,
			expectedNDCG:    1.0,
		},
		{
			name:            "binary relevance partial",
			retrieved:       []string{"a", "x", "b"},
			expected:        []string{"a", "b", "c"},
			relevanceScores: []float64{},
			k:               3,
			expectedNDCG:    0.70, // DCG = 1 + 0 + 1/log2(4) = 1.5, IDCG = 1 + 1/log2(3) + 1/log2(4) ≈ 1 + 0.63 + 0.5 = 2.13, nDCG = 1.5/2.13 ≈ 0.70
		},
		{
			name:            "no relevant retrieved",
			retrieved:       []string{"x", "y", "z"},
			expected:        []string{"a", "b", "c"},
			relevanceScores: []float64{},
			k:               3,
			expectedNDCG:    0.0,
		},
		{
			name:            "k=1",
			retrieved:       []string{"a", "b"},
			expected:        []string{"a", "b"},
			relevanceScores: []float64{1.0, 0.5},
			k:               1,
			expectedNDCG:    1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ndcg := ComputeNDCGAtK(tt.retrieved, tt.expected, tt.relevanceScores, tt.k)
			// Allow small floating point differences
			if ndcg < tt.expectedNDCG-0.02 || ndcg > tt.expectedNDCG+0.02 {
				t.Errorf("nDCG@%d: expected ~%.4f, got %.4f", tt.k, tt.expectedNDCG, ndcg)
			}
		})
	}
}

func TestMetrics_LatencyPercentiles(t *testing.T) {
	latencies := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		40 * time.Millisecond,
		50 * time.Millisecond,
		60 * time.Millisecond,
		70 * time.Millisecond,
		80 * time.Millisecond,
		90 * time.Millisecond,
		100 * time.Millisecond,
	}

	stats := ComputeLatencyStats(latencies)

	if stats.Count != 10 {
		t.Errorf("Count: expected 10, got %d", stats.Count)
	}
	if stats.Min != 10 {
		t.Errorf("Min: expected 10, got %.2f", stats.Min)
	}
	if stats.Max != 100 {
		t.Errorf("Max: expected 100, got %.2f", stats.Max)
	}
	if stats.Mean != 55 {
		t.Errorf("Mean: expected 55, got %.2f", stats.Mean)
	}
	if stats.Median != 55 {
		t.Errorf("Median: expected 55, got %.2f", stats.Median)
	}
	if stats.P50 != 55 {
		t.Errorf("P50: expected 55, got %.2f", stats.P50)
	}
	// P90 should be 91 (90th percentile of 10 items)
	if stats.P90 < 90 || stats.P90 > 92 {
		t.Errorf("P90: expected ~91, got %.2f", stats.P90)
	}
}

func TestMetrics_EmptyLatencies(t *testing.T) {
	stats := ComputeLatencyStats([]time.Duration{})
	if stats.Count != 0 {
		t.Errorf("Expected count 0 for empty latencies")
	}
}

func TestEvaluateQuery(t *testing.T) {
	golden := GoldenEntry{
		Query:           "test query",
		ExpectedIDs:     []string{"a", "b", "c"},
		RelevanceScores: []float64{1.0, 0.8, 0.6},
		ID:              "q1",
	}

	result := RetrievalResult{
		QueryID:      "q1",
		Query:        "test query",
		RetrievedIDs: []string{"a", "b", "x", "c"},
		Scores:       []float64{0.9, 0.8, 0.7, 0.6},
		Latency:      10 * time.Millisecond,
		Mode:         "hybrid",
	}

	eval := EvaluateQuery(result, golden, DefaultMetricsConfig())

	if eval.QueryID != "q1" {
		t.Errorf("QueryID mismatch")
	}
	if eval.MRR != 1.0 { // First result is correct
		t.Errorf("MRR: expected 1.0, got %.4f", eval.MRR)
	}
	if eval.RecallAtK[3] != 2.0/3.0 { // a, b in top 3
		t.Errorf("Recall@3: expected 0.667, got %.4f", eval.RecallAtK[3])
	}
	if eval.LatencyMs != 10 {
		t.Errorf("LatencyMs: expected 10, got %.2f", eval.LatencyMs)
	}
}

func TestHarness_Run(t *testing.T) {
	// Create a mock golden set
	content := `{"query": "test query 1", "expected_ids": ["doc1", "doc2"], "id": "q1"}
{"query": "test query 2", "expected_ids": ["doc3"], "id": "q2"}
`
	tmpFile, err := os.CreateTemp("", "golden_*.jsonl")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	cfg := HarnessConfig{
		GoldenSetPath: tmpFile.Name(),
		Modes:         []SearchMode{ModeVector},
		TopK:          10,
		MetricsConfig: DefaultMetricsConfig(),
	}

	harness, err := NewHarness(cfg)
	if err != nil {
		t.Fatalf("NewHarness failed: %v", err)
	}

	// Register mock engine
	mock := NewMockSearchEngine("mock-vector", ModeVector)
	mock.SetFixedResults("test query 1", []string{"doc1", "doc2", "doc3"})
	mock.SetFixedResults("test query 2", []string{"doc3", "doc4"})
	harness.RegisterEngine(ModeVector, mock)

	agg, err := harness.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if agg.TotalQueries != 2 {
		t.Errorf("TotalQueries: expected 2, got %d", agg.TotalQueries)
	}
	if agg.SuccessfulQueries != 2 {
		t.Errorf("SuccessfulQueries: expected 2, got %d", agg.SuccessfulQueries)
	}
	if agg.MRR != 1.0 {
		t.Errorf("MRR: expected 1.0, got %.4f", agg.MRR)
	}
}

func TestHarness_RunMultipleModes(t *testing.T) {
	content := `{"query": "test", "expected_ids": ["doc1"], "id": "q1"}
`
	tmpFile, err := os.CreateTemp("", "golden_*.jsonl")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	cfg := HarnessConfig{
		GoldenSetPath: tmpFile.Name(),
		Modes:         []SearchMode{ModeVector, ModeBM25, ModeHybrid},
		TopK:          10,
		MetricsConfig: DefaultMetricsConfig(),
	}

	harness, err := NewHarness(cfg)
	if err != nil {
		t.Fatalf("NewHarness failed: %v", err)
	}

	for _, mode := range []SearchMode{ModeVector, ModeBM25, ModeHybrid} {
		mock := NewMockSearchEngine("mock-"+string(mode), mode)
		mock.SetFixedResults("test", []string{"doc1", "doc2"})
		harness.RegisterEngine(mode, mock)
	}

	agg, err := harness.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if agg.TotalQueries != 3 { // 1 query * 3 modes
		t.Errorf("TotalQueries: expected 3, got %d", agg.TotalQueries)
	}
	if len(agg.ByMode) != 3 {
		t.Errorf("ByMode: expected 3 modes, got %d", len(agg.ByMode))
	}
}

func TestHarness_OutputFormats(t *testing.T) {
	content := `{"query": "test", "expected_ids": ["doc1"], "id": "q1"}
`
	tmpFile, err := os.CreateTemp("", "golden_*.jsonl")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	cfg := HarnessConfig{
		GoldenSetPath: tmpFile.Name(),
		Modes:         []SearchMode{ModeVector},
		TopK:          10,
		MetricsConfig: DefaultMetricsConfig(),
		OutputFormat:  FormatJSON,
	}

	harness, err := NewHarness(cfg)
	if err != nil {
		t.Fatalf("NewHarness failed: %v", err)
	}

	mock := NewMockSearchEngine("mock-vector", ModeVector)
	mock.SetFixedResults("test", []string{"doc1"})
	harness.RegisterEngine(ModeVector, mock)

	agg, err := harness.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Test JSON output
	jsonFile, err := os.CreateTemp("", "output_*.json")
	if err != nil {
		t.Fatalf("Failed to create output file: %v", err)
	}
	defer os.Remove(jsonFile.Name())
	jsonFile.Close()

	cfg.OutputPath = jsonFile.Name()
	cfg.OutputFormat = FormatJSON
	harness.config = cfg

	if err := harness.WriteResults(agg); err != nil {
		t.Errorf("WriteResults JSON failed: %v", err)
	}

	// Test Markdown output
	mdFile, err := os.CreateTemp("", "output_*.md")
	if err != nil {
		t.Fatalf("Failed to create output file: %v", err)
	}
	defer os.Remove(mdFile.Name())
	mdFile.Close()

	cfg.OutputPath = mdFile.Name()
	cfg.OutputFormat = FormatMarkdown
	harness.config = cfg

	if err := harness.WriteResults(agg); err != nil {
		t.Errorf("WriteResults Markdown failed: %v", err)
	}

	// Test CSV output
	csvFile, err := os.CreateTemp("", "output_*.csv")
	if err != nil {
		t.Fatalf("Failed to create output file: %v", err)
	}
	defer os.Remove(csvFile.Name())
	csvFile.Close()

	cfg.OutputPath = csvFile.Name()
	cfg.OutputFormat = FormatCSV
	harness.config = cfg

	if err := harness.WriteResults(agg); err != nil {
		t.Errorf("WriteResults CSV failed: %v", err)
	}
}

func TestCompareModes(t *testing.T) {
	modeA := &AggregateMetrics{
		MRR:          0.8,
		RecallAtK:    map[int]float64{10: 0.7},
		NDCGAtK:      map[int]float64{10: 0.75},
		LatencyStats: LatencyStats{P50: 50, P95: 100},
	}

	modeB := &AggregateMetrics{
		MRR:          0.9,
		RecallAtK:    map[int]float64{10: 0.85},
		NDCGAtK:      map[int]float64{10: 0.85},
		LatencyStats: LatencyStats{P50: 60, P95: 120},
	}

	comp := CompareModes(modeA, modeB, "mrr")
	if comp.ValueA != 0.8 || comp.ValueB != 0.9 {
		t.Errorf("MRR comparison values wrong: %v", comp)
	}
	if comp.Diff < 0.099 || comp.Diff > 0.101 {
		t.Errorf("MRR diff wrong: expected 0.1, got %.4f", comp.Diff)
	}
	if comp.PctChange < 12.4 || comp.PctChange > 12.6 {
		t.Errorf("MRR pct change wrong: expected 12.5%%, got %.1f%%", comp.PctChange)
	}

	comp = CompareModes(modeA, modeB, "recall@10")
	if comp.ValueA != 0.7 || comp.ValueB != 0.85 {
		t.Errorf("Recall@10 comparison values wrong: %v", comp)
	}

	comp = CompareModes(modeA, modeB, "latency_p50")
	if comp.ValueA != 50 || comp.ValueB != 60 {
		t.Errorf("Latency P50 comparison values wrong: %v", comp)
	}
	// Higher latency is worse, so diff should be positive (modeB slower)
	if comp.Diff != 10 {
		t.Errorf("Latency P50 diff wrong: expected 10, got %.4f", comp.Diff)
	}
}
