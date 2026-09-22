---
title: "Command Surface & Architecture Decisions"
updatedAt: "2026-09-16T00:00:00.000Z"
---

# Command Surface & Architecture Decisions

## Current prompts (11 total, in `prompts/`)
`start`, `assess`, `implement`, `optimize`, `improve`, `loop`, `fix`, `continue`, `save`, `doc`, `generate`.

## Repo-wide fix, round 3: multi-platform awareness + docs/ + templates/ (2026-09-16)
User reminded that Plaesy Spec-Kit supports 20 AI platforms (`scripts/configs/
platform.json` → `platforms.*`), not just Claude Code — `.claude/commands` is one
of 20 different `prompts` destinations (`.opencode/prompts`, `.cursor/rules`,
`.github/prompts`, etc.). Re-audited the whole repo, not just `instructions/`:
- `how-to-create-prompt.instructions.md` (portable, copied to every consuming
  project) still defaulted to `.claude/commands/` as *the* mirror path in two
  places. Fixed to name Claude Code as one of 20 examples and instruct: check
  which platform directory actually exists in the project being edited, don't
  assume Claude Code.
- `templates/srs.template.md` had a `prompts/doc.md` dead reference (same class
  as the `instructions/` ones fixed earlier) — templates get copied into every
  consuming project too, same basename-preserved-different-prefix rule applies.
  Fixed to `/doc`, re-synced to `.plaesy/templates/srs.template.md`.
- `docs/scripts/plaesy-init.md` documented **two things that don't match actual
  script behavior**, verified against `scripts/bash/plaesy-init.sh`: (1) an
  example install log showing chatmodes copied to `.claude/roles` — traced the
  code, that never happens for any platform, chatmodes always go to
  `.plaesy/roles/` via a separate function `setup_platform_config()` doesn't
  touch; fixed the example log. (2) A "File Copying Strategy" section presenting
  `platform.json`'s per-platform `chatmodes` mapping as real behavior — added an
  explicit "Known discrepancy" callout instead of silently rewriting it, since
  this is describing an actual gap between config intent and implementation, not
  just a wrong path.
- Flagged but **not fixed** (needs deeper verification before editing):
  `docs/scripts/plaesy-clean.md`'s example output claims
  `✅ Removed platform claude_code chatmodes: .claude/roles` — `plaesy-clean.sh`
  DOES include `"chatmodes"` in its `mapping_types` array (unlike `plaesy-init.sh`),
  so unlike the init-side bug, this might be intentional (attempting cleanup of a
  path that happens to never exist, i.e. a harmless no-op) rather than wrong
  documentation. Didn't have enough confidence to edit without reading
  `plaesy-clean.sh`'s removal logic in full — left as a note for a future session.
- Swept the rest of the repo (root `README.md`, `CLAUDE.md`,
  `instructions/agents.instructions.md`, `instructions/plaesy.instructions.md` —
  the actual `core` files copied into every consuming project — plus all
  `templates/*.md`, `checklists/*.md`, `chatmodes/*.md`) for the same three bug
  classes (stale `prompts/x.md`, missing `.plaesy/` prefix, `.claude/`-only
  assumptions): all clean, nothing else found.
**Established rule, extended again**: never assume Claude Code / `.claude/` when
writing something meant to be portable (an `instructions/*.md` or
`templates/*.md` file, or documentation describing general behavior) — this repo
supports 20 platforms with genuinely different target directories per
`platform.json`. Name Claude Code only as one example among several, or better,
point at "whichever platform directory is actually present."

## Repo-wide fix, round 2: missing `.plaesy/` prefix + a factually wrong example (2026-09-16)
Follow-up sweep after the `prompts/*.md` fix below, checking every install-copied
category (templates, checklists, scripts, chatmodes/roles), not just prompts:
- **6 bare path refs** missing the `.plaesy/` prefix — basename correct (templates
  and checklists ARE copied with filename unchanged, unlike instructions/chatmodes/
  prompts), but directory wrong: `assess-design.md` (`templates/design.template.md`),
  `assess-management.md` (`checklists/pm.checklist.md`), `assess-product.md` (2x
  checklist refs), `assess-technical.md` (2x checklist refs). Fixed by adding
  `.plaesy/` prefix — basename didn't need to change.
