package graph

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"
)

// newBuildRoot creates a fixture root one level below t.TempDir(). This
// machine's %TEMP% is C:\msys64\tmp, and a build writes its output dir two
// levels under the root, so a flat TempDir plus a long test name overflows
// MAX_PATH. Nothing here ever writes inside the real repository.
func newBuildRoot(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "plaesy-test", "project")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q): %v", root, err)
	}
	return root
}

func writeFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		full := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("MkdirAll(%q): %v", filepath.Dir(full), err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatalf("WriteFile(%q): %v", full, err)
		}
	}
}

func TestResolvePaths(t *testing.T) {
	root := newBuildRoot(t)

	t.Run("default out dir", func(t *testing.T) {
		p, err := ResolvePaths(Options{RepoPath: root})
		if err != nil {
			t.Fatalf("ResolvePaths() error = %v", err)
		}
		if p.RepoRoot != root {
			t.Errorf("RepoRoot = %q, want %q", p.RepoRoot, root)
		}
		if p.RepoRootSlash != toSlash(root) {
			t.Errorf("RepoRootSlash = %q, want %q", p.RepoRootSlash, toSlash(root))
		}
		if want := filepath.Join(root, ".plaesy", "analysis"); p.OutFull != want {
			t.Errorf("OutFull = %q, want %q", p.OutFull, want)
		}
		if want := filepath.Join(p.OutFull, "project.graph.json"); p.ProjectJSON != want {
			t.Errorf("ProjectJSON = %q, want %q", p.ProjectJSON, want)
		}
		if want := filepath.Join(p.OutFull, ".fingerprint"); p.FingerprintFile != want {
			t.Errorf("FingerprintFile = %q, want %q", p.FingerprintFile, want)
		}
	})

	t.Run("custom out dir is joined onto the repo root", func(t *testing.T) {
		p, err := ResolvePaths(Options{RepoPath: root, OutDir: "analysis"})
		if err != nil {
			t.Fatalf("ResolvePaths() error = %v", err)
		}
		if want := filepath.Join(root, "analysis"); p.OutFull != want {
			t.Errorf("OutFull = %q, want %q", p.OutFull, want)
		}
	})

	// An absolute --outdir names a directory outside the repo, and must be
	// honoured as given. filepath.Join does not do this: it folds
	// "/tmp/kg" into "<repo>/tmp/kg" on Unix, and on Windows it mangles the
	// drive letter into a path segment, so a real `graph build --outdir
	// C:/tmp/kg` died with `mkdir <repo>\C::`.
	t.Run("absolute out dir is used as given", func(t *testing.T) {
		outside := filepath.Join(t.TempDir(), "plaesy-test", "kg")
		p, err := ResolvePaths(Options{RepoPath: root, OutDir: outside})
		if err != nil {
			t.Fatalf("ResolvePaths() error = %v", err)
		}
		if p.OutFull != outside {
			t.Errorf("OutFull = %q, want the absolute path %q", p.OutFull, outside)
		}
		// It must genuinely be outside the repo, not merely spelled that way.
		if rel, relErr := filepath.Rel(root, p.OutFull); relErr == nil && !strings.HasPrefix(rel, "..") {
			t.Errorf("OutFull = %q resolved inside the repo (rel %q); an absolute --outdir must stay outside", p.OutFull, rel)
		}
	})

	// The root that every internal comparison is made against has to be the
	// absolute one, even when the user passed a relative --path. The Windows
	// defect this package already had once (a root that came back as
	// "/C:/repo" instead of "C:/repo") lives entirely in this value.
	t.Run("a relative repo path is resolved against the working directory", func(t *testing.T) {
		abs, err := filepath.Abs(root)
		if err != nil {
			t.Fatalf("Abs: %v", err)
		}
		old, err := os.Getwd()
		if err != nil {
			t.Fatalf("Getwd: %v", err)
		}
		if err := os.Chdir(filepath.Dir(abs)); err != nil {
			t.Fatalf("Chdir: %v", err)
		}
		t.Cleanup(func() {
			if err := os.Chdir(old); err != nil {
				t.Fatalf("restoring cwd: %v", err)
			}
		})
		// The expected root is derived from a fresh Getwd after the Chdir,
		// not from the pre-Chdir `abs`: on macOS, a tmp dir under
		// /var/folders/... is reached through /var, a symlink to
		// /private/var, and Getwd (unlike filepath.Abs on an already-absolute
		// path) returns the kernel's resolved, symlink-free view —
		// "/private/var/folders/...". Comparing against the unresolved `abs`
		// would make this assertion fail on every macOS run for a reason
		// that has nothing to do with ResolvePaths' own correctness.
		wd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Getwd after Chdir: %v", err)
		}
		want := filepath.Join(wd, filepath.Base(abs))
		p, err := ResolvePaths(Options{RepoPath: filepath.Base(abs)})
		if err != nil {
			t.Fatalf("ResolvePaths() error = %v", err)
		}
		if p.RepoRoot != want {
			t.Errorf("RepoRoot = %q, want %q", p.RepoRoot, want)
		}
		if p.RepoRootSlash != toSlash(want) {
			t.Errorf("RepoRootSlash = %q, want %q", p.RepoRootSlash, toSlash(want))
		}
		if !strings.HasPrefix(p.ProjectJSON, p.RepoRootSlash+"/") && p.RepoRootSlash != toSlash(p.RepoRoot) {
			t.Errorf("ProjectJSON %q is not under RepoRootSlash %q", p.ProjectJSON, p.RepoRootSlash)
		}
		if got := outDirNorm(Options{RepoPath: filepath.Base(abs)}, p); got != ".plaesy/analysis" {
			t.Errorf("outDirNorm() = %q, want %q", got, ".plaesy/analysis")
		}
	})
}

