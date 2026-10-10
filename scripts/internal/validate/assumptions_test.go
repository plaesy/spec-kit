package validate

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// runGit mirrors the exec.Command("git", "-C", dir, ...) pattern used by
// internal/common's own tests (common_test.go) and internal/featurepath's.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestAssumptions_FindsTags(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "plan.md", "Some text.\n\nASSUMED — use the default retry count\n\nMore text.\n")

	res, err := Assumptions(dir, 0)
	if err != nil {
		t.Fatalf("Assumptions: %v", err)
	}
	if res.FilesScanned != 1 {
		t.Fatalf("FilesScanned = %d, want 1", res.FilesScanned)
	}
	if len(res.Tags) != 1 {
		t.Fatalf("len(Tags) = %d, want 1", len(res.Tags))
	}
	if res.Tags[0].Text != "use the default retry count" {
		t.Errorf("Tags[0].Text = %q", res.Tags[0].Text)
	}
	if res.Tags[0].Line != 3 {
		t.Errorf("Tags[0].Line = %d, want 3", res.Tags[0].Line)
	}
	if len(res.Stale) != 0 {
		t.Errorf("Stale = %v, want empty when reviewDays=0", res.Stale)
	}
}

func TestAssumptions_HyphenVariantAccepted(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "plan.md", "ASSUMED - plain hyphen variant\n")

	res, err := Assumptions(dir, 0)
	if err != nil {
		t.Fatalf("Assumptions: %v", err)
	}
	if len(res.Tags) != 1 {
		t.Fatalf("len(Tags) = %d, want 1", len(res.Tags))
	}
}

func TestAssumptions_SkipsGitDir(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, filepath.Join(".git", "COMMIT_EDITMSG.md"), "ASSUMED — should not be scanned\n")

	res, err := Assumptions(dir, 0)
	if err != nil {
		t.Fatalf("Assumptions: %v", err)
	}
	if len(res.Tags) != 0 {
		t.Fatalf("len(Tags) = %d, want 0 (skip .git)", len(res.Tags))
	}
}

func TestAssumptions_StaleWhenReviewWindowExceeded(t *testing.T) {
	dir := t.TempDir()
	runGit(t, dir, "init", "-q")
	write(t, dir, "old.md", "ASSUMED — a decision from long ago\n")
	runGit(t, dir, "add", "old.md")

	old := "2020-01-01T00:00:00"
	cmd := exec.Command("git", "-C", dir, "commit", "-q", "-m", "old decision",
		"--date="+old)
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com",
		"GIT_COMMITTER_DATE="+old,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}

	res, err := Assumptions(dir, 30) // 30-day review window
	if err != nil {
		t.Fatalf("Assumptions: %v", err)
	}
	if len(res.Tags) != 1 {
		t.Fatalf("len(Tags) = %d, want 1", len(res.Tags))
	}
	if len(res.Stale) != 1 {
		t.Fatalf("len(Stale) = %d, want 1 (committed in 2020, window 30 days)", len(res.Stale))
	}
}

func TestAssumptions_NotStaleWithinWindow(t *testing.T) {
	dir := t.TempDir()
	runGit(t, dir, "init", "-q")
	write(t, dir, "fresh.md", "ASSUMED — a decision made just now\n")
	runGit(t, dir, "add", "fresh.md")
	runGit(t, dir, "commit", "-q", "-m", "fresh decision")

	res, err := Assumptions(dir, 30)
	if err != nil {
		t.Fatalf("Assumptions: %v", err)
	}
	if len(res.Stale) != 0 {
		t.Fatalf("len(Stale) = %d, want 0 (just committed)", len(res.Stale))
	}
}

func write(t *testing.T, dir, rel, content string) {
	t.Helper()
	path := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}
