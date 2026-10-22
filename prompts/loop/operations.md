---
description: "Auto-fixable deploy, release, and observability findings"
subagent: true
---

# `/loop:operations` command instructions

Scope fixed to **operations**: the autonomous assess → fix → verify cycle run
against deployment, release, reliability, observability, and IaC findings.

- **Loads**: `.plaesy/instructions/assess-operations.md` — apply its **Persona
  Protocol** first, then follow the full protocol in `/loop`.

An operations finding set that stops shrinking is a plateau, not a finish line:
the measurement stopped discriminating, so the axis is wrong, not the project.
Widen the axis the way `/loop` prescribes rather than ending the run. Report
non-convergence as a finding only once the loop has actually stopped on
`consecutive_failures` — a run the user ended is not a convergence failure.

## Success Contract

Before any work begins, write the task checklist as `contracts.json` in the task folder: each item has a `run` command and an `expected`. Run `plaesy tasks validate <task>` to record the results; the task is done when every checklist item is `passed`. Rule: `.plaesy/instructions/success-contract.md`. Do not report completion while a checklist item is not `passed` (constitution NN-08).
