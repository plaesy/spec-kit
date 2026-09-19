---
title: "Cross-Dimensional Generalization + Path Consistency (2026-09-16)"
updatedAt: "2026-09-16T00:00:00.000Z"
---

## Trigger

User asked (in Indonesian) whether spec-kit needed instructions to make assessment
easier from documents/internet (researched firecrawl/anydoc, kepano/defuddle — both
already covered by WebFetch's built-in summarization + Read; added a small
"Document & Source Ingestion" section to `prompts/assess.md` instead of a new
instruction file). That led into a broader ask: "check semua prompt saya, saya
ingin spec-kit saya tidak hanya untuk technical project."

## Root Finding

`prompts/assess.md` was already cross-dimensional (technical/design/business/
marketing/legal/financial/management/product). The other 7 execution prompts
(`start`, `continue`, `loop`, `implement`, `optimize`, `doc`, `fix`) assumed
`project == codebase` as an axiom — not a partial bias, a structural one.
`start.md`'s constitution generation was the root: it fed code-only defaults
(90% coverage, <200ms) into every downstream prompt regardless of project type.

## Fixes Applied (all in `prompts/*.md` source, not installed copies)

1. **`start.md`**: Phase 0 now detects project type first (software/business/
   legal/marketing/design/product/mixed) and generates per-dimension constitution
   defaults instead of always code metrics.
2. **`continue.md`**: state machine + auto-routing now reads the constitution's
   active dimension(s) instead of hardcoding `/implement`'s TDD cycle for
   everything.
3. **`loop.md`**: Phase 1 no longer hardcodes `/assess:technical`; all 6
   Autonomous Decision Rules got a non-software variant.
4. **`implement.md`**: added a "Non-Code Deliverable Cycle" (Draft-Review-Revise)
   parallel to TDD for non-software dimensions.
5. **`optimize.md`**: performance-targets table marked software-only; added
   "Optimization Targets by Other Dimension" table.
6. **`doc.md`**: added a Handoff Spec path for non-code deliverables (points at
   `/assess`'s Design Spine table instead of file:line source citation).
7. **`fix.md`**: defect classification extended to non-software types (unsourced
   claim, untraceable clause, broken assumption, brand inconsistency); legal
   escalation pattern generalized into an explicit cross-dimension template.
8. **`instructions/rust.instructions.md`**: was completely empty (0 bytes) —
   written to the same depth/pattern as go/java instructions.
9. **Threshold consistency**: `assess.md`'s Auto-Decision Rules had `<85%` test
   coverage while every other file in the repo says 90% — fixed to 90%.
   `optimize.md`'s "≥80% (no reduction)" clarified to "≥90% target, or ≥
   pre-optimization baseline if already below 90%".

## Syntax Consistency Fix (user-flagged)

User noticed `/optimize --design` used dash-flag syntax while `/assess:technical`
already established colon-scope syntax — asked why, wanted `:design` everywhere.
Standardized: `:{dimension}` always selects one of the same 8 dimensions as
`/assess:{scope}`; `--focus {aspect}` only narrows within a dimension (e.g.
`/optimize:technical --focus performance`), never selects the dimension itself.
Added a "Usage Format" block to `optimize.md`, `fix.md`, `implement.md` (matching
`assess.md`'s existing one); `loop.md` gained `/loop:{dimension}` replacing its
old `--focus {dimension}` flag.

## Path-Reference Bug (user-flagged, second round)

User asked why a prompt said "see `prompts/assess.md`" when that path doesn't
exist post-install. Checked `scripts/configs/platform.json` +
`scripts/bash/plaesy-init.sh`: `instructions/` and `chatmodes/` mirror to a
stable platform-agnostic path (`.plaesy/instructions/`, `.plaesy/roles/`)
regardless of AI tool, but **`prompts/` does not** — it maps straight to a
platform-specific folder (`.claude/commands`, `.opencode/prompts`, `.cursor/
rules`, `.github/prompts`, etc., all different). There is no `.plaesy/prompts/`.

Web research (Catalin's Tech "AI Slash Commands" guide, VS Code Prompt Files
docs) confirms the fix: cross-platform slash-command systems abstract the
command name from its file location — reference by `/assess`, never by the
source file path. Fixed all 16 `prompts/assess.md`/`prompts/start.md`
references across the 9 prompt files to plain command names.

Also fixed 5 secondary instances of the same bug class: `checklists/X.md` and
`templates/X.md` referenced without the `.plaesy/` prefix, even though those two
(unlike `prompts/`) DO mirror stably to `.plaesy/checklists/` and
`.plaesy/templates/` — same root cause (source path used without accounting for
install-time relocation), smaller blast radius since the filename itself was
still correct.

## Established Rule (new)

When writing any `prompts/*.md` content that needs to point at another prompt's
behavior, reference it by command name (`/assess`, `/start`) — never by its
source file path (`prompts/assess.md`). `instructions/`, `chatmodes/`,
`templates/`, and `checklists/` DO have a stable post-install path under
`.plaesy/` and may be referenced that way (with the `.plaesy/` prefix and, for
instructions only, the `.instructions` suffix stripped) — see
[[assess-dimension-coverage]] for the instructions-specific filename-transform
rule this extends.

## Status

All fixes applied and verified via `git status` (only `prompts/*.md` +
`instructions/rust.instructions.md` touched, never `.claude/commands/` or
`.plaesy/`). Not yet committed — user has not asked for a commit this session.
