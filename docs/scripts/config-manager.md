# plaesy config

Centralized configuration management for Plaesy Spec-Kit platforms and AI
assistants. Reads `scripts/configs/platform.json` to detect the active AI
platform, resolve file mapping targets, and compute cleanup paths.

Source: `scripts/cmd/plaesy/config.go` (config subcommand wiring) +
`scripts/internal/config/config.go` — a single Go implementation, replacing
the previous bash + PowerShell pair.

## Purpose

- Resolve per-platform mapping targets (where core instructions, prompts, agents get installed)
- Compute cleanup files/directories for a platform (used by `plaesy clean`)
- Validate `platform.json` syntax

## Commands

These six read `scripts/configs/platform.json` and ask things about **the
file**. The four platform questions — which platform am I on, which exist, what
does one look like, what is its config value — moved to `plaesy platforms`;
see [platforms.md](./platforms.md).

| Command | Purpose |
|---------|---------|
| `plaesy config get-mapping <section> <mapping_type>` | Get mapping value with value/excludes structure |
| `plaesy config get-excludes <section> <mapping_type>` | Get exclude patterns for a mapping |
| `plaesy config get-clean-files [platform]` | Get files to clean for platform |
| `plaesy config get-clean-dirs [platform]` | Get directories to clean for platform |
| `plaesy config get-structure <component>` | Get Plaesy structure config |
| `plaesy config validate` | Validate platform configuration |

Run `plaesy config --help` or `plaesy config <command> --help` for the exact,
current flags on each subcommand.

## get-clean-dirs behavior

For the target platform, `plaesy config get-clean-dirs` resolves
`mapping.core`, `mapping.prompts`, and `mapping.agents`, takes the
directory portion of each resolved target, and dedupes them into a list. If
the named platform yields no directories, it falls back to computing core
directories across *all* platforms via the platform list — this matches the
original bash/PowerShell behavior exactly (ported 1:1, including the fixed
`get-platform-config`-delegation behavior the bash version was patched to
use).

A platform name the config does not declare is an **error**, not a fallback.
The fallback answers "this platform has no directories of its own"; using it
for a typo made `get-clean-dirs cluade` return the union of every platform's
directories and exit 0.

Shorthand names are resolved to the ids `platform.json` declares
(`claude_code` → `claude`, `copilot` → `github_copilot`, `cursor` →
`cursor_ai`, `windsurf` → `windsurf_ai`, `generic` → `generic_ai`), so the
spellings used throughout these docs work on every subcommand that takes a
platform name — `get-clean-dirs` here, and `platforms get` /
`platforms show` over in [platforms.md](./platforms.md). A configured id or a
display name works too. An unrecognized name is reported as an error rather
than resolved to a plausible-looking platform.

`get-structure` likewise errors on a component it does not know,
listing the four it does, rather than printing an empty line — a blank line
and exit 0 is indistinguishable from a component whose value is empty.

`get-clean-files` is **not** derived from JSON — it returns a hardcoded
fallback list (`CLAUDE.md`, `.cursorrules`, `.github/copilot-instructions.md`)
regardless of platform, since cleanup file configuration doesn't yet exist in
`platform.json`. This is unchanged from the original scripts.

## Usage examples

```bash

# Get a mapping value (handles value/excludes structure)
plaesy config get-mapping claude prompts
# Result: .claude/prompts/

# Get exclude patterns for a mapping
plaesy config get-excludes claude core

# Get files to clean for a platform
plaesy config get-clean-files claude

# Get directories to clean for a platform
plaesy config get-clean-dirs claude

# Get a Plaesy structure component
plaesy config get-structure core_directories

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
        "agents": ".claude/agents/",
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

- **`plaesy init`** — calls into the same `internal/config` package to determine where to install core/prompts/agents
- **`plaesy clean`** — calls `get-clean-files` and `get-clean-dirs` for platform-specific cleanup
- **`plaesy platforms detect`** — detects the active platform via the same detection logic

## File locations

- Config data: `scripts/configs/platform.json`
- Implementation: `scripts/internal/config/config.go` (single Go implementation, Go standard library `encoding/json` only)

## Troubleshooting

```bash

# Platform not detected -> check detection files exist in the project
plaesy platforms detect

# Invalid JSON -> check syntax in platform.json
plaesy config validate

# Missing key -> verify the key exists in platform.json
plaesy platforms get claude missing_key
```
