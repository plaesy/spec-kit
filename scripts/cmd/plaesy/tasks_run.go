package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/plaesy/spec-kit/internal/config"
	"github.com/plaesy/spec-kit/internal/delegate"
	"github.com/plaesy/spec-kit/internal/tasks"
	"github.com/spf13/cobra"
)

// newTasksRunCmd hands claimed tasks of one spec to a delegate provider or
// chain, one after another, running in the spec's project root (its worktree
// when the spec has one).
//
// A run never marks a task done. Its output goes to <spec>/runs/, and the task
// stays in doing until someone reviews it and runs `tasks complete`. A failed
// run moves the task to blocked with the reason.
func newTasksRunCmd() *cobra.Command {
	var provider, chainName string
	var maxTasks, parallel int
	var timeout time.Duration
	var confirm bool

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Hand todo tasks of the current spec to a delegate provider or chain",
		Long: `Claim todo tasks of the spec (highest priority first) and send each to a
delegate provider or chain from .plaesy/config.json, running in the project root.

  plaesy tasks run --chain review
  plaesy tasks run --provider claude-cli --max 3 --parallel 2

Each claim is an atomic move todo -> doing, so several runs can share a spec.
The output of a run is saved under <spec>/runs/. The task stays in doing for
review. Review it, then run "plaesy tasks complete <file>".

A run that fails moves the task to blocked with the reason. Use
"plaesy tasks unblock <file>" to try it again.

Configuration is read before any task is claimed, so a config error leaves
the tasks where they are.`,
		Example: `  plaesy tasks run --chain review
  plaesy tasks run --provider claude-cli --max 2 --timeout 15m`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if provider == "" && chainName == "" {
				return fmt.Errorf("exactly one of --provider or --chain is required (see `plaesy delegate --list`)")
			}
			tasksDir, err := tasksDirOrErr()
			if err != nil {
				return err
			}
			root, err := tasks.FindProjectRoot()
			if err != nil {
				return err
			}

			cfg, err := loadDelegateConfig(root)
			if err != nil {
				return err
			}
			names, err := chainNames(cfg, provider, chainName)
			if err != nil {
				return err
			}

			if parallel < 1 {
				parallel = 1
			}
			// Each worker claims, runs, and records one task at a time. Claims
			// are exclusive across workers and across processes, so the pool
			// never runs one task twice.
			var (
				started  atomic.Int64
				ran      atomic.Int64
				mu       sync.Mutex
				firstErr error
				wg       sync.WaitGroup
			)
			for w := 0; w < parallel; w++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for {
						if started.Add(1) > int64(maxTasks) {
							return
						}
						file, found, err := tasks.ClaimTask(tasksDir, currentSpec())
						if err == nil && !found {
							return
						}
						if err == nil {
							err = runClaimedTask(cmd.Context(), tasksDir, root, file, cfg, names, chainName, confirm, timeout)
							ran.Add(1)
						}
						if err != nil {
							mu.Lock()
							if firstErr == nil {
								firstErr = err
							}
							mu.Unlock()
							return
						}
					}
				}()
			}
			wg.Wait()
			if firstErr == nil && ran.Load() == 0 {
				fmt.Println("no todo or backlog tasks in this spec")
			}
			return firstErr
		},
	}

	cmd.Flags().StringVar(&provider, "provider", "", "provider name from delegate.providers")
	cmd.Flags().StringVar(&chainName, "chain", "", "chain name from delegate.chains")
	cmd.Flags().IntVar(&maxTasks, "max", 1, "stop after this many tasks")
	cmd.Flags().IntVar(&parallel, "parallel", 1, "run this many tasks at once, each in its own goroutine")
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Minute, "stop each call after this long")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "allow a billable provider to run")
	return cmd
}

