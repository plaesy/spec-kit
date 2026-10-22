---
description: "Deployment, release, runbook, IaC implementation"
subagent: true
---

# `/implement:operations` command instructions

Scope fixed to **operations**: Deploy pipeline, release, IaC, observability, runbook, incident readiness.

- **Loads**: `.plaesy/instructions/assess-operations.md` — apply its **Persona
  Protocol** first, then follow the full protocol in `/implement`.

Build the environment reproducibly: an environment that cannot be recreated from
the repository is not implemented, it is a snapshot. Follow the TDD cycle for any
script or check you add, and gate on the constitution's coverage minimum as
`/implement` requires.

## Success Contract

Before any work begins, write the task checklist as `contracts.json` in the task folder: each item has a `run` command and an `expected`. Run `plaesy tasks validate <task>` to record the results; the task is done when every checklist item is `passed`. Rule: `.plaesy/instructions/success-contract.md`. Do not report completion while a checklist item is not `passed` (constitution NN-08).
