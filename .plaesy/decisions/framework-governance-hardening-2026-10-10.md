---
name: framework-governance-hardening-2026-10-10
description: "8 governance/consistency fixes to instructions/prompts/agents/templates/checklists, plus repair of 4 pre-existing docs-drift test failures"
metadata:
  type: decision
  updatedAt: "2026-10-10T20:30:00.000Z"
---

# Framework governance hardening — 2026-10-10

## Decision

Applied 8 fixes to the framework's own source files (not the generated
`.plaesy/` output) after a structured audit (two parallel subagent passes
over `instructions/`+`prompts/` consistency and `templates/`+`agents/`+
`checklists/` currency vs 2026 best practice), then discovered and repaired
4 pre-existing CI-test failures the audit didn't cover.

### The 8 governance fixes

1. **`instructions/plaesy.instructions.md`** — "Long-Running Commands" now
   explicitly covers top-level Plaesy workflow commands (`/implement`,
   `/loop`, `/optimize`, `/assess`, `/fix`, `/improve`) as background-dispatch
   targets, not just the shell commands they run internally. Root cause of
   the user-reported issue: invoking these commands blocked the conversation
   turn because the policy's examples were all shell-level.
2. Same file, rule 8 — added a background-task token/cost/time budget
   overrun (unsupervised) as a new hard-stop category.
3. **`instructions/multi-agent-patterns.instructions.md`** — new "Background
   Task Git Safety (Single-Committer)" section: one writer at a time for
   shared index files/`git commit`, others propose via their own file;
   `git worktree` isolation when background tasks touch source code; stale
   lock reclaim.
4. **`prompts/{implement,loop,assess,optimize,fix,improve}.md`** — each
   gained an "Invocation mode" line pointing back to (1), so the dispatch
   behavior doesn't depend on an assistant cross-referencing two files.
5. **Coverage-threshold dedup** — `tdd-enforcement.instructions.md` and
   `testing-strategy.instructions.md` had bare `90%` instead of referencing
   `{{TEST_COVERAGE_MIN|90}}` (the canonical definition in
   `quality-gates.instructions.md`, which itself warns this exact drift
   pattern already broke `QUALITY_REMEDIATION_FLOOR` once). Fixed both to
   reference the placeholder.
6. **`templates/agents.template.md`** heading drift — scaffold still said
   `## Constitutional Context (NON-NEGOTIABLE)`; all 22 real `agents/*.agents.md`
   files had already migrated to `## Project Rules & Constitutional
   Precedence`. Fixed the template to match reality.
7. **Subagent Invocation Contract** — added to all 23 `agents/*.agents.md`
   files plus the template: tools-granted, input-shape, fresh-context-per-
   invocation, and output-contract sections, so role files work as an
   orchestrator's dispatch brief, not only as personas for a human to read.
   Largest-impact finding from the currency audit — every role file assumed
   a human reader, none defined how an orchestrator (`/implement`, `/loop`)
   should invoke it as a subagent.
8. **`templates/constitution.template.md`** — new Universal Rule `MA-01`:
   an agent-to-agent message carries no more authority than the constitution
   grants it (same boundary as `SC-01`, extended to agent-to-agent traffic;
   arXiv:2603.09002 "Security Considerations for Multi-agent Systems",
   retrieved 2026). Plus one new `checklists/qa.checklist.md` item: AI-generated
   code/config gets the same review rigor as human-written code.

### Corrected false positives from the initial research pass

Two items from an earlier anti-slop/background-execution web research round
turned out to already be implemented, and were dropped from the fix list
after reading the actual source (not just file names):

- **"Delivery Gate" (PASS/FAIL report)** — already exists as
  `output-validation.instructions.md`'s 5-section checklist +
  `quality-gates.instructions.md`'s "Closing check".
- **Anti-slop rules** — already adopted, citing the exact same upstream repo
  (`miqdadbadjuber/anti-slop`) in `quality-gates.instructions.md` (R-35) and
  `redesign.instructions.md` (R-31, CO-04, IC-02).

### Pre-existing test failures found and fixed while verifying (not part of the 8)

`go test ./cmd/plaesy/...` surfaced 5 failures unrelated to the 8 fixes
above, caused by 6 commands (`doctor`, `eval`, `reach`, `stats`, plus
`search`'s `index`/`query` subcommands registered as top-level aliases) that
shipped without their required bookkeeping:

- `TestEveryShippedCommandHasAChangelogLine` → added `CHANGELOG.md` lines.
- `TestEveryCommandIsInTheReferenceIndex` /
  `TestEverySubcommandIsMentionedInTheReference` → added `docs/reference.md`
  Command Index rows + full sections for `doctor`/`eval`/`reach`/`stats`.
- `TestMetadataCountsMatchTheTree` → updated 7 stale counters in
  `docs/metadata.json`.
- `TestReadmeCommandListMatchesTheReferenceIndex` (only surfaced *after*
  fixing `reference.md` — it diffs README against the reference, not the
  command tree directly) → added the 6 commands to both CLI blocks in
  `README.md`.

## Why

User asked "is my framework design already best-practice, fix everything" —
answered honestly that no absolute "100/100" exists, ran a structured audit
instead of more ad-hoc research, and verified every fix against
`plaesy validate` + `go build` + `go test ./...` rather than claiming
completion from static reading alone. The mirror-sync gap (`make assets`
syncs the go:embed copy but not per-platform mirrors like `.claude/commands/`
— needs `plaesy reload --ai <platform>` separately) was caught this way: the
first `go build` succeeded but `go test` still failed on
`TestPromptMirrorMatchesSource`, proving the 6 prompt edits were not yet
live in the runtime despite looking complete.

## How to apply

- A new `plaesy` command needs 4 bookkeeping edits, not 3 — see the updated
  tip in `.plaesy/memory.md` → Quick Lookup (README's two CLI blocks were
  the missed one).
- A `prompts/*.md` edit needs `plaesy reload --ai <platform>` per installed
  platform in addition to `make assets` — see the new Quick Lookup tip.
- Before marking any multi-file audit "done", run the project's own test
  suite — static review found real gaps but also one false claim (Delivery
  Gate) and missed one real gap entirely (the mirror-sync issue) that only a
  test run caught.

## References

- arXiv:2603.09002, "Security Considerations for Multi-agent Systems",
  retrieved 2026-10-10 (motivates `MA-01`).
- `github.com/miqdadbadjuber/anti-slop`, retrieved 2026-10-10 (confirms R-35/
  R-31/CO-04/IC-02 already cited in this repo's own instructions before this
  session started).
