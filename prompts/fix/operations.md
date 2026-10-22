---
description: "Deployment, release, observability, runbook defects"
subagent: true
---

# `/fix:operations` command instructions

Scope fixed to **operations**: Failed deploys, pipeline breakage, rollback and restore failures, alert and incident noise.

- **Loads**: `.plaesy/instructions/assess-operations.md` — apply its **Persona
  Protocol** first, then follow the full protocol in `/fix`.

A defect in application code rather than in the pipeline, infrastructure, or the
runbook is a `technical` finding — route it to `/fix:technical` rather than
fixing it under this scope.

## Success Contract

Before any work begins, write the task checklist as `contracts.json` in the task folder: each item has a `run` command and an `expected`. Run `plaesy tasks validate <task>` to record the results; the task is done when every checklist item is `passed`. Rule: `.plaesy/instructions/success-contract.md`. Do not report completion while a checklist item is not `passed` (constitution NN-08).
