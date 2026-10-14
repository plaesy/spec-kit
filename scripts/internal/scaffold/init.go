package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/plaesy/spec-kit/internal/common"
	"github.com/plaesy/spec-kit/internal/config"
)

// Options configures a call to Init.
type Options struct {
	// TargetDir is the project directory to initialize (default ".").
	TargetDir string
	// AIPlatform selects the AI platform id (e.g. "claude"), a display
	// name, or "none" for manual development. Empty means "none" -- this
	// port does not carry over the bash script's interactive menu or
	// stack-based auto-detection (see package doc / task report).
	AIPlatform string
	// PlaesyHome overrides the Plaesy repo root lookup (--plaesy-home).
	PlaesyHome string
}

// Init ports plaesy-init.sh's main(): resolves the Plaesy repo root, loads
// platform.json, validates and normalizes the target directory, builds the
// .plaesy/ structure, and (when an AI platform was selected) copies its
// platform-specific core/prompt files.
func Init(opts Options) error {
	targetDir := opts.TargetDir
	if targetDir == "" {
		targetDir = "."
	}
	absTarget, err := filepath.Abs(targetDir)
	if err != nil {
		return fmt.Errorf("resolving target directory: %w", err)
	}
	if err := validateTargetDir(absTarget); err != nil {
		return err
	}

	home, err := FindHome(opts.PlaesyHome)
	if err != nil {
		return err
	}
	common.LogInfo("%s", "Plaesy home: "+home)
	common.LogInfo("%s", "Target directory: "+absTarget)

	cfg, err := LoadPlatformConfig(home)
	if err != nil {
		return fmt.Errorf("%w (resolved Plaesy home %q has no usable platform.json)", err, home)
	}

	platform := normalizePlatform(opts.AIPlatform, cfg)
	if platform != "none" && !cfg.HasPlatform(platform) {
		return fmt.Errorf("invalid AI platform: %s (available: %v)", platform, cfg.PlatformNames())
	}
	common.LogSuccess("%s", "Selected: "+cfg.DisplayName(platform))

	if err := CreateStructure(home, absTarget, cfg); err != nil {
		return err
	}

	if platform != "none" {
		if err := SetupPlatformConfig(home, absTarget, platform, cfg); err != nil {
			return err
		}
	}

	common.LogSuccess("Plaesy Spec-Kit initialization completed!")
	common.LogInfo("%s", "Project: "+absTarget)
	common.LogInfo("%s", "Platform: "+cfg.DisplayName(platform))
	return nil
}

// validateTargetDir ports validate_target_dir(): the directory must exist
// and be writable. Unlike the bash version this does not attempt an
// interactive "mkdir -p it for me" suggestion beyond the error message.
func validateTargetDir(targetDir string) error {
	info, err := os.Stat(targetDir)
	if err != nil {
		return fmt.Errorf("target directory does not exist: %s", targetDir)
	}
	if !info.IsDir() {
		return fmt.Errorf("target path is not a directory: %s", targetDir)
	}

	probe := filepath.Join(targetDir, ".plaesy-init-write-check")
	f, err := os.Create(probe)
	if err != nil {
		return fmt.Errorf("target directory is not writable: %s", targetDir)
	}
	f.Close()
	os.Remove(probe)

	if info, err := os.Stat(filepath.Join(targetDir, ".plaesy")); err == nil && info.IsDir() {
		common.LogWarning("Target directory already has .plaesy/ (existing files will be preserved, not overwritten)")
	}
	return nil
}

// normalizePlatform resolves common aliases (claude, cursor, copilot, none,
// manual, ...) to their canonical platform id, then falls back to matching
// against configured platform ids / display names, matching
// normalize_platform()'s behavior. Empty input normalizes to "none".
//
// The alias table itself lives in internal/config, because `plaesy clean --ai`
// needs the same resolution: a shorthand that only `init` understood is a name
// the config has no key for, and a lookup that misses removes nothing while
// reporting that it did. This function only adds what that table cannot: the
// dashed spellings, and mapping the no-platform family onto "none".
func normalizePlatform(input string, cfg *PlatformConfig) string {
	if input == "" {
		return "none"
	}

	// "none" is the sentinel for "configure no AI platform". It is checked
	// before the shared table because config.NormalizePlatform returns "" for
	// it, and SetupPlatformConfig reads the literal "none".
	switch lower(input) {
	case "none", "manual", "generic", "generic_ai", "generic-ai":
		return "none"
	}
	// The shared table takes the underscore spelling; "claude-code" is a name
	// people type and has to reach the same platform.
	//
	// "id != lower(input)" alone used to be enough to tell "the table
	// resolved something" apart from "here is your input back unchanged".
	// It stopped being enough once "claude" and "kilo" became canonical ids:
	// their own alias entries now map to themselves, so resolving "CLAUDE"
	// produces id "claude", which equals lower(input) and looks like a
	// no-op -- falling through to a case-sensitive display-name match that
	// "CLAUDE" cannot win. IsKnownAlias catches that case by asking the
	// table directly rather than inferring a hit from whether the spelling
	// changed.
	normalized := strings.ReplaceAll(lower(input), "-", "_")
	if id := config.NormalizePlatform(normalized, nil); id != lower(input) || config.IsKnownAlias(normalized) {
		return id
	}
	if cfg != nil {
		for _, id := range cfg.PlatformNames() {
			if input == id || input == cfg.DisplayName(id) {
				return id
			}
		}
	}
	return input
}

func lower(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + ('a' - 'A')
		}
	}
	return string(b)
}
