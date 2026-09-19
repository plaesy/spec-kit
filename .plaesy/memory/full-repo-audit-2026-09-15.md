---
title: "Full Repo Audit (ultracode workflow) + Fixes"
updatedAt: "2026-09-15T00:00:00.000Z"
---

# Full Repo Audit (ultracode workflow) + Fixes

## Method
User asked "is spec-kit fully the best?" — ran a `Workflow` (ultracode) with 7
parallel finder agents (tech instructions, chatmodes, bash/PS1 script parity,
checklists, templates, docs/, repo-wide orphan/broken-reference sweep) + a
completeness-critic round proposing 4 follow-ups (task-model docs vs
plaesy-task-manage reality, plaesy-trim.sh bash/ps1 parity, Makefile safety,
test coverage gaps) + adversarial verification (3 skeptics per finding,
majority-refute kills it). 41/43 raw findings survived verification across two
rounds. 141 subagents, ~6.8M tokens, ~15 min wall clock.

## Fixed this session (of the 41 confirmed findings)

**Functional bug** (highest impact): `detect-stack.sh`/`.ps1` matched keywords
by plain substring, so short/common keywords false-positive-matched inside
longer words — `"java"` inside `"javascript"`, `"pr"` inside `"prepare"`/
`"private"`. Fixed both to word-boundary matching (bash: glob char-class
boundary test; PowerShell: `\b`-anchored regex). Verified both catch the false
positive and still match the real case.

