package main

import "testing"

// TestKeepKindFiltersByLabel covers the bug where --kind symbol still printed doc
// rows: a Markdown heading is labelled doc even though the symbol index produced it.
func TestKeepKindFiltersByLabel(t *testing.T) {
	rows := []row{
		{kind: "symbol", name: "validate"},
		{kind: "doc", name: "heading from symbol index"},
		{kind: "symbol", name: "parse"},
	}

	got := keepKind(rows, "symbol")
	if len(got) != 2 || got[0].name != "validate" || got[1].name != "parse" {
		t.Errorf("keepKind(symbol) = %+v, want the two symbol rows in order", got)
	}

	if got := keepKind(rows, "doc"); len(got) != 1 || got[0].kind != "doc" {
		t.Errorf("keepKind(doc) = %+v, want the one doc row", got)
	}
}

// TestParseExtsNormalizes checks that --ext accepts dots, mixed case, and blank terms.
func TestParseExtsNormalizes(t *testing.T) {
	got := parseExts(" .Go, ts ,,.md,")
	want := []string{"go", "ts", "md"}
	if len(got) != len(want) {
		t.Fatalf("parseExts = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("parseExts = %v, want %v", got, want)
		}
	}
}

// TestKeepExtORsAcrossKinds checks that --ext keeps a row when its extension is any
// of the listed ones, for both symbol and doc rows, and that an empty list keeps all.
func TestKeepExtORsAcrossKinds(t *testing.T) {
	rows := []row{
		{kind: "symbol", ext: "go", name: "a"},
		{kind: "doc", ext: "md", name: "b"},
		{kind: "symbol", ext: "ts", name: "c"},
		{kind: "doc", ext: "html", name: "d"},
	}

	got := keepExt(rows, []string{"md", "go"})
	if len(got) != 2 || got[0].name != "a" || got[1].name != "b" {
		t.Errorf("keepExt(md,go) = %+v, want rows a and b in order", got)
	}

	if got := keepExt(rows, nil); len(got) != len(rows) {
		t.Errorf("keepExt(nil) = %d rows, want all %d", len(got), len(rows))
	}
}