func TestOutDirNorm(t *testing.T) {
	root := newBuildRoot(t)
	tests := []struct {
		name   string
		outDir string
		want   string
	}{
		{name: "empty falls back to the default", outDir: "", want: ".plaesy/analysis"},
		{name: "default", outDir: ".plaesy/analysis", want: ".plaesy/analysis"},
		{name: "non-dot output dir", outDir: "analysis", want: "analysis"},
		{name: "leading dot-slash is cleaned by Join", outDir: "./analysis", want: "analysis"},
		{name: "nested", outDir: "out/graph", want: "out/graph"},
		{name: "trailing slash", outDir: "analysis/", want: "analysis"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			opts := Options{RepoPath: root, OutDir: tc.outDir}
			p, err := ResolvePaths(opts)
			if err != nil {
				t.Fatalf("ResolvePaths() error = %v", err)
			}
			if got := outDirNorm(opts, p); got != tc.want {
				t.Fatalf("outDirNorm(%q) = %q, want %q", tc.outDir, got, tc.want)
			}
		})
	}

	// --outdir pointing outside the repo cannot be expressed as a relative
	// path, so outDirNorm gives up and returns "". The scan then excludes
	// nothing, which is correct: the output is not inside the tree being
	// scanned, so there is nothing to skip.
	t.Run("output dir outside the repo yields the no-exclusion sentinel", func(t *testing.T) {
		outside := filepath.Join(filepath.Dir(root), "elsewhere")
		opts := Options{RepoPath: root, OutDir: filepath.Join("..", filepath.Base(outside))}
		p, err := ResolvePaths(opts)
		if err != nil {
			t.Fatalf("ResolvePaths() error = %v", err)
		}
		if p.OutFull != outside {
			t.Fatalf("OutFull = %q, want %q", p.OutFull, outside)
		}
		if got := outDirNorm(opts, p); got != "../elsewhere" {
			t.Fatalf("outDirNorm() = %q, want %q", got, "../elsewhere")
		}
	})
}

