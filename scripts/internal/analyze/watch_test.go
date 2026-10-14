package analyze

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/plaesy/spec-kit/internal/graph"
)

// watchFingerprintFor resolves the paths the way Watch does and returns the
// combined fingerprint for dir.
func watchFingerprintFor(t *testing.T, dir string) string {
	t.Helper()
	gopts := graph.Options{RepoPath: dir}
	paths, err := graph.ResolvePaths(gopts)
	if err != nil {
		t.Fatal(err)
	}
	fp, err := watchFingerprint(gopts, paths, dir)
	if err != nil {
		t.Fatal(err)
	}
	return fp
}

func writeAt(t *testing.T, path, body string, mtime time.Time) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, mtime, mtime); err != nil {
		t.Fatal(err)
	}
}

// A change to a document-index file that the graph does not collect (JSON,
// HTML) must change the watch fingerprint, or watch leaves that index stale.
func TestWatchFingerprintSeesDocumentIndexFiles(t *testing.T) {
	dir := t.TempDir()
	base := time.Now().Add(-time.Hour)
	writeAt(t, filepath.Join(dir, "src", "main.go"), "package main\n", base)
	writeAt(t, filepath.Join(dir, "docs", "guide.html"), "<p>old</p>", base)

	before := watchFingerprintFor(t, dir)

	writeAt(t, filepath.Join(dir, "docs", "guide.html"), "<p>new</p>", time.Now())
	after := watchFingerprintFor(t, dir)

	if before == after {
		t.Fatalf("editing docs/guide.html did not change the watch fingerprint: %s", after)
	}
}

// Ordinary source edits still change it, as they did before the document half
// was added.
func TestWatchFingerprintSeesSourceFiles(t *testing.T) {
	dir := t.TempDir()
	base := time.Now().Add(-time.Hour)
	writeAt(t, filepath.Join(dir, "src", "main.go"), "package main\n", base)

	before := watchFingerprintFor(t, dir)

	writeAt(t, filepath.Join(dir, "src", "main.go"), "package main\n// x\n", time.Now())
	after := watchFingerprintFor(t, dir)

	if before == after {
		t.Fatalf("editing src/main.go did not change the watch fingerprint: %s", after)
	}
}
