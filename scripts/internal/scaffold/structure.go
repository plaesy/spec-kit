package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/plaesy/spec-kit/internal/common"
)

const tasksReadme = `# Task Management

Tasks are organized by status:

- **backlog/** — Ideas, features, bugs (unscheduled)
- **todo/** — Ready to start (next in queue)
- **doing/** — In active work
- **done/** — Completed and validated
- **blocked/** — Waiting on dependency

## Task Lifecycle

` + "```" + `
backlog → todo → doing → [quality-gates] → done / blocked
` + "```" + `

Move tasks between directories as status changes. Use ` + "`/continue`" + ` or ` + "`/loop`" + ` for automated workflow.

**See ` + "`.plaesy/instructions/tasks.md`" + ` for detailed task management instructions.**
`

// createTaskStructure ports create_task_structure().
func createTaskStructure(targetDir string) error {
	common.LogInfo("Creating task management structure...")
	for _, d := range []string{"backlog", "todo", "doing", "done", "blocked"} {
		if err := os.MkdirAll(filepath.Join(targetDir, "tasks", d), 0o755); err != nil {
			return err
		}
	}
	readmePath := filepath.Join(targetDir, "tasks", "README.md")
	if err := os.WriteFile(readmePath, []byte(tasksReadme), 0o644); err != nil {
		return err
	}
	common.LogSuccess("Task structure created")
	return nil
}

