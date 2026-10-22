---
description: "Only unused Plaesy Spec-Kit capabilities + version drift"
subagent: true
---

# `/improve:adoption` command instructions

This is the **adoption**-scoped entry point into `/improve` (Project Mode). Follow
the full protocol defined in `/improve`, narrowed to: unused Plaesy
Spec-Kit capabilities and version drift only — not a project quality dimension, so
it is not part of the 9-dimension scope set shared with `/assess:{scope}`.

## Success Contract

Before any work begins, write the task checklist as `contracts.json` in the task folder: each item has a `run` command and an `expected`. Run `plaesy tasks validate <task>` to record the results; the task is done when every checklist item is `passed`. Rule: `.plaesy/instructions/success-contract.md`. Do not report completion while a checklist item is not `passed` (constitution NN-08).
