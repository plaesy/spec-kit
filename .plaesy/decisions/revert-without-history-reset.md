---
title: "Revert uses pre-edit snapshots, never history reset"
description: "/fix, /loop and /optimize revert from snapshots under .plaesy/tmp; git reset, checkout and clean are banned for revert"
updatedAt: "2026-10-14T05:10:00.000Z"
---

# Decision: Revert uses pre-edit snapshots

## Context

The loop and fix prompts told agents to "roll back" or "revert" a failed change. Read literally,
`git reset --hard` or `git checkout -- .` also discards the user's uncommitted work. The rule also
said "restore from the snapshot taken before the change", but no step created one.

## Decision

- Before the first edit of a finding, copy every file the change will touch to
  `.plaesy/tmp/snapshots/<finding-id>/`, keeping relative paths.
- Revert copies those files back. It never runs `git reset --hard`, `git checkout -- .`, or
  `git clean`. Resetting committed history is destructive and needs the user's confirmation
  (`plaesy.md` rule 8).
- `.plaesy/tmp/` is gitignored. The snapshot folder is deleted once the finding is verified.

## Options considered

1. Keep `git reset` as the revert. Rejected: it destroys uncommitted work.
2. Commit before every fix so `git revert` works. Rejected: commits are opt-in (`--auto-commit`).

## References

- `prompts/fix.md` (Revert Rule), `prompts/loop.md` (Revert Rule), `prompts/optimize.md` (Error
  Recovery), accessed 2026-10-14
- `.gitignore` entry for `.plaesy/tmp/`, accessed 2026-10-14
