---
title: "Task checklist model: folders, spec.md, contracts.json, run/expected"
description: "Replaces the frozen Success Contract and file-based tasks"
updatedAt: "2026-10-22T00:00:00.000Z"
decided: "2026-10-22"
status: "Accepted and implemented; tasks are global, spec is the branch name"
---

# Decision: task checklist model

**Status**: Accepted and implemented in the Go tree. **Decided**: 2026-10-22.
**Deciders**: user (chose the model) and session agent.
**Related**: [spec-folder-timestamp-and-spec-tasks](spec-folder-timestamp-and-spec-tasks.md), [task-claim-lock-and-run-review](task-claim-lock-and-run-review.md)

## Context

The first contract design froze targets in git (commit before work, anchored by history). The user
pointed out that projects may have no git, and that the real goal is only a checklist that says
whether a task is done. The freeze, amendment and waiver machinery was removed.

## Decision

- A task is a folder: `tasks/<status>/<priority>_<slug>/` holding `spec.md` and `contracts.json`.
- `contracts.json` is a checklist. Each item has `id`, `check`, `run` (shell command, run from the
  project root), `expected` (what the last output line must reach), `status` (`pending`, `passed`,
  `failed`) and `histories` (every failed run: `at`, `actual`, `run`).
- `run` exits 0 for `pass`. Other expectations judge the last output line (`zero`, `yes`, `N of N`,
  or a number with a comparison such as `>= 9`).
- A task is done when every item is `passed`. `plaesy tasks validate` moves a finished task to `done`.
- `blocked` is only for a task waiting on something outside the work. An unfinished task stays in `doing`.
- `tasks create` creates the spec first when it does not exist, then the task inside it. A failed spec
  creation removes what it made.

## Options considered

- Keep the frozen contract in git: rejected, because it needs git and a commit before work.
- Store a timestamp in the file: rejected as proof, since a file can be edited with its timestamp.
- Name the run field `run` (GitHub Actions uses `run:` for shell commands). Source:
  [GitHub Actions cheatsheet](https://www.glukhov.org/developer-tools/ci-cd/github-actions-cheatsheet/),
  accessed 2026-10-22.

## Decided on 2026-10-22

- Tasks are global: `.plaesy/tasks/<status>/<task>/`. A task records its spec in its `spec.md`.
- A spec folder is named after its git branch, `.plaesy/specs/<branch>/`, with no timestamp. Two
  branches cannot share a spec, because a branch name is unique.

## References

- Commit stomping: git dates are set by the committer, so a commit date is not proof on its own.
  [cybersecuritynews](https://cybersecuritynews.com/commit-stomping/amp/), accessed 2026-10-22.
