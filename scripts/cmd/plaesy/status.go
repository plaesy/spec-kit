package main

import (
	"fmt"

	"github.com/plaesy/spec-kit/internal/common"
	"github.com/spf13/cobra"
)

func init() { register(newStatusCmd()) }

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show Plaesy CLI version and environment status",
		// Takes no arguments. Without this, cobra's default ArbitraryArgs
		// silently swallows `plaesy status stray`, so a typo reads as success.
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			common.PrintBanner("Plaesy Constitution Kit", "version "+common.Version)

			if err := common.ValidateCommandExists("git", "command"); err != nil {
				fmt.Println(common.CheckFile("", "git available"))
			} else {
				fmt.Println("  ✓ git available")
			}

			if root, err := common.GetRepoRoot(); err == nil {
				fmt.Printf("  ✓ git repository: %s\n", root)
			} else {
				fmt.Println("  ✗ not inside a git repository")
			}

			return nil
		},
	}
}
