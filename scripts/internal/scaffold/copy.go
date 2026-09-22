package scaffold

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/plaesy/spec-kit/internal/common"
)

// defaultInstructions mirrors plaesy-init.sh's copy_instructions() last-resort
// fallback list (used when mapping.json's always_load can't be read). This
// port does not carry over detect-stack.sh's stack-detection selection or
// mapping.json's always_load list -- see the "not ported 1:1" note in the
// command's help text / task report.
var defaultInstructions = []string{
	"plaesy.instructions.md",
	"plaesy-trim.instructions.md",
	"plaesy-graph.instructions.md",
	"tasks.instructions.md",
	"quality-gates.instructions.md",
	"error-recovery.instructions.md",
	"date-system.instructions.md",
	"assess-technical.instructions.md",
	"assess-design.instructions.md",
	"assess-business.instructions.md",
	"assess-financial.instructions.md",
	"assess-marketing.instructions.md",
	"assess-legal.instructions.md",
	"assess-management.instructions.md",
	"assess-product.instructions.md",
	"dimension-mapping.instructions.md",
	"error-recovery-predictive.instructions.md",
	"output-validation.instructions.md",
	"universal-orchestrator.instructions.md",
}

// copyFile copies src to dst, skipping (not overwriting) if dst already
// exists -- matching every copy_* function's "already exists, skipping"
// behavior throughout plaesy-init.sh.
func copyFile(src, dst string) (copied bool, err error) {
	if _, err := os.Stat(dst); err == nil {
		common.LogWarning("  ⊘ " + filepath.Base(dst) + " (already exists, skipping)")
		return false, nil
	}

	in, err := os.Open(src)
	if err != nil {
		return false, err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return false, err
	}

	out, err := os.Create(dst)
	if err != nil {
		return false, err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return false, err
	}
	return true, nil
}

// copyInstructions ports copy_instructions(): populates
// <target>/instructions/ from <home>/instructions/*.instructions.md,
// stripping the ".instructions" segment from the destination filename
// (quality-gates.instructions.md -> quality-gates.md).
func copyInstructions(home, targetDir string) error {
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

	common.LogInfo("  (using default always-load instructions)")
	for _, name := range defaultInstructions {
		srcFile := filepath.Join(sourceDir, name)
		if _, err := os.Stat(srcFile); err != nil {
			common.LogWarning("  ? " + name + " (source not found, skipping)")
			continue
		}
		base := strings.TrimSuffix(filepath.Base(srcFile), ".instructions.md")
		dstFile := filepath.Join(instructionsDir, base+".md")
		copied, err := copyFile(srcFile, dstFile)
		if err != nil {
			common.LogWarning("  ✗ " + base + ".md (failed to copy)")
			continue
		}
		if copied {
			common.LogInfo("  ✓ " + base + ".md")
		}
	}

	memoryTemplate := filepath.Join(home, "templates", "memory.template.md")
	memoryTarget := filepath.Join(targetDir, "memory.md")
	if _, err := os.Stat(memoryTarget); err == nil {
		common.LogWarning("  ⊘ memory.md (already exists, skipping)")
	} else if _, err := copyFile(memoryTemplate, memoryTarget); err == nil {
		if _, statErr := os.Stat(memoryTarget); statErr == nil {
			common.LogInfo("  ✓ memory.md")
		} else {
			common.LogWarning("  ! memory.template.md not found, creating empty file")
			os.WriteFile(memoryTarget, nil, 0o644)
			common.LogInfo("  ✓ memory.md (empty)")
		}
	}

	common.LogSuccess("Instructions copied")
	return nil
}

// copyFlatDir copies every file directly under sourceDir (non-recursive,
// matching the bash glob loops) into targetDir/label, skipping existing
// files. exts, if non-empty, restricts to those extensions (with the dot);
// nil means "all files" (used for checklists which is *.md only via caller).
func copyFlatDir(home, targetDir, sourceSub, label string, exts []string) error {
	sourceDir := filepath.Join(home, sourceSub)
	if !dirHasFiles(sourceDir) {
		common.LogWarning("Skipping " + label + " copy - source directory issues")
		return nil
	}

	common.LogInfo("Copying " + label + " to .plaesy/" + label + "/...")
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
		copied, err := copyFile(src, dst)
		if err != nil {
			common.LogWarning("  ✗ " + e.Name() + " (failed to copy)")
			continue
		}
		if copied {
			copiedCount++
		}
	}
	if copiedCount > 0 {
		common.LogInfo("  ✓ " + label + " copied (" + itoa(copiedCount) + " files)")
	}
	common.LogSuccess(strings.Title(label) + " copied")
	return nil
}

// copyTemplates ports copy_templates(): flat copy of templates/*.{md,json,yaml}.
func copyTemplates(home, targetDir string) error {
	return copyFlatDir(home, targetDir, "templates", "templates", []string{".md", ".json", ".yaml"})
}

// copyChecklists ports copy_checklists(): flat copy of checklists/*.md.
func copyChecklists(home, targetDir string) error {
	return copyFlatDir(home, targetDir, "checklists", "checklists", []string{".md"})
}

// copyScripts ports copy_scripts(): copies scripts/bash/*.sh,
// scripts/configs/* and scripts/powershell/*.ps1 into
// <target>/scripts/{bash,configs,powershell}/.
func copyScripts(home, targetDir string) error {
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
			copied, err := copyFile(filepath.Join(src, e.Name()), filepath.Join(dstDir, e.Name()))
			if err != nil {
				common.LogWarning("  ✗ " + e.Name() + " (failed to copy)")
				continue
			}
			if copied {
				count++
			}
		}
		if count > 0 {
			common.LogInfo("  ✓ " + label + " copied (" + itoa(count) + " files)")
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
