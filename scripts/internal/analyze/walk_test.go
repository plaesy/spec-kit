package analyze

import (
	"errors"
	"io/fs"
	"path/filepath"
	"reflect"
	"testing"
)

func TestIsExcludedDirName(t *testing.T) {
	cases := []struct {
		name string
		dir  string
		want bool
	}{
		{"dot directory", ".git", true},
		{"dot prefix of a normal name", ".config", true},
		{"just a dot", ".", true},
		{"node_modules", "node_modules", true},
		{"vendor is NOT excluded (only dot dirs and node_modules)", "vendor", false},
		{"build is NOT excluded", "build", false},
		{"ordinary directory", "lib", false},
		{"directory that merely contains a dot", "my.dir", false},
		{"empty name", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isExcludedDirName(tc.dir); got != tc.want {
				t.Errorf("isExcludedDirName(%q) = %v, want %v", tc.dir, got, tc.want)
			}
		})
	}
}

func TestWalkProject(t *testing.T) {
	root := newFixture(t, map[string]string{
		"main.go":            "package main",
		"README.md":          "# hi",
		"lib/a.go":           "package lib",
		"lib/b.py":           "print(1)",
		"lib/deep/c.ts":      "export {}",
		".git/config":        "[core]",
		".hidden/secret.go":  "package hidden",
		"node_modules/pkg/i": "module.exports = {}",
	})
	res, err := walkProject(root)
	if err != nil {
		t.Fatalf("walkProject: %v", err)
	}

	wantFiles := []string{"README.md", "lib/a.go", "lib/b.py", "lib/deep/c.ts", "main.go"}
	if got := sortedCopy(relPaths(res.files)); !reflect.DeepEqual(got, wantFiles) {
		t.Errorf("walked files\n got: %v\nwant: %v", got, wantFiles)
	}

	gotDirs := map[string]int{}
	for _, d := range res.dirs {
		gotDirs[d.relPath] = d.fileCount
	}
	wantDirs := map[string]int{"lib": 2, "lib/deep": 1}
	if !reflect.DeepEqual(gotDirs, wantDirs) {
		t.Errorf("directory file counts\n got: %v\nwant: %v", gotDirs, wantDirs)
	}

	// relative paths are always slash separated, and absPath is usable.
	for _, f := range res.files {
		if f.relPath == "" {
			t.Errorf("empty relPath in %+v", f)
		}
		if f.absPath == "" {
			t.Errorf("empty absPath in %+v", f)
		}
		if f.size == 0 {
			t.Errorf("expected a size for %q", f.relPath)
		}
		if f.modUnix <= 0 {
			t.Errorf("expected a unix mtime for %q, got %d", f.relPath, f.modUnix)
		}
	}
	if res.newestUnix <= 0 {
		t.Errorf("newestUnix should be populated, got %d", res.newestUnix)
	}
}

func TestWalkProjectEmptyDirectory(t *testing.T) {
	root := t.TempDir()
	res, err := walkProject(root)
	if err != nil {
		t.Fatalf("walkProject on empty dir: %v", err)
	}
	if len(res.files) != 0 || len(res.dirs) != 0 {
		t.Errorf("expected an empty result, got %+v", res)
	}
	if res.newestUnix != 0 {
		t.Errorf("newestUnix = %d, want 0", res.newestUnix)
	}
}

// TestWalkProjectReportsAMissingRoot: the walk callback used to swallow the root
// stat error and return an empty result, which made analyze.go's
// `walk project: %w` branch unreachable. "This directory has no files" and
// "this directory does not exist" are different answers, and only one of them
// is worth reporting as a success.
func TestWalkProjectReportsAMissingRoot(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	res, err := walkProject(missing)
	if err == nil {
		t.Fatalf("walkProject(missing) = %+v, want an error", res)
	}
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("walkProject(missing) error = %v, want one wrapping fs.ErrNotExist", err)
	}
}

func TestWalkProjectCountsOnlyDirectChildren(t *testing.T) {
	root := newFixture(t, map[string]string{
		"a.txt":       "1",
		"b.txt":       "2",
		"sub/c.txt":   "3",
		"sub/d/e.txt": "4",
	})
	res, err := walkProject(root)
	if err != nil {
		t.Fatalf("walkProject: %v", err)
	}
	counts := map[string]int{}
	for _, d := range res.dirs {
		counts[d.relPath] = d.fileCount
	}
	want := map[string]int{"sub": 1, "sub/d": 1}
	if !reflect.DeepEqual(counts, want) {
		t.Errorf("direct-children counts\n got: %v\nwant: %v", counts, want)
	}
	// The root's own file count is not exposed as a dirInfo entry.
	if _, ok := counts["."]; ok {
		t.Errorf("walk result should not contain a root dir entry: %v", counts)
	}
}

func TestWalkProjectVisitsLexicallyAndStably(t *testing.T) {
	files := map[string]string{}
	for _, n := range []string{"z.go", "a.go", "m/b.go", "m/a.go", "B.go"} {
		files[n] = "x"
	}
	root := newFixture(t, files)
	first := walkedPaths(t, root)
	for i := 0; i < 3; i++ {
		if got := walkedPaths(t, root); !reflect.DeepEqual(got, first) {
			t.Fatalf("walk order is not stable:\n%v\nvs\n%v", got, first)
		}
	}
	want := []string{"B.go", "a.go", "m/a.go", "m/b.go", "z.go"}
	if !reflect.DeepEqual(first, want) {
		t.Errorf("walk order\n got: %v\nwant: %v", first, want)
	}
}

// walkedPaths returns the relative paths produced by a walk of root, in visit
// order.
func walkedPaths(t *testing.T, root string) []string {
	t.Helper()
	res, err := walkProject(root)
	if err != nil {
		t.Fatalf("walkProject: %v", err)
	}
	return relPaths(res.files)
}
