package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/plaesy/spec-kit/internal/common"
)

// ReloadOptions configures a call to Reload.
type ReloadOptions struct {
	// TargetDir is the project to refresh. Defaults to ".".
	TargetDir string

	// AIPlatform refreshes the platform-specific files too (the platform core
	// file, the prompt directory, and the agent roles). Empty means "only
	// refresh .plaesy/", which is the safe default: the platform is not
	// recorded anywhere on disk, so guessing it could write a second platform's
	// files into the project.
	AIPlatform string

	// PlaesyHome overrides the repo root holding the sources to copy from.
	PlaesyHome string

	// DryRun reports what would change without writing anything.
	DryRun bool

	// Prune lists the files in tool-owned trees that the sources no longer
	// produce — a mirror that outlived a renamed or deleted prompt. On its own
	// it deletes nothing; it is the report that makes the orphan visible.
	Prune bool

	// PruneApply turns Prune's report into deletions. It has no effect unless
	// Prune is set. Deletion is gated twice because these trees are gitignored
	// and version control cannot undo it.
	PruneApply bool
}

// ReloadReport is the per-file outcome of a reload.
//
// The distinction between Updated and Unchanged is the point of the command:
// `init` cannot report it (it skips everything that exists), so before reload
// there was no way to ask "is my .plaesy/ current?" and get an answer.
type ReloadReport struct {
	DryRun    bool
	Created   []string
	Updated   []string
	Unchanged []string
	Protected []string
	Failed    []string

	// Stale lists owned-tree files with no counterpart in the sources.
	// Pruned lists the ones that were actually deleted.
	Stale  []string
	Pruned []string
}

// Changed reports whether reload would alter (or has altered) anything.
func (r *ReloadReport) Changed() bool {
	return len(r.Created) > 0 || len(r.Updated) > 0 || len(r.Pruned) > 0
}

// Summary renders a one-line-per-category summary for the command output.
func (r *ReloadReport) Summary() string {
	prefix := ""
	if r.DryRun {
		prefix = "Would "
	}
	parts := []string{
		fmt.Sprintf("%s%d created", prefix, len(r.Created)),
		fmt.Sprintf("%d updated", len(r.Updated)),
		fmt.Sprintf("%d unchanged", len(r.Unchanged)),
	}
	if len(r.Protected) > 0 {
		parts = append(parts, fmt.Sprintf("%d preserved (your data)", len(r.Protected)))
	}
	if len(r.Stale) > 0 {
		parts = append(parts, fmt.Sprintf("%d stale (no source)", len(r.Stale)))
	}
	if len(r.Pruned) > 0 {
		parts = append(parts, fmt.Sprintf("%d pruned", len(r.Pruned)))
	}
	out := strings.Join(parts, ", ")
	if len(r.Failed) > 0 {
		out += fmt.Sprintf(", %d FAILED", len(r.Failed))
	}
	return out
}

// reloadProtectedPaths are the generated-shaped files that hold user data.
//
// memory.md, context.md and state.json are each *created* from a template and
// then edited by the workflow, so "it exists" does not mean "it is generated"
// and overwriting them would destroy the thing the user is trying to keep. They
// are still created when missing: a deleted memory.md is a repair, not a reason
// to stay deleted.
//
// Everything else under .plaesy/ that the workflow writes — specs/, tasks/,
// analysis/, memory/ — is never a copy target, so it needs no entry here. The
// lists are exhaustive by construction, not by convention: reload only ever
// calls the copy_* helpers, and those only ever write where a source file
// exists in the repo.
func reloadProtectedPaths(baseDir string) map[string]bool {
	paths := make(map[string]bool, 4)
	for _, name := range []string{"memory.md", "context.md", "state.json"} {
		paths[filepath.Join(baseDir, name)] = true
	}
	return paths
}

// reloadGeneratedDirs are the directories reload refreshes. They are created
// if missing, which repairs a partially deleted tree, but nothing outside this
// list is touched.
var reloadGeneratedDirs = []string{
	"instructions",
	"templates",
	"checklists",
	"scripts",
}