// createMemoryStructure ports create_memory_structure().
func createMemoryStructure(home, targetDir string, pol copyPolicy) error {
	common.LogInfo("Creating memory and analysis directories...")
	if err := os.MkdirAll(filepath.Join(targetDir, "memory"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(targetDir, "analysis"), 0o755); err != nil {
		return err
	}

	contextTarget := filepath.Join(targetDir, "context.md")
	contextTemplate := filepath.Join(home, "templates", "context.template.md")
	switch {
	case fileExists(contextTarget) && pol.isProtected(contextTarget):
		// Session state the user has been writing into. The copyFile path
		// would also protect it, but falling through to the empty-file branch
		// below (because copyFile reports copied=false) blanked it — a reload
		// has no business destroying context.md.
		common.LogWarning("  ⊘ context.md (your file, left alone)")
		pol.report.record(contextTarget, "protected")
	case fileExists(contextTarget) && !pol.overwrite:
		common.LogWarning("  ⊘ context.md (already exists, skipping)")
		pol.report.record(contextTarget, "unchanged")
	case fileExists(contextTemplate):
		if copied, err := copyFile(contextTemplate, contextTarget, pol); err == nil && copied && !pol.dry() {
			common.LogInfo("  ✓ context.md")
		}
	case pol.dry():
		common.LogInfo("  + context.md (would create empty)")
		pol.report.record(contextTarget, "created")
	default:
		// Template missing or the copy failed: still leave a context.md on
		// disk, because an absent file reads as "never initialised" to
		// every command that looks for it.
		common.LogWarning("  ! context.template.md not found, creating empty file")
		os.WriteFile(contextTarget, nil, 0o644)
		common.LogInfo("  ✓ context.md (empty)")
	}

	common.LogSuccess("Memory and analysis structure created")
	return nil
}

// createDecisionStructure ports create_decision_structure(): creates the
// decisions/ directory and seeds it with an empty decisions.md index from
// the decisions.template.md template. The index is a navigation surface,
// not a container — it stays empty until decision topic files exist.
func createDecisionStructure(home, targetDir string, pol copyPolicy) error {
	common.LogInfo("Creating decision management structure...")

	decisionsDir := filepath.Join(targetDir, "decisions")
	if err := os.MkdirAll(decisionsDir, 0o755); err != nil {
		return err
	}

	decisionsTarget := filepath.Join(targetDir, "decisions.md")
	decisionsTemplate := filepath.Join(home, "templates", "decisions.template.md")
	switch {
	case fileExists(decisionsTarget) && pol.isProtected(decisionsTarget):
		// User content. Never write, not even an empty scaffold over it.
		common.LogWarning("  ⊘ decisions.md (your file, left alone)")
		pol.report.record(decisionsTarget, "protected")
	case fileExists(decisionsTarget) && !pol.overwrite:
		common.LogWarning("  ⊘ decisions.md (already exists, skipping)")
		pol.report.record(decisionsTarget, "unchanged")
	case !fileExists(decisionsTemplate):
		// The template's absence is a distinct case from a failed copy, and
		// it is the documented one: decisions.md is scaffolded empty.
		common.LogWarning("  ! decisions.template.md not found, creating empty file")
		if err := os.WriteFile(decisionsTarget, nil, 0o644); err != nil {
			common.LogWarning("  ✗ decisions.md (could not create: %v)", err)
		} else {
			common.LogInfo("  ✓ decisions.md (empty)")
		}
	default:
		if _, err := copyFile(decisionsTemplate, decisionsTarget, pol); err != nil {
			common.LogWarning("  ✗ decisions.md (failed to copy: %v)", err)
		} else if !pol.dry() {
			common.LogInfo("  ✓ decisions.md")
		}
	}

	common.LogSuccess("Decision management structure created")
	return nil
}

// createLoopState ports create_loop_state(): copies state.template.json,
// replacing the [TIMESTAMP] placeholder with the current UTC time.
func createLoopState(home, targetDir string, pol copyPolicy) error {

	stateTarget := filepath.Join(targetDir, "state.json")
	if _, err := os.Stat(stateTarget); err == nil {
		common.LogWarning("  ⊘ state.json (already exists, skipping)")
		pol.report.record(stateTarget, "protected")
		return nil
	}

	stateTemplate := filepath.Join(home, "templates", "state.template.json")
	data, err := os.ReadFile(stateTemplate)
	if err != nil {
		common.LogWarning("  ! state.template.json not found, creating empty file")
		os.WriteFile(stateTarget, nil, 0o644)
		common.LogInfo("  ✓ state.json (empty)")
		return nil
	}

	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	// Both spellings, because the template ships the canonical {{ }} form and
	// the old [TIMESTAMP] form is still what a hand-written or third-party
	// template may carry. Substituting only "[TIMESTAMP]" was a silent no-op
	// against the real template: `plaesy init` wrote a literal {{TIMESTAMP}}
	// into a fresh project's .plaesy/state.json, which is the file /loop reads
	// for its loop state. The regression test asserts on the real template
	// rather than on a fixture, because a fixture with [TIMESTAMP] is what let
	// the mismatch survive in the first place.
	rendered := strings.ReplaceAll(string(data), "[TIMESTAMP]", timestamp)
	rendered = strings.ReplaceAll(rendered, "{{TIMESTAMP}}", timestamp)
	if pol.dry() {
		common.LogInfo("  + state.json (would create)")
		pol.report.record(stateTarget, "created")
		return nil
	}
	if err := os.WriteFile(stateTarget, []byte(rendered), 0o644); err != nil {
		return err
	}
	common.LogInfo("  ✓ state.json")
	pol.report.record(stateTarget, "created")
	return nil
}

// CreateStructure ports create_structure(): builds <targetDir>/<base_directory>
// (default .plaesy) with its core directories, populates it via the copy_*
// helpers, and creates the project directories (docs, specs by default)
// alongside it.
func CreateStructure(home, targetDir string, cfg *PlatformConfig) error {
	return createStructure(home, targetDir, cfg, copyPolicy{})
}

// createStructure is CreateStructure with an explicit copy policy. Init passes
// the zero policy, which reproduces the original skip-if-exists behaviour
// exactly; reload passes an overwriting, protected-path policy.
func createStructure(home, targetDir string, cfg *PlatformConfig, pol copyPolicy) error {
	common.LogInfo("Creating Plaesy structure...")

	baseDir := filepath.Join(targetDir, cfg.Plaesy.BaseDirectory)
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return err
	}

	for _, dir := range cfg.Plaesy.CoreDirectories {
		if dir == "" {
			continue
		}
		common.LogInfo("%s", "Creating core directory: '"+filepath.Join(cfg.Plaesy.BaseDirectory, dir)+"'")
		if err := os.MkdirAll(filepath.Join(baseDir, dir), 0o755); err != nil {
			return err
		}
	}

	if err := copyInstructions(home, baseDir, pol); err != nil {
		return err
	}
	if err := copyTemplates(home, baseDir, pol); err != nil {
		return err
	}
	if err := copyChecklists(home, baseDir, pol); err != nil {
		return err
	}
	if err := copyScripts(home, baseDir, pol); err != nil {
		return err
	}
	// createTaskStructure is deliberately NOT policy-gated: it overwrites
	// tasks/README.md unconditionally (os.WriteFile, not copyFile) because
	// that file is constant generated prose. Reload does not call this path at
	// all — a refresh must not rewrite anything the user owns.
	if err := createTaskStructure(baseDir); err != nil {
		return err
	}
	if err := createMemoryStructure(home, baseDir, pol); err != nil {
		return err
	}
	if err := createDecisionStructure(home, baseDir, pol); err != nil {
		return err
	}
	if err := createLoopState(home, baseDir, pol); err != nil {
		return err
	}

	for _, dir := range cfg.Plaesy.ProjectDirectories {
		if dir == "" {
			continue
		}
		common.LogInfo("%s", "Creating project directory: '"+dir+"'")
		if err := os.MkdirAll(filepath.Join(targetDir, dir), 0o755); err != nil {
			return err
		}
	}

	common.LogSuccess("Plaesy structure created with dynamic configuration")
	return nil
}
