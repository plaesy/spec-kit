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

Before any work begins, write the task checklist as `contracts.json` in the task folder: each item has a `run` command and an `expected`. Run `plaesy tasks validate <task>` to record the results; the task is done when every checklist item is `passed`. Rule: `.plaesy/instructions/success-contract.md`. Do not report completion while a checklist item is not `passed` (constitution NN-08).
