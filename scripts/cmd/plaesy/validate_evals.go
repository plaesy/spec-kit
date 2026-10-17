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
		Short: "Run the instruction case files: phrase checks, lint fixtures, and optional model briefs",
		Long: `Run every case file under evals/cases.

  static   the instruction text must contain required phrases and avoid forbidden
           ones. This checks that wording is present. It does not check meaning.
  outputs  hand-written fixture directories under evals/fixtures are linted with
           design lint and checked against the expected result. This tests the
           lint on known inputs. It is not model output.
  briefs   the only case that tests model behaviour. It needs PLAESY_EVAL_RUNNER,
           a command run as "<command> <brief-id> <output-dir>" with the brief in
           PLAESY_EVAL_BRIEF. It must write at least one lintable file, with no FAIL
           findings. PLAESY_EVAL_RUNNER is a JSON array of strings, or a command line
           where "quoted groups" are one argument; backslashes are kept, so Windows
           paths work.

Without PLAESY_EVAL_RUNNER every brief is reported SKIP, and the summary says that
no model behaviour was checked. A skip is never a pass. The command exits 0 when
nothing failed, including when briefs were skipped.`,
		Example: `  plaesy validate evals
  PLAESY_EVAL_RUNNER='"C:\tools\my runner.exe"' plaesy validate evals
  PLAESY_EVAL_RUNNER='["C:\\tools\\my runner.exe", "--fast"]' plaesy validate evals`,
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
			if skip > 0 {
				fmt.Printf("[SKIP] %d brief case(s) not run: PLAESY_EVAL_RUNNER is not set. No model behaviour was checked.\n", skip)
			}
			if fail > 0 {
				fmt.Println("[✗] evals failed")
				return fmt.Errorf("%d eval case(s) failed", fail)
			}
			fmt.Println("[OK] evals passed (static and fixture checks)")
			return nil
		},
	}
	return cmd
}
