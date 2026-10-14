package main

import (
	"fmt"

	"github.com/plaesy/spec-kit/internal/featurepath"
	"github.com/spf13/cobra"
)

// `plaesy features` is the single entry point for the feature-branch workflow:
// creating one, listing them, printing the current one's paths, and checking it
// is ready for task generation.
//
// Those were three top-level commands with three different shapes
// (`create-new-feature`, `get-feature-paths`, `check-task-prerequisites`), none
// of which told a reader they belonged together. Consolidating them under one
// resource noun is the same move `plaesy validate <target>` made at
// constitution 1.3.0, and the naming follows the convention the rest of this
// tree is being moved onto: **resource plural, then verb** — `features create`,
// `tasks start`, `images create`, `context update`.
//
// The old spellings were removed rather than aliased. That is deliberate and
// different from the `validate-*` collapse, which kept deprecated aliases for a
// documented window: these had no deprecation window, and a leftover alias
// would be invisible — deprecated commands are hidden from help and from the
// documentation bar the drift tests enforce, so nothing would notice it still
// working.
func init() { register(newFeaturesCmd()) }

func newFeaturesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "features",
		Short: "List, create, and inspect feature branches",
		Long: `Work with feature branches.

  plaesy features                       list every feature in .plaesy/specs/
  plaesy features create <description>  branch, directories, and spec.md
  plaesy features paths                 paths for the current feature
  plaesy features validate              is the current feature ready for tasks

With no subcommand, lists the features that exist.`,
		Example: `  plaesy features
  plaesy features create "Add dark mode toggle"
  plaesy features paths
  plaesy features validate --json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			features, err := featurepath.ListFeatures()
			if err != nil {
				return err
			}
			fmt.Print(featurepath.ListReport(features))
			return nil
		},
	}
	cmd.AddCommand(
		newFeaturesCreateCmd(),
		newFeaturesPathsCmd(),
		newFeaturesValidateCmd(),
	)
	return cmd
}
