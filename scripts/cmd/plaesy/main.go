// Command plaesy is the single cross-platform binary that replaces the
// scripts/bash and scripts/powershell dispatch scripts.
package main

import (
	"fmt"
	"os"

	"github.com/plaesy/spec-kit/internal/common"
	"github.com/spf13/cobra"
)

// newRoot builds the root command. main and the tests share it so that the
// command tree a test sees is the one the binary runs.
func newRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "plaesy",
		Short:         "Plaesy Constitution Kit CLI",
		Version:       common.Version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(registry...)
	return root
}

func main() {
	if err := newRoot().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
