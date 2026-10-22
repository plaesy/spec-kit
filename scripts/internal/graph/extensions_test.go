package graph

import (
	"reflect"
	"testing"
)

// TestCollectFiles_CFamilyHeaderAndSourceExtensions pins the bug where .hh and
// .cxx had symbol extraction but were never collected into the graph.
func TestCollectFiles_CFamilyHeaderAndSourceExtensions(t *testing.T) {
	root := writeScanFS(t, map[string]string{
		"include/foo.hh": "",
		"src/main.cxx":   "",
		"src/util.cc":    "",
		"notes.txt":      "",
	})
	got, err := collectFiles(root, "")
	if err != nil {
		t.Fatalf("collectFiles() error = %v", err)
	}
	want := []string{"include/foo.hh", "src/main.cxx", "src/util.cc"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("collectFiles() = %v, want %v", got, want)
	}
}

// TestNodeTypeOf_CFamilyIsSource checks that .hh and .cxx are labelled source,
// like the other C-family extensions.
func TestNodeTypeOf_CFamilyIsSource(t *testing.T) {
	for _, rel := range []string{"include/foo.hh", "src/main.cxx", "src/util.cc"} {
		if got := nodeTypeOf(rel); got != "source" {
			t.Errorf("nodeTypeOf(%q) = %q, want %q", rel, got, "source")
		}
	}
}

// TestSymbolsOf_CFamilyHeaderAndSourceExtract checks that .hh and .cxx files
// yield their class symbols, so their graph nodes are not empty. The C-family
// extractor reads type declarations only (cFamilySymbols), so the fixture is a
// class.
func TestSymbolsOf_CFamilyHeaderAndSourceExtract(t *testing.T) {
	content := "class Foo {\n public:\n  void bar();\n};\n"
	for _, rel := range []string{"include/foo.hh", "src/main.cxx"} {
		got := symbolsOf(rel, content)
		if !reflect.DeepEqual(got, []string{"Foo"}) {
			t.Errorf("symbolsOf(%q) = %v, want [Foo]", rel, got)
		}
	}
}
