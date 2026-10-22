// Package tasks manages the task lifecycle of a spec. A task is a folder:
//
//	.plaesy/specs/<spec>/tasks/<status>/<priority>_<slug>/
//	  spec.md          what the task is for
//	  contracts.json   the checklist that says when it is done
//
// The status is the folder the task sits in. Moving a task between statuses
// moves its whole folder.
package tasks

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/plaesy/spec-kit/internal/common"
)

// TaskFileName is the description file every task folder holds (spec.md).
const TaskFileName = "spec.md"

// Statuses are the status folders of a spec's tasks, in lifecycle order.
var Statuses = []string{"backlog", "todo", "doing", "done", "blocked"}

// priorities are the prefixes a task name may start with, most urgent first.
var priorities = []string{"critical", "high", "medium", "low"}

// ValidateStatus returns an error unless status is one of Statuses.
func ValidateStatus(status string) error {
	for _, s := range Statuses {
		if status == s {
			return nil
		}
	}
	return fmt.Errorf("invalid status: %s (must be: %s)", status, strings.Join(Statuses, ", "))
}

// ValidatePriority returns an error unless priority is one of the known prefixes.
func ValidatePriority(priority string) error {
	for _, p := range priorities {
		if priority == p {
			return nil
		}
	}
	return fmt.Errorf("invalid priority: %s (must be: %s)", priority, strings.Join(priorities, ", "))
}

// FindProjectRoot walks up from the working directory to the nearest .plaesy folder.
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

// TasksDir returns the project's task folder: .plaesy/tasks. Every task lives here,
// whichever spec it belongs to; the spec is recorded in the task's spec.md.
func TasksDir(projectRoot string) string {
	return filepath.Join(projectRoot, ".plaesy", "tasks")
}

// SpecOf returns the spec a task belongs to, read from the spec field of its spec.md.
func SpecOf(tasksDir, status, task string) string {
	return readFrontmatter(filepath.Join(tasksDir, status, task, TaskFileName))["spec"]
}

// ResolveSpec returns the spec folder whose name ends in the git branch.
func ResolveSpec(projectRoot, branch string) (string, error) {
	return common.FindSpecName(projectRoot, branch)
}

// CreateStatusDirs creates the status folders under a spec's tasks folder.
func CreateStatusDirs(tasksDir string) error {
	for _, s := range Statuses {
		if err := os.MkdirAll(filepath.Join(tasksDir, s), 0o755); err != nil {
			return err
		}
	}
	return nil
}

// taskBodyTemplate is the spec.md of a new task. Only the title is filled in.
const taskBodyTemplate = `---
spec: %s
title: %s
phase: implement
status: backlog
---

## Goal

## Scope

## Context

## Description

## Decision

## KPI

## Verification

## Reference

## Dependencies

## Notes
`

