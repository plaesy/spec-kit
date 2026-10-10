package main

import (
	"fmt"

	"github.com/plaesy/spec-kit/internal/taskmanage"
	"github.com/spf13/cobra"
)

// `plaesy tasks` — the task lifecycle under .plaesy/tasks/.
//
// It was `plaesy task-manage`, which spells the resource as `task` and then
// repeats the idea in `-manage`: `tasks list` already says everything
// `task-manage list` does, one word shorter. The convention this tree is
// settling on is resource plural then verb, so the parent is the resource and
// the nine subcommands are the verbs. The old spelling was removed, not
// aliased; see features.go for why a leftover alias would be invisible.
func init() { register(newTasksCmd()) }

func newTasksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tasks",
		Short: "Task lifecycle under .plaesy/tasks/ (backlog/todo/doing/done/blocked)",
		Long: `Work with the tasks under .plaesy/tasks/.

  plaesy tasks list                 every task, across every status
  plaesy tasks list-status <status> tasks in one status
  plaesy tasks next                 the next todo task
  plaesy tasks start <file>         move a task to doing
  plaesy tasks complete <file>      move a task to done
  plaesy tasks block <file> [why]   move a task to blocked
  plaesy tasks unblock <file>       move a task back to todo
  plaesy tasks move <file> <from> <to>
  plaesy tasks show <file> [status]

With no subcommand, prints this list.`,
		Example: `  plaesy tasks list
  plaesy tasks next
  plaesy tasks complete 003-add-dark-mode-tasks.md`,
		Args: cobra.NoArgs,
	}

	cmd.AddCommand(
		newTasksListCmd(),
		newTasksListStatusCmd(),
		newTasksNextCmd(),
		newTasksStartCmd(),
		newTasksCompleteCmd(),
		newTasksBlockCmd(),
		newTasksUnblockCmd(),
		newTasksMoveCmd(),
		newTasksShowCmd(),
	)

	return showHelpWhenBare(cmd)
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

func newTasksListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all tasks across every status",
		Args:  cobra.NoArgs,
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

func newTasksListStatusCmd() *cobra.Command {
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

func newTasksNextCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "next",
		Short: "Pop the next task from backlog into todo",
		Args:  cobra.NoArgs,
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
			// One labeled line, like every sibling move. This used to print the
			// bare file name on a second line as well, which read as two
			// different things having happened.
			fmt.Printf("[✓] Moved: %s (backlog → todo)\n", taskFile)
			return nil
		},
	}
}

func newTasksStartCmd() *cobra.Command {
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

func newTasksCompleteCmd() *cobra.Command {
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

func newTasksBlockCmd() *cobra.Command {
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

func newTasksUnblockCmd() *cobra.Command {
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

func newTasksMoveCmd() *cobra.Command {
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

func newTasksShowCmd() *cobra.Command {
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
