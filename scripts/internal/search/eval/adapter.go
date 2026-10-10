package eval

import (
	"time"
)

// MockSearchEngine is a mock implementation for testing.
type MockSearchEngine struct {
	name         string
	mode         SearchMode
	fixedResults map[string][]string
	latency      time.Duration
}

// NewMockSearchEngine creates a new mock search engine.
func NewMockSearchEngine(name string, mode SearchMode) *MockSearchEngine {
	return &MockSearchEngine{
		name:         name,
		mode:         mode,
		fixedResults: make(map[string][]string),
		latency:      10 * time.Millisecond,
	}
}

// Name returns the engine name.
func (m *MockSearchEngine) Name() string {
	return m.name
}

// SetFixedResults sets fixed results for specific queries.
func (m *MockSearchEngine) SetFixedResults(query string, ids []string) {
	m.fixedResults[query] = ids
}

// SetLatency sets the mock latency.
func (m *MockSearchEngine) SetLatency(latency time.Duration) {
	m.latency = latency
}

// Search returns mock results.
func (m *MockSearchEngine) Search(query string, topK int, mode SearchMode) ([]string, []float64, error) {
	// Simulate latency
	time.Sleep(m.latency)

	if ids, ok := m.fixedResults[query]; ok {
		scores := make([]float64, len(ids))
		for i := range scores {
			scores[i] = 1.0 - float64(i)*0.1
		}
		if len(ids) > topK {
			ids = ids[:topK]
			scores = scores[:topK]
		}
		return ids, scores, nil
	}

	// Default: return empty results
	return []string{}, []float64{}, nil
}
