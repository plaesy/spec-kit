package main

import (
	"fmt"

	"github.com/plaesy/spec-kit/internal/common"
	"github.com/plaesy/spec-kit/internal/scaffold"
	"github.com/spf13/cobra"
)

func init() { register(newReloadCmd()) }

func newReloadCmd() *cobra.Command {
	var aiPlatform string
	var plaesyHome string
	var dryRun bool
	var prune bool
	var pruneApply bool

	cmd := &cobra.Command{
		Use:   "reload [directory]",
		Short: "Refresh generated .plaesy/ files, keeping memory and context",
		Long: `Refresh the generated parts of .plaesy/ from the current sources.

'plaesy init' never overwrites a file that already exists, so a project
initialised before an instruction or template was edited keeps the stale copy
permanently: repair and upgrade are stubs, so there was no command that would
bring it forward. This is that command.

Reload overwrites generated files and leaves everything else alone. It will not
overwrite memory.md, context.md or state.json, and it never writes specs/,
tasks/, analysis/ or memory/. A file with no counterpart in the sources is left
in place, so a hand-written template survives.

That last rule is right for files you wrote and wrong for files reload wrote.
A prompt renamed in prompts/ leaves a complete, still-invokable command
behind in the mirror, so --prune lists orphans in the tool-owned trees (the
agent roles and the platform prompt mirror) and --prune-apply removes them.
It touches nothing else: no .plaesy/ content, no memory, and never a
destination the platform shares with hand-written files. Deletion is opt-in
twice because these directories are gitignored and version control cannot undo
it.

The AI platform is not guessed -- it is not recorded anywhere on disk, so a
wrong guess would write a second platform's files into the project. Pass --ai
to refresh the platform core file, prompts and agent roles as well. The prompt
mirror is only pruned for the platform you name, for the same reason.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := "."
			if len(args) == 1 {
				target = args[0]
			}

			report, err := scaffold.Reload(scaffold.ReloadOptions{
				TargetDir:  target,
				AIPlatform: aiPlatform,
				PlaesyHome: plaesyHome,
				DryRun:     dryRun,
				Prune:      prune,
				PruneApply: pruneApply,
			})
			if err != nil {
				return err
			}

			common.LogInfo("")

			if len(report.Protected) > 0 {
				common.LogInfo("Preserved (your data, not regenerated):")
				for _, p := range report.Protected {
					common.LogInfo("%s", "  · "+p)
				}
			}
			if len(report.Failed) > 0 {
				common.LogError("Failed:")
				for _, p := range report.Failed {
					common.LogError("%s", "  ✗ "+p)
				}
			}

			if len(report.Stale) > 0 {
				if len(report.Pruned) > 0 {
					common.LogInfo("Pruned (removed, no counterpart in the sources):")
					for _, p := range report.Pruned {
						common.LogInfo("%s", "  · "+p)
					}
				} else {
					common.LogInfo("Stale (generated, no counterpart in the sources — NOT removed):")
					for _, p := range report.Stale {
						common.LogInfo("%s", "  · "+p)
					}
				}
			}
			common.LogInfo("%s", "Summary: "+report.Summary())

			if len(report.Failed) > 0 {
				// Non-zero exit: a partial refresh that silently reported
				// success is how drift becomes permanent in the first place.
				return fmt.Errorf("reload completed with %d failure(s); see the list above",
					len(report.Failed))
			}
			if report.DryRun {
				if report.Changed() {
					common.LogSuccess("Dry run complete — re-run without --dry-run to apply.")
				} else {
					common.LogSuccess("Dry run complete — everything is already up to date.")
				}
				return nil
			}
			if report.Changed() {
				common.LogSuccess("Reload complete.")
			} else {
				common.LogSuccess("Nothing to do — .plaesy/ already matches the sources.")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&aiPlatform, "ai", "", "Also refresh this AI platform's core file, prompts and roles (e.g. claude, cursor_ai, github_copilot, kilo); default: skip platform files")
	cmd.Flags().StringVar(&plaesyHome, "plaesy-home", "", "Override the Plaesy repo root to copy sources from")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Report what would change without writing anything")
	cmd.Flags().BoolVar(&prune, "prune", false, "List generated files that no longer have a source (agent roles, prompt mirror); deletes nothing")
	cmd.Flags().BoolVar(&pruneApply, "prune-apply", false, "With --prune, actually delete the stale files listed")

	return cmd
}
