# Changelog

All notable changes to Plaesy Constitution Kit are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and
this project uses [Semantic Versioning](https://semver.org/) once it reaches 1.0.0.

## [Unreleased]

### Go CLI Migration

- Migrated from bash/PowerShell scripts to a single cross-platform Go binary
  (`plaesy`). All `scripts/bash/*.sh` and `scripts/powershell/*.ps1` files have
  been removed; their behavior is preserved 1:1 in the Go port under
  `scripts/cmd/plaesy/` + `scripts/internal/`. Built with `go 1.20` (stdlib only,
  no external HTTP/db dependencies; Cobra for CLI framework).
- `plaesy install` copies the running binary to `%LOCALAPPDATA%\Plaesy\bin`
  (Windows) or `$HOME/.local/bin` (Unix); `plaesy status` / `plaesy uninstall`
  manage the installation. `plaesy repair` / `plaesy upgrade` print "not yet
  implemented" — build a new binary from source and re-run `plaesy install`.
- `plaesy config detect-platform` auto-detects the current AI platform
  (18 supported: claude_code, opencode, cursor_ai, github_copilot, etc.).
- `plaesy detect-stack` selects relevant `instructions/*.instructions.md`
  files based on the target project's tech stack, using
  `instructions/mapping.json` as the keyword registry.
- `plaesy analyze` generates project.json, project.structure.json, overview.md,
  and the dependency graph in one run. Fingerprint-based fast path skips
  regeneration when source is unchanged (use `--force` to override).
- `plaesy graph` builds a knowledge graph with query, explain, path-query,
  impact-check, semantic-queue, and watch modes.
- `plaesy trim` provides token/context compression (run/compress/llm-queue/
  apply-llm/report subcommands).
- `plaesy init [dir] --ai <platform>` scaffolds `.plaesy/` structure +
  platform-specific config files from `scripts/configs/platform.json`.
- `plaesy inject-ai-headers` injects platform-specific YAML front-matter into
  prompt/chatmode/instructions files.
- `plaesy clean` removes Plaesy framework files with safe/thorough/complete
  levels, backup, and dry-run support.
- `plaesy task-manage` manages task lifecycle across backlog/todo/doing/done/
  blocked directories.
- `plaesy create-new-feature` creates feature branches + spec.md from templates.
- `plaesy get-feature-paths` prints shell-sourceable feature paths.
- `plaesy update-agent-context` syncs agent context files (CLAUDE.md,
  AGENTS.md, .cursor/rules, etc.) with the current feature's plan.md.
- `plaesy validate-memory` scans `.plaesy/memory/` for external references.
- `plaesy generate-image` generates images via OpenAI or Gemini APIs.

### Fixes

- Fixed Windows path resolution: `git rev-parse --show-toplevel` returns
  MSYS2-style paths (`/c/Users/...`) on Windows; `GetRepoRoot()` now normalizes
  these to native `C:\Users\...` paths so all commands work correctly.
- Fixed GitHub Actions release workflow: corrected build working-directory
  (`working-directory: scripts`) and output filename (was `$os/$arch`, now
  `$os-$arch` matching install.sh/install.ps1 download URLs).
- Added `scripts/install.ps1` — PowerShell equivalent of `install.sh` for
  downloading release binaries on Windows without requiring `sh`.
- Updated `instructions/mapping.json` to reference `plaesy init` instead of
  `plaesy-init.sh/.ps1`.

## [0.0.1] - Initial release

- Initial public commit of the Plaesy Spec-Kit framework: prompts, instructions,
  chatmodes/roles, checklists, templates, and bash/PowerShell automation scripts.
