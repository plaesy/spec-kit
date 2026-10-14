package scaffold

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/plaesy/spec-kit/internal/common"
)

// loadAlwaysLoad reads the always_load list from instructions/mapping.json.
//
// mapping.json is the SINGLE SOURCE OF TRUTH for which instruction files
// `plaesy init` copies as always-loaded (shared across every project). This
// port reads it at runtime instead of hard-coding a duplicate list: the bash
// original kept a "last-resort fallback" list, but that duplicate is the exact
// anti-pattern this framework forbids (no second source of truth) and it
// silently drifts — the earlier Go port ALWAYS used the hard-coded list and
// ignored mapping.json's always_load entirely, so edits to mapping.json had no
// effect on install. A missing/corrupt mapping.json is a framework
// misconfiguration; we log a warning and skip the instruction copy rather than
// silently copying a stale list.
func loadAlwaysLoad(home string) ([]string, error) {
	return loadMappingList(home, "always_load")
}

// loadScopeLoad reads the scope_load list from instructions/mapping.json.
// These files are copied to .plaesy/instructions/ (so /assess:{scope} and
// equivalent scoped commands can read them on-demand) but are NOT
// always-loaded into the agent's active context — the agent reads them only when
// a specific dimension scope is invoked, keeping the base context lean. This
// mirrors how role files (.plaesy/roles/*.md) are available but not always in
// context.
func loadScopeLoad(home string) ([]string, error) {
	return loadMappingList(home, "scope_load")
}

func loadMappingList(home, key string) ([]string, error) {
	data, err := os.ReadFile(filepath.Join(home, "instructions", "mapping.json"))
	if err != nil {
		return nil, fmt.Errorf("read instructions/mapping.json: %w", err)
	}
	var m struct {
		Mappings struct {
			AlwaysLoad []string `json:"always_load"`
			ScopeLoad  []string `json:"scope_load"`
		} `json:"mappings"`
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse instructions/mapping.json: %w", err)
	}
	switch key {
	case "always_load":
		if len(m.Mappings.AlwaysLoad) == 0 {
			return nil, fmt.Errorf("instructions/mapping.json mappings.always_load is empty")
		}
		return m.Mappings.AlwaysLoad, nil
	case "scope_load":
		return m.Mappings.ScopeLoad, nil
	default:
		return nil, fmt.Errorf("unknown mapping list key: %s", key)
	}
}

// copyPolicy controls how a copy treats a destination that already exists.
//
// The zero value is `init`'s historical behaviour, byte for byte: skip
// anything already on disk. Every field exists for `reload`, which exists
// precisely because that skip is what lets the generated tree drift forever —
// `init` can never refresh a file it wrote before the source changed, and
// `repair`/`upgrade` are stubs, so a project that ran `init` once before an
// instruction was edited keeps the stale copy forever.
type copyPolicy struct {
	// overwrite replaces an existing destination instead of skipping it.
	overwrite bool

	// dryRun reports the outcome it would have produced without writing.
	dryRun bool

	// protected holds absolute destination paths that are never overwritten.
	// They are still *created* when missing — a deleted memory.md is a repair,
	// not a reason to stay deleted — but an existing one is left alone. This
	// is how reload protects user data that happens to be generated-shaped:
	// memory.md and context.md are created from templates and then edited.
	protected map[string]bool

	// report, when non-nil, accumulates one entry per file touched.
	report *fileReport
}

// isProtected reports whether dst must not be overwritten.
func (p copyPolicy) isProtected(dst string) bool {
	if len(p.protected) == 0 {
		return false
	}
	return p.protected[filepath.Clean(dst)]
}

func (p copyPolicy) dry() bool { return p.dryRun }

// fileReport accumulates the per-file outcome of a reload so the command can
// print a summary and, more importantly, so a caller can assert on it.
type fileReport struct {
	base      string
	created   []string
	updated   []string
	unchanged []string
	protected []string
	failed    []string
}

func (r *fileReport) record(dst, outcome string) {
	if r == nil {
		return
	}
	display := dst
	if r.base != "" {
		if rel, err := filepath.Rel(r.base, dst); err == nil && !strings.HasPrefix(rel, "..") {
			display = rel
		}
	}
	switch outcome {
	case "created":
		r.created = append(r.created, display)
	case "updated":
		r.updated = append(r.updated, display)
	case "unchanged":
		r.unchanged = append(r.unchanged, display)
	case "protected":
		r.protected = append(r.protected, display)
	case "failed":
		r.failed = append(r.failed, display)
	}
}

