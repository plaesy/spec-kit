---
description: "Reliability/SLO, release cadence, toil"
subagent: true
---

# `/optimize:operations` command instructions

Scope fixed to **operations**: Deploy and release throughput, build duration, SLO attainment, alert quality, IaC drift.

- **Loads**: `.plaesy/instructions/assess-operations.md` — apply its **Persona
  Protocol** first, then follow the full protocol in `/optimize`.

Optimize a measure, never a guess: establish the baseline for the target first
(per the canonical Baseline step in `/optimize`), and record the before/after
numbers. A deploy that is fast because it skips verification is a regression,
not an optimization.

## Success Contract

Before any work begins, write the task checklist as `contracts.json` in the task folder: each item has a `run` command and an `expected`. Run `plaesy tasks validate <task>` to record the results; the task is done when every checklist item is `passed`. Rule: `.plaesy/instructions/success-contract.md`. Do not report completion while a checklist item is not `passed` (constitution NN-08).
