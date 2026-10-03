package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/plaesy/spec-kit/internal/common"
)

// writeConfig writes body to <dir>/platform.json and returns the path.
func writeConfig(t *testing.T, dir, body string) string {
	t.Helper()
	path := filepath.Join(dir, "platform.json")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

// chdir switches the process working directory for the duration of the test.
// DetectPlatform resolves its detection patterns with os.Stat against the
// process cwd, so the cwd is the only seam available (Go 1.20 has no
// t.Chdir). Callers must not use t.Parallel.
func chdir(t *testing.T, dir string) {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(prev); err != nil {
			t.Errorf("restore cwd to %s: %v", prev, err)
		}
	})
}

// platformJSON is a compact but realistic platform.json: an object-form
// plaesy.mapping entry with excludes and description, a bare-string
// platforms.*.mapping, one platform with no detection patterns (the fallback
// platform) and two that could both match.
const platformJSON = `{
  "version": "1.0",
  "description": "test fixture",
  "plaesy": {
    "base_directory": ".plaesy",
    "core_directories": ["memory", "instructions", "tasks"],
    "project_directories": [".plaesy/specs"],
    "mapping": {
      "core": {"value": "instructions/agents.instructions.md"},
      "memory": {"value": "memory/*", "excludes": [], "description": "memory files"},
      "instructions": {"value": "instructions/*", "excludes": ["agents.instructions.md"]}
    }
  },
  "platforms": {
    "zeta_first": {
      "name": "Zeta",
      "provider": "ZetaCorp",
      "category": "cli_tools",
      "detection": ["zeta.marker"],
      "mapping": {"core": "ZETA.md", "prompts": ".zeta/commands"}
    },
    "alpha_second": {
      "name": "Alpha",
      "provider": "AlphaCorp",
      "category": "ide_tools",
      "detection": ["alpha.marker", ".alpha"],
      "mapping": {"core": "ALPHA.md", "instructions": ".alpha/instructions", "agents": ".alpha/agents"}
    },
    "no_detection": {
      "name": "Fallback",
      "provider": "Various",
      "category": "fallback",
      "detection": [],
      "mapping": {"core": "AI-INSTRUCTIONS.md"}
    }
  }
}`

// fallbackBody is a config whose only mapping with a directory component is a
// platform's "core", so GetCleanDirs' cross-platform fallback has something to
// find. The other platforms contribute nothing (bare filenames), which keeps
// the expected result deterministic despite Go's randomized map iteration.
const fallbackBody = `{
  "platforms": {
    "alpha": {"mapping": {"core": "alpha-dir/ALPHA.md", "prompts": "AI-INSTRUCTIONS.md"}},
    "beta":  {"mapping": {"core": "BETA.md"}},
    "gamma": {"mapping": {"core": "GAMMA.md"}}
  }
}`

func loadFixture(t *testing.T, body string) (*PlatformConfig, string) {
	t.Helper()
	dir := t.TempDir()
	path := writeConfig(t, dir, body)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return cfg, path
}

func TestMappingEntryUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    MappingEntry
		wantErr bool
	}{
		{
			name: "bare string (platforms.*.mapping shape)",
			in:   `".claude/instructions"`,
			want: MappingEntry{Value: ".claude/instructions"},
		},
		{
			name: "object with value and excludes (plaesy.mapping shape)",
			in:   `{"value": "instructions/*", "excludes": ["a.md", "b.md"]}`,
			want: MappingEntry{Value: "instructions/*", Excludes: []string{"a.md", "b.md"}},
		},
		{
			name: "object with description",
			in:   `{"value": "roles/*", "excludes": [], "description": "agents"}`,
			want: MappingEntry{Value: "roles/*", Excludes: []string{}, Description: "agents"},
		},
		{
			name: "empty object",
			in:   `{}`,
			want: MappingEntry{},
		},
		{
			name: "null yields zero value",
			in:   `null`,
			want: MappingEntry{},
		},
		{
			name:    "number rejected",
			in:      `42`,
			wantErr: true,
		},
		{
			name:    "array rejected",
			in:      `["a"]`,
			wantErr: true,
		},
		{
			name:    "object with wrong field type rejected",
			in:      `{"value": 5}`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got MappingEntry
			err := got.UnmarshalJSON([]byte(tt.in))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error for %s, got %+v", tt.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("UnmarshalJSON(%s): %v", tt.in, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		missing  bool
		wantErr  string
		wantPath bool // the error message should name the config path
		check    func(t *testing.T, cfg *PlatformConfig)
	}{
		{
			name: "full fixture",
			body: platformJSON,
			check: func(t *testing.T, cfg *PlatformConfig) {
				if cfg.Version != "1.0" {
					t.Errorf("version = %q, want 1.0", cfg.Version)
				}
				if cfg.Description != "test fixture" {
					t.Errorf("description = %q", cfg.Description)
				}
				if cfg.Plaesy.BaseDirectory != ".plaesy" {
					t.Errorf("base_directory = %q", cfg.Plaesy.BaseDirectory)
				}
				if got, want := cfg.Plaesy.CoreDirectories, []string{"memory", "instructions", "tasks"}; !reflect.DeepEqual(got, want) {
					t.Errorf("core_directories = %v, want %v", got, want)
				}
				if got, want := cfg.Plaesy.ProjectDirectories, []string{".plaesy/specs"}; !reflect.DeepEqual(got, want) {
					t.Errorf("project_directories = %v, want %v", got, want)
				}
				if got := cfg.Plaesy.Mapping["memory"]; got.Description != "memory files" {
					t.Errorf("description field not parsed: %+v", got)
				}
				if got := cfg.Platforms["zeta_first"].Detection; !reflect.DeepEqual(got, []string{"zeta.marker"}) {
					t.Errorf("detection = %v", got)
				}
				if got := cfg.Platforms["zeta_first"].Mapping["core"]; got != "ZETA.md" {
					t.Errorf("platform mapping = %q, want ZETA.md", got)
				}
			},
		},
		{
			name: "minimal object",
			body: `{}`,
			check: func(t *testing.T, cfg *PlatformConfig) {
				// Absent "platforms" key leaves the map nil (json does not
				// initialize it); only a present-but-empty object decodes to
				// an allocated map.
				if cfg.Platforms != nil {
					t.Errorf("platforms = %v, want nil for a config without the key", cfg.Platforms)
				}
				if cfg.Plaesy.Mapping != nil {
					t.Errorf("plaesy.mapping = %v, want nil for a config without the key", cfg.Plaesy.Mapping)
				}
			},
		},
		{
			name: "empty platforms object",
			body: `{"platforms": {}}`,
			check: func(t *testing.T, cfg *PlatformConfig) {
				if cfg.Platforms == nil {
					t.Error("platforms should be an allocated empty map")
				}
			},
		},
		{
			name:    "empty file",
			body:    ``,
			wantErr: "invalid JSON syntax in platform configuration",
		},
		{
			name:    "malformed json",
			body:    `{"version": "1.0",}`,
			wantErr: "invalid JSON syntax in platform configuration",
		},
		{
			name:    "wrong top level type",
			body:    `[]`,
			wantErr: "invalid JSON syntax in platform configuration",
		},
		{
			name:     "missing file",
			missing:  true,
			wantErr:  "platform configuration file not found",
			wantPath: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "platform.json")
			if !tt.missing {
				path = writeConfig(t, dir, tt.body)
			}

			cfg, err := Load(path)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected an error containing %q, got config %+v", tt.wantErr, cfg)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %q, want it to contain %q", err, tt.wantErr)
				}
				if tt.wantPath && !strings.Contains(err.Error(), path) {
					t.Errorf("error should name the path, got %q", err)
				}
				if cfg != nil {
					t.Error("config must be nil on error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			tt.check(t, cfg)
		})
	}
}

