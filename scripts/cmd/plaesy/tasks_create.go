package main

import (
	"fmt"

	"github.com/plaesy/spec-kit/internal/common"
	"github.com/plaesy/spec-kit/internal/featurepath"
	"github.com/plaesy/spec-kit/internal/tasks"
	"github.com/spf13/cobra"
)

// newTasksCreateCmd creates a task. Creating a task makes a new resource, so the verb is
// create; "add" is reserved for putting an existing item into a collection.
func newTasksCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create <priority> <slug> <title>",
		Short: "Create a task in the backlog",
		Long: `Create a task folder in .plaesy/tasks/backlog/<priority>_<slug>/, holding spec.md
and an empty contracts.json. The task records the spec it belongs to.

The spec is the current git branch, which is also the spec's name. If the spec does
not exist yet, it is created first, with spec.md, in .plaesy/specs/<spec>/. --spec
names another spec.

Priorities: critical, high, medium, low. The task name is <priority>_<slug>, and it
must be unique in the project.`,
		Example: `  plaesy tasks create high login-form "Build the login form"`,
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := tasks.FindProjectRoot()
			if err != nil {
				return err
			}
			branch, err := common.GetCurrentBranch()
			if err != nil {
				return err
			}
			specName := tasksSpec
			if specName == "" {
				specName = branch
			}
			if _, err := featurepath.EnsureSpec(root, specName, branch); err != nil {
				return err
			}
			tasksDir := tasks.TasksDir(root)
			if err := tasks.CreateStatusDirs(tasksDir); err != nil {
				return err
			}
			name, err := tasks.AddTask(tasksDir, specName, args[0], args[1], args[2])
			if err != nil {
				return err
			}
			fmt.Printf("[✓] Created %s for spec %s\n", name, specName)
			return nil
		},
	}
}
