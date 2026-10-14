package delegate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/plaesy/spec-kit/internal/config"
)

func intp(n int) *int { return &n }

// usageFixture has a record without the chain field (an older shape), a damaged
// line, and two current records, one with unknown token counts.
const usageFixture = `{"at":"2026-10-13T09:00:00Z","provider":"anthropic","billable":true,"input_tokens":100,"output_tokens":50,"wall_ms":900}
this line is not json
{"at":"2026-10-13T10:00:00Z","provider":"claude-cli","chain":"review","billable":false,"input_tokens":null,"output_tokens":null,"wall_ms":2000}
{"at":"2026-10-13T11:00:00Z","provider":"anthropic","chain":"review","billable":true,"input_tokens":30,"output_tokens":20,"wall_ms":700}
`

// TestReadUsage_OldRecordsStillParse checks a log written before the chain field
// existed still loads, and a damaged line is skipped instead of hiding the rest.
func TestReadUsage_OldRecordsStillParse(t *testing.T) {
	path := filepath.Join(t.TempDir(), UsageFile)
	if err := os.WriteFile(path, []byte(usageFixture), 0o644); err != nil {
		t.Fatal(err)
	}
	recs, err := ReadUsage(path)
	if err != nil {
		t.Fatalf("ReadUsage: %v", err)
	}
	if len(recs) != 3 {
		t.Fatalf("want 3 records (damaged line skipped), got %d", len(recs))
	}
	if recs[0].Chain != "" {
		t.Errorf("old record should have no chain, got %q", recs[0].Chain)
	}
}

// TestReadUsage_MissingFileIsEmpty checks a project with no history is not an error.
func TestReadUsage_MissingFileIsEmpty(t *testing.T) {
	recs, err := ReadUsage(filepath.Join(t.TempDir(), "absent.jsonl"))
	if err != nil || len(recs) != 0 {
		t.Fatalf("want empty log and no error, got %d records, err=%v", len(recs), err)
	}
}

// TestSummarize_MatchesFixture checks totals per provider and per chain match the
// records in the fixture. Unknown counts are counted separately, not as zero.
func TestSummarize_MatchesFixture(t *testing.T) {
	path := filepath.Join(t.TempDir(), UsageFile)
	if err := os.WriteFile(path, []byte(usageFixture), 0o644); err != nil {
		t.Fatal(err)
	}
	recs, err := ReadUsage(path)
	if err != nil {
		t.Fatal(err)
	}
	byProvider, byChain := Summarize(recs)

	a := byProvider["anthropic"]
	if a.Calls != 2 || a.InputTokens != 130 || a.OutputTokens != 70 || a.WallMs != 1600 {
		t.Errorf("anthropic totals = %+v", a)
	}
	c := byProvider["claude-cli"]
	if c.Calls != 1 || c.UnknownTokens != 1 || c.InputTokens != 0 {
		t.Errorf("claude-cli totals = %+v (a call without counts must not count as zero)", c)
	}
	review := byChain["review"]
	if review.Calls != 2 {
		t.Errorf("review chain calls = %d, want 2", review.Calls)
	}
	if direct := byChain[""]; direct.Calls != 1 {
		t.Errorf("direct calls = %d, want 1", direct.Calls)
	}
}

// TestCheckBudget_RefusesBeforeCall checks a call that would exceed the daily
// budget is refused, and that a budget of zero does not limit the call further.
func TestCheckBudget_RefusesBeforeCall(t *testing.T) {
	day := time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)
	recs := []UsageRecord{{At: "2026-10-13T09:00:00Z", Provider: "anthropic", Billable: true,
		InputTokens: intp(900), OutputTokens: intp(0)}}
	entry := config.DelegateProviderConfig{Name: "anthropic", Billable: true, DailyTokenBudget: 1000}

	big := Request{Prompt: strings.Repeat("x", 4000), MaxOutputTokens: 500} // about 1500 tokens
	if err := CheckBudget(entry, big, recs, day); err == nil {
		t.Fatal("want refusal: 900 used plus about 1500 exceeds 1000")
	}

	small := Request{Prompt: "hi", MaxOutputTokens: 10} // about 910 with the estimate
	if err := CheckBudget(entry, small, recs, day); err != nil {
		t.Fatalf("910 is under the 1000 budget, got %v", err)
	}

	unlimited := config.DelegateProviderConfig{Name: "anthropic", Billable: true, DailyTokenBudget: 0}
	if err := CheckBudget(unlimited, big, recs, day); err != nil {
		t.Fatalf("zero budget should not add a limit beyond --confirm, got %v", err)
	}
}

// TestCheckBudget_ResetsOnNewDay checks yesterday's tokens do not count today.
func TestCheckBudget_ResetsOnNewDay(t *testing.T) {
	recs := []UsageRecord{{At: "2026-10-12T23:00:00Z", Provider: "anthropic", Billable: true,
		InputTokens: intp(5000), OutputTokens: intp(0)}}
	entry := config.DelegateProviderConfig{Name: "anthropic", Billable: true, DailyTokenBudget: 1000}
	today := time.Date(2026, 10, 13, 1, 0, 0, 0, time.UTC)
	// A small request, so the only thing that could refuse it is yesterday's usage.
	small := Request{Prompt: "hi", MaxOutputTokens: 10}
	if err := CheckBudget(entry, small, recs, today); err != nil {
		t.Fatalf("yesterday's usage should not count today, got %v", err)
	}
}

// TestAppendUsage_RoundTrips checks an appended record reads back intact.
func TestAppendUsage_RoundTrips(t *testing.T) {
	path := filepath.Join(t.TempDir(), "analysis", UsageFile)
	rec := UsageRecord{At: "2026-10-13T12:00:00Z", Provider: "ollama", Chain: "local", WallMs: 42,
		InputTokens: intp(3), OutputTokens: intp(4)}
	if err := AppendUsage(path, rec); err != nil {
		t.Fatal(err)
	}
	recs, err := ReadUsage(path)
	if err != nil || len(recs) != 1 {
		t.Fatalf("want 1 record, got %d err=%v", len(recs), err)
	}
	if recs[0].Provider != "ollama" || recs[0].Chain != "local" || *recs[0].InputTokens != 3 {
		t.Errorf("round trip changed the record: %+v", recs[0])
	}
}
