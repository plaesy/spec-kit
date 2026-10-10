package installer

// Install and Uninstall are the two functions that write to and delete from
// the user's filesystem, and both were at 0% coverage. The rest of this
// package's tests are written to skip on Windows, which is what left them
// untested: InstallDir reads LOCALAPPDATA on Windows and $HOME elsewhere, and
// both are ordinary environment variables, so redirectInstallDir can point
// either at a temp directory and the whole pair becomes testable on every
// platform the project builds for.
//
// These are characterization tests. They record what the code does, including
// the property the self-copy guard exists to protect, so that changing the
// implementation cannot quietly break it.

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// redirectInstallDir points InstallDir at a temp directory for the duration of
// the test and returns it. Without this, a test of Install would copy the test
// binary into the real ~/.local/bin or %LOCALAPPDATA%\Plaesy\bin.
func redirectInstallDir(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	if runtime.GOOS == "windows" {
		// InstallDir prefers LOCALAPPDATA and only falls back to the home
		// directory when it is empty, so this is the variable that matters.
		t.Setenv("LOCALAPPDATA", tmp)
	} else {
		t.Setenv("HOME", tmp)
	}
	dir, err := InstallDir()
	if err != nil {
		t.Fatalf("InstallDir: %v", err)
	}
	if inside, err := filepath.Rel(tmp, dir); err != nil || len(inside) >= 2 && inside[:2] == ".." {
		t.Fatalf("InstallDir() = %q, which is not inside the temp dir %q; "+
			"this test would write to the real install location", dir, tmp)
	}
	return dir
}

func writeSource(t *testing.T, dir, content string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}
	src := filepath.Join(dir, "plaesy-source")
	if err := os.WriteFile(src, []byte(content), 0o755); err != nil {
		t.Fatalf("write source: %v", err)
	}
	return src
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func TestInstallFromCopiesAndCreatesTheDirectory(t *testing.T) {
	dir := redirectInstallDir(t)
	src := writeSource(t, t.TempDir(), "binary-v1")

	dst, err := installFrom(src)
	if err != nil {
		t.Fatalf("installFrom: %v", err)
	}
	if dst != filepath.Join(dir, binaryFileName()) {
		t.Errorf("destination = %q, want %q", dst, filepath.Join(dir, binaryFileName()))
	}
	if got := readFile(t, dst); got != "binary-v1" {
		t.Errorf("installed content = %q, want %q", got, "binary-v1")
	}
}

func TestInstallFromOverwritesAnEarlierVersion(t *testing.T) {
	dir := redirectInstallDir(t)
	dst := filepath.Join(dir, binaryFileName())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(dst, []byte("binary-v0-stale"), 0o755); err != nil {
		t.Fatalf("seed destination: %v", err)
	}
	src := writeSource(t, t.TempDir(), "binary-v1")

	if _, err := installFrom(src); err != nil {
		t.Fatalf("installFrom: %v", err)
	}
	if got := readFile(t, dst); got != "binary-v1" {
		t.Errorf("installed content = %q, want the new binary %q; install must "+
			"replace an earlier version rather than append or skip", got, "binary-v1")
	}
}

// The property the self-copy guard exists for. copyFile opens its destination
// with O_TRUNC, so an install whose source is the destination would empty the
// binary it is executing — which on Windows means a binary that no longer
// runs, and on Unix one that fails on the next exec. This asserts the file
// still has its content afterwards, not merely that installFrom returned.
func TestInstallFromDoesNotTruncateWhenSourceIsTheDestination(t *testing.T) {
	dir := redirectInstallDir(t)
	dst := filepath.Join(dir, binaryFileName())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	const original = "binary-already-installed"
	if err := os.WriteFile(dst, []byte(original), 0o755); err != nil {
		t.Fatalf("seed destination: %v", err)
	}

	got, err := installFrom(dst)
	if err != nil {
		t.Fatalf("installFrom: %v", err)
	}
	if got != dst {
		t.Errorf("destination = %q, want %q", got, dst)
	}
	if content := readFile(t, dst); content != original {
		t.Errorf("installed content = %q, want %q: running `plaesy install` "+
			"from the installed copy truncated the binary it was executing",
			content, original)
	}
}

