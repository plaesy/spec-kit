package scaffold

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// PlatformConfig mirrors the relevant shape of scripts/configs/platform.json.
// Only the fields plaesy-init.sh actually reads are modeled.
type PlatformConfig struct {
	Plaesy    PlaesySection            `json:"plaesy"`
	Platforms map[string]PlatformEntry `json:"platforms"`
}

// PlaesySection is the config's "plaesy" block: base directory, the
// directories always created under it, the project-level directories
// created alongside it, and the source-side mapping used by both the
// universal copy step and the per-platform setup step.
type PlaesySection struct {
	BaseDirectory      string                  `json:"base_directory"`
	CoreDirectories    []string                `json:"core_directories"`
	ProjectDirectories []string                `json:"project_directories"`
	Mapping            map[string]MappingEntry `json:"mapping"`
}

// MappingEntry is one entry of plaesy.mapping (source side) or a
// platform's mapping (destination side, expressed as a bare string).
type MappingEntry struct {
	Value    string   `json:"value"`
	Excludes []string `json:"excludes"`
}

// PlatformEntry describes one AI platform: display name plus its
// destination mapping (core/instructions/prompts/chatmodes -> target path).
type PlatformEntry struct {
	Name    string            `json:"name"`
	Mapping map[string]string `json:"mapping"`
}

// defaultBaseDirectory, defaultCoreDirectories and defaultProjectDirectories
// mirror plaesy-init.sh's create_structure() fallbacks when platform.json
// cannot be read or parsed.
const defaultBaseDirectory = ".plaesy"

var (
	defaultCoreDirectories    = []string{"memory", "instructions"}
	defaultProjectDirectories = []string{"docs", "specs"}
)

// LoadPlatformConfig reads scripts/configs/platform.json from the given
// Plaesy home. On any read/parse error it returns a config populated with
// just the plaesy-init.sh defaults, matching the bash script's fallback
// behavior (it never hard-fails just because the config is missing).
func LoadPlatformConfig(home string) (*PlatformConfig, error) {
	path := filepath.Join(home, "scripts", "configs", "platform.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return fallbackConfig(), fmt.Errorf("reading %s: %w", path, err)
	}

	var cfg PlatformConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fallbackConfig(), fmt.Errorf("parsing %s: %w", path, err)
	}

	if cfg.Plaesy.BaseDirectory == "" {
		cfg.Plaesy.BaseDirectory = defaultBaseDirectory
	}
	if len(cfg.Plaesy.CoreDirectories) == 0 {
		cfg.Plaesy.CoreDirectories = defaultCoreDirectories
	}
	if len(cfg.Plaesy.ProjectDirectories) == 0 {
		cfg.Plaesy.ProjectDirectories = defaultProjectDirectories
	}
	return &cfg, nil
}

func fallbackConfig() *PlatformConfig {
	return &PlatformConfig{
		Plaesy: PlaesySection{
			BaseDirectory:      defaultBaseDirectory,
			CoreDirectories:    defaultCoreDirectories,
			ProjectDirectories: defaultProjectDirectories,
		},
	}
}

// PlatformNames returns the configured platform ids (map iteration order is
// randomized, so this sorts for deterministic --help/listing output).
func (c *PlatformConfig) PlatformNames() []string {
	names := make([]string, 0, len(c.Platforms))
	for name := range c.Platforms {
		names = append(names, name)
	}
	for i := 1; i < len(names); i++ {
		for j := i; j > 0 && names[j] < names[j-1]; j-- {
			names[j], names[j-1] = names[j-1], names[j]
		}
	}
	return names
}

// DisplayName returns the platform's configured display name, or the raw id
// if the platform is unknown (matching _get_platform_name_fallback's final
// `*) echo "$platform"` case).
func (c *PlatformConfig) DisplayName(platform string) string {
	if entry, ok := c.Platforms[platform]; ok && entry.Name != "" {
		return entry.Name
	}
	return platform
}

// HasPlatform reports whether platform is a known platform id.
func (c *PlatformConfig) HasPlatform(platform string) bool {
	_, ok := c.Platforms[platform]
	return ok
}
