package graph

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// normalizePath must keep the root it was given. It is a lexical realpath, so
// on Windows it has to return "C:/repo/docs/a.md" and not "/C:/repo/docs/a.md":
// every caller compares the result against `repoRootSlash + "/"`, and a path
// that grew a leading slash matches nothing. The result was that markdown links,
// dot-sources, JS/Python/Dart imports and Java imports resolved to "" on Windows
// and the graph came out with no structural edges at all — a silent difference
// between the three platforms CI builds, in the one command whose whole job is
// to find references.
func TestNormalizePathKeepsTheRootItWasGiven(t *testing.T) {
	tests := []struct {
		name, base, p, want string
	}{
		{"relative", "/repo", "docs/a.md", "/repo/docs/a.md"},
		{"dot segment", "/repo", "./docs/a.md", "/repo/docs/a.md"},
		{"parent segment", "/repo/docs", "../a.md", "/repo/a.md"},
		{"parent above the root clamps at the root", "/repo", "../../a.md", "/a.md"},
		{"absolute p ignores base", "/repo", "/other/a.md", "/other/a.md"},
		{"empty p is the base", "/repo", "", "/repo"},
		{"trailing slash", "/repo", "docs/", "/repo/docs"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizePath(tt.base, tt.p); got != tt.want {
				t.Errorf("normalizePath(%q, %q) = %q, want %q", tt.base, tt.p, got, tt.want)
			}
		})
	}
}

func TestNormalizePathPreservesWindowsDrive(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("a drive letter only exists on Windows")
	}
	tests := []struct{ name, base, p, want string }{
		{"drive preserved", `C:/repo`, "docs/a.md", `C:/repo/docs/a.md`},
		{"parent keeps the drive", `C:/repo/docs`, "../a.md", `C:/repo/a.md`},
		{"above the drive is dropped", `C:/repo`, "../../../a.md", `C:/a.md`},
		{"absolute p wins", `C:/repo`, `D:/other/a.md`, `D:/other/a.md`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizePath(tt.base, tt.p); got != tt.want {
				t.Errorf("normalizePath(%q, %q) = %q, want %q", tt.base, tt.p, got, tt.want)
			}
		})
	}
}

// The platform-independent statement of the same invariant: whatever the root
// is, the resolved path still starts with it. This is the form that fails on
// Windows and passes on Linux, which is why it is the one worth keeping.
func TestNormalizePathResultStaysUnderTheRoot(t *testing.T) {
	root := toSlash(t.TempDir())
	if root == "" {
		t.Skip("no temp dir")
	}
	for _, p := range []string{"docs/a.md", "./b.md", "c/d/e.md"} {
		got := normalizePath(root, p)
		if !strings.HasPrefix(got, root+"/") {
			t.Errorf("normalizePath(%q, %q) = %q, which is not under the root %q", root, p, got, root)
		}
	}
}

// resolveRelLink is the first caller to notice, so it gets the end-to-end test:
// a link in a document must resolve to the file it names, on every platform.
func TestResolveRelLinkFindsTheTargetFile(t *testing.T) {
	root := writeRepoTree(t, map[string]string{
		"README.md":    "See [spec](docs/spec.md).\n",
		"docs/spec.md": "Back to [readme](../README.md).\n",
	})
	known := map[string]bool{"README.md": true, "docs/spec.md": true}
	tests := []struct{ name, from, link, want string }{
		{"down a folder", "README.md", "docs/spec.md", "docs/spec.md"},
		{"up a folder", "docs/spec.md", "../README.md", "README.md"},
		{"anchor is stripped", "README.md", "docs/spec.md#usage", "docs/spec.md"},
		{"external link", "README.md", "https://example.com", ""},
		{"anchor only", "README.md", "#usage", ""},
		{"unknown target", "README.md", "docs/nope.md", ""},
		{"escapes the repo", "README.md", "../outside.md", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveRelLink(root, tt.from, tt.link, known)
			if got != tt.want {
				t.Errorf("resolveRelLink(%q, %q) = %q, want %q (root %q)", tt.from, tt.link, got, tt.want, filepath.Base(root))
			}
		})
	}
}
