package retriever

import (
	"context"
	"strings"
	"testing"

	"github.com/plaesy/spec-kit/internal/config"
	"github.com/plaesy/spec-kit/internal/search/textindex"
	"github.com/plaesy/spec-kit/internal/search/vectorstore"
)

func TestParseFilter_RejectsBadSpecs(t *testing.T) {
	cases := map[string]string{
		"no colon":      "markdown",
		"unknown key":   "owner:alice",
		"empty value":   "source:",
		"bad date":      "date:>2024-13-40",
		"date no value": "date:>",
	}
	for name, spec := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseFilter(spec); err == nil {
				t.Errorf("ParseFilter(%q) = nil error, want an error", spec)
			}
		})
	}
}

func TestParseFilter_EmptySpecKeepsEverything(t *testing.T) {
	f, err := ParseFilter("  ")
	if err != nil {
		t.Fatalf("ParseFilter: %v", err)
	}
	if !f.Match(map[string]string{}) {
		t.Error("empty filter rejected a chunk with no metadata")
	}
}

func TestFilter_Source(t *testing.T) {
	f, _ := ParseFilter("source:markdown")
	if !f.Match(map[string]string{"source": "markdown"}) {
		t.Error("source:markdown rejected a markdown chunk")
	}
	if f.Match(map[string]string{"source": "html"}) {
		t.Error("source:markdown accepted an html chunk")
	}
}

// The stored path is absolute, so a path filter has to match a segment anywhere
// in it. It must also match whole segments only.
func TestFilter_PathPrefixMatchesWholeSegments(t *testing.T) {
	f, _ := ParseFilter("path:internal/search")
	cases := map[string]bool{
		"/repo/internal/search/a.go":     true,
		"internal/search/b.md":           true,
		"/repo/internal/searching/c.go":  false,
		"/repo/other/internal/search.go": false,
	}
	for path, want := range cases {
		if got := f.Match(map[string]string{"path": path}); got != want {
			t.Errorf("path:internal/search on %q = %v, want %v", path, got, want)
		}
	}
}

func TestFilter_FileType(t *testing.T) {
	for _, spec := range []string{"file_type:md", "file_type:.md"} {
		f, _ := ParseFilter(spec)
		if !f.Match(map[string]string{"path": "/repo/docs/guide.MD"}) {
			t.Errorf("%s rejected guide.MD", spec)
		}
		if f.Match(map[string]string{"path": "/repo/docs/guide.html"}) {
			t.Errorf("%s accepted guide.html", spec)
		}
	}
}

func TestFilter_DateOperators(t *testing.T) {
	meta := func(d string) map[string]string { return map[string]string{ModifiedKey: d} }
	cases := []struct {
		spec, date string
		want       bool
	}{
		{"date:>2024-01-01", "2024-01-02", true},
		{"date:>2024-01-01", "2024-01-01", false},
		{"date:>=2024-01-01", "2024-01-01", true},
		{"date:<2024-01-01", "2023-12-31", true},
		{"date:<2024-01-01", "2024-01-01", false},
		{"date:<=2024-01-01", "2024-01-01", true},
		{"date:2024-01-01", "2024-01-01", true},
		{"date:2024-01-01", "2024-01-02", false},
		// A date that is missing or unreadable fails closed.
		{"date:>2000-01-01", "", false},
	}
	for _, c := range cases {
		f, err := ParseFilter(c.spec)
		if err != nil {
			t.Fatalf("ParseFilter(%q): %v", c.spec, err)
		}
		if got := f.Match(meta(c.date)); got != c.want {
			t.Errorf("%s on %q = %v, want %v", c.spec, c.date, got, c.want)
		}
	}
	f, _ := ParseFilter("date:>2000-01-01")
	if f.Match(map[string]string{}) {
		t.Error("a chunk with no modified date matched a date range")
	}
}

func TestFilter_TermsAreAnded(t *testing.T) {
	f, _ := ParseFilter("source:markdown,path:docs,date:>2024-01-01")
	ok := map[string]string{"source": "markdown", "path": "/r/docs/a.md", ModifiedKey: "2024-06-01"}
	if !f.Match(ok) {
		t.Fatal("chunk meeting every term was rejected")
	}
	bad := map[string]string{"source": "markdown", "path": "/r/docs/a.md", ModifiedKey: "2023-06-01"}
	if f.Match(bad) {
		t.Error("chunk failing the date term was accepted")
	}
}

// Predicate filters run before fusion. A chunk that fails them must not reach
// the result, even when it is the best match on both sides.
func TestRetrieveWith_PredicateDropsCandidatesBeforeFusion(t *testing.T) {
	cfg := config.RetrieverConfig{
		TopK:   5,
		Hybrid: config.HybridConfig{RRFK: 60, VectorWeight: 1, BM25Weight: 1},
	}
	vs := &mockVectorStoreForRetriever{data: map[string]vectorstore.SearchResult{
		"keep": {ID: "keep", Score: 0.5, Content: "keep", Metadata: map[string]string{"doc_id": "d1", "path": "/r/docs/a.md"}},
		"drop": {ID: "drop", Score: 0.9, Content: "drop", Metadata: map[string]string{"doc_id": "d2", "path": "/r/old/b.md"}},
	}}
	ti := &mockTextIndexForRetriever{data: map[string]textindex.SearchResult{
		"drop": {ID: "drop", Score: 9, Content: "drop", Metadata: map[string]string{"doc_id": "d2", "path": "/r/old/b.md"}},
		"keep": {ID: "keep", Score: 3, Content: "keep", Metadata: map[string]string{"doc_id": "d1", "path": "/r/docs/a.md"}},
	}}
	h := NewHybridRetriever(cfg, &mockEmbedderForRetriever{dimension: 3}, vs, ti, &mockDocStoreForRetriever{})

	where, err := ParseFilter("path:docs")
	if err != nil {
		t.Fatal(err)
	}
	got, err := h.RetrieveWith(context.Background(), "q", QueryOptions{Where: where})
	if err != nil {
		t.Fatalf("RetrieveWith: %v", err)
	}
	for _, r := range got {
		if strings.Contains(r.ID, "drop") {
			t.Errorf("filtered chunk %q reached the result", r.ID)
		}
	}
	if len(got) != 1 || got[0].ID != "keep" {
		t.Errorf("result = %+v, want only the kept chunk", got)
	}
}