func TestListPlatforms(t *testing.T) {
	t.Run("preserves declaration order", func(t *testing.T) {
		cfg, path := loadFixture(t, platformJSON)
		got, err := cfg.ListPlatforms(path)
		if err != nil {
			t.Fatalf("ListPlatforms: %v", err)
		}
		want := []string{"zeta_first", "alpha_second", "no_detection"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("ListPlatforms = %v, want declaration order %v", got, want)
		}
	})

	// A nested "platforms" object is picked up before the top-level one.
	// The ordered scan looks for the first string token equal to the section
	// name anywhere in the document rather than only at depth 1, so keys from
	// a nested object are mixed in. platform.json never nests the name today,
	// which is the only reason this has not bitten. Asserted as-is: the test
	// documents the limitation so a fix to orderedTopLevelKeys is visible.
	t.Run("matches the first section of that name at any depth", func(t *testing.T) {
		body := `{
		  "wrapper": {"platforms": {"nested": {}}},
		  "platforms": {"only": {"name": "Only"}}
		}`
		cfg, path := loadFixture(t, body)
		got, err := cfg.ListPlatforms(path)
		if err != nil {
			t.Fatalf("ListPlatforms: %v", err)
		}
		if !reflect.DeepEqual(got, []string{"nested", "only"}) {
			t.Errorf("ListPlatforms = %v, want [nested only] (nested keys are not skipped)", got)
		}
	})

	t.Run("falls back to sorted keys when the section is absent", func(t *testing.T) {
		// The file has no "platforms" key at all, so the ordered scan yields
		// nothing and ListPlatforms sorts the map keys instead.
		body := `{"version": "1.0", "description": "no platforms section"}`
		path := writeConfig(t, t.TempDir(), body)
		cfg := &PlatformConfig{
			Platforms: map[string]Platform{"b": {}, "a": {}, "c": {}},
		}
		got, err := cfg.ListPlatforms(path)
		if err != nil {
			t.Fatalf("ListPlatforms: %v", err)
		}
		if !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
			t.Errorf("ListPlatforms = %v, want sorted [a b c]", got)
		}
	})

	t.Run("falls back to sorted keys on a truncated document", func(t *testing.T) {
		// Truncated right after the section name: the ordered scan hits EOF
		// looking for the section's '{' and returns what it has (nothing).
		path := writeConfig(t, t.TempDir(), `{"platforms"`)
		cfg := &PlatformConfig{Platforms: map[string]Platform{"b": {}, "a": {}}}
		got, err := cfg.ListPlatforms(path)
		if err != nil {
			t.Fatalf("ListPlatforms: %v", err)
		}
		if !reflect.DeepEqual(got, []string{"a", "b"}) {
			t.Errorf("ListPlatforms = %v, want sorted [a b]", got)
		}
	})

	t.Run("empty platform map yields an empty list", func(t *testing.T) {
		cfg, path := loadFixture(t, `{"platforms": {}}`)
		got, err := cfg.ListPlatforms(path)
		if err != nil {
			t.Fatalf("ListPlatforms: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("ListPlatforms = %v, want empty", got)
		}
	})

	t.Run("unreadable path errors", func(t *testing.T) {
		cfg, _ := loadFixture(t, platformJSON)
		if _, err := cfg.ListPlatforms(filepath.Join(t.TempDir(), "missing.json")); err == nil {
			t.Fatal("expected an error for a missing platform.json")
		}
	})
}

func TestGetPlatform(t *testing.T) {
	cfg, _ := loadFixture(t, platformJSON)

	t.Run("known platform", func(t *testing.T) {
		p, err := cfg.GetPlatform("alpha_second")
		if err != nil {
			t.Fatalf("GetPlatform: %v", err)
		}
		if p.Name != "Alpha" {
			t.Errorf("name = %q, want Alpha", p.Name)
		}
	})

	t.Run("unknown platform", func(t *testing.T) {
		p, err := cfg.GetPlatform("nope")
		if err == nil {
			t.Fatalf("expected an error, got %+v", p)
		}
		if err.Error() != "unknown platform: nope" {
			t.Errorf("error = %q, want %q", err, "unknown platform: nope")
		}
		if !reflect.DeepEqual(p, Platform{}) {
			t.Errorf("platform = %+v, want the zero value on error", p)
		}
	})
}

