---
description: "Assess cost structure, pricing, profitability"
---

# `/assess:financial` command instructions

This is the **financial**-scoped entry point into the universal assessment orchestrator.
Follow the full protocol defined in `/assess`, with scope fixed to:

- **Scope**: `financial`
- **Loads**: `.plaesy/instructions/assess-financial.md`
- **Delegates to**: `.plaesy/roles/ba.md`, `.plaesy/roles/pm.md`
- **Use case**: Pricing, cost structure, profitability, funding

Run every step of `/assess` (Mode 1/2/3, Uncertainty Surfacing Protocol, Output Format,
Error Recovery, Success Criteria) exactly as written there, treating `$ARGUMENTS` (if any)
as extra scoping detail on top of the `financial` dimension — do not broaden scope to
other dimensions unless the user explicitly asks with a comma-separated scope.
