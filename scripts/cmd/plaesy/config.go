package main

import (
	"fmt"
	"strings"

	"github.com/plaesy/spec-kit/internal/config"
	"github.com/spf13/cobra"
)

func init() { register(newConfigCmd()) }

// configPath resolves --config, falling back to the repo-relative default
// (scripts/configs/platform.json), mirroring config-manager.sh's
// $SCRIPT_DIR/../configs/platform.json.
func configPath(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	return config.DefaultConfigPath()
}

func newConfigCmd() *cobra.Command {
	var configFile string

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Plaesy platform configuration management",
		Long:  "Centralized configuration management using scripts/configs/platform.json (port of config-manager.sh).",
	}
	cmd.PersistentFlags().StringVar(&configFile, "config", "", "path to platform.json (default: scripts/configs/platform.json under the repo root)")

	cmd.AddCommand(
		newConfigDetectPlatformCmd(&configFile),
		newConfigListPlatformsCmd(&configFile),
		newConfigGetPlatformConfigCmd(&configFile),
		newConfigGetMappingValueCmd(&configFile),
		newConfigGetMappingExcludesCmd(&configFile),
		newConfigGetCleanFilesCmd(),
		newConfigGetCleanDirsCmd(&configFile),
		newConfigGetPlaesyStructureCmd(&configFile),
		newConfigShowPlatformInfoCmd(&configFile),
		newConfigValidateCmd(&configFile),
	)
	return cmd
}

func loadConfig(configFile string) (*config.PlatformConfig, string, error) {
	path, err := configPath(configFile)
	if err != nil {
		return nil, "", err
	}
	cfg, err := config.Load(path)
	if err != nil {
		return nil, "", err
	}
	return cfg, path, nil
}

func newConfigDetectPlatformCmd(configFile *string) *cobra.Command {
	return &cobra.Command{
		Use:   "detect-platform",
		Short: "Detect current AI platform",
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

func newConfigListPlatformsCmd(configFile *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list-platforms",
		Short: "List all available platforms",
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

func newConfigGetPlatformConfigCmd(configFile *string) *cobra.Command {
	return &cobra.Command{
		Use:   "get-platform-config <platform> <key>",
		Short: "Get specific platform configuration",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := loadConfig(*configFile)
			if err != nil {
				return err
			}
			value, err := cfg.GetPlatformConfig(args[0], args[1])
			if err != nil {
				return err
			}
			fmt.Println(value)
			return nil
		},
	}
}

func newConfigGetMappingValueCmd(configFile *string) *cobra.Command {
	return &cobra.Command{
		Use:   "get-mapping-value <section> <mapping_type>",
		Short: "Get mapping value with support for the value/excludes structure",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := loadConfig(*configFile)
			if err != nil {
				return err
			}
			// The bash original reads plaesy.mapping[section] first, then
			// falls back to platforms[section].mapping[mapping_type]; the
			// section argument only ever names a plaesy.mapping key in
			// practice, so mapping_type is accepted for CLI-surface parity
			// but only the section's mapping value is used.
			value, err := cfg.GetMappingValue(args[0])
			if err != nil {
				return err
			}
			if value == "" {
				value, err = cfg.GetPlatformConfig(args[0], "mapping."+args[1])
				if err != nil {
					return err
				}
			}
			fmt.Println(value)
			return nil
		},
	}
}

func newConfigGetMappingExcludesCmd(configFile *string) *cobra.Command {
	return &cobra.Command{
		Use:   "get-mapping-excludes <section> <mapping_type>",
		Short: "Get exclude patterns for a mapping",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := loadConfig(*configFile)
			if err != nil {
				return err
			}
			for _, e := range cfg.GetMappingExcludes(args[0]) {
				fmt.Println(e)
			}
			return nil
		},
	}
}

func newConfigGetCleanFilesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get-clean-files [platform]",
		Short: "Get files to clean for platform",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(strings.Join(config.GetCleanFiles(), " "))
			return nil
		},
	}
}

func newConfigGetCleanDirsCmd(configFile *string) *cobra.Command {
	return &cobra.Command{
		Use:   "get-clean-dirs [platform]",
		Short: "Get directories to clean for platform",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, path, err := loadConfig(*configFile)
			if err != nil {
				return err
			}
			platform := ""
			if len(args) > 0 {
				platform = args[0]
			} else {
				platform, _ = cfg.DetectPlatform(path)
			}
			dirs, err := cfg.GetCleanDirs(platform)
			if err != nil {
				return err
			}
			fmt.Println(strings.Join(dirs, " "))
			return nil
		},
	}
}

func newConfigGetPlaesyStructureCmd(configFile *string) *cobra.Command {
	return &cobra.Command{
		Use:   "get-plaesy-structure <component>",
		Short: "Get Plaesy structure configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := loadConfig(*configFile)
			if err != nil {
				return err
			}
			value, err := cfg.PlaesyStructureComponent(args[0])
			if err != nil {
				return err
			}
			fmt.Println(value)
			return nil
		},
	}
}

func newConfigShowPlatformInfoCmd(configFile *string) *cobra.Command {
	return &cobra.Command{
		Use:   "show-platform-info [platform]",
		Short: "Show detailed platform information",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, path, err := loadConfig(*configFile)
			if err != nil {
				return err
			}
			platform := ""
			if len(args) > 0 {
				platform = args[0]
			} else {
				platform, _ = cfg.DetectPlatform(path)
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

func newConfigValidateCmd(configFile *string) *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate platform configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := configPath(*configFile)
			if err != nil {
				return err
			}
			if err := config.Validate(path); err != nil {
				return err
			}
			fmt.Println("Platform configuration is valid")
			return nil
		},
	}
}
