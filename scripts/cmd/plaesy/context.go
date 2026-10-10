package main

import (
	"github.com/plaesy/spec-kit/internal/agentcontext"
	"github.com/spf13/cobra"
)

// `plaesy context update` — the agent context files (CLAUDE.md, AGENTS.md,
// .github/copilot-instructions.md, …) that a coding agent reads, synced with
// the current feature's plan.md.
//
// It was `plaesy update-agent-context`: a three-word bare verb with no
// resource in it, which is the one shape this tree is being moved off. The
// resource is `context` and the action is `update`, so it is now a two-word
// subcommand under a one-word parent — the same shape as `features create` and
// `images create`. The old spelling was removed, not aliased; see features.go
// for why a leftover alias would be invisible.
func init() { register(newContextCmd()) }

func newContextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Agent context files that plaesy keeps in sync with plan.md",
		Long: `Work with the agent context files.

  plaesy context update [agent]   sync an existing context file with plan.md

With no subcommand, prints this list.`,
		Args: cobra.NoArgs,
	}
	cmd.AddCommand(newContextUpdateCmd())
	return showHelpWhenBare(cmd)
}

func newContextUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update [claude|gemini|copilot|cursor|qwen|opencode]",
		Short: "Sync agent context files with the current feature's plan.md",
		Long: `Sync an EXISTING agent context file with the current feature's plan.md.

This never creates a context file. The agent-file template was removed on
2026-09-25, so scaffolding one is your call — copy an existing CLAUDE.md /
AGENTS.md, or let your AI platform create its own. A missing target is an error
rather than a silent skip.

A file with none of the tracked sections ("## Active Technologies",
"## Recent Changes", "Last updated: YYYY-MM-DD") is also an error, because a
run that changes nothing must not report success. A file already in sync says
so and exits 0.

With no argument, every context file that already exists is updated.`,
		Example: `  plaesy context update
  plaesy context update claude`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			agentType := ""
			if len(args) == 1 {
				agentType = args[0]
			}
			return agentcontext.Update(agentType)
		},
	}
}
