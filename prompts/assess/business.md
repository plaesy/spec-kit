---
description: "Assess market fit, business model, revenue"
subagent: true
---

# `/assess:business` command instructions

Scope fixed to **business**: market fit, business model, revenue.

- **Loads**: `.plaesy/instructions/assess-business.md` — apply its **Persona
  Protocol** first, then follow the full protocol in `/assess`.
- **Delegates to**: `.plaesy/roles/ba.md`, `.plaesy/roles/pm.md`

## Success Contract

Before any work begins, write the task checklist as `contracts.json` in the task folder: each item has a `run` command and an `expected`. Run `plaesy tasks validate <task>` to record the results; the task is done when every checklist item is `passed`. Rule: `.plaesy/instructions/success-contract.md`. Do not report completion while a checklist item is not `passed` (constitution NN-08).
