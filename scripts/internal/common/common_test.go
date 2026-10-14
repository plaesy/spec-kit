package common

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// chdir switches the process working directory for the duration of the test.
// GetRepoRoot and GetCurrentBranch shell out to git with no -C flag, so the cwd
// is the only seam (go.mod declares go 1.20, which has no t.Chdir). Callers must
// not use t.Parallel.
func chdir(t *testing.T, dir string) {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(prev); err != nil {
			t.Errorf("restore cwd to %s: %v", prev, err)
		}
	})
}

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skipf("git not available: %v", err)
	}
}

// initRepo creates a git repository in a fresh temp dir, on a branch named
// after the test's needs, and returns its path. Kept shallow: git and Windows
// both have path-length limits, and t.TempDir() may already be short here.
func initRepo(t *testing.T, branch string) string {
	t.Helper()
	requireGit(t)
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init", "-q"},
		{"config", "user.email", "test@example.invalid"},
		{"config", "user.name", "Test"},
		{"checkout", "-q", "-b", branch},
		{"commit", "-q", "--allow-empty", "-m", "init"},
	} {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	return dir
}

// initRepoWithoutCommit creates a repository that has a branch but no revision
// on it. This is not a broken fixture: `plaesy features create` leaves exactly
// this state, and it is the state the very next command runs against.
func initRepoWithoutCommit(t *testing.T, branch string) string {
	t.Helper()
	requireGit(t)
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init", "-q"},
		{"config", "user.email", "test@example.invalid"},
		{"config", "user.name", "Test"},
		{"checkout", "-q", "-b", branch},
	} {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	return dir
}

// `git rev-parse --abbrev-ref HEAD` fails on an unborn branch: HEAD names a
// branch that has no revision to resolve. `plaesy features create` makes
// exactly such a repository, so the command immediately after it could not read
// the branch back. `git branch --show-current` answers the same question there,
// and is what GetCurrentBranch falls back to.
func TestGetCurrentBranchOnAnUnbornBranch(t *testing.T) {
	dir := initRepoWithoutCommit(t, "001-alpha")
	chdir(t, dir)

	branch, err := GetCurrentBranch()
	if err != nil {
		t.Fatalf("GetCurrentBranch on a branch with no commit: %v", err)
	}
	if branch != "001-alpha" {
		t.Errorf("GetCurrentBranch = %q, want 001-alpha", branch)
	}
}

// The fallback is not a second code path with its own idea of the answer: on a
// repository that has commits, both queries name the same branch.
func TestGetCurrentBranchAgreesOnACommittedBranch(t *testing.T) {
	dir := initRepo(t, "002-beta")
	chdir(t, dir)

	branch, err := GetCurrentBranch()
	if err != nil {
		t.Fatalf("GetCurrentBranch: %v", err)
	}
	if branch != "002-beta" {
		t.Errorf("GetCurrentBranch = %q, want 002-beta", branch)
	}
}

// Outside a repository both queries fail, and the error has to say so rather
// than reporting an empty branch name.
func TestGetCurrentBranchOutsideARepository(t *testing.T) {
	requireGit(t)
	chdir(t, t.TempDir())

	branch, err := GetCurrentBranch()
	if err == nil {
		t.Fatalf("GetCurrentBranch outside a repo = %q, want an error", branch)
	}
	if !strings.Contains(err.Error(), "unable to determine current branch") {
		t.Errorf("error %q does not explain the cause", err)
	}
}

