package main

import (
	"fmt"

	"github.com/plaesy/spec-kit/internal/featurepath"
	"github.com/spf13/cobra"
)

func newFeaturesPathsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "paths",
		Short: "Print paths for the current feature branch without creating anything",
		// Takes no arguments; without this a typo is silently accepted and the
		// user sees a successful report of paths they did not ask for.
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Print(featurepath.GetPathsReport())
			return nil
		},
	}
}
