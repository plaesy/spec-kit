package graph

import (
	"reflect"
	"testing"
)

// polyglotFixture is the tree the extraction tests below share. The three
// manifest files (go.mod, pubspec.yaml) are deliberately listed in `known`
// even though collectFiles would not return them: that is the point, see
// TestExtractEdgesFindsGoAndDartPackageImportsInARealBuild.
func polyglotFixture() map[string]string {
	return map[string]string{
		"go.mod":                       "module example.com/demo\n",
		"main.go":                      "package main\n\nimport (\n\t\"example.com/demo/internal/util\"\n)\n\nfunc main() {}\n",
		"single.go":                    "package main\n\nimport \"example.com/demo/internal/util\"\n",
		"internal/util/helper.go":      "package util\n",
		"internal/util/helper_test.go": "package util\n",
		"pubspec.yaml":                 "name: app\n",
		"lib/model.dart":               "class Model {}\n",
		"app/main.dart":                "import 'package:app/model.dart';\nimport '../lib/other.dart';\nimport 'dart:io';\n",
		// A single source-root folder: indexJavaFiles drops the first path
		// segment, so this is the shape its dotted names actually line up with.
		"src/com/example/App.java":         "package com.example;\nimport com.example.util.Helper;\nimport java.util.List;\nimport static com.example.util.Constants.MAX;\nimport com.example.util.*;\n",
		"src/com/example/util/Helper.java": "package com.example.util;\n",
	}
}

func knownSet(files map[string]string) (fileList []string, known map[string]bool) {
	known = map[string]bool{}
	for rel := range files {
		fileList = append(fileList, rel)
		known[rel] = true
	}
	return fileList, known
}

// resolveDartImport resolves a Dart spec to a node id or "". It is the only
// resolver with a package-registry form ("package:name/path"), which is how
// real Dart code imports anything outside its own directory, so both forms
// plus the escapes need pinning.
func TestResolveDartImport(t *testing.T) {
	root := "/repo"
	known := map[string]bool{
		"lib/model.dart":                true,
		"lib/src/deep.dart":             true,
		"packages/mylib/lib/thing.dart": true,
	}
	pkgs := map[string]string{
		"app":   "./lib",
		"mylib": "packages/mylib/lib",
	}
	tests := []struct{ name, rel, spec, want string }{
		{name: "package spec resolves through the pubspec's lib dir", rel: "app/main.dart", spec: "package:app/model.dart", want: "lib/model.dart"},
		{name: "package spec with a nested path", rel: "app/main.dart", spec: "package:app/src/deep.dart", want: "lib/src/deep.dart"},
		{name: "package spec for a second workspace package", rel: "app/main.dart", spec: "package:mylib/thing.dart", want: "packages/mylib/lib/thing.dart"},
		{name: "package spec with no path resolves to the lib dir itself", rel: "app/main.dart", spec: "package:app", want: ""},
		{name: "unknown package", rel: "app/main.dart", spec: "package:other/model.dart", want: ""},
		{name: "package spec whose file is not a node", rel: "app/main.dart", spec: "package:app/missing.dart", want: ""},
		{name: "relative spec against the importing file's dir", rel: "app/main.dart", spec: "model.dart", want: ""},
		{name: "relative spec climbing out of the repo", rel: "app/main.dart", spec: "../../../etc/passwd", want: ""},
		{name: "relative spec reaching a real file", rel: "lib/src/other.dart", spec: "deep.dart", want: "lib/src/deep.dart"},
		{name: "relative spec from the repo root", rel: "main.dart", spec: "lib/model.dart", want: "lib/model.dart"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveDartImport(root, tc.rel, tc.spec, pkgs, known); got != tc.want {
				t.Errorf("resolveDartImport(%q, %q) = %q, want %q", tc.rel, tc.spec, got, tc.want)
			}
		})
	}
}

