# plaesy update-agent-context

Synchronizes AI assistant context files (`CLAUDE.md`, `GEMINI.md`, Copilot
instructions, etc.) with the current feature's `plan.md` — technology stack,
recent changes, and project structure.

Source: `scripts/cmd/plaesy/update_agent.go` + `scripts/internal/agentcontext/update.go`.

## Prerequisites

- Run from a feature branch with `specs/<branch>/plan.md` present (created from `templates/plan.template.md`).
- `git` on PATH.
- No external interpreter needed — the regex-based Active Technologies /
  Recent Changes rewrite is implemented natively in Go
  (`scripts/internal/agentcontext/update.go`), unlike the old bash script
  which shelled out to `python3` for the same edit (a previously
  undocumented hard dependency). This is the one command where the Go port
  is strictly simpler than either predecessor.

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
# Update all detected context files
plaesy update-agent-context

# Update one platform
plaesy update-agent-context claude
plaesy update-agent-context copilot
```

An unknown agent-type argument exits with an error.

## How It Works

1. **Extract from `plan.md`**: `Language/Version`, `Primary Dependencies`, `Storage`, `Project Type`.
2. **New context file**: copied from `.plaesy/templates/agent-file-template.md` with placeholders filled in (project name, date, tech stack, structure, test/lint commands, language guidance).
3. **Existing context file**: `## Active Technologies` gets new language/framework/DB lines appended if not already present; `## Recent Changes` gets the current branch's change prepended and is trimmed to the last 3 entries; the `Last updated: YYYY-MM-DD` line is refreshed.

## Manual Additions Preservation

A user-editable block is preserved across regeneration:

```html
<!-- MANUAL ADDITIONS START -->
... your custom content ...
<!-- MANUAL ADDITIONS END -->
```

The command captures the block, strips it before editing the Active
Technologies / Recent Changes sections, then re-appends it after — content
between the markers always survives a regeneration.

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

Console summary:

```
- Added language: Python 3.11
- Added framework: FastAPI
- Added database: PostgreSQL
```

## Troubleshooting

**`no plan.md found`** — create `specs/<branch>/plan.md` from `templates/plan.template.md` first.

**`template not found`** — ensure `.plaesy/templates/agent-file-template.md` exists.

**Unknown agent type** — pass one of `claude|gemini|copilot|cursor|qwen|opencode`, or omit the argument to update all existing context files.

## Related Commands

- **`plaesy create-new-feature`** — creates the feature branch/spec that `plan.md` lives under (see [create-new-feature.md](./create-new-feature.md))
- **`plaesy init`** — initializes AI context files for a new project (see [plaesy-init.md](./plaesy-init.md))
- **`plaesy get-feature-paths`** — resolves the current feature's paths (see [get-feature-paths.md](./get-feature-paths.md))
