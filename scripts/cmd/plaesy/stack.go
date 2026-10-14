package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/plaesy/spec-kit/internal/common"
	"github.com/plaesy/spec-kit/internal/detectstack"
	"github.com/spf13/cobra"
)

func init() { register(newStackCmd()) }

// `plaesy stack detect` was the bare `plaesy detect-stack`.
//
// "detect-stack" is a compound of the two things the rest of this tree was just
// un-bundling: it detects a technology STACK, and it does so from marker FILES
// in a target project. Both are real nouns here — `platforms` is a target, a
// `stack` is a target, and a `detect` is a verb, so `stack detect` says the same
// thing while matching `features validate`, `images create` and `tasks
// update`. "detect-stack" also read as a typo'd `detect --stack`.
//
// `stack` currently has exactly one subcommand, which is normally a sign that a
// parent is premature. It earns the wrapper because the noun is the stable part:
// reading a stack to pick instructions is one of several things a future
// `stack` may do, and a future `stack explain` or `stack check` then lands
// without another top-level rename.
func newStackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stack",
		Short: "Inspect a target project's technology stack",
		Long: `Inspect a target project's technology stack.

  plaesy stack detect [target-dir]   instruction files matching that stack
  plaesy stack detect --install      also copy them into <target>/.plaesy/instructions/

With no subcommand, prints this list.`,
		Args: cobra.NoArgs,
	}
	cmd.AddCommand(newStackDetectCmd())
	return showHelpWhenBare(cmd)
}

// newStackDetectCmd was the top-level `plaesy detect-stack`.
func newStackDetectCmd() *cobra.Command {
	var plaesyRoot string
	var install bool

	cmd := &cobra.Command{
		Use:   "detect [target-dir]",
		Short: "List instruction files relevant to a target project's tech stack",
		Long: `List instruction files relevant to a target project's tech stack.

  plaesy stack detect [target-dir]   instruction files matching that stack
  plaesy stack detect --install      also copy them into <target>/.plaesy/instructions/

Without --install this only prints. Detection alone does not put a file on disk:
an instruction file that is never copied cannot be loaded by any command, so a
project whose stack is known should be installed once after plaesy init.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetDir := "."
			if len(args) == 1 {
				targetDir = args[0]
			}

			root := plaesyRoot
			if root == "" {
				root = os.Getenv("PLAESY_ROOT")
			}
			if root == "" {
				if r, err := common.GetRepoRoot(); err == nil {
					root = r
				}
			}
			if root == "" {
				return fmt.Errorf("could not determine PLAESY_ROOT: not a git repository and PLAESY_ROOT is unset")
			}

			files, err := detectstack.Detect(targetDir, root)
			if err != nil {
				return err
			}

			if !install {
				for _, f := range files {
					fmt.Println(f)
				}
				return nil
			}

			destDir := filepath.Join(targetDir, ".plaesy", "instructions")
			if err := os.MkdirAll(destDir, 0o755); err != nil {
				return fmt.Errorf("creating %s: %w", destDir, err)
			}

			installed := 0
			for _, f := range files {
				src := filepath.Join(root, "instructions", f)
				base := strings.TrimSuffix(filepath.Base(f), ".instructions.md")
				dst := filepath.Join(destDir, base+".md")

				// never clobber a hand-edited installed file
				if _, err := os.Stat(dst); err == nil {
					fmt.Printf("= %s (exists, kept)\n", base+".md")
					continue
				}

				if err := copyDetected(src, dst); err != nil {
					fmt.Printf("! %s (%v)\n", base+".md", err)
					continue
				}
				fmt.Printf("+ %s\n", base+".md")
				installed++
			}
			fmt.Printf("\n%d file(s) installed into %s, %d skipped.\n", installed, destDir, len(files)-installed)
			fmt.Println("Re-run `plaesy stack detect --install` after adding a dependency that changes the stack.")
			return nil
		},
	}

	cmd.Flags().StringVar(&plaesyRoot, "plaesy-root", "", "path to the plaesy repo root (defaults to $PLAESY_ROOT, then git repo root)")
	cmd.Flags().BoolVar(&install, "install", false, "copy the detected instruction files into <target>/.plaesy/instructions/ (never overwrites an existing file)")
	return cmd
}

// copyDetected copies one detected instruction file into the project's
// .plaesy/instructions/. The destination name drops the ".instructions" segment,
// matching scaffold.copyInstructions so both install paths produce the same
// flat <name>.md layout and never disagree about what a file is called.
func copyDetected(src, dst string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, b, 0o644)
}
