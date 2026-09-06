# Plaesy Config Manager

Centralized configuration management for Plaesy Spec-Kit platforms and AI assistants. Reads `scripts/configs/platform.json` to detect the active AI platform, resolve file mapping targets, and compute cleanup paths.

Two equivalent implementations:
- `scripts/bash/config-manager.sh` — pure bash + awk/grep/sed, no external dependencies
- `scripts/powershell/config-manager.ps1` — PowerShell with built-in `ConvertFrom-Json`

## Purpose

- Detect which AI assistant platform (Claude Code, GitHub Copilot, Cursor, etc.) a project is using
- Resolve per-platform mapping targets (where core instructions, prompts, chatmodes get installed)
- Compute cleanup files/directories for a platform (used by `plaesy-clean`)
- Validate `platform.json` syntax

## Commands

| Command | Purpose |
|---------|---------|
| `detect-platform` | Auto-detect current AI platform from `platforms.<name>.detection` patterns |
| `list-platforms` | List all platform names defined in `platform.json` |
| `get-platform-config <platform> <key>` | Get a config value for a platform (supports dotted keys like `mapping.core`) |
| `get-mapping-value <section> <mapping_type>` | Get a mapping target (handles both old plain-string and new `{value, excludes}` structures) |
| `get-mapping-excludes <section> <mapping_type>` | Get the `excludes` array for a mapping entry |
| `get-clean-files [platform]` | Get files to remove for a platform (currently a fixed fallback list, not read from JSON) |
| `get-clean-dirs [platform]` | Get directories to remove for a platform, derived from `mapping.core`/`mapping.prompts`/`mapping.chatmodes` |
| `get-plaesy-structure <component>` | Get a `plaesy` section value (e.g. `core_directories`, `project_directories`, `base_directory`) |
| `show-platform-info [platform]` | Print name/provider/category for a platform |
| `validate` | Validate `platform.json` syntax |
| `help` | Show usage |

## get-clean-dirs behavior (verified from source)

Both `get-clean-dirs` (bash) and `Get-CleanDirs` (PowerShell) work the same way: for the target platform, they call `get-platform-config`/`Get-PlatformConfigValue` on `mapping.core`, `mapping.prompts`, and `mapping.chatmodes`, take the directory portion of each resolved target (`dirname`/`Split-Path -Parent`), and dedupe them into a space-separated list. If no directories are found for the platform, it falls back to computing core directories across *all* platforms via `list-platforms`.

Bash previously had a bug: `get-clean-dirs` called a broken 3-level nested JSON lookup directly instead of delegating to `get-platform-config`. It has been fixed to call `get-platform-config "$platform" "mapping.core"` (etc.), which correctly handles nested key lookups via its `awk`-based parser. This doc reflects the fixed, current behavior — `get-clean-dirs` is no longer implemented as its own JSON parser.

`get-clean-files` / `Get-CleanFiles` are **not** derived from JSON — both implementations return a hardcoded fallback list (`CLAUDE.md .cursorrules .github/copilot-instructions.md`) regardless of platform, since cleanup file configuration doesn't yet exist in `platform.json`.

## Usage examples (bash)

```bash
# Detect current AI platform
./scripts/bash/config-manager.sh detect-platform

# List all available platforms
./scripts/bash/config-manager.sh list-platforms

# Get a platform config value (supports dotted/nested keys)
./scripts/bash/config-manager.sh get-platform-config claude mapping.core
# Result: CLAUDE.md

# Get a mapping value (handles value/excludes structure)
./scripts/bash/config-manager.sh get-mapping-value claude prompts
# Result: .claude/prompts/

# Get exclude patterns for a mapping
./scripts/bash/config-manager.sh get-mapping-excludes claude core

# Get files to clean for a platform
./scripts/bash/config-manager.sh get-clean-files claude

# Get directories to clean for a platform
./scripts/bash/config-manager.sh get-clean-dirs claude

# Get a Plaesy structure component
./scripts/bash/config-manager.sh get-plaesy-structure core_directories

# Show detailed platform info
./scripts/bash/config-manager.sh show-platform-info claude

# Validate platform.json syntax
./scripts/bash/config-manager.sh validate
```

## Usage examples (PowerShell)

The PowerShell version takes the command as a positional argument, with `-Platform` and `-Key` as named parameters:

```powershell
# Detect current AI platform
.\scripts\powershell\config-manager.ps1 detect-platform

# List all available platforms
.\scripts\powershell\config-manager.ps1 list-platforms

# Get a platform config value (supports dotted keys)
.\scripts\powershell\config-manager.ps1 get-platform-config -Platform claude -Key mapping.core

# Get a mapping value
.\scripts\powershell\config-manager.ps1 get-mapping-value -Platform claude -Key prompts

# Get exclude patterns for a mapping
.\scripts\powershell\config-manager.ps1 get-mapping-excludes -Platform claude -Key core

# Get files to clean for a platform
.\scripts\powershell\config-manager.ps1 get-clean-files -Platform claude

# Get directories to clean for a platform
.\scripts\powershell\config-manager.ps1 get-clean-dirs -Platform claude

# Get a Plaesy structure component
.\scripts\powershell\config-manager.ps1 get-plaesy-structure -Key core_directories

# Show detailed platform info
.\scripts\powershell\config-manager.ps1 show-platform-info -Platform claude

# Validate platform.json syntax
.\scripts\powershell\config-manager.ps1 validate
```

## Configuration structure

```json
{
  "platforms": {
    "claude": {
      "name": "Claude Code",
      "provider": "Anthropic",
      "category": "ai_assistant",
      "detection": [".claude", "CLAUDE.md"],
      "mapping": {
        "core": "CLAUDE.md",
        "prompts": ".claude/prompts/",
        "chatmodes": ".claude/chatmodes/",
        "instructions": ".claude/instructions/"
      }
    }
  },
  "plaesy": {
    "base_directory": ".plaesy",
    "core_directories": ["memory"],
    "project_directories": ["docs", "specs"]
  }
}
```

## Integration with other scripts

- **plaesy-init** — calls `detect-platform` and `get-mapping-value` to determine where to install core/prompts/chatmodes
- **plaesy-clean** — calls `get-clean-files` and `get-clean-dirs` for platform-specific cleanup
- **update-agent-context** — calls `detect-platform` to determine active platform

## File locations

- Primary config: `scripts/configs/platform.json`
- Bash implementation: `scripts/bash/config-manager.sh` (no external dependencies)
- PowerShell implementation: `scripts/powershell/config-manager.ps1` (uses built-in `ConvertFrom-Json`)

## Troubleshooting

```bash
# Platform not detected -> check detection files exist in the project
./scripts/bash/config-manager.sh detect-platform

# Invalid JSON -> check syntax in platform.json
./scripts/bash/config-manager.sh validate

# Missing key -> verify the key exists in platform.json
./scripts/bash/config-manager.sh get-platform-config claude missing_key
```
