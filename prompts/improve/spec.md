---
description: "Only spec/plan/tasks quality + traceability"
subagent: true
---

# `/improve:spec` command instructions

This is the **spec**-scoped entry point into `/improve` (Project Mode). Follow the
full protocol defined in `/improve`, narrowed to: spec/plan/tasks quality
and traceability only — not a project quality dimension, so it is not part of the
9-dimension scope set shared with `/assess:{scope}`.

When the spec/plan/task artifact being reviewed is itself a prompt or an
AI-facing instruction, delegate that portion to `.plaesy/roles/pe.md`; see the
second bullet of `/improve` Step 2.

## Success Contract

Before any work begins, write the task checklist as `contracts.json` in the task folder: each item has a `run` command and an `expected`. Run `plaesy tasks validate <task>` to record the results; the task is done when every checklist item is `passed`. Rule: `.plaesy/instructions/success-contract.md`. Do not report completion while a checklist item is not `passed` (constitution NN-08).
