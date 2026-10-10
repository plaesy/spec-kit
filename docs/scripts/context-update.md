# plaesy context update

Synchronizes **existing** AI assistant context files (`CLAUDE.md`, `GEMINI.md`,
Copilot instructions, etc.) with the current feature's `plan.md` — technology
stack and recent changes.

It patches, it does not scaffold. The agent-file template was removed on
2026-09-25, so the command no longer creates a context file: if the target file
does not exist, it exits non-zero and tells you to create it. A run that changed
nothing never reports success.

Source: `scripts/cmd/plaesy/context.go` + `scripts/internal/agentcontext/update.go`.

## Prerequisites

- Run from a feature branch with `.plaesy/specs/<branch>/plan.md` present (created from `templates/plan.template.md`).
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
| *(none)* | Updates every context file that already exists; if none exist, exits non-zero listing the paths it looked for |

## Quick Start

```bash

# Update all detected context files
plaesy context update

# Update one platform
plaesy context update claude
plaesy context update copilot
```

An unknown agent-type argument exits with an error.

## How It Works

1. **Extract from `plan.md`**: `Language/Version`, `Primary Dependencies`, `Storage`, `Project Type`.
2. **Missing context file**: the command stops with an error naming the path. It does not create one — there is no template to create it from, by design.
3. **Existing context file**: `## Active Technologies` gets a new
   language/framework line when the plan introduces a language not yet listed,
   plus a storage line for a new database; `## Recent Changes` gets the current
   branch's change and is trimmed to the last 3 entries; the
   `Last updated: YYYY-MM-DD` line is refreshed.

`## Recent Changes` holds **one line per branch**. Running the command again on
the same branch replaces that branch's line rather than adding a second one, so
the section is safe to re-run and a plan that changed since the last run is
reflected instead of duplicated. A plan that has not resolved its **Language**
(`NEEDS CLARIFICATION`) records only the parts that are known — the line reads
`- 001-feature-auth: Added FastAPI`, not `Added  + FastAPI`.

### Known limitation

A plan whose **language** is already listed adds nothing to `## Active Technologies`,
even when its **framework** is new — the language acts as the marker for the pair.
Tracked in `scripts/internal/agentcontext/update_test.go`
(`TestUpdateAgentFileKeepsKnownLanguageOnce`), which exists to make a change to that
behavior deliberate rather than silent.

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

```text

- Added language: Python 3.11
- Added framework: FastAPI
- Added database: PostgreSQL
```

## Troubleshooting

**`no plan.md found`** — create `.plaesy/specs/<branch>/plan.md` from `templates/plan.template.md` first.

**`no <agent> context file at <path>`** — the file does not exist. Create it (copy an
existing context file, or let your AI platform create its own) and re-run. The
command only updates files that already exist.

**Unknown agent type** — pass one of `claude|gemini|copilot|cursor|qwen|opencode`, or omit the argument to update all existing context files.

## Related Commands

- **`plaesy features create`** — creates the feature branch/spec that `plan.md` lives under (see [features.md](./features.md))
- **`plaesy init`** — initializes AI context files for a new project (see [plaesy-init.md](./plaesy-init.md))
- **`plaesy features paths`** — resolves the current feature's paths (see [features-paths.md](./features-paths.md))
