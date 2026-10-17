// Package taskmanage ports scripts/bash/plaesy-task-manage.sh: task lifecycle
// management for the .plaesy/specs/<spec>/tasks/ directory (backlog/todo/doing/done/blocked).
// Task files are moved between status subdirectories by simple os.Rename;
// the bash script never parses task file content except to append a
// "## Blocked Reason" section on task_block, so this port does the same.
package taskmanage

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/plaesy/spec-kit/internal/common"
)

// Statuses are the five valid task status directories, in bash case-statement order.
var Statuses = []string{"backlog", "todo", "doing", "done", "blocked"}

// ValidateStatus mirrors validate_status().
func ValidateStatus(status string) error {
	for _, s := range Statuses {
		if status == s {
			return nil
		}
	}
	return fmt.Errorf("invalid status: %s (must be: backlog, todo, doing, done, blocked)", status)
}

// FindProjectRoot mirrors find_project_root(): walk up from cwd looking for
// a .plaesy directory.
func FindProjectRoot() (string, error) {
	current, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if info, err := os.Stat(filepath.Join(current, ".plaesy")); err == nil && info.IsDir() {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("project root not found (.plaesy directory not found)")
		}
		current = parent
	}
}

// SpecsDir returns PROJECT_ROOT/.plaesy/specs, the parent of every spec.
func SpecsDir(projectRoot string) string {
	return filepath.Join(projectRoot, ".plaesy", "specs")
}

// TasksDir returns the task directory owned by one spec:
// PROJECT_ROOT/.plaesy/specs/<specID>/tasks. Tasks belong to a spec, so there
// is no project-wide task directory.
func TasksDir(projectRoot, specID string) string {
	return filepath.Join(SpecsDir(projectRoot), specID, "tasks")
}

// ResolveSpec returns the spec folder name for a git branch. The rule lives in
// common.FindSpecName so features and tasks agree on it.
func ResolveSpec(projectRoot, branch string) (string, error) {
	return common.FindSpecName(projectRoot, branch)
}

// CreateStatusDirs creates the five status directories under a spec's tasks
// directory. It is called when a spec is created, so every spec starts with
// the same lifecycle layout.
func CreateStatusDirs(tasksDir string) error {
	for _, s := range Statuses {
		if err := os.MkdirAll(filepath.Join(tasksDir, s), 0o755); err != nil {
			return err
		}
	}
	return nil
}