// buildFixture is the tree most Build tests share: one markdown link pair, one
// shell dot-source, one JS import and one duplicate basename (the mirror edge).
// Every extension here is in INCLUDE_EXT_RE, so every file becomes a node.
func buildFixture() map[string]string {
	return map[string]string{
		"README.md":          "See [the spec](docs/spec.md).\n",
		"docs/spec.md":       "## Spec\n\nBack to [readme](../README.md).\n",
		"scripts/run.sh":     ". ./lib.sh\n",
		"scripts/lib.sh":     "lib_helper() {\n  echo hi\n}\n",
		"app/main.js":        "import {x} from './util';\n",
		"app/util.js":        "export const x = 1;\n",
		"other/lib.sh":       "echo other\n",
		"notes/unmatched.md": "nothing in here\n",
	}
}

func edgeSet(edges []Edge) []string {
	out := make([]string, 0, len(edges))
	for _, e := range edges {
		out = append(out, e.Source+" -> "+e.Target+" ("+e.Type+"/"+e.Confidence+")")
	}
	sort.Strings(out)
	return out
}

func TestBuild(t *testing.T) {
	root := newBuildRoot(t)
	writeFiles(t, root, buildFixture())

	opts := Options{RepoPath: root}
	p, err := ResolvePaths(opts)
	if err != nil {
		t.Fatalf("ResolvePaths: %v", err)
	}
	g, err := Build(opts, p)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// The output dir is created by Build, and it must not become a node.
	if _, err := os.Stat(p.OutFull); err != nil {
		t.Errorf("Build did not create the output dir: %v", err)
	}

	t.Run("one node per collected file, in sorted id order", func(t *testing.T) {
		var ids []string
		for _, n := range g.Nodes {
			ids = append(ids, n.ID)
		}
		want := []string{
			"README.md", "app/main.js", "app/util.js", "docs/spec.md",
			"notes/unmatched.md", "other/lib.sh", "scripts/lib.sh", "scripts/run.sh",
		}
		if !reflect.DeepEqual(ids, want) {
			t.Fatalf("node ids = %v, want %v", ids, want)
		}
	})

	t.Run("metadata is derived per node", func(t *testing.T) {
		byID := map[string]Node{}
		for _, n := range g.Nodes {
			byID[n.ID] = n
		}
		tests := []struct {
			id                       string
			label, typ, group, layer string
		}{
			{id: "README.md", label: "README.md", typ: "other", group: "README.md", layer: ""},
			{id: "docs/spec.md", label: "spec.md", typ: "doc", group: "docs", layer: ""},
			{id: "scripts/run.sh", label: "run.sh", typ: "script", group: "scripts", layer: ""},
			{id: "app/util.js", label: "util.js", typ: "source", group: "app", layer: "Utility"},
		}
		for _, tc := range tests {
			n, ok := byID[tc.id]
			if !ok {
				t.Errorf("no node for %q", tc.id)
				continue
			}
			if n.Label != tc.label || n.Type != tc.typ || n.Group != tc.group || n.Layer != tc.layer {
				t.Errorf("node %q = {label:%q type:%q group:%q layer:%q}, want {%q %q %q %q}",
					tc.id, n.Label, n.Type, n.Group, n.Layer, tc.label, tc.typ, tc.group, tc.layer)
			}
		}
	})

	// Every structural reference kind the fixture contains. The two
	// "lib.sh" files also produce a mirrors edge, which is the point of
	// including them.
	t.Run("structural edges", func(t *testing.T) {
		want := []string{
			"README.md -> docs/spec.md (references/EXTRACTED)",
			"app/main.js -> app/util.js (imports/EXTRACTED)",
			"docs/spec.md -> README.md (references/EXTRACTED)",
			"other/lib.sh -> scripts/lib.sh (mirrors/EXTRACTED)",
			"scripts/run.sh -> scripts/lib.sh (sources/EXTRACTED)",
		}
		if got := edgeSet(g.Edges); !reflect.DeepEqual(got, want) {
			t.Fatalf("edges = %v, want %v", got, want)
		}
	})

	// A markdown link is also a plain-text mention of the same path. The weak
	// mention must be dropped in favour of the real reference, and the
	// unmatched file must keep no edges at all.
	t.Run("weak mentions are dropped in favour of real edges", func(t *testing.T) {
		for _, e := range g.Edges {
			if e.Type == "mentions" {
				t.Errorf("unexpected mention edge %+v; every mention in this fixture duplicates a real edge", e)
			}
		}
		if d := g.Edges; len(d) != 5 {
			t.Fatalf("edge count = %d, want 5", len(d))
		}
	})

	t.Run("degree counts both endpoints of every edge", func(t *testing.T) {
		byID := map[string]int{}
		for _, n := range g.Nodes {
			byID[n.ID] = n.Degree
		}
		// The two markdown files link to each other, so each is at both ends
		// of two edges; scripts/lib.sh is at the end of a sources edge and a
		// mirrors edge; notes/unmatched.md is an orphan.
		want := map[string]int{
			"README.md": 2, "docs/spec.md": 2, "scripts/run.sh": 1,
			"scripts/lib.sh": 2, "other/lib.sh": 1, "app/main.js": 1,
			"app/util.js": 1, "notes/unmatched.md": 0,
		}
		if !reflect.DeepEqual(byID, want) {
			t.Fatalf("degrees = %v, want %v", byID, want)
		}
	})

	// Communities come from edge structure, not folder names: the mirror edge
	// is what puts other/lib.sh in the same community as scripts/lib.sh.
	t.Run("communities are derived from edges, not folders", func(t *testing.T) {
		comm := map[string]int{}
		for _, n := range g.Nodes {
			comm[n.ID] = n.Community
		}
		if comm["other/lib.sh"] != comm["scripts/lib.sh"] {
			t.Errorf("mirrored files are in different communities: %d vs %d", comm["other/lib.sh"], comm["scripts/lib.sh"])
		}
		if comm["README.md"] != comm["docs/spec.md"] {
			t.Errorf("linked files are in different communities: %d vs %d", comm["README.md"], comm["docs/spec.md"])
		}
		if comm["app/main.js"] != comm["app/util.js"] {
			t.Errorf("importing files are in different communities: %d vs %d", comm["app/main.js"], comm["app/util.js"])
		}
		// Four components, none of them merged: 0..3, each used once or twice.
		seen := map[int]int{}
		for _, c := range comm {
			seen[c]++
		}
		if len(seen) != 4 {
			t.Fatalf("got %d communities %v, want 4", len(seen), seen)
		}
		for id := 0; id < 4; id++ {
			if seen[id] == 0 {
				t.Errorf("community ids are not contiguous from zero: %v", seen)
			}
		}
	})

	t.Run("symbols are collected per file", func(t *testing.T) {
		byID := map[string][]string{}
		for _, n := range g.Nodes {
			if len(n.Symbols) > 0 {
				byID[n.ID] = n.Symbols
			}
		}
		want := map[string][]string{
			"docs/spec.md":   {"Spec"},
			"scripts/lib.sh": {"lib_helper"},
		}
		if !reflect.DeepEqual(byID, want) {
			t.Fatalf("symbols = %v, want %v", byID, want)
		}
	})

	t.Run("header fields describe the build", func(t *testing.T) {
		if g.Root != root {
			t.Errorf("Root = %q, want %q", g.Root, root)
		}
		if len(g.Nodes) != 8 {
			t.Errorf("Description should count 8 files: %q", g.Description)
		}
		if !strings.Contains(g.Description, filepath.Base(root)) {
			t.Errorf("Description %q should name the project %q", g.Description, filepath.Base(root))
		}
		if _, err := time.Parse("2006-01-02 15:04:05", g.GeneratedAt); err != nil {
			t.Errorf("GeneratedAt %q is not a timestamp: %v", g.GeneratedAt, err)
		}
	})

	// The bash original iterated a hash-ordered associative array; this port
	// sorts, so two builds of the same tree must be byte-identical. A
	// regression here would make project.graph.json churn on every run.
	t.Run("two builds of the same tree are identical", func(t *testing.T) {
		second, err := Build(opts, p)
		if err != nil {
			t.Fatalf("Build() error = %v", err)
		}
		if !reflect.DeepEqual(edgeSet(second.Edges), edgeSet(g.Edges)) {
			t.Errorf("edge set changed between runs:\n%v\n%v", edgeSet(second.Edges), edgeSet(g.Edges))
		}
		strip := func(x *Graph) []Node {
			out := append([]Node(nil), x.Nodes...)
			for i := range out {
				out[i].Community = 0
			}
			return out
		}
		if !reflect.DeepEqual(strip(second), strip(g)) {
			t.Error("node set changed between runs")
		}
	})
}

