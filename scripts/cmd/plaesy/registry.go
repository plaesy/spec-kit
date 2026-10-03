package main

import "github.com/spf13/cobra"

// registry collects subcommands via init() in each command's own file, so
// parallel additions never conflict on a shared AddCommand() call site.
var registry []*cobra.Command

func register(cmd *cobra.Command) {
	registry = append(registry, cmd)
}

// showHelpWhenBare makes a noun parent usable with no subcommand while still
// rejecting a misspelled one.
//
// The problem is cobra's ordering. A parent with no Run/RunE is not Runnable,
// and execute() returns flag.ErrHelp before it ever calls ValidateArgs — so
// `Args: cobra.NoArgs` on such a parent is dead code, and `plaesy platforms
// detectx` prints the help text and exits 0. That is the worst possible answer
// for a group: a user who mistypes a subcommand, or who types one of the
// pre-rename spellings this tree removed, gets a wall of help as if it were
// the result. Captured in a shell it becomes the value.
//
// Giving the parent a Run makes it Runnable, ValidateArgs runs, and cobra's
// own NoArgs produces `unknown command "detectx" for "plaesy platforms"`. With
// no arguments it falls through to Help, which is the documented behaviour for
// every one of these parents.
func showHelpWhenBare(cmd *cobra.Command) *cobra.Command {
	cmd.RunE = func(c *cobra.Command, args []string) error {
		return c.Help()
	}
	if cmd.Args == nil {
		cmd.Args = cobra.NoArgs
	}
	return cmd
}
