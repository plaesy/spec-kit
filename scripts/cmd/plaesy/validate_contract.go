package main

import (
	"fmt"

	"github.com/plaesy/spec-kit/internal/contract"
	"github.com/spf13/cobra"
)

func newContractCmd() *cobra.Command {
	var gate, requireGit bool
	cmd := &cobra.Command{
		Use:   "contract <file>",
		Short: "Validate a Success Contract; with --gate, check the recorded results",
		Long: `Validate one Success Contract file (instructions/success-contract.instructions.md).

Without --gate: the schema, measurable targets (a comparison with one number and an optional
unit, or a binary condition), no vague words and no template placeholders, a real method, and
the freeze. Every change to a target after freeze needs a recorded amendment with a reason; a
blocking item cannot be loosened or removed; at least one item must be blocking unless a
no-blocking waiver is recorded.

With --gate: every item needs a recorded result (met, not_met or not_measurable). A met item
needs an actual value and evidence: a repo-relative path that exists (optionally :line), or
cmd:<the command run>. A numeric claim is checked against its target, unit included; a binary
claim is checked against a fixed vocabulary. A blocking item that is not met, or not measurable,
fails the gate. A contract whose items are all non-blocking fails the gate if one is not met.

The freeze is checked against git history when the contract is committed: the frozen
snapshot must match the first committed version, that version must hold no results, and the
amendment log must only grow. Outside git, or before the contract is committed, the freeze is
reported as NOT VERIFIED; --require-git makes that an error.

Exit status is non-zero when any error is found. Warnings are printed but do not fail.
The contract is a markdown file with exactly one fenced json block; the rest is prose.`,
		Example: `  plaesy validate contract .plaesy/contracts/fix-empty-slug.contract.md
  plaesy validate contract .plaesy/contracts/fix-empty-slug.contract.md --gate --require-git`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rep, err := contract.Check(args[0], contract.Options{Gate: gate, RequireGit: requireGit})
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
			if rep.Anchor.Verified {
				fmt.Printf("[INFO] freeze verified against git: %s\n", rep.Anchor.Note)
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
	cmd.Flags().BoolVar(&requireGit, "require-git", false, "treat a freeze that cannot be verified against git as an error")
	return cmd
}
