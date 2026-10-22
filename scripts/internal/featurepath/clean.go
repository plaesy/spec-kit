package featurepath

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/plaesy/spec-kit/internal/common"
)

// WorktreeStatus is what CleanWorktrees found and did for one feature worktree.
type WorktreeStatus struct {
	Path   string
	Branch string
	// Merged reports whether the branch is already in the trunk branch.
	Merged bool
	// Removed reports whether the worktree was removed. It is only set when
	// apply is true.
	Removed bool
	// Reason explains a skip or a failure. It is empty for a clean removal.
	Reason string
}

// CleanWorktrees lists the feature worktrees of this project and, when apply
// is true, removes those whose branch is merged into the trunk branch.
//
// Removal never forces. `git worktree remove` refuses a worktree with
// uncommitted changes, and that refusal is reported, not overridden. A dry
// run (apply false) changes nothing.
func CleanWorktrees(apply bool) ([]WorktreeStatus, error) {
	mainRoot, err := common.MainCheckoutRoot()
	if err != nil {
		return nil, err
	}
	projectDir, err := ProjectWorktreeDir(mainRoot)
	if err != nil {
		return nil, err
	}
	trunk, err := trunkBranch(mainRoot)
	if err != nil {
		return nil, err
	}

	out, err := gitIn(mainRoot, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}

	var statuses []WorktreeStatus
	for _, wt := range parseWorktreeList(out) {
		if !isInside(wt.path, projectDir) {
			continue
		}
		st := WorktreeStatus{Path: wt.path, Branch: wt.branch}
		switch {
		case wt.branch == "":
			st.Reason = "detached HEAD; remove by hand if it is no longer needed"
		default:
			st.Merged = gitOK(mainRoot, "merge-base", "--is-ancestor", wt.branch, trunk)
			if !st.Merged {
				st.Reason = "not merged into " + trunk
			}
		}

		// `git worktree remove` deletes git-ignored files without asking, and a
		// spec folder is ignored on machines whose global gitignore covers
		// .plaesy/. Such a worktree is kept, so no uncommitted spec is lost.
		if st.Merged {
			if ign := ignoredPaths(wt.path); len(ign) > 0 {
				st.Merged = true
				st.Reason = fmt.Sprintf("%d git-ignored path(s) would be deleted with it (e.g. %s); not removed", len(ign), ign[0])
				statuses = append(statuses, st)
				continue
			}
		}

		if apply && st.Merged {
			if out, err := gitIn(mainRoot, "worktree", "remove", wt.path); err != nil {
				st.Reason = fmt.Sprintf("git worktree remove refused: %v: %s", err, strings.TrimSpace(out))
			} else {
				st.Removed = true
			}
		}
		statuses = append(statuses, st)
	}

	if apply {
		// Drop admin records for worktrees whose folders were deleted by hand.
		_, _ = gitIn(mainRoot, "worktree", "prune")
	}
	return statuses, nil
}

// trunkBranch returns main, or master when main does not exist.
func trunkBranch(mainRoot string) (string, error) {
	for _, b := range []string{"main", "master"} {
		if gitOK(mainRoot, "rev-parse", "--verify", "--quiet", "refs/heads/"+b) {
			return b, nil
		}
	}
	return "", fmt.Errorf("no trunk branch: neither main nor master exists")
}

type worktreeEntry struct {
	path   string
	branch string
}

// parseWorktreeList reads `git worktree list --porcelain`. Each entry starts
// with a "worktree <path>" line and may carry "branch refs/heads/<name>".
func parseWorktreeList(out string) []worktreeEntry {
	var entries []worktreeEntry
	var cur *worktreeEntry
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimRight(line, "\r")
		switch {
		case strings.HasPrefix(line, "worktree "):
			entries = append(entries, worktreeEntry{path: filepath.Clean(strings.TrimPrefix(line, "worktree "))})
			cur = &entries[len(entries)-1]
		case strings.HasPrefix(line, "branch ") && cur != nil:
			cur.branch = strings.TrimPrefix(strings.TrimPrefix(line, "branch "), "refs/heads/")
		}
	}
	return entries
}

// ignoredPaths lists the git-ignored paths in a worktree, as reported by
// `git status --ignored`. A git failure reports none, so the caller falls back
// to git's own refusal rules.
func ignoredPaths(dir string) []string {
	out, err := gitIn(dir, "status", "--porcelain", "--ignored")
	if err != nil {
		return nil
	}
	var paths []string
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "!! ") {
			paths = append(paths, strings.TrimSpace(strings.TrimPrefix(line, "!! ")))
		}
	}
	return paths
}

func gitIn(dir string, args ...string) (string, error) {
	full := append([]string{"-C", dir}, args...)
	out, err := exec.Command("git", full...).CombinedOutput()
	return string(out), err
}

func gitOK(dir string, args ...string) bool {
	_, err := gitIn(dir, args...)
	return err == nil
}
