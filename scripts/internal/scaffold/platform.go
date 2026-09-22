package scaffold

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/plaesy/spec-kit/internal/common"
)

// SetupPlatformConfig ports setup_platform_config(): copies chatmodes into
// <targetDir>/.plaesy/roles/ (platform-agnostic) and then, per the
// platform's mapping in platform.json, copies the platform-specific core
// file (e.g. CLAUDE.md) and prompt files (e.g. .claude/commands/*.md),
// remapping the prompt extension per platform (github_copilot -> .prompt.md,
// everything else stays .md).
func SetupPlatformConfig(home, targetDir, platform string, cfg *PlatformConfig) error {
	common.LogInfo("Setting up configuration for: " + cfg.DisplayName(platform) + " based on platform.json mapping...")

	// Chatmodes -> .plaesy/roles/, always at the .plaesy base dir regardless
	// of platform (mirrors the bash script hardcoding ".plaesy/roles").
	if err := copyChatmodes(home, filepath.Join(targetDir, cfg.Plaesy.BaseDirectory, "roles")); err != nil {
		return err
	}

	entry, ok := cfg.Platforms[platform]
	if !ok {
		common.LogWarning("Unknown platform, skipping platform-specific file setup: " + platform)
		return nil
	}

	if corePath, ok := entry.Mapping["core"]; ok && corePath != "" {
		if err := copyPlatformCore(home, targetDir, corePath, cfg); err != nil {
			return err
		}
	}

	if promptsPath, ok := entry.Mapping["prompts"]; ok && promptsPath != "" {
		if err := copyPlatformPrompts(home, targetDir, promptsPath, platform, cfg); err != nil {
			return err
		}
	}

	common.LogSuccess("AI-specific configuration completed based on platform.json mapping")
	return nil
}

// copyChatmodes copies <home>/chatmodes/*.chatmode.md into rolesDir,
// stripping the ".chatmode" segment from the destination filename.
func copyChatmodes(home, rolesDir string) error {
	common.LogInfo("Copying chatmodes to .plaesy/roles/...")
	sourceDir := filepath.Join(home, "chatmodes")

	if err := os.MkdirAll(rolesDir, 0o755); err != nil {
		return err
	}

	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		common.LogWarning("Source chatmodes directory not found: " + sourceDir)
		return nil
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".chatmode.md") {
			continue
		}
		base := strings.TrimSuffix(e.Name(), ".chatmode.md")
		dst := filepath.Join(rolesDir, base+".md")
		copied, err := copyFile(filepath.Join(sourceDir, e.Name()), dst)
		if err != nil {
			common.LogWarning("  ✗ " + base + ".md (failed to copy)")
			continue
		}
		if copied {
			common.LogInfo("  ✓ " + base + ".md")
		}
	}
	return nil
}

// copyPlatformCore copies the shared core file (plaesy.mapping.core.value,
// e.g. instructions/agents.instructions.md) to the platform's destination
// (e.g. CLAUDE.md, AGENTS.md), skipping if the destination already exists.
func copyPlatformCore(home, targetDir, destRelPath string, cfg *PlatformConfig) error {
	coreMapping, ok := cfg.Plaesy.Mapping["core"]
	if !ok || coreMapping.Value == "" {
		return nil
	}

	srcFile := filepath.Join(home, filepath.FromSlash(coreMapping.Value))
	if _, err := os.Stat(srcFile); err != nil {
		common.LogWarning("Source core file not found: " + srcFile)
		return nil
	}

	dstFile := filepath.Join(targetDir, filepath.FromSlash(destRelPath))
	if dstDir := filepath.Dir(dstFile); dstDir != targetDir {
		if err := os.MkdirAll(dstDir, 0o755); err != nil {
			return err
		}
	}

	if _, err := os.Stat(dstFile); err == nil {
		common.LogInfo("Platform-specific core exists, skipping: " + destRelPath)
		return nil
	}

	if _, err := copyFile(srcFile, dstFile); err != nil {
		return err
	}
	common.LogSuccess("Created platform-specific core file: " + destRelPath)
	return nil
}

// copyPlatformPrompts copies every *.md under <home>/prompts (recursively,
// preserving relative subdirectory structure) into the platform's prompts
// destination directory, remapping the extension per platform.
func copyPlatformPrompts(home, targetDir, destRelDir, platform string, cfg *PlatformConfig) error {
	promptMapping, ok := cfg.Plaesy.Mapping["prompts"]
	if !ok || promptMapping.Value == "" {
		return nil
	}
	sourceDir := filepath.Join(home, filepath.FromSlash(strings.TrimSuffix(promptMapping.Value, "/*")))

	destDir := filepath.Join(targetDir, filepath.FromSlash(destRelDir))
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	ext := promptExtension(platform)

	return filepath.WalkDir(sourceDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		rel, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return nil
		}
		rel = strings.TrimSuffix(rel, ".md")
		dst := filepath.Join(destDir, rel+ext)

		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if _, err := copyFile(path, dst); err != nil {
			common.LogWarning("  ✗ " + rel + ext + " (failed to copy)")
			return nil
		}
		common.LogSuccess("Copied prompt: " + filepath.Join(destRelDir, rel+ext))
		return nil
	})
}

// promptExtension mirrors the case statement in setup_platform_config():
// github_copilot gets .prompt.md, claude_code/cursor_ai/everything else
// keeps .md.
func promptExtension(platform string) string {
	switch platform {
	case "github_copilot":
		return ".prompt.md"
	default:
		return ".md"
	}
}
