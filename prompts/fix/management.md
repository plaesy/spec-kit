---
description: "Broken process, role, or org-health claims"
subagent: true
---

# `/fix:management` command instructions

Scope fixed to **management**: Process steps that fail in practice, ownership gaps, or capacity assumptions that proved wrong.

- **Loads**: `.plaesy/instructions/assess-management.md` — apply its **Persona
  Protocol** first, then follow the full protocol in `/fix`.

## Success Contract

Before any work begins, write the task checklist as `contracts.json` in the task folder: each item has a `run` command and an `expected`. Run `plaesy tasks validate <task>` to record the results; the task is done when every checklist item is `passed`. Rule: `.plaesy/instructions/success-contract.md`. Do not report completion while a checklist item is not `passed` (constitution NN-08).
