# Merge branches into main, then fix committed conflict markers

**Date:** 2026-10-17
**Status:** accepted

## Context

The user asked to merge every branch into `main`. All local branches (`ao/*`,
`integration/spec-kit-2-5`) and remote branches (`origin/dev`,
`origin/feature/001-mcp-server`) were already ancestors of `main`, so no merge was needed.

`main` was 43 commits ahead of `origin/main`. The user asked for everything to be done, so
`main` was pushed, the merged branches were deleted, and worktrees were removed.

Afterward the user reported conflicts. `git grep` found `<<<<<<<` / `>>>>>>>` markers in
`docs/metadata.json` and `scripts/cmd/plaesy/repo_hygiene_test.go`. They came from the stash
pop in merge commit `7318581` and had been pushed.

## Decision

- Keep the upstream (HEAD) side of each conflict. The stash side was an older version, and in
  `repo_hygiene_test.go` it was empty.
- Drop the `taskDirs` read in `repo_hygiene_test.go`. It was unused after the spec walk replaced
  it, and it broke the build.
- Update the five drifted counts in `docs/metadata.json` to the values the drift test measures,
  without hand-editing them. The file says counts must come from the test.
- Verify with targeted tests only, as the Test Scope rule requires.

## Alternatives considered

- Keep the stash side. Rejected: it is older, and the `repo_hygiene_test.go` side deletes the
  committed tracking entries from `b150a00`.
- Revert merge `7318581`. Rejected: it would also undo the stash contents that were restored
  correctly.

## Consequences

- `main` is at `74e9128`. The Kilo worktree `sly-vanadium` appeared after cleanup.
- `high_delegate-provider-interface` is still in `specs/legacy/tasks/todo/`, although the context
  file calls it complete. Not verified in this session.
- Before any merge or stash pop, check the result with `git grep -nI '^<<<<<<< '`.

## References

- Repository: https://github.com/plaesy/spec-kit (accessed 2026-10-17)
- Commits: `7318581` (merge with leftover markers), `74e9128` (fix)
