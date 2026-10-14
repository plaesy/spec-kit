package graph

import (
	"reflect"
	"sort"
	"testing"
)

// Dart directives that name another file are edges: import, export and part.
// "part of" names the parent library and must not create an edge of its own.
func TestDartDirectivesBecomeEdges(t *testing.T) {
	files := map[string]string{
		"lib/main.dart":     "import 'src/util.dart';\nexport 'src/util.dart';\npart 'src/part.dart';\n",
		"lib/src/util.dart": "class Util {}\n",
		"lib/src/part.dart": "part of '../main.dart';\n",
	}
	root := writeRepoTree(t, files)
	fileList, known := knownSet(files)
	got := map[edgeKey]bool{}
	for _, e := range extractEdges(root, fileList, known) {
		got[edgeKey{e.Source, e.Target, e.Type}] = true
	}
	for _, k := range []edgeKey{
		{"lib/main.dart", "lib/src/util.dart", "imports"},
		{"lib/main.dart", "lib/src/part.dart", "imports"},
	} {
		if !got[k] {
			t.Errorf("missing edge %+v; got %v", k, keysOf(got))
		}
	}
	for k := range got {
		if k.src == "lib/src/part.dart" {
			t.Errorf("part of must not create an edge from the part file; got %+v", k)
		}
	}
}

// symbolsOf must name every Dart type declaration form, not just classes.
func TestDartDeclarationSymbols(t *testing.T) {
	src := "" +
		"class Plain {}\n" +
		"abstract class Shape {}\n" +
		"final class Locked {}\n" +
		"sealed class Result {}\n" +
		"mixin Logs {}\n" +
		"mixin class Both {}\n" +
		"enum Mode { a, b }\n" +
		"typedef Handler = void Function(int);\n" +
		"extension Money on int {}\n" +
		"extension type Meters(int value) {\n" +
		"}\n"
	got := symbolsOf("lib/types.dart", src)
	want := []string{"Plain", "Shape", "Locked", "Result", "Logs", "Both", "Mode", "Handler", "Money", "Meters"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("symbolsOf() = %v\nwant %v", got, want)
	}
}

// A top-level function is named once, from its declaration line, and a method
// call inside a body is not a declaration.
func TestDartFunctionSymbols(t *testing.T) {
	src := "" +
		"void main() {\n" +
		"  runApp();\n" +
		"}\n" +
		"\n" +
		"class App {\n" +
		"  Widget build(BuildContext context) {\n" +
		"    return Text('x');\n" +
		"  }\n" +
		"}\n"
	got := symbolsOf("lib/main.dart", src)
	want := []string{"main", "App", "build"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("symbolsOf() = %v\nwant %v", got, want)
	}
}

// The graph must leave out the folders and generated files a Flutter or
// Python project keeps beside its source, whichever walk reaches them.
func TestCollectFilesExcludesToolingAndGitignored(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		".gitignore":                       "*.g.dart\n/out/\n",
		"lib/main.dart":                    "void main() {}\n",
		"lib/src/model.g.dart":             "class Model {}\n",
		"lib/src/model.freezed.dart":       "class Frozen {}\n",
		"out/gen.dart":                     "class Gen {}\n",
		"build/app/gen.dart":               "class B {}\n",
		".dart_tool/cache.dart":            "class C {}\n",
		"target/x.dart":                    "class W {}\n",
		"venv/lib/site.py":                 "x = 1\n",
		"ios/Pods/Pod.swift":               "class P {}\n",
		"ios/Flutter/ephemeral/host.swift": "class H {}\n",
		"linux/flutter/ephemeral/host.cc":  "int h;\n",
		"node_modules/pkg/index.js":        "module.exports = {}\n",
		"ios/Runner/AppDelegate.swift":     "class AppDelegate {}\n",
		"coverage/lcov.dart":               "class L {}\n",
	})
	got, err := collectFiles(root, "")
	if err != nil {
		t.Fatalf("collectFiles: %v", err)
	}
	sort.Strings(got)
	want := []string{
		"ios/Runner/AppDelegate.swift",
		"lib/main.dart",
		"lib/src/model.freezed.dart",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("collectFiles() = %v\nwant %v", got, want)
	}
}
