package main

import (
	"github.com/plaesy/spec-kit/internal/agentcontext"
	"github.com/spf13/cobra"
)

func init() { register(newUpdateAgentContextCmd()) }

func newUpdateAgentContextCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update-agent-context [claude|gemini|copilot|cursor|qwen|opencode]",
		Short: "Sync agent context files with the current feature's plan.md",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			agentType := ""
			if len(args) == 1 {
				agentType = args[0]
			}
			return agentcontext.Update(agentType)
		},
	}
}
