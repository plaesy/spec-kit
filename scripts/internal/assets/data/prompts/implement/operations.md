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

Before the first action, write the contract and freeze it: rule in `.plaesy/instructions/success-contract.md`, template in
`.plaesy/templates/success-contract.template.md`, check with `plaesy validate contract <file>`. At the end, compare every item
with its result (Gate 0 and the Contract comparison table in `.plaesy/instructions/quality-gates.md`). Do not report completion
while a blocking item is unmet or unmeasured.