func TestBuildEmptyRepo(t *testing.T) {
	root := newBuildRoot(t)
	opts := Options{RepoPath: root}
	p, err := ResolvePaths(opts)
	if err != nil {
		t.Fatalf("ResolvePaths: %v", err)
	}
	g, err := Build(opts, p)
	if err == nil {
		t.Fatalf("Build() on an empty tree = %v, want an error", g)
	}
	if !strings.Contains(err.Error(), "no files found under") || !strings.Contains(err.Error(), root) {
		t.Errorf("error %q should name the root it scanned", err)
	}
}

// A tree whose only candidate files live under an excluded or the output
// directory has nothing to graph, and says so rather than emitting an empty
// graph.
func TestBuildRepoWithNoCandidateFiles(t *testing.T) {
	root := newBuildRoot(t)
	writeFiles(t, root, map[string]string{
		"data.json":           "{}",
		"node_modules/dep.md": "# dep\n",
		".git/hooks/pre.md":   "# hook\n",
		"analysis/project.go": "package main\n",
		"analysis/data.json":  "{}",
	})
	opts := Options{RepoPath: root, OutDir: "analysis"}
	p, err := ResolvePaths(opts)
	if err != nil {
		t.Fatalf("ResolvePaths: %v", err)
	}
	if _, err := Build(opts, p); err == nil {
		t.Fatal("Build() = nil error, want \"no files found\"")
	}
}

