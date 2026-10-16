package main

import (
	"fmt"
	"os"

	"github.com/plaesy/spec-kit/internal/common"
	"github.com/plaesy/spec-kit/internal/evals"
	"github.com/spf13/cobra"
)

func newEvalsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "evals",
		Short: "Run behavioural cases for instruction files (static, output fixtures, optional briefs)",
		Long: `Run every case file under evals/cases.

  static   the instruction text must contain required phrases and avoid forbidden ones
  outputs  fixture directories under evals/fixtures are linted with design lint
  briefs   optional: PLAESY_EVAL_RUNNER names a command that writes an output for
           each brief. It is run as "<command> <brief-id> <output-dir>" with the
           brief text in PLAESY_EVAL_BRIEF, and its output is linted. Without it,
           every brief is reported SKIP. A skip is never a pass.

The command fails if any case fails. Briefs are not a pass without a runner.`,
		Example: `  plaesy validate evals
  PLAESY_EVAL_RUNNER="./scripts/my-llm-runner" plaesy validate evals`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := common.GetRepoRoot()
			if err != nil {
				root = "."
			}
			outcomes, err := evals.Run(root, evals.Options{Runner: os.Getenv("PLAESY_EVAL_RUNNER")})
			if err != nil {
				fmt.Println("[✗]", err)
				return err
			}
			for _, o := range outcomes {
				fmt.Printf("%-4s %-26s %-7s %-34s %s\n", o.Status, o.File, o.Kind, o.ID, o.Detail)
			}
			pass, fail, skip := evals.Summary(outcomes)
			fmt.Printf("\n%d pass, %d fail, %d skip\n", pass, fail, skip)
			if fail > 0 {
				fmt.Println("[✗] evals failed")
				return fmt.Errorf("%d eval case(s) failed", fail)
			}
			fmt.Println("[OK] evals passed")
			return nil
		},
	}
	return cmd
}
