---
description: "Auto-fixable claim-sourcing and channel-fit findings"
subagent: true
---

# `/loop:marketing` command instructions

Scope fixed to **marketing**: Marketing findings, re-measured each iteration.
A plateau is a reason to widen the axis, never to stop; `/loop` owns when a run ends.

- **Loads**: `.plaesy/instructions/assess-marketing.md` — apply its **Persona
  Protocol** first, then follow the full protocol in `/loop`.

## Success Contract

Before any work begins, write the task checklist as `contracts.json` in the task folder: each item has a `run` command and an `expected`. Run `plaesy tasks validate <task>` to record the results; the task is done when every checklist item is `passed`. Rule: `.plaesy/instructions/success-contract.md`. Do not report completion while a checklist item is not `passed` (constitution NN-08).
