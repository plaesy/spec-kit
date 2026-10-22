---
description: "Tech stack, architecture pattern, tooling currency"
subagent: true
---

# `/improve:technical` command instructions

Scope fixed to **technical**: Tech stack, architecture pattern, tooling currency.

- **Loads**: `.plaesy/instructions/assess-technical.md` — apply its **Persona
  Protocol** first, then follow the full protocol in `/improve`.

## Success Contract

Before any work begins, write the task checklist as `contracts.json` in the task folder: each item has a `run` command and an `expected`. Run `plaesy tasks validate <task>` to record the results; the task is done when every checklist item is `passed`. Rule: `.plaesy/instructions/success-contract.md`. Do not report completion while a checklist item is not `passed` (constitution NN-08).
