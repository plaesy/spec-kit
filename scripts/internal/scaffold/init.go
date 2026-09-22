package scaffold

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/plaesy/spec-kit/internal/common"
)

// Options configures a call to Init.
type Options struct {
	// TargetDir is the project directory to initialize (default ".").
	TargetDir string
	// AIPlatform selects the AI platform id (e.g. "claude_code"), a display
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
	common.LogInfo("Plaesy home: " + home)
	common.LogInfo("Target directory: " + absTarget)

	cfg, err := LoadPlatformConfig(home)
	if err != nil {
		common.LogWarning(err.Error() + " (using built-in defaults)")
	}

	platform := normalizePlatform(opts.AIPlatform, cfg)
	if platform != "none" && !cfg.HasPlatform(platform) {
		return fmt.Errorf("invalid AI platform: %s (available: %v)", platform, cfg.PlatformNames())
	}
	common.LogSuccess("Selected: " + cfg.DisplayName(platform))

	if err := CreateStructure(home, absTarget, cfg); err != nil {
		return err
	}

	if platform != "none" {
		if err := SetupPlatformConfig(home, absTarget, platform, cfg); err != nil {
			return err
		}
	}

	common.LogSuccess("Plaesy Spec-Kit initialization completed!")
	common.LogInfo("Project: " + absTarget)
	common.LogInfo("Platform: " + cfg.DisplayName(platform))
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
func normalizePlatform(input string, cfg *PlatformConfig) string {
	if input == "" {
		return "none"
	}

	switch lower(input) {
	case "claude", "claude_code", "claude-code", "anthropic":
		return "claude_code"
	case "cursor", "cursor_ai", "cursor-ai":
		return "cursor_ai"
	case "github", "copilot", "github_copilot", "github-copilot", "gh-copilot":
		return "github_copilot"
	case "none", "manual", "generic", "generic_ai", "generic-ai":
		return "none"
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
