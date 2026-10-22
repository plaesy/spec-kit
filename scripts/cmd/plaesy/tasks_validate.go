package main

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/plaesy/spec-kit/internal/checklist"
	"github.com/plaesy/spec-kit/internal/tasks"
	"github.com/spf13/cobra"
)

// newTasksValidateCmd runs the checklist of every task, or of one task folder. Each
// item's run command is executed, its result is written back to contracts.json, and a task
// whose blocking items are all met is moved to done.
func newTasksValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate [task]",
		Short: "Run each task's checklist, record the results, and move finished tasks to done",
		Long: `Run the contracts.json checklist of each task folder under
.plaesy/specs/<spec>/tasks/<status>/<task>/.

Every item's run command is run from the project root. The result is written back into
contracts.json. When every blocking item is met, the task folder moves to done.

With no argument, every task is checked. With a task folder name, only that task.
--spec limits the search to one spec.

  plaesy tasks validate
  plaesy tasks validate implement-contract-record`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := tasks.FindProjectRoot()
			if err != nil {
				return err
			}
			tasksDir := tasks.TasksDir(root)
			refs, err := tasks.ListContracts(tasksDir, tasksSpec)
			if err != nil {
				return err
			}
			checked, failed := 0, 0
			for _, ref := range refs {
				if len(args) == 1 && ref.Task != args[0] {
					continue
				}
				checked++
				if err := validateTaskChecklist(root, tasksDir, ref); err != nil {
					failed++
					fmt.Printf("  [!] %s/%s: %v\n", ref.Spec, ref.Task, err)
				}
			}
			if checked == 0 {
				fmt.Println("[!] No task with a contracts.json was found")
				return nil
			}
			fmt.Printf("\n[INFO] %d task(s) checked, %d could not be run\n", checked, failed)
			if failed > 0 {
				return fmt.Errorf("%d task checklist(s) could not be run", failed)
			}
			return nil
		},
	}
}

// validateTaskChecklist runs one task's checklist, saves the results, and moves the task
// to done when it is finished. A task that is not finished stays where it is.
func validateTaskChecklist(root, tasksDir string, ref tasks.ContractRef) error {
	f, err := checklist.Load(ref.Path)
	if err != nil {
		return err
	}
	checklist.Run(&f, func(run string) (int, string, error) { return runCommand(root, run) }, time.Now())
	if err := checklist.Save(ref.Path, f); err != nil {
		return err
	}
	passed, total := checklist.Progress(f)
	fmt.Printf("%s/%s (%s): %d/%d checks passed\n", ref.Spec, ref.Task, ref.Status, passed, total)
	if checklist.Done(f) && ref.Status != "done" {
		if err := tasks.MoveTask(tasksDir, ref.Task, ref.Status, "done"); err != nil {
			return err
		}
		fmt.Printf("  moved to done\n")
	}
	return nil
}

// runCommand runs one command through bash from the project root and returns its exit
// code and standard output. A command that cannot start is an error, not a failed item.
func runCommand(root, run string) (int, string, error) {
	cmd := exec.Command("bash", "-c", run)
	cmd.Dir = root
	out, err := cmd.Output()
	if err == nil {
		return 0, string(out), nil
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return ee.ExitCode(), string(out), nil
	}
	return 0, "", err
}
