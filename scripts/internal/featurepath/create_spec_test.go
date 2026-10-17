package featurepath

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// fixedClock is the timestamp used by tests that check a spec folder name.
func fixedClock() time.Time {
	return time.Date(2026, 10, 15, 14, 30, 22, 0, time.UTC)
}

func TestCreateNewFeatureRejectsDescriptionWithoutSlug(t *testing.T) {
	root := t.TempDir()
	installGitShim(t, root, "main")
	if _, err := CreateNewFeatureWith("***", CreateOptions{Now: fixedClock}); err == nil {
		t.Fatal("a description with no letters or digits created a feature, want an error")
	}
}

func TestCreateNewFeatureInWorktreeUsesTheShim(t *testing.T) {
	root := t.TempDir()
	home := t.TempDir()
	t.Setenv(worktreeHomeEnv, home)
	log := installGitShim(t, root, "main")
	writeFile(t, filepath.Join(root, "templates", "spec.template.md"), defaultSpecTemplate)

	res, err := CreateNewFeatureWith("Add user login", CreateOptions{Worktree: true, Now: fixedClock})
	if err != nil {
		t.Fatalf("CreateNewFeatureWith(worktree): %v", err)
	}

	if res.BranchName != "add-user-login" {
		t.Errorf("branch = %q, want add-user-login", res.BranchName)
	}
	if res.SpecName != "20261015-143022-add-user-login" {
		t.Errorf("spec name = %q, want the timestamped folder", res.SpecName)
	}
	// The worktree lives under the worktree home, not inside the repository.
	if !isInside(res.WorktreePath, home) {
		t.Errorf("worktree %q is not under the worktree home %q", res.WorktreePath, home)
	}
	if isInside(res.WorktreePath, root) {
		t.Errorf("worktree %q is inside the repository %q", res.WorktreePath, root)
	}
	// The spec sits inside the worktree, on the feature branch.
	wantSpec := filepath.Join(res.WorktreePath, ".plaesy", "specs", "20261015-143022-add-user-login", "spec.md")
	if res.SpecFile != wantSpec {
		t.Errorf("spec file = %q, want %q", res.SpecFile, wantSpec)
	}
	if _, err := os.Stat(wantSpec); err != nil {
		t.Errorf("spec.md not created in the worktree: %v", err)
	}
	// The template comes from the main checkout, where git-ignored files live.
	if b, _ := os.ReadFile(wantSpec); string(b) != defaultSpecTemplate {
		t.Errorf("spec.md content = %q, want the main checkout's template", b)
	}

	// git was asked for a worktree on the branch, and never for a checkout in
	// the current tree.
	got := gitBranchLog(t, log)
	if len(got) != 1 || !strings.HasPrefix(got[0], "worktree add add-user-login ") {
		t.Errorf("git calls = %v, want one worktree add for add-user-login", got)
	}
}

func TestWorktreeHomeOverride(t *testing.T) {
	t.Setenv(worktreeHomeEnv, "/custom/place")
	got, err := WorktreeHome()
	if err != nil {
		t.Fatalf("WorktreeHome: %v", err)
	}
	if got != "/custom/place" {
		t.Errorf("WorktreeHome = %q, want the override", got)
	}
}

func TestProjectWorktreeDirIsStablePerRepo(t *testing.T) {
	t.Setenv(worktreeHomeEnv, t.TempDir())
	// Hermetic: a global ignore file on this machine must not decide the result.
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	a, err := ProjectWorktreeDir(filepath.Join("/work", "one", "app"))
	if err != nil {
		t.Fatal(err)
	}
	b, _ := ProjectWorktreeDir(filepath.Join("/work", "one", "app"))
	c, _ := ProjectWorktreeDir(filepath.Join("/other", "app"))
	if a != b {
		t.Errorf("the same repo gave two folders: %q and %q", a, b)
	}
	if a == c {
		t.Errorf("two repos named app share the folder %q", a)
	}
	if !strings.HasPrefix(filepath.Base(a), "app-") {
		t.Errorf("project folder %q does not start with the repo name", filepath.Base(a))
	}
}

