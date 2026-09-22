// Package config ports scripts/bash/config-manager.sh and
// scripts/powershell/config-manager.ps1: centralized reads of
// scripts/configs/platform.json (AI-platform detection/mapping and the
// Plaesy directory layout) used by plaesy-init and install.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/plaesy/spec-kit/internal/common"
)

// MappingEntry is a platform.json mapping value. In plaesy.mapping it is
// always an object ({"value": ..., "excludes": [...]}); in platforms.*.mapping
// it is always a bare string. UnmarshalJSON accepts either shape so one type
// serves both, mirroring the bash script's "new structure vs old structure"
// fallback (get-mapping-value).
type MappingEntry struct {
	Value       string   `json:"value"`
	Excludes    []string `json:"excludes,omitempty"`
	Description string   `json:"description,omitempty"`
}

func (m *MappingEntry) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		m.Value = s
		return nil
	}
	type alias MappingEntry
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*m = MappingEntry(a)
	return nil
}

// PlaesyStructure is the platform.json "plaesy" section: the base directory
// layout and the source->target mapping used to populate a platform's
// instruction/prompt/chatmode directories.
type PlaesyStructure struct {
	BaseDirectory      string                  `json:"base_directory"`
	CoreDirectories    []string                `json:"core_directories"`
	ProjectDirectories []string                `json:"project_directories"`
	Mapping            map[string]MappingEntry `json:"mapping"`
}

// Platform is one entry under platform.json "platforms".
type Platform struct {
	Name      string            `json:"name"`
	Provider  string            `json:"provider"`
	Category  string            `json:"category"`
	Detection []string          `json:"detection"`
	Mapping   map[string]string `json:"mapping"`
}

// PlatformConfig is the full parsed contents of platform.json.
type PlatformConfig struct {
	Version     string              `json:"version"`
	Description string              `json:"description"`
	Plaesy      PlaesyStructure     `json:"plaesy"`
	Platforms   map[string]Platform `json:"platforms"`
}

// DefaultConfigPath locates scripts/configs/platform.json relative to the
// repo root, mirroring config-manager.sh's $SCRIPT_DIR/../configs/platform.json.
func DefaultConfigPath() (string, error) {
	root, err := common.GetRepoRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "scripts", "configs", "platform.json"), nil
}

// Load reads and parses platform.json from path.
func Load(path string) (*PlatformConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("platform configuration file not found: %s", path)
	}
	var cfg PlatformConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid JSON syntax in platform configuration: %w", err)
	}
	return &cfg, nil
}

// ListPlatforms returns platform keys in the order platform.json defines
// them (Go maps don't preserve order, so it re-reads the raw JSON key order).
func (c *PlatformConfig) ListPlatforms(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// Decode with json.Decoder/Token to recover declaration order of the
	// "platforms" object's top-level keys.
	dec := json.NewDecoder(strings.NewReader(string(data)))
	names, err := orderedTopLevelKeys(dec, "platforms")
	if err != nil || len(names) == 0 {
		// Fall back to sorted map keys (still correct, just reordered).
		names = make([]string, 0, len(c.Platforms))
		for k := range c.Platforms {
			names = append(names, k)
		}
		sort.Strings(names)
	}
	return names, nil
}

// orderedTopLevelKeys scans a JSON document for object `section` and returns
// the immediate child keys in declaration order.
func orderedTopLevelKeys(dec *json.Decoder, section string) ([]string, error) {
	depth := 0
	inSection := false
	sectionDepth := 0
	var keys []string
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case json.Delim:
			switch t {
			case '{', '[':
				depth++
			case '}', ']':
				depth--
				if inSection && depth < sectionDepth {
					inSection = false
				}
			}
		case string:
			if !inSection && t == section {
				// Next token should be the opening '{' of the section.
				next, err := dec.Token()
				if err != nil {
					return keys, nil
				}
				if d, ok := next.(json.Delim); ok && d == '{' {
					depth++
					inSection = true
					sectionDepth = depth
				}
				continue
			}
			if inSection && depth == sectionDepth {
				keys = append(keys, t)
			}
		}
	}
	return keys, nil
}

// GetPlatform returns the named platform or an error if it is unknown.
func (c *PlatformConfig) GetPlatform(name string) (Platform, error) {
	p, ok := c.Platforms[name]
	if !ok {
		return Platform{}, fmt.Errorf("unknown platform: %s", name)
	}
	return p, nil
}