// Build must not swallow its own output: on the second run the JSON, HTML and
// report it wrote must not become nodes. Only reports.md could qualify (the
// other two are not in INCLUDE_EXT_RE), which is exactly why the exclusion
// exists at all.
func TestBuildExcludesItsOwnOutput(t *testing.T) {
	root := newBuildRoot(t)
	writeFiles(t, root, map[string]string{
		"README.md":    "# Title\n\nSee [spec](docs/spec.md).\n",
		"docs/spec.md": "## Spec\n",
	})
	opts := Options{RepoPath: root, OutDir: "analysis"}
	p, err := ResolvePaths(opts)
	if err != nil {
		t.Fatalf("ResolvePaths: %v", err)
	}
	for pass := 1; pass <= 2; pass++ {
		g, err := Build(opts, p)
		if err != nil {
			t.Fatalf("Build() pass %d error = %v", pass, err)
		}
		for _, n := range g.Nodes {
			if strings.HasPrefix(n.ID, "analysis/") {
				t.Fatalf("pass %d: the graph's own output %q became a node", pass, n.ID)
			}
		}
		if len(g.Nodes) != 2 {
			t.Fatalf("pass %d: %d nodes, want 2", pass, len(g.Nodes))
		}
		// Write the artifacts a real run writes, so pass 2 has them to skip.
		if err := SaveJSON(g, p.ProjectJSON); err != nil {
			t.Fatalf("SaveJSON: %v", err)
		}
		if err := WriteReport(g, p.OutFull, ""); err != nil {
			t.Fatalf("WriteReport: %v", err)
		}
		if err := WriteHTML(g, p.OutFull); err != nil {
			t.Fatalf("WriteHTML: %v", err)
		}
	}
}

