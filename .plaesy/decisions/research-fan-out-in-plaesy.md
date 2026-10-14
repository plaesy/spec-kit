---
title: "Research fan-out runs inside plaesy, not through an external skill"
description: "/assess Mode 1 spawns researcher subagents per facet and synthesizes in-run; no external deep-research dependency"
updatedAt: "2026-10-14T05:10:00.000Z"
---

# Decision: Research fan-out runs inside plaesy

## Context

Mode 1 research handled 3+ facet questions by delegating to `anthropic-skills:deep-research`. That
made plaesy's research quality depend on an external skill, and it was not installed everywhere
plaesy is. The goal was a research procedure plaesy owns.

## Decision

- Facets (three or more) fan out to one `researcher` subagent each, under
  `.plaesy/roles/researcher.md` (from `agents/researcher.agents.md`). Briefs carry the sub-question
  and the spec's constraints. Output goes to `.plaesy/memory/research/<topic>/<facet>.md`.
- Without a subagent facility, facets run one after another with the same briefs.
- The lead agent writes `summary.md` from the facet files. Synthesis stays in the run.
- Every citation is tiered (T1 to T3), verified before use, and held to the same rules whoever
  wrote it.

## Options considered

1. Keep delegating to the external skill. Rejected at the user's request.
2. Single-agent research only. Rejected: it loses the independent-angles benefit for broad
   questions.

## References

- `prompts/assess.md` (Mode 1, Step 1), accessed 2026-10-14
- `agents/researcher.agents.md`, accessed 2026-10-14
