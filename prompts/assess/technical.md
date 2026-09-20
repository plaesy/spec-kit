---
description: "Assess code quality, architecture, tests, security, infrastructure, deployment"
---

# `/assess:technical` command instructions

This is the **technical**-scoped entry point into the universal assessment orchestrator.
Follow the full protocol defined in `/assess`, with scope fixed to:

- **Scope**: `technical`
- **Loads**: `.plaesy/instructions/assess-technical.md`
- **Delegates to**: `.plaesy/roles/dev.md`, `.plaesy/roles/devsecops.md`
- **Use case**: Code quality, tests, security, performance, infrastructure, deployment

Run every step of `/assess` (Mode 1/2/3, Uncertainty Surfacing Protocol, Output Format,
Error Recovery, Success Criteria) exactly as written there, treating `$ARGUMENTS` (if any)
as extra scoping detail on top of the `technical` dimension — do not broaden scope to
other dimensions unless the user explicitly asks with a comma-separated scope.
