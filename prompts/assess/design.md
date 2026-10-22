---
description: "Assess UI/UX, accessibility, design systems"
subagent: true
---

# `/assess:design` command instructions

Scope fixed to **design**: UI/UX, accessibility, design systems.

- **Loads**: `.plaesy/instructions/assess-design.md` — apply its **Persona
  Protocol** first, then follow the full protocol in `/assess`.
- **Delegates to**: `.plaesy/roles/designer.md`, `.plaesy/roles/accessibility.md`

## Success Contract

Before any work begins, write the task checklist as `contracts.json` in the task folder: each item has a `run` command and an `expected`. Run `plaesy tasks validate <task>` to record the results; the task is done when every checklist item is `passed`. Rule: `.plaesy/instructions/success-contract.md`. Do not report completion while a checklist item is not `passed` (constitution NN-08).
