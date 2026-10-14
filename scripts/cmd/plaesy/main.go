// Command plaesy is the single cross-platform binary that replaces the
// scripts/bash and scripts/powershell dispatch scripts.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:           "plaesy",
		Short:         "Plaesy Constitution Kit CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(registry...)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
