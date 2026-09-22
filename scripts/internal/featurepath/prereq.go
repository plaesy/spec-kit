package featurepath

import (
	"fmt"
	"os"
	"strings"

	"github.com/plaesy/spec-kit/internal/common"
)

// GetPathsReport mirrors get-feature-paths.sh's best-effort text report
// (never errors hard — always produces informational output, same as the
// bash version's fallback branches).
func GetPathsReport() string {
	var b strings.Builder

	fp, err := common.GetFeaturePaths()
	if err != nil {
		cwd, _ := os.Getwd()
		fmt.Fprintf(&b, "REPO_ROOT: %s\n", cwd)
		fmt.Fprintf(&b, "BRANCH: unknown\n")
		fmt.Fprintf(&b, "FEATURE_DIR: Not available\n")
		fmt.Fprintf(&b, "FEATURE_SPEC: Not available\n")
		fmt.Fprintf(&b, "IMPL_PLAN: Not available\n")
		fmt.Fprintf(&b, "TASKS: Not available\n")
		fmt.Fprintf(&b, "INFO: Unable to determine feature paths\n")
		return b.String()
	}

	if err := common.CheckFeatureBranch(fp.CurrentBranch); err != nil {
		fmt.Fprintf(&b, "REPO_ROOT: %s\n", fp.RepoRoot)
		fmt.Fprintf(&b, "BRANCH: %s\n", fp.CurrentBranch)
		fmt.Fprintf(&b, "FEATURE_DIR: %s\n", fp.FeatureDir)
		fmt.Fprintf(&b, "FEATURE_SPEC: %s\n", fp.FeatureSpec)
		fmt.Fprintf(&b, "IMPL_PLAN: %s\n", fp.ImplPlan)
		fmt.Fprintf(&b, "TASKS: %s\n", fp.Tasks)
		fmt.Fprintf(&b, "INFO: Not on a feature branch (format: XXX-feature-name)\n")
		return b.String()
	}

	fmt.Fprintf(&b, "REPO_ROOT: %s\n", fp.RepoRoot)
	fmt.Fprintf(&b, "BRANCH: %s\n", fp.CurrentBranch)
	fmt.Fprintf(&b, "FEATURE_DIR: %s\n", fp.FeatureDir)
	fmt.Fprintf(&b, "FEATURE_SPEC: %s\n", fp.FeatureSpec)
	fmt.Fprintf(&b, "IMPL_PLAN: %s\n", fp.ImplPlan)
	fmt.Fprintf(&b, "TASKS: %s\n", fp.Tasks)
	return b.String()
}

// PrereqResult mirrors check-task-prerequisites.sh's --json output.
type PrereqResult struct {
	FeatureDir     string   `json:"FEATURE_DIR"`
	AvailableDocs  []string `json:"AVAILABLE_DOCS"`
}

// CheckTaskPrerequisites mirrors check-task-prerequisites.sh: validates the
// feature dir and plan.md exist, then reports which optional design docs
// are present.
func CheckTaskPrerequisites() (*PrereqResult, error) {
	fp, err := common.GetFeaturePaths()
	if err != nil {
		return nil, err
	}
	if err := common.CheckFeatureBranch(fp.CurrentBranch); err != nil {
		return nil, err
	}
	if info, err := os.Stat(fp.FeatureDir); err != nil || !info.IsDir() {
		return nil, fmt.Errorf("feature directory not found: %s (run /start first to create the feature structure)", fp.FeatureDir)
	}
	if _, err := os.Stat(fp.ImplPlan); err != nil {
		return nil, fmt.Errorf("plan.md not found in %s (create plan.md from templates/plan.template.md first)", fp.FeatureDir)
	}

	var docs []string
	if _, err := os.Stat(fp.Research); err == nil {
		docs = append(docs, "research.md")
	}
	if _, err := os.Stat(fp.DataModel); err == nil {
		docs = append(docs, "data-model.md")
	}
	if entries, err := os.ReadDir(fp.ContractsDir); err == nil && len(entries) > 0 {
		docs = append(docs, "contracts/")
	}
	if _, err := os.Stat(fp.Quickstart); err == nil {
		docs = append(docs, "quickstart.md")
	}

	return &PrereqResult{FeatureDir: fp.FeatureDir, AvailableDocs: docs}, nil
}

// TextReport renders the non-JSON report format of check-task-prerequisites.sh.
func (r *PrereqResult) TextReport(fp *common.FeaturePaths) string {
	var b strings.Builder
	fmt.Fprintf(&b, "FEATURE_DIR:%s\n", r.FeatureDir)
	fmt.Fprintf(&b, "AVAILABLE_DOCS:\n")
	fmt.Fprintln(&b, common.CheckFile(fp.Research, "research.md"))
	fmt.Fprintln(&b, common.CheckFile(fp.DataModel, "data-model.md"))
	fmt.Fprintln(&b, common.CheckDir(fp.ContractsDir, "contracts/"))
	fmt.Fprintln(&b, common.CheckFile(fp.Quickstart, "quickstart.md"))
	return b.String()
}
