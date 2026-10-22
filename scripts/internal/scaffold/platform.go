package scaffold

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/plaesy/spec-kit/internal/common"
)

// SetupPlatformConfig ports setup_platform_config(): copies agents into
// the platform-specific agent directory (e.g. .claude/agents, .opencode/agents)
// per the platform's mapping in platform.json, and then copies the
// platform-specific core file (e.g. CLAUDE.md) and prompt files
// (e.g. .claude/commands/*.md), remapping the prompt extension per platform
// (github_copilot -> .prompt.md, everything else stays .md).
func SetupPlatformConfig(home, targetDir, platform string, cfg *PlatformConfig) error {
	return setupPlatformConfig(home, targetDir, platform, cfg, copyPolicy{})
}

// setupPlatformConfig is SetupPlatformConfig with an explicit copy policy. Init
// passes the zero policy (skip existing); reload passes an overwriting one.
func setupPlatformConfig(home, targetDir, platform string, cfg *PlatformConfig, pol copyPolicy) error {
	common.LogInfo("%s", "Setting up configuration for: "+cfg.DisplayName(platform)+" based on platform.json mapping...")

	// Agents -> platform-specific directory per platform.json mapping.
	// Falls back to .plaesy/roles/ when the platform declares no agents mapping
	// (e.g. platforms that don't support custom agents natively).
	rolesDir := filepath.Join(targetDir, cfg.Plaesy.BaseDirectory, "roles")
	if entry, ok := cfg.Platforms[platform]; ok {
		if agentsDest, has := entry.Mapping["agents"]; has && agentsDest != "" {
			rolesDir = filepath.Join(targetDir, filepath.FromSlash(agentsDest))
		}
	}
	if err := copyAgents(home, rolesDir, platform, pol); err != nil {
		return err
	}

	entry, ok := cfg.Platforms[platform]
	if !ok {
		common.LogWarning("%s", "Unknown platform, skipping platform-specific file setup: "+platform)
		return nil
	}

	if corePath, ok := entry.Mapping["core"]; ok && corePath != "" {
		if err := copyPlatformCore(home, targetDir, corePath, cfg, pol); err != nil {
			return err
		}
	}

	if promptsPath, ok := entry.Mapping["prompts"]; ok && promptsPath != "" {
		if err := copyPlatformPrompts(home, targetDir, promptsPath, platform, cfg, pol); err != nil {
			return err
		}
	}

	if err := installPlatformHooks(targetDir, platform, cfg, pol); err != nil {
		return err
	}

	common.LogSuccess("AI-specific configuration completed based on platform.json mapping")
	return nil
}

// copyAgents copies <home>/agents/*.agents.md into the platform-specific
// destination directory, stripping the ".agents" segment from the destination
// filename and remapping the extension per platform.
func copyAgents(home, rolesDir, platform string, pol copyPolicy) error {
	// Determine the destination directory label for logging.
	label := filepath.ToSlash(rolesDir)
	if rel, err := filepath.Rel(filepath.Dir(rolesDir), rolesDir); err == nil {
		label = filepath.ToSlash(rel)
	}
	common.LogInfo("Copying agents to %s...", label)

	sourceDir := filepath.Join(home, "agents")

	if err := os.MkdirAll(rolesDir, 0o755); err != nil {
		return err
	}

	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		common.LogWarning("%s", "Source agents directory not found: "+sourceDir)
		return nil
	}

	// Agent extension remapping: github_copilot uses .agent.md, everything
	// else keeps .md.
	ext := agentExtension(platform)

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".agents.md") {
			continue
		}
		base := strings.TrimSuffix(e.Name(), ".agents.md")
		dst := filepath.Join(rolesDir, base+ext)
		copied, err := copyFile(filepath.Join(sourceDir, e.Name()), dst, pol)
		if err != nil {
			common.LogWarning("%s", "  ✗ "+base+ext+" (failed to copy)")
			continue
		}
		if copied {
			common.LogInfo("%s", "  ✓ "+base+ext)
		}
	}
	return nil
}

// copyPlatformCore copies the shared core file (plaesy.mapping.core.value,
// e.g. instructions/agents.instructions.md) to the platform's destination
// (e.g. CLAUDE.md, AGENTS.md), skipping if the destination already exists.
func copyPlatformCore(home, targetDir, destRelPath string, cfg *PlatformConfig, pol copyPolicy) error {
	coreMapping, ok := cfg.Plaesy.Mapping["core"]
	if !ok || coreMapping.Value == "" {
		return nil
	}

	srcFile := filepath.Join(home, filepath.FromSlash(coreMapping.Value))
	if _, err := os.Stat(srcFile); err != nil {
		common.LogWarning("%s", "Source core file not found: "+srcFile)
		return nil
	}

	dstFile := filepath.Join(targetDir, filepath.FromSlash(destRelPath))
	if dstDir := filepath.Dir(dstFile); dstDir != targetDir {
		if err := os.MkdirAll(dstDir, 0o755); err != nil {
			return err
		}
	}

	if _, err := os.Stat(dstFile); err == nil {
		common.LogInfo("%s", "Platform-specific core exists, skipping: "+destRelPath)
		return nil
	}

	if _, err := copyFile(srcFile, dstFile, pol); err != nil {
		return err
	}
	common.LogSuccess("%s", "Created platform-specific core file: "+destRelPath)
	return nil
}

// copyPlatformPrompts copies every *.md under <home>/prompts (recursively,
// preserving relative subdirectory structure) into the platform's prompts
// destination directory, remapping the extension per platform.
func copyPlatformPrompts(home, targetDir, destRelDir, platform string, cfg *PlatformConfig, pol copyPolicy) error {
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
		if _, err := copyFile(path, dst, pol); err != nil {
			common.LogWarning("%s", "  ✗ "+rel+ext+" (failed to copy)")
			return nil
		}
		common.LogSuccess("%s", "Copied prompt: "+filepath.Join(destRelDir, rel+ext))
		return nil
	})
}

// promptExtension mirrors the case statement in setup_platform_config():
// github_copilot gets .prompt.md, claude/cursor_ai/everything else
// keeps .md.
func promptExtension(platform string) string {
	switch platform {
	case "github_copilot":
		return ".prompt.md"
	default:
		return ".md"
	}
}

// agentExtension mirrors the case statement in setup_platform_config():
// github_copilot agents use .agent.md, everything else keeps .md.
func agentExtension(platform string) string {
	switch platform {
	case "github_copilot":
		return ".agent.md"
	default:
		return ".md"
	}
}