// The Go, Dart and Java passes of extractEdges are each guarded by a pre-scan
// (go.mod / pubspec.yaml / the .java index) and each has a skip rule for a
// spec that is not a real repository reference. None of the three is reached
// by the existing md/ps1/sh/js/py fixture, so every branch below is
// unexercised without this test.
func TestExtractEdgesGoDartJava(t *testing.T) {
	files := polyglotFixture()
	root := writeRepoTree(t, files)
	fileList, known := knownSet(files)
	// The manifests are not INCLUDE_EXT_RE matches; they are in `known` only
	// because the pre-scans read them by name.
	known["go.mod"] = true
	known["pubspec.yaml"] = true

	got := map[edgeKey]bool{}
	for _, e := range extractEdges(root, fileList, known) {
		if e.Confidence != "EXTRACTED" {
			t.Errorf("extractEdges produced a %q edge, want EXTRACTED: %+v", e.Confidence, e)
		}
		got[edgeKey{e.Source, e.Target, e.Type}] = true
	}

	for _, k := range []edgeKey{
		{"main.go", "internal/util/helper.go", "imports"},
		{"app/main.dart", "lib/model.dart", "imports"},
		{"src/com/example/App.java", "src/com/example/util/Helper.java", "imports"},
	} {
		if !got[k] {
			t.Errorf("missing edge %+v; got %v", k, keysOf(got))
		}
	}

	// The skip rules, each of which is a spec that names something other than
	// a file in this repository: a Java wildcard import has no single file, a
	// Java stdlib import is not in the index, and a Dart SDK import is not a
	// repository path.
	for _, k := range []edgeKey{
		{"src/com/example/App.java", "com.example.util.Constants.MAX", "imports"},
		{"src/com/example/App.java", "java.util.List", "imports"},
		{"app/main.dart", "dart:io", "imports"},
		{"app/main.dart", "lib/other.dart", "imports"},
	} {
		if got[k] {
			t.Errorf("edge for a non-repository spec: %+v", k)
		}
	}
}

// FIXED BEHAVIOUR: indexJavaFiles now finds the longest standard source-root
// prefix (src/main/java/, src/test/java/, src/, java/) and keys by the
// package+class name. A Maven/Gradle file at
// "maven/src/main/java/com/example/util/Helper.java" is now indexed as
// "com.example.util.Helper", which `import com.example.util.Helper;` names.
func TestIndexJavaFilesMavenLayout(t *testing.T) {
	files := []string{
		"maven/src/main/java/com/example/Use.java",
		"maven/src/main/java/com/example/util/Helper.java",
	}
	got := indexJavaFiles(files)
	if got["com.example.util.Helper"] != "maven/src/main/java/com/example/util/Helper.java" {
		t.Fatalf("indexJavaFiles = %v, want key com.example.util.Helper", got)
	}
	if got["com.example.Use"] != "maven/src/main/java/com/example/Use.java" {
		t.Fatalf("indexJavaFiles = %v, want key com.example.Use", got)
	}

	// The same fix seen from the extractor: the edge the import names is
	// now produced.
	tree := writeRepoTree(t, map[string]string{
		"maven/src/main/java/com/example/Use.java":         "package com.example;\nimport com.example.util.Helper;\n",
		"maven/src/main/java/com/example/util/Helper.java": "package com.example.util;\n",
	})
	edges := extractEdges(tree, files, map[string]bool{
		"maven/src/main/java/com/example/Use.java":         true,
		"maven/src/main/java/com/example/util/Helper.java": true,
	})
	found := false
	for _, e := range edges {
		if e.Type == "imports" && e.Target == "maven/src/main/java/com/example/util/Helper.java" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("FIXED BEHAVIOUR: expected an imports edge to Helper.java, got %v", edges)
	}
}

