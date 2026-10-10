package semantic

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/plaesy/spec-kit/internal/graph"
)

func TestStoredFingerprint_NoFile_ReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	fp, err := StoredFingerprint(dir)
	if err != nil {
		t.Fatalf("StoredFingerprint error: %v", err)
	}
	if fp != "" {
		t.Errorf("expected empty fingerprint, got %q", fp)
	}
}

func TestRecordAndStoredFingerprint_RoundTrip(t *testing.T) {
	repoRoot := t.TempDir()
	// Create a couple of files so SourceFingerprint has something to hash
	writeFile(t, filepath.Join(repoRoot, "a.go"), "package main\nfunc A() {}\n")
	writeFile(t, filepath.Join(repoRoot, "b.go"), "package main\nfunc B() {}\n")

	outDir := t.TempDir()
	opts := graph.Options{RepoPath: repoRoot, OutDir: ".plaesy/analysis"}
	paths, err := graph.ResolvePaths(opts)
	if err != nil {
		t.Fatalf("ResolvePaths: %v", err)
	}
	// Override OutFull to our temp dir
	paths.OutFull = outDir

	if err := RecordFingerprint(opts, paths); err != nil {
		t.Fatalf("RecordFingerprint: %v", err)
	}

	fp, err := StoredFingerprint(outDir)
	if err != nil {
		t.Fatalf("StoredFingerprint: %v", err)
	}
	if fp == "" {
		t.Fatal("expected non-empty fingerprint after RecordFingerprint")
	}
	// Format should be "file_count|newest_mtime"
	if len(fp) < 3 || fp[1] != '|' {
		t.Errorf("fingerprint format unexpected: %q", fp)
	}
}

func TestIsIndexStale_NoPriorIndex_ReturnsTrue(t *testing.T) {
	repoRoot := t.TempDir()
	writeFile(t, filepath.Join(repoRoot, "a.go"), "package main\nfunc A() {}\n")

	outDir := t.TempDir()
	opts := graph.Options{RepoPath: repoRoot, OutDir: ".plaesy/analysis"}
	paths, err := graph.ResolvePaths(opts)
	if err != nil {
		t.Fatalf("ResolvePaths: %v", err)
	}
	paths.OutFull = outDir

	stale, err := IsIndexStale(opts, paths, outDir)
	if err != nil {
		t.Fatalf("IsIndexStale error: %v", err)
	}
	if !stale {
		t.Error("expected stale=true when no prior index exists")
	}
}

func TestIsIndexStale_FreshIndex_ReturnsFalse(t *testing.T) {
	repoRoot := t.TempDir()
	writeFile(t, filepath.Join(repoRoot, "a.go"), "package main\nfunc A() {}\n")

	outDir := t.TempDir()
	opts := graph.Options{RepoPath: repoRoot, OutDir: ".plaesy/analysis"}
	paths, err := graph.ResolvePaths(opts)
	if err != nil {
		t.Fatalf("ResolvePaths: %v", err)
	}
	paths.OutFull = outDir

	// Record initial fingerprint
	if err := RecordFingerprint(opts, paths); err != nil {
		t.Fatalf("RecordFingerprint: %v", err)
	}

	// Immediately check - should be fresh
	stale, err := IsIndexStale(opts, paths, outDir)
	if err != nil {
		t.Fatalf("IsIndexStale error: %v", err)
	}
	if stale {
		t.Error("expected stale=false for fresh index")
	}
}

func TestIsIndexStale_FileContentChange_DetectedViaMtime(t *testing.T) {
	repoRoot := t.TempDir()
	aPath := filepath.Join(repoRoot, "a.go")
	writeFile(t, aPath, "package main\nfunc A() {}\n")

	outDir := t.TempDir()
	opts := graph.Options{RepoPath: repoRoot, OutDir: ".plaesy/analysis"}
	paths, err := graph.ResolvePaths(opts)
	if err != nil {
		t.Fatalf("ResolvePaths: %v", err)
	}
	paths.OutFull = outDir

	// Record initial fingerprint
	if err := RecordFingerprint(opts, paths); err != nil {
		t.Fatalf("RecordFingerprint: %v", err)
	}

	// Modify the file and force a strictly newer mtime. Relying on the write
	// alone to advance mtime is flaky: filesystem timestamp granularity differs
	// across platforms and mounts (1s on some Windows/network volumes), so a fast
	// write can land inside the same second and the fingerprint compares equal,
	// failing the test for a reason that has nothing to do with the code.
	writeFile(t, aPath, "package main\nfunc A() {}\nfunc B() {}\n")
	newer := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(aPath, newer, newer); err != nil {
		t.Fatalf("Chtimes: %v", err)
	}

	stale, err := IsIndexStale(opts, paths, outDir)
	if err != nil {
		t.Fatalf("IsIndexStale error: %v", err)
	}
	if !stale {
		t.Error("expected stale=true after file modification (mtime changed)")
	}
}

func TestFingerprintPath_Format(t *testing.T) {
	got := FingerprintPath(filepath.Join("tmp", "out"))
	want := filepath.Join("tmp", "out", "embeddings", fingerprintFileName)
	if got != want {
		t.Errorf("FingerprintPath = %q, want %q", got, want)
	}
}

// writeFile is defined once in semantic_test.go; this file reuses it rather than
// redeclaring it (a duplicate declaration in the same package is a build error).

// TestBuildIndexThenRecord_IndexIsFresh covers the real `--index` sequence from
// cmd/plaesy/search.go: BuildIndex writes the collection, the caller then records
// the source fingerprint, and a query immediately afterwards is not stale.
//
// BuildIndex deliberately does NOT record a fingerprint itself — the caller does,
// after the build succeeds, so a failed build never leaves a fingerprint claiming
// the index is fresh. A test that expects BuildIndex alone to produce one is
// asserting a contract that does not exist.
func TestBuildIndexThenRecord_IndexIsFresh(t *testing.T) {
	outDir := t.TempDir()
	repoRoot := t.TempDir()
	writeFile(t, filepath.Join(repoRoot, "auth", "validate.go"), `package auth

// ValidateUser checks credentials.
func ValidateUser(user, pass string) bool { return true }
`)

	opts := graph.Options{RepoPath: repoRoot, OutDir: ".plaesy/analysis"}
	paths, err := graph.ResolvePaths(opts)
	if err != nil {
		t.Fatalf("ResolvePaths: %v", err)
	}
	paths.OutFull = outDir

	if err := BuildIndex(context.Background(), testGraph(), outDir, repoRoot, &fakeEmbedder{}); err != nil {
		t.Fatalf("BuildIndex: %v", err)
	}

	fp, err := StoredFingerprint(outDir)
	if err != nil {
		t.Fatalf("StoredFingerprint: %v", err)
	}
	if fp != "" {
		t.Errorf("BuildIndex recorded a fingerprint itself (%q); only the caller should", fp)
	}

	if err := RecordFingerprint(opts, paths); err != nil {
		t.Fatalf("RecordFingerprint: %v", err)
	}

	stale, err := IsIndexStale(opts, paths, outDir)
	if err != nil {
		t.Fatalf("IsIndexStale: %v", err)
	}
	if stale {
		t.Error("index should be fresh immediately after build + record")
	}
}
