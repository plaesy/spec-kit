package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/plaesy/spec-kit/internal/featurepath"
	"github.com/spf13/cobra"
)

// Registered as a subcommand of `plaesy features`: see features.go.
func newFeaturesCreateCmd() *cobra.Command {
	var jsonMode bool
	var worktree bool

	cmd := &cobra.Command{
		Use:   "create <feature description>",
		Short: "Create a new feature branch, spec folder and spec.md from template",
		Long: `Create a feature: a branch named after the description, and a spec folder
named {TIMESTAMP}-{BRANCH} under .plaesy/specs/ with spec.md and tasks/.

By default the branch is checked out in the current working tree.
With --worktree the feature gets its own git worktree, outside the repository,
so several features can be in progress at once. Run "plaesy features clean"
to remove the worktrees of merged features.`,
		Example: `  plaesy features create "Add login form"
  plaesy features create --worktree "Add billing page"`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := featurepath.CreateNewFeatureWith(strings.Join(args, " "), featurepath.CreateOptions{Worktree: worktree})
			if err != nil {
				return err
			}

			if jsonMode {
				out, err := json.Marshal(map[string]string{
					"BRANCH_NAME":   result.BranchName,
					"SPEC_NAME":     result.SpecName,
					"SPEC_FILE":     result.SpecFile,
					"WORKTREE_PATH": result.WorktreePath,
				})
				if err != nil {
					return err
				}
				fmt.Println(string(out))
				return nil
			}

			fmt.Printf("BRANCH_NAME: %s\n", result.BranchName)
			fmt.Printf("SPEC_NAME: %s\n", result.SpecName)
			fmt.Printf("SPEC_FILE: %s\n", result.SpecFile)
			if result.WorktreePath != "" {
				fmt.Printf("WORKTREE_PATH: %s\n", result.WorktreePath)
			}
			if result.Warning != "" {
				fmt.Fprintf(os.Stderr, "WARNING: %s\n", result.Warning)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonMode, "json", false, "output as JSON")
	cmd.Flags().BoolVar(&worktree, "worktree", false, "create the feature in its own git worktree outside the repository")
	return cmd
}
