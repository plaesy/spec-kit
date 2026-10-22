---
title: "Role mirror is always refreshed; role memory protocol removed"
updatedAt: "2026-10-20T00:00:00.000Z"
---

# Role mirror is always refreshed; role memory protocol removed

**Decision 1 — `.plaesy/roles/` is refreshed on every `init` and `reload`, with or without `--ai`.**
`scripts/internal/scaffold/reload.go` and `init.go` call `copyAgents` into `.plaesy/roles/`. `prune.go`
checks that folder for orphans as well, even when the platform keeps its agents in another folder.

**Why:** the framework's instructions delegate to `.plaesy/roles/*.md` (for example `assess-technical`).
The mirror was written only for platforms with no agents folder, so a Claude project kept a stale copy
and the delegation pointed at old text.

**Decision 2 — the "Memory Protocol" section and its `.plaesy/memory/roles/` path are removed.**
The section is removed from the 28 `agents/*.agents.md` files and from `templates/agents.template.md`, and
the "Role Memory" section is removed from `instructions/plaesy.instructions.md`.

**Why:** the memory layout in `plaesy.instructions.md` is flat, so a `roles/` subfolder was never allowed.
No role memory file existed, no code read it, and the pointer to "Memory Hierarchy → Role Memory" named a
section that was never written. The durable records are `memory.md`, `decisions/`, and `context.md`.

**Decision 3 — NN references are `NN-01..08`, and the stack examples are not tied to one stack.**
Source only; `.plaesy/` copies are regenerated with `plaesy reload`.

**References (repo paths, consulted 2026-10-20):**
- `scripts/internal/scaffold/reload.go`, `init.go`, `prune.go` (role mirror)
- `scripts/cmd/plaesy/reload.go` (help text)
- `scripts/internal/scaffold/reload_test.go` (`TestReloadRefreshesRoleMirrorWithoutAI`)
- `.plaesy/memory/constitution.md` (NN-08 row, version 2.2.0, 2026-10-17)

Related: [installers-add-binary-dir-to-path](installers-add-binary-dir-to-path.md).