// copyFile copies src to dst, skipping (not overwriting) if dst already
// exists -- matching every copy_* function's "already exists, skipping"
// behavior throughout plaesy-init.sh.
//
// The source's permission bits are carried over, which is what the bash
// original got from `cp`: without this, a 0755 scripts/bash/*.sh would land
// in .plaesy/scripts/bash/ as 0644 and stop being runnable.
func copyFile(src, dst string, pol copyPolicy) (copied bool, err error) {
	if _, statErr := os.Stat(dst); statErr == nil {
		if pol.isProtected(dst) {
			common.LogWarning("%s", "  ⊘ "+filepath.Base(dst)+" (your file, left alone)")
			pol.report.record(dst, "protected")
			return false, nil
		}
		if !pol.overwrite {
			common.LogWarning("%s", "  ⊘ "+filepath.Base(dst)+" (already exists, skipping)")
			pol.report.record(dst, "unchanged")
			return false, nil
		}
		// Overwrite: compare first so the log distinguishes a real refresh from
		// a no-op re-copy. Without this, every reload prints a wall of "✓"
		// lines and the user cannot tell what actually moved.
		same, cmpErr := sameFileContents(src, dst)
		if cmpErr != nil {
			common.LogWarning("%s", "  ! "+filepath.Base(dst)+" (could not compare, copying anyway)")
		} else if same {
			common.LogInfo("%s", "  = "+filepath.Base(dst)+" (unchanged)")
			pol.report.record(dst, "unchanged")
			return false, nil
		}
		if pol.dry() {
			common.LogInfo("%s", "  ~ "+filepath.Base(dst)+" (would update)")
			pol.report.record(dst, "updated")
			return true, nil
		}
		if err := writeCopy(src, dst); err != nil {
			pol.report.record(dst, "failed")
			return false, err
		}
		common.LogInfo("%s", "  ↑ "+filepath.Base(dst)+" (updated)")
		pol.report.record(dst, "updated")
		return true, nil
	} else if pol.dry() {
		common.LogInfo("%s", "  + "+filepath.Base(dst)+" (would create)")
		pol.report.record(dst, "created")
		return true, nil
	}

	if err := writeCopy(src, dst); err != nil {
		pol.report.record(dst, "failed")
		return false, err
	}
	pol.report.record(dst, "created")
	return true, nil
}

// sameFileContents reports whether src and dst have identical bytes. A read
// error on either side is returned rather than reported as "different", so a
// permission problem surfaces as itself instead of silently forcing a copy.
func sameFileContents(src, dst string) (bool, error) {
	a, err := os.ReadFile(src)
	if err != nil {
		return false, err
	}
	b, err := os.ReadFile(dst)
	if err != nil {
		return false, err
	}
	return string(a) == string(b), nil
}

// writeCopy writes src's contents and permission bits to dst.
func writeCopy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	srcInfo, err := in.Stat()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode().Perm())
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	// OpenFile only applies the mode when it creates the file, and umask may
	// have cleared bits from it, so set it explicitly.
	if err := os.Chmod(dst, srcInfo.Mode().Perm()); err != nil {
		common.LogWarning("%s", "  ! "+filepath.Base(dst)+" (could not set permissions)")
	}
	return nil
}

