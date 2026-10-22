package graph

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
)

// collectFiles is documented as a best-effort scan: an entry it cannot read is
// skipped rather than aborting the walk. The only way to make a directory
// unreadable on POSIX is its mode, and Windows has no equivalent that a test
// can set up without changing the machine's ACLs — so this branch is asserted
// where it can be reached and skipped where it cannot.
func TestCollectFilesSkipsAnUnreadableDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("a directory's mode does not stop a Windows directory listing")
	}
	root := newBuildRoot(t)
	writeFiles(t, root, map[string]string{
		"a.md":                 "# A\n",
		"locked/secret.md":     "# Secret\n",
		"locked/deep/other.md": "# Other\n",
	})
	if err := os.Chmod(filepath.Join(root, "locked"), 0o000); err != nil {
		t.Skipf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(root, "locked"), 0o755) })

	got, err := collectFiles(root, "")
	if err != nil {
		t.Fatalf("collectFiles() error = %v; an unreadable directory must not abort the scan", err)
	}
	sort.Strings(got)
	if len(got) != 1 || got[0] != "a.md" {
		t.Fatalf("collectFiles() = %v, want just the readable file", got)
	}
}

// A walk whose root cannot be read is the same contract one level up: nothing
// found, no error. The root here is a path that does not exist, which is what
// a stale --path or a deleted working directory looks like.
func TestCollectFilesToleratesAMissingRoot(t *testing.T) {
	got, err := collectFiles(filepath.Join(t.TempDir(), "gone", "deeper"), "")
	if err != nil {
		t.Fatalf("collectFiles() error = %v, want nil", err)
	}
	if len(got) != 0 {
		t.Fatalf("collectFiles() = %v, want empty", got)
	}
}
