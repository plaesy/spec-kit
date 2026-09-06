# Update Agent Context Script

Synchronizes AI assistant context files (`CLAUDE.md`, `GEMINI.md`, Copilot instructions, etc.) with the current feature's `plan.md` — technology stack, recent changes, and project structure.

Available in two implementations:
- `scripts/bash/update-agent-context.sh`
- `scripts/powershell/update-agent-context.ps1`

## Prerequisites

- Run from a feature branch with `specs/<branch>/plan.md` present (created by `/plan`).
- `git` on PATH.
- **Bash version only**: requires `python3` on PATH — it shells out to a Python heredoc to perform the Active Technologies / Recent Changes regex replacement (see `update_agent_file()` in the script). This is an undocumented but hard dependency; the script will fail without `python3` available.
- **PowerShell version**: no external interpreter needed — it uses native `-replace` / `[regex]::Replace` for the same edits.

## Supported Agent Types

| Value | Context File |
|---|---|
| `claude` | `CLAUDE.md` |
| `gemini` | `GEMINI.md` |
| `copilot` | `.github/copilot-instructions.md` |
| `cursor` | `.cursor/rules/specify-rules.mdc` |
| `qwen` | `QWEN.md` |
| `opencode` | `AGENTS.md` |
| *(none)* | Updates every context file that already exists; if none exist, creates `CLAUDE.md` |

## Quick Start

```bash
# Bash — update all detected context files
./scripts/bash/update-agent-context.sh

# Bash — update one platform
./scripts/bash/update-agent-context.sh claude
./scripts/bash/update-agent-context.sh copilot
```

```powershell
# PowerShell — update all detected context files
./scripts/powershell/update-agent-context.ps1

# PowerShell — update one platform
./scripts/powershell/update-agent-context.ps1 claude
./scripts/powershell/update-agent-context.ps1 copilot
```

Unknown agent-type argument exits with an error in both versions (bash: exit 1 with usage message; PowerShell: `Write-Error` with usage message).

## How It Works

1. **Extract from `plan.md`**: `Language/Version`, `Primary Dependencies`, `Storage`, `Project Type` (bash also reads these via `grep`/`sed`; PowerShell via `Select-String` in `Get-PlanValue`).
2. **New context file**: copied from `.plaesy/templates/agent-file-template.md` with placeholders filled in (project name, date, tech stack, structure, test/lint commands, language guidance).
3. **Existing context file**: `## Active Technologies` gets new language/framework/DB lines appended if not already present; `## Recent Changes` gets the current branch's change prepended and is trimmed to the last 3 entries; the `Last updated: YYYY-MM-DD` line is refreshed.

## Manual Additions Preservation

Both implementations preserve a user-editable block across regeneration:

```html
<!-- MANUAL ADDITIONS START -->
... your custom content ...
<!-- MANUAL ADDITIONS END -->
```

- **Bash** (`update_agent_file`): captures the line range between the markers with `grep -n` before the Python rewrite, then after rewriting strips any stale marker block from the output and re-appends the captured content.
- **PowerShell** (`Update-AgentFile`): matches `(?s)<!-- MANUAL ADDITIONS START -->.*?<!-- MANUAL ADDITIONS END -->` with a regex, removes it from `$content` before editing the Active Technologies / Recent Changes sections, then appends the captured block back after the edits.

Both scripts implement this equivalently — content between the markers always survives a regeneration in either version.

## Output Example

```markdown
## Active Technologies
- Python 3.11 + FastAPI (001-feature-auth)
- PostgreSQL (001-feature-auth)

## Recent Changes
- 001-feature-auth: Added Python 3.11 + FastAPI
- 002-feature-api: Added SQLAlchemy + Pydantic

<!-- MANUAL ADDITIONS START -->
... preserved user content ...
<!-- MANUAL ADDITIONS END -->

Last updated: 2026-08-09
```

Console summary printed by both versions:

```
- Added language: Python 3.11
- Added framework: FastAPI
- Added database: PostgreSQL
```

## Troubleshooting

**`No plan.md found`** — run `/plan` first to generate `specs/<branch>/plan.md`.

**`Template not found`** — ensure `.plaesy/templates/agent-file-template.md` exists.

**Bash fails with `python3: command not found`** — install Python 3, or use the PowerShell version instead (no Python dependency).

**Unknown agent type** — pass one of `claude|gemini|copilot|cursor|qwen|opencode`, or omit the argument to update all existing context files.

## Related Scripts

- `create-new-feature` — creates the feature branch/spec that `plan.md` lives under.
- `plaesy-init` — initializes AI context files for a new project.
- `get-feature-paths` — resolves the current feature's paths.
