package eval

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"
)

// GoldenEntry represents a single query with expected results.
type GoldenEntry struct {
	Query           string            `json:"query"`
	ExpectedIDs     []string          `json:"expected_ids"`
	RelevanceScores []float64         `json:"relevance_scores,omitempty"` // Optional graded relevance
	Metadata        map[string]string `json:"metadata,omitempty"`         // Additional metadata
	ID              string            `json:"id,omitempty"`               // Auto-generated if empty
}

// GoldenSet holds a collection of golden entries.
type GoldenSet struct {
	Entries     []GoldenEntry `json:"entries"`
	Version     string        `json:"version"`               // Schema version
	CreatedAt   string        `json:"created_at"`            // ISO8601 timestamp
	CorpusHash  string        `json:"corpus_hash,omitempty"` // Hash of indexed corpus
	Description string        `json:"description,omitempty"` // Human-readable description
}

// GoldenSetConfig holds configuration for loading golden sets.
type GoldenSetConfig struct {
	Path             string // Path to JSONL file
	MinEntries       int    // Minimum required entries
	RequireScores    bool   // Whether relevance scores are required
	AllowEmptyScores bool   // Allow empty relevance scores (binary relevance)
}

// DefaultGoldenSetConfig returns default configuration.
func DefaultGoldenSetConfig() GoldenSetConfig {
	return GoldenSetConfig{
		MinEntries:       1,
		RequireScores:    false,
		AllowEmptyScores: true,
	}
}

// LoadGoldenSet loads a golden set from a JSONL file.
func LoadGoldenSet(cfg GoldenSetConfig) (*GoldenSet, error) {
	if cfg.Path == "" {
		return nil, errors.New("golden set path is required")
	}

	file, err := os.Open(cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open golden set file: %w", err)
	}
	defer file.Close()

	gs := &GoldenSet{
		Entries:   make([]GoldenEntry, 0),
		Version:   "1.0",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		var entry GoldenEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return nil, fmt.Errorf("line %d: failed to parse JSON: %w", lineNum, err)
		}

		// Validate entry
		if err := validateEntry(entry, cfg, lineNum); err != nil {
			return nil, err
		}

		// Auto-generate ID if missing
		if entry.ID == "" {
			entry.ID = fmt.Sprintf("q%d", lineNum)
		}

		gs.Entries = append(gs.Entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	if len(gs.Entries) < cfg.MinEntries {
		return nil, fmt.Errorf("golden set has %d entries, minimum required: %d", len(gs.Entries), cfg.MinEntries)
	}

	return gs, nil
}

// validateEntry validates a single golden entry.
func validateEntry(entry GoldenEntry, cfg GoldenSetConfig, lineNum int) error {
	if strings.TrimSpace(entry.Query) == "" {
		return fmt.Errorf("line %d: query is required", lineNum)
	}

	if len(entry.ExpectedIDs) == 0 {
		return fmt.Errorf("line %d: at least one expected_id is required", lineNum)
	}

	// Check for duplicate expected IDs
	seen := make(map[string]bool)
	for _, id := range entry.ExpectedIDs {
		if seen[id] {
			return fmt.Errorf("line %d: duplicate expected_id: %s", lineNum, id)
		}
		seen[id] = true
	}

	if cfg.RequireScores && len(entry.RelevanceScores) == 0 {
		return fmt.Errorf("line %d: relevance scores required but not provided", lineNum)
	}

	if len(entry.RelevanceScores) > 0 && len(entry.RelevanceScores) != len(entry.ExpectedIDs) {
		return fmt.Errorf("line %d: relevance_scores length (%d) must match expected_ids length (%d)",
			lineNum, len(entry.RelevanceScores), len(entry.ExpectedIDs))
	}

	// Validate score ranges if provided
	for i, score := range entry.RelevanceScores {
		if score < 0 || score > 1 {
			return fmt.Errorf("line %d: relevance_scores[%d] = %.3f must be in [0, 1]", lineNum, i, score)
		}
	}

	return nil
}

// SaveGoldenSet saves a golden set to a JSONL file.
func SaveGoldenSet(gs *GoldenSet, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create golden set file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, entry := range gs.Entries {
		if err := encoder.Encode(entry); err != nil {
			return fmt.Errorf("failed to encode entry: %w", err)
		}
	}

	return nil
}

// GetEntryByID retrieves an entry by its ID.
func (gs *GoldenSet) GetEntryByID(id string) (*GoldenEntry, bool) {
	for i := range gs.Entries {
		if gs.Entries[i].ID == id {
			return &gs.Entries[i], true
		}
	}
	return nil, false
}

// FilterEntries returns entries matching the given predicate.
func (gs *GoldenSet) FilterEntries(predicate func(GoldenEntry) bool) []GoldenEntry {
	var filtered []GoldenEntry
	for _, entry := range gs.Entries {
		if predicate(entry) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// SortByQuery sorts entries alphabetically by query.
func (gs *GoldenSet) SortByQuery() {
	sort.Slice(gs.Entries, func(i, j int) bool {
		return gs.Entries[i].Query < gs.Entries[j].Query
	})
}

// Stats returns statistics about the golden set.
func (gs *GoldenSet) Stats() GoldenSetStats {
	stats := GoldenSetStats{
		TotalQueries:     len(gs.Entries),
		TotalExpectedIDs: 0,
		WithScores:       0,
	}

	for _, entry := range gs.Entries {
		stats.TotalExpectedIDs += len(entry.ExpectedIDs)
		if len(entry.RelevanceScores) > 0 {
			stats.WithScores++
		}
	}

	if stats.TotalQueries > 0 {
		stats.AvgExpectedPerQuery = float64(stats.TotalExpectedIDs) / float64(stats.TotalQueries)
	}

	return stats
}

// GoldenSetStats holds statistics about a golden set.
type GoldenSetStats struct {
	TotalQueries        int
	TotalExpectedIDs    int
	AvgExpectedPerQuery float64
	WithScores          int
}

// MergeGoldenSets merges multiple golden sets into one.
func MergeGoldenSets(sets []*GoldenSet) *GoldenSet {
	merged := &GoldenSet{
		Version:   "1.0",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	seen := make(map[string]bool)
	for _, gs := range sets {
		for _, entry := range gs.Entries {
			key := entry.Query + "|" + strings.Join(entry.ExpectedIDs, ",")
			if !seen[key] {
				seen[key] = true
				merged.Entries = append(merged.Entries, entry)
			}
		}
	}

	return merged
}

// ReadGoldenSetFromReader reads a golden set from an io.Reader.
func ReadGoldenSetFromReader(r io.Reader) (*GoldenSet, error) {
	gs := &GoldenSet{
		Entries:   make([]GoldenEntry, 0),
		Version:   "1.0",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	scanner := bufio.NewScanner(r)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		var entry GoldenEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return nil, fmt.Errorf("line %d: failed to parse JSON: %w", lineNum, err)
		}

		if entry.ID == "" {
			entry.ID = fmt.Sprintf("q%d", lineNum)
		}

		gs.Entries = append(gs.Entries, entry)
	}

	return gs, scanner.Err()
}
