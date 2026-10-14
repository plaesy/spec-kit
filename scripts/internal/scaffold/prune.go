package scaffold

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/plaesy/spec-kit/internal/common"
)

// Prune exists because `reload` copies but never deletes, and the two
// properties that are right for a user-owned tree are wrong for a
// tool-owned one.
//
// A file in .plaesy/ with no counterpart in the sources is usually the user's
// work: a hand-written template, a note, a spec. Deleting it on sight would
// destroy exactly the thing reload is supposed to protect.
//
// The same file in a generated mirror is not user work. It is a copy that
// outlived its source, and it is worse than dead weight: it is a *complete,
// loadable command*. When /create:doc:design moved to /spec:design,
// prompts/create/doc/design.md was renamed and
// .kilo/commands/create/doc/design.md stayed behind, still titled
// "/create:doc:design command instructions" and still invokable. Both
// commands answered. One of them was a file the project no longer had, and an
// agent that picked it would report having followed the project's own
// convention.
//
// So the discriminator is ownership, not path shape, and this file implements
// only the "tool owns this tree outright" half. Mixed trees stay
// manifest-tracked rather than path-pruned: a manifest records the path and
// content hash of every file reload wrote, and only an orphan that still
// matches its recorded hash is garbage. An orphan the user has since edited is
// work, not garbage, and path-based pruning cannot tell those two apart.
//
// Deletion is opt-in twice. --prune reports, --prune --apply deletes, because
// these trees are gitignored (.kilo*, .claude*) and version control cannot
// bring back a wrongly deleted file.

// pruneRoot is a destination tree that reload owns in full: every file in it
// is a copy of a file under srcDir, and reload is the only thing that writes
// there.
type pruneRoot struct {
	// label names the tree in the command output.
	label string

	// destDir is the absolute destination root.
	destDir string

	// srcDir is the absolute source root the mirror was copied from.
	srcDir string

	// srcStrip is the source filename suffix that is replaced by dstExt.
	// ".agents.md" -> ".md" for roles; ".md" -> ".md" for a prompt mirror,
	// which is why a single rule covers both.
	srcStrip string

	// dstExt is the extension given to the copied file.
	dstExt string
}

// sharedPromptDirs are prompt destinations that are NOT dedicated to copied
// prompts. A file dropped in one of these by hand is indistinguishable from a
// mirror of a prompt that has since been renamed, so these are excluded from
// pruning regardless of what the platform's mapping claims.
//
// The platform config marks the exclusive ones with "prune_prompts": "true";
// this list is the deny list that keeps the mark honest. A platform that is
// absent from both is treated as shared, because the failure mode of guessing
// wrong is an unrecoverable deletion and the failure mode of being cautious
// is a stale file someone can delete by hand.
var sharedPromptDirs = map[string]bool{
	".cursor/rules": true,
	".qoder/rules":  true,
}

// pruneRoots returns the trees that are safe to prune for this invocation.
//
// It returns nothing for the prompt mirror when no platform was named. Reload
// does not guess the AI platform — it is not recorded anywhere on disk — and a
// guessed platform would prune a directory belonging to some *other* tool.
func pruneRoots(home, absTarget, baseDir, platform string, cfg *PlatformConfig) []pruneRoot {
	roots := []pruneRoot{{
		label:    "agent roles",
		destDir:  filepath.Join(baseDir, "roles"),
		srcDir:   filepath.Join(home, "agents"),
		srcStrip: ".agents.md",
		dstExt:   ".md",
	}}

	if platform == "" {
		return roots
	}

	entry, ok := cfg.Platforms[platform]
	if !ok {
		return roots
	}
	destRel, ok := entry.Mapping["prompts"]
	if !ok || destRel == "" || sharedPromptDirs[destRel] {
		return roots
	}
	excl, ok := entry.Mapping["prune_prompts"]
	if !ok {
		return roots
	}
	allowed, err := strconv.ParseBool(excl)
	if err != nil || !allowed {
		return roots
	}

	promptMapping, ok := cfg.Plaesy.Mapping["prompts"]
	if !ok || promptMapping.Value == "" {
		return roots
	}

	roots = append(roots, pruneRoot{
		label:    "prompt mirror",
		destDir:  filepath.Join(absTarget, filepath.FromSlash(destRel)),
		srcDir:   filepath.Join(home, filepath.FromSlash(strings.TrimSuffix(promptMapping.Value, "/*"))),
		srcStrip: ".md",
		dstExt:   promptExtension(platform),
	})
	return roots
}

