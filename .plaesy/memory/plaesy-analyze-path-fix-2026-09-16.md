---
title: "plaesy-analyze path fix + chatmodes core_directories bug"
updatedAt: "2026-09-16T00:00:00.000Z"
---

# plaesy-analyze context.md/memory.md path fix + chatmodes core_directories bug (2026-09-16)

## Bug 1: `.plaesy/chatmodes` empty folder created on every init

`scripts/configs/platform.json` (+ `.plaesy/scripts/configs/platform.json` copy) listed
`"chatmodes"` in `core_directories` alongside `"roles"`. Init loops `core_directories` and
`mkdir`s every entry, but the actual chatmode-file copy step (`plaesy-init.sh` ~L685-712)
only ever writes to `.plaesy/roles/` — so `.plaesy/chatmodes/` was created and left
permanently empty on every fresh install.

Note: `mapping.chatmodes` (a *different* key, the per-platform mapping object at
`platform.json` L49+, e.g. `.claude/roles`, `.github/chatmodes`) is unrelated and still
needed — do not confuse the two when touching this file again.

**Fix**: removed `"chatmodes"` from `core_directories` in both platform.json copies.

## Bug 2: plaesy-analyze wrote context.md/memory.md to the wrong path

`scripts/bash/plaesy-analyze.sh` and `scripts/powershell/plaesy-analyze.ps1` (+ their
`.plaesy/scripts/` copies) generated `context.md` and `memory.md` inside
`$MEMORY_DIR` = `.plaesy/memory/` — but every other part of the framework
(`instructions/plaesy.instructions.md`, `testing/bash/run.sh`,
`testing/smoke/smoke-e2e.sh`, `AGENTS.md`, `docs/scripts/plaesy-init.md`) expects them
at `.plaesy/context.md` and `.plaesy/memory.md` (repo root, not the `memory/`
subfolder). The PS1 version additionally used the filename `overview.md` instead of
`memory.md`, a second inconsistency vs. the bash version.

**Fix**: both generator functions in both scripts now target
`$PROJECT_PATH/.plaesy/{context,memory}.md`. PS1's `New-OverviewMd` renamed output to
`memory.md`. `.plaesy/memory/[topic].md` files (individual topic memory, e.g. this
file) are unaffected — that subfolder convention is correct and unchanged.

**Also changed the overwrite behavior**: previously both scripts skipped entirely if
the target file already existed ("preserving existing content"). Per user request,
changed to **append**: each run now adds a dated `## Analysis Update (<ISO timestamp>)`
section to the existing file instead of doing nothing (or, in the old buggy path,
nothing user-visible since the file was in the wrong place anyway). Verified via a
throwaway `/tmp` test project: first run creates the file with a `# Project
Context`/`# Project Overview` header + one section; second run appends a second
dated section, `wc -l` growing accordingly, original content untouched above.

## Files touched
- `scripts/configs/platform.json`, `.plaesy/scripts/configs/platform.json`
- `scripts/bash/plaesy-analyze.sh`, `.plaesy/scripts/bash/plaesy-analyze.sh`
- `scripts/powershell/plaesy-analyze.ps1`, `.plaesy/scripts/powershell/plaesy-analyze.ps1`
- `docs/scripts/plaesy-analyze.md` (location doc fixed: was `.plaesy/memory/memory.md`)

## Why this matters going forward
Confirms [[feedback_ps1_bash_parity]] again — bash and PS1 had drifted (memory.md vs
overview.md) and needed the same audit + fix on both sides. Also a reminder to
grep the *whole* codebase for a path reference before trusting one script's own
comments/docs about where it writes — the docs here were themselves wrong and would
have led to "fixing" the working files at root instead of the broken generator.
