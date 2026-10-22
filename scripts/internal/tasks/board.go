package tasks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TaskRef is one task folder with the fields the board shows.
type TaskRef struct {
	Spec   string
	Status string
	Name   string
	Title  string
	Phase  string
}

// Problem is one inconsistency found by CheckIntegrity.
type Problem struct {
	Task    string
	Message string
}

// ListAllTasks returns every task of the project under tasksDir, ordered by status in
// Statuses order, then by name.
func ListAllTasks(tasksDir string) ([]TaskRef, error) {
	var refs []TaskRef
	for _, status := range Statuses {
		names, _, err := ListTasks(tasksDir, status)
		if err != nil {
			return nil, err
		}
		for _, name := range names {
			fm := readFrontmatter(filepath.Join(tasksDir, status, name, TaskFileName))
			refs = append(refs, TaskRef{Spec: fm["spec"], Status: status, Name: name, Title: fm["title"], Phase: fm["phase"]})
		}
	}
	return refs, nil
}

// CheckIntegrity reports three kinds of inconsistency: a task folder without its
// spec.md, a task whose frontmatter status differs from its status folder, and a task
// name that appears in two status folders. A task name is unique in the project.
func CheckIntegrity(tasksDir string) ([]Problem, error) {
	var problems []Problem
	seen := map[string]string{}
	for _, status := range Statuses {
		names, _, err := ListTasks(tasksDir, status)
		if err != nil {
			return nil, err
		}
		for _, name := range names {
			if prev, ok := seen[name]; ok {
				problems = append(problems, Problem{Task: name, Message: fmt.Sprintf("in both %s and %s", prev, status)})
			} else {
				seen[name] = status
			}
			doc := filepath.Join(tasksDir, status, name, TaskFileName)
			if _, err := os.Stat(doc); err != nil {
				problems = append(problems, Problem{Task: name, Message: "no spec.md in the task folder"})
				continue
			}
			declared, ok := readFrontmatter(doc)["status"]
			switch {
			case !ok:
				problems = append(problems, Problem{Task: name, Message: fmt.Sprintf("no status in frontmatter, folder is %s", status)})
			case declared != status:
				problems = append(problems, Problem{Task: name, Message: fmt.Sprintf("frontmatter status %s, folder is %s", declared, status)})
			}
		}
	}
	return problems, nil
}

// readFrontmatter returns the top-level "key: value" pairs of a file's YAML
// frontmatter, or an empty map when the file has none.
func readFrontmatter(path string) map[string]string {
	out := map[string]string{}
	data, err := os.ReadFile(path)
	if err != nil {
		return out
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return out
	}
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "---" {
			break
		}
		key, value, found := strings.Cut(line, ":")
		if !found || strings.HasPrefix(line, " ") {
			continue
		}
		out[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), `"'`)
	}
	return out
}
