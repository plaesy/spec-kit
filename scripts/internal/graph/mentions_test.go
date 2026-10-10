package graph

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestMentionEdges(t *testing.T) {
	newRoot := func(t *testing.T, files map[string]string) string {
		t.Helper()
		root := t.TempDir()
		for rel, content := range files {
			full := filepath.Join(root, filepath.FromSlash(rel))
			if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		return filepath.ToSlash(root)
	}

	t.Run("forward-slash mention becomes an INFERRED edge", func(t *testing.T) {
		root := newRoot(t, map[string]string{"a.md": "see scripts/install.ps1 for details\n"})
		got := mentionEdges(root, []string{"a.md"}, []string{"a.md", "scripts/install.ps1"})
		want := []Edge{{Source: "a.md", Target: "scripts/install.ps1", Type: "mentions", Confidence: "INFERRED"}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("mentionEdges() = %v, want %v", got, want)
		}
	})

	t.Run("backslash mention is matched too", func(t *testing.T) {
		root := newRoot(t, map[string]string{"a.md": `open scripts\install.ps1` + "\n"})
		got := mentionEdges(root, []string{"a.md"}, []string{"scripts/install.ps1"})
		if len(got) != 1 || got[0].Target != "scripts/install.ps1" {
			t.Fatalf("mentionEdges() = %v, want one edge to scripts/install.ps1", got)
		}
		if got[0].Confidence != "INFERRED" || got[0].Type != "mentions" {
			t.Fatalf("mentionEdges() = %v, want mentions/INFERRED", got[0])
		}
	})

	t.Run("self mention is skipped", func(t *testing.T) {
		root := newRoot(t, map[string]string{"a.md": "a.md mentions a.md\n"})
		if got := mentionEdges(root, []string{"a.md"}, []string{"a.md"}); got != nil {
			t.Fatalf("mentionEdges() = %v, want nil", got)
		}
	})

	t.Run("no mention yields no edges", func(t *testing.T) {
		root := newRoot(t, map[string]string{"a.md": "nothing here\n"})
		if got := mentionEdges(root, []string{"a.md"}, []string{"b.md"}); got != nil {
			t.Fatalf("mentionEdges() = %v, want nil", got)
		}
	})

	t.Run("empty and unreadable files are skipped", func(t *testing.T) {
		root := newRoot(t, map[string]string{"empty.md": "", "ok.md": "b.md\n"})
		files := []string{"empty.md", "missing.md", "ok.md"}
		got := mentionEdges(root, files, []string{"a.md", "b.md"})
		want := []Edge{{Source: "ok.md", Target: "b.md", Type: "mentions", Confidence: "INFERRED"}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("mentionEdges() = %v, want %v", got, want)
		}
	})

	t.Run("every hit in the whole file is reported", func(t *testing.T) {
		root := newRoot(t, map[string]string{"a.md": "x.md\n... y.md ...\n"})
		got := mentionEdges(root, []string{"a.md"}, []string{"x.md", "y.md", "z.md"})
		if len(got) != 2 || got[0].Target != "x.md" || got[1].Target != "y.md" {
			t.Fatalf("mentionEdges() = %v, want x.md then y.md", got)
		}
	})

	t.Run("edges follow file then candidate order", func(t *testing.T) {
		root := newRoot(t, map[string]string{
			"one.md": "b.md c.md\n",
			"two.md": "b.md\n",
		})
		got := mentionEdges(root, []string{"one.md", "two.md"}, []string{"b.md", "c.md"})
		var pairs [][2]string
		for _, e := range got {
			pairs = append(pairs, [2]string{e.Source, e.Target})
		}
		want := [][2]string{{"one.md", "b.md"}, {"one.md", "c.md"}, {"two.md", "b.md"}}
		if !reflect.DeepEqual(pairs, want) {
			t.Fatalf("mentionEdges() = %v, want %v", pairs, want)
		}
	})

	t.Run("no files yields no edges", func(t *testing.T) {
		root := newRoot(t, nil)
		if got := mentionEdges(root, nil, []string{"a.md"}); got != nil {
			t.Fatalf("mentionEdges() = %v, want nil", got)
		}
	})
}

func TestDropWeakMentions(t *testing.T) {
	tests := []struct {
		name  string
		edges []Edge
		want  []Edge
	}{
		{
			name:  "nil input yields nil output",
			edges: nil,
			want:  nil,
		},
		{
			name:  "no mentions are kept",
			edges: []Edge{{Source: "a", Target: "b", Type: "imports", Confidence: "EXTRACTED"}},
			want:  []Edge{{Source: "a", Target: "b", Type: "imports", Confidence: "EXTRACTED"}},
		},
		{
			name: "mention dropped when strong edge exists for same pair",
			edges: []Edge{
				{Source: "a", Target: "b", Type: "mentions", Confidence: "INFERRED"},
				{Source: "a", Target: "b", Type: "imports", Confidence: "EXTRACTED"},
			},
			want: []Edge{{Source: "a", Target: "b", Type: "imports", Confidence: "EXTRACTED"}},
		},
		{
			name: "strong edge listed first is still detected",
			edges: []Edge{
				{Source: "a", Target: "b", Type: "calls", Confidence: "EXTRACTED"},
				{Source: "a", Target: "b", Type: "mentions", Confidence: "INFERRED"},
			},
			want: []Edge{{Source: "a", Target: "b", Type: "calls", Confidence: "EXTRACTED"}},
		},
		{
			name: "reverse direction is not treated as strong",
			edges: []Edge{
				{Source: "b", Target: "a", Type: "imports", Confidence: "EXTRACTED"},
				{Source: "a", Target: "b", Type: "mentions", Confidence: "INFERRED"},
			},
			want: []Edge{
				{Source: "b", Target: "a", Type: "imports", Confidence: "EXTRACTED"},
				{Source: "a", Target: "b", Type: "mentions", Confidence: "INFERRED"},
			},
		},
		{
			name:  "exact duplicates are deduped",
			edges: []Edge{{Source: "a", Target: "b", Type: "mentions", Confidence: "INFERRED"}, {Source: "a", Target: "b", Type: "mentions", Confidence: "INFERRED"}},
			want:  []Edge{{Source: "a", Target: "b", Type: "mentions", Confidence: "INFERRED"}},
		},
		{
			name: "different confidence on same pair is not a duplicate",
			edges: []Edge{
				{Source: "a", Target: "b", Type: "mentions", Confidence: "INFERRED"},
				{Source: "a", Target: "b", Type: "mentions", Confidence: "EXTRACTED"},
			},
			want: []Edge{
				{Source: "a", Target: "b", Type: "mentions", Confidence: "INFERRED"},
				{Source: "a", Target: "b", Type: "mentions", Confidence: "EXTRACTED"},
			},
		},
		{
			name: "unrelated mentions survive in order",
			edges: []Edge{
				{Source: "a", Target: "b", Type: "mentions", Confidence: "INFERRED"},
				{Source: "a", Target: "c", Type: "mentions", Confidence: "INFERRED"},
			},
			want: []Edge{
				{Source: "a", Target: "b", Type: "mentions", Confidence: "INFERRED"},
				{Source: "a", Target: "c", Type: "mentions", Confidence: "INFERRED"},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := dropWeakMentions(tc.edges)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("dropWeakMentions() = %v, want %v", got, tc.want)
			}
		})
	}
}