func TestGetPlatformConfig(t *testing.T) {
	cfg, _ := loadFixture(t, platformJSON)
	tests := []struct {
		name     string
		platform string
		key      string
		want     string
		wantErr  bool
	}{
		{name: "name field", platform: "zeta_first", key: "name", want: "Zeta"},
		{name: "provider field", platform: "zeta_first", key: "provider", want: "ZetaCorp"},
		{name: "category field", platform: "zeta_first", key: "category", want: "cli_tools"},
		{name: "mapping.core", platform: "zeta_first", key: "mapping.core", want: "ZETA.md"},
		{name: "mapping.prompts", platform: "zeta_first", key: "mapping.prompts", want: ".zeta/commands"},
		{name: "mapping key absent for the platform", platform: "zeta_first", key: "mapping.agents", want: ""},
		{name: "unknown key returns empty", platform: "zeta_first", key: "version", want: ""},
		{name: "empty key returns empty", platform: "zeta_first", key: "", want: ""},
		{name: "mapping with empty suffix", platform: "zeta_first", key: "mapping.", want: ""},
		{name: "unknown platform errors", platform: "nope", key: "name", wantErr: true},
		{name: "unknown platform errors even for mapping keys", platform: "nope", key: "mapping.core", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cfg.GetPlatformConfig(tt.platform, tt.key)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetPlatformConfig: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetMappingValue(t *testing.T) {
	cfg, _ := loadFixture(t, platformJSON)
	tests := []struct {
		name    string
		section string
		want    string
	}{
		{name: "value only", section: "core", want: "instructions/agents.instructions.md"},
		{name: "value with excludes", section: "memory", want: "memory/*"},
		{name: "missing section yields empty and no error", section: "prompts", want: ""},
		{name: "platform mapping is not consulted", section: "zeta_first", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cfg.GetMappingValue(tt.section)
			if err != nil {
				t.Fatalf("GetMappingValue: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetMappingExcludes(t *testing.T) {
	cfg, _ := loadFixture(t, platformJSON)
	tests := []struct {
		name    string
		section string
		want    []string
	}{
		{name: "excludes present", section: "instructions", want: []string{"agents.instructions.md"}},
		{name: "empty excludes is an empty non-nil slice", section: "memory", want: []string{}},
		{name: "no excludes key at all", section: "core", want: nil},
		{name: "missing section yields nil", section: "prompts", want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfg.GetMappingExcludes(tt.section)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestDetectPlatform(t *testing.T) {
	tests := []struct {
		name  string
		files []string // created in the cwd the detection runs from
		dirs  []string
		want  string
	}{
		{name: "no markers detects nothing", want: ""},
		{name: "first declaration wins", files: []string{"zeta.marker", "alpha.marker"}, want: "zeta_first"},
		{name: "later declaration when only it matches", files: []string{"alpha.marker"}, want: "alpha_second"},
		{name: "any pattern in the list matches", files: []string{".alpha"}, want: "alpha_second"},
		{name: "directory markers count", dirs: []string{".alpha"}, want: "alpha_second"},
		{name: "platform without detection patterns is never detected", files: []string{"AI-INSTRUCTIONS.md"}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := writeConfig(t, dir, platformJSON)
			for _, f := range tt.files {
				if err := os.WriteFile(filepath.Join(dir, f), []byte("x"), 0o644); err != nil {
					t.Fatalf("seed %s: %v", f, err)
				}
			}
			for _, d := range tt.dirs {
				if err := os.MkdirAll(filepath.Join(dir, d), 0o755); err != nil {
					t.Fatalf("seed dir %s: %v", d, err)
				}
			}
			chdir(t, dir)

			cfg, err := Load(path)
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			got, err := cfg.DetectPlatform(path)
			if err != nil {
				t.Fatalf("DetectPlatform: %v", err)
			}
			if got != tt.want {
				t.Errorf("DetectPlatform = %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("unreadable config errors", func(t *testing.T) {
		dir := t.TempDir()
		chdir(t, dir)
		cfg, err := Load(writeConfig(t, t.TempDir(), platformJSON))
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if _, err := cfg.DetectPlatform(filepath.Join(dir, "missing.json")); err == nil {
			t.Fatal("expected an error when platform.json cannot be read")
		}
	})
}

// TestDetectPlatformResolvesAgainstCwd pins where the detection patterns are
// looked up: the current working directory, not the directory holding
// platform.json. The config path is read only to recover the declaration order
// of the platform keys, so the two markers below are found in the first case
// and not the second.
//
// This reads like a bug — the argument is called `path`, and the markers sit
// right next to it. It is not one: `plaesy config detect-platform` answers
// "which platform is this project using", and the project is the directory the
// user is standing in. The parameter is named configPath now so that the next
// reader does not "fix" it into searching configs/.
func TestDetectPlatformResolvesAgainstCwd(t *testing.T) {
	configDir := t.TempDir()
	path := writeConfig(t, configDir, platformJSON)
	if err := os.WriteFile(filepath.Join(configDir, "zeta.marker"), []byte("x"), 0o644); err != nil {
		t.Fatalf("seed marker: %v", err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Same config, cwd holding the marker: detected.
	chdir(t, configDir)
	got, err := cfg.DetectPlatform(path)
	if err != nil {
		t.Fatalf("DetectPlatform: %v", err)
	}
	if got != "zeta_first" {
		t.Fatalf("DetectPlatform from the config dir = %q, want zeta_first", got)
	}

	// Same config, unrelated cwd: the marker is invisible even though the
	// path argument still points at the config that declares it.
	other := t.TempDir()
	chdir(t, other)
	got, err = cfg.DetectPlatform(path)
	if err != nil {
		t.Fatalf("DetectPlatform: %v", err)
	}
	if got != "" {
		t.Errorf("DetectPlatform from an unrelated cwd = %q, want \"\" (patterns resolve against the cwd)", got)
	}
}

func TestGetCleanFiles(t *testing.T) {
	want := []string{"CLAUDE.md", ".cursorrules", ".github/copilot-instructions.md"}
	got := GetCleanFiles()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetCleanFiles = %v, want %v", got, want)
	}
	if got2 := GetCleanFiles(); &got2[0] == &got[0] {
		t.Error("GetCleanFiles must return a fresh slice the caller cannot use to mutate later calls")
	}
}

func TestGetSourceRules(t *testing.T) {
	tests := []struct {
		name     string
		fileType string
		want     []string
	}{
		{
			name: "core", fileType: "core",
			want: []string{"instructions/plaesy.instructions.md"},
		},
		{
			name: "instructions", fileType: "instructions",
			want: []string{"instructions/*.instructions.md", "plaesy.instructions.md"},
		},
		{name: "prompts", fileType: "prompts", want: []string{"prompts/*.prompt.md"}},
		{name: "agents", fileType: "agents", want: []string{"agents/*.agents.md"}},
		{name: "unknown type yields nil", fileType: "roles", want: nil},
		{name: "empty type yields nil", fileType: "", want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetSourceRules(tt.fileType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSourceRules(%q) = %v, want %v", tt.fileType, got, tt.want)
			}
		})
	}
}

func TestGetCleanDirs(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		platform string
		want     []string
		wantErr  string
	}{
		{
			name: "core prompts and agents targets",
			body: platformJSON, platform: "alpha_second",
			want: []string{".alpha"},
		},
		{
			name: "core target without a directory component is skipped",
			body: platformJSON, platform: "zeta_first",
			want: []string{".zeta"},
		},
		{
			// A name the config does not declare used to fall through to the
			// cross-platform union, so `get-clean-dirs typo` answered a specific
			// question with every platform's directories and exited 0. A wrong
			// answer that looks confident is worse than a refusal.
			name:     "unknown platform is an error, not the cross-platform union",
			body:     fallbackBody,
			platform: "not_a_platform",
			wantErr:  "unknown platform: not_a_platform",
		},
		{
			name: "empty platform name also uses the fallback",
			body: fallbackBody, platform: "",
			want: []string{"alpha-dir"},
		},
		{
			name: "platform yielding nothing falls back to every core mapping",
			body: `{"platforms": {
			  "a": {"mapping": {"core": "A.md", "prompts": "null", "agents": ""}},
			  "b": {"mapping": {"core": "b-dir/A.md"}}
			}}`,
			platform: "a",
			want:     []string{"b-dir"},
		},
		{
			name:    "no usable mapping anywhere",
			body:    `{"platforms": {"a": {"mapping": {"core": "A.md"}}}}`,
			wantErr: "no clean directories found",
		},
		{
			name:    "no platforms at all",
			body:    `{"platforms": {}}`,
			wantErr: "no clean directories found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, path := loadFixture(t, tt.body)
			cfg, err := Load(path)
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			got, err := cfg.GetCleanDirs(tt.platform)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected an error containing %q, got %v", tt.wantErr, got)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("error = %q, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetCleanDirs: %v", err)
			}
			if len(tt.want) == 0 {
				if len(got) != 0 {
					t.Errorf("GetCleanDirs = %v, want none", got)
				}
				return
			}
			// Map iteration order is unspecified for the fallback scan, so
			// compare as a set.
			if len(got) != len(tt.want) {
				t.Fatalf("GetCleanDirs = %v, want %v", got, tt.want)
			}
			for _, w := range tt.want {
				found := false
				for _, g := range got {
					if g == w {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GetCleanDirs = %v, missing %q", got, w)
				}
			}
		})
	}
}

// TestGetCleanDirsDeduplicates asserts a directory shared by two mapping
// targets is listed once.
func TestGetCleanDirsDeduplicates(t *testing.T) {
	body := `{"platforms": {"a": {"mapping": {
	  "core": "shared/A.md", "prompts": "shared/p.md", "agents": "shared/agents.md"
	}}}}`
	_, path := loadFixture(t, body)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	got, err := cfg.GetCleanDirs("a")
	if err != nil {
		t.Fatalf("GetCleanDirs: %v", err)
	}
	if !reflect.DeepEqual(got, []string{"shared"}) {
		t.Errorf("GetCleanDirs = %v, want [shared] exactly once", got)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		missing bool
		wantErr string
	}{
		{name: "valid config", body: platformJSON},
		{name: "valid but empty config", body: `{}`},
		{name: "invalid json", body: `{`, wantErr: "invalid JSON syntax in platform configuration"},
		{name: "missing file", missing: true, wantErr: "platform configuration file not found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "platform.json")
			if !tt.missing {
				path = writeConfig(t, dir, tt.body)
			}
			err := Validate(path)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected an error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}

func TestPlaesyStructureComponent(t *testing.T) {
	cfg, _ := loadFixture(t, platformJSON)
	tests := []struct {
		name      string
		component string
		want      string
		wantErr   bool
	}{
		{name: "core directories are space joined", component: "core_directories", want: "memory instructions tasks"},
		{name: "project directories are space joined", component: "project_directories", want: ".plaesy/specs"},
		{name: "memory subdirectories are not modeled", component: "memory_subdirectories", want: ""},
		{name: "base directory is a single value", component: "base_directory", want: ".plaesy"},
		// "unknown yields empty" was the old contract, and it was the wrong one:
		// the command printed a blank line and exited 0, which a caller cannot
		// tell apart from a component whose value genuinely is empty.
		{name: "unknown component is an error", component: "nope", wantErr: true},
		{name: "empty component is an error", component: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cfg.PlaesyStructureComponent(tt.component)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("PlaesyStructureComponent(%q) = %q, want an error", tt.component, got)
				}
				// The message has to name the component and the alternatives, or
				// the user cannot tell a typo from a missing feature.
				if !strings.Contains(err.Error(), "component") {
					t.Errorf("error %q does not say what was wrong", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("PlaesyStructureComponent: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("empty array joins to an empty string", func(t *testing.T) {
		empty := &PlatformConfig{}
		got, err := empty.PlaesyStructureComponent("core_directories")
		if err != nil || got != "" {
			t.Errorf("got (%q, %v), want (\"\", nil)", got, err)
		}
	})
}

// A project created by `plaesy init` keeps its scripts tree under .plaesy, so
// platform.json is at .plaesy/scripts/configs/platform.json there and not at
// scripts/configs/platform.json. Resolving only the second left every
// scaffolded project unable to run `plaesy platforms` or `plaesy config`.
//
// defaultConfigPath takes the root so both shapes can be built in a temp dir.
// The old test could not do this: it ran DefaultConfigPath against whatever
// repository it happened to be in, which is this one, where the hardcoded path
// is correct by construction — so the case that was broken was the one case it
// could not express.
func TestDefaultConfigPathPicksTheLayoutTheRootActuallyIs(t *testing.T) {
	t.Run("source tree", func(t *testing.T) {
		root := t.TempDir()
		want := filepath.Join(root, "scripts", "configs", "platform.json")
		if got := defaultConfigPath(root); got != want {
			t.Errorf("defaultConfigPath = %q, want %q for a root with no .plaesy", got, want)
		}
	})

	t.Run("scaffolded project", func(t *testing.T) {
		root := t.TempDir()
		if err := os.MkdirAll(filepath.Join(root, ".plaesy", "scripts", "configs"), 0o755); err != nil {
			t.Fatalf("scaffold: %v", err)
		}
		want := filepath.Join(root, ".plaesy", "scripts", "configs", "platform.json")
		got := defaultConfigPath(root)
		if got != want {
			t.Errorf("defaultConfigPath = %q, want %q for a scaffolded project", got, want)
		}
		// The path has to be loadable, not merely plausible: the bug was that
		// it named a file that was not there.
		if err := os.WriteFile(got, []byte(`{"platforms":{}}`), 0o644); err != nil {
			t.Fatalf("seed platform.json: %v", err)
		}
		if _, err := Load(got); err != nil {
			t.Errorf("the resolved path is not loadable: %v", err)
		}
	})
}

// In this repository the default must keep working, and must still point at a
// file that exists — the source-tree layout is the one the tests and the
// `--fix` and `--write-baseline` paths all run against.
func TestDefaultConfigPathInThisRepositoryExists(t *testing.T) {
	if _, err := common.GetRepoRoot(); err != nil {
		t.Skipf("not inside a git repository: %v", err)
	}
	got, err := DefaultConfigPath()
	if err != nil {
		t.Fatalf("DefaultConfigPath: %v", err)
	}
	if _, err := os.Stat(got); err != nil {
		t.Errorf("the default path should exist in this repository: %v", err)
	}
}

// A shorthand that only one command understood is a name the config has no key
// for. `plaesy init` resolved aliases locally; `plaesy clean --ai claude` did
// not, so it accepted the name, printed it as the chosen platform, and then
// matched no mapping — a claude-specific clean that removed nothing and a plan
// that reported nothing to remove. The alias table therefore lives here, where
// both can reach it.
func TestNormalizePlatformResolvesShorthandToAConfigKey(t *testing.T) {
	// The shipped platform.json, not a fixture: the property that matters is
	// that every alias lands on a key the real config declares. A fixture would
	// let the table and the config drift apart unnoticed.
	cfg, err := Load(filepath.Join("..", "..", "configs", "platform.json"))
	if err != nil {
		t.Fatalf("the shipped platform.json must load: %v", err)
	}

	aliases := map[string]string{
		"claude":      "claude",
		"claude_code": "claude",
		"anthropic":   "claude",
		"copilot":     "github_copilot",
		"github":      "github_copilot",
		"cursor":      "cursor_ai",
		"windsurf":    "windsurf_ai",
		"continue":    "continue_dev",
		"kilo":        "kilo",
		"kilo_code":   "kilo",
		"trae":        "trae_ai",
		"generic":     "generic_ai",
	}
	for input, want := range aliases {
		got := NormalizePlatform(input, cfg)
		if got != want {
			t.Errorf("NormalizePlatform(%q) = %q, want %q", input, got, want)
			continue
		}
		// The whole point: a resolved name must be a real key, or the lookup
		// that consumes it silently misses.
		if !cfg.HasPlatform(got) {
			t.Errorf("NormalizePlatform(%q) resolved to %q, which platform.json does not declare", input, got)
		}
	}

	// A canonical id passes through untouched rather than being re-resolved.
	for _, id := range []string{"claude", "github_copilot", "windsurf_ai", "generic_ai"} {
		if got := NormalizePlatform(id, cfg); got != id {
			t.Errorf("NormalizePlatform(%q) = %q, want it unchanged", id, got)
			continue
		}
		if !cfg.HasPlatform(id) {
			t.Errorf("HasPlatform(%q) is false for an id the config declares", id)
		}
	}

	// "none" is not a platform id and must not become one.
	if got := NormalizePlatform("none", cfg); got != "" {
		t.Errorf(`NormalizePlatform("none") = %q, want "" so callers read it as "no platform"`, got)
	}
}

func TestNormalizePlatformAcceptsIdsAndDisplayNames(t *testing.T) {
	cfg, err := Load(writeConfig(t, t.TempDir(), platformJSON))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct{ in, want string }{
		{"zeta_first", "zeta_first"},
		{"Zeta", "zeta_first"},
		{"alpha_second", "alpha_second"},
		{"Alpha", "alpha_second"},
		{"Fallback", "no_detection"},
		// Surrounding whitespace is not part of the name; a copied flag value
		// often carries it.
		{"  zeta_first  ", "zeta_first"},
		{"ZETA_FIRST", "zeta_first"},
	}
	for _, tc := range cases {
		if got := NormalizePlatform(tc.in, cfg); got != tc.want {
			t.Errorf("NormalizePlatform(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNormalizePlatformLeavesUnknownNamesAlone(t *testing.T) {
	cfg, err := Load(writeConfig(t, t.TempDir(), platformJSON))
	if err != nil {
		t.Fatal(err)
	}
	// An unknown name must survive so the caller can reject it by name. Silently
	// resolving it to "generic" would turn a typo into a generic_ai clean, which
	// is the more destructive of the two.
	for _, input := range []string{"mystery", "claud", "zed"} {
		if got := NormalizePlatform(input, cfg); got != input {
			t.Errorf("NormalizePlatform(%q) = %q, want it unchanged so the caller can reject it", input, got)
		}
		if cfg.HasPlatform(NormalizePlatform(input, cfg)) {
			t.Errorf("an unknown name resolved to a configured platform: %q", input)
		}
	}
	// A nil config is the "config could not be loaded" path and must not panic.
	if got := NormalizePlatform("claude", nil); got != "claude" {
		t.Errorf("NormalizePlatform with a nil config = %q", got)
	}
	if got := NormalizePlatform("mystery", nil); got != "mystery" {
		t.Errorf("NormalizePlatform(%q, nil) = %q", "mystery", got)
	}
}

func TestNormalizePlatformTreatsEmptyAndNoneAsNoPlatform(t *testing.T) {
	cfg, err := Load(writeConfig(t, t.TempDir(), platformJSON))
	if err != nil {
		t.Fatal(err)
	}
	// "no platform" is a real answer, distinct from "unknown": it is what
	// `plaesy init --ai none` means, and it must not be reported as a typo.
	for _, input := range []string{"", "  ", "none", "NONE"} {
		got := NormalizePlatform(input, cfg)
		if got != "" && got != "none" {
			t.Errorf("NormalizePlatform(%q) = %q, want no platform", input, got)
		}
	}
}

func TestPlatformNamesIsSortedAndHasPlatformAgrees(t *testing.T) {
	cfg, err := Load(writeConfig(t, t.TempDir(), platformJSON))
	if err != nil {
		t.Fatal(err)
	}
	names := cfg.PlatformNames()
	want := []string{"alpha_second", "no_detection", "zeta_first"}
	if !reflect.DeepEqual(names, want) {
		t.Errorf("PlatformNames() = %v, want %v", names, want)
	}
	for _, name := range names {
		if !cfg.HasPlatform(name) {
			t.Errorf("HasPlatform(%q) is false for a listed platform", name)
		}
	}
	if cfg.HasPlatform("mystery") {
		t.Error("HasPlatform is true for a platform the config does not declare")
	}
}
