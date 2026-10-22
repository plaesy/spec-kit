package trimmer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// EstimateTokens mirrors estimate_tokens(): a rough chars/4 heuristic, no
// tokenizer dependency.
func EstimateTokens(text string) int {
	return (len(text) + 3) / 4
}

// PctSaved mirrors pct_saved(): percentage reduction from before to after,
// rounded to one decimal place, matching `awk '%.1f'`.
func PctSaved(before, after int) float64 {
	if before <= 0 {
		return 0
	}
	pct := (1 - float64(after)/float64(before)) * 100
	// Round to 1 decimal, same precision as the bash awk printf.
	return roundTo(pct, 1)
}

func roundTo(v float64, decimals int) float64 {
	mult := 1.0
	for i := 0; i < decimals; i++ {
		mult *= 10
	}
	return float64(int(v*mult+sign(v)*0.5)) / mult
}

func sign(v float64) float64 {
	if v < 0 {
		return -1
	}
	return 1
}

// StatRecord mirrors one entry of token-stats.json, written by add_stat_record().
type StatRecord struct {
	Timestamp string  `json:"timestamp"`
	Layer     string  `json:"layer"`
	Path      string  `json:"path"`
	Before    int     `json:"before"`
	After     int     `json:"after"`
	SavedPct  float64 `json:"saved_pct"`
}

// AddStatRecord mirrors add_stat_record(): appends one record to the
// cumulative token-stats.json file at statsPath, creating it if needed.
func AddStatRecord(statsPath, layer, target string, before, after int) error {
	if err := os.MkdirAll(filepath.Dir(statsPath), 0o755); err != nil {
		return err
	}
	var records []StatRecord
	if data, err := os.ReadFile(statsPath); err == nil {
		_ = json.Unmarshal(data, &records) // corrupt/empty file -> start fresh, same as bash's best-effort append
	}
	records = append(records, StatRecord{
		Timestamp: time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		Layer:     layer,
		Path:      target,
		Before:    before,
		After:     after,
		SavedPct:  PctSaved(before, after),
	})
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(statsPath, append(data, '\n'), 0o644)
}

// Report mirrors cmd_report(): a per-layer and grand-total summary of
// cumulative token savings recorded at statsPath.
func Report(statsPath string) (string, error) {
	data, err := os.ReadFile(statsPath)
	if err != nil {
		return "No token stats yet. Run 'compress' or 'run' first.\n", nil
	}
	var records []StatRecord
	if err := json.Unmarshal(data, &records); err != nil || len(records) == 0 {
		return "No token stats yet. Run 'compress' or 'run' first.\n", nil
	}

	type totals struct {
		count  int
		before int
		after  int
	}
	byLayer := map[string]*totals{}
	var layers []string
	grandBefore, grandAfter := 0, 0
	for _, r := range records {
		t, ok := byLayer[r.Layer]
		if !ok {
			t = &totals{}
			byLayer[r.Layer] = t
			layers = append(layers, r.Layer)
		}
		t.count++
		t.before += r.Before
		t.after += r.After
		grandBefore += r.Before
		grandAfter += r.After
	}
	sort.Strings(layers)

	out := "plaesy-trim report\n"
	out += "====================\n"
	for _, layer := range layers {
		t := byLayer[layer]
		out += fmt.Sprintf("%s: %d runs, %d -> %d tokens (%.1f%% avg saved)\n", layer, t.count, t.before, t.after, PctSaved(t.before, t.after))
	}
	out += "--------------------\n"
	out += fmt.Sprintf("total: %d -> %d tokens (%.1f%% saved)\n", grandBefore, grandAfter, PctSaved(grandBefore, grandAfter))
	return out, nil
}
