---
description: "Auto-fixable code quality, security, and test findings"
subagent: true
---

# `/loop:technical` command instructions

Scope fixed to **technical**: Findings the technical assessment surfaces, re-measured each iteration.
A plateau is a reason to widen the axis, never to stop; `/loop` owns when a run ends.

- **Loads**: `.plaesy/instructions/assess-technical.md` — apply its **Persona
  Protocol** first, then follow the full protocol in `/loop`.

## Success Contract

Before the first action, write the contract and freeze it: rule in `.plaesy/instructions/success-contract.md`, template in
`.plaesy/templates/success-contract.template.md`, check with `plaesy validate contract <file>`. At the end, compare every item
with its result (Gate 0 and the Contract comparison table in `.plaesy/instructions/quality-gates.md`). Do not report completion
while a blocking item is unmet or unmeasured.