// expectedPrunePaths maps every source file the mirror would contain, as a
// destination-relative path.
//
// It returns an error when srcDir cannot be walked. Callers must not prune in
// that case: an unreadable source produces an empty expected set, and an
// empty expected set makes every file in the tree look like an orphan. This is
// the one place where "delete what has no source" turns a misconfigured
// --plaesy-home into data loss.
func expectedPrunePaths(root pruneRoot) (map[string]bool, error) {
	expected := make(map[string]bool)

	err := filepath.WalkDir(root.srcDir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), root.srcStrip) {
			return nil
		}
		rel, relErr := filepath.Rel(root.srcDir, p)
		if relErr != nil {
			return relErr
		}
		name := strings.TrimSuffix(d.Name(), root.srcStrip) + root.dstExt
		expected[filepath.Clean(filepath.Join(filepath.Dir(rel), name))] = true
		return nil
	})
	if err != nil {
		return nil, err
	}
	return expected, nil
}

// withinRoot reports whether p is a real path inside root, and is the check
// that keeps a pruning bug from reaching outside the owned tree.
func withinRoot(root, p string) bool {
	rel, err := filepath.Rel(root, p)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// findPruneCandidates lists files in the owned trees that the sources no
// longer produce, as paths relative to the project root.
func findPruneCandidates(roots []pruneRoot, absTarget string) []string {
	var found []string

	for _, root := range roots {
		if info, err := os.Stat(root.srcDir); err != nil || !info.IsDir() {
			common.LogWarning("%s", "Skipping "+root.label+" prune — source not found: "+root.srcDir)
			continue
		}
		// An empty source directory is as dangerous as a missing one: the
		// expected set comes out empty, so every file in the tree looks like an
		// orphan. That happens for a real reason (a --plaesy-home pointing at a
		// tree whose sources have not been populated yet), and the cost of
		// guessing wrong is a deletion git cannot undo.
		if !dirHasFiles(root.srcDir) {
			common.LogWarning("%s", "Skipping "+root.label+" prune — source is empty: "+root.srcDir)
			continue
		}
		expected, err := expectedPrunePaths(root)
		if err != nil {
			common.LogWarning("%s", "Skipping "+root.label+" prune — source unreadable: "+err.Error())
			continue
		}
		if _, err := os.Stat(root.destDir); err != nil {
			continue
		}

		_ = filepath.WalkDir(root.destDir, func(p string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if !withinRoot(root.destDir, p) {
				return nil
			}
			rel, relErr := filepath.Rel(root.destDir, p)
			if relErr != nil {
				return nil
			}
			if expected[filepath.Clean(rel)] {
				return nil
			}
			if rel, relErr = filepath.Rel(absTarget, p); relErr == nil &&
				!strings.HasPrefix(rel, "..") {
				found = append(found, filepath.ToSlash(rel))
			}
			return nil
		})
	}

	sort.Strings(found)
	return found
}

// applyPrune removes the candidates and the directories they leave empty.
//
// The directory sweep is not cosmetic. The stale mirror this was written for
// left an entire .kilo/commands/create/doc/ tree behind, and a directory tree
// with no files in it is still a directory the runtime walks.
func applyPrune(roots []pruneRoot, absTarget string, candidates []string) []string {
	removed := make([]string, 0, len(candidates))

	for _, rel := range candidates {
		p := filepath.Join(absTarget, filepath.FromSlash(rel))
		owned := false
		for _, root := range roots {
			if withinRoot(root.destDir, p) {
				owned = true
				break
			}
		}
		if !owned {
			common.LogWarning("%s", "Refusing to remove "+rel+" — outside every owned tree")
			continue
		}
		if err := os.Remove(p); err != nil {
			common.LogWarning("%s", "  ✗ "+rel+" ("+err.Error()+")")
			continue
		}
		removed = append(removed, rel)
	}

	for _, root := range roots {
		removeEmptyDirs(root.destDir)
	}
	return removed
}

// removeEmptyDirs deletes directories under root that contain no files,
// deepest first. It stops at root itself.
func removeEmptyDirs(root string) {
	var dirs []string
	_ = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() || p == root {
			return nil
		}
		dirs = append(dirs, p)
		return nil
	})
	// WalkDir is lexically ordered, so a reverse pass is deepest-first.
	for i := len(dirs) - 1; i >= 0; i-- {
		entries, err := os.ReadDir(dirs[i])
		if err != nil || len(entries) > 0 {
			continue
		}
		_ = os.Remove(dirs[i])
	}
}
