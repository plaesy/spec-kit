package main

import "github.com/spf13/cobra"

// `plaesy images create` — the one asset this CLI generates.
//
// It was `plaesy generate-image`, then briefly `plaesy create image`. The
// settled shape is the same one `features create` and `tasks start` use:
// **resource plural, then verb**. `create` was left with a single child and was
// the only top-level verb in a tree that is otherwise resource-nouns
// (`config`, `features`, `tasks`, `context`, `validate`), so it went too.
func init() { register(newImagesCmd()) }

func newImagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "images",
		Short: "Image assets generated through a provider API",
		Long: `Work with generated image assets.

  plaesy images create --prompt <text> --out <path>

With no subcommand, prints this list.`,
		Args: cobra.NoArgs,
	}
	cmd.AddCommand(newImagesCreateCmd())
	return showHelpWhenBare(cmd)
}