// AddTask writes a new task file into a spec's backlog and returns its file
// name. The name follows {priority}_{slug}.md from .plaesy/instructions/tasks.md.
// It refuses to overwrite an existing file, so an added task never replaces
// one that is already in another status.
func AddTask(tasksDir, priority, slug, title string) (string, error) {
	if err := ValidatePriority(priority); err != nil {
		return "", err
	}
	if slug == "" || strings.ContainsAny(slug, `/\`) {
		return "", fmt.Errorf("invalid task slug %q", slug)
	}
	fileName := priority + "_" + slug + ".md"
	if _, _, found := FindTask(tasksDir, fileName); found {
		return "", fmt.Errorf("task already exists: %s", fileName)
	}
	backlog := filepath.Join(tasksDir, "backlog")
	if err := os.MkdirAll(backlog, 0o755); err != nil {
		return "", err
	}
	body := fmt.Sprintf("---\ntitle: %s\nphase: implement\nstatus: backlog\n---\n\n## Description\n\n## Acceptance Criteria\n\n- [ ] \n", title)
	f, err := os.OpenFile(filepath.Join(backlog, fileName), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := f.WriteString(body); err != nil {
		return "", err
	}
	return fileName, nil
}

// ValidatePriority accepts the priority prefixes used in task file names.
func ValidatePriority(priority string) error {
	switch priority {
	case "critical", "high", "medium", "low", "backlog":
		return nil
	}
	return fmt.Errorf("invalid priority: %s (must be: critical, high, medium, low, backlog)", priority)
}

// GetNextTask mirrors get_next_task(): the alphabetically-last *.md file in
// backlog (bash: `ls | sort -r | head -1`, i.e. reverse-sorted, first entry).
// Returns ("", false) when backlog is empty, matching the bash "no tasks" case.
func GetNextTask(tasksDir string) (string, bool, error) {
	backlogDir := filepath.Join(tasksDir, "backlog")
	if info, err := os.Stat(backlogDir); err != nil || !info.IsDir() {
		return "", false, fmt.Errorf("backlog directory not found: %s", backlogDir)
	}

	entries, err := os.ReadDir(backlogDir)
	if err != nil {
		return "", false, err
	}

	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			names = append(names, e.Name())
		}
	}
	if len(names) == 0 {
		return "", false, nil
	}
	sort.Sort(sort.Reverse(sort.StringSlice(names)))
	return names[0], true, nil
}

// MoveTask mirrors move_task(): moves a task file between status directories.
func MoveTask(tasksDir, taskFile, fromStatus, toStatus string) error {
	if err := ValidateStatus(fromStatus); err != nil {
		return err
	}
	if err := ValidateStatus(toStatus); err != nil {
		return err
	}

	fromPath := filepath.Join(tasksDir, fromStatus, taskFile)
	toPath := filepath.Join(tasksDir, toStatus, taskFile)

	if info, err := os.Stat(fromPath); err != nil || info.IsDir() {
		return fmt.Errorf("task not found: %s", fromPath)
	}

	if err := os.MkdirAll(filepath.Dir(toPath), 0o755); err != nil {
		return err
	}
	if err := os.Rename(fromPath, toPath); err != nil {
		return err
	}
	return nil
}

// PopNextTask mirrors pop_next_task(): moves the next backlog task to todo
// and returns its filename. found is false when backlog was empty.
func PopNextTask(tasksDir string) (taskFile string, found bool, err error) {
	taskFile, found, err = GetNextTask(tasksDir)
	if err != nil {
		return "", false, err
	}
	if !found {
		return "", false, nil
	}
	if err := MoveTask(tasksDir, taskFile, "backlog", "todo"); err != nil {
		return "", false, err
	}
	return taskFile, true, nil
}

// TaskStart mirrors task_start(): todo -> doing.
func TaskStart(tasksDir, taskFile string) error {
	return MoveTask(tasksDir, taskFile, "todo", "doing")
}

// TaskComplete mirrors task_complete(): doing -> done.
func TaskComplete(tasksDir, taskFile string) error {
	return MoveTask(tasksDir, taskFile, "doing", "done")
}

// TaskBlock mirrors task_block(): doing -> blocked, then appends a
// "## Blocked Reason" section to the moved file.
func TaskBlock(tasksDir, taskFile, reason string) error {
	if reason == "" {
		reason = "No reason provided"
	}
	if err := MoveTask(tasksDir, taskFile, "doing", "blocked"); err != nil {
		return err
	}

	blockedPath := filepath.Join(tasksDir, "blocked", taskFile)
	if info, err := os.Stat(blockedPath); err == nil && !info.IsDir() {
		f, err := os.OpenFile(blockedPath, os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := fmt.Fprintf(f, "\n## Blocked Reason\n%s\n", reason); err != nil {
			return err
		}
	}
	return nil
}

// TaskUnblock mirrors task_unblock(): blocked -> todo.
func TaskUnblock(tasksDir, taskFile string) error {
	return MoveTask(tasksDir, taskFile, "blocked", "todo")
}

// ListTasks mirrors list_tasks(): the *.md basenames (без ".md") in a status
// directory, sorted. found is false when the directory does not exist.
func ListTasks(tasksDir, status string) (tasks []string, found bool, err error) {
	if err := ValidateStatus(status); err != nil {
		return nil, false, err
	}

	dir := filepath.Join(tasksDir, status)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			tasks = append(tasks, strings.TrimSuffix(e.Name(), ".md"))
		}
	}
	sort.Strings(tasks)
	return tasks, true, nil
}

// FindTask mirrors show_task()'s search-all-statuses behavior: returns the
// status directory a task file lives in, searching in Statuses order.
func FindTask(tasksDir, taskFile string) (status string, filePath string, found bool) {
	for _, s := range Statuses {
		p := filepath.Join(tasksDir, s, taskFile)
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return s, p, true
		}
	}
	return "", "", false
}

// ReadTask mirrors show_task(): resolves the task file (in the given status,
// or by searching all statuses when status is empty) and returns its raw
// content plus the status it was found in.
func ReadTask(tasksDir, taskFile, status string) (content string, resolvedStatus string, err error) {
	var filePath string
	if status != "" {
		if err := ValidateStatus(status); err != nil {
			return "", "", err
		}
		filePath = filepath.Join(tasksDir, status, taskFile)
		resolvedStatus = status
	} else {
		var found bool
		resolvedStatus, filePath, found = FindTask(tasksDir, taskFile)
		if !found {
			return "", "", fmt.Errorf("task not found: %s", taskFile)
		}
	}

	if info, err := os.Stat(filePath); err != nil || info.IsDir() {
		return "", "", fmt.Errorf("task not found: %s", taskFile)
	}

	b, err := os.ReadFile(filePath)
	if err != nil {
		return "", "", err
	}
	return string(b), resolvedStatus, nil
}
