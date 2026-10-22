package exclude

import (
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestDirName(t *testing.T) {
	cases := []struct {
		name string
		dir  string
		want bool
	}{
		{"dot directory", ".git", true},
		{"tool cache", ".dart_tool", true},
		{"just a dot", ".", true},
		{"node_modules", "node_modules", true},
		{"vendor", "vendor", true},
		{"dist", "dist", true},
		{"build", "build", true},
		{"target", "target", true},
		{"pycache", "__pycache__", true},
		{"venv", "venv", true},
		{"Pods", "Pods", true},
		{"DerivedData", "DerivedData", true},
		{"coverage", "coverage", true},
		{"ordinary source directory", "lib", false},
		{"directory that merely contains a dot", "my.dir", false},
		{"bin is not excluded (often a source folder)", "bin", false},
		{"empty name", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := DirName(tc.dir); got != tc.want {
				t.Errorf("DirName(%q) = %v, want %v", tc.dir, got, tc.want)
			}
		})
	}
}

func TestRelPath(t *testing.T) {
	cases := []struct {
		rel  string
		want bool
	}{
		{"ios/Flutter/ephemeral", true},
		{"linux/flutter/ephemeral", true},
		{"macos/Flutter/ephemeral", true},
		{"flutter/ephemeral", true},
		{"ios/Runner", false},
		{"lib/flutter/ephemeral_helper", false},
			{"scripts/internal/assets/data", true},
			{"scripts/internal/assets/data/prompts/save.md", true},
			{"scripts/internal/assets/gen", false},
			{"prompts/save.md", false},
	}
	for _, tc := range cases {
		if got := RelPath(tc.rel); got != tc.want {
			t.Errorf("RelPath(%q) = %v, want %v", tc.rel, got, tc.want)
		}
	}
}

func TestGitignoreMatch(t *testing.T) {
	rules := []string{
		"# a comment",
		"",
		"*.g.dart",
		"build/",
		"/out",
		"docs/*.tmp",
		"secrets/**",
		"**/cache",
		"!keep.g.dart",
	}
	ig := &Ignore{}
	for _, l := range rules {
		if r, ok := parseRule(l); ok {
			ig.rules = append(ig.rules, r)
		}
	}
	cases := []struct {
		name  string
		rel   string
		isDir bool
		want  bool
	}{
		{"generated dart at any depth", "lib/src/model.g.dart", false, true},
		{"negated re-include", "lib/keep.g.dart", false, false},
		{"dir-only rule matches a directory", "app/build", true, true},
		{"dir-only rule ignores a file of the same name", "app/build", false, false},
		{"anchored rule matches only at root", "out", true, true},
		{"anchored rule does not match nested", "x/out", true, false},
		{"inner slash anchors the pattern", "docs/a.tmp", false, true},
		{"inner slash does not match deeper", "docs/deep/a.tmp", false, false},
		{"** suffix matches inside", "secrets/keys/id.txt", false, true},
		{"** prefix matches at any depth", "a/b/cache", true, true},
		{"unrelated file", "lib/main.dart", false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ig.Match(tc.rel, tc.isDir); got != tc.want {
				t.Errorf("Match(%q, dir=%v) = %v, want %v", tc.rel, tc.isDir, got, tc.want)
			}
		})
	}
}

func TestNilIgnoreMatchesNothing(t *testing.T) {
	var ig *Ignore
	if ig.Match("anything.g.dart", false) || ig.SkipDir("x", "x") || ig.SkipFile("x") {
		t.Fatal("a nil Ignore must not exclude anything beyond the built-in names")
	}
	if !ig.SkipDir("x/node_modules", "node_modules") {
		t.Fatal("a nil Ignore must still apply the built-in directory names")
	}
}

func TestLoadGitignore(t *testing.T) {
	t.Run("missing file excludes nothing extra", func(t *testing.T) {
		ig := LoadGitignore(t.TempDir())
		if ig.Match("model.g.dart", false) {
			t.Fatal("no .gitignore must mean no extra rules")
		}
	})
	t.Run("reads CRLF line endings", func(t *testing.T) {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte("*.g.dart\r\nbuild/\r\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		ig := LoadGitignore(root)
		if !ig.Match("lib/model.g.dart", false) {
			t.Error("*.g.dart must match with CRLF endings")
		}
		if !ig.SkipDir("app/build", "build") {
			t.Error("build/ must skip a directory with CRLF endings")
		}
	})
}

func TestSkipIndexDirAndFile(t *testing.T) {
	var ig *Ignore
	dirs := []struct {
		rel  string
		want bool
	}{
		{".plaesy", false},
		{".plaesy/memory", false},
		{".plaesy/memory/roles", false},
		{".plaesy/decisions", false},
		{".plaesy/analysis", true},
		{".plaesy/tasks", true},
		{".git", true},
		{"docs", false},
		{"scripts/internal/assets/data", true},
	}
	for _, tc := range dirs {
		if got := ig.SkipIndexDir(tc.rel, path.Base(tc.rel)); got != tc.want {
			t.Errorf("SkipIndexDir(%q) = %v, want %v", tc.rel, got, tc.want)
		}
	}
	files := []struct {
		rel  string
		want bool
	}{
		{".plaesy/memory/topic.md", false},
		{".plaesy/decisions/hms-not-adopted.md", false},
		{".plaesy/context.md", true},
		{".plaesy/analysis/docs/own.md", true},
		{"README.md", false},
	}
	for _, tc := range files {
		if got := ig.SkipIndexFile(tc.rel); got != tc.want {
			t.Errorf("SkipIndexFile(%q) = %v, want %v", tc.rel, got, tc.want)
		}
	}
}

// TestSkipIndexFileWithRepoGitignore uses this repository's own .gitignore, whose
// `.plaesy/` rule must not hide the memory and decision records from the index.
func TestSkipIndexFileWithRepoGitignore(t *testing.T) {
	ig := LoadGitignore("../../..")
	for _, rel := range []string{".plaesy/memory/session-2026-10-17.md", ".plaesy/decisions/hms-not-adopted.md"} {
		if ig.SkipIndexFile(rel) {
			t.Errorf("SkipIndexFile(%q) = true with the repo .gitignore, want false", rel)
		}
	}
}
