package main

import (
	"fmt"

	"github.com/plaesy/spec-kit/internal/common"
	"github.com/plaesy/spec-kit/internal/tasks"
	"github.com/spf13/cobra"
)

// `plaesy tasks` — the task lifecycle under .plaesy/specs/<spec>/tasks/.
// Every spec owns its tasks; --spec picks one, the current branch by default.
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
		Short: "Task lifecycle under .plaesy/specs/<spec>/tasks/ (backlog/todo/doing/done/blocked)",
		Long: `Work with the tasks of one spec, under .plaesy/specs/<spec>/tasks/.

The spec is the one whose folder name ends in the current git branch. Use
--spec <folder> to act on another spec.

  plaesy tasks create <priority> <slug> <title>  create a task in the backlog
  plaesy tasks list                 every task, across every status
  plaesy tasks list-status <status> tasks in one status
  plaesy tasks next                 the next todo task
  plaesy tasks start <file>         move a task to doing
  plaesy tasks complete <file>      move a task to done
  plaesy tasks block <file> [why]   move a task to blocked
  plaesy tasks unblock <file>       move a task back to todo
  plaesy tasks move <file> <from> <to>
  plaesy tasks show <file> [status]
  plaesy tasks run --chain <name>   hand todo tasks to a delegate provider

With no subcommand, prints this list.`,
		Example: `  plaesy tasks create high login-form "Build the login form"
  plaesy tasks list
  plaesy tasks next
  plaesy tasks --spec 20261015-143022-auth-login complete high_login-form`,
		Args: cobra.NoArgs,
	}

	cmd.PersistentFlags().StringVar(&tasksSpec, "spec", "", "spec to act on (default: the current git branch, which is the spec's name)")

	cmd.AddCommand(
		newTasksCreateCmd(),
		newTasksListCmd(),
		newTasksListStatusCmd(),
		newTasksNextCmd(),
		newTasksStartCmd(),
		newTasksCompleteCmd(),
		newTasksBlockCmd(),
		newTasksUnblockCmd(),
		newTasksMoveCmd(),
		newTasksShowCmd(),
		newTasksRunCmd(),
		newTasksValidateCmd(),
	)

	return showHelpWhenBare(cmd)
}

// tasksSpec is the --spec flag. Empty means "the spec for the current branch".
var tasksSpec string

// tasksDirOrErr returns the project's task directory, .plaesy/tasks.
// --spec names it directly; otherwise the current git branch selects it.
func tasksDirOrErr() (string, error) {
	root, err := tasks.FindProjectRoot()
	if err != nil {
		return "", err
	}
	return tasks.TasksDir(root), nil
}

// currentSpec is the spec the command acts on: --spec, or else the current git branch,
// which is the spec's name.
func currentSpec() string {
	if tasksSpec != "" {
		return tasksSpec
	}
	branch, _ := common.GetCurrentBranch()
	return branch
}

func printTaskList(tasksDir, status string) error {
	names, found, err := tasks.ListTasks(tasksDir, status)
	if err != nil {
		return err
	}
	if !found {
		fmt.Printf("[!] No tasks in %s\n", status)
		return nil
	}
	fmt.Printf("%s (%d)\n", status, len(names))
	for _, t := range names {
		fmt.Printf("  - %s\n", t)
	}
	return nil
}

func newTasksListCmd() *cobra.Command {
	var all, syncContext bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all tasks across every status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if syncContext {
				return syncInFlight()
			}
			if all {
				return printTaskBoard()
			}
			tasksDir, err := tasksDirOrErr()
			if err != nil {
				return err
			}
			fmt.Println("📋 Task Summary")
			fmt.Println()
			for _, status := range tasks.Statuses {
				if err := printTaskList(tasksDir, status); err != nil {
					return err
				}
				fmt.Println()
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "show every spec's tasks as one board, with title and phase, and report integrity problems")
	cmd.Flags().BoolVar(&syncContext, "sync-context", false, "rewrite the In-Flight Tasks section of .plaesy/context.md from the task files, without an AI step")
	return cmd
}

