package common

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// A spec folder is named {TIMESTAMP}-{BRANCH} under .plaesy/specs/, where
// TIMESTAMP is the UTC creation time as YYYYMMDD-HHMMSS and BRANCH is the git
// branch the spec lives on. Folders created before timestamps existed are
// named after the branch alone (NNN-slug), so both forms must resolve.
const SpecTimestampLayout = "20060102-150405"

var (
	specStampRE     = regexp.MustCompile(`^[0-9]{8}-[0-9]{6}-`)
	defaultBranches = map[string]bool{"main": true, "master": true, "HEAD": true, "": true}
)

// NewSpecName returns {TIMESTAMP}-{BRANCH} for a branch, using the given time
// in UTC.
func NewSpecName(now time.Time, branch string) string {
	return now.UTC().Format(SpecTimestampLayout) + "-" + branch
}

// SpecBranch returns the branch a spec folder belongs to: the folder name with
// its timestamp prefix removed. A legacy folder with no timestamp is its own
// branch.
func SpecBranch(folder string) string {
	if loc := specStampRE.FindStringIndex(folder); loc != nil {
		return folder[loc[1]:]
	}
	return folder
}

// FindSpecName returns the name of the spec folder that belongs to branch,
// searching repoRoot/.plaesy/specs. It errors when no folder matches, or when
// more than one does (two specs on one branch cannot be told apart).
func FindSpecName(repoRoot, branch string) (string, error) {
	entries, err := os.ReadDir(filepath.Join(repoRoot, ".plaesy", "specs"))
	if err != nil {
		return "", fmt.Errorf("no specs directory: %w", err)
	}
	var matches []string
	for _, e := range entries {
		if e.IsDir() && SpecBranch(e.Name()) == branch {
			matches = append(matches, e.Name())
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no spec for branch %q under %s", branch, filepath.Join(repoRoot, ".plaesy", "specs"))
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("branch %q matches several specs: %s", branch, strings.Join(matches, ", "))
	}
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
