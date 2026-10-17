package main

import (
	"fmt"
	"path/filepath"

	"github.com/plaesy/spec-kit/internal/common"
	"github.com/plaesy/spec-kit/internal/contract"
	"github.com/spf13/cobra"
)

func newContractsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "contracts",
		Short: "Check every Success Contract under .plaesy/contracts (closed ones at the gate)",
		Long: `Check every *.contract.md under .plaesy/contracts. An open contract is checked for its schema
and its freeze. A closed contract is checked at the gate, in full. With no contracts, the check is
skipped and says so. This runs as part of bare 'plaesy validate'.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := common.GetRepoRoot()
			if err != nil {
				root = "."
			}
			results, err := contract.CheckDir(root)
			if err != nil {
				fmt.Println("[✗]", err)
				return err
			}
			if len(results) == 0 {
				fmt.Printf("[SKIP] no contracts under %s\n", filepath.Join(root, filepath.FromSlash(contract.Dir)))
				return nil
			}
			failed := 0
			for _, r := range results {
				mode := "schema and freeze"
				if r.Gate {
					mode = "gate"
				}
				fmt.Printf("%-6s %s (%s, %d error(s), %d warning(s))\n", statusWord(r.Report.Errors()), r.Report.Path, mode, r.Report.Errors(), r.Report.Warnings())
				for _, i := range r.Report.Issues {
					if i.Severity == contract.SevError {
						fmt.Printf("       %s %s: %s\n", i.Severity, i.Item, i.Message)
					}
				}
				if r.Report.Errors() > 0 {
					failed++
				}
			}
			if failed > 0 {
				return fmt.Errorf("%d of %d contract(s) failed", failed, len(results))
			}
			fmt.Printf("[OK] %d contract(s) passed\n", len(results))
			return nil
		},
	}
}

func statusWord(errors int) string {
	if errors > 0 {
		return "[✗]"
	}
	return "[OK]"
}