func TestParseWorktreeList(t *testing.T) {
	out := "worktree /repo\nHEAD abc\nbranch refs/heads/main\n\n" +
		"worktree /home/wt/20261015-143022-login\nHEAD def\nbranch refs/heads/login\n\n" +
		"worktree /home/wt/detached\nHEAD 123\ndetached\n"
	got := parseWorktreeList(out)
	if len(got) != 3 {
		t.Fatalf("entries = %d, want 3: %+v", len(got), got)
	}
	if got[1].path != filepath.Clean("/home/wt/20261015-143022-login") || got[1].branch != "login" {
		t.Errorf("entry 1 = %+v", got[1])
	}
	if got[2].branch != "" {
		t.Errorf("detached worktree reported branch %q", got[2].branch)
	}
}

// TestWorktreeLifecycleWithGit runs the real git: a feature worktree is made,
// committed, merged, and removed by CleanWorktrees. It skips when git is absent.
func TestWorktreeLifecycleWithGit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	repo := t.TempDir()
	t.Setenv(worktreeHomeEnv, t.TempDir())
	// Hermetic: a global ignore file on this machine must not decide the result.
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)

	runGitIn(t, repo, "init", "-b", "main")
	runGitIn(t, repo, "config", "user.email", "test@example.com")
	runGitIn(t, repo, "config", "user.name", "test")
	writeFile(t, filepath.Join(repo, "README.md"), "x\n")
	runGitIn(t, repo, "add", "README.md")
	runGitIn(t, repo, "commit", "-m", "init")
	writeFile(t, filepath.Join(repo, "templates", "spec.template.md"), defaultSpecTemplate)
	t.Chdir(repo)

	res, err := CreateNewFeatureWith("Add login", CreateOptions{Worktree: true, Now: fixedClock})
	if err != nil {
		t.Fatalf("CreateNewFeatureWith: %v", err)
	}

	// Not merged yet: the feature has its own commit.
	runGitIn(t, res.WorktreePath, "add", "-A")
	runGitIn(t, res.WorktreePath, "commit", "-m", "spec for login")

	st, err := CleanWorktrees(true)
	if err != nil {
		t.Fatalf("CleanWorktrees(apply) before merge: %v", err)
	}
	if len(st) != 1 || st[0].Merged || st[0].Removed {
		t.Fatalf("before merge = %+v, want one unmerged worktree kept", st)
	}
	if _, err := os.Stat(res.WorktreePath); err != nil {
		t.Fatalf("unmerged worktree was removed: %v", err)
	}

	// Merged: a dry run reports it and changes nothing.
	runGitIn(t, repo, "merge", "--no-edit", "add-login")
	st, err = CleanWorktrees(false)
	if err != nil {
		t.Fatalf("CleanWorktrees(dry run): %v", err)
	}
	if len(st) != 1 || !st[0].Merged || st[0].Removed {
		t.Fatalf("dry run = %+v, want one merged worktree not removed", st)
	}
	if _, err := os.Stat(res.WorktreePath); err != nil {
		t.Fatalf("dry run removed the worktree: %v", err)
	}

	// Applying removes it.
	st, err = CleanWorktrees(true)
	if err != nil {
		t.Fatalf("CleanWorktrees(apply): %v", err)
	}
	if len(st) != 1 || !st[0].Removed {
		t.Fatalf("apply = %+v, want the merged worktree removed", st)
	}
	if _, err := os.Stat(res.WorktreePath); !os.IsNotExist(err) {
		t.Errorf("worktree folder still exists after removal: %v", err)
	}
}

