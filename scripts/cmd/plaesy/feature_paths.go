package main

import (
	"fmt"

	"github.com/plaesy/spec-kit/internal/featurepath"
	"github.com/spf13/cobra"
)

func init() { register(newGetFeaturePathsCmd()) }

func newGetFeaturePathsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get-feature-paths",
		Short: "Print paths for the current feature branch without creating anything",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Print(featurepath.GetPathsReport())
			return nil
		},
	}
}