// GetPlatformConfig mirrors get-platform-config: fetch a top-level field
// (name/provider/category) or a "mapping.<type>" value for a platform.
func (c *PlatformConfig) GetPlatformConfig(platform, key string) (string, error) {
	p, err := c.GetPlatform(platform)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(key, "mapping.") {
		mappingType := strings.TrimPrefix(key, "mapping.")
		return p.Mapping[mappingType], nil
	}
	switch key {
	case "name":
		return p.Name, nil
	case "provider":
		return p.Provider, nil
	case "category":
		return p.Category, nil
	}
	return "", nil
}

// GetMappingValue mirrors get-mapping-value: the mapping value for a section
// (plaesy structure section, e.g. "core") and mapping type. It reads from
// the plaesy.mapping structure ({"value": ...}); the bash "old structure"
// fallback onto platforms.*.mapping.* does not apply to plaesy.mapping keys
// and is not reachable via the plaesy structure, so is omitted here.
func (c *PlatformConfig) GetMappingValue(section string) (string, error) {
	entry, ok := c.Plaesy.Mapping[section]
	if !ok {
		return "", nil
	}
	return entry.Value, nil
}

// GetMappingExcludes mirrors get-mapping-excludes.
func (c *PlatformConfig) GetMappingExcludes(section string) []string {
	entry, ok := c.Plaesy.Mapping[section]
	if !ok {
		return nil
	}
	return entry.Excludes
}

// DetectPlatform mirrors detect-platform: the first platform (in
// declaration order) whose detection patterns match a file or directory
// under cwd. Returns "" with no error when nothing matches (interactive
// selection trigger in the bash original).
func (c *PlatformConfig) DetectPlatform(path string) (string, error) {
	names, err := c.ListPlatforms(path)
	if err != nil {
		return "", err
	}
	for _, name := range names {
		p := c.Platforms[name]
		if len(p.Detection) == 0 {
			continue
		}
		for _, pattern := range p.Detection {
			if _, err := os.Stat(pattern); err == nil {
				return name, nil
			}
		}
	}
	return "", nil
}

// GetCleanFiles mirrors get-clean-files: hardcoded fallback list (the bash
// original notes cleanup configuration is not present in platform.json).
func GetCleanFiles() []string {
	return []string{"CLAUDE.md", ".cursorrules", ".github/copilot-instructions.md"}
}

// GetSourceRules mirrors get-source-rules: hardcoded per-file-type globs
// used by plaesy-init when copying source files.
func GetSourceRules(fileType string) []string {
	switch fileType {
	case "core":
		return []string{"instructions/plaesy.instructions.md"}
	case "instructions":
		return []string{"instructions/*.instructions.md", "plaesy.instructions.md"}
	case "prompts":
		return []string{"prompts/*.prompt.md"}
	case "chatmodes":
		return []string{"chatmodes/*.chatmode.md"}
	}
	return nil
}

// GetCleanDirs mirrors get-clean-dirs: derive cleanup directories from a
// platform's mapping (core/prompts/chatmodes target dirs), falling back to
// scanning every platform's core mapping when the given platform yields none.
func (c *PlatformConfig) GetCleanDirs(platform string) ([]string, error) {
	var dirs []string
	add := func(target string) {
		if target == "" || target == "null" {
			return
		}
		dir := filepath.Dir(filepath.ToSlash(target))
		if dir == "." {
			return
		}
		for _, existing := range dirs {
			if existing == dir {
				return
			}
		}
		dirs = append(dirs, dir)
	}

	if platform != "" {
		if p, ok := c.Platforms[platform]; ok {
			add(p.Mapping["core"])
			add(p.Mapping["prompts"])
			add(p.Mapping["chatmodes"])
		}
	}

	if len(dirs) == 0 {
		for _, p := range c.Platforms {
			add(p.Mapping["core"])
		}
	}

	if len(dirs) == 0 {
		return nil, fmt.Errorf("no clean directories found")
	}
	return dirs, nil
}

// Validate mirrors validate: platform.json must exist and parse as JSON.
func Validate(path string) error {
	_, err := Load(path)
	return err
}

// PlaesyStructureComponent mirrors get-plaesy-structure: returns either a
// space-joined list (for the three array components) or a single value.
func (c *PlatformConfig) PlaesyStructureComponent(component string) (string, error) {
	switch component {
	case "core_directories":
		return strings.Join(c.Plaesy.CoreDirectories, " "), nil
	case "project_directories":
		return strings.Join(c.Plaesy.ProjectDirectories, " "), nil
	case "memory_subdirectories":
		// Not present in the current platform.json; the bash original
		// returns an empty match for this case too.
		return "", nil
	case "base_directory":
		return c.Plaesy.BaseDirectory, nil
	}
	return "", nil
}
