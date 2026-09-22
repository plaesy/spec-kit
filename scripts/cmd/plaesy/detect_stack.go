package main

import (
	"fmt"
	"os"

	"github.com/plaesy/spec-kit/internal/common"
	"github.com/plaesy/spec-kit/internal/detectstack"
	"github.com/spf13/cobra"
)

func init() { register(newDetectStackCmd()) }

func newDetectStackCmd() *cobra.Command {
	var plaesyRoot string

	cmd := &cobra.Command{
		Use:   "detect-stack [target-dir]",
		Short: "List instruction files relevant to a target project's tech stack",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetDir := "."
			if len(args) == 1 {
				targetDir = args[0]
			}

			root := plaesyRoot
			if root == "" {
				root = os.Getenv("PLAESY_ROOT")
			}
			if root == "" {
				if r, err := common.GetRepoRoot(); err == nil {
					root = r
				}
			}
			if root == "" {
				return fmt.Errorf("could not determine PLAESY_ROOT: not a git repository and PLAESY_ROOT is unset")
			}

			files, err := detectstack.Detect(targetDir, root)
			if err != nil {
				return err
			}
			for _, f := range files {
				fmt.Println(f)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&plaesyRoot, "plaesy-root", "", "path to the plaesy repo root (defaults to $PLAESY_ROOT, then git repo root)")
	return cmd
}
