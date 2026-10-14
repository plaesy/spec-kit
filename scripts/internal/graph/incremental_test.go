package graph

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

// buildPair builds the repo incrementally (into the default analysis directory,
// which keeps the cache under test) and in full (into a separate directory, so
// its result does not touch that cache). It returns both graphs. fileWork is
// left counting only the incremental build.
func buildPair(t *testing.T, repo string) (inc, full *Graph) {
	t.Helper()
	incOpts := Options{RepoPath: repo}
	incPaths, err := ResolvePaths(incOpts)
	if err != nil {
		t.Fatal(err)
	}
	inc, err = Build(incOpts, incPaths)
	if err != nil {
		t.Fatalf("incremental build: %v", err)
	}
	incWork := fileWork
	assertFileResultsMatchReference(t, repo, incPaths)

	fullOpts := Options{RepoPath: repo, OutDir: ".plaesy/expected", Full: true}
	fullPaths, err := ResolvePaths(fullOpts)
	if err != nil {
		t.Fatal(err)
	}
	full, err = Build(fullOpts, fullPaths)
	if err != nil {
		t.Fatalf("full build: %v", err)
	}
	// The full build recomputes everything by design. Keep the incremental count.
	fileWork = incWork
	return inc, full
}

// assertFileResultsMatchReference checks the cached per-file results against the
// original, uncached functions. Those are the oracle. A full build reuses the
// same per-file code as the cached build, so comparing two builds cannot catch
// a bug in that shared code.
func assertFileResultsMatchReference(t *testing.T, repo string, paths Paths) {
	t.Helper()
	opts := Options{RepoPath: repo}
	files, err := collectFiles(paths.RepoRoot, outDirNorm(opts, paths))
	if err != nil {
		t.Fatal(err)
	}
	known := make(map[string]bool, len(files))
	for _, f := range files {
		known[f] = true
	}

	wantEdges := append(extractEdges(paths.RepoRootSlash, files, known), mentionEdges(paths.RepoRootSlash, files, files)...)
	wantSymbols := map[string][]string{}
	for _, rel := range files {
		content := readAll(joinRoot(paths.RepoRootSlash, rel))
		if content == "" {
			continue
		}
		if syms := symbolsOf(rel, content); len(syms) > 0 {
			wantSymbols[rel] = syms
		}
	}

	gotEdges, gotSymbols, err := fileResults(opts, paths, files, known)
	if err != nil {
		t.Fatal(err)
	}
	if !sameEdges(gotEdges, wantEdges) {
		t.Errorf("cached edges differ from the reference extraction\ncached:    %v\nreference: %v", gotEdges, wantEdges)
	}
	if !reflect.DeepEqual(gotSymbols, wantSymbols) {
		t.Errorf("cached symbols differ from the reference extraction\ncached:    %v\nreference: %v", gotSymbols, wantSymbols)
	}
}

// sameEdges compares two edge lists as multisets, ignoring order.
func sameEdges(a, b []Edge) bool {
	if len(a) != len(b) {
		return false
	}
	key := func(e Edge) string { return e.Source + "|" + e.Target + "|" + e.Type + "|" + e.Confidence }
	count := map[string]int{}
	for _, e := range a {
		count[key(e)]++
	}
	for _, e := range b {
		count[key(e)]--
	}
	for _, n := range count {
		if n != 0 {
			return false
		}
	}
	return true
}

func assertSameGraph(t *testing.T, inc, full *Graph) {
	t.Helper()
	if !reflect.DeepEqual(inc.Nodes, full.Nodes) {
		t.Errorf("incremental nodes differ from a full build\nincremental: %+v\nfull:        %+v", inc.Nodes, full.Nodes)
	}
	if !reflect.DeepEqual(inc.Edges, full.Edges) {
		t.Errorf("incremental edges differ from a full build\nincremental: %+v\nfull:        %+v", inc.Edges, full.Edges)
	}
}

