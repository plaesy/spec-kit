---
description: "Assess positioning, messaging, go-to-market"
---

# `/assess:marketing` command instructions

This is the **marketing**-scoped entry point into the universal assessment orchestrator.
Follow the full protocol defined in `/assess`, with scope fixed to:

- **Scope**: `marketing`
- **Loads**: `.plaesy/instructions/assess-marketing.md`
- **Delegates to**: `.plaesy/roles/market-research-analyst.md`, `.plaesy/roles/pm.md`
- **Use case**: Positioning, messaging, GTM

Run every step of `/assess` (Mode 1/2/3, Uncertainty Surfacing Protocol, Output Format,
Error Recovery, Success Criteria) exactly as written there, treating `$ARGUMENTS` (if any)
as extra scoping detail on top of the `marketing` dimension — do not broaden scope to
other dimensions unless the user explicitly asks with a comma-separated scope.
