---
description: "Feature/roadmap inconsistencies"
subagent: true
---

# `/fix:product` command instructions

Scope fixed to **product**: Unmet acceptance criteria, incorrect prioritization, or a feature that missed its stated outcome.

- **Loads**: `.plaesy/instructions/assess-product.md` — apply its **Persona
  Protocol** first, then follow the full protocol in `/fix`.

## Success Contract

Before the first action, write the contract and freeze it: rule in `.plaesy/instructions/success-contract.md`, template in
`.plaesy/templates/success-contract.template.md`, check with `plaesy validate contract <file>`. At the end, compare every item
with its result (Gate 0 and the Contract comparison table in `.plaesy/instructions/quality-gates.md`). Do not report completion
while a blocking item is unmet or unmeasured.