func TestNormalizeVersion(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"1.2.3", "1.2.3"},
		{"  1.2.3  ", "1.2.3"},
		{"0.0.0", "0.0.0"},
		{"10.20.30", "10.20.30"},
		{"", "0.0.0"},
		{"1.2", "0.0.0"},
		{"1.2.3-rc1", "0.0.0"},
		{"v1.2.3", "0.0.0"},
		{"1.2.3\n", "1.2.3"},
		{"1.2.3.4", "0.0.0"},
		{"latest", "0.0.0"},
	}
	for _, tc := range cases {
		if got := NormalizeVersion(tc.in); got != tc.want {
			t.Errorf("NormalizeVersion(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestVersionDefaultIsNotSemver(t *testing.T) {
	// Version is injected at build time; the fallback must itself be valid, or
	// `plaesy --version` reports something NormalizeVersion would reject.
	if got := NormalizeVersion(Version); got != Version {
		t.Errorf("NormalizeVersion(Version=%q) = %q, want it unchanged", Version, got)
	}
}

func TestCheckFeatureBranch(t *testing.T) {
	cases := []struct {
		branch  string
		wantErr bool
	}{
		{"001-feature-name", false},
		{"042-add-oauth", false},
		{"999-", false},
		{"001", true},
		{"01-short", true},
		{"1-no-pad", true},
		{"main", true},
		{"", true},
		{"x001-name", true},
		{"001x-name", true},
	}
	for _, tc := range cases {
		err := CheckFeatureBranch(tc.branch)
		if tc.wantErr != (err != nil) {
			t.Errorf("CheckFeatureBranch(%q) error = %v, wantErr %v", tc.branch, err, tc.wantErr)
		}
		if err != nil && !strings.Contains(err.Error(), tc.branch) {
			t.Errorf("CheckFeatureBranch(%q) error %q does not name the branch", tc.branch, err)
		}
	}
}

func TestGetFeatureDir(t *testing.T) {
	got := GetFeatureDir(filepath.Join("repo", "root"), "001-x")
	want := filepath.Join("repo", "root", ".plaesy", "specs", "001-x")
	if got != want {
		t.Errorf("GetFeatureDir = %q, want %q", got, want)
	}
}

func TestShellLinesQuotesValuesForTheShell(t *testing.T) {
	p := &FeaturePaths{
		RepoRoot:      "/home/me/repo",
		CurrentBranch: "001-x",
		FeatureDir:    "/home/me/repo/.plaesy/specs/001-x",
		FeatureSpec:   "/home/me/repo/.plaesy/specs/001-x/spec.md",
		ImplPlan:      "/home/me/repo/.plaesy/specs/001-x/plan.md",
		Tasks:         "/home/me/repo/.plaesy/specs/001-x/tasks.md",
		Research:      "/home/me/repo/.plaesy/specs/001-x/research.md",
		DataModel:     "/home/me/repo/.plaesy/specs/001-x/data-model.md",
		Quickstart:    "/home/me/repo/.plaesy/specs/001-x/quickstart.md",
		ContractsDir:  "/home/me/repo/.plaesy/specs/001-x/contracts",
	}
	out := p.ShellLines()
	for _, key := range []string{
		"REPO_ROOT", "CURRENT_BRANCH", "FEATURE_DIR", "FEATURE_SPEC", "IMPL_PLAN",
		"TASKS", "RESEARCH", "DATA_MODEL", "QUICKSTART", "CONTRACTS_DIR",
	} {
		if !strings.Contains(out, key+"=") {
			t.Errorf("ShellLines missing %s:\n%s", key, out)
		}
	}
	if !strings.Contains(out, `REPO_ROOT='/home/me/repo'`) {
		t.Errorf("values are not single-quoted:\n%s", out)
	}
	if strings.Count(out, "\n") != 10 {
		t.Errorf("ShellLines has %d lines, want 10:\n%s", strings.Count(out, "\n"), out)
	}

	// A single quote in a value must not be able to end the quoting.
	p.RepoRoot = `/home/o'brien/repo`
	quoted := p.ShellLines()
	if !strings.Contains(quoted, `'/home/o'\''brien/repo'`) {
		t.Errorf("embedded quote not escaped POSIX-style:\n%s", quoted)
	}
}

func TestCheckFileAndCheckDir(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "present.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	full := filepath.Join(root, "full")
	empty := filepath.Join(root, "empty")
	for _, d := range []string{full, empty} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(full, "a"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}

	if got := CheckFile(file, "config"); got != "  ✓ config" {
		t.Errorf("CheckFile(present) = %q", got)
	}
	if got := CheckFile(filepath.Join(root, "missing"), "config"); got != "  ✗ config" {
		t.Errorf("CheckFile(missing) = %q", got)
	}
	if got := CheckDir(full, "dir"); got != "  ✓ dir" {
		t.Errorf("CheckDir(non-empty) = %q", got)
	}
	// An empty directory is a ✗: the report is about content being there.
	if got := CheckDir(empty, "dir"); got != "  ✗ dir" {
		t.Errorf("CheckDir(empty) = %q", got)
	}
	if got := CheckDir(filepath.Join(root, "missing"), "dir"); got != "  ✗ dir" {
		t.Errorf("CheckDir(missing) = %q", got)
	}
	// A file is not a directory, even though it exists.
	if got := CheckDir(file, "dir"); got != "  ✗ dir" {
		t.Errorf("CheckDir(file) = %q", got)
	}
}

func TestValidateFileAndDirectoryExists(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "a.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := ValidateFileExists(file, ""); err != nil {
		t.Errorf("ValidateFileExists(present) = %v", err)
	}
	if err := ValidateFileExists(root, "the config"); err == nil ||
		!strings.Contains(err.Error(), "required the config not found") {
		t.Errorf("a directory is not a file: err = %v", err)
	}
	err := ValidateFileExists(filepath.Join(root, "nope"), "")
	if err == nil || !strings.Contains(err.Error(), "required file not found") {
		t.Errorf("empty description should default to \"file\": err = %v", err)
	}

	if err := ValidateDirectoryExists(root, ""); err != nil {
		t.Errorf("ValidateDirectoryExists(present) = %v", err)
	}
	if err := ValidateDirectoryExists(file, "the dir"); err == nil ||
		!strings.Contains(err.Error(), "required the dir not found") {
		t.Errorf("a file is not a directory: err = %v", err)
	}
	err = ValidateDirectoryExists(filepath.Join(root, "nope"), "")
	if err == nil || !strings.Contains(err.Error(), "required directory not found") {
		t.Errorf("empty description should default to \"directory\": err = %v", err)
	}
}

func TestValidateCommandExists(t *testing.T) {
	requireGit(t)
	if err := ValidateCommandExists("git", ""); err != nil {
		t.Errorf("ValidateCommandExists(git) = %v", err)
	}
	if err := ValidateCommandExists("git", "the VCS"); err != nil {
		t.Errorf("ValidateCommandExists(git, description) = %v", err)
	}
	err := ValidateCommandExists("plaesy-no-such-command-xyz", "")
	if err == nil || !strings.Contains(err.Error(), "required command not found") {
		t.Errorf("missing command: err = %v", err)
	}
	err = ValidateCommandExists("plaesy-no-such-command-xyz", "the tool")
	if err == nil || !strings.Contains(err.Error(), "required the tool not found") {
		t.Errorf("named command: err = %v", err)
	}
}

func TestValidateNotRoot(t *testing.T) {
	err := ValidateNotRoot()
	if os.Geteuid() == 0 {
		// Containers and CI runners are commonly root; the check must catch it.
		if err == nil {
			t.Error("running as root should be refused")
		}
		return
	}
	if err != nil {
		t.Errorf("ValidateNotRoot as a normal user = %v, want nil", err)
	}
}

func TestValidateEnvironment(t *testing.T) {
	requireGit(t)
	// Outside a repository it warns rather than failing: the command may still
	// be doing something useful in a plain directory.
	plain := t.TempDir()
	chdir(t, plain)
	if err := ValidateEnvironment(); err != nil {
		t.Errorf("ValidateEnvironment outside a repo = %v, want nil", err)
	}

	// Inside one it passes too; the difference is the warning, not the result.
	chdir(t, initRepo(t, "001-env"))
	if err := ValidateEnvironment(); err != nil {
		t.Errorf("ValidateEnvironment inside a repo = %v, want nil", err)
	}
}

func TestGetRepoRootAndBranch(t *testing.T) {
	dir := initRepo(t, "001-alpha")
	chdir(t, dir)

	root, err := GetRepoRoot()
	if err != nil {
		t.Fatalf("GetRepoRoot: %v", err)
	}
	// git may report the path with symlinks resolved (macOS /tmp), so compare
	// the resolved forms rather than the raw strings.
	wantRoot, err := filepath.EvalSymlinks(dir)
	if err != nil {
		wantRoot = dir
	}
	gotRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		gotRoot = root
	}
	if gotRoot != wantRoot {
		t.Errorf("GetRepoRoot = %q, want %q", root, dir)
	}

	branch, err := GetCurrentBranch()
	if err != nil {
		t.Fatalf("GetCurrentBranch: %v", err)
	}
	if branch != "001-alpha" {
		t.Errorf("GetCurrentBranch = %q, want 001-alpha", branch)
	}
}

