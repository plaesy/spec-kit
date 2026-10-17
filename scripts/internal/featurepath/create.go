// Package featurepath implements the feature-branch workflow helpers ported
// from create-new-feature.sh, get-feature-paths.sh and
// check-task-prerequisites.sh.
package featurepath

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/plaesy/spec-kit/internal/common"
	"github.com/plaesy/spec-kit/internal/taskmanage"
)

// CreateResult describes a new feature.
type CreateResult struct {
	// BranchName is the git branch, e.g. "auth-login".
	BranchName string
	// SpecName is the spec folder, {TIMESTAMP}-{BRANCH}, e.g.
	// "20261015-143022-auth-login".
	SpecName string
	// SpecFile is the spec.md inside the spec folder.
	SpecFile string
	// WorktreePath is the linked worktree the feature lives in. It is empty
	// when the feature was created in the current working tree.
	WorktreePath string
	// Warning is set when the new spec would not be committed, for example
	// because git ignores it. It is empty when there is nothing to report.
	Warning string
}

// CreateOptions controls where a feature is created.
type CreateOptions struct {
	// Worktree creates the feature in its own git worktree outside the
	// repository, so several features can be in progress at once. Without it
	// the feature branch is checked out in the current working tree.
	Worktree bool
	// Now supplies the timestamp. Nil means the current time. Tests set it.
	Now func() time.Time
}

var nonAlnumRE = regexp.MustCompile(`[^a-z0-9]+`)

// CreateNewFeature creates a feature in the current working tree.
func CreateNewFeature(description string) (*CreateResult, error) {
	return CreateNewFeatureWith(description, CreateOptions{})
}

// CreateNewFeatureWith creates a feature. Its branch is named after the first
// words of the description, and its spec folder is {TIMESTAMP}-{BRANCH}.
func CreateNewFeatureWith(description string, opts CreateOptions) (*CreateResult, error) {
	description = strings.TrimSpace(description)
	if description == "" {
		return nil, fmt.Errorf("feature description is required")
	}

	slug := branchSlug(description)
	if slug == "" {
		return nil, fmt.Errorf("feature description %q has no letters or digits to name the branch", description)
	}

	repoRoot, err := common.GetRepoRoot()
	if err != nil {
		return nil, err
	}
	// Templates are git-ignored, so only the main checkout has them. A
	// worktree reads them from there.
	mainRoot, err := common.MainCheckoutRoot()
	if err != nil {
		return nil, err
	}

	now := time.Now
	if opts.Now != nil {
		now = opts.Now
	}
	specName := common.NewSpecName(now(), slug)

	if opts.Worktree {
		return createInWorktree(mainRoot, slug, specName)
	}
	return createInPlace(repoRoot, mainRoot, slug, specName)
}

// createInPlace checks the feature branch out in the current working tree.
func createInPlace(repoRoot, mainRoot, branch, specName string) (*CreateResult, error) {
	specsDir := filepath.Join(repoRoot, ".plaesy", "specs")
	if err := os.MkdirAll(specsDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating specs dir: %w", err)
	}

	if out, err := exec.Command("git", "checkout", "-b", branch).CombinedOutput(); err != nil {
		return nil, fmt.Errorf("git checkout -b %s: %w: %s", branch, err, string(out))
	}

	specDir := filepath.Join(specsDir, specName)
	return seedSpec(mainRoot, repoRoot, specDir, branch, specName, "")
}

// createInWorktree adds a linked worktree for the feature branch, outside the
// repository, and creates the spec inside it. The branch starts from the
// main checkout's current commit.
func createInWorktree(mainRoot, branch, specName string) (*CreateResult, error) {
	projectDir, err := ProjectWorktreeDir(mainRoot)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating worktree dir: %w", err)
	}

	wtPath := filepath.Join(projectDir, specName)
	if _, err := os.Stat(wtPath); err == nil {
		return nil, fmt.Errorf("worktree path already exists: %s", wtPath)
	}

	if out, err := exec.Command("git", "-C", mainRoot, "worktree", "add", "-b", branch, wtPath).CombinedOutput(); err != nil {
		return nil, fmt.Errorf("git worktree add -b %s %s: %w: %s", branch, wtPath, err, string(out))
	}

	specDir := filepath.Join(wtPath, ".plaesy", "specs", specName)
	result, err := seedSpec(mainRoot, wtPath, specDir, branch, specName, wtPath)
	if err != nil {
		// Undo only what this call created: the new worktree and its branch.
		_ = exec.Command("git", "-C", mainRoot, "worktree", "remove", "--force", wtPath).Run()
		_ = exec.Command("git", "-C", mainRoot, "branch", "-D", branch).Run()
		return nil, err
	}
	return result, nil
}

