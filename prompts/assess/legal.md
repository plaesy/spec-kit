---
description: "Assess regulatory compliance, data privacy, risks"
---

# `/assess:legal` command instructions

This is the **legal**-scoped entry point into the universal assessment orchestrator.
Follow the full protocol defined in `/assess`, with scope fixed to:

- **Scope**: `legal`
- **Loads**: `.plaesy/instructions/assess-legal.md`
- **Delegates to**: `.plaesy/roles/privacy-legal.md`, `.plaesy/roles/compliance.md`
- **Use case**: Regulatory, compliance, data privacy, risk

Run every step of `/assess` (Mode 1/2/3, Uncertainty Surfacing Protocol, Output Format,
Error Recovery, Success Criteria) exactly as written there, treating `$ARGUMENTS` (if any)
as extra scoping detail on top of the `legal` dimension — do not broaden scope to
other dimensions unless the user explicitly asks with a comma-separated scope.
