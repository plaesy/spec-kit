package main

import (
	"fmt"

	"github.com/plaesy/spec-kit/internal/contract"
	"github.com/spf13/cobra"
)

func newContractCmd() *cobra.Command {
	var gate bool
	cmd := &cobra.Command{
		Use:   "contract <file>",
		Short: "Validate a Success Contract; with --gate, check the recorded results",
		Long: `Validate one Success Contract file (instructions/success-contract.instructions.md).

Without --gate: the schema, the measurability rules (numeric or binary targets, no vague
words, a real method), and the freeze. Every change to a target after freeze must have a
recorded amendment with a reason. Blocking items cannot be loosened.

With --gate: every item must have a recorded result (met, not_met or not_measurable).
A met item needs its actual value and evidence, and for a numeric target the validator
checks that the actual value satisfies it. A blocking item that is not met, or not
measurable, fails the gate: the phase must not report completion.

Exit status is non-zero when any error is found. Warnings are printed but do not fail.
The contract is a markdown file with one fenced json block; the rest is prose.`,
		Example: `  plaesy validate contract .plaesy/contracts/fix-empty-slug.contract.md
  plaesy validate contract .plaesy/contracts/fix-empty-slug.contract.md --gate`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rep, err := contract.Check(args[0], gate)
			if err != nil {
				fmt.Println("[✗]", err)
				return err
			}
			for _, i := range rep.Issues {
				loc := i.Item
				if loc == "" {
					loc = "(contract)"
				}
				fmt.Printf("%-7s %-12s %s\n", i.Severity, loc, i.Message)
			}
			mode := "schema and freeze"
			if gate {
				mode = "gate"
			}
			fmt.Printf("\n[INFO] %s check of %s: %d line(s), %d error(s), %d warning(s)\n", mode, args[0], rep.Lines, rep.Errors(), rep.Warnings())
			if rep.Errors() > 0 {
				fmt.Println("[✗] contract failed")
				return fmt.Errorf("contract %s has %d error(s)", args[0], rep.Errors())
			}
			fmt.Println("[OK] contract passed")
			return nil
		},
	}
	cmd.Flags().BoolVar(&gate, "gate", false, "check the recorded results as the final gate")
	return cmd
}
