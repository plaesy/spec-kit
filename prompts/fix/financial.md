---
description: "Broken cost, pricing, or profitability figures"
subagent: true
---

# `/fix:financial` command instructions

Scope fixed to **financial**: Incorrect figures, broken pricing logic, or funding projections that no longer hold.

- **Loads**: `.plaesy/instructions/assess-financial.md` — apply its **Persona
  Protocol** first, then follow the full protocol in `/fix`.

## Success Contract

Before any work begins, write the task checklist as `contracts.json` in the task folder: each item has a `run` command and an `expected`. Run `plaesy tasks validate <task>` to record the results; the task is done when every checklist item is `passed`. Rule: `.plaesy/instructions/success-contract.md`. Do not report completion while a checklist item is not `passed` (constitution NN-08).
