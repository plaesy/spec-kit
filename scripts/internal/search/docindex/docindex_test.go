package docindex

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// writeFile creates path under root with content, making parent directories.
func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestCollect_IncludesDocsAndSkipsExcludedDirs checks that documents are found
// and that .plaesy, .git, and dependency folders are not entered.
func TestCollect_IncludesDocsAndSkipsExcludedDirs(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "README.md", "# Title")
	writeFile(t, root, "docs/guide.html", "<p>guide</p>")
	writeFile(t, root, "data/config.json", "{}")
	writeFile(t, root, "main.go", "package main")
	writeFile(t, root, ".plaesy/analysis/docs/own.md", "generated")
	writeFile(t, root, ".git/notes.md", "git")
	writeFile(t, root, "node_modules/pkg/readme.md", "dep")
	writeFile(t, root, "vendor/lib/doc.md", "vendored")

	got, err := Collect(root)
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	want := []string{"README.md", "data/config.json", "docs/guide.html"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Collect = %v, want %v", got, want)
	}
}

// TestCollect_IndexesPlaesyMemoryOnly checks that the memory and decision
// records under .plaesy are indexed, and the rest of .plaesy is not.
func TestCollect_IndexesPlaesyMemoryOnly(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".plaesy/memory/topic.md", "memory")
	writeFile(t, root, ".plaesy/decisions/hms-not-adopted.md", "decision")
	writeFile(t, root, ".plaesy/context.md", "context")
	writeFile(t, root, ".plaesy/analysis/docs/own.md", "generated")
	writeFile(t, root, ".plaesy/tasks/todo/t.md", "task")

	got, err := Collect(root)
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	want := []string{".plaesy/decisions/hms-not-adopted.md", ".plaesy/memory/topic.md"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Collect = %v, want %v", got, want)
	}
}

// TestFingerprint_ChangesWithEditAddAndDelete: the fingerprint is the claim the
// index matches the documents, so each kind of change must move it.
func TestFingerprint_ChangesWithEditAddAndDelete(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "a.md", "one")
	base, err := Fingerprint(root)
	if err != nil {
		t.Fatal(err)
	}

	writeFile(t, root, "a.md", "two")
	edited, _ := Fingerprint(root)
	if edited == base {
		t.Error("editing a document did not change the fingerprint")
	}

	writeFile(t, root, "b.md", "new")
	added, _ := Fingerprint(root)
	if added == edited {
		t.Error("adding a document did not change the fingerprint")
	}

	if err := os.Remove(filepath.Join(root, "b.md")); err != nil {
		t.Fatal(err)
	}
	deleted, _ := Fingerprint(root)
	if deleted != edited {
		t.Error("deleting the added document did not restore the earlier fingerprint")
	}
}

// TestIsStale_MissingIndexIsStale: with no fingerprint on disk, the index is
// refused, not treated as fresh.
func TestIsStale_MissingIndexIsStale(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "a.md", "text")
	stale, err := IsStale(root, filepath.Join(root, ".plaesy", "analysis"))
	if err != nil {
		t.Fatal(err)
	}
	if !stale {
		t.Error("IsStale = false with no index built; want true")
	}
}

// TestIsStale_EditAfterBuildIsStale: an index whose fingerprint matches the
// documents is fresh, and editing one document makes it stale.
func TestIsStale_EditAfterBuildIsStale(t *testing.T) {
	root := t.TempDir()
	analysisDir := filepath.Join(root, ".plaesy", "analysis")
	writeFile(t, root, "a.md", "original text")

	fp, err := Fingerprint(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(Dir(analysisDir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(Dir(analysisDir), fingerprintName), []byte(fp+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if stale, _ := IsStale(root, analysisDir); stale {
		t.Fatal("freshly recorded index reported stale")
	}
	writeFile(t, root, "a.md", "edited text")
	if stale, _ := IsStale(root, analysisDir); !stale {
		t.Error("index not refused after a document was edited")
	}
}

// TestSnippet_CutsOnRuneBoundary: a passage cut in the middle of a multi-byte
// character would print as invalid UTF-8.
func TestSnippet_CutsOnRuneBoundary(t *testing.T) {
	s := strings.Repeat("é", 50) // 2 bytes per rune
	got := snippet(s, 25)
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("snippet %q has no ellipsis", got)
	}
	body := strings.TrimSuffix(got, "...")
	if !strings.HasPrefix(s, body) {
		t.Errorf("snippet body %q is not a prefix of the input", body)
	}
	for _, r := range body {
		if r != 'é' {
			t.Errorf("snippet contains a broken rune %q", r)
		}
	}
}
