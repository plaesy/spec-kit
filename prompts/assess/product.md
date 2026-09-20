---
description: "Assess feature set, roadmap, competitive advantage"
---

# `/assess:product` command instructions

This is the **product**-scoped entry point into the universal assessment orchestrator.
Follow the full protocol defined in `/assess`, with scope fixed to:

- **Scope**: `product`
- **Loads**: `.plaesy/instructions/assess-product.md`
- **Delegates to**: `.plaesy/roles/pm.md`, `.plaesy/roles/ba.md`
- **Use case**: Feature set, roadmap, competitive advantage

Run every step of `/assess` (Mode 1/2/3, Uncertainty Surfacing Protocol, Output Format,
Error Recovery, Success Criteria) exactly as written there, treating `$ARGUMENTS` (if any)
as extra scoping detail on top of the `product` dimension — do not broaden scope to
other dimensions unless the user explicitly asks with a comma-separated scope.