// seedSpec creates the spec folder with its status directories and spec.md.
// templateRoot is where the template is read from; gitDir is the working tree
// the spec belongs to, used to warn when git ignores the new spec.
func seedSpec(templateRoot, gitDir, specDir, branch, specName, worktreePath string) (*CreateResult, error) {
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating feature dir: %w", err)
	}
	// Tasks belong to the spec, so the spec starts with its own status folders.
	if err := taskmanage.CreateStatusDirs(filepath.Join(specDir, "tasks")); err != nil {
		return nil, fmt.Errorf("creating task dirs: %w", err)
	}
	specFile := filepath.Join(specDir, "spec.md")
	if err := seedSpecFromTemplate(templateRoot, specFile); err != nil {
		return nil, err
	}
	return &CreateResult{
		BranchName:   branch,
		SpecName:     specName,
		SpecFile:     specFile,
		WorktreePath: worktreePath,
		Warning:      ignoredWarning(gitDir, specFile),
	}, nil
}

// ignoredWarning returns a warning when git ignores the spec file, so the
// spec would never be committed. It returns "" when the file is tracked-able
// or when git cannot tell.
func ignoredWarning(gitDir, specFile string) string {
	cmd := exec.Command("git", "-C", gitDir, "check-ignore", "-q", specFile)
	err := cmd.Run()
	if err == nil {
		return fmt.Sprintf("git ignores %s, so the spec will not be committed. Check your gitignore rules for .plaesy/ (a global ignore file often has one).", specFile)
	}
	return ""
}

// branchSlug turns a description into a branch name: its first three words,
// lower-cased and joined by hyphens.
func branchSlug(description string) string {
	slug := strings.Trim(nonAlnumRE.ReplaceAllString(strings.ToLower(description), "-"), "-")
	var words []string
	for _, w := range strings.Split(slug, "-") {
		if w == "" {
			continue
		}
		words = append(words, w)
		if len(words) == 3 {
			break
		}
	}
	return strings.Join(words, "-")
}

// specTemplateName is the file seeded into .plaesy/specs/<spec>/spec.md.
const specTemplateName = "spec.template.md"

// specTemplateCandidates lists, in order, where the spec template may live
// relative to the repository root.
//
// The installed location comes first: `plaesy init` copies templates into
// <repo>/.plaesy/templates (see scaffold.copyTemplates). The bare
// <repo>/templates path is the spec-kit source tree itself, which is only
// correct when the framework is dogfooding its own repository.
//
// This ordering matters: reading <repo>/templates unconditionally made every
// installed project silently produce a zero-byte spec.md, because the read
// failed and the error was discarded.
func specTemplateCandidates(repoRoot string) []string {
	return []string{
		filepath.Join(repoRoot, ".plaesy", "templates", specTemplateName),
		filepath.Join(repoRoot, "templates", specTemplateName),
	}
}

// seedSpecFromTemplate copies the spec template to specFile.
//
// A missing template is a hard error: silently writing an empty spec.md
// produced artifacts that looked complete but carried no content, and the
// caller had no signal that anything was wrong.
func seedSpecFromTemplate(repoRoot, specFile string) error {
	var tried []string
	for _, template := range specTemplateCandidates(repoRoot) {
		tried = append(tried, template)
		data, err := os.ReadFile(template)
		if err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("reading spec template %s: %w", template, err)
			}
			continue
		}
		if err := os.WriteFile(specFile, data, 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", specFile, err)
		}
		return nil
	}
	return fmt.Errorf(
		"spec template %s not found; looked in:\n  %s\nrun `plaesy init <dir> --ai <platform>` to install the framework before creating a feature",
		specTemplateName, strings.Join(tried, "\n  "),
	)
}