// TestCleanKeepsADirtyWorktree shows that removal never forces: a merged
// worktree with uncommitted changes is kept and the refusal is reported.
func TestCleanKeepsADirtyWorktree(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	repo := t.TempDir()
	t.Setenv(worktreeHomeEnv, t.TempDir())
	// Hermetic: a global ignore file on this machine must not decide the result.
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)

	runGitIn(t, repo, "init", "-b", "main")
	runGitIn(t, repo, "config", "user.email", "test@example.com")
	runGitIn(t, repo, "config", "user.name", "test")
	writeFile(t, filepath.Join(repo, "README.md"), "x\n")
	runGitIn(t, repo, "add", "README.md")
	runGitIn(t, repo, "commit", "-m", "init")
	writeFile(t, filepath.Join(repo, "templates", "spec.template.md"), defaultSpecTemplate)
	t.Chdir(repo)

	res, err := CreateNewFeatureWith("Add dirty", CreateOptions{Worktree: true, Now: fixedClock})
	if err != nil {
		t.Fatalf("CreateNewFeatureWith: %v", err)
	}
	// The spec was never committed, so the worktree holds untracked files.
	// Its branch is at main, so it counts as merged, but git must refuse the
	// removal and CleanWorktrees must keep it.
	runGitIn(t, repo, "merge", "--no-edit", "add-dirty")

	st, err := CleanWorktrees(true)
	if err != nil {
		t.Fatalf("CleanWorktrees: %v", err)
	}
	if len(st) != 1 || st[0].Removed {
		t.Fatalf("dirty worktree = %+v, want kept", st)
	}
	if !strings.Contains(st[0].Reason, "refused") {
		t.Errorf("reason = %q, want the refusal reported", st[0].Reason)
	}
	if _, err := os.Stat(res.WorktreePath); err != nil {
		t.Errorf("dirty worktree was removed: %v", err)
	}
}

// runGitIn runs git in dir and fails the test on error.
func runGitIn(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v in %s: %v: %s", args, dir, err, out)
	}
}

// TestCreateWarnsWhenTheSpecIsIgnored: a repository that ignores .plaesy/
// would silently drop the spec from commits, so the create reports it.
func TestCreateWarnsWhenTheSpecIsIgnored(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	repo := t.TempDir()
	t.Setenv(worktreeHomeEnv, t.TempDir())
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)

	runGitIn(t, repo, "init", "-b", "main")
	runGitIn(t, repo, "config", "user.email", "test@example.com")
	runGitIn(t, repo, "config", "user.name", "test")
	writeFile(t, filepath.Join(repo, "README.md"), "x\n")
	runGitIn(t, repo, "add", "README.md")
	runGitIn(t, repo, "commit", "-m", "init")
	writeFile(t, filepath.Join(repo, ".git", "info", "exclude"), ".plaesy/\n")
	writeFile(t, filepath.Join(repo, "templates", "spec.template.md"), defaultSpecTemplate)
	t.Chdir(repo)

	res, err := CreateNewFeatureWith("Add ignored", CreateOptions{Worktree: true, Now: fixedClock})
	if err != nil {
		t.Fatalf("CreateNewFeatureWith: %v", err)
	}
	if !strings.Contains(res.Warning, "will not be committed") {
		t.Errorf("warning = %q, want the not-committed warning", res.Warning)
	}
}

// TestCleanKeepsAWorktreeWithIgnoredSpec: git removes ignored files without
// asking, so a merged worktree holding an ignored spec is kept.
func TestCleanKeepsAWorktreeWithIgnoredSpec(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	repo := t.TempDir()
	t.Setenv(worktreeHomeEnv, t.TempDir())
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)

	runGitIn(t, repo, "init", "-b", "main")
	runGitIn(t, repo, "config", "user.email", "test@example.com")
	runGitIn(t, repo, "config", "user.name", "test")
	writeFile(t, filepath.Join(repo, "README.md"), "x\n")
	runGitIn(t, repo, "add", "README.md")
	runGitIn(t, repo, "commit", "-m", "init")
	writeFile(t, filepath.Join(repo, ".git", "info", "exclude"), ".plaesy/\n")
	writeFile(t, filepath.Join(repo, "templates", "spec.template.md"), defaultSpecTemplate)
	t.Chdir(repo)

	res, err := CreateNewFeatureWith("Add ignored spec", CreateOptions{Worktree: true, Now: fixedClock})
	if err != nil {
		t.Fatalf("CreateNewFeatureWith: %v", err)
	}
	// Merged: the branch is at main. The spec is ignored, so removal must not happen.
	st, err := CleanWorktrees(true)
	if err != nil {
		t.Fatalf("CleanWorktrees: %v", err)
	}
	if len(st) != 1 || st[0].Removed {
		t.Fatalf("worktree with ignored spec = %+v, want kept", st)
	}
	if !strings.Contains(st[0].Reason, "git-ignored") {
		t.Errorf("reason = %q, want the ignored-files reason", st[0].Reason)
	}
	if _, err := os.Stat(res.SpecFile); err != nil {
		t.Errorf("ignored spec was deleted: %v", err)
	}
}
