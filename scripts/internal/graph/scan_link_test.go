package graph

import (
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestResolveRelLink(t *testing.T) {
	known := map[string]bool{
		"docs/b.md":     true,
		"docs/sub/c.md": true,
		"top.md":        true,
		"README.md":     true,
	}
	const root = "/repo"

	tests := []struct {
		name    string
		fromRel string
		link    string
		want    string
	}{
		{name: "sibling resolves against from dir", fromRel: "docs/a.md", link: "b.md", want: "docs/b.md"},
		{name: "nested relative", fromRel: "docs/a.md", link: "sub/c.md", want: "docs/sub/c.md"},
		{name: "dot-slash prefix", fromRel: "docs/a.md", link: "./b.md", want: "docs/b.md"},
		{name: "parent traversal within repo", fromRel: "docs/sub/a.md", link: "../b.md", want: "docs/b.md"},
		{name: "root-relative file from nested dir", fromRel: "docs/sub/a.md", link: "../../top.md", want: "top.md"},
		{name: "fromRel without slash uses repo root", fromRel: "top.md", link: "README.md", want: "README.md"},
		{name: "anchor is stripped", fromRel: "docs/a.md", link: "b.md#sec-2", want: "docs/b.md"},
		{name: "anchor-only link ignored", fromRel: "docs/a.md", link: "#sec", want: ""},
		{name: "empty link ignored", fromRel: "docs/a.md", link: "", want: ""},
		{name: "http ignored", fromRel: "docs/a.md", link: "http://x/b.md", want: ""},
		{name: "https ignored", fromRel: "docs/a.md", link: "https://x/b.md#frag", want: ""},
		{name: "unresolved target returns empty", fromRel: "docs/a.md", link: "missing.md", want: ""},
		{name: "escaping the repo returns empty", fromRel: "docs/a.md", link: "../../../etc/passwd", want: ""},
		{name: "anchor-only after slash path", fromRel: "top.md", link: "#a#b", want: ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveRelLink(root, tc.fromRel, tc.link, known); got != tc.want {
				t.Fatalf("resolveRelLink(%q, %q) = %q, want %q", tc.fromRel, tc.link, got, tc.want)
			}
		})
	}

	// A link that resolves to the repo root is a directory, not a node. The
	// resolver used to let that case bypass the prefix check, and since
	// TrimPrefix cannot strip a missing trailing slash it returned an absolute
	// path: never a key in `known` (which holds repo-relative ids), and one that
	// would embed a machine-specific path in project.graph.json if it ever were.
	t.Run("link resolving to the repo root is not a node", func(t *testing.T) {
		if got := resolveRelLink(root, "docs/a.md", "..", map[string]bool{root: true}); got != "" {
			t.Fatalf("resolveRelLink(root-level link) = %q, want \"\"", got)
		}
	})
}

func TestReadAll(t *testing.T) {
	t.Run("existing file returns content", func(t *testing.T) {
		p := filepath.Join(t.TempDir(), "f.txt")
		if err := os.WriteFile(p, []byte("hello\nworld"), 0o644); err != nil {
			t.Fatal(err)
		}
		if got := readAll(p); got != "hello\nworld" {
			t.Fatalf("readAll() = %q", got)
		}
	})

	t.Run("missing file returns empty string", func(t *testing.T) {
		if got := readAll(filepath.Join(t.TempDir(), "nope.txt")); got != "" {
			t.Fatalf("readAll() = %q, want empty", got)
		}
	})

	t.Run("directory returns empty string", func(t *testing.T) {
		if got := readAll(t.TempDir()); got != "" {
			t.Fatalf("readAll(dir) = %q, want empty", got)
		}
	})
}

func TestJoinRoot(t *testing.T) {
	tests := []struct {
		name string
		root string
		rel  string
	}{
		{name: "nested rel", root: "/repo", rel: "docs/a.md"},
		{name: "root itself", root: "/repo", rel: ""},
		{name: "dot-slash rel is cleaned", root: "/repo", rel: "./docs/a.md"},
		{name: "parent in rel is cleaned", root: "/repo", rel: "docs/../a.md"},
		{name: "windows-style root", root: `C:/repo`, rel: "docs/a.md"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			want := filepath.FromSlash(path.Join(tc.root, tc.rel))
			if got := joinRoot(tc.root, tc.rel); got != want {
				t.Fatalf("joinRoot(%q, %q) = %q, want %q", tc.root, tc.rel, got, want)
			}
		})
	}
}
