---
description: "WCAG violations, token inconsistencies"
subagent: true
---

# `/fix:design` command instructions

Scope fixed to **design**: WCAG violations, token inconsistencies.

- **Loads**: `.plaesy/instructions/assess-design.md` — apply its **Persona
  Protocol** first, then follow the full protocol in `/fix`.

## Success Contract

Before any work begins, after any routing or persona step above, write the contract, freeze it and commit it: rule in
`.plaesy/instructions/success-contract.md`, template in `.plaesy/templates/success-contract.template.md`, check with
`plaesy validate contract <file>`. At the end, compare every item with its result (Gate 0 and the Contract comparison
table in `.plaesy/instructions/quality-gates.md`). Do not report completion while a blocking item is unmet or unmeasured
(constitution NN-08).