func TestSourceFingerprint(t *testing.T) {
	// Pinned mtimes make the second half of the "count|mtime" pair a fixed
	// value instead of "whenever this test happened to run".
	fixed := time.Date(2021, 3, 4, 5, 6, 7, 0, time.UTC)
	stamp := func(t *testing.T, root string, rels ...string) {
		t.Helper()
		for _, rel := range rels {
			if err := os.Chtimes(filepath.Join(root, filepath.FromSlash(rel)), fixed, fixed); err != nil {
				t.Fatalf("Chtimes(%q): %v", rel, err)
			}
		}
	}

	t.Run("count and newest mtime of the scanned files", func(t *testing.T) {
		root := newBuildRoot(t)
		writeFiles(t, root, map[string]string{
			"a.md":        "# A\n",
			"docs/b.md":   "# B\n",
			"ignored.txt": "not a node\n",
		})
		stamp(t, root, "a.md", "docs/b.md", "ignored.txt")
		opts := Options{RepoPath: root}
		p, err := ResolvePaths(opts)
		if err != nil {
			t.Fatalf("ResolvePaths: %v", err)
		}
		got, err := SourceFingerprint(opts, p)
		if err != nil {
			t.Fatalf("SourceFingerprint() error = %v", err)
		}
		want := "2|" + itoa(fixed.Unix())
		if got != want {
			t.Fatalf("SourceFingerprint() = %q, want %q", got, want)
		}
	})

	// --if-changed and --watch both rebuild when this string moves, so both a
	// new file and an edited file have to change it.
	t.Run("adding a file changes it", func(t *testing.T) {
		root := newBuildRoot(t)
		writeFiles(t, root, map[string]string{"a.md": "# A\n"})
		stamp(t, root, "a.md")
		opts := Options{RepoPath: root}
		p, _ := ResolvePaths(opts)
		before, err := SourceFingerprint(opts, p)
		if err != nil {
			t.Fatalf("SourceFingerprint() error = %v", err)
		}
		writeFiles(t, root, map[string]string{"b.md": "# B\n"})
		stamp(t, root, "b.md")
		after, err := SourceFingerprint(opts, p)
		if err != nil {
			t.Fatalf("SourceFingerprint() error = %v", err)
		}
		if before == after {
			t.Fatalf("fingerprint did not change after adding a file: %q", after)
		}
	})

	t.Run("touching a file changes it", func(t *testing.T) {
		root := newBuildRoot(t)
		writeFiles(t, root, map[string]string{"a.md": "# A\n"})
		stamp(t, root, "a.md")
		opts := Options{RepoPath: root}
		p, _ := ResolvePaths(opts)
		before, _ := SourceFingerprint(opts, p)
		later := fixed.Add(2 * time.Hour)
		if err := os.Chtimes(filepath.Join(root, "a.md"), later, later); err != nil {
			t.Fatalf("Chtimes: %v", err)
		}
		after, err := SourceFingerprint(opts, p)
		if err != nil {
			t.Fatalf("SourceFingerprint() error = %v", err)
		}
		if after == before {
			t.Fatalf("fingerprint did not change after touching a file: %q", after)
		}
		if want := "1|" + itoa(later.Unix()); after != want {
			t.Fatalf("SourceFingerprint() = %q, want %q", after, want)
		}
	})

	// The fingerprint is what --if-changed compares after a build, and a build
	// writes project.graph.json and reports.md into the output dir. If those
	// counted, every run would invalidate the next one.
	t.Run("the output dir is not part of the signature", func(t *testing.T) {
		root := newBuildRoot(t)
		writeFiles(t, root, map[string]string{"a.md": "# A\n"})
		opts := Options{RepoPath: root, OutDir: "analysis"}
		p, _ := ResolvePaths(opts)
		before, _ := SourceFingerprint(opts, p)
		g := &Graph{Nodes: []Node{{ID: "a.md", Type: "other", Group: "a.md"}}}
		if err := os.MkdirAll(p.OutFull, 0o755); err != nil {
			t.Fatalf("MkdirAll(%q): %v", p.OutFull, err)
		}
		if err := SaveJSON(g, p.ProjectJSON); err != nil {
			t.Fatalf("SaveJSON: %v", err)
		}
		if err := WriteReport(g, p.OutFull, ""); err != nil {
			t.Fatalf("WriteReport: %v", err)
		}
		after, err := SourceFingerprint(opts, p)
		if err != nil {
			t.Fatalf("SourceFingerprint() error = %v", err)
		}
		if after != before {
			t.Fatalf("writing the graph's own output changed the fingerprint: %q -> %q", before, after)
		}
	})

	// An empty tree is a real state (a project whose files are all excluded),
	// and it has to produce a signature rather than an error.
	t.Run("empty tree", func(t *testing.T) {
		root := newBuildRoot(t)
		opts := Options{RepoPath: root}
		p, _ := ResolvePaths(opts)
		got, err := SourceFingerprint(opts, p)
		if err != nil {
			t.Fatalf("SourceFingerprint() error = %v", err)
		}
		// "0|0": with no files there is no newest timestamp, and zero says so.
		// This used to be "0|-62135596800" — the Unix value of the zero
		// time.Time, i.e. 0001-01-01 — because maxMtime was never assigned when
		// the loop had nothing to iterate. It was stable, so --if-changed still
		// worked, but the value is written into a user-visible .fingerprint file
		// and reads as corruption rather than as "nothing here yet".
		if got != "0|0" {
			t.Fatalf("SourceFingerprint() = %q, want %q", got, "0|0")
		}
	})

	// A file the walk can list but the stat cannot follow is counted and then
	// skipped, not fatal: --watch would otherwise abort on a dangling symlink
	// or a file deleted between the walk and the stat.
	t.Run("a file that cannot be stat'ed is skipped, not fatal", func(t *testing.T) {
		root := newBuildRoot(t)
		writeFiles(t, root, map[string]string{"a.md": "# A\n"})
		if err := os.Symlink(filepath.Join(root, "missing.md"), filepath.Join(root, "dangling.md")); err != nil {
			t.Skipf("this machine cannot create symlinks: %v", err)
		}
		opts := Options{RepoPath: root}
		p, _ := ResolvePaths(opts)
		got, err := SourceFingerprint(opts, p)
		if err != nil {
			t.Fatalf("SourceFingerprint() error = %v", err)
		}
		// Counted (the walk lists it), but it contributes no mtime — so the
		// timestamp half of the signature is the zero time's.
		if !strings.HasPrefix(got, "2|") {
			t.Fatalf("SourceFingerprint() = %q, want the dangling link counted", got)
		}
	})
}

