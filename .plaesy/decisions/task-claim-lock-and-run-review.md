---
title: "Task claims use a mkdir lock; delegated runs stay in doing until reviewed"
description: "How tasks run claims a task, and why a run does not mark it done"
updatedAt: "2026-10-15T00:00:00.000Z"
decided: "2026-10-15"
status: "Accepted and implemented"
---

# Decision: task claims use a mkdir lock; delegated runs stay in doing until reviewed

**Status**: Accepted and implemented. **Decided**: 2026-10-15.
**Deciders**: session agent, with user approval of `tasks run`.
**Related**: [spec-folder-timestamp-and-spec-tasks](spec-folder-timestamp-and-spec-tasks.md)

## Context

Several `plaesy tasks run` workers may share a spec. The first design claimed a task with a
rename from `todo/` to `doing/`. A test showed that on this Windows machine, concurrent renames
of one file all reported success in 200 of 200 rounds, so the rename does not exclude.

## Decision

- A claim creates `.claims/<task>.lock` with `os.Mkdir`, which gave exactly one winner in
  200 of 200 rounds. The rename into `doing/` happens while the lock is held.
- A lock older than 30 minutes belonged to a dead worker and is taken over.
- A successful run writes `<spec>/runs/<task>-<HHMMSS>.md` and leaves the task in `doing/`.
  Review, then `plaesy tasks complete`.
- A failed run moves the task to `blocked/` with the reason, including a budget refusal.

## Why

Exclusion must come from something that is actually exclusive on every target OS. A run's
output is text, and no code is verified by the run itself. Marking it done would claim
acceptance criteria nobody checked.

## Consequences

- API providers return text only and do not edit files. Only CLI providers can change code,
  and that depends on how the CLI is run.
- Claims are tested within one process and across three processes in a smoke test. Multi-machine
  claims are not covered.

## Access

- 2026-10-15: implemented in `internal/taskmanage/claim.go` and `cmd/plaesy/tasks_run.go`.
