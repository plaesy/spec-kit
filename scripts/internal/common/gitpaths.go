package common

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

var featureBranchRE = regexp.MustCompile(`^[0-9]{3}-`)

// msysDriveRE matches an MSYS2/Git-Bash style drive prefix like "/c/" or "/D/".
var msysDriveRE = regexp.MustCompile(`^/([a-zA-Z])/`)

// toNativePath converts a path returned by git (which on Windows may use
// MSYS2/Git-Bash Unix-style paths such as "/c/Users/...") into a native
// OS path so that filepath.Join and os.Stat work correctly.
func toNativePath(p string) string {
	if runtime.GOOS == "windows" {
		if m := msysDriveRE.FindStringSubmatch(p); len(m) > 1 {
			// /c/Users/... → C:\Users\...
			p = strings.ToUpper(m[1]) + ":\\" + p[3:]
		}
		p = filepath.FromSlash(p)
	}
	return p
}

// GetRepoRoot mirrors get_repo_root: git rev-parse --show-toplevel.
func GetRepoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository: %w", err)
	}
	root := toNativePath(strings.TrimSpace(string(out)))
	if fixed, ok := rootViaCwd(root); ok {
		return fixed, nil
	}
	return root, nil
}

// rootViaCwd repairs a root that names a POSIX mount rather than a drive. Under
// Git Bash on Windows git prints its own view of the path: a repository under
// %TEMP% comes back as "/tmp/repo", which toNativePath turns into "	mpepo" —
// a path the OS reads as "<current drive>:	mpepo", a different directory
// that usually does not exist. Every command that resolves the repo root then
// reads and writes nothing, or the wrong tree, while the path it prints looks
// perfectly plausible.
//
// `git rev-parse --show-cdup` gives the same root relative to the working
// directory, and the working directory is a native path, so joining the two
// reconstructs the real root without guessing a mount table. The candidate is
// only accepted once it exists, so this never invents a path.
func rootViaCwd(root string) (string, bool) {
	if runtime.GOOS != "windows" {
		return "", false
	}
	if _, err := os.Stat(root); err == nil {
		return root, true
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", false
	}
	out, err := exec.Command("git", "rev-parse", "--show-cdup").Output()
	if err != nil {
		return "", false
	}
	cdup := strings.TrimSpace(string(out))
	candidate := wd
	if cdup != "" {
		candidate = filepath.Join(wd, filepath.FromSlash(cdup))
	}
	if _, err := os.Stat(candidate); err != nil {
		return "", false
	}
	return candidate, true
}

// gitOutput runs a git subcommand and returns its stdout. The error carries
// git's own stderr, so "not a git repository" and "unknown revision" are told
// apart by the caller instead of both collapsing into "exit status 128".
func gitOutput(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, msg)
		}
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return string(out), nil
}

// GetCurrentBranch mirrors get_current_branch: git rev-parse --abbrev-ref HEAD.
//
// That form fails on an unborn branch — a repository with no commits yet, where
// `git checkout -b 001-x` has set HEAD to 001-x but there is no revision to
// resolve. `plaesy features create` makes exactly that repository, so the very
// next command that asks for the branch failed. `git branch --show-current`
// answers the same question and works there, so it is the fallback.
//
// Both are run through the same helper as the primary call, so a git that is
// absent or a directory that is not a repository produces the same error either
// way rather than a second, differently-worded one.
func GetCurrentBranch() (string, error) {
	if out, err := gitOutput("rev-parse", "--abbrev-ref", "HEAD"); err == nil {
		return strings.TrimSpace(out), nil
	}
	out, err := gitOutput("branch", "--show-current")
	if err != nil {
		return "", fmt.Errorf("unable to determine current branch: %w", err)
	}
	return strings.TrimSpace(out), nil
}

// CheckFeatureBranch mirrors check_feature_branch: enforces NNN-name naming.
func CheckFeatureBranch(branch string) error {
	if !featureBranchRE.MatchString(branch) {
		return fmt.Errorf("not on a feature branch (current: %s); feature branches should be named like: 001-feature-name", branch)
	}
	return nil
}

// GetFeatureDir mirrors get_feature_dir.
func GetFeatureDir(repoRoot, branch string) string {
	return filepath.Join(repoRoot, ".plaesy", "specs", branch)
}

// FeaturePaths mirrors the KEY=value set emitted by get_feature_paths().
type FeaturePaths struct {
	RepoRoot      string
	CurrentBranch string
	FeatureDir    string
	FeatureSpec   string
	ImplPlan      string
	Tasks         string
	Research      string
	DataModel     string
	Quickstart    string
	ContractsDir  string
}

// GetFeaturePaths mirrors get_feature_paths().
func GetFeaturePaths() (*FeaturePaths, error) {
	repoRoot, err := GetRepoRoot()
	if err != nil {
		return nil, err
	}
	branch, err := GetCurrentBranch()
	if err != nil {
		return nil, err
	}
	featureDir := GetFeatureDir(repoRoot, branch)

	return &FeaturePaths{
		RepoRoot:      repoRoot,
		CurrentBranch: branch,
		FeatureDir:    featureDir,
		FeatureSpec:   filepath.Join(featureDir, "spec.md"),
		ImplPlan:      filepath.Join(featureDir, "plan.md"),
		Tasks:         filepath.Join(featureDir, "tasks.md"),
		Research:      filepath.Join(featureDir, "research.md"),
		DataModel:     filepath.Join(featureDir, "data-model.md"),
		Quickstart:    filepath.Join(featureDir, "quickstart.md"),
		ContractsDir:  filepath.Join(featureDir, "contracts"),
	}, nil
}

// ShellLines renders the paths as shell-sourceable KEY=value lines, matching
// the bash get_feature_paths() output format 1:1 for backward compatibility
// with anything that sources/evals this output. Go does not need printf %q
// escaping to protect its own eval (there is none), but callers piping this
// into a shell still need safe quoting, so values are single-quoted with
// embedded quotes escaped POSIX-style.
func (p *FeaturePaths) ShellLines() string {
	q := func(s string) string {
		return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "REPO_ROOT=%s\n", q(p.RepoRoot))
	fmt.Fprintf(&b, "CURRENT_BRANCH=%s\n", q(p.CurrentBranch))
	fmt.Fprintf(&b, "FEATURE_DIR=%s\n", q(p.FeatureDir))
	fmt.Fprintf(&b, "FEATURE_SPEC=%s\n", q(p.FeatureSpec))
	fmt.Fprintf(&b, "IMPL_PLAN=%s\n", q(p.ImplPlan))
	fmt.Fprintf(&b, "TASKS=%s\n", q(p.Tasks))
	fmt.Fprintf(&b, "RESEARCH=%s\n", q(p.Research))
	fmt.Fprintf(&b, "DATA_MODEL=%s\n", q(p.DataModel))
	fmt.Fprintf(&b, "QUICKSTART=%s\n", q(p.Quickstart))
	fmt.Fprintf(&b, "CONTRACTS_DIR=%s\n", q(p.ContractsDir))
	return b.String()
}
