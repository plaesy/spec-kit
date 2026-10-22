---
title: "Spec folders are {TIMESTAMP}-{BRANCH}; tasks live inside each spec"
description: "Replaces NNN numbering and the global .plaesy/tasks/ folder"
updatedAt: "2026-10-15T00:00:00.000Z"
decided: "2026-10-15"
status: "Superseded on 2026-10-22 by layout-global-tasks-and-configs"
---

# Decision: spec folders are {TIMESTAMP}-{BRANCH}; tasks live inside each spec

**Status**: Accepted and implemented. **Decided**: 2026-10-15.
**Deciders**: user (asked for the layout) and session agent.
**Related**: [spec-worktrees-outside-repo](spec-worktrees-outside-repo.md), [task-claim-lock-and-run-review](task-claim-lock-and-run-review.md), [spec-tasks-and-worktrees](../memory/spec-tasks-and-worktrees.md)

## Context

Feature numbers came from scanning `.plaesy/specs/` for the highest `NNN`. Two
parallel branches could pick the same number. Tasks sat in one project-wide
`.plaesy/tasks/`, so a second feature had no place of its own.

## Decision

- Branch: `{BRANCH}`, the first three words of the description (`auth-login`).
- Spec folder: `.plaesy/specs/{YYYYMMDD-HHMMSS}-{BRANCH}/` in UTC. No scan is needed.
- Tasks: `.plaesy/specs/<spec>/tasks/{backlog,todo,doing,done,blocked}/`. A spec owns its tasks.
- `.plaesy/tasks/` is gone. Its 22 files moved with `git mv` to `.plaesy/specs/legacy/tasks/`,
  which has no branch. Old `NNN-slug` folders still resolve: a spec is matched to its branch
  by its name with any timestamp prefix removed.

## Why

A timestamp needs no lookup and does not collide across machines in practice. Tasks inside a
spec travel with the branch, so they cannot leak between features.

## Consequences

- Two spec folders on one branch are an error, not a silent pick.
- The spec's `tasks/` is the only task state. Keeping a second global copy would drift.

## Access

- 2026-10-15: implemented in `CreateNewFeatureWith`, `common.FindSpecName`, `taskmanage.TasksDir`.