// An unreadable file must not cost the whole graph. collectFiles can list a
// file whose content is empty or unreadable, and the symbols pass is the one
// that would otherwise index a bogus symbol list for it.
func TestBuildSkipsUnreadableFiles(t *testing.T) {
	root := newBuildRoot(t)
	writeFiles(t, root, map[string]string{
		"README.md": "# Title\n",
		"empty.md":  "",
		"broken.md": "# Broken\n",
	})
	if err := os.Symlink(filepath.Join(root, "missing.md"), filepath.Join(root, "dangling.md")); err != nil {
		t.Logf("no dangling symlink in the fixture: %v", err)
	}
	opts := Options{RepoPath: root}
	p, _ := ResolvePaths(opts)
	g, err := Build(opts, p)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	byID := map[string]Node{}
	for _, n := range g.Nodes {
		byID[n.ID] = n
	}
	// The empty file and the unreadable link are nodes with no content: they
	// must exist and carry no symbols rather than abort the build.
	for _, id := range []string{"empty.md", "dangling.md"} {
		n, ok := byID[id]
		if !ok {
			t.Errorf("no node for %q", id)
			continue
		}
		if len(n.Symbols) != 0 {
			t.Errorf("node %q has symbols %v, want none", id, n.Symbols)
		}
	}
	if got := byID["README.md"].Symbols; !reflect.DeepEqual(got, []string{"Title"}) {
		t.Errorf("README.md symbols = %v, want [Title]", got)
	}
	if got := byID["broken.md"].Symbols; !reflect.DeepEqual(got, []string{"Broken"}) {
		t.Errorf("broken.md symbols = %v, want [Broken]", got)
	}
}

