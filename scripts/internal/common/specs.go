package common

import (
	"fmt"
	"os"
	"path/filepath"
)

// A spec folder is named after its git branch: .plaesy/specs/<branch>/. A branch
// name is unique, so a spec name is unique too, and two branches cannot share a spec.
var defaultBranches = map[string]bool{"main": true, "master": true, "HEAD": true, "": true}

// FindSpecName returns the spec folder for branch: the folder named after it. It
// errors when that folder does not exist.
func FindSpecName(repoRoot, branch string) (string, error) {
	specsDir := filepath.Join(repoRoot, ".plaesy", "specs")
	if _, err := os.ReadDir(specsDir); err != nil {
		return "", fmt.Errorf("no specs directory: %w", err)
	}
	info, err := os.Stat(filepath.Join(specsDir, branch))
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("no spec for branch %q under %s", branch, specsDir)
	}
	return branch, nil
}

// MainCheckoutRoot returns the root of the main working tree, even when run
// from a linked worktree. Files that git ignores (templates, instructions)
// exist only in the main checkout, so worktrees read them from here.
func MainCheckoutRoot() (string, error) {
	out, err := gitOutput("rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", err
	}
	// The common dir is <main>/.git; its parent is the main working tree.
	return filepath.Dir(toNativePath(out)), nil
}

// IsDefaultBranch reports whether a branch is a trunk branch rather than a
// feature. Feature commands refuse to run on these.
func IsDefaultBranch(branch string) bool {
	return defaultBranches[branch]
}
