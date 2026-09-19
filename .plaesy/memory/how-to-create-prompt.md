---
title: "How to Write a Plaesy Slash-Command Prompt"
updatedAt: "2026-09-16T00:00:00.000Z"
---

# How to Write a Plaesy Slash-Command Prompt

**Moved here from `instructions/how-to-create-prompt.instructions.md` (2026-09-16)** —
originally created as a portable, keyword-triggered `instructions/*.instructions.md`
file (installed into every consuming project's `.plaesy/instructions/`). User judged
this content as guidance for maintaining **this repo's own** prompt authoring
convention, not a generically reusable skill worth shipping to every consuming
project — moved to `.plaesy/memory/` per this repo's own established rule (see
[[command-surface]]): *"Instructions install to `.plaesy/instructions/` — `.plaesy/memory/`
is for constitution.md, project memory.md, and `/doc` outputs only."* Removed from
`instructions/mapping.json`'s `prompt-authoring` entry and from
`.plaesy/instructions/` accordingly — this content no longer installs into any
consuming project.

Create or revise a Plaesy-style slash-command prompt file (a `/command` an agent
executes — not end-user chat guidance, not project documentation). In this repo
these live under `prompts/*.md`, mirrored by `plaesy init` to whichever directory
the active AI platform uses (`.claude/commands` for Claude Code, `.cursor/rules`
for Cursor, `.opencode/prompts` for OpenCode, `.github/prompts` for GitHub Copilot,
and 16 more — see `scripts/configs/platform.json`'s `platforms.*.mapping.prompts`).

The goal is an instruction file the agent follows reliably on the first read.
Every section should answer: "would the agent do this wrong without it?" If not,
cut it — length itself degrades instruction-following (see Evidence below).

## Required Shape

1. **Frontmatter**: `description:` one line — used elsewhere as the command's
   one-line summary (README's command table, `docs/prompts/README.md`).
2. **Title + run mode**: `# \`/name\` command instructions`, then `⚡ **Run with**:
   \`ultracode\`` (multi-agent) or state `standard` explicitly if not.
3. **Usage Format**: a fenced `bash` block of every real invocation shape
   (`/name`, `/name:{scope}`, flags). If the command shares the 8-dimension scope
   set (`technical`/`design`/`business`/`legal`/`marketing`/`financial`/
   `management`/`product`) with `/assess`, say so in one line — don't re-list the
   8 dimensions' definitions again, `/assess` already owns that.
4. **Objective**: what the command does and, critically, what it does **not** do —
   name the adjacent commands it must route to instead of reimplementing their job.
5. **Protocol**: the actual steps, as short numbered/lettered lists. One state
   machine gets **one** diagram or **one** table — never both saying the same thing.
6. **Anti-Patterns**: only non-obvious mistakes an agent would actually make.
   Don't pad this list to look thorough.
7. **Output Format**: one example of the completion report shape.
8. **Footer**: `**Follow shared protocols**: \`.plaesy/instructions/quality-gates.md\`
   → \`.plaesy/instructions/error-recovery.md\`` (add
   `→ [Global Routing](\`.plaesy/instructions/plaesy.md#autonomous-routing-rules\`)`
   only if the command routes to other commands).

## The One Rule That Matters Most: Say It Once

The single most common defect found across this repo's own prompt files (assess,
doc, implement, fix, optimize, continue — all fixed 2026-09-16) was the same
routing logic restated 3-5 times in different shapes: a table, an ASCII flowchart,
a "MANDATORY Next Step" callout, a "Phase Loop" diagram, a worked example — all
saying "after this command, run `/assess`, then route by finding type." Every
extra restatement was cut with zero loss of instruction, because none of them
added new information.

**Test before adding a section**: if you can point to another section (in this
file or another) that already states the same rule, delete the new one and link
to the original instead. `dimension-mapping.instructions.md` is the canonical
routing table — command files point to it, they don't re-derive it.

## Evidence (Why This Rule, Not Just Style Preference)

Researched 2026-09-16 against Anthropic's own prompt-engineering guidance and
2025-2026 instruction-following benchmarks:

- Constraint/rule density measurably degrades instruction-following — models don't
  flag internal contradictions, they silently pick one interpretation. More rules
  is not more reliable compliance.
  Source: arxiv.org/pdf/2605.10039 (ConInstruct/CodeIF-Bench), retrieved 2026-09-16.
- `NEVER`/`ALWAYS`/`MANDATORY`/`CRITICAL` are statistical weights to the model, not
  a logical priority system — repeating them doesn't enforce order, it adds noise.
  Source: wordman.dev/agent-instructions, retrieved 2026-09-16.
- Production agent-instruction files that work well in practice are surprisingly
  short (~120-200 lines) even for large codebases — length is a cost, not a proxy
  for thoroughness.
  Source: wordman.dev/agent-instructions, retrieved 2026-09-16.
- Anthropic's own guidance: be explicit and direct rather than relying on
  decoration; apply the "colleague test" — if a person skimming the file would be
  confused or bored, the model's adherence degrades too.
  Source: aiwithgrant.com/guides/anthropic-prompt-engineering-overview, retrieved
  2026-09-16.

## Anti-Patterns (Found and Fixed in This Repo — Don't Reintroduce)

- ❌ Same routing logic in a table AND a flowchart AND a "MANDATORY Next Step"
  callout AND a "Phase Loop" diagram — pick **one** shape, delete the rest
- ❌ A "Success Criteria"/"Critical Rules" section that just re-lists the
  Anti-Patterns or Self-Audit checklist already stated earlier in the same file
- ❌ Restating another dimension/command's full routing table inline instead of
  pointing to `dimension-mapping.instructions.md` (or that command's own file)
- ❌ Referencing a `.plaesy/` path that doesn't correspond to any real source file
  or template — verify the target exists before citing it (a dead reference here
  teaches the agent to hallucinate the missing content)
- ❌ Referencing another *prompt* by its source-repo file path (`prompts/x.md`)
  instead of its command name (`/x`) — only `instructions/`, `chatmodes/`,
  `templates/`, `checklists/` get a stable post-install mirror under `.plaesy/`;
  `prompts/` maps to a different, platform-specific folder per AI tool (20
  platforms, see `scripts/configs/platform.json`) and has no universal path
- ❌ Assuming `.claude/commands/` (or any single platform's directory) is *the*
  mirror — check which platform directory actually exists in the project
- ❌ Hardcoding "requires a project/constitution" into a command whose job could
  reasonably apply to a single standalone artifact — check whether `/improve`'s
  Mode Detection pattern (Project Mode vs Artifact Mode) applies before assuming
  a full spec-kit scaffold exists

## When Revising an Existing Prompt

Don't rewrite blindly. Read the whole file first, keep every section that states
something not said elsewhere, and cut only genuine restatement. After editing,
copy the file to every platform mirror actually present in this repo (currently
just `.claude/commands/<name>.md`, since this repo uses Claude Code) — the source
(`prompts/<name>.md`) and every present mirror must stay identical; `plaesy init`
creates the mirror once at setup, there is no ongoing build step that re-syncs it.
If the command is user-facing, also update its row in `README.md`'s command table
and its section in `docs/prompts/README.md`.