// syncInFlight rewrites the In-Flight Tasks section of .plaesy/context.md. Only that
// section changes; the rest of the file is kept byte for byte.
func syncInFlight() error {
	root, err := tasks.FindProjectRoot()
	if err != nil {
		return err
	}
	refs, err := tasks.ListAllTasks(tasks.TasksDir(root))
	if err != nil {
		return err
	}
	changed, err := tasks.SyncInFlight(root, refs)
	if err != nil {
		return err
	}
	if !changed {
		fmt.Println("[✓] In-Flight Tasks already up to date")
		return nil
	}
	fmt.Println("[✓] In-Flight Tasks written to .plaesy/context.md")
	return nil
}

// checkTaskIntegrity is the `validate` check for tasks. It prints each problem and
// fails when there is at least one, so the bare `plaesy validate` run reports it.
func checkTaskIntegrity(tasksDir string) error {
	problems, err := tasks.CheckIntegrity(tasksDir)
	if err != nil {
		return err
	}
	for _, p := range problems {
		fmt.Printf("[ERROR] %s: %s\n", p.Task, p.Message)
	}
	if len(problems) > 0 {
		return fmt.Errorf("%d task integrity problem(s)", len(problems))
	}
	fmt.Println("[OK] tasks: status matches folder, no duplicates")
	return nil
}

// printTaskBoard prints every task of every spec, then the integrity problems
// found across them. It reads the whole .plaesy/specs tree, not one spec.
func printTaskBoard() error {
	root, err := tasks.FindProjectRoot()
	if err != nil {
		return err
	}
	tasksDir := tasks.TasksDir(root)
	refs, err := tasks.ListAllTasks(tasksDir)
	if err != nil {
		return err
	}
	fmt.Printf("📋 Task Board (%d tasks)\n\n", len(refs))
	currentSpec := ""
	for _, r := range refs {
		if r.Spec != currentSpec {
			currentSpec = r.Spec
			fmt.Printf("%s\n", r.Spec)
		}
		fmt.Printf("  %-6s %-40s %-10s %s\n", r.Status, r.Name, r.Phase, r.Title)
	}

	problems, err := tasks.CheckIntegrity(tasksDir)
	if err != nil {
		return err
	}
	fmt.Println()
	if len(problems) == 0 {
		fmt.Println("[✓] No integrity problems")
		return nil
	}
	fmt.Printf("[!] %d integrity problem(s)\n", len(problems))
	for _, p := range problems {
		fmt.Printf("  - %s: %s\n", p.Task, p.Message)
	}
	return nil
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
			taskFile, found, err := tasks.PopNextTask(tasksDir, currentSpec())
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
		Use:   "start <task>",
		Short: "Move a task from todo to doing",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tasksDir, err := tasksDirOrErr()
			if err != nil {
				return err
			}
			if err := tasks.TaskStart(tasksDir, args[0]); err != nil {
				return err
			}
			fmt.Printf("[✓] Moved: %s (todo → doing)\n", args[0])
			return nil
		},
	}
}

func newTasksCompleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "complete <task>",
		Short: "Move a task from doing to done",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tasksDir, err := tasksDirOrErr()
			if err != nil {
				return err
			}
			if err := tasks.TaskComplete(tasksDir, args[0]); err != nil {
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
		Use:   "block <task> [reason]",
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
			if err := tasks.TaskBlock(tasksDir, args[0], reason); err != nil {
				return err
			}
			fmt.Printf("[✓] Moved: %s (doing → blocked)\n", args[0])
			return nil
		},
	}
}

func newTasksUnblockCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unblock <task>",
		Short: "Move a task from blocked back to todo",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tasksDir, err := tasksDirOrErr()
			if err != nil {
				return err
			}
			if err := tasks.TaskUnblock(tasksDir, args[0]); err != nil {
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
		Use:   "move <task> <from_status> <to_status>",
		Short: "Move a task file between arbitrary status directories",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			tasksDir, err := tasksDirOrErr()
			if err != nil {
				return err
			}
			if err := tasks.MoveTask(tasksDir, args[0], args[1], args[2]); err != nil {
				return err
			}
			fmt.Printf("[✓] Moved: %s (%s → %s)\n", args[0], args[1], args[2])
			return nil
		},
	}
}

func newTasksShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <task> [status]",
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
			content, resolvedStatus, err := tasks.ReadTask(tasksDir, args[0], status)
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
