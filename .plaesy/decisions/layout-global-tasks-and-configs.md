---
title: "Layout: global tasks, branch-named specs, configs at the repository root"
description: "Tasks live in .plaesy/tasks, a spec is named after its branch, and configs live in the repository root and are read from the Plaesy home, not copied into projects"
updatedAt: "2026-10-22T00:00:00.000Z"
decided: "2026-10-22"
status: "Accepted and implemented"
---

# Decision: layout of tasks, specs and configs

**Status**: Accepted and implemented. Updated 2026-10-22: init and reload no longer copy configs into a project. **Decided**: 2026-10-22.
**Deciders**: user (chose global tasks, branch names and the configs location) and session agent.
**Related**: [task-checklist-model](task-checklist-model.md), [spec-folder-timestamp-and-spec-tasks](spec-folder-timestamp-and-spec-tasks.md) (superseded)

## Context

Tasks were stored per spec, under `.plaesy/specs/<spec>/tasks/`, with timestamped spec folder names.
Config was stored under `scripts/configs/` in the repository, which is also the Go module folder, and
copied to `.plaesy/scripts/configs/`, a layout left over from the bash era. `init` and `reload` also
created `.plaesy/specs/`, which a project does not need until a feature is made.

## Decision

- **Tasks** are global: `.plaesy/tasks/<status>/<task>/`, each with `spec.md` and `contracts.json`. The
  spec a task belongs to is recorded as `spec:` in its `spec.md`. A task name is unique in the project.
- **Specs** are named after their git branch, `.plaesy/specs/<branch>/`, with no timestamp. A branch name
  is unique, so two branches cannot share a spec.
- **Configs** move to `configs/` at the repository root, next to `templates/` and `instructions/`. They are read from the Plaesy home and are
  not copied into a project. `init` and `reload` no longer write `.plaesy/configs/`. The bash and powershell
  copy steps are removed, because the repository has no sources for them.
- **`init` and `reload` never create `.plaesy/specs/`.** Default project directories are empty, and
  `reload` no longer lists `scripts` among its generated folders. A regression test covers this.

## Options considered

- Keep tasks under each spec: rejected by the user, who wanted one list of tasks across the project.
- Keep timestamped spec names: rejected, because a branch name already identifies the spec and the
  timestamp only added ordering.
- Keep `scripts/configs/`: rejected, because `scripts/` is the Go module and the name suggested bash.

## References

- Go project layout: one `cmd/<binary>/`, the rest in `internal/`, and no folder for the sake of a
  pattern. [appliedgo.net](https://learn.appliedgo.net/blog/go-project-layout), accessed 2026-10-22.
