---
description: "Assess team capacity, project management, processes"
---

# `/assess:management` command instructions

This is the **management**-scoped entry point into the universal assessment orchestrator.
Follow the full protocol defined in `/assess`, with scope fixed to:

- **Scope**: `management`
- **Loads**: `.plaesy/instructions/assess-management.md`
- **Delegates to**: `.plaesy/roles/pm.md`, `.plaesy/roles/sm.md`
- **Use case**: Team capacity, project management, processes, organizational health

Run every step of `/assess` (Mode 1/2/3, Uncertainty Surfacing Protocol, Output Format,
Error Recovery, Success Criteria) exactly as written there, treating `$ARGUMENTS` (if any)
as extra scoping detail on top of the `management` dimension — do not broaden scope to
other dimensions unless the user explicitly asks with a comma-separated scope.
