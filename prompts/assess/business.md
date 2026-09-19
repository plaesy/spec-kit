---
description: "Assess market fit, business model, revenue"
---

# `/assess:business` command instructions

This is the **business**-scoped entry point into the universal assessment orchestrator.
Follow the full protocol defined in `/assess`, with scope fixed to:

- **Scope**: `business`
- **Loads**: `.plaesy/instructions/assess-business.md`
- **Delegates to**: `.plaesy/roles/ba.md`, `.plaesy/roles/pm.md`
- **Use case**: Market fit, business model, viability

Run every step of `/assess` (Mode 1/2/3, Uncertainty Surfacing Protocol, Output Format,
Error Recovery, Success Criteria) exactly as written there, treating `$ARGUMENTS` (if any)
as extra scoping detail on top of the `business` dimension — do not broaden scope to
other dimensions unless the user explicitly asks with a comma-separated scope.
