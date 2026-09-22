package main

import "github.com/spf13/cobra"

// registry collects subcommands via init() in each command's own file, so
// parallel additions never conflict on a shared AddCommand() call site.
var registry []*cobra.Command

func register(cmd *cobra.Command) {
	registry = append(registry, cmd)
}