- **6 more** in `excel.instructions.md`/`powerpoint.instructions.md`/
  `word.instructions.md` — `scripts/bash/validate-{xlsx,pptx,docx}.sh` and the
  PowerShell twins, same missing-prefix pattern (scripts copy to
  `.plaesy/scripts/bash/`+`.plaesy/scripts/powershell/`, filename unchanged).
- **One factually wrong example**, not just a missing prefix, in
  `plaesy-graph.instructions.md`'s explanation of the graph tool's `mirrors` edge
  type: claimed `chatmodes/*.chatmode.md` → `.claude/roles/*.chatmode.md`. Traced
  and found **both wrong**: destination is actually `.plaesy/roles/*.md` (suffix
  stripped) via `plaesy-init.sh`'s dedicated chatmode-copy function; the
  `.claude/roles` mapping the example cited comes from `platform.json`'s
  `chatmodes` entry, which `plaesy-init.sh` never actually reads (`platform_mappings`
  only iterates `"prompts"` + optional `"core"` — same class of dead-config bug as
  the previously-found `"instructions"` dead path). Replaced the example with a
  verified-correct one (`prompts/assess.md` → `.claude/commands/assess.md`, a true
  filename-preserving mirror) and added a note that chatmodes/instructions are
  *not* filename-preserving, so the `mirrors` edge's plain-filename-match logic
  won't catch them — and a standing caution to verify against `plaesy-init.sh`'s
  actual copy functions, never against `platform.json` alone.
- All 12 touched files re-synced to `.plaesy/instructions/`.
**Established rule, extended**: when citing a `templates/`, `checklists/`, or
`scripts/` path from inside an `instructions/*.md` file, always prefix
`.plaesy/` — those three categories preserve the source basename on copy, so only
the directory needs correcting (unlike instructions/chatmodes/prompts, which also
rename). And never trust `platform.json` alone as evidence a copy path is real —
confirm the corresponding `plaesy-init.sh` function actually reads that mapping.

## Repo-wide fix: `prompts/*.md` path references in `instructions/*.md` (2026-09-16)
User asked to check other instructions files for the same class of bug caught in
`how-to-create-prompt.instructions.md` (a path only valid in the source repo, dead
once copied into a consuming project). Found it was pre-existing and repo-wide,
not something introduced this session:
- **6 of 8** `assess-*.instructions.md` files had a stale `**Referenced from**:
  `.plaesy/memory/assess.md`` header — wrong on two counts: prompts never installed
  to `.plaesy/memory/` (only `.plaesy/instructions/`), and even that convention is
  for instructions files, not prompts (which install to `.claude/commands/` or the
  platform equivalent, never literally as `.plaesy/` anything).