// copyInstructions ports copy_instructions(): populates
// <target>/instructions/ from <home>/instructions/*.instructions.md,
// stripping the ".instructions" segment from the destination filename
// (quality-gates.instructions.md -> quality-gates.md).
func copyInstructions(home, targetDir string, pol copyPolicy) error {
	sourceDir := filepath.Join(home, "instructions")
	if !dirHasFiles(sourceDir) {
		common.LogWarning("Skipping instruction copy - source directory issues")
		return nil
	}

	common.LogInfo("Copying instructions to .plaesy/instructions/...")
	instructionsDir := filepath.Join(targetDir, "instructions")
	if err := os.MkdirAll(instructionsDir, 0o755); err != nil {
		return err
	}

	common.LogInfo("  (loading always-load instructions from mapping.json)")
	names, loadErr := loadAlwaysLoad(home)
	if loadErr != nil {
		common.LogWarning("%s", "  ✗ could not load mapping.json always_load list: "+loadErr.Error())
		return nil
	}
	for _, name := range names {
		srcFile := filepath.Join(sourceDir, name)
		if _, err := os.Stat(srcFile); err != nil {
			common.LogWarning("%s", "  ? "+name+" (source not found, skipping)")
			continue
		}
		base := strings.TrimSuffix(filepath.Base(srcFile), ".instructions.md")
		dstFile := filepath.Join(instructionsDir, base+".md")
		copied, err := copyFile(srcFile, dstFile, pol)
		if err != nil {
			common.LogWarning("%s", "  ✗ "+base+".md (failed to copy)")
			continue
		}
		if copied && !pol.dry() {
			common.LogInfo("%s", "  ✓ "+base+".md")
		}
	}

	// scope_load files: copied to .plaesy/instructions/ (available for on-demand
	// loading by /assess:{scope} and equivalent commands) but NOT always-loaded
	// into the agent's active context. Mirrors how role files are available but
	// not always in context.
	common.LogInfo("  (loading scope-load instructions from mapping.json)")
	scopeNames, scopeErr := loadScopeLoad(home)
	if scopeErr != nil {
		common.LogWarning("%s", "  could not load mapping.json scope_load list: "+scopeErr.Error())
	} else {
		for _, name := range scopeNames {
			srcFile := filepath.Join(sourceDir, name)
			if _, err := os.Stat(srcFile); err != nil {
				common.LogWarning("%s", "  ? "+name+" (source not found, skipping)")
				continue
			}
			base := strings.TrimSuffix(filepath.Base(srcFile), ".instructions.md")
			dstFile := filepath.Join(instructionsDir, base+".md")
			copied, err := copyFile(srcFile, dstFile, pol)
			if err != nil {
				common.LogWarning("%s", "  ✗ "+base+".md (failed to copy)")
				continue
			}
			if copied && !pol.dry() {
				common.LogInfo("%s", "  ✓ "+base+".md (scope-load)")
			}
		}
	}

	memoryTemplate := filepath.Join(home, "templates", "memory.template.md")
	memoryTarget := filepath.Join(targetDir, "memory.md")
	switch {
	case fileExists(memoryTarget) && pol.isProtected(memoryTarget):
		// User content. Never write, not even an empty scaffold over it.
		common.LogWarning("  ⊘ memory.md (your file, left alone)")
		pol.report.record(memoryTarget, "protected")
	case fileExists(memoryTarget) && !pol.overwrite:
		common.LogWarning("  ⊘ memory.md (already exists, skipping)")
		pol.report.record(memoryTarget, "unchanged")
	case !fileExists(memoryTemplate):
		// The template's absence is a distinct case from a failed copy, and
		// it is the documented one: memory.md is scaffolded empty. The old
		// chain folded this into a `if stat(memoryTarget) == nil` branch
		// *inside* `else if copyFile(...) == nil`, and copyFile only returns
		// a nil error when the destination already exists — which the Stat
		// above had just ruled out. So the branch was unreachable: a missing
		// templates/memory.template.md (a partial PLAESY_HOME) and a genuine
		// copy failure both fell through the whole chain silently, and the
		// function went on to report "Instructions copied" with no memory.md
		// on disk at all.
		//
		// The empty scaffold is written ONLY when there is nothing to lose.
		// The old ordering reached this branch for an existing file too, so a
		// reload against a home without the template blanked out the user's
		// memory.
		common.LogWarning("  ! memory.template.md not found, creating empty file")
		if err := os.WriteFile(memoryTarget, nil, 0o644); err != nil {
			common.LogWarning("  ✗ memory.md (could not create: %v)", err)
		} else {
			common.LogInfo("  ✓ memory.md (empty)")
		}
	default:
		if _, err := copyFile(memoryTemplate, memoryTarget, pol); err != nil {
			common.LogWarning("  ✗ memory.md (failed to copy: %v)", err)
		} else if !pol.dry() {
			common.LogInfo("  ✓ memory.md")
		}
	}

	common.LogSuccess("Instructions copied")
	return nil
}

