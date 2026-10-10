package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/plaesy/spec-kit/internal/cleaner"
	"github.com/plaesy/spec-kit/internal/common"
	"github.com/plaesy/spec-kit/internal/config"
	"github.com/spf13/cobra"
)

func init() { register(newCleanCmd()) }

func newCleanCmd() *cobra.Command {
	var (
		autoConfirm bool
		dryRun      bool
		backup      bool
		noBackup    bool
		level       string
		aiChoice    string
		verbose     bool
		configFile  string
	)

	cmd := &cobra.Command{
		Use:   "clean [TARGET_DIR]",
		Short: "Remove Plaesy Constitution Kit framework files and directories",
		Long: `Remove Plaesy Spec-Kit framework files and directories with safety checks.

CLEANUP LEVELS:
  safe        Remove framework files only, preserve user code (default)
  thorough    Remove framework + specs, preserve user code
  complete    Remove everything Plaesy-related (DANGEROUS)

Note: This operation removes Plaesy framework files but preserves your source code.
Always use --dry-run first to preview what will be removed.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cleaner.ValidLevel(level) {
				return fmt.Errorf("invalid cleanup level: %s (available levels: safe, thorough, complete)", level)
			}

			targetDir := "."
			if len(args) == 1 {
				targetDir = args[0]
			}

			// --backup defaults to true, so the flag's own value cannot be the
			// answer: only an explicit --no-backup (or --backup=false) turns it
			// off. Reading "if noBackup ... else if backup" made `--backup=false`
			// fall through to the default and keep the backup anyway, which is the
			// opposite of what the command line said.
			backupEnabled := true
			if noBackup {
				backupEnabled = false
			} else if cmd.Flags().Changed("backup") {
				backupEnabled = backup
			}

			opts := cleaner.Options{
				TargetDir:   targetDir,
				AutoConfirm: autoConfirm,
				DryRun:      dryRun,
				Backup:      backupEnabled,
				Level:       cleaner.Level(level),
				AIChoice:    aiChoice,
				Verbose:     verbose,
			}

			path, err := configPath(configFile)
			if err != nil {
				return err
			}

			if aiChoice != "" {
				cfg, err := config.Load(path)
				if err != nil {
					return err
				}
				// Resolve the shorthand to the id platform.json declares before
				// anything looks a platform up by key. "--ai claude" used to be
				// accepted, printed as the chosen platform, and then matched no
				// mapping at all, so a claude-specific clean removed nothing and
				// the plan said there was nothing to remove.
				opts.AIChoice = config.NormalizePlatform(aiChoice, cfg)
				if !cfg.HasPlatform(opts.AIChoice) {
					return fmt.Errorf("invalid AI assistant: %s (available assistants: %v)",
						aiChoice, cfg.PlatformNames())
				}
			}

			c, err := cleaner.New(opts, path)
			if err != nil {
				return err
			}

			if verbose {
				fmt.Println("Configuration:")
				fmt.Printf("  Cleanup Level: %s\n", opts.Level)
				fmt.Printf("  Dry Run: %v\n", opts.DryRun)
				fmt.Printf("  Backup: %v\n", opts.Backup)
				platformDisplay := opts.AIChoice
				if platformDisplay == "" {
					platformDisplay = "auto-detect"
				}
				fmt.Printf("  Platform: %s\n", platformDisplay)
				fmt.Printf("  Target Directory: %s\n", c.Opts.TargetDir)
				fmt.Println()
			}

			resolvedAI := opts.AIChoice
			if resolvedAI == "" {
				fmt.Println("Detecting AI platforms for cleanup...")
				platforms, detectErr := cleaner.DetectAllPlatforms(c.Config, c.ConfigPath, c.Opts.TargetDir)
				if detectErr != nil {
					// A detection marker that could not be read is not the same
					// as a platform that is absent, and this switch would turn
					// the first into the second: a wrong platform set, and
					// therefore a wrong deletion set, reported as detection.
					return detectErr
				}
				switch {
				case len(platforms) == 1 && platforms[0] != "generic_ai":
					fmt.Printf("Detected AI platform: %s\n", platforms[0])
					fmt.Println("Using platform-adapted cleanup...")
					resolvedAI = platforms[0]
				case len(platforms) > 1:
					fmt.Printf("Detected multiple AI platforms: %s\n", strings.Join(platforms, ", "))
					fmt.Println("Using multi-platform cleanup...")
					resolvedAI = strings.Join(platforms, " ")
				default:
					fmt.Println("No specific AI platform detected. Using generic cleanup...")
					// Keep the config's own fallback id. This used to be the
					// literal "generic", which is the --ai "no platform"
					// sentinel that BuildPlan and Remove both skip, so a project
					// with no AI markers had its generic_ai files spared at every
					// level -- including `complete`, the level whose whole job is
					// not sparing them.
					resolvedAI = platforms[0]
				}
				c.Opts.AIChoice = resolvedAI
			} else {
				fmt.Printf("Using specified AI: %s\n", resolvedAI)
			}

			if err := config.Validate(c.ConfigPath); err != nil {
				return fmt.Errorf("configuration validation failed: %w", err)
			}

			common.PrintBanner("Plaesy Clean", "Remove plaesy framework directories")
			if err := c.ValidateEnvironment(); err != nil {
				return err
			}

			plan := c.BuildPlan()
			c.PrintPlan(plan)
			if plan.Empty {
				common.LogSuccess("No Plaesy directories or files found to remove.")
				return nil
			}

			if !cleaner.Confirm(c.Opts, bufio.NewReader(os.Stdin), os.Stdout) {
				fmt.Println("Clean operation cancelled.")
				return nil
			}

			if c.Opts.Backup && !c.Opts.DryRun {
				if _, err := c.CreateBackup(); err != nil {
					return err
				}
			}

			if !c.Opts.DryRun {
				if _, err := c.Remove(); err != nil {
					return err
				}
			}

			fmt.Println()
			common.LogSuccess("\U0001F9F9 Plaesy clean completed!")

			if c.Opts.AIChoice != "" {
				fmt.Println()
				fmt.Println("Platform-specific cleanup guidance:")
				// The ids are what detection and --ai both produce, so the switch
				// used to be keyed on names ("claude", "copilot") that no code path
				// could ever set: a detected clean always fell to "other platforms".
				// Keep the shorthands too, since a caller can still set Opts directly.
				switch strings.Fields(c.Opts.AIChoice)[0] {
				case "claude", "claude_code":
					fmt.Println("For Claude Code: Use /clean command in future")
				case "cursor", "cursor_ai":
					fmt.Println("For Cursor AI: Use --clean command in future")
				case "windsurf", "windsurf_ai":
					fmt.Println("For Windsurf AI: Use clean command in future")
				case "copilot", "github_copilot":
					fmt.Println("For GitHub Copilot: Continue using plaesy clean command")
				default:
					fmt.Println("For other platforms: Use plaesy clean command")
				}
			}
			fmt.Println()
			fmt.Printf("Cleaned directory: %s\n", c.Opts.TargetDir)
			fmt.Println()
			fmt.Println("To reinitialize the project:")
			fmt.Println("   plaesy init [--ai=your-choice]")
			fmt.Println()

			return nil
		},
	}

	cmd.Flags().BoolVarP(&autoConfirm, "yes", "y", false, "auto-confirm deletion (skip confirmation prompt)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be removed without actually removing")
	cmd.Flags().BoolVar(&backup, "backup", true, "create backup before removal (enabled by default)")
	cmd.Flags().BoolVar(&noBackup, "no-backup", false, "skip backup creation")
	cmd.Flags().StringVar(&level, "level", "safe", "cleanup level (safe|thorough|complete)")
	cmd.Flags().StringVar(&aiChoice, "ai", "", "AI platform to clean (auto-detected if not specified)")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "show detailed progress")
	cmd.Flags().StringVar(&configFile, "config", "", "path to platform.json (default: scripts/configs/platform.json under the repo root)")

	return cmd
}
