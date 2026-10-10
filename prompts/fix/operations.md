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
