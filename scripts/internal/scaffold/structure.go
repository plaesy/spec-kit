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
func createMemoryStructure(home, targetDir string) error {
	common.LogInfo("Creating memory and analysis directories...")
	if err := os.MkdirAll(filepath.Join(targetDir, "memory"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(targetDir, "analysis"), 0o755); err != nil {
		return err
	}

	contextTarget := filepath.Join(targetDir, "context.md")
	if _, err := os.Stat(contextTarget); err == nil {
		common.LogWarning("  ⊘ context.md (already exists, skipping)")
	} else {
		contextTemplate := filepath.Join(home, "templates", "context.template.md")
		if copied, err := copyFile(contextTemplate, contextTarget); err == nil && copied {
			common.LogInfo("  ✓ context.md")
		} else {
			common.LogWarning("  ! context.template.md not found, creating empty file")
			os.WriteFile(contextTarget, nil, 0o644)
			common.LogInfo("  ✓ context.md (empty)")
		}
	}

	common.LogSuccess("Memory and analysis structure created")
	return nil
}

// createLoopState ports create_loop_state(): copies state.template.json,
// replacing the [TIMESTAMP] placeholder with the current UTC time.
func createLoopState(home, targetDir string) error {
	common.LogInfo("Creating autonomous loop configuration...")

	stateTarget := filepath.Join(targetDir, "state.json")
	if _, err := os.Stat(stateTarget); err == nil {
		common.LogWarning("  ⊘ state.json (already exists, skipping)")
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
	rendered := strings.ReplaceAll(string(data), "[TIMESTAMP]", timestamp)
	if err := os.WriteFile(stateTarget, []byte(rendered), 0o644); err != nil {
		return err
	}
	common.LogInfo("  ✓ state.json")
	return nil
}

// CreateStructure ports create_structure(): builds <targetDir>/<base_directory>
// (default .plaesy) with its core directories, populates it via the copy_*
// helpers, and creates the project directories (docs, specs by default)
// alongside it.
func CreateStructure(home, targetDir string, cfg *PlatformConfig) error {
	common.LogInfo("Creating Plaesy structure...")

	baseDir := filepath.Join(targetDir, cfg.Plaesy.BaseDirectory)
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return err
	}

	for _, dir := range cfg.Plaesy.CoreDirectories {
		if dir == "" {
			continue
		}
		common.LogInfo("Creating core directory: '" + filepath.Join(cfg.Plaesy.BaseDirectory, dir) + "'")
		if err := os.MkdirAll(filepath.Join(baseDir, dir), 0o755); err != nil {
			return err
		}
	}

	if err := copyInstructions(home, baseDir); err != nil {
		return err
	}
	if err := copyTemplates(home, baseDir); err != nil {
		return err
	}
	if err := copyChecklists(home, baseDir); err != nil {
		return err
	}
	if err := copyScripts(home, baseDir); err != nil {
		return err
	}
	if err := createTaskStructure(baseDir); err != nil {
		return err
	}
	if err := createMemoryStructure(home, baseDir); err != nil {
		return err
	}
	if err := createLoopState(home, baseDir); err != nil {
		return err
	}

	for _, dir := range cfg.Plaesy.ProjectDirectories {
		if dir == "" {
			continue
		}
		common.LogInfo("Creating project directory: '" + dir + "'")
		if err := os.MkdirAll(filepath.Join(targetDir, dir), 0o755); err != nil {
			return err
		}
	}

	common.LogSuccess("Plaesy structure created with dynamic configuration")
	return nil
}
