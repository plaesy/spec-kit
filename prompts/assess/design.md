---
description: "Assess UI/UX, accessibility, design systems"
---

# `/assess:design` command instructions

This is the **design**-scoped entry point into the universal assessment orchestrator.
Follow the full protocol defined in `/assess`, with scope fixed to:

- **Scope**: `design`
- **Loads**: `.plaesy/instructions/assess-design.md`
- **Delegates to**: `.plaesy/roles/designer.md`, `.plaesy/roles/accessibility.md`
- **Use case**: UI/UX, WCAG compliance, design systems

Run every step of `/assess` (Mode 1/2/3, Uncertainty Surfacing Protocol, Output Format,
Error Recovery, Success Criteria) exactly as written there, treating `$ARGUMENTS` (if any)
as extra scoping detail on top of the `design` dimension — do not broaden scope to other
dimensions unless the user explicitly asks with a comma-separated scope.
