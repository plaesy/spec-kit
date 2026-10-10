---
description: "Decision rule for when a new capability becomes an instruction file vs. a Claude Code Skill"
applyTo: "instruction-authoring, capability-design"
---

# Skills Packaging Convention

## Purpose

Documents when a new Plaesy capability should be packaged as:

- an **instruction file** (`*.instructions.md`) — always-loaded or stack-detected
  via `mapping.json`, appropriate when needed on essentially every task
- a **Claude Code Skill** (`SKILL.md`) — loaded only on invocation, appropriate
  when narrow, large, or triggered by request-shape rather than filesystem
  signals

## Decision Rule

| Capability fits… | Package as | Why |
|---|---|---|
| Needed on essentially every task | instruction file | Always-loaded cost is justified by universal relevance |
| Reliably triggered by `mapping.json` stack detection (file extension, framework keyword, config filename) | instruction file, stack-detected | Loads only when stack matches; zero always-loaded cost |
| Narrow (e.g., one specific tool or workflow) | Skill | Body loads only on invocation; description stays cheap in context |
| Large (substantial reference material) | Skill | Avoids always-loaded token cost while remaining discoverable |
| Triggered by request-shape, not filesystem signals | Skill | `mapping.json` cannot detect request-shape triggers |

## SKILL.md Shape

A Plaesy Skill follows Claude Code's `SKILL.md` convention:

```yaml

---
name: "descriptive-skill-name"
description: "One-line description always visible in context"
disable-model-invocation: false  # true if the skill is reference-only
user-invocable: true             # true if the user should trigger it directly
allowed-tools: ["Read", "Grep", "Edit"]  # restrict to minimum needed
---
```

- **`description`** — stays in context always; make it specific enough to
  distinguish from other skills.
- **`disable-model-invocation`** — set `true` for reference-only skills
  (e.g., a glossary or style guide the agent shouldn't auto-invoke).
- **`user-invocable`** — set `false` for skills that only fire as subagents.
- **`allowed-tools`** — restrict to the minimum toolset the skill needs.

## What This Does Not Change

- No existing instruction file is moved or converted — this is a
  forward-only decision rule for new capabilities.
- `plaesy init` scaffolding is untouched — Skills are a Claude Code
  host-level mechanism; Plaesy's `.instructions.md` mechanism remains the
  portable, platform-agnostic distribution layer.
- Adopting an actual `.claude/skills/` directory in this repo is a separate
  structural decision, not implied by this document.

## See Also

- `.plaesy/memory/agentic-best-practices-research-2026-09-25.md` — research part 1, §2
- `.plaesy/memory/research-consolidation-priority-plan-2026-09-25.md` — Tier 1, backport D
