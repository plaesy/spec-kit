package main

import (
	"fmt"

	"github.com/plaesy/spec-kit/internal/taskmanage"
	"github.com/spf13/cobra"
)

// newTasksAddCmd adds a task to a spec's backlog. A spec can take tasks at any
// point in its life, not only when it is created.
func newTasksAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <priority> <slug> <title>",
		Short: "Add a task to the backlog of the current spec",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			tasksDir, err := tasksDirOrErr()
			if err != nil {
				return err
			}
			fileName, err := taskmanage.AddTask(tasksDir, args[0], args[1], args[2])
			if err != nil {
				return err
			}
			fmt.Println(fileName)
			return nil
		},
	}
}