// Build creates its output directory before it scans. When that directory
// cannot be created the run has to fail there, not later with a confusing
// "no files found" — the user needs to know the output path is the problem.
func TestBuildFailsWhenTheOutputDirCannotBeCreated(t *testing.T) {
	root := newBuildRoot(t)
	writeFiles(t, root, map[string]string{"README.md": "# Title\n"})
	// A regular file where a directory would have to be.
	writeFiles(t, root, map[string]string{"blocker": "x"})
	opts := Options{RepoPath: root, OutDir: "blocker/analysis"}
	p, _ := ResolvePaths(opts)
	if _, err := Build(opts, p); err == nil {
		t.Fatal("Build() = nil error; want the mkdir failure")
	}
}

// Two edges between the same pair of files with different types (a PowerShell
// file that both dot-sources and calls the same script) must both survive and
// be ordered by type, which is the branch the edge sort's type tie-break
// exists for.
func TestBuildOrdersEdgesByTypeForTheSamePair(t *testing.T) {
	root := newBuildRoot(t)
	writeFiles(t, root, map[string]string{
		"run.ps1":         ". ./scripts/lib.ps1\n& \"scripts/lib.ps1\"\n",
		"scripts/lib.ps1": "Write-Host 1\n",
	})
	opts := Options{RepoPath: root}
	p, _ := ResolvePaths(opts)
	g, err := Build(opts, p)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	want := []string{
		"run.ps1 -> scripts/lib.ps1 (calls/EXTRACTED)",
		"run.ps1 -> scripts/lib.ps1 (sources/EXTRACTED)",
	}
	if got := edgeSet(g.Edges); !reflect.DeepEqual(got, want) {
		t.Fatalf("edges = %v, want %v (same source and target, ordered by type)", got, want)
	}
}

// The other sort tie-break: one source pointing at two different targets. The
// comparator reaches the target comparison exactly when the sources match, so
// a graph where every source has at most one edge never runs it.
func TestBuildOrdersEdgesByTargetForTheSameSource(t *testing.T) {
	root := newBuildRoot(t)
	writeFiles(t, root, map[string]string{
		"index.md": "See [a](a.md) and [b](b.md).\n",
		"a.md":     "# A\n",
		"b.md":     "# B\n",
	})
	opts := Options{RepoPath: root}
	p, _ := ResolvePaths(opts)
	g, err := Build(opts, p)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	want := []string{
		"index.md -> a.md (references/EXTRACTED)",
		"index.md -> b.md (references/EXTRACTED)",
	}
	if got := edgeSet(g.Edges); !reflect.DeepEqual(got, want) {
		t.Fatalf("edges = %v, want %v", got, want)
	}
}

// outDirNorm is the only guard that keeps the graph from eating its own
// reports.md when the output dir is somewhere other than a dot-directory. When
// it cannot express the output dir as a relative path it has to say "exclude
// nothing" rather than return a prefix that would match unrelated folders.
func TestOutDirNormWhenRelFails(t *testing.T) {
	var p Paths
	if runtime.GOOS == "windows" {
		// Rel refuses to relate two paths on different volumes.
		p = Paths{RepoRoot: `C:\repo`, OutFull: `D:\out`}
	} else {
		// Rel refuses to relate an absolute path to a relative one.
		p = Paths{RepoRoot: "/repo", OutFull: "out"}
	}
	if got := outDirNorm(Options{RepoPath: p.RepoRoot}, p); got != "" {
		t.Fatalf("outDirNorm() = %q, want the no-exclusion sentinel", got)
	}
}

// itoa keeps the fingerprint assertions readable without pulling strconv into
// the test file's import list for two call sites.
func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		return "-" + string(digits)
	}
	return string(digits)
}
