package main

import (
	"fmt"

	"github.com/plaesy/spec-kit/internal/taskmanage"
	"github.com/spf13/cobra"
)

func init() { register(newTaskManageCmd()) }

func newTaskManageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task-manage",
		Short: "Task lifecycle management (.plaesy/tasks/ backlog/todo/doing/done/blocked)",
	}

	cmd.AddCommand(
		newTaskManageListCmd(),
		newTaskManageListStatusCmd(),
		newTaskManageNextCmd(),
		newTaskManageStartCmd(),
		newTaskManageCompleteCmd(),
		newTaskManageBlockCmd(),
		newTaskManageUnblockCmd(),
		newTaskManageMoveCmd(),
		newTaskManageShowCmd(),
	)

	return cmd
}

func tasksDirOrErr() (string, error) {
	root, err := taskmanage.FindProjectRoot()
	if err != nil {
		return "", err
	}
	return taskmanage.TasksDir(root), nil
}

func printTaskList(tasksDir, status string) error {
	tasks, found, err := taskmanage.ListTasks(tasksDir, status)
	if err != nil {
		return err
	}
	if !found {
		fmt.Printf("[!] No tasks in %s\n", status)
		return nil
	}
	fmt.Printf("%s (%d)\n", status, len(tasks))
	for _, t := range tasks {
		fmt.Printf("  - %s\n", t)
	}
	return nil
}

func newTaskManageListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all tasks across every status",
		RunE: func(cmd *cobra.Command, args []string) error {
			tasksDir, err := tasksDirOrErr()
			if err != nil {
				return err
			}
			fmt.Println("📋 Task Summary")
			fmt.Println()
			for _, status := range taskmanage.Statuses {
				if err := printTaskList(tasksDir, status); err != nil {
					return err
				}
				fmt.Println()
			}
			return nil
		},
	}
}

func newTaskManageListStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-status <backlog|todo|doing|done|blocked>",
		Short: "List tasks in one status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tasksDir, err := tasksDirOrErr()
			if err != nil {
				return err
			}
			return printTaskList(tasksDir, args[0])
		},
	}
}

func newTaskManageNextCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "next",
		Short: "Pop the next task from backlog into todo",
		RunE: func(cmd *cobra.Command, args []string) error {
			tasksDir, err := tasksDirOrErr()
			if err != nil {
				return err
			}
			taskFile, found, err := taskmanage.PopNextTask(tasksDir)
			if err != nil {
				return err
			}
			if !found {
				fmt.Println("[!] No tasks in backlog")
				return nil
			}
			fmt.Printf("[✓] Moved: %s (backlog → todo)\n", taskFile)
			fmt.Println(taskFile)
			return nil
		},
	}
}

func newTaskManageStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start <task_file>",
		Short: "Move a task from todo to doing",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tasksDir, err := tasksDirOrErr()
			if err != nil {
				return err
			}
			if err := taskmanage.TaskStart(tasksDir, args[0]); err != nil {
				return err
			}
			fmt.Printf("[✓] Moved: %s (todo → doing)\n", args[0])
			return nil
		},
	}
}

func newTaskManageCompleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "complete <task_file>",
		Short: "Move a task from doing to done",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tasksDir, err := tasksDirOrErr()
			if err != nil {
				return err
			}
			if err := taskmanage.TaskComplete(tasksDir, args[0]); err != nil {
				return err
			}
			fmt.Printf("[✓] Moved: %s (doing → done)\n", args[0])
			fmt.Printf("[✓] Task completed: %s\n", args[0])
			return nil
		},
	}
}

func newTaskManageBlockCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "block <task_file> [reason]",
		Short: "Move a task from doing to blocked, recording a reason",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			tasksDir, err := tasksDirOrErr()
			if err != nil {
				return err
			}
			reason := ""
			if len(args) > 1 {
				reason = args[1]
			}
			if err := taskmanage.TaskBlock(tasksDir, args[0], reason); err != nil {
				return err
			}
			fmt.Printf("[✓] Moved: %s (doing → blocked)\n", args[0])
			return nil
		},
	}
}

func newTaskManageUnblockCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unblock <task_file>",
		Short: "Move a task from blocked back to todo",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tasksDir, err := tasksDirOrErr()
			if err != nil {
				return err
			}
			if err := taskmanage.TaskUnblock(tasksDir, args[0]); err != nil {
				return err
			}
			fmt.Printf("[✓] Moved: %s (blocked → todo)\n", args[0])
			fmt.Printf("[✓] Task unblocked: %s\n", args[0])
			return nil
		},
	}
}

func newTaskManageMoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "move <task_file> <from_status> <to_status>",
		Short: "Move a task file between arbitrary status directories",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			tasksDir, err := tasksDirOrErr()
			if err != nil {
				return err
			}
			if err := taskmanage.MoveTask(tasksDir, args[0], args[1], args[2]); err != nil {
				return err
			}
			fmt.Printf("[✓] Moved: %s (%s → %s)\n", args[0], args[1], args[2])
			return nil
		},
	}
}

func newTaskManageShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <task_file> [status]",
		Short: "Show a task's raw content",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			tasksDir, err := tasksDirOrErr()
			if err != nil {
				return err
			}
			status := ""
			if len(args) > 1 {
				status = args[1]
			}
			content, resolvedStatus, err := taskmanage.ReadTask(tasksDir, args[0], status)
			if err != nil {
				return err
			}
			fmt.Printf("📄 Task: %s\n", args[0])
			fmt.Printf("Status: %s\n", resolvedStatus)
			fmt.Println()
			fmt.Print(content)
			return nil
		},
	}
}
