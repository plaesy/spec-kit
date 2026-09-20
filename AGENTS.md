# Plaesy Spec-Kit — Agent Instructions

## Quick Reference

| Topic | Command / Path |
|-------|----------------|
| **Primary config** | `.plaesy/instructions/plaesy.md` (read at session start) |
| **Memory index** | `.plaesy/memory.md` → `.plaesy/memory/*.md` |
| **Context (session)** | `.plaesy/context.md` (≤100 lines) |
| **Dogfood reload** | `make reload` — **deletes `.claude/`, `.plaesy/`, `CLAUDE.md` unconditionally** |
| **Run all tests** | `bash testing/bash/run.sh` • `pwsh testing/powershell/run.ps1` |
| **Smoke tests** | `bash testing/smoke/smoke-e2e.sh` • `pwsh testing/smoke/smoke-powershell.ps1` |
| **CI pipeline** | Bash lint → PowerShell lint → smoke tests (both shells) |

## Repository Structure

```
scripts/bash/         # Linux/macOS automation (26 scripts)
scripts/powershell/   # Windows/PowerShell automation (26 scripts, parity required)
.paesy/
  instructions/       # 35 instruction files + mapping.json (auto-load logic)
  memory/             # Topic knowledge files (flat, update memory.md index)
  tasks/              # Kanban: backlog/, todo/, doing/, done/, blocked/
  analysis/           # plaesy-analyze output: project.graph.json, project.html, reports.md
  context.md          # Session-only context (≤100 lines)
  state.json          # Persistent agent state
prompts/              # 9 workflow prompts (start, assess, implement, optimize, fix, doc, loop, continue, save)
chatmodes/            # 24 AI role configs (*.chatmode.md)
checklists/           # 7 quality gates (*.checklist.md)
instructions/         # 43 technology/methodology instructions (*.instructions.md)
templates/            # Project/document templates (*.template.md)
```

## Critical Conventions

### 1. Bash/PowerShell Parity (Mandatory)
**Any behavioral change must be made in both `scripts/bash/` and `scripts/powershell/`**. The two implementations must stay functionally identical. CI runs both lint suites.

### 2. Instruction Auto-Loading (`instructions/mapping.json`)
- `always_load`: 24 files loaded unconditionally (plaesy.*, quality-gates, assess-*, tasks, etc.)
- `frameworks/`: Keyword + regex detection (Next.js, NestJS, React, Angular, Flutter, Spring Boot, Rails, Office automation)
- `languages/`: Go, Java, C#, Rust, Dart
- `cross_cutting/`: Git, performance, security, DevOps, K8s, Terraform, testing, TDD, SQL, tech-validation
- `methodologies/`: Brainstorming, changelog, markdown, brandkit, redesign
- **Detection scans**: `specs/*/context.md`, `specs/*/requirements.md`, `.plaesy/analysis/project.json`, `package.json`, `go.mod`, `pom.xml`, `build.gradle`, `Cargo.toml`, `pubspec.yaml`, `Gemfile`, `requirements.txt`
- **New instruction files MUST be added to `mapping.json`** (under `always_load` or a keyword category) or they silently never install.

### 3. Memory Management
- `.plaesy/memory.md` = index only (TOC lines: `- [Title](file.md) — description`)
- `.plaesy/memory/*.md` = detailed topic files (flat, no subdirs)
- `.plaesy/context.md` = current session only (≤100 lines, archive to memory/ when full)
- Update `memory.md` index when creating new `memory/*.md` files.

### 4. Anti-Duplication Protocol
**Default: modify/extend existing code, never duplicate** unless user explicitly requests new.
- Search first: `grep -r "similar purpose"` or `find . -name "*.{sh,ps1,md}"`
- If match found → modify (add param, extend return, keep compat)
- If not extendable → ask user with found match(es)

### 5. File Naming Conventions
| Type | Suffix | Location |
|------|--------|----------|
| Prompts | `.md` | `prompts/` |
| Instructions | `.instructions.md` | `instructions/` |
| Chatmodes | `.chatmode.md` | `chatmodes/` |
| Checklists | `.checklist.md` | `checklists/` |
| Templates | `.template.md` | `templates/` / `.plaesy/templates/` |

### 6. Docs Links
All `docs/*/README.md` are at depth 2. Links to repo root need `../../`, links to sibling top-level folders need `../../x/`.

## Workflow Commands (AI Assistant)

| Command | Purpose |
|---------|---------|
| `/start <desc>` | Full workflow: constitution → research → spec → implement → QA → optimize → fix → doc → save |
| `/assess [role] [path]` | Unified assessment (research mode upfront, assessment mode after implement) |
| `/continue` | Resume from last incomplete phase |
| `/implement` | TDD implementation with smart instruction loading |
| `/optimize` | Performance + design + code optimization |
| `/fix` | Bug resolution with root cause analysis |
| `/doc` | Generate documentation from code/specs |
| `/save` | Persist session + update memory |
| `/loop` | Autonomous assess→fix→verify loop |

## Assessment Dimensions (Phase 4 & 6)
Technical, Design, Business, Product, Marketing, Legal, Financial, Operations — each scores 0-100 with findings.

## Testing Notes
- `testing/bash/run.sh` (61KB) and `testing/powershell/run.ps1` (68KB) are the main suites
- Smoke tests are fast E2E validation of `plaesy-graph` + `plaesy-analyze`
- No unit test framework — scripts are tested via syntax check + functional smoke tests
- `detect-stack` tests missing from PowerShell suite (known gap)

## Known Gotchas
- `make reload` has **no confirmation prompt** — backs up `.claude/`, `.plaesy/`, `CLAUDE.md` first
- `platform-detector.ps1` has no bash twin (gap)
- `testing/powershell/run.ps1` lacks tests for `plaesy-task-manage` and `plaesy-validate-memory`
- `docs/` had systemic broken links (depth-2 path issue) — fixed in recent audit
- `mapping.json` `scan_locations` must match what `detect-stack.sh/.ps1` actually read
- Office automation instructions (PowerPoint/Excel/Word) are in `instructions/` + `mapping.json` frameworks

## Reference Files
- `.plaesy/instructions/plaesy.md` — Global mandatory instructions (342 lines)
- `.plaesy/instructions/tasks.md` — Task management instructions
- `.plaesy/instructions/quality-gates.md` — Quality gate definitions
- `.plaesy/instructions/universal-orchestrator.md` — Phase orchestration logic
- `instructions/mapping.json` — Auto-load detection rules (228 lines)
- `CONTRIBUTING.md` — Development setup, parity rule, PR checklist