**Missing chatmode** (functional gap, not just a doc bug): `@nara` ("Decision
Oracle") is referenced as a real, usable capability in `prompts/fix.md` and
`prompts/implement.md` with full usage examples, but `chatmodes/nara.chatmode.md`
never existed. Created it (RACE-framework format matching the other 22
chatmodes) — multi-perspective decision synthesis for patch-vs-refactor,
symptom-vs-root-cause, architecture/library choice calls.

**Broken links, systemic**: every one of `docs/{instructions,chatmodes,
checklists,templates,testing,prompts}/README.md` had `../README.md` (should be
`../../README.md` — these READMEs live two levels deep). `docs/chatmodes/README.md`
additionally had ALL 23 chatmode links as `./x.chatmode.md` (should be
`../../chatmodes/x.chatmode.md`) — not just the nara one, every single row.
`docs/checklists/README.md` had the same for its 7 checklist links + the Usage
section's script paths. `docs/templates/README.md` had it for ~40 template
links. Root `README.md` linked to `docs/README.md` which didn't exist — created
it as a proper hub linking all 8 docs subdirectories. Fixed `docs/tasks/README.md`
stale `templates/tasks/` path (templates are flat, not nested).

**Orphaned checklists wired in** (7 files, previously referenced only from
`docs/checklists/README.md`, never from any prompt/instruction driving actual
agent behavior): `story-draft`/`story-done` → `prompts/implement.md`
(pre-implementation + Definition-of-Done gates); `pm` → `assess-management.
instructions.md`; `po`/`update` → `assess-product.instructions.md`; `sa`/`qa`
→ `assess-technical.instructions.md`; `update` also → `prompts/fix.md` (scope
creep beyond a one-line patch).

**Dead code removed**: `scripts/bash/config-manager.sh` had `parse_json_array`
defined twice (keyed-extraction version, then a single-arg version) — bash
silently lets the second redefine the first, so the first 40-line
implementation could never run (only call site uses the single-arg form).
Deleted the dead first definition.

**Bash/PowerShell parity fix**: `plaesy-trim.sh`'s `full`/`ultra` level was
missing the leading-whitespace-strip transform that `plaesy-trim.ps1` has —
added `sed -E 's/^[[:space:]]+//'` to match.

**PowerShell parity fix**: `detect-stack.ps1` hardcoded the 4 category names it
scans (`frameworks`/`languages`/`cross_cutting`/`methodologies`) via explicit
calls; the bash twin is fully generic. Changed to iterate every
`mapping.mappings` property generically (skipping `always_load`), matching bash
— a new mapping.json category is now picked up on both platforms without a
script edit. Verified: both now return 19 matches for this repo.

**Contradictory task-model docs fixed**: `templates/README.md` described an
epic/story/JSON-status task model that `plaesy-task-manage.sh`/`.ps1` never
implements (those only move flat `.md` files between `.plaesy/tasks/{status}/`).
Annotated the 3 templates describing that model as "design reference, not the
implemented model" with a full explanatory note, pointing at the real flat-file
model instead.

**`template-registry.json` completed**: it claimed to be "comprehensive" but
was missing `agent-file-template`, `spec-structure-folder`, `state`,
`task-structure-guide`, `task-readme`, `task-status`, `design` (templates), plus
`iso25010-quality-model`, `research`, `change-request`, `integration-examples`,
and both `market-research-analyst` and `nara` chatmodes. Added all of them.
Confirmed via repo-wide grep that nothing actually reads this file
programmatically (pure documentation-completeness fix, zero behavior risk).

**Orphaned templates linked**: `templates/chatmode.template.md` (the scaffold
for authoring a new chatmode) was referenced nowhere at all — added to
`templates/README.md`'s Platform-specific assets section.

**Unbounded-growth guard added** (same pattern as the `docs/components.md` fix
from an earlier session): `templates/srs.template.md` (Functional Requirements,
FR-XXX) and `templates/test-plan.template.md` (Test Cases) now carry an
"at scale, split into topic files + keep an index" note, threshold ~20-30 items,
same index-links-to-topic-files pattern as `.plaesy/memory.md`.

**Orphaned scripts wired in**: `plaesy-task-manage.sh`'s header comment
referenced `/evolve` — a command removed in an earlier session (see
[[command-surface]]) — fixed the comment and pointed at `/continue`/`/implement`
as the actual owners of most task transitions. Added discoverability references
to both `plaesy-task-manage.sh` and `plaesy-validate-memory.sh` in
`instructions/tasks.instructions.md`.

**Makefile documented, not changed**: `make`/`make reload` unconditionally
`rm -rf`s `.claude/`, `.plaesy/`, and `CLAUDE.md` with no confirmation and no
mention anywhere in CONTRIBUTING.md. Documented it in CONTRIBUTING.md
("Development setup" step 4) rather than changing the destructive default
behavior — that's a design decision for the user to make, not something to
silently alter.

## Deliberately NOT fixed — flagged for user decision or larger scope
- **`scripts/powershell/platform-detector.ps1` has no bash twin** — every other
  scripts/powershell/*.ps1 file does. Fixing means authoring a new ~equivalent
  bash script from scratch, not a small edit.
- **Test coverage gaps**: `testing/powershell/run.ps1` has zero tests for
  `detect-stack.ps1` (bash suite has a dedicated block); neither suite tests
  `plaesy-task-manage`/`plaesy-validate-memory` at all (bash or PowerShell).
  Needs new test authorship, not flagged as urgent enough to do blind.
- **`task-status.template.json`'s schema is fully disconnected** from the real
  task lifecycle scripts (they never read/write its `quality_gates`/
  `progress_notes`/etc fields) — documented as a "reference, not implemented"
  in `templates/README.md`, but making `plaesy-task-manage.sh` actually use
  that schema would be a real feature addition, not a bug fix.
- 2 refuted findings (not included in the 41): discarded after adversarial
  verification found them unsupported.

## Established rule for future sessions
Every `docs/{x}/README.md` lives at depth 2 (`docs/x/`) — any link back to repo
root needs `../../`, not `../`. Any link to a sibling top-level folder
(`chatmodes/`, `templates/`, `checklists/`, `scripts/`) also needs `../../x/`,
not `./x` or `../x/`. Check this explicitly whenever adding a new docs/*/README.md
or a new cross-reference in an existing one.
