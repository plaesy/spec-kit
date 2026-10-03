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
		Short: "Read and validate scripts/configs/platform.json",
		Long: `Read and validate scripts/configs/platform.json (port of config-manager.sh).

  plaesy config get-mapping <section>    which files a section maps to
  plaesy config get-excludes <section>   files a section deliberately skips
  plaesy config get-clean-files <path>    vendor-style files to ignore
  plaesy config get-clean-dirs <path>    build directories to ignore
  plaesy config get-structure            the full .plaesy/ layout
  plaesy config validate                 is platform.json well-formed?

These stay here because the object of each is the CONFIG FILE, not a platform.
Asking which platform this project uses is "plaesy platforms detect".

With no subcommand, prints this list.`,
		Args: cobra.NoArgs,
	}
	cmd.PersistentFlags().StringVar(&configFile, "config", "", "path to platform.json (default: scripts/configs/platform.json under the repo root)")

	cmd.AddCommand(
		newConfigGetMappingValueCmd(&configFile),
		newConfigGetMappingExcludesCmd(&configFile),
		newConfigGetCleanFilesCmd(),
		newConfigGetCleanDirsCmd(&configFile),
		newConfigGetPlaesyStructureCmd(&configFile),
		newConfigValidateCmd(&configFile),
	)
	return showHelpWhenBare(cmd)
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

func newConfigGetMappingValueCmd(configFile *string) *cobra.Command {
	return &cobra.Command{
		Use:   "get-mapping <section> <mapping_type>",
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
		Use:   "get-excludes <section> <mapping_type>",
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
				platform = config.NormalizePlatform(args[0], cfg)
			} else {
				// Detection can fail, and the discarded error is not a
				// formatting detail: on failure platform stays "", and
				// GetCleanDirs("") answers for the generic platform rather than
				// for the one in front of the user. The command then prints a
				// directory list that looks like a detection result and exits 0.
				platform, err = cfg.DetectPlatform(path)
				if err != nil {
					return err
				}
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
		Use:   "get-structure <component>",
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

func newConfigValidateCmd(configFile *string) *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate platform configuration",
		// The file comes from --config; a stray argument would be ignored and
		// the command would report a different file as valid.
		Args: cobra.NoArgs,
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