// AddTask creates a task folder in a spec's backlog, named <priority>_<slug>, with
// its spec.md and an empty contracts.json. It returns the folder name. It refuses a
// name that already exists in any status.
func AddTask(tasksDir, spec, priority, slug, title string) (string, error) {
	if err := ValidatePriority(priority); err != nil {
		return "", err
	}
	if slug == "" || strings.ContainsAny(slug, `/\`) {
		return "", fmt.Errorf("invalid task slug %q", slug)
	}
	name := priority + "_" + slug
	if _, _, found := FindTask(tasksDir, name); found {
		return "", fmt.Errorf("task already exists: %s", name)
	}
	dir := filepath.Join(tasksDir, "backlog", name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(dir, TaskFileName), []byte(fmt.Sprintf(taskBodyTemplate, spec, title)), 0o644); err != nil {
		return "", err
	}
	empty := fmt.Sprintf("{\n  \"id\": %q,\n  \"checklist\": []\n}\n", name)
	if err := os.WriteFile(filepath.Join(dir, ContractFileName), []byte(empty), 0o644); err != nil {
		return "", err
	}
	return name, nil
}

// GetNextTask returns the most urgent task in backlog: the lowest priority rank,
// then the name order within a rank. spec limits the choice to tasks of that spec;
// an empty spec allows any. found is false when no task qualifies.
func GetNextTask(tasksDir, spec string) (string, bool, error) {
	names, found, err := ListTasks(tasksDir, "backlog")
	if err != nil || !found {
		return "", false, err
	}
	names = forSpec(tasksDir, "backlog", names, spec)
	if len(names) == 0 {
		return "", false, nil
	}
	sort.SliceStable(names, func(i, j int) bool {
		return priorityRank(names[i]) < priorityRank(names[j])
	})
	return names[0], true, nil
}

// forSpec keeps the tasks in names that belong to spec. An empty spec keeps them all.
func forSpec(tasksDir, status string, names []string, spec string) []string {
	if spec == "" {
		return names
	}
	var out []string
	for _, n := range names {
		if SpecOf(tasksDir, status, n) == spec {
			out = append(out, n)
		}
	}
	return out
}

// MoveTask moves one task folder from one status to another.
func MoveTask(tasksDir, task, fromStatus, toStatus string) error {
	if err := ValidateStatus(fromStatus); err != nil {
		return err
	}
	if err := ValidateStatus(toStatus); err != nil {
		return err
	}
	from := filepath.Join(tasksDir, fromStatus, task)
	to := filepath.Join(tasksDir, toStatus, task)
	if info, err := os.Stat(from); err != nil || !info.IsDir() {
		return fmt.Errorf("task not found in %s: %s", fromStatus, task)
	}
	if _, err := os.Stat(to); err == nil {
		return fmt.Errorf("task already exists in %s: %s", toStatus, task)
	}
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		return err
	}
	if err := os.Rename(from, to); err != nil {
		return err
	}
	return setFrontmatterStatus(filepath.Join(to, TaskFileName), toStatus)
}

// setFrontmatterStatus rewrites the status line of a task's frontmatter, so the file
// agrees with the folder it now sits in. A file with no status line is left alone.
func setFrontmatterStatus(path, status string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil // no spec.md, nothing to update
	}
	lines := strings.Split(string(data), "\n")
	inFront := false
	for i, line := range lines {
		if line == "---" {
			if inFront {
				break
			}
			inFront = true
			continue
		}
		if inFront && strings.HasPrefix(line, "status:") {
			lines[i] = "status: " + status
			return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644)
		}
	}
	return nil
}

// PopNextTask moves the next backlog task of spec to todo and returns its name. found is
// false when no task qualifies.
func PopNextTask(tasksDir, spec string) (task string, found bool, err error) {
	task, found, err = GetNextTask(tasksDir, spec)
	if err != nil || !found {
		return "", false, err
	}
	if err := MoveTask(tasksDir, task, "backlog", "todo"); err != nil {
		return "", false, err
	}
	return task, true, nil
}

// TaskStart moves a task from todo to doing.
func TaskStart(tasksDir, task string) error {
	return MoveTask(tasksDir, task, "todo", "doing")
}

// TaskComplete moves a task from doing to done.
func TaskComplete(tasksDir, task string) error {
	return MoveTask(tasksDir, task, "doing", "done")
}

// TaskBlock moves a task from doing to blocked and records why in its spec.md.
func TaskBlock(tasksDir, task, reason string) error {
	if reason == "" {
		reason = "No reason provided"
	}
	if err := MoveTask(tasksDir, task, "doing", "blocked"); err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(tasksDir, "blocked", task, TaskFileName), os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "\n## Blocked Reason\n%s\n", reason)
	return err
}

// TaskUnblock moves a task from blocked back to todo.
func TaskUnblock(tasksDir, task string) error {
	return MoveTask(tasksDir, task, "blocked", "todo")
}

// ListTasks returns the task folder names in one status, sorted. found is false
// when the status folder does not exist.
func ListTasks(tasksDir, status string) (tasks []string, found bool, err error) {
	if err := ValidateStatus(status); err != nil {
		return nil, false, err
	}
	entries, err := os.ReadDir(filepath.Join(tasksDir, status))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	for _, e := range entries {
		if e.IsDir() {
			tasks = append(tasks, e.Name())
		}
	}
	sort.Strings(tasks)
	return tasks, true, nil
}

// FindTask returns the status folder that holds a task, searching in Statuses order.
func FindTask(tasksDir, task string) (status, path string, found bool) {
	for _, s := range Statuses {
		p := filepath.Join(tasksDir, s, task)
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			return s, p, true
		}
	}
	return "", "", false
}

// ReadTask returns the spec.md of a task. status narrows the search to one status
// folder; an empty status searches all of them.
func ReadTask(tasksDir, task, status string) (content, resolvedStatus string, err error) {
	if status == "" {
		s, _, found := FindTask(tasksDir, task)
		if !found {
			return "", "", fmt.Errorf("task not found: %s", task)
		}
		status = s
	} else if err := ValidateStatus(status); err != nil {
		return "", "", err
	}
	data, err := os.ReadFile(filepath.Join(tasksDir, status, task, TaskFileName))
	if err != nil {
		return "", "", err
	}
	return string(data), status, nil
}
