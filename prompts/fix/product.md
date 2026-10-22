---
description: "Feature/roadmap inconsistencies"
subagent: true
---

# `/fix:product` command instructions

Scope fixed to **product**: Unmet acceptance criteria, incorrect prioritization, or a feature that missed its stated outcome.

- **Loads**: `.plaesy/instructions/assess-product.md` — apply its **Persona
  Protocol** first, then follow the full protocol in `/fix`.

## Success Contract

Before any work begins, write the task checklist as `contracts.json` in the task folder: each item has a `run` command and an `expected`. Run `plaesy tasks validate <task>` to record the results; the task is done when every checklist item is `passed`. Rule: `.plaesy/instructions/success-contract.md`. Do not report completion while a checklist item is not `passed` (constitution NN-08).
