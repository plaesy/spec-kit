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
// instruction/prompt/agent directories.
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

// DefaultConfigPath locates platform.json relative to the repo root, mirroring
// config-manager.sh's $SCRIPT_DIR/../configs/platform.json.
func DefaultConfigPath() (string, error) {
	root, err := common.GetRepoRoot()
	if err != nil {
		return "", err
	}
	return defaultConfigPath(root), nil
}

// defaultConfigPath is DefaultConfigPath with the root named, so both layouts
// can be tested without a checkout that happens to be shaped like one of them.
//
// platform.json lives in two places, and which one is correct depends on what
// the root is. In this framework's own source tree it is
// scripts/configs/platform.json, next to the code that reads it. In a project
// created by `plaesy init` it is .plaesy/scripts/configs/platform.json, because
// init copies the scripts tree under .plaesy. Resolving only the first meant
// that every project the tool scaffolded was a project in which `plaesy
// platforms` and `plaesy config` could not run — the tool's own output was not
// a valid input to two of its command families. No test caught it because the
// only test could run against this repository, where the hardcoded path is
// correct by construction.
func defaultConfigPath(root string) string {
	if _, err := os.Stat(filepath.Join(root, ".plaesy")); err == nil {
		return filepath.Join(root, ".plaesy", "scripts", "configs", "platform.json")
	}
	return filepath.Join(root, "scripts", "configs", "platform.json")
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
// configPath is the platform.json to read the declaration order from — not a
// directory to search. Both functions that take it are named for what it is,
// because the previous name (`path`) invited a reader to "fix" DetectPlatform
// into resolving its patterns against it, which would search for platform
// markers inside the configs/ folder.
func (c *PlatformConfig) ListPlatforms(configPath string) ([]string, error) {
	data, err := os.ReadFile(configPath)
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

// HasPlatform reports whether name is a configured platform id. It is the
// existence check that pairs with NormalizePlatform: normalize first, then ask
// whether what came back is real.
func (c *PlatformConfig) HasPlatform(name string) bool {
	_, ok := c.Platforms[name]
	return ok
}

// PlatformNames returns the configured platform ids in sorted order, so an
// error message that lists them reads the same way twice.
func (c *PlatformConfig) PlatformNames() []string {
	names := make([]string, 0, len(c.Platforms))
	for name := range c.Platforms {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// platformAliases maps the shorthand names people actually type to the ids
// platform.json declares. Anything that looks a platform up by name has to
// resolve the shorthand first -- otherwise the lookup silently misses and
// the caller cleans nothing while reporting that it did.
//
// Every right-hand side is a key in platform.json. That is the whole
// contract, and it was broken in one direction: the table used to map
// "claude" to "claude" while platform.json also keyed it "claude", so the
// *long* form the docs, the --ai flag help and promptExtension() all named
// -- `claude_code`, `cursor_ai`, `github_copilot` -- resolved to nothing at
// all. `plaesy platforms show claude_code` printed "Name: Unknown" and
// `plaesy clean --ai claude_code` cleaned nothing while reporting that it
// had. See TestEveryPlatformAliasResolves and
// TestTheLongFormSpellingOfEveryPlatformResolves.
//
// "claude" and "kilo" are themselves the canonical ids now (platform.json
// used to key them "claude_code"/"kilo_code"); "claude_code" and "kilo_code"
// stay here as aliases so a config, script or muscle memory that still
// types the old long form keeps resolving rather than silently breaking.
var platformAliases = map[string]string{
	"claude":      "claude",
	"claude_code": "claude",
	"anthropic":   "claude",
	"copilot":     "github_copilot",
	"github":      "github_copilot",
	"gh":          "github_copilot",
	"gh_copilot":  "github_copilot",
	"cursor":      "cursor_ai",
	"windsurf":    "windsurf_ai",
	"continue":    "continue_dev",
	"kilo":        "kilo",
	"kilo_code":   "kilo",
	"trae":        "trae_ai",
	"generic":     "generic_ai",
	"none":        "",
}

// NormalizePlatform resolves a user-supplied platform name to the id
// platform.json uses: the alias table first, then a configured id, then a
// configured display name (case-insensitively), and finally the input
// unchanged so the caller can report it as unknown.
//
// The empty string and every alias for it ("none") return "", which callers read
// as "no platform" -- that is what `plaesy init --ai none` means, and it is not
// the same as an unknown name, which must not be swallowed.
// IsKnownAlias reports whether name -- already lowercased, with dashes
// turned into underscores -- is a key platformAliases recognizes. It exists
// because NormalizePlatform's return value alone cannot tell a caller "the
// table matched something" from "nothing matched, here is your input back":
// since "claude" and "kilo" became canonical ids, their own alias entries
// now map to themselves, so a caller that infers a hit from "the output
// differs from the input" misses exactly those two, while a display name
// that merely differs in case (no table entry at all) would wrongly look
// like a hit if inferred the opposite way. See scaffold.normalizePlatform,
// the one caller that needs this distinction.
func IsKnownAlias(name string) bool {
	_, ok := platformAliases[name]
	return ok
}

func NormalizePlatform(input string, c *PlatformConfig) string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return ""
	}
	lowered := strings.ToLower(trimmed)
	if id, ok := platformAliases[lowered]; ok {
		return id
	}
	if c != nil {
		for id, p := range c.Platforms {
			if strings.EqualFold(id, trimmed) || strings.EqualFold(p.Name, trimmed) {
				return id
			}
		}
	}
	return trimmed
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

// DetectPlatform mirrors detect-platform: the first platform (in declaration
// order) whose detection patterns match a file or directory under the current
// working directory. Returns "" with no error when nothing matches (interactive
// selection trigger in the bash original).
//
// The patterns are relative to the cwd on purpose — the command answers "which
// platform is this project using", and the project is the directory you are
// standing in. configPath is the platform.json, read only for declaration
// order: a caller that passes the config's own directory here is asking the
// wrong question, and every marker would be invisible.
func (c *PlatformConfig) DetectPlatform(configPath string) (string, error) {
	names, err := c.ListPlatforms(configPath)
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
	case "agents":
		return []string{"agents/*.agents.md"}
	}
	return nil
}

// GetCleanDirs mirrors get-clean-dirs: derive cleanup directories from a
// platform's mapping (core/prompts/agents target dirs), falling back to
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
		p, ok := c.Platforms[platform]
		if !ok {
			// Falling through to the cross-platform union here would answer a
			// specific question with every platform's directories and exit 0, so
			// a typo in a platform name looked like a real, confident answer.
			return nil, fmt.Errorf("unknown platform: %s", platform)
		}
		add(p.Mapping["core"])
		add(p.Mapping["prompts"])
		add(p.Mapping["agents"])
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
//
// An unrecognized component is an error. It used to return an empty string with
// no error, which the command printed as a blank line and exited 0 over — the
// one result a caller cannot tell apart from "this component is empty".
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
	return "", fmt.Errorf("unknown plaesy structure component: %s (available: base_directory, core_directories, project_directories, memory_subdirectories)", component)
}