// CURRENT BEHAVIOUR — suspected bug: the Go pass only matches a quoted path on
// a line of its own (goImportLineRE is `^\s*"..."\s*$` applied per line), so a
// single-line `import "path"` — no parentheses — is not seen. Only a grouped
// `import ( ... )` block is read. What it should be: also accept the
// `import "path"` form, which gofmt keeps for a file with one import and which
// is the shape a small generated file has.
func TestExtractEdgesSingleLineGoImport(t *testing.T) {
	files := map[string]string{
		"go.mod":                  "module example.com/demo\n",
		"single.go":               "package main\n\nimport \"example.com/demo/internal/util\"\n",
		"internal/util/helper.go": "package util\n",
	}
	root := writeRepoTree(t, files)
	fileList, known := knownSet(files)
	known["go.mod"] = true
	got := map[edgeKey]bool{}
	for _, e := range extractEdges(root, fileList, known) {
		got[edgeKey{e.Source, e.Target, e.Type}] = true
	}
	// The idiomatic spelling for a file with one dependency. The old pattern
	// required a bare quoted line, so it only matched inside `import ( ... )`
	// and a file written this way contributed nothing to the graph.
	if !got[edgeKey{"single.go", "internal/util/helper.go", "imports"}] {
		t.Errorf("the single-line `import \"...\"` form must resolve like a grouped one; got %v", keysOf(got))
	}
}

// A JS import of a bare specifier (a package name, not a path) is skipped by
// the "." prefix guard rather than resolved, so it must not reach the
// resolver and be reported as an unresolved relative path.
func TestExtractEdgesSkipsBareJSSpecifiers(t *testing.T) {
	files := map[string]string{
		"app/main.js": "import React from 'react';\nimport {x} from './util';\n",
		"app/util.js": "export const x = 1;\n",
	}
	root := writeRepoTree(t, files)
	fileList, known := knownSet(files)
	got := map[edgeKey]bool{}
	for _, e := range extractEdges(root, fileList, known) {
		got[edgeKey{e.Source, e.Target, e.Type}] = true
	}
	if !got[edgeKey{"app/main.js", "app/util.js", "imports"}] {
		t.Errorf("relative import missing; got %v", keysOf(got))
	}
	if got[edgeKey{"app/main.js", "react", "imports"}] {
		t.Error("a bare package specifier must not become a node")
	}
}

// A file the walk could not read (deleted between collectFiles and the read)
// is skipped, not fatal: extractEdges is the tolerant half of a best-effort
// scan and one unreadable file must not cost the whole graph.
func TestExtractEdgesToleratesUnreadableFiles(t *testing.T) {
	files := map[string]string{"a.md": "See [b](b.md).\n", "b.md": "# B\n"}
	root := writeRepoTree(t, files)
	fileList, known := knownSet(files)
	fileList = append(fileList, "gone.md") // listed but not on disk
	known["gone.md"] = true
	got := extractEdges(root, fileList, known)
	if len(got) != 1 || got[0].Target != "b.md" {
		t.Fatalf("extractEdges() = %v, want the single a.md -> b.md edge", got)
	}
}

