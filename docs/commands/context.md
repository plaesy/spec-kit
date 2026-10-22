# `context`

```bash

plaesy context update [claude|gemini|copilot|cursor|qwen|opencode]
```

## `context update`

Sync **existing** agent context files with the current feature's `plan.md`. With no
argument, updates every detected context file; if none exist, exits non-zero listing
the paths it looked for. It does not create a context file — the agent-file template
was removed on 2026-09-25. Delegates to `internal/agentcontext`
(`scripts/internal/agentcontext/update.go`).

It was `plaesy update-agent-context`: a three-word bare verb with no resource in
it. The resource is `context` and the action is `update`, matching
`features create` and `images create`. With no subcommand, `plaesy context`
prints this list. The old spelling was removed, not aliased.

**Example**: `plaesy context update claude`

See [CLI Reference](../reference.md).
