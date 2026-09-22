package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/plaesy/spec-kit/internal/featurepath"
	"github.com/spf13/cobra"
)

func init() { register(newCreateNewFeatureCmd()) }

func newCreateNewFeatureCmd() *cobra.Command {
	var jsonMode bool

	cmd := &cobra.Command{
		Use:   "create-new-feature <feature description>",
		Short: "Create a new feature branch, directory structure and spec.md from template",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := featurepath.CreateNewFeature(strings.Join(args, " "))
			if err != nil {
				return err
			}

			if jsonMode {
				out, err := json.Marshal(map[string]string{
					"BRANCH_NAME": result.BranchName,
					"SPEC_FILE":   result.SpecFile,
					"FEATURE_NUM": result.FeatureNum,
				})
				if err != nil {
					return err
				}
				fmt.Println(string(out))
				return nil
			}

			fmt.Printf("BRANCH_NAME: %s\n", result.BranchName)
			fmt.Printf("SPEC_FILE: %s\n", result.SpecFile)
			fmt.Printf("FEATURE_NUM: %s\n", result.FeatureNum)
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonMode, "json", false, "output as JSON")
	return cmd
}
