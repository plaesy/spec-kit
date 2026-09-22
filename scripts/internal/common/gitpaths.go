package common

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var featureBranchRE = regexp.MustCompile(`^[0-9]{3}-`)

// GetRepoRoot mirrors get_repo_root: git rev-parse --show-toplevel.
func GetRepoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// GetCurrentBranch mirrors get_current_branch: git rev-parse --abbrev-ref HEAD.
func GetCurrentBranch() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("unable to determine current branch: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
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
	return filepath.Join(repoRoot, "specs", branch)
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
