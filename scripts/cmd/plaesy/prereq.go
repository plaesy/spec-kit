package main

import (
	"encoding/json"
	"fmt"

	"github.com/plaesy/spec-kit/internal/common"
	"github.com/plaesy/spec-kit/internal/featurepath"
	"github.com/spf13/cobra"
)

func init() { register(newCheckTaskPrerequisitesCmd()) }

func newCheckTaskPrerequisitesCmd() *cobra.Command {
	var jsonMode bool

	cmd := &cobra.Command{
		Use:   "check-task-prerequisites",
		Short: "Validate the current feature has a plan.md and report available design docs",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := featurepath.CheckTaskPrerequisites()
			if err != nil {
				return err
			}

			if jsonMode {
				out, err := json.Marshal(result)
				if err != nil {
					return err
				}
				fmt.Println(string(out))
				return nil
			}

			fp, err := common.GetFeaturePaths()
			if err != nil {
				return err
			}
			fmt.Print(result.TextReport(fp))
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonMode, "json", false, "output as JSON")
	return cmd
}
