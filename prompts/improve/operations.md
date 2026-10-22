---
description: "Deployment, release, observability, IaC practice currency"
subagent: true
---

# `/improve:operations` command instructions

Scope fixed to **operations**: Deployment practice, release hygiene, observability maturity, IaC currency.

- **Loads**: `.plaesy/instructions/assess-operations.md` — apply its **Persona
  Protocol** first, then follow the full protocol in `/improve`.

"Behind best practice" must be replaced with a concrete, checkable criterion
before it counts as a finding. Name the practice, the source, and what the
project does today.

## Success Contract

Before any work begins, write the task checklist as `contracts.json` in the task folder: each item has a `run` command and an `expected`. Run `plaesy tasks validate <task>` to record the results; the task is done when every checklist item is `passed`. Rule: `.plaesy/instructions/success-contract.md`. Do not report completion while a checklist item is not `passed` (constitution NN-08).
