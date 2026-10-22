---
title: "Repo hygiene and test paths"
description: "What broke repeatedly while moving tasks, specs and configs, and how to avoid it"
updatedAt: "2026-10-22T00:00:00.000Z"
metadata:
  type: project
---

# Repo hygiene and test paths

- **Tracked files that were removed from the working tree fail the hygiene test.**
  `TestNoBuildArtifactIsTracked` lists them. After deleting or moving tracked files, run
  `git rm --cached -- <path>` for each missing path. This only changes the index.
  **Why:** the commit from `make reset` tracks everything, so a later deletion looks like a leftover.
  **How to apply:** after any move, run the loop `git ls-files | while read f; do [ -e "$f" ] || git rm --cached -q -- "$f"; done`.
- **Test paths depend on the package depth.** From `scripts/cmd/plaesy` or `scripts/internal/<pkg>`, the
  repository root is `../../..`. A `..\..\configs` path points at `scripts/configs`, which no longer exists.
- **Drift tests read the command pages.** `TestEverySubcommandIsMentionedInTheReference` reads
  `docs/commands/*.md` as well as `docs/reference.md`. A subcommand that appears only in its page is found.
- **Line endings.** `.gitattributes` sets `* text=auto eol=lf`. Files written with CRLF show a warning
  until they are converted to LF. Convert changed files before committing.
- **Editing code with node scripts is fragile.** Escaped backticks and `\n` inside template strings
  broke several edits silently or produced invalid Go. Use the Edit tool, or write the replacement text to
  a file and splice it with awk.
- **`plaesy reload` must not recreate removed folders.** Its list of generated directories
  (`reloadGeneratedDirs`) decides what it creates. Keep `scripts` and `specs` out of it.

Related: [layout-global-tasks-and-configs](../decisions/layout-global-tasks-and-configs.md).