// Reload refreshes the generated parts of an existing .plaesy/ tree from the
// current sources, overwriting what is there.
//
// It exists because `init` cannot: copyFile has always returned early when the
// destination exists, `plaesy repair` and `plaesy upgrade` are stubs that
// print "not yet implemented", and so a project initialised before an
// instruction was edited keeps the stale copy permanently. Measured on this
// repo after one day of edits, 13 of 33 installed instruction files and 4 of 58
// installed templates differed from source with no command that would fix it.
//
// What reload will NOT do:
//
//   - it never deletes anything from a tree the user owns. A file in .plaesy/
//     with no counterpart in the sources is left alone, so a hand-written
//     template survives. Pass --prune to report orphans in the tool-owned
//     trees (the agent roles and the platform prompt mirror), and
//     --prune --apply to remove them. See prune.go for why the two kinds of
//     tree need opposite treatment;
//   - it never overwrites memory.md, context.md or state.json;
//   - it never writes specs/, tasks/, analysis/ or memory/;
//   - it does not guess the AI platform. Pass --ai if you want the platform
//     core file, prompts and roles refreshed too.
func Reload(opts ReloadOptions) (*ReloadReport, error) {
	targetDir := opts.TargetDir
	if targetDir == "" {
		targetDir = "."
	}
	absTarget, err := filepath.Abs(targetDir)
	if err != nil {
		return nil, fmt.Errorf("resolving target directory: %w", err)
	}

	home, err := FindHome(opts.PlaesyHome)
	if err != nil {
		return nil, err
	}

	cfg, err := LoadPlatformConfig(home)
	if err != nil {
		return nil, fmt.Errorf("%w (resolved Plaesy home %q has no usable platform.json)", err, home)
	}

	baseDir := filepath.Join(absTarget, cfg.Plaesy.BaseDirectory)
	if info, statErr := os.Stat(baseDir); statErr != nil || !info.IsDir() {
		if statErr != nil && os.IsNotExist(statErr) {
			return nil, fmt.Errorf("no %s directory in %s — nothing to reload; run `plaesy init` first",
				cfg.Plaesy.BaseDirectory, absTarget)
		}
		return nil, fmt.Errorf("%s is not a directory: %s", cfg.Plaesy.BaseDirectory, baseDir)
	}

	report := &ReloadReport{DryRun: opts.DryRun}
	pol := copyPolicy{
		overwrite: true,
		dryRun:    opts.DryRun,
		protected: reloadProtectedPaths(baseDir),
		report:    &fileReport{base: baseDir},
	}

	if opts.DryRun {
		common.LogInfo("Dry run — no files will be written.")
	}

	common.LogInfo("%s", "Refreshing generated files in "+cfg.Plaesy.BaseDirectory+"/ from "+home+"...")

	for _, dir := range reloadGeneratedDirs {
		if err := os.MkdirAll(filepath.Join(baseDir, dir), 0o755); err != nil {
			return nil, err
		}
	}

	if err := copyInstructions(home, baseDir, pol); err != nil {
		return nil, err
	}
	if err := copyTemplates(home, baseDir, pol); err != nil {
		return nil, err
	}
	if err := copyChecklists(home, baseDir, pol); err != nil {
		return nil, err
	}
	if err := copyScripts(home, baseDir, pol); err != nil {
		return nil, err
	}

	// memory/, analysis/, context.md and state.json are user state, not
	// generated content. createMemoryStructure and createLoopState are what
	// init uses to create them, and both already no-op when the file exists,
	// so calling them here only repairs a tree that is missing them. Neither
	// is allowed to overwrite: pol.protected covers the three files.
	if err := createMemoryStructure(home, baseDir, pol); err != nil {
		return nil, err
	}
	if err := createLoopState(home, baseDir, pol); err != nil {
		return nil, err
	}

	platform := ""
	if opts.AIPlatform != "" && opts.AIPlatform != "none" {
		platform = normalizePlatform(opts.AIPlatform, cfg)
		if !cfg.HasPlatform(platform) {
			return nil, fmt.Errorf("invalid AI platform: %s (available: %v)", opts.AIPlatform, cfg.PlatformNames())
		}
		common.LogSuccess("%s", "Selected: "+cfg.DisplayName(platform))
		if err := setupPlatformConfig(home, absTarget, platform, cfg, pol); err != nil {
			return nil, err
		}
	} else {
		common.LogInfo("Skipping platform files — pass --ai <platform> to refresh the core file, prompts and roles.")
	}

	if opts.Prune {
		// The copy step above is what makes a candidate provable: the expected
		// set is derived from the same sources the mirror was just written
		// from, so an unreadable source shows up as a failure to copy, not as
		// a reason to delete.
		roots := pruneRoots(home, absTarget, baseDir, platform, cfg)
		report.Stale = findPruneCandidates(roots, absTarget)

		if len(report.Stale) > 0 && opts.PruneApply && !opts.DryRun {
			report.Pruned = applyPrune(roots, absTarget, report.Stale)
		}
	}

	report.Created = pol.report.created
	report.Updated = pol.report.updated
	report.Unchanged = pol.report.unchanged
	report.Protected = pol.report.protected
	report.Failed = pol.report.failed

	return report, nil
}
