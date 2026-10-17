---
description: "Assess code quality, architecture, tests, security, performance"
subagent: true
---

# `/assess:technical` command instructions

Scope fixed to **technical**: code quality, architecture, tests, security, performance.

- **Loads**: `.plaesy/instructions/assess-technical.md` — apply its **Persona
  Protocol** first, then follow the full protocol in `/assess`.
- **Delegates to**: `.plaesy/roles/dev.md`, `.plaesy/roles/devsecops.md`
-  **Also delegates to, when triggered**: `.plaesy/roles/qa.md` (test strategy/coverage), `.plaesy/roles/tw.md` (doc accuracy), `.plaesy/roles/data-engineer.md` (pipelines/ETL/warehouse/streaming),
  `.plaesy/roles/ai-architect.md` (AI/ML in design or prod), `.plaesy/roles/mlops.md` (models trained/served/versioned/monitored)

## Success Contract

Before the first action, write the contract and freeze it: rule in `.plaesy/instructions/success-contract.md`, template in
`.plaesy/templates/success-contract.template.md`, check with `plaesy validate contract <file>`. At the end, compare every item
with its result (Gate 0 and the Contract comparison table in `.plaesy/instructions/quality-gates.md`). Do not report completion
while a blocking item is unmet or unmeasured.
