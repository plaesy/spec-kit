---
title: "Feature worktrees live outside the repository; cleanup never forces"
description: "Where --worktree features are kept, and when plaesy features clean removes one"
updatedAt: "2026-10-15T00:00:00.000Z"
decided: "2026-10-15"
status: "Accepted and implemented"
---

# Decision: feature worktrees live outside the repository; cleanup never forces

**Status**: Accepted and implemented. **Decided**: 2026-10-15.
**Deciders**: user (asked for parallel features) and session agent.
**Related**: [spec-folder-timestamp-and-spec-tasks](spec-folder-timestamp-and-spec-tasks.md)

## Context

Several features in progress need several checkouts. A worktree inside the repository would
nest a full checkout in a tracked tree, and `.plaesy/specs/` is tracked.

## Decision

- Location: `PLAESY_WORKTREE_HOME` if set; Windows `%LOCALAPPDATA%\Plaesy\worktrees\`;
  elsewhere `$XDG_DATA_HOME/plaesy/worktrees/` or `~/.local/share/plaesy/worktrees/`.
- Per project: `{project}-{hash8}`, the hash of the main checkout's path.
- `features create` is serial by default. `--worktree` opts in.
- `features clean` is a dry run by default. `--apply` removes only worktrees whose branch is
  merged into main or master, that hold no git-ignored files, and that git accepts without
  `--force`. It never deletes branches.

## Why

`git worktree remove` deletes ignored files silently. A spec is ignored on any machine whose
global gitignore covers `.plaesy/`, so the check is needed for data safety, not tidiness.
Squash-merged branches are not ancestors of main and are kept on purpose.

## Access

- 2026-10-15: implemented in `worktree.go`, `clean.go`. Real-git tests cover create, merge,
  dirty refusal, and ignored-spec refusal.
