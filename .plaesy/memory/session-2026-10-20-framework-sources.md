---
title: "Session 2026-10-20 (later): framework source fixes, role mirror, role memory removed"
updatedAt: "2026-10-20T00:00:00.000Z"
---

# Session 2026-10-20 (later): framework sources

## What changed (uncommitted, source only)

- **NN range.** Every `NN-01..07` in `agents/`, `instructions/` now reads `NN-01..08`. NN-08 (Success Contract) was added to the constitution on 2026-10-17 but the role text was not updated.
- **Stack-locked examples.** The `dev` TDD example used pytest. It now says "the project's test framework". The `researcher` example was Go-specific and now uses `{{TOOL}}`. The story-done checklist lists Go doc comments and Dart `///`. The CI security template's Node commands are `{{NAME|default}}` placeholders, with Node as the default.
- **Role mirror.** `reload` used to refresh `.plaesy/roles/` only for platforms with no agents folder. The framework's instructions delegate to `.plaesy/roles/*.md`, so Claude projects kept a stale mirror. Now `init` and every `reload` write it, with or without `--ai`. `prune` also checks it.
- **Role memory removed.** The `Memory Protocol` section in 28 role files and the template pointed to `.plaesy/memory/roles/`, a folder the flat memory layout does not allow. No role memory file ever existed and no code read it, so the section was removed.

## Learnings

- **Check the rule before the example.** The "Memory Hierarchy → Role Memory" pointer was never defined anywhere. A pointer to a missing section passes every link check, because the file still exists.
- **Count the source, not the copies.** Editing `.claude/`, `.kilo/`, or `.plaesy/` output is wrong. Edit the source, regenerate the embed with `go run ./internal/assets/gen`, then run `reload`.
- **A stale mirror can be silent.** The instructions depended on a directory that the tool never refreshed. Only a test that compares the repo's own generated output with the sources caught it (`TestReloadOnThisRepoIsIdempotent`).
- **Help text can lie.** `reload --help` said `--ai` refreshed the agent roles. It refreshed the platform's agents folder. The text now matches the code.
- **Build the binary you test.** The installed `plaesy` was older than the source. Use `go build -o <scratch>` for testing a change.

## Open

- Installed `plaesy` is still the old build. Reinstall after the commit.
- `.kilo/agents/` and `.plaesy/roles/` in other projects still show the old text until those projects run `reload`.
- Nothing committed yet. See [session-2026-10-20-installers](session-2026-10-20-installers.md) for the branch and commits already made.

Related: [session-2026-10-20-installers](session-2026-10-20-installers.md), [installers-add-binary-dir-to-path](../decisions/installers-add-binary-dir-to-path.md).
