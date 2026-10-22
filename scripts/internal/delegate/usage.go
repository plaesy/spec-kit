package delegate

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/plaesy/spec-kit/internal/config"
)

// UsageFile is the usage log, one JSON record per line, kept under the analysis
// directory. Lines are appended, so a crash never rewrites earlier records.
const UsageFile = "delegate-usage.jsonl"

// UsageRecord is one answered call. Token counts are nil for providers that do
// not report them (the CLI providers), and nil counts are not summed as zero.
type UsageRecord struct {
	At           string `json:"at"` // RFC 3339, UTC
	Provider     string `json:"provider"`
	Chain        string `json:"chain,omitempty"`
	Billable     bool   `json:"billable"`
	InputTokens  *int   `json:"input_tokens"`
	OutputTokens *int   `json:"output_tokens"`
	WallMs       int64  `json:"wall_ms"`
}

// UsagePath returns the usage log path under a project's analysis directory.
func UsagePath(analysisDir string) string {
	return filepath.Join(analysisDir, UsageFile)
}

// AppendUsage adds one record to the log, creating the file and its directory
// when needed.
func AppendUsage(path string, rec UsageRecord) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating usage directory: %w", err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("opening usage log: %w", err)
	}
	defer f.Close()
	line, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("encoding usage record: %w", err)
	}
	if _, err := f.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("writing usage log: %w", err)
	}
	return nil
}

// ReadUsage loads the log. A missing file is an empty log. A line that does not
// parse is skipped, so one damaged line cannot hide the rest of the history.
func ReadUsage(path string) ([]UsageRecord, error) {
	f, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("opening usage log: %w", err)
	}
	defer f.Close()
	var out []UsageRecord
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		var rec UsageRecord
		if err := json.Unmarshal(sc.Bytes(), &rec); err != nil || rec.Provider == "" {
			continue
		}
		out = append(out, rec)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("reading usage log: %w", err)
	}
	return out, nil
}

// Totals is the sum of usage records for one provider or one chain.
type Totals struct {
	Calls          int   `json:"calls"`
	InputTokens    int   `json:"input_tokens"`
	OutputTokens   int   `json:"output_tokens"`
	UnknownTokens  int   `json:"calls_without_token_counts"`
	WallMs         int64 `json:"wall_ms"`
}

func (t *Totals) add(rec UsageRecord) {
	t.Calls++
	t.WallMs += rec.WallMs
	if rec.InputTokens == nil || rec.OutputTokens == nil {
		t.UnknownTokens++
		return
	}
	t.InputTokens += *rec.InputTokens
	t.OutputTokens += *rec.OutputTokens
}

// Summarize totals the records per provider and per chain. A call made without
// a chain is grouped under the empty name.
func Summarize(recs []UsageRecord) (byProvider, byChain map[string]Totals) {
	byProvider = map[string]Totals{}
	byChain = map[string]Totals{}
	for _, r := range recs {
		p := byProvider[r.Provider]
		p.add(r)
		byProvider[r.Provider] = p
		c := byChain[r.Chain]
		c.add(r)
		byChain[r.Chain] = c
	}
	return byProvider, byChain
}

// SortedKeys returns the keys of a totals map in name order, for stable output.
func SortedKeys(m map[string]Totals) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// tokensOn sums input and output tokens recorded for one provider on the UTC
// date of day. Records without token counts add nothing.
func tokensOn(recs []UsageRecord, provider string, day time.Time) int {
	want := day.UTC().Format("2006-01-02")
	total := 0
	for _, r := range recs {
		if r.Provider != provider || r.InputTokens == nil || r.OutputTokens == nil {
			continue
		}
		at, err := time.Parse(time.RFC3339, r.At)
		if err != nil || at.UTC().Format("2006-01-02") != want {
			continue
		}
		total += *r.InputTokens + *r.OutputTokens
	}
	return total
}

// EstimateTokens is a rough pre-call figure for the budget check: about four
// characters per token for the prompt, plus the output allowance. It is an
// estimate, not a count, and the budget check is only as accurate as this.
func EstimateTokens(req Request) int {
	out := req.MaxOutputTokens
	if out <= 0 {
		out = defaultMaxOutputTokens
	}
	return len(req.Prompt)/4 + out
}

// CheckBudget refuses a billable call when today's tokens plus this call's
// estimate would exceed the provider's daily budget. A provider with no budget
// configured (0) is not limited here; a billable call still needs --confirm.
// The refusal is Terminal, so a chain stops instead of spending elsewhere.
func CheckBudget(entry config.DelegateProviderConfig, req Request, recs []UsageRecord, now time.Time) error {
	if !entry.Billable || entry.DailyTokenBudget <= 0 {
		return nil
	}
	used := tokensOn(recs, entry.Name, now)
	need := EstimateTokens(req)
	if used+need > entry.DailyTokenBudget {
		return &Error{Class: Terminal, Err: fmt.Errorf(
			"provider %q: daily token budget of %d would be exceeded (used %d today, this call needs about %d); no request was sent",
			entry.Name, entry.DailyTokenBudget, used, need)}
	}
	return nil
}
