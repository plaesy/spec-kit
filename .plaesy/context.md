---
title: "Session Context"
updatedAt: "2026-09-19T00:00:00.000Z"
phase: [implement]
status: [done]
---

## Context
- Working on plaesy-spec-kit itself (framework repo), improving the framework via
  Claude Code, using the repo's own `plaesy init` to dogfood it.
- Session history (oldest→newest, condensed, full detail in linked topic files):
  command-surface cleanup → Office automation parity → assess dimension coverage
  → prompt quality + design.md system → full repo audit → cross-dimensional
  prompt generalization → chatmodes/analyze path bugs → memory.md file-counts bug
  → analyze overview.md refactor → mapping.json extension-based detection →
  colon-scope sub-commands → stack detection fixes → `/generate:images` command +
  static-site framework gap → design.md spec alignment + memory migration (this
  session, below). See `.plaesy/memory.md` Core Reference section for the full
  linked list.
- This session (2026-09-19): (1) researched the real Google Labs DESIGN.md spec
  (`github.com/google-labs-code/design.md`) and found `.plaesy/memory/design.md`'s
  template/instructions had been built off a blog summary, not the spec — fixed
  field names (`rounded` not `radii`) and body-section order (8 fixed sections),
  added `instructions/how-to-create-designmd.instructions.md`. Detail:
  [designmd-spec-alignment-2026-09-19.md](memory/designmd-spec-alignment-2026-09-19.md).
  (2) Migrated all still-valid project memory out of Claude's cross-session
  auto-memory folder (`~/.claude/projects/.../memory/`) into this repo's
  `.plaesy/memory/`, then emptied that Claude-side folder entirely — project
  memory must live only in-repo now, not in a tool-specific location, so it
  travels with the repo across machines/tools. New rule recorded in
  `.plaesy/memory.md` Established Rules.

## Recent Decisions
- Project-specific facts/decisions/feedback belong ONLY in `.plaesy/memory/` +
  `.plaesy/context.md` — never in an AI tool's own cross-session memory. See
  `.plaesy/memory.md` Established Rules (added 2026-09-19).
- Always edit `scripts/{bash,powershell}/*.sh|.ps1` (source), never
  `.plaesy/scripts/**` (gitignored `plaesy init` output — edits there are lost).
- `context.md`/`memory.md` live at `.plaesy/` root, never inside `.plaesy/memory/`
  (topic files only). `core_directories` must not list `"chatmodes"` — only
  `"roles"`.
- plaesy-analyze reruns only replace `.plaesy/analysis/overview.md` (full
  overwrite). `context.md`/`memory.md` are manual/curated — analyze never writes
  to them.
- Prompt cross-references use command names (`/assess`), never source paths
  (`prompts/assess.md`).
- Every execution prompt reads active dimension(s) from
  `.plaesy/memory/constitution.md` instead of assuming software.
- Any new `instructions/*.instructions.md` MUST be added to `mapping.json` or it
  silently never installs — EXCEPT `agents.instructions.md`. See
  [[stack-detection-fixes-2026-09-17]] for detection-field details
  (`extensions`/`filenames`) and known PS1/bash gotchas.
- A scoped prompt that executes something shells out to a real script under
  `scripts/{bash,powershell}/` — never fabricates a result; missing precondition
  → stop, say what's missing. See [[generate-images-and-static-site-2026-09-18]].

## Doing
- Nothing in flight.

## Next
- Not yet committed — awaiting user's go-ahead to `git commit`. Pending batches:
  (1) 58 new `prompts/` wrapper files + docs update
  ([[colon-scope-commands-2026-09-17]]); (2) stack-detection fixes
  ([[stack-detection-fixes-2026-09-17]]); (3) `/generate:images` +
  `static-site.instructions.md` ([[generate-images-and-static-site-2026-09-18]]);
  (4) this session's design.md spec alignment (5 files, see
  [[designmd-spec-alignment-2026-09-19]]) + 4 new `.plaesy/memory/*.md` files.
- `/generate:images` untested end-to-end (no API key in this session).
- `.claude/commands/{fix,implement,optimize,loop,improve}/` still missing their
  colon-scope mirrors in this repo's live session — needs a `plaesy init` re-run.
- 3 items from the 2026-09-15 audit still deliberately unfixed — see
  [[full-repo-audit-2026-09-15]].

## Analysis Reference
Project stats (file counts, tech stack, components) live in
`.plaesy/analysis/{overview.md,project.json,project.structure.json}` —
`plaesy analyze` fully replaces `overview.md` each run, never writes here or to
`memory.md` (see [[analysis-overview-md-refactor-2026-09-16]]).
