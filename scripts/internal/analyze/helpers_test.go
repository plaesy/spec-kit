package analyze

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// writeTree materializes a fixture tree under root. Keys are slash-separated
// relative paths; a key ending in "/" creates an empty directory. All fixture
// trees live inside t.TempDir() so the repository tree is never touched.
func writeTree(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if strings.HasSuffix(rel, "/") {
			if err := os.MkdirAll(p, 0o755); err != nil {
				t.Fatalf("mkdir %s: %v", rel, err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("mkdir parent of %s: %v", rel, err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
}

// newFixture creates an empty temp dir pre-populated with files and returns it.
func newFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	writeTree(t, root, files)
	return root
}

// readFileString reads a whole file, failing the test when unreadable.
func readFileString(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

// logCapture replaces the package-level Logf/Successf seams with collectors so
// tests can assert on the progress narration without touching stdout. The
// original functions are restored via t.Cleanup, so tests using it must not be
// parallel.
type logCapture struct {
	Lines []string
}

func (c *logCapture) format(tag, format string, a ...interface{}) {
	c.Lines = append(c.Lines, tag+" "+fmt.Sprintf(format, a...))
}

func (c *logCapture) joined() string { return strings.Join(c.Lines, "\n") }

func captureLogs(t *testing.T) *logCapture {
	t.Helper()
	cap := &logCapture{}
	origLog, origSuccess := Logf, Successf
	Logf = func(format string, a ...interface{}) { cap.format("[INFO]", format, a...) }
	Successf = func(format string, a ...interface{}) { cap.format("[SUCCESS]", format, a...) }
	t.Cleanup(func() { Logf, Successf = origLog, origSuccess })
	return cap
}

// sortedCopy returns a sorted copy so assertions do not depend on the order
// filepath.WalkDir happens to visit entries in.
func sortedCopy(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}

// relPaths extracts the relative paths from a walk result.
func relPaths(files []fileInfo) []string {
	out := make([]string, len(files))
	for i, f := range files {
		out[i] = f.relPath
	}
	return out
}

func hasString(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

func mustContain(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Errorf("expected output to contain %q, got:\n%s", want, got)
	}
}

func mustNotContain(t *testing.T, got, unwanted string) {
	t.Helper()
	if strings.Contains(got, unwanted) {
		t.Errorf("expected output NOT to contain %q, got:\n%s", unwanted, got)
	}
}
