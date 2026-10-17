---
description: "Assess deployment, release, observability, runbooks, IaC"
subagent: true
---

# `/assess:operations` command instructions

Scope fixed to **operations**: Deployment, release, reliability/SLOs, observability, infrastructure as code, runbooks.

- **Loads**: `.plaesy/instructions/assess-operations.md` — apply its **Persona
  Protocol** first, then follow the full protocol in `/assess`.
- **Delegates to**: `.plaesy/roles/devops.md`, `.plaesy/roles/sre.md`

The technical/operations boundary is in the Boundary table of the loaded file —
use it, do not re-derive it here.

## Success Contract

Before any work begins, after any routing or persona step above, write the contract, freeze it and commit it: rule in
`.plaesy/instructions/success-contract.md`, template in `.plaesy/templates/success-contract.template.md`, check with
`plaesy validate contract <file>`. At the end, compare every item with its result (Gate 0 and the Contract comparison
table in `.plaesy/instructions/quality-gates.md`). Do not report completion while a blocking item is unmet or unmeasured
(constitution NN-08).