// copyFlatDir copies every file directly under sourceDir (non-recursive,
// matching the bash glob loops) into targetDir/label, skipping existing
// files. exts, if non-empty, restricts to those extensions (with the dot);
// nil means "all files" (used for checklists which is *.md only via caller).
func copyFlatDir(home, targetDir, sourceSub, label string, exts []string, pol copyPolicy) error {
	sourceDir := filepath.Join(home, sourceSub)
	if !dirHasFiles(sourceDir) {
		common.LogWarning("%s", "Skipping "+label+" copy - source directory issues")
		return nil
	}

	common.LogInfo("%s", "Copying "+label+" to .plaesy/"+label+"/...")
	dstDir := filepath.Join(targetDir, label)
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return err
	}

	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return err
	}

	copiedCount := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if len(exts) > 0 && !hasAnyExt(e.Name(), exts) {
			continue
		}
		src := filepath.Join(sourceDir, e.Name())
		dst := filepath.Join(dstDir, e.Name())
		copied, err := copyFile(src, dst, pol)
		if err != nil {
			common.LogWarning("%s", "  ✗ "+e.Name()+" (failed to copy)")
			continue
		}
		if copied {
			copiedCount++
		}
	}
	if copiedCount > 0 && !pol.dry() {
		common.LogInfo("%s", "  ✓ "+label+" copied ("+itoa(copiedCount)+" files)")
	}
	common.LogSuccess("%s", strings.Title(label)+" copied")
	return nil
}

// copyTemplates ports copy_templates(): flat copy of templates/*.{md,json,yaml}.
func copyTemplates(home, targetDir string, pol copyPolicy) error {
	return copyFlatDir(home, targetDir, "templates", "templates", []string{".md", ".json", ".yaml"}, pol)
}

// copyChecklists ports copy_checklists(): flat copy of checklists/*.md.
func copyChecklists(home, targetDir string, pol copyPolicy) error {
	return copyFlatDir(home, targetDir, "checklists", "checklists", []string{".md"}, pol)
}

// copyScripts ports copy_scripts(): copies scripts/bash/*.sh,
// scripts/configs/* and scripts/powershell/*.ps1 into
// <target>/scripts/{bash,configs,powershell}/.
func copyScripts(home, targetDir string, pol copyPolicy) error {
	sourceDir := filepath.Join(home, "scripts")
	if !dirHasFiles(sourceDir) {
		common.LogWarning("Skipping script copy - source directory issues")
		return nil
	}

	common.LogInfo("Copying scripts to .plaesy/scripts/...")

	bashDst := filepath.Join(targetDir, "scripts", "bash")
	configsDst := filepath.Join(targetDir, "scripts", "configs")
	psDst := filepath.Join(targetDir, "scripts", "powershell")
	for _, d := range []string{bashDst, configsDst, psDst} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}

	copyExt := func(sub, dstDir, label string, exts []string) {
		src := filepath.Join(sourceDir, sub)
		entries, err := os.ReadDir(src)
		if err != nil {
			return
		}
		count := 0
		for _, e := range entries {
			if e.IsDir() || (len(exts) > 0 && !hasAnyExt(e.Name(), exts)) {
				continue
			}
			copied, err := copyFile(filepath.Join(src, e.Name()), filepath.Join(dstDir, e.Name()), pol)
			if err != nil {
				common.LogWarning("%s", "  ✗ "+e.Name()+" (failed to copy)")
				continue
			}
			if copied {
				count++
			}
		}
		if count > 0 && !pol.dry() {
			common.LogInfo("%s", "  ✓ "+label+" copied ("+itoa(count)+" files)")
		}
	}

	copyExt("bash", bashDst, "Bash scripts", []string{".sh"})
	copyExt("configs", configsDst, "Configuration files", nil)
	copyExt("powershell", psDst, "PowerShell scripts", []string{".ps1"})

	common.LogSuccess("Scripts copied")
	return nil
}

func hasAnyExt(name string, exts []string) bool {
	for _, ext := range exts {
		if strings.EqualFold(filepath.Ext(name), ext) {
			return true
		}
	}
	return false
}

func dirHasFiles(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return false
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() {
			return true
		}
	}
	// Allow dirs whose files are nested (e.g. templates with subdirs) -- a
	// shallow ReadDir miss doesn't necessarily mean "empty" the way the bash
	// `find | head -1 | wc -l` check did across the whole tree.
	return dirHasAnyFileRecursive(dir)
}

func dirHasAnyFileRecursive(dir string) bool {
	found := false
	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}

// fileExists reports whether path exists and is readable. os.Stat's error is
// deliberately collapsed to a bool: every caller is asking "is there something
// here to protect or skip", and a stat failure means there is not.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
