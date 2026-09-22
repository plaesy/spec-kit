# plaesy config

Centralized configuration management for Plaesy Spec-Kit platforms and AI
assistants. Reads `scripts/configs/platform.json` to detect the active AI
platform, resolve file mapping targets, and compute cleanup paths.

Source: `scripts/cmd/plaesy/main.go` (config subcommand wiring) +
`scripts/internal/config/config.go` — a single Go implementation, replacing
the previous bash + PowerShell pair.

## Purpose

- Detect which AI assistant platform (Claude Code, GitHub Copilot, Cursor, etc.) a project is using
- Resolve per-platform mapping targets (where core instructions, prompts, chatmodes get installed)
- Compute cleanup files/directories for a platform (used by `plaesy clean`)
- Validate `platform.json` syntax

## Commands

| Command | Purpose |
|---------|---------|
| `plaesy config detect-platform` | Auto-detect current AI platform from `platforms.<name>.detection_patterns` |
| `plaesy config list-platforms` | List all platform names defined in `platform.json` |
| `plaesy config get-platform-config <platform> <key>` | Get a config value for a platform (supports dotted keys like `mapping.core`) |
| `plaesy config get-mapping-value <section> <mapping_type>` | Get a mapping target (handles both old plain-string and new `{value, excludes}` structures) |
| `plaesy config get-mapping-excludes <section> <mapping_type>` | Get the `excludes` array for a mapping entry |
| `plaesy config get-clean-files [platform]` | Get files to remove for a platform (a fixed fallback list, not read from JSON) |
| `plaesy config get-clean-dirs [platform]` | Get directories to remove for a platform, derived from `mapping.core`/`mapping.prompts`/`mapping.chatmodes` |
| `plaesy config get-plaesy-structure <component>` | Get a `plaesy` section value (e.g. `core_directories`, `project_directories`, `base_directory`) |
| `plaesy config show-platform-info [platform]` | Print name/provider/category for a platform |
| `plaesy config validate` | Validate `platform.json` syntax |

Run `plaesy config --help` or `plaesy config <command> --help` for the exact,
current flags on each subcommand.

## get-clean-dirs behavior

For the target platform, `plaesy config get-clean-dirs` resolves
`mapping.core`, `mapping.prompts`, and `mapping.chatmodes`, takes the
directory portion of each resolved target, and dedupes them into a list. If
no directories are found for the platform, it falls back to computing core
directories across *all* platforms via the platform list — this matches the
original bash/PowerShell behavior exactly (ported 1:1, including the fixed
`get-platform-config`-delegation behavior the bash version was patched to
use).

`get-clean-files` is **not** derived from JSON — it returns a hardcoded
fallback list (`CLAUDE.md`, `.cursorrules`, `.github/copilot-instructions.md`)
regardless of platform, since cleanup file configuration doesn't yet exist in
`platform.json`. This is unchanged from the original scripts.

## Usage examples

```bash
# Detect current AI platform
plaesy config detect-platform

# List all available platforms
plaesy config list-platforms

# Get a platform config value (supports dotted/nested keys)
plaesy config get-platform-config claude mapping.core
# Result: CLAUDE.md

# Get a mapping value (handles value/excludes structure)
plaesy config get-mapping-value claude prompts
# Result: .claude/prompts/

# Get exclude patterns for a mapping
plaesy config get-mapping-excludes claude core

# Get files to clean for a platform
plaesy config get-clean-files claude

# Get directories to clean for a platform
plaesy config get-clean-dirs claude

# Get a Plaesy structure component
plaesy config get-plaesy-structure core_directories

# Show detailed platform info
plaesy config show-platform-info claude

# Validate platform.json syntax
plaesy config validate
```

By default `plaesy config` reads `scripts/configs/platform.json` under the
repo root; override the location with `--config <path>`.

## Configuration structure

```json
{
  "platforms": {
    "claude": {
      "name": "Claude Code",
      "provider": "Anthropic",
      "category": "ai_assistant",
      "detection_patterns": [".claude", "CLAUDE.md"],
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

## Integration with other commands

- **`plaesy init`** — calls into the same `internal/config` package to determine where to install core/prompts/chatmodes
- **`plaesy clean`** — calls `get-clean-files` and `get-clean-dirs` for platform-specific cleanup
- **`plaesy update-agent-context`** — detects the active platform via the same detection logic

## File locations

- Config data: `scripts/configs/platform.json`
- Implementation: `scripts/internal/config/config.go` (single Go implementation, Go standard library `encoding/json` only)

## Troubleshooting

```bash
# Platform not detected -> check detection files exist in the project
plaesy config detect-platform

# Invalid JSON -> check syntax in platform.json
plaesy config validate

# Missing key -> verify the key exists in platform.json
plaesy config get-platform-config claude missing_key
```
