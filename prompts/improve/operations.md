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

Before any work begins, after any routing or persona step above, write the contract, freeze it and commit it: rule in
`.plaesy/instructions/success-contract.md`, template in `.plaesy/templates/success-contract.template.md`, check with
`plaesy validate contract <file>`. At the end, compare every item with its result (Gate 0 and the Contract comparison
table in `.plaesy/instructions/quality-gates.md`). Do not report completion while a blocking item is unmet or unmeasured
(constitution NN-08).
