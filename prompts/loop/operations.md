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

Before any work begins, after any routing or persona step above, write the contract, freeze it and commit it: rule in
`.plaesy/instructions/success-contract.md`, template in `.plaesy/templates/success-contract.template.md`, check with
`plaesy validate contract <file>`. At the end, compare every item with its result (Gate 0 and the Contract comparison
table in `.plaesy/instructions/quality-gates.md`). Do not report completion while a blocking item is unmet or unmeasured
(constitution NN-08).
