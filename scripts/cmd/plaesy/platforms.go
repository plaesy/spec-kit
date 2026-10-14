package main

import (
	"fmt"

	"github.com/plaesy/spec-kit/internal/config"
	"github.com/spf13/cobra"
)

// `plaesy platforms` — the AI platform a project targets (claude,
// cursor_ai, github_copilot, …).
//
// These four subcommands were inside `plaesy config`, which grouped them by the
// FILE they read (scripts/configs/platform.json) rather than by the OBJECT a
// user is thinking about. Nobody asks "what does platform.json say" — they ask
// "which platform am I on" and "which platforms exist". Grouping by resource is
// what kubectl (`get configmap`), `gh` (`gh repo`) and terraform (`terraform
// workspace`) do, and it completes the noun family this CLI now has:
// `features`, `tasks`, `images`, `platforms`, `context`, `config`, `stack`.
//
// What stayed in `config` is the part that genuinely is about the file:
// `get-mapping`, `get-excludes`, `get-clean-files`, `get-clean-dirs`,
// `get-structure` and `validate`. `platforms get-structure` or
// `platforms validate` would name a resource the command does not touch.
//
// The old `config detect-platform`, `config list-platforms`,
// `config show-platform-info` and `config get-platform-config` spellings were
// removed rather than aliased — see features.go for why a leftover alias would
// be invisible.
func init() { register(newPlatformsCmd()) }

func newPlatformsCmd() *cobra.Command {
	var configFile string

	cmd := &cobra.Command{
		Use:   "platforms",
		Short: "AI platforms plaesy can target (claude, cursor_ai, …)",
		Long: `Work with the AI platform a project targets.

  plaesy platforms detect             which platform this project is on
  plaesy platforms list               every platform plaesy knows about
  plaesy platforms show [platform]    one platform in detail
  plaesy platforms get <platform> <key>

With no subcommand, prints this list.`,
		Example: `  plaesy platforms detect
  plaesy platforms list
  plaesy platforms show claude
  plaesy platforms get claude mapping.core`,
		Args: cobra.NoArgs,
	}
	cmd.PersistentFlags().StringVar(&configFile, "config", "", "path to platform.json (default: scripts/configs/platform.json under the repo root)")

	cmd.AddCommand(
		newPlatformsDetectCmd(&configFile),
		newPlatformsListCmd(&configFile),
		newPlatformsShowCmd(&configFile),
		newPlatformsGetCmd(&configFile),
	)
	return showHelpWhenBare(cmd)
}

// newPlatformsDetectCmd was `plaesy config detect-platform`.
func newPlatformsDetectCmd(configFile *string) *cobra.Command {
	return &cobra.Command{
		Use:   "detect",
		Short: "Detect current AI platform",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, path, err := loadConfig(*configFile)
			if err != nil {
				return err
			}
			platform, err := cfg.DetectPlatform(path)
			if err != nil {
				return err
			}
			if platform == "" {
				return fmt.Errorf("no platform detected")
			}
			fmt.Println(platform)
			return nil
		},
	}
}

// newPlatformsListCmd was `plaesy config list-platforms`.
func newPlatformsListCmd(configFile *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all available platforms",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, path, err := loadConfig(*configFile)
			if err != nil {
				return err
			}
			names, err := cfg.ListPlatforms(path)
			if err != nil {
				return err
			}
			for _, n := range names {
				fmt.Println(n)
			}
			return nil
		},
	}
}

// newPlatformsGetCmd was `plaesy config get-platform-config`.
func newPlatformsGetCmd(configFile *string) *cobra.Command {
	return &cobra.Command{
		Use:   "get <platform> <key>",
		Short: "Get a specific platform's configuration value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := loadConfig(*configFile)
			if err != nil {
				return err
			}
			// Resolving aliases here (e.g. the legacy "claude_code" long form) is
			// the same fix `clean --ai` needed, applied to the read side.
			value, err := cfg.GetPlatformConfig(config.NormalizePlatform(args[0], cfg), args[1])
			if err != nil {
				return err
			}
			fmt.Println(value)
			return nil
		},
	}
}

// newPlatformsShowCmd was `plaesy config show-platform-info`.
func newPlatformsShowCmd(configFile *string) *cobra.Command {
	return &cobra.Command{
		Use:   "show [platform]",
		Short: "Show detailed platform information",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, path, err := loadConfig(*configFile)
			if err != nil {
				return err
			}
			platform := ""
			if len(args) > 0 {
				platform = config.NormalizePlatform(args[0], cfg)
			} else {
				// A failed detection used to be discarded here, leaving platform
				// empty. GetPlatform("") then fails, and the handler below turned
				// that failure into a page of "Unknown" fields and exit 0 —
				// indistinguishable from a platform whose config genuinely has
				// no name, and a way to get the wrong answer without an error.
				platform, err = cfg.DetectPlatform(path)
				if err != nil {
					return err
				}
			}

			fmt.Printf("Platform Information for: %s\n", platform)
			fmt.Println("================================")

			p, err := cfg.GetPlatform(platform)
			if err != nil {
				fmt.Println("Name: Unknown")
				fmt.Println("Provider: Unknown")
				fmt.Println("Category: Unknown")
				return nil
			}
			printOrUnknown := func(label, value string) {
				if value == "" {
					value = "Unknown"
				}
				fmt.Printf("%s: %s\n", label, value)
			}
			printOrUnknown("Name", p.Name)
			printOrUnknown("Provider", p.Provider)
			printOrUnknown("Category", p.Category)
			return nil
		},
	}
}
