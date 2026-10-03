package graph

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func writeScanFS(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, content := range files {
		full := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("MkdirAll(%q): %v", filepath.Dir(full), err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatalf("WriteFile(%q): %v", full, err)
		}
	}
	return root
}

func TestCollectFiles(t *testing.T) {
	root := writeScanFS(t, map[string]string{
		"b.md":                  "",
		"a.md":                  "",
		"sub/c.go":              "",
		"sub/deep/d.py":         "",
		"notes.txt":             "",
		".git/config.md":        "",
		"node_modules/e.md":     "",
		"dist/f.md":             "",
		"build/g.md":            "",
		"__pycache__/h.md":      "",
		"vendor/i.md":           "",
		"graphify-out/j.json":   "",
		"graphify-out/sub/k.md": "",
	})

	t.Run("only candidate extensions, sorted, no excluded dirs", func(t *testing.T) {
		got, err := collectFiles(root, "")
		if err != nil {
			t.Fatalf("collectFiles() error = %v", err)
		}
		want := []string{"a.md", "b.md", "graphify-out/sub/k.md", "sub/c.go", "sub/deep/d.py"}
		sort.Strings(want)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("collectFiles() = %v, want %v", got, want)
		}
	})

	t.Run("outDirNorm subtree is skipped", func(t *testing.T) {
		got, err := collectFiles(root, "graphify-out")
		if err != nil {
			t.Fatalf("collectFiles() error = %v", err)
		}
		want := []string{"a.md", "b.md", "sub/c.go", "sub/deep/d.py"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("collectFiles() = %v, want %v", got, want)
		}
	})

	t.Run("sibling dir with outDir as prefix is kept", func(t *testing.T) {
		got, err := collectFiles(root, "graph")
		if err != nil {
			t.Fatalf("collectFiles() error = %v", err)
		}
		want := []string{"a.md", "b.md", "graphify-out/sub/k.md", "sub/c.go", "sub/deep/d.py"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("collectFiles(outDirNorm=graph) = %v, want %v", got, want)
		}
	})

	t.Run("empty tree yields no files and no error", func(t *testing.T) {
		got, err := collectFiles(t.TempDir(), "")
		if err != nil {
			t.Fatalf("collectFiles() error = %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("collectFiles() = %v, want empty", got)
		}
	})

	t.Run("missing root is tolerated (best-effort scan)", func(t *testing.T) {
		got, err := collectFiles(filepath.Join(t.TempDir(), "nope"), "")
		if err != nil {
			t.Fatalf("collectFiles() error = %v, want nil", err)
		}
		if len(got) != 0 {
			t.Fatalf("collectFiles() = %v, want empty", got)
		}
	})
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name string
		base string
		p    string
		want string
	}{
		{name: "relative joins onto base", base: "/repo", p: "a/b.md", want: "/repo/a/b.md"},
		{name: "empty p is base itself", base: "/repo", p: "", want: "/repo"},
		{name: "dot segments dropped", base: "/repo", p: "./a/./b", want: "/repo/a/b"},
		{name: "trailing slash dropped", base: "/repo", p: "a/b/", want: "/repo/a/b"},
		{name: "parent pops previous segment", base: "/repo", p: "a/../b", want: "/repo/b"},
		{name: "parent above base is dropped", base: "/repo", p: "../x", want: "/x"},
		{name: "repeated parents at root", base: "/repo/sub", p: "../../../x", want: "/x"},
		{name: "absolute p ignores base", base: "/repo", p: "/other/x", want: "/other/x"},
		{name: "empty base with relative p", base: "", p: "a", want: "/a"},
		{name: "both empty", base: "", p: "", want: "/"},
		{name: "duplicate slashes collapse", base: "/repo", p: "a//b", want: "/repo/a/b"},
		{name: "windows drive treated as absolute", base: "/repo", p: "C:/x/y", want: "C:/x/y"},
		{name: "windows root is preserved", base: "C:/repo", p: "docs/a.md", want: "C:/repo/docs/a.md"},
		{name: "windows parent keeps the drive", base: "C:/repo/docs", p: "../a.md", want: "C:/repo/a.md"},
		{name: "climbing above a drive stops at the drive", base: "C:/repo", p: "../../../a.md", want: "C:/a.md"},
		{name: "absolute drive p wins over base", base: "C:/repo", p: "D:/other/a.md", want: "D:/other/a.md"},
		{name: "bare drive root", base: "", p: "C:/", want: "C:/"},
		{name: "parent of sole segment yields filesystem root", base: "/repo", p: "..", want: "/"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizePath(tc.base, tc.p); got != tc.want {
				t.Fatalf("normalizePath(%q, %q) = %q, want %q", tc.base, tc.p, got, tc.want)
			}
		})
	}
}

func TestIsWindowsAbs(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"C:", true}, {"c:/x", true}, {"Z:\\y", true}, {"C", false}, {"/x", false}, {"", false}, {"1:", false}, {"C:", true},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			if got := isWindowsAbs(tc.in); got != tc.want {
				t.Fatalf("isWindowsAbs(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}