func TestGetRepoRootOutsideARepository(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skipf("git not available: %v", err)
	}
	chdir(t, t.TempDir())
	if root, err := GetRepoRoot(); err == nil {
		t.Errorf("GetRepoRoot outside a repo = %q, want an error", root)
	} else if !strings.Contains(err.Error(), "not a git repository") {
		t.Errorf("error %q does not explain the cause", err)
	}
}

func TestGetFeaturePaths(t *testing.T) {
	dir := initRepo(t, "001-beta")
	chdir(t, dir)

	paths, err := GetFeaturePaths()
	if err != nil {
		t.Fatalf("GetFeaturePaths: %v", err)
	}
	if paths.CurrentBranch != "001-beta" {
		t.Errorf("CurrentBranch = %q", paths.CurrentBranch)
	}
	if !strings.HasSuffix(paths.FeatureDir, filepath.Join(".plaesy", "specs", "001-beta")) {
		t.Errorf("FeatureDir = %q", paths.FeatureDir)
	}
	if filepath.Base(paths.FeatureSpec) != "spec.md" ||
		filepath.Base(paths.ImplPlan) != "plan.md" ||
		filepath.Base(paths.Tasks) != "tasks.md" ||
		filepath.Base(paths.Research) != "research.md" ||
		filepath.Base(paths.DataModel) != "data-model.md" ||
		filepath.Base(paths.Quickstart) != "quickstart.md" ||
		filepath.Base(paths.ContractsDir) != "contracts" {
		t.Errorf("unexpected feature file layout: %+v", paths)
	}
	// Every path must sit inside the feature dir, or a command writing to one
	// writes outside the feature.
	for name, p := range map[string]string{
		"FeatureSpec": paths.FeatureSpec, "ImplPlan": paths.ImplPlan,
		"Tasks": paths.Tasks, "Research": paths.Research,
		"DataModel": paths.DataModel, "Quickstart": paths.Quickstart,
		"ContractsDir": paths.ContractsDir,
	} {
		if !strings.HasPrefix(p, paths.FeatureDir+string(os.PathSeparator)) {
			t.Errorf("%s = %q is outside FeatureDir %q", name, p, paths.FeatureDir)
		}
	}
}