func writeRepoFile(t *testing.T, repo, rel, body string) {
	t.Helper()
	path := filepath.Join(repo, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// fixtureRepo writes a small repo with a Go package, JS imports, a Markdown link
// and a plain-text mention, so every kind of per-file edge is exercised.
func fixtureRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	writeRepoFile(t, repo, "go.mod", "module example.com/fixture\n")
	writeRepoFile(t, repo, "src/a.go", "package src\n\n// A does the first thing.\nfunc A() {}\n")
	writeRepoFile(t, repo, "src/a.js", "export function AJS() {}\n")
	writeRepoFile(t, repo, "src/b.js", "import { AJS } from './a.js'\nexport function B() {}\n")
	writeRepoFile(t, repo, "docs/guide.md", "See [a](../src/a.go) and the file src/b.js for details.\n")
	return repo
}

func TestIncrementalBuildMatchesFullBuild(t *testing.T) {
	repo := fixtureRepo(t)

	// First build: no cache yet, so every file is computed.
	inc, full := buildPair(t, repo)
	assertSameGraph(t, inc, full)

	// Nothing changed: the cache answers for every file.
	fileWork = 0
	inc, full = buildPair(t, repo)
	assertSameGraph(t, inc, full)
	if fileWork != 0 {
		t.Errorf("unchanged repo recomputed %d files, want 0", fileWork)
	}

	// One source file edited: only that file is recomputed.
	writeRepoFile(t, repo, "src/a.go", "package src\n\n// A does the first thing.\nfunc A() {}\n\n// B2 is new.\nfunc B2() {}\n")
	fileWork = 0
	inc, full = buildPair(t, repo)
	assertSameGraph(t, inc, full)
	if fileWork != 1 {
		t.Errorf("one edited file recomputed %d files, want 1", fileWork)
	}
}

func TestIncrementalBuildTouchWithoutEditSkipsWork(t *testing.T) {
	repo := fixtureRepo(t)
	buildPair(t, repo)

	// A new mtime with identical content is not an edit: the hash matches.
	later := time.Now().Add(time.Hour)
	if err := os.Chtimes(filepath.Join(repo, "src", "a.go"), later, later); err != nil {
		t.Fatal(err)
	}
	fileWork = 0
	inc, full := buildPair(t, repo)
	assertSameGraph(t, inc, full)
	if fileWork != 0 {
		t.Errorf("touched but unedited file recomputed %d files, want 0", fileWork)
	}
}

func TestIncrementalBuildFallsBackWhenFileSetChanges(t *testing.T) {
	repo := fixtureRepo(t)
	buildPair(t, repo)

	// A new file can become the target of an existing import or mention, so
	// every file is recomputed.
	writeRepoFile(t, repo, "docs/new.md", "New page that mentions src/a.go.\n")
	fileWork = 0
	inc, full := buildPair(t, repo)
	assertSameGraph(t, inc, full)
	if fileWork != len(inc.Nodes) {
		t.Errorf("file added: recomputed %d files, want all %d", fileWork, len(inc.Nodes))
	}

	// A removed file is the same case in reverse.
	if err := os.Remove(filepath.Join(repo, "src", "a.js")); err != nil {
		t.Fatal(err)
	}
	fileWork = 0
	inc, full = buildPair(t, repo)
	assertSameGraph(t, inc, full)
	if fileWork != len(inc.Nodes) {
		t.Errorf("file removed: recomputed %d files, want all %d", fileWork, len(inc.Nodes))
	}
}

func TestIncrementalBuildFallsBackWhenManifestChanges(t *testing.T) {
	repo := fixtureRepo(t)
	buildPair(t, repo)

	// go.mod decides how Go imports resolve, so changing it recomputes everything.
	writeRepoFile(t, repo, "go.mod", "module example.com/renamed\n")
	fileWork = 0
	inc, full := buildPair(t, repo)
	assertSameGraph(t, inc, full)
	if fileWork != len(inc.Nodes) {
		t.Errorf("manifest changed: recomputed %d files, want all %d", fileWork, len(inc.Nodes))
	}
}

func TestIncrementalBuildIgnoresUnreadableCache(t *testing.T) {
	repo := fixtureRepo(t)
	buildPair(t, repo)

	// A cache that cannot be parsed is treated as absent, not as an error.
	cache := filepath.Join(repo, ".plaesy", "analysis", buildCacheName)
	if err := os.WriteFile(cache, []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	fileWork = 0
	inc, full := buildPair(t, repo)
	assertSameGraph(t, inc, full)
	if fileWork != len(inc.Nodes) {
		t.Errorf("corrupt cache: recomputed %d files, want all %d", fileWork, len(inc.Nodes))
	}
}
