---
title: "Decision: Extract bloated always_load instructions into scope_load siblings"
description: "Split self-contained sub-topics out of always_load files instead of letting them grow; buggy/fictional content gets deleted outright"
updatedAt: "2026-10-03T17:10:00Z"
decided: "2026-10-03"
status: "Accepted"
---

# Decision: Extract bloated always_load instructions into scope_load siblings

**Status**: Accepted
**Decided**: 2026-10-03
**Deciders**: session agent
**Related**: none

## Context and Problem Statement

`instructions/*.instructions.md` files listed in `mapping.json`'s `always_load`
are loaded into every agent session regardless of task — unlike `scope_load`,
which loads on a detected trigger. Three of the eight `always_load` files had
grown large (`plaesy.instructions.md` 587 lines, `error-recovery.instructions.md`
792 lines, `universal-orchestrator.instructions.md` 446 lines) through organic
accretion: merged files, duplicated content between siblings, and — in
`universal-orchestrator.md`'s case — a whole section describing CLI flags
(`--learn-from`, `--predict`) that were never implemented anywhere in the
codebase (confirmed via repo-wide grep, zero hits outside that one file).

## Decision Drivers

- Every line in an `always_load` file is a token cost paid on every single
  session, whether or not that content is relevant to the task at hand.
- A self-contained sub-topic (e.g. predictive error-recovery pattern
  detection) doesn't need to be *in context by default* — it needs to be
  *reachable* when its trigger fires. `scope_load` already exists for this.
- Content describing a nonexistent feature is not "aspirational
  documentation" in an instructions file — an agent reading it may try to
  invoke a flag that errors, or report a capability the project does not have.

## Considered Options

### Option 1: Leave files as-is, rely on `plaesy trim compress`

Measured result on this repo's own instruction files: 0–1.7% size reduction
(see `.plaesy/memory/instructions-maintenance.md`). The prose here is already
dense — there is little filler for a heuristic compressor to strip. Rejected:
doesn't move the number that matters (bytes paid every session).

### Option 2: Extract self-contained sections to new `scope_load` files

Move a section that already describes itself as a distinct mode/topic
(predictive error recovery, the 9-phase workflow ladder, multi-agent
orchestration patterns) into its own file, register it in `mapping.json`'s
`scope_load` array, and add a trigger row to `plaesy.instructions.md`'s
"Shared Protocol Files" table so it stays reachable.

**Pros**: the `always_load` file shrinks by exactly the content that wasn't
needed every session; nothing is lost, just relocated behind a trigger.
**Cons**: adds a file; requires updating the trigger table in the same change
or the content becomes unreachable (this is enforced by
`mapping.json`'s own `always_load_note`).

### Option 3: Delete content that isn't backed by real implementation

For content that doesn't just belong elsewhere but doesn't exist at all
(the `--learn-from`/`--predict` section, an "Integration Template" and
"Recovery Metrics" section in `error-recovery.md` with zero references
anywhere else in the repo) — delete outright rather than relocate.

## Decision Outcome

**Chosen option**: Option 2 for self-contained real content, Option 3 for
fictional/unreferenced content. Applied to:

- `error-recovery.instructions.md`: 792 → 378 lines. Predictive section
  extracted to `error-recovery-predictive.instructions.md` (`scope_load`,
  triggered by `mode: both`/`mode: predictive`). Unused "Prevention &
  Self-Healing", "Recovery Metrics", and "Integration Template" sections
  deleted (zero references found anywhere else in the repo).
- `plaesy.instructions.md`: 587 → 382 lines. "Workflow Phases" and
  "Multi-Agent Orchestration Patterns" extracted to their own `scope_load`
  files; duplicate "Context Window Management" content merged into one
  section.
- `universal-orchestrator.instructions.md`: 446 → 140 lines. Fictional
  `--learn-from`/`--predict` section deleted; duplicated per-dimension
  descriptions (already canonical in `dimension-mapping.md`) replaced with a
  pointer; six near-identical "Workflow Commands" subsections and two long
  narrative "Integration Examples" condensed to one of each.
- `tdd-enforcement.instructions.md` (`scope_load`, not `always_load`, but the
  single largest file in the repo at 1049 lines): ~21 near-identical
  multi-language (JS/Java/Python/Go/Rust/C#) code examples of the same AAA
  pattern, repeated across 5 sections, reduced to one JS/TS example per
  section with a pointer to the existing "Language-Specific TDD
  Considerations" table. 1049 → 693 lines.

**Rationale**: `always_load` cost is paid unconditionally; `scope_load` cost
is paid only when the trigger fires. Moving self-contained content from the
former to the latter is a pure win with no information loss. Deleting
unreferenced/fictional content is a pure win with no loss at all.

**Accepted risks**: every file edited under `instructions/` must be
re-synced to `scripts/internal/assets/data/` (`go run ./internal/assets/gen`)
and this repo's own `.plaesy/instructions/` must be refreshed (`plaesy
reload`, and `plaesy reload --ai claude` for the `.claude/commands/` mirror)
before `go test ./...` is green — this repo dogfoods itself, so editing the
source without reloading the installed copy produces real (not spurious)
test failures. See `.plaesy/memory/instructions-maintenance.md`.

## Constitutional Compliance & Quality Gates

- Framework alignment: reduces per-session token cost for `always_load`
  content without reducing capability — every relocated section stays
  reachable via its trigger row.
- [x] Maintainability improved — removed 3-4x duplicated procedure
  restatements and one systemic dangling-reference bug class (see
  `.plaesy/memory/instructions-maintenance.md`).
- [ ] Performance requirements met (N/A — documentation change)
- [ ] Security review passed (N/A — documentation change)

## References

- `scripts/internal/scaffold/copy.go:244-310` — `copyInstructions`, the
  mechanism that reads `mapping.json`'s `always_load`/`scope_load` arrays.
- `instructions/mapping.json` — `always_load_note` (the enforcement rule:
  moving a file out of `always_load` requires a trigger row in the same
  change).
- `instructions/plaesy.instructions.md` → "Shared Protocol Files" (the
  trigger table this decision added rows to).

---

**Document Control**: Extract bloated always_load instructions · v1.0 · Created 2026-10-03

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 1.0 | 2026-10-03 | session agent | Initial version |
