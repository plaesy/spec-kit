---
description: "Contract / policy drafting"
subagent: true
---

# `/implement:legal` command instructions

Scope fixed to **legal**: Contract / policy drafting.

- **Loads**: `.plaesy/instructions/assess-legal.md` — apply its **Persona
  Protocol** first, then follow the full protocol in `/implement`.

## Success Contract

Before the first action, write the contract and freeze it: rule in `.plaesy/instructions/success-contract.md`, template in
`.plaesy/templates/success-contract.template.md`, check with `plaesy validate contract <file>`. At the end, compare every item
with its result (Gate 0 and the Contract comparison table in `.plaesy/instructions/quality-gates.md`). Do not report completion
while a blocking item is unmet or unmeasured.
