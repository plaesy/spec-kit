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

Before any work begins, after any routing or persona step above, write the contract, freeze it and commit it: rule in
`.plaesy/instructions/success-contract.md`, template in `.plaesy/templates/success-contract.template.md`, check with
`plaesy validate contract <file>`. At the end, compare every item with its result (Gate 0 and the Contract comparison
table in `.plaesy/instructions/quality-gates.md`). Do not report completion while a blocking item is unmet or unmeasured
(constitution NN-08).
