package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// `.kilo/commands/` is the copy the agent runtime actually reads. `prompts/` is
// the source an author edits. Nothing connected them.
//
// That gap was worth 68 of 69 files: every prompt edited in this project over
// two sessions was mirrored stale, and the mirror still carried
// `design-spine.md` — a file that has never existed — in six files, months after
// the source was corrected. `plaesy reload` does not close it either: without
// `--ai <platform>` it copies instructions, templates and checklists, then logs
// "Skipping platform files" and leaves every prompt behind. So the C3 fix was
// real in the source and absent from the artefact in use.
//
// The audit that reported C3 as fixed was scoped to the source tree, with
// `.kilo/` explicitly excluded as "generated". Excluding generated files is
// right for counting duplication — a mirror would inflate it — and wrong for
// asserting a fix, because the mirror is what runs.
//
// This is the guard that makes the two trees one contract.
func TestPromptMirrorMatchesSource(t *testing.T) {
	root := corpusRoot(t)
	srcRoot := filepath.Join(root, "prompts")
	mirrorRoot := filepath.Join(root, ".kilo", "commands")

	if _, err := os.Stat(mirrorRoot); err != nil {
		t.Skipf("no generated mirror at %s; nothing to compare against", relTo(root, mirrorRoot))
	}

	checked := 0
	var missing, differing []string
	err := filepath.Walk(srcRoot, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(p, ".md") {
			return nil
		}
		rel, relErr := filepath.Rel(srcRoot, p)
		if relErr != nil {
			return relErr
		}
		checked++

		mirror := filepath.Join(mirrorRoot, rel)
		mirrorBytes, readErr := os.ReadFile(mirror)
		if readErr != nil {
			missing = append(missing, filepath.ToSlash(rel))
			return nil
		}
		srcBytes, readErr := os.ReadFile(p)
		if readErr != nil {
			return readErr
		}
		if !bytes.Equal(srcBytes, mirrorBytes) {
			differing = append(differing, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if checked == 0 {
		t.Fatal("no prompt files found under prompts/; this would pass without checking anything")
	}

	for _, m := range missing {
		t.Errorf("prompts/%s has no copy at .kilo/commands/%s.\n"+
			"  The mirror is what the agent runtime loads, so the prompt does not exist as far as "+
			"  behaviour is concerned. Run `plaesy reload --ai <platform>`.",
			m, m)
	}
	for _, d := range differing {
		t.Errorf("prompts/%s differs from .kilo/commands/%s.\n"+
			"  The runtime reads the mirror, so the edit is not in effect. Every fix verified "+
			"  against the source tree alone is unverified where it counts.\n"+
			"  Run `plaesy reload --ai <platform>`.",
			d, d)
	}
	if len(missing) == 0 && len(differing) == 0 {
		t.Logf("%d prompt file(s) mirrored byte-identically", checked)
	}
}

// TestMirrorCarriesNoReferencesMissingFromSource catches the inverse direction:
// the mirror kept `design-spine.md` for months because the source stopped
// referencing it and nobody compared. Checking only source links would not see
// it, because the stale reference lives only in the mirror.

// TestMirrorHasNoFileMissingFromSource closes the other gap: a file that exists
// in the mirror and *not* in the source.
//
// `TestPromptMirrorMatchesSource` walks `prompts/` and asks whether each source
// file has a matching copy, so by construction it never sees a file the source
// does not have. `plaesy reload` has no prune step — it copies what exists and
// leaves everything else — so renaming or deleting a prompt leaves the old
// copy behind, still complete and still loadable by the agent runtime.
//
// The cost is that the deletion is invisible. When `/create:doc:design` moved to
// `/spec:design`, `prompts/create/doc/design.md` was gone and
// `.kilo/commands/create/doc/design.md` remained, still titled
// "`/create:doc:design` command instructions" and still invokable. Both
// commands answer; one of them is a file the project no longer has.
//
// The stale copy is not merely redundant: it is a full working command, so an
// agent that discovered it would use it and report having followed the project's
// own convention. Compare it against the source tree, which is the only place
// the answer to "does this command still exist" is recorded.
func TestMirrorHasNoFileMissingFromSource(t *testing.T) {
	root := corpusRoot(t)
	mirrorRoot := filepath.Join(root, ".kilo", "commands")

	if _, err := os.Stat(mirrorRoot); err != nil {
		t.Skipf("no generated mirror at %s; nothing to compare against", relTo(root, mirrorRoot))
	}

	srcRoot := filepath.Join(root, "prompts")
	var orphans []string
	scanned := 0
	err := filepath.Walk(mirrorRoot, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(p, ".md") {
			return nil
		}
		scanned++
		rel, relErr := filepath.Rel(mirrorRoot, p)
		if relErr != nil {
			return relErr
		}
		if _, statErr := os.Stat(filepath.Join(srcRoot, rel)); statErr != nil {
			orphans = append(orphans, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(orphans)

	for _, o := range orphans {
		t.Errorf("%s exists in the mirror but not in prompts/.\n"+
			"  `plaesy reload` copies and overwrites but never prunes, so a renamed or deleted "+
			"  prompt leaves its old copy behind — complete, loadable, and titled with the old "+
			"  command name. An agent that finds it uses a file the project no longer has, and "+
			"  reports having followed the project's own convention. Delete the stale file, or "+
			"  teach `plaesy reload` to prune.",
			o)
	}
	t.Logf("%d mirror file(s) checked, %d with no source", scanned, len(orphans))
}

func TestMirrorReferencesNoPathMissingFromSource(t *testing.T) {
	root := corpusRoot(t)
	mirrorRoot := filepath.Join(root, ".kilo", "commands")
	if _, err := os.Stat(mirrorRoot); err != nil {
		t.Skipf("no generated mirror at %s", relTo(root, mirrorRoot))
	}

	// Paths that have never existed in this project, in any generation.
	knownGhosts := []string{
		"design-spine.md", // C3: 43 references, file never created
	}

	found := 0
	scanned := 0
	err := filepath.Walk(mirrorRoot, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(p, ".md") {
			return nil
		}
		scanned++
		raw, readErr := os.ReadFile(p)
		if readErr != nil {
			return readErr
		}
		for _, ghost := range knownGhosts {
			if bytes.Contains(raw, []byte(ghost)) {
				found++
				t.Errorf("%s references %q, which does not exist in the source tree.\n"+
					"  This is a stale mirror: the source was corrected and the generated copy "+
					"  was not. `plaesy reload` without `--ai <platform>` does not refresh "+
					"  prompts, so this survives a normal reload.",
					relTo(root, p), ghost)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%d ghost reference(s) across %d mirror file(s)", found, scanned)
}

// TestNoFileDuplicatesAnotherOutsideTheKnownMirrors closes the one direction
// the two tests above cannot see: neither of them knows whether a file is
// *allowed* to exist twice.
//
// It was written because `docs/CHANGELOG.md` turned out to be a byte-identical
// copy of `.plaesy/instructions.md` — an instructions index filed under a
// changelog name, untracked, unreferenced, and 60-odd relative links away from
// anything real. It was found only incidentally, by the dead-link guard
// reporting sixty link failures with a common cause nobody had looked for. A
// stray copy of a file that happens to contain no relative links would have
// been invisible to everything in this package.
//
// The repository duplicates content on purpose, and always in pairs: an
// editable source and the copy a tool actually reads. That is 207 pairs across
// six families. The property worth protecting is not "no duplicates" — it is
// "every duplicate is one of the pairs the build already expects", so a copy
// nobody asked for cannot accumulate beside a hand-maintained file and drift
// silently out of agreement with it.
func TestNoFileDuplicatesAnotherOutsideTheKnownMirrors(t *testing.T) {
	root := corpusRoot(t)

	// Text-ish suffixes only. Hashing the tree including the 14.9 MB tracked
	// `plaesy` binary and any .git objects would cost minutes to learn nothing.
	hashable := map[string]bool{
		".md": true, ".markdown": true, ".json": true, ".yaml": true, ".yml": true,
		".go": true, ".sh": true, ".ps1": true, ".txt": true, ".toml": true,
	}
	skipDirs := map[string]bool{".git": true, "node_modules": true, "graphify-out": true}
	// scripts/internal/assets/data is itself a generated, seventh mirror
	// family: a 1:1 go:embed copy of templates/, instructions/, prompts/,
	// agents/, checklists/ and scripts/configs (see internal/assets/gen),
	// kept current by `go run ./internal/assets/gen` rather than by the
	// per-platform install flow mirrorFamily already knows about. Every file
	// under it is expected to be byte-identical to its real source, so
	// including it here would just duplicate TestAssetsDataMatchesSource's
	// job with a worse error message.
	const assetsDataPrefix = "scripts/internal/assets/data/"

	byDigest := map[string][]string{}
	scanned := 0
	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel := relTo(root, p)
		if info.IsDir() {
			if skipDirs[info.Name()] || rel+"/" == assetsDataPrefix {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(rel, assetsDataPrefix) {
			return nil
		}
		if !hashable[strings.ToLower(filepath.Ext(p))] {
			return nil
		}
		// An empty file is not a copy of anything; several exist by design.
		if info.Size() == 0 {
			return nil
		}
		raw, readErr := os.ReadFile(p)
		if readErr != nil {
			return readErr
		}
		sum := sha256.Sum256(raw)
		byDigest[fmt.Sprintf("%x", sum[:])] = append(byDigest[fmt.Sprintf("%x", sum[:])], rel)
		scanned++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if scanned == 0 {
		t.Fatal("no hashable files found; this would pass without checking anything")
	}

	digests := make([]string, 0, len(byDigest))
	for d := range byDigest {
		digests = append(digests, d)
	}
	sort.Strings(digests)

	pairs := 0
	for _, d := range digests {
		group := byDigest[d]
		if len(group) < 2 {
			continue
		}
		sort.Strings(group)

		// Every member of a legitimate duplicate group carries the same family
		// key, and a non-member carries none. Anything else is a copy that
		// nothing in the build asked for.
		key := mirrorFamily(group[0])
		legitimate := key != ""
		for _, f := range group[1:] {
			if mirrorFamily(f) != key {
				legitimate = false
			}
		}
		if legitimate {
			pairs++
			continue
		}

		described := make([]string, 0, len(group))
		for _, f := range group {
			if k := mirrorFamily(f); k != "" {
				described = append(described, f+" (mirrors "+k+")")
			} else {
				described = append(described, f+" (belongs to no known source/mirror family)")
			}
		}
		t.Errorf("these %d files are byte-identical and are not a known source/mirror pair:\n    %s\n"+
			"A copy nobody asked for drifts out of agreement with the file it was copied from and "+
			"nothing reports it. If the second file is deliberate, it is a mirror family this test "+
			"does not know about — add it to mirrorFamily. If it is not, delete it.",
			len(group), strings.Join(described, "\n    "))
	}
	t.Logf("%d file(s) hashed; %d byte-identical group(s), all known source/mirror pairs", scanned, pairs)
}

// mirrorFamily returns the shared identity of a source/mirror pair, or "" for a
// file that has no counterpart.
//
// The key is deliberately coarse: family plus base name, ignoring the
// directory and the extension decoration (`x.instructions.md` in the source
// becomes `x.md` when installed). Pairing is a property of the two paths, so
// anything finer would be an assertion about content that the parity tests
// above already own.
func mirrorFamily(rel string) string {
	// The platform core file is a copy of exactly one source — the file
	// `plaesy.mapping.core.value` names — installed to whatever path each
	// platform reads its instructions from. Two platforms may even share a
	// target (`opencode` and `kilo` both install to `AGENTS.md`), so the
	// set of legitimate targets is read from the config rather than listed here:
	// a list would have to be edited every time a platform is added, and a
	// forgotten entry turns the generated core file into a "stray duplicate"
	// that fails a guard describing work the framework just did correctly.
	if coreMirrorTargets[toSlash(rel)] {
		return "platform core file"
	}
	if prefix, ok := matchPromptMirrorPrefix(rel); ok {
		return "prompt " + strings.TrimPrefix(rel, prefix)
	}
	switch {
	case strings.HasPrefix(rel, "prompts/"):
		return "prompt " + strings.TrimPrefix(rel, "prompts/")
	case strings.HasPrefix(rel, "instructions/"):
		return "instruction " + strings.TrimSuffix(strings.TrimPrefix(rel, "instructions/"), ".instructions.md")
	case strings.HasPrefix(rel, ".plaesy/instructions/"):
		return "instruction " + strings.TrimSuffix(strings.TrimPrefix(rel, ".plaesy/instructions/"), ".md")
	case strings.HasPrefix(rel, "agents/"):
		return "agent " + strings.TrimSuffix(strings.TrimPrefix(rel, "agents/"), ".agents.md")
	case strings.HasPrefix(rel, ".plaesy/roles/"):
		return "agent " + strings.TrimSuffix(strings.TrimPrefix(rel, ".plaesy/roles/"), ".md")
	case strings.HasPrefix(rel, "templates/"):
		return "template " + strings.TrimPrefix(rel, "templates/")
	case strings.HasPrefix(rel, ".plaesy/templates/"):
		return "template " + strings.TrimPrefix(rel, ".plaesy/templates/")
	case strings.HasPrefix(rel, "checklists/"):
		return "checklist " + strings.TrimPrefix(rel, "checklists/")
	case strings.HasPrefix(rel, ".plaesy/checklists/"):
		return "checklist " + strings.TrimPrefix(rel, ".plaesy/checklists/")
	case strings.HasPrefix(rel, "scripts/configs/"):
		return "config " + strings.TrimPrefix(rel, "scripts/configs/")
	case strings.HasPrefix(rel, ".plaesy/scripts/configs/"):
		return "config " + strings.TrimPrefix(rel, ".plaesy/scripts/configs/")
	}
	return ""
}

// coreMirrorTargets is the set of repo-relative paths the shipped platform.json
// installs the core instruction file to, plus the source itself. It is loaded
// once from the real config; an unreadable or absent config yields the source
// only, so the guard reports the core file as an unexplained duplicate rather
// than silently accepting every file in the tree.
var coreMirrorTargets = loadCoreMirrorTargets()

func loadCoreMirrorTargets() map[string]bool {
	targets := map[string]bool{}
	raw, err := os.ReadFile(filepath.Join("..", "..", "configs", "platform.json"))
	if err != nil {
		return targets
	}
	var doc struct {
		Plaesy struct {
			Mapping struct {
				Core struct {
					Value string `json:"value"`
				} `json:"core"`
			} `json:"mapping"`
		} `json:"plaesy"`
		Platforms map[string]struct {
			Mapping map[string]string `json:"mapping"`
		} `json:"platforms"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return targets
	}
	if src := doc.Plaesy.Mapping.Core.Value; src != "" {
		targets[toSlash(src)] = true
	}
	for _, p := range doc.Platforms {
		if core := p.Mapping["core"]; core != "" {
			targets[toSlash(core)] = true
		}
	}
	return targets
}

// promptMirrorPrefixes is every platform's prompts destination directory
// (e.g. ".claude/commands/", ".kilo/commands/", ".cursor/rules/"), read from
// the shipped platform.json rather than hardcoded. Hardcoding one platform's
// directory here is exactly the bug this file otherwise guards against: a
// second platform's prompt mirror (".claude/commands/" was missing until this
// was generalized) would fail as an "unexplained duplicate" even though it is
// `plaesy init`'s own, correct output.
var promptMirrorPrefixes = loadPromptMirrorPrefixes()

func loadPromptMirrorPrefixes() []string {
	var prefixes []string
	raw, err := os.ReadFile(filepath.Join("..", "..", "configs", "platform.json"))
	if err != nil {
		return prefixes
	}
	var doc struct {
		Platforms map[string]struct {
			Mapping map[string]string `json:"mapping"`
		} `json:"platforms"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return prefixes
	}
	for _, p := range doc.Platforms {
		if dest := p.Mapping["prompts"]; dest != "" {
			prefixes = append(prefixes, strings.TrimSuffix(toSlash(dest), "/")+"/")
		}
	}
	return prefixes
}

// matchPromptMirrorPrefix reports whether rel sits under some platform's
// prompts destination, and the matching prefix (so the caller can strip it
// off and key on the same relative sub-path prompts/ uses).
func matchPromptMirrorPrefix(rel string) (string, bool) {
	for _, prefix := range promptMirrorPrefixes {
		if strings.HasPrefix(rel, prefix) {
			return prefix, true
		}
	}
	return "", false
}

func toSlash(p string) string { return strings.ReplaceAll(filepath.ToSlash(p), "\\", "/") }