- `assess-marketing.instructions.md`, `dimension-mapping.instructions.md`,
  `quality-gates.instructions.md`, `universal-orchestrator.instructions.md` had
  `prompts/{name}.md` path references (some in "Referenced from"/"See Also"
  metadata, some inline in active instructions like "the Mode 1 anti-hallucination
  protocol in `prompts/assess.md`").
- Fix applied everywhere: replaced every `prompts/{name}.md` / stale
  `.plaesy/memory/assess.md` reference with the command name itself (`/assess`,
  `/implement`, etc.) — always resolves correctly regardless of repo vs. consuming
  project vs. platform, since it's just naming the slash command, not a file path.
  `.plaesy/instructions/dimension-mapping.md`-style references were left as-is —
  those ARE correct, that file genuinely does get copied to that exact path.
- All 12 touched files re-synced to `.plaesy/instructions/`.
**Established rule going forward**: never reference another *prompt* by its
source-repo file path (`prompts/x.md`) from inside an `instructions/*.md` file —
name the command (`/x`) instead. Referencing another *instructions* file by its
`.plaesy/instructions/x.md` copied-name is fine, since that path is real in every
consuming project.

## This repo's own maintenance note (NOT for `instructions/` — repo-specific)
When adding/renaming a command in **this** repo (the Plaesy Spec-Kit source, not a
consuming project): also update the command's row in this repo's own `README.md`
and its section in `docs/prompts/README.md`. Neither path exists in a consuming
project's `.plaesy/` (that `docs/` is an empty per-project scaffold, unrelated to
this repo's own docs) — this note was originally written into
`how-to-create-prompt.instructions.md` itself, which is wrong: that file gets
copied into every consuming project via `mapping.json`, so it must stay fully
portable (same pattern as `how-to-create-agentsmd.instructions.md`, which never
references this repo's specific paths either). Moved here instead, 2026-09-16,
after the user caught the mixing of portable-skill content with repo-specific
facts in the same instructions file.

## Correction: `how-to-create-prompt` relocated to `.plaesy/memory/` (2026-09-16)
The entry below created this as `instructions/how-to-create-prompt.instructions.md`
(portable, installed into every consuming project). User judged it as this repo's
own maintenance convention rather than a generically reusable skill worth shipping
everywhere — moved to `.plaesy/memory/how-to-create-prompt.md`, removed from
`instructions/mapping.json`'s `prompt-authoring` entry and from
`.plaesy/instructions/`. This directly applies the rule already on record in
`.plaesy/memory.md` line 36 ("Instructions install to `.plaesy/instructions/` —
`.plaesy/memory/` is for constitution.md, project memory.md, and `/doc` outputs
only") — a rule this session initially failed to apply to its own new file.
See [[how-to-create-prompt]] for the current, relocated content.

## Framework gap closed (superseded, see correction above): `how-to-create-prompt.instructions.md` (2026-09-16)
The framework gap `/improve` surfaced (no dedicated instructions file for writing
a `prompts/*.md` slash command, unlike the adjacent `how-to-create-agentsmd.md`/
`how-to-create-markdown-document.md`) — created. Encodes this session's actual
findings as durable rules: the "say routing logic once, not 3-5 times" anti-
pattern (with cited 2026 evidence: constraint density degrades instruction-
following, repeated MANDATORY/CRITICAL is noise not enforcement, short files
outperform long ones), the concrete anti-patterns found and fixed today (fake
"Success Criteria" duplicate sections, dead `.plaesy/` path references, hardcoded
project/constitution requirements where Artifact Mode should apply instead), and
the required shape/footer/mirror-sync checklist for any new or revised prompt.
Registered in `instructions/mapping.json` under a new `prompt-authoring`
keyword-triggered entry (same pattern as `agentsmd`/`markdown`, not `always_load`
— it's opt-in, triggered by "write a prompt"/"new /command" keywords). Synced to
`.plaesy/instructions/how-to-create-prompt.md`.

## `/improve` self-run: all prompt files as artifacts (2026-09-16)
User ran `/improve` with "improve semua prompt spec-kit saya" — Artifact Mode
target was the prompt files themselves (no `word`/`powerpoint`/`excel`/Figma
match — flagged as a framework gap: **no dedicated instructions file exists for
"how to write a Plaesy slash-command prompt"**, unlike `how-to-create-agentsmd.md`/
`how-to-create-markdown-document.md` which cover adjacent formats). Evaluated
against the 2026 prompt-engineering research already gathered this session
(constraint-density degrades adherence; repeated routing diagrams are noise, not
signal). Found the same "route → detail → MANDATORY restatement → ASCII diagram"
quadruple-redundancy pattern — already fixed once in `assess.md`/`doc.md` — had
also shipped in `optimize.md`, `implement.md`'s Autonomous Routing section, and
`fix.md`'s Autonomous Routing section (missed in the earlier pass). Collapsed each
to one short paragraph + the existing detail table. Also found and removed:
`fix.md`'s "Success Criteria" section (pure restatement of its own "Self-Audit"
checklist — the second such duplicate found in this file, after "Critical Rules"
removed earlier), and `continue.md`'s "State Categories & Workflow Mapping"
(9-item prose list restating the same routing as its own "Phase Detection +
Auto-Routing Logic" pseudo-code). Net: `optimize.md` 314→281, `implement.md`
374→346, `fix.md` 355→323 (then 362→355 from the Success Criteria removal, prior
turn's dead-ref fix already counted), `continue.md` 225→204 lines. `loop.md`
(385, now the largest) and `start.md`/`save.md` were re-checked and found already
free of this pattern — not touched. All `.claude/commands/` mirrors resynced.

## Cross-command check: same "requires a project" gap elsewhere? (2026-09-16)
User asked to check the rest of the command surface for the same implicit
assumption `/improve` had (needing a constitution/spec/tasks project to be useful).
Audited all 9 other prompts:
- `/assess`, `/fix`, `/doc` — already fine standalone, no constitution dependency
  in their Objective/Usage sections; `/doc` already had a non-software carve-out.
- `/optimize`, `/implement` — same gap class as `/improve` had: Objective is
  explicitly scoped to "dimension(s) active in the constitution", and
  `/optimize`'s "Pre-Optimization Validation" checklist (code quality gates ≥80,
  tests, production traffic) is 100% software-specific with no artifact-mode
  fallback. Fixed with a one-line pointer in each (not a duplicated Artifact Mode
  — that would violate the just-completed DRY trim pass): "no constitution / one
  standalone artifact instead → use `/improve` (Artifact Mode)". Kept the logic
  itself in one place (`/improve`'s Artifact Mode) rather than copying the
  artifact-type table into `/optimize.md`/`implement.md`.
- `/loop`, `/continue`, `/save`, `/start` — correctly project/session-scoped by
  design (their whole job is managing project state); no fix needed, not a gap.

## `/improve` — added Artifact Mode (2026-09-16, fourth revision same day)
User pointed out the previous design implicitly required a project with a
constitution/spec/tasks system — but Plaesy Spec-Kit deliverables are often a
single non-code artifact (a MOM doc, a Figma file, a pitch deck) with no project
scaffolding at all. Added explicit **Mode Detection**: Artifact Mode triggers when
the request names a specific file/link, or no constitution exists — skips the
`/assess`-baseline gate (no project score applies to one document) and instead
matches the artifact's type to `word`/`powerpoint`/`excel`.instructions.md or Figma
MCP + design roles, checking it against current convention for that document type
(cited). If no instructions/template matches the artifact type at all, that's
reported as a **framework gap** (not silently worked around). Project Mode's Spec
Quality check was also narrowed to fire only when a spec/tasks system has real
content — an empty/absent one now reports N/A instead of implicitly gating the
whole command (root cause of the confusion: this repo's own `/improve` run
correctly found "N/A" for Spec Quality but read as if the command needed it).
Concrete gap surfaced during that same run, not yet acted on: **no MOM/meeting-
minutes template exists anywhere in `templates/`** — first real case Artifact Mode
would flag as a framework gap.

## `/improve` — widened to all dimensions (2026-09-16, third revision same day)
Was scoped to technical/adoption/spec only; user asked for it to cover "semua aspek"
(every aspect) a spec-kit-consuming project has — not just code. Added the full
8-dimension scope set (`/improve:technical`, `:design`, `:business`, `:legal`,
`:marketing`, `:financial`, `:management`, `:product`, same pattern as `/assess`),
each dimension's "behind current best practice" table (Technical: stack/tooling
currency; Design: design-system/a11y-pattern currency; Business: pricing/model
currency; Legal: compliance-practice currency vs. letter-of-the-law; Marketing: GTM
practice currency; Financial: FinOps practice currency; Management: process/org
practice currency; Product: roadmap-framework currency), each requiring cited
evidence (Context7 for technical, WebSearch/WebFetch otherwise — same discipline as
`/assess` Mode 1). The gate logic (route defects to `/fix`, known gaps to
`/optimize`, unmeasured state to `/assess`) is unchanged; only Step 2's scope grew
from "tech-stack + spec" to "every active dimension + spec." Framework Adoption
Audit and Spec Quality Review stay as the two dimension-agnostic checks that run
every time regardless of scope.

## `/improve` — original gate design (2026-09-16, second revision)
Started as a prospective advisor; redesigned per user request into a **triage
gate**: it does not score (`/assess`'s job), does not act on known perf/quality
findings (`/optimize`'s job), and does not debug defects (`/fix`'s job) — it routes
to whichever of those three owns the finding, and directly handles only the
remainder: framework adoption gaps (unused instructions/chatmodes/templates),
version drift, tech-stack/design modernization, and spec traceability/ambiguity.
Gap identified via web research (2026-09-16) comparing against GitHub Spec Kit's
`/clarify`/`/analyze` and the 2026 "living spec"/traceability trend — see
`prompts/improve.md`. No dedicated `instructions/improve.instructions.md` was
created (same pattern as `/optimize`, which has no dedicated instructions file
either). Mirrored to `.claude/commands/improve.md` and registered in `README.md`'s
command table and `docs/prompts/README.md`. Not added to `instructions/mapping.json`
— that file maps tech-stack keywords to instruction files for
`detect-stack.sh`/`.ps1`, has no prompt/command registry, and `/optimize` itself has
no entry there either.

## Prompt-file trim pass (2026-09-16)
Web research on 2026 prompt-engineering best practices (Anthropic guidance +
empirical findings on constraint density degrading instruction-following) found the
biggest prompt files were bloated mostly by **pure duplication**, not necessary
detail. Fixed:
- `save.md`/`continue.md`: dead reference to `.plaesy/memory/shared-protocols.md`
  (never existed) → corrected to `.plaesy/instructions/tasks.md`.
- `fix.md`: removed a "Critical Rules" section that fully restated the
  Anti-Patterns/Self-Audit checklists already above it.
- `assess.md` (476→299 lines): the "Autonomous Findings Routing" section repeated
  the same routing logic 5 ways (table, ASCII flowchart, example output, pseudo-code,
  batching example) despite `dimension-mapping.instructions.md` already being the
  stated canonical source — collapsed to one priority-order paragraph pointing there.
- `doc.md` (403→329): collapsed 5 near-identical "Gap-Based Routing Examples" plus a
  duplicate handoff checklist into one short routing paragraph.
- `implement.md` (395→370): removed a verbatim-duplicate "Need Help Deciding? / call
  @nara" block (stated twice) and collapsed per-dimension routing examples to a
  pointer at `dimension-mapping.instructions.md` (same canonical-source fix as
  `assess.md`).
All three trims removed only content that was a repeat of something stated
elsewhere in the same file or already canonical in another file — no unique
instruction, example, or rule was deleted. `.claude/commands/` mirrors kept in sync.

## `dimension-mapping.instructions.md` trim (2026-09-16, follow-up)
889 → 64 lines. The 8 per-dimension sections (Assessment/Implementation/
Optimization/Error-Recovery prose + bullet lists, ~90-100 lines each) restated
content that already lives in each `assess-{dimension}.instructions.md` (what
gets measured) and each prompt's own tables (`/optimize`'s Optimization Targets,
etc.) — this file's actual job is only dimension→command routing. Collapsed to one
8-row routing table. Also cut: the 4-example "Usage Examples" walkthrough section
(~90 lines, illustrated logic the table already states) and the "Reference
Commands (Summary)" cheat-sheet (duplicated README.md's command table and every
prompt's own Usage Format section). Fixed stale flag syntax throughout
(`--security` → `--focus security`, `--dimension` prefix → `:dimension`) to match
the `--focus` convention every prompt file now uses. Synced to
`.plaesy/instructions/dimension-mapping.md` (this repo's own dogfooded per-project
copy, generated by `plaesy init`/`analyze` — kept in sync with the source manually
here since there's no running init/upgrade in this session).

## Removed commands (folded elsewhere — do not recreate)
- **`/clarify`** — folded into `/assess` Mode 1 (Ambiguity Resolution step: ≤5 questions,
  safe defaults labeled `ASSUMED`, hard forks escalate).
- **`/design`** — folded into `/assess`'s **Design Spine** (Mode 1 production, Mode 2
  audit) across 5 dimensions (UI/UX, architecture, business model, org structure,
  process); redesign/refactor when the audit fails goes to `/optimize:design`.
- **`/research`** — folded into `/assess` Mode 1 (web/Context7-backed evidence
  requirement, citation-mandatory, anti-hallucination).
- **`/evolve`** — never existed as a real command; its job (pop backlog task, advance
  lifecycle) was already `/continue`'s.
- **`/flow`** — removed, not needed (user decision).

## `/loop` — real command, was missing its `prompts/` entrypoint
Was fully speced in `instructions/loop.instructions.md` but had no `prompts/loop.md`,
so it was unreachable. Moved to `prompts/loop.md` (canonical now); reads/writes
`.plaesy/state.json` for iteration control (`max_iterations`, `consecutive_failures`,
`checkpoint_interval` — from `templates/state.template.json`).

## Constitution (new)
`/start` Phase 0 generates `.plaesy/memory/constitution.md` from
`templates/constitution.template.md` — per-project quality defaults (coverage,
performance, security bar) that override hardcoded defaults in `implement.md`/`assess.md`.

## Verifier separation (new)
`/implement` spawns a fresh review pass (no implementation context) before reporting
complete — checks the diff against the spec, not the implementer's self-report.

## Task system — single source of truth
`.plaesy/tasks/{backlog,todo,doing,done,blocked}/*.md` (one file per task, YAML
frontmatter, folder = status) is the **only** task/backlog system — matches current
best practice (Backlog.md, Agent Kanban pattern, researched 2026-09-15). A flat
`.plaesy/memory/backlog.md` bootstrapped from a template that never existed
(`templates/backlog.template.md`) was dead code — removed from both init scripts.

## Install destination change
Instructions now install to `.plaesy/instructions/[name].md` (was `.plaesy/memory/[name].md`).
`.plaesy/memory/` is reserved for constitution.md, project memory.md, `/doc` outputs
(overview.md, architecture.md, etc.), and loop bootstrap files (backlog/state, now removed).

## Fixed: `plaesy init` was silently breaking for every user
Two bash anti-patterns, both under `set -euo pipefail`:
1. `((counter++))` where counter starts at 0 evaluates falsy → `set -e` kills the
   script immediately. Existed in `plaesy-init.sh`, `common.sh`, `config-manager.sh`,
   `plaesy-clean.sh`, `plaesy-validate-memory.sh`. Fixed everywhere to `x=$((x+1))`.
2. `validate_source_directory()` checked `-maxdepth 1 -type f`, which is always 0 for
   `scripts/` (only has subdirectories) — `copy_scripts` always skipped, for everyone,
   always. Fixed to recursive `find -type f`.
Both together meant `.claude/commands/` and `CLAUDE.md` were never created for any
`claude_code` init. Verified fixed end-to-end (bash + PowerShell, fresh scratch dirs).

## Templates & checklists — were never installed
`platform.json` declared `templates`/`checklists` as `core_directories` with mapping
sources, but no copy function ever existed for them (only `core`, `instructions`,
`prompts` mapping types were handled). Added `copy_templates()`/`copy_checklists()`
(bash) and `Copy-Templates`/`Copy-Checklists` (PowerShell) — real feature completion,
not just a doc fix, since `update-agent-context.sh` actually reads
`.plaesy/templates/agent-file-template.md` at runtime and would always error otherwise.
