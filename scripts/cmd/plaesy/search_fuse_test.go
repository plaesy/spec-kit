package main

import "testing"

// fuse must interleave the two lists by rank, so the best symbol and the best
// document both come before the second of either. A plain concatenation would
// bury the documents behind every symbol.
func TestFuse_InterleavesByRank(t *testing.T) {
	symbols := []row{{kind: "symbol", name: "s1"}, {kind: "symbol", name: "s2"}}
	docs := []row{{kind: "doc", name: "d1"}, {kind: "doc", name: "d2"}}

	got := fuse(symbols, docs)
	want := []string{"s1", "d1", "s2", "d2"}
	if len(got) != len(want) {
		t.Fatalf("fuse returned %d rows, want %d", len(got), len(want))
	}
	for i, name := range want {
		if got[i].name != name {
			t.Errorf("row %d = %q, want %q", i, got[i].name, name)
		}
	}
	for i := 1; i < len(got); i++ {
		if got[i].score > got[i-1].score {
			t.Errorf("scores not descending at %d: %v then %v", i, got[i-1].score, got[i].score)
		}
	}
}

// --kind must reject anything other than symbol or doc before touching an index.
func TestSearchQuery_RejectsUnknownKind(t *testing.T) {
	err := searchQuery(nil, "text", ".", ".plaesy/analysis", 10, "code", "", "", false)
	if err == nil {
		t.Fatal("searchQuery accepted --kind code")
	}
}