// A relative source and an absolute destination are the same file. The guard
// compares filepath.Clean on both, so "…/bin/plaesy" and "…/bin/./plaesy" must
// still be recognized as one; a guard that missed it would truncate.
func TestInstallFromRecognisesTheDestinationThroughADifferentSpelling(t *testing.T) {
	dir := redirectInstallDir(t)
	dst := filepath.Join(dir, binaryFileName())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	const original = "binary-already-installed"
	if err := os.WriteFile(dst, []byte(original), 0o755); err != nil {
		t.Fatalf("seed destination: %v", err)
	}

	messy := filepath.Join(dir, "bin", "..", binaryFileName())
	if _, err := installFrom(messy); err != nil {
		t.Fatalf("installFrom: %v", err)
	}
	if content := readFile(t, dst); content != original {
		t.Errorf("installed content = %q, want %q: the self-copy guard compares "+
			"cleaned paths, and a different spelling of the destination "+
			"slipped past it", content, original)
	}
}

func TestInstallFromMissingSourceIsAnError(t *testing.T) {
	redirectInstallDir(t)
	missing := filepath.Join(t.TempDir(), "not-here")

	dst, err := installFrom(missing)
	if err == nil {
		t.Fatalf("installFrom(%q) = %q with no error; a source that cannot be "+
			"read must not report a successful install", missing, dst)
	}
	if _, statErr := os.Stat(dst); statErr == nil {
		t.Errorf("a destination file was created at %q despite the failure", dst)
	}
}

func TestInstallFromLeavesNoTemporaryFileBehind(t *testing.T) {
	dir := redirectInstallDir(t)
	src := writeSource(t, t.TempDir(), "binary-v1")

	if _, err := installFrom(src); err != nil {
		t.Fatalf("installFrom: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read install dir: %v", err)
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("install left %q in the install directory; copyFile writes "+
				"to dst+\".tmp\" and renames it, so anything still ending in "+
				".tmp\" is residue in a directory that is on the user's PATH",
				filepath.Join(dir, e.Name()))
		}
	}
}

func TestUninstallRemovesTheInstalledBinary(t *testing.T) {
	dir := redirectInstallDir(t)
	path := filepath.Join(dir, binaryFileName())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("binary-v1"), 0o755); err != nil {
		t.Fatalf("seed: %v", err)
	}

	got, err := Uninstall()
	if err != nil {
		t.Fatalf("Uninstall: %v", err)
	}
	if got != path {
		t.Errorf("reported path = %q, want %q", got, path)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("file still present after Uninstall (stat error: %v)", err)
	}
}

// Uninstall is destructive, so the case that matters most is the one where
// there is nothing installed. It has to be a no-op, not an error: a user
// running uninstall twice should not see a failure the second time.
func TestUninstallWhenNothingIsInstalledIsANoOp(t *testing.T) {
	redirectInstallDir(t)

	path, err := Uninstall()
	if err != nil {
		t.Fatalf("Uninstall with nothing installed returned %v, want nil so that "+
			"running it twice does not report a failure", err)
	}
	if path != filepath.Join(mustInstallDir(t), binaryFileName()) {
		t.Errorf("reported path = %q, want the path that would have been installed", path)
	}
}

func mustInstallDir(t *testing.T) string {
	t.Helper()
	dir, err := InstallDir()
	if err != nil {
		t.Fatalf("InstallDir: %v", err)
	}
	return dir
}

// Characterization: what happens when something that is not a file sits where
// the binary would be. Recorded rather than asserted as desirable, because
// the destructive path here is the question, not the formatting.
func TestUninstallWhenThePathIsADirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("os.Remove on a directory is not permitted the same way on Windows; " +
			"this records Unix behaviour and the difference is the point")
	}
	dir := redirectInstallDir(t)
	path := filepath.Join(dir, binaryFileName())
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if _, err := Uninstall(); err != nil {
		t.Fatalf("Uninstall: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("the directory at %q survived Uninstall (stat error: %v)", path, err)
	}
}

// Install itself resolves its source with os.Executable, which is why its two
// error paths cannot be provoked. The happy path can still be exercised, and it
// is worth doing: installFrom is the part that copies, and a regression in the
// path resolution above it — an EvalSymlinks that is dropped, a destination
// that is computed from the wrong base — would otherwise only show up when a
// user runs the command.
func TestInstallCopiesTheRunningBinaryToTheInstallDir(t *testing.T) {
	dir := redirectInstallDir(t)

	dst, err := Install()
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if dst != filepath.Join(dir, binaryFileName()) {
		t.Errorf("destination = %q, want %q", dst, filepath.Join(dir, binaryFileName()))
	}
	self, err := os.Executable()
	if err != nil {
		t.Skipf("os.Executable unavailable: %v", err)
	}
	selfData, err := os.ReadFile(self)
	if err != nil {
		t.Fatalf("read running binary: %v", err)
	}
	installed, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read installed binary: %v", err)
	}
	if len(installed) != len(selfData) {
		t.Errorf("installed binary is %d bytes, running binary is %d; the "+
			"installed copy is not the running executable", len(installed), len(selfData))
	}
}
