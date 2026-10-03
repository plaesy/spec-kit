package trimmer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// goCommand builds a portable subprocess invocation: the `go` binary is
// guaranteed to be on PATH for a test that is running under it, and unlike
// `sh -c` it behaves the same on every platform this project builds on.
func goCommand(args ...string) []string {
	return append([]string{"go"}, args...)
}

func readStats(t *testing.T, path string) []StatRecord {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read stats: %v", err)
	}
	var records []StatRecord
	if err := json.Unmarshal(data, &records); err != nil {
		t.Fatalf("parse stats: %v\n%s", err, data)
	}
	return records
}

func readDirEntries(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

// writeFile creates path (and its parents) with content.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
