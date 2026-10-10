package graph

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// edgeKey identifies one extracted edge for comparison in a table.
type edgeKey struct{ src, tgt, typ string }

func TestCleanPS1Target(t *testing.T) {
	tests := []struct{ name, in, want string }{
		{"backslash prefix", `$PSScriptRoot\lib\a.ps1`, "lib/a.ps1"},
		{"slash prefix", `$PSScriptRoot/lib/a.ps1`, "lib/a.ps1"},
		{"windows separators", `scripts\bash\a.ps1`, "scripts/bash/a.ps1"},
		{"already clean", "lib/a.ps1", "lib/a.ps1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanPS1Target(tt.in); got != tt.want {
				t.Errorf("cleanPS1Target(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestResolveJSImport(t *testing.T) {
	root := "/repo"
	known := map[string]bool{
		"src/app.ts":        true,
		"src/lib/util.ts":   true,
		"src/lib/index.ts":  true,
		"src/node_mod/x.js": true,
	}
	tests := []struct {
		name, rel, spec, want string
	}{
		{"exact file", "src/main.ts", "./app", "src/app.ts"},
		{"extension added", "src/main.ts", "./lib/util", "src/lib/util.ts"},
		{"directory index", "src/main.ts", "./lib", "src/lib/index.ts"},
		{"absolute-ish path", "src/main.ts", "./node_mod/x.js", "src/node_mod/x.js"},
		{"bare specifier is not resolved", "src/main.ts", "react", ""},
		{"outside the repo", "src/main.ts", "../../etc/passwd", ""},
		{"unknown target", "src/main.ts", "./missing", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveJSImport(root, tt.rel, tt.spec, known)
			if got != tt.want {
				t.Errorf("resolveJSImport(%q, %q) = %q, want %q", tt.rel, tt.spec, got, tt.want)
			}
		})
	}
}

func TestResolvePyImport(t *testing.T) {
	known := map[string]bool{"pkg/mod.py": true, "pkg/pkg/inner.py": true}
	tests := []struct{ name, rel, spec, want string }{
		{"absolute spec resolves from the repo root", "pkg/main.py", "pkg.mod", "pkg/mod.py"},
		{"a spec next to the importer wins", "pkg/main.py", "pkg.inner", "pkg/pkg/inner.py"},
		{"relative dots are ignored", "pkg/main.py", ".mod", ""},
		{"empty", "pkg/main.py", "", ""},
		{"outside the repo", "pkg/main.py", "os.path", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolvePyImport("/repo", tt.rel, tt.spec, known); got != tt.want {
				t.Errorf("resolvePyImport(%q) = %q, want %q", tt.spec, got, tt.want)
			}
		})
	}
}

func TestIndexGoDirFiles(t *testing.T) {
	files := []string{
		"main.go", "main_test.go",
		"internal/graph/graph.go", "internal/graph/graph_test.go",
		"docs/readme.md",
	}
	got := indexGoDirFiles(files)
	if len(got[""]) != 1 || got[""][0] != "main.go" {
		t.Errorf("root dir = %v, want [main.go] (test files are excluded)", got[""])
	}
	if len(got["internal/graph"]) != 1 || got["internal/graph"][0] != "internal/graph/graph.go" {
		t.Errorf("internal/graph = %v, want [internal/graph/graph.go]", got["internal/graph"])
	}
	if _, ok := got["docs"]; ok {
		t.Errorf("a non-Go file must not create an entry, got %v", got["docs"])
	}
}

func TestResolveGoImport(t *testing.T) {
	gomod := map[string]string{
		"":        "github.com/plaesy/spec-kit",
		"scripts": "github.com/plaesy/spec-kit/scripts",
	}
	dirFiles := map[string][]string{
		"internal/graph":     {"internal/graph/a.go", "internal/graph/b.go"},
		"scripts/cmd/plaesy": {"scripts/cmd/plaesy/main.go"},
	}
	tests := []struct {
		name, rel, spec string
		want            []string
	}{
		{
			name: "package fans out to every non-test file in the dir",
			rel:  "cmd/plaesy/x.go", spec: "github.com/plaesy/spec-kit/internal/graph",
			want: []string{"internal/graph/a.go", "internal/graph/b.go"},
		},
		{
			name: "a file never imports itself",
			rel:  "internal/graph/a.go", spec: "github.com/plaesy/spec-kit/internal/graph",
			want: []string{"internal/graph/b.go"},
		},
		{
			name: "the longest matching module wins",
			rel:  "scripts/cmd/plaesy/main.go", spec: "github.com/plaesy/spec-kit/scripts/cmd/plaesy",
			want: nil,
		},
		{
			name: "a spec outside every module is dropped",
			rel:  "internal/graph/a.go", spec: "github.com/other/repo",
			want: nil,
		},
		{
			name: "a file outside any module has no imports",
			rel:  "docs/x.go", spec: "github.com/plaesy/spec-kit/internal/graph",
			want: []string{"internal/graph/a.go", "internal/graph/b.go"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveGoImport(tt.rel, tt.spec, gomod, dirFiles)
			if len(got) != len(tt.want) {
				t.Fatalf("resolveGoImport(%q) = %v, want %v", tt.spec, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("resolveGoImport(%q) = %v, want %v", tt.spec, got, tt.want)
				}
			}
		})
	}
}

func TestIndexJavaFiles(t *testing.T) {
	files := []string{
		"src/main/java/com/example/App.java",
		"src/main/java/com/example/util/Helper.java",
		"Root.java", // no source-root folder: not indexed
	}
	// The key drops the standard source-root prefix (src/main/java/) and turns
	// the rest into a dotted name, which is what `import com.example.App;` writes.
	got := indexJavaFiles(files)
	if got["com.example.App"] != "src/main/java/com/example/App.java" {
		t.Errorf("indexJavaFiles[com.example.App] = %q", got["com.example.App"])
	}
	if got["com.example.util.Helper"] != "src/main/java/com/example/util/Helper.java" {
		t.Errorf("indexJavaFiles[com.example.util.Helper] = %q", got["com.example.util.Helper"])
	}
	if _, ok := got["Root"]; ok {
		t.Errorf("a root-level .java file must not be indexed, got %v", got)
	}
}

// writeRepoTree materializes a repo fixture and returns its absolute path in
// forward-slash form, which is the shape every extractor takes.
func writeRepoTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, content := range files {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", rel, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
	return toSlash(root)
}

// The scan finds manifests on disk rather than among the symbol-extraction
// candidates, because go.mod is not a source extension and could never appear in
// `known` — which is why this used to return an empty map for every real project.
//
// vendor/ is deliberately NOT scanned. The source scan excludes it too, so a
// vendored module can never be a node in the graph; recording its module path
// would let a Go import resolve to a node that does not exist.
func TestScanGoModules(t *testing.T) {
	root := writeRepoTree(t, map[string]string{
		"go.mod":          "module github.com/plaesy/spec-kit\n\ngo 1.22\n",
		"scripts/go.mod":  "module github.com/plaesy/spec-kit/scripts\n",
		"vendor/go.mod":   "module example.com/vendored\n",
		"scripts/main.go": "package main\n",
		"scripts/go.sum":  "",
	})
	got := scanGoModules(root, map[string]bool{"scripts/main.go": true})
	want := map[string]string{
		"":        "github.com/plaesy/spec-kit",
		"scripts": "github.com/plaesy/spec-kit/scripts",
	}
	if len(got) != len(want) {
		t.Fatalf("scanGoModules = %v, want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("scanGoModules[%q] = %q, want %q", k, got[k], v)
		}
	}
	if _, ok := got["vendor"]; ok {
		t.Errorf("a vendored module must not be recorded: the source scan excludes vendor/, so it can never be a node (%v)", got)
	}
}

func TestScanDartPubspecs(t *testing.T) {
	root := writeRepoTree(t, map[string]string{
		"pubspec.yaml":              "name: app\n",
		"packages/lib/pubspec.yaml": "name: mylib\n",
		"packages/lib/lib/a.dart":   "class A {}\n",
	})
	got := scanDartPubspecs(root, map[string]bool{
		"pubspec.yaml": true, "packages/lib/pubspec.yaml": true, "packages/lib/lib/a.dart": true,
	})
	if got["app"] != "./lib" {
		t.Errorf("root package lib dir = %q, want %q", got["app"], "./lib")
	}
	if got["mylib"] != "packages/lib/lib" {
		t.Errorf("nested package lib dir = %q, want %q", got["mylib"], "packages/lib/lib")
	}
}

func TestExtractEdges(t *testing.T) {
	root := writeRepoTree(t, map[string]string{
		"README.md":         "See [the spec](docs/spec.md) and [gone](docs/missing.md).\n",
		"docs/spec.md":      "Back to [readme](../README.md).\n",
		"scripts/run.sh":    ". ./lib.sh\nsource ./other.sh\n",
		"scripts/lib.sh":    "echo hi\n",
		"scripts/other.sh":  "echo bye\n",
		"scripts/build.ps1": ". $PSScriptRoot\\lib.ps1\n& \"scripts/build.ps1\"\n",
		"scripts/lib.ps1":   "Write-Host 1\n",
		"app/main.js":       "import {x} from './util';\nconst y = require('./missing');\n",
		"app/util.js":       "export const x = 1;\n",
		"pkg/main.py":       "import mod\nfrom pkg.helpers import h\n",
		"pkg/mod.py":        "x = 1\n",
		"pkg/helpers.py":    "def h(): pass\n",
	})
	files := []string{
		"README.md", "docs/spec.md", "scripts/run.sh", "scripts/lib.sh", "scripts/other.sh",
		"scripts/build.ps1", "scripts/lib.ps1", "app/main.js", "app/util.js",
		"pkg/main.py", "pkg/mod.py", "pkg/helpers.py",
	}
	known := map[string]bool{}
	for _, f := range files {
		known[f] = true
	}

	edges := extractEdges(root, files, known)
	got := map[edgeKey]bool{}
	for _, e := range edges {
		if e.Confidence != "EXTRACTED" {
			t.Errorf("extractEdges produced a %q edge, want EXTRACTED only: %+v", e.Confidence, e)
		}
		got[edgeKey{e.Source, e.Target, e.Type}] = true
	}
	want := []edgeKey{
		{"README.md", "docs/spec.md", "references"},
		{"docs/spec.md", "README.md", "references"},
		{"scripts/run.sh", "scripts/lib.sh", "sources"},
		{"scripts/run.sh", "scripts/other.sh", "sources"},
		{"scripts/build.ps1", "scripts/lib.ps1", "sources"},
		{"app/main.js", "app/util.js", "imports"},
		{"pkg/main.py", "pkg/mod.py", "imports"},
		{"pkg/main.py", "pkg/helpers.py", "imports"},
	}
	for _, k := range want {
		if !got[k] {
			t.Errorf("missing edge %+v; got %v", k, keysOf(got))
		}
	}
	// A link to a file that is not in the file set is not a node, so it is not
	// an edge: docs/missing.md and app/missing are absent from the fixture.
	for _, k := range []edgeKey{
		{"README.md", "docs/missing.md", "references"},
		{"app/main.js", "app/missing.js", "imports"},
	} {
		if got[k] {
			t.Errorf("edge to a file outside the file set: %+v", k)
		}
	}
	// No self-edge: a file that links to itself is not a relationship.
	for k := range got {
		if k.src == k.tgt {
			t.Errorf("self edge %+v", k)
		}
	}
}

func keysOf(m map[edgeKey]bool) []string {
	var out []string
	for k := range m {
		out = append(out, k.src+" -> "+k.tgt+" ("+k.typ+")")
	}
	sort.Strings(out)
	return out
}
