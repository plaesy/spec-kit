package main

import (
	"github.com/plaesy/spec-kit/internal/scaffold"
	"github.com/spf13/cobra"
)

func init() { register(newInitCmd()) }

func newInitCmd() *cobra.Command {
	var aiPlatform string
	var plaesyHome string

	cmd := &cobra.Command{
		Use:   "init [directory]",
		Short: "Scaffold a new Plaesy project (.plaesy/ structure + AI platform files)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := "."
			if len(args) == 1 {
				target = args[0]
			}
			return scaffold.Init(scaffold.Options{
				TargetDir:  target,
				AIPlatform: aiPlatform,
				PlaesyHome: plaesyHome,
			})
		},
	}

	cmd.Flags().StringVar(&aiPlatform, "ai", "", "AI platform to configure (e.g. claude_code, cursor_ai, github_copilot); default: none")
	cmd.Flags().StringVar(&plaesyHome, "plaesy-home", "", "Override the Plaesy repo root (also read from PLAESY_HOME env var)")

	return cmd
}