// CURRENT BEHAVIOUR — suspected bug: scanGoModules and scanDartPubspecs are
// driven by `known`, which Build fills from collectFiles — and collectFiles
// only returns paths matching INCLUDE_EXT_RE, which lists neither ".mod" nor
// ".yaml". So in a real build both pre-scans see an empty `known` for their
// manifests and return an empty map, the `if len(gomodPath) > 0` guard in the
// .go pass is never true, and resolveDartImport's "package:" form always
// returns "". Result: a Go or Dart project gets NO import edges at all, while
// the unit tests for the two scanners pass because they hand `known` a
// go.mod/pubspec.yaml entry that collectFiles cannot produce. What it should
// be: the scanners should look for go.mod/pubspec.yaml on disk (or `known`
// should be built from the walk before the extension filter).
func TestExtractEdgesFindsGoAndDartPackageImportsInARealBuild(t *testing.T) {
	files := polyglotFixture()
	root := newBuildRoot(t)
	writeFiles(t, root, files)
	opts := Options{RepoPath: root}
	p, err := ResolvePaths(opts)
	if err != nil {
		t.Fatalf("ResolvePaths: %v", err)
	}
	g, err := Build(opts, p)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	got := map[edgeKey]bool{}
	for _, e := range g.Edges {
		got[edgeKey{e.Source, e.Target, e.Type}] = true
	}
	for _, k := range []edgeKey{
		// Java works: indexJavaFiles is built from the collected files, and
		// the fixture uses a single source-root folder.
		{"src/com/example/App.java", "src/com/example/util/Helper.java", "imports"},
	} {
		if !got[k] {
			t.Errorf("missing edge %+v; got %v", k, keysOf(got))
		}
	}
	for _, k := range []edgeKey{
		// Go: the go.mod is found on disk, so the module path is known and a
		// grouped import resolves. This edge did not exist before: the
		// manifest was looked for among the symbol-extraction candidates,
		// where a .mod file can never appear, so every Go project built a
		// dependency-free graph and still reported success.
		{"main.go", "internal/util/helper.go", "imports"},
		// Same, via the single-line import spelling.
		{"single.go", "internal/util/helper.go", "imports"},
		// Dart: the pubspec's `name:` is the package the `package:` import
		// form resolves against. Previously unreachable for the same reason.
		{"app/main.dart", "lib/model.dart", "imports"},
	} {
		if !got[k] {
			t.Errorf("missing edge %+v; got %v", k, keysOf(got))
		}
	}
	// Every pass must now contribute, not just the one that already worked.
	if len(g.Edges) != 4 {
		t.Fatalf("Build() produced %d edges (%v), want 4 (Java + 2 Go + Dart)", len(g.Edges), edgeSet(g.Edges))
	}
}

func TestScanGoModulesIgnoresManifestsWithoutAModuleLine(t *testing.T) {
	root := writeRepoTree(t, map[string]string{
		"go.mod":         "go 1.20\n", // no `module` line
		"sub/go.mod":     "module example.com/sub\n",
		"scripts/go.sum": "",
	})
	got := scanGoModules(root, map[string]bool{"go.mod": true, "sub/go.mod": true, "scripts/go.sum": true})
	want := map[string]string{"sub": "example.com/sub"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("scanGoModules() = %v, want %v", got, want)
	}
}

func TestScanDartPubspecsIgnoresManifestsWithoutAName(t *testing.T) {
	root := writeRepoTree(t, map[string]string{
		"pubspec.yaml":              "dependencies:\n  flutter: ^3.0\n", // no `name:` line
		"packages/lib/pubspec.yaml": "name: mylib\n",
	})
	got := scanDartPubspecs(root, map[string]bool{
		"pubspec.yaml": true, "packages/lib/pubspec.yaml": true,
	})
	want := map[string]string{"mylib": "packages/lib/lib"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("scanDartPubspecs() = %v, want %v", got, want)
	}
}

// resolveGoImport returns nil when no go.mod governs the importing file. The
// existing test only has a root module, which governs everything, so the
// "no module at all" case — the one a nested project hits — was unexercised.
func TestResolveGoImportWithoutAnyModule(t *testing.T) {
	got := resolveGoImport("root/x.go", "example.com/demo/pkg", map[string]string{"sub": "example.com/demo"}, map[string][]string{
		"sub": {"sub/y.go"},
	})
	if got != nil {
		t.Fatalf("resolveGoImport() = %v, want nil when no go.mod governs the file", got)
	}
	// A go.mod at the repo root governs everything again, and the module's own
	// package fans out to every non-test file in the root directory except the
	// importing file itself. The _test.go filter lives in indexGoDirFiles, not
	// here, so this builds the map the way extractEdges does.
	got = resolveGoImport("main.go", "example.com/demo", map[string]string{"": "example.com/demo"},
		indexGoDirFiles([]string{"main.go", "main_test.go", "other.go", "docs/x.md"}))
	if !reflect.DeepEqual(got, []string{"other.go"}) {
		t.Fatalf("resolveGoImport() = %v, want the other non-test file", got)
	}
}
