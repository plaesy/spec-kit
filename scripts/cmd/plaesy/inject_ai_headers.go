package main

import (
	"fmt"
	"path/filepath"

	"github.com/plaesy/spec-kit/internal/aiheaders"
	"github.com/plaesy/spec-kit/internal/common"
	"github.com/spf13/cobra"
)

func init() { register(newInjectAIHeadersCmd()) }

func newInjectAIHeadersCmd() *cobra.Command {
	var (
		aiPlatform string
		targetDir  string
		headersDir string
		dryRun     bool
		force      bool
		merge      bool
		backup     bool
		listOnly   bool
		patterns   []string
		excludes   []string
	)

	cmd := &cobra.Command{
		Use:   "inject-ai-headers",
		Short: "Inject platform-specific AI headers into prompt/chatmode/instructions files",
		Long: `Inject platform-specific headers into prompt files based on the chosen AI platform.

Headers are stored as YAML files in templates/ai-headers/:
  <platform>.prompts.yaml       Headers for prompt files
  <platform>.chatmodes.yaml     Headers for chatmode files
  <platform>.instructions.yaml  Headers for instruction files
  <platform>.header.yaml        Generic headers for the platform

Supported AI platforms: ` + fmt.Sprint(aiheaders.ValidPlatforms),
		RunE: func(cmd *cobra.Command, args []string) error {
			if headersDir == "" {
				root, err := common.GetRepoRoot()
				if err != nil {
					return fmt.Errorf("could not determine repo root (use --headers-dir to override): %w", err)
				}
				headersDir = filepath.Join(root, "templates", "ai-headers")
			}

			opts := &aiheaders.Options{
				AIPlatform: aiPlatform,
				TargetDir:  targetDir,
				HeadersDir: headersDir,
				DryRun:     dryRun,
				Force:      force,
				Merge:      merge,
				Backup:     backup,
				ListOnly:   listOnly,
				Patterns:   patterns,
				Excludes:   excludes,
				Stdout:     cmd.OutOrStdout(),
				Stderr:     cmd.ErrOrStderr(),
			}

			_, err := aiheaders.Run(opts)
			return err
		},
	}

	cmd.Flags().StringVar(&aiPlatform, "ai", "", "AI platform to configure headers for (required)")
	cmd.Flags().StringVar(&targetDir, "target", "", "Target directory containing prompt files (required)")
	cmd.Flags().StringVar(&headersDir, "headers-dir", "", "Directory containing header YAML files (default: <repo-root>/templates/ai-headers)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be changed without making changes")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing headers without confirmation")
	cmd.Flags().BoolVar(&backup, "backup", false, "Create backup of original files")
	cmd.Flags().BoolVar(&merge, "merge", false, "Merge header keys into existing front-matter (add missing keys)")
	cmd.Flags().BoolVar(&listOnly, "list-only", false, "List files and their resolved header mapping, then exit")
	cmd.Flags().StringArrayVar(&patterns, "pattern", nil, "Glob pattern of files to process (repeatable; default: *.prompt.md, *.chatmode.md, *.instructions.md)")
	cmd.Flags().StringArrayVar(&excludes, "exclude", nil, "Glob pattern of paths to exclude (repeatable)")

	return cmd
}