// runClaimedTask sends one claimed task to the chain, saves the run, and moves
// the task on. The returned error is an I/O failure, not a failed run: a failed
// run is recorded on the task and does not stop the batch.
func runClaimedTask(parent context.Context, tasksDir, root, file string, cfg *config.DelegateConfig, names []string, chainName string, confirm bool, timeout time.Duration) error {
	specDir := filepath.Join(tasks.SpecsDir(root), tasks.SpecOf(tasksDir, "doing", file))
	taskDir := filepath.Join(tasksDir, "doing", file)
	taskBody, err := os.ReadFile(filepath.Join(taskDir, "spec.md"))
	if err != nil {
		return err
	}
	specBody, _ := os.ReadFile(filepath.Join(specDir, "spec.md")) // optional context

	prompt := buildTaskPrompt(filepath.Base(specDir), file, string(taskBody), string(specBody))

	usagePath := delegate.UsagePath(filepath.Join(root, analysisRel))
	recs, err := delegate.ReadUsage(usagePath)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	resp, attempts, runErr := delegate.RunChain(ctx, cfg, names, delegate.Request{
		Prompt:  prompt,
		Timeout: timeout,
		Dir:     root,
	}, delegate.ChainOptions{
		Chain:   chainName,
		Confirm: confirm,
		Build:   delegate.EnvBuild(),
		Check: func(entry config.DelegateProviderConfig, req delegate.Request) error {
			return delegate.CheckBudget(entry, req, recs, time.Now())
		},
	})

	runPath, err := writeRunRecord(taskDir, file, resp, attempts, runErr)
	if err != nil {
		return err
	}

	if runErr != nil {
		reason := fmt.Sprintf("delegate run failed: %v (see %s)", runErr, runPath)
		if err := tasks.TaskBlock(tasksDir, file, reason); err != nil {
			return err
		}
		fmt.Printf("blocked   %s — %v\n", file, runErr)
		return nil
	}

	if err := delegate.AppendUsage(usagePath, usageRecord(cfg, resp, chainName)); err != nil {
		return err
	}
	fmt.Printf("ran       %s via %s — review %s, then: plaesy tasks complete %s\n", file, resp.Provider, runPath, file)
	return nil
}

// buildTaskPrompt is what the provider receives: the task, the spec for
// context, and the instruction to report what changed.
func buildTaskPrompt(specName, taskFile, taskBody, specBody string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "You are implementing one task of the feature spec %q.\n", specName)
	b.WriteString("Work in the current directory and change only what the task needs.\n")
	b.WriteString("When you finish, reply with a short summary: what changed, and how each acceptance criterion was checked.\n\n")
	fmt.Fprintf(&b, "## Task (%s)\n\n%s\n\n", taskFile, taskBody)
	if specBody != "" {
		fmt.Fprintf(&b, "## Feature spec\n\n%s\n", specBody)
	}
	return b.String()
}

// writeRunRecord saves a run's output, or its error, under <spec>/runs/.
// The file name carries the time, so runs of one task do not overwrite each other.
func writeRunRecord(taskDir, file string, resp delegate.Response, attempts []delegate.Attempt, runErr error) (string, error) {
	dir := filepath.Join(taskDir, "runs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	stem := file
	path := filepath.Join(dir, fmt.Sprintf("%s-%s.md", stem, time.Now().UTC().Format("150405")))

	var b strings.Builder
	fmt.Fprintf(&b, "# Run of %s\n\n", file)
	fmt.Fprintf(&b, "- Time: %s\n", time.Now().UTC().Format(time.RFC3339))
	if resp.Provider != "" {
		fmt.Fprintf(&b, "- Provider: %s\n", resp.Provider)
	}
	if runErr != nil {
		fmt.Fprintf(&b, "- Result: failed\n\n## Error\n\n%v\n", runErr)
	} else {
		fmt.Fprintf(&b, "- Result: ok — not yet reviewed\n\n## Output\n\n%s\n", resp.Text)
	}
	if len(attempts) > 0 {
		fmt.Fprintf(&b, "\n## Attempts\n\n%d\n", len(attempts))
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return "", err
	}
	return path, nil
}
