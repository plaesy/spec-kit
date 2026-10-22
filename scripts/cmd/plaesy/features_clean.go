package main

import (
	"fmt"

	"github.com/plaesy/spec-kit/internal/featurepath"
	"github.com/spf13/cobra"
)

// newFeaturesCleanCmd lists the feature worktrees of this project and removes
// the merged ones with --apply. Without --apply it changes nothing.
func newFeaturesCleanCmd() *cobra.Command {
	var apply bool

	cmd := &cobra.Command{
		Use:   "clean",
		Short: "List feature worktrees; remove merged ones with --apply",
		Long: `List the feature worktrees of this project. A worktree is removed only when
its branch is merged into main (or master) and --apply is given.

Removal never forces: a worktree with uncommitted changes is kept and the
reason is printed. Branches are not deleted.

Squash-merged branches are not ancestors of main, so they show as not merged
and are kept. Remove them with "git worktree remove" once you have checked them.`,
		Example: `  plaesy features clean
  plaesy features clean --apply`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			statuses, err := featurepath.CleanWorktrees(apply)
			if err != nil {
				return err
			}
			if len(statuses) == 0 {
				fmt.Println("no feature worktrees for this project")
				return nil
			}
			for _, st := range statuses {
				state := "keep"
				switch {
				case st.Removed:
					state = "removed"
				case st.Merged && !apply:
					state = "would remove"
				}
				fmt.Printf("%-13s %s (%s)", state, st.Path, st.Branch)
				if st.Reason != "" {
					fmt.Printf(" — %s", st.Reason)
				}
				fmt.Println()
			}
			if !apply {
				fmt.Println("dry run: pass --apply to remove the merged worktrees")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&apply, "apply", false, "remove merged feature worktrees")
	return cmd
}
