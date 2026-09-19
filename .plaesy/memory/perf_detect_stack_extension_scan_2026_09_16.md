---
name: perf_detect_stack_extension_scan_2026_09_16
description: detect-stack.sh/.ps1 extension-presence check ran one directory walk per extension; batched into a single walk in both scripts
metadata:
  type: project
---

## What was slow

The `extensions`-array presence check added in [[mapping-json-extensions-audit-2026-09-16]]
ran one recursive directory walk (`find` / `Get-ChildItem -Recurse`) **per extension per
mapping.json entry** — 13 extensions across 6 entries meant up to 13 separate tree walks
just for that one feature, on top of the manifest-scan walks already in the script.

## Fix

Collect every entry's extensions up front, then do **one** recursive walk that finds
every matching extension at once; check membership against that result set instead of
re-walking per extension.
- Bash (`scripts/bash/detect-stack.sh`): build `all_ext_args` (one `-o -iname "*.ext"`
  per unique extension) and issue a single `find`; also merged the separate `*.csproj`
  manifest-content scan into the main manifest `find` (wildcard `-name` join, `mindepth`
  unified to 1) instead of a second full-tree pass.
- PowerShell (`scripts/powershell/detect-stack.ps1`): collect `$allExtensions` across all
  categories, one `Get-ChildItem -Recurse -Include <patterns>` call, then `Test-Category`
  checks `$foundExtensions.Contains($ext)` instead of calling `Get-ChildItem` itself.

Also merged the separate `.tsx`/`.jsx` presence check into the same single bash `find`
call (was a 2nd tree walk right after the extensions one) — down to exactly 1 `find` for
all extension-based detection combined.

## Verified

Output byte-identical before/after in both scripts (`diff`/`Compare-Object` on sorted
output, this repo as target — 21 bash lines / 22 ps1 lines, zero diff), re-confirmed
after the tsx/jsx merge too.

Timing — **corrected after multi-round, warm-cache re-measurement** (the first pass's
numbers below were cold-cache noise, not a reliable measurement; see caveat):
- **bash/MSYS**: initial single-shot measurement showed 5.1s → 3.35s, but 5 back-to-back
  rounds with warm OS page cache put both before and after at ~1.1-1.5s with no
  consistent winner — the wall-clock difference is inside measurement noise on this
  repo's size. The algorithmic improvement (≤15 `find` calls → 2, then → 1 after the
  tsx/jsx merge) is real and correct, but **do not repeat the "34% faster" claim for
  bash** — it wasn't reproducible.
- **powershell**: 4 back-to-back rounds, consistently 15.4-20.4s before → 7.6-15.1s
  after (settling to ~8s once warm) — a real, reproducible ~50-60% improvement. Windows
  `Get-ChildItem -Recurse` is expensive enough per call that cutting ~13 walks to 1 has
  a measurable effect even accounting for `powershell.exe` startup overhead.

**Caveat for future perf claims in this repo**: always take multiple back-to-back
timed runs (not one before + one after) before reporting a percentage — a single
cold-vs-warm comparison overstates the win. This was a real mistake made and corrected
in the same session; see [[feedback_measure_perf_multiple_rounds]] if that memory
exists, otherwise treat this note as the record of it.

## Why this matters for future edits

Any future addition to `mapping.json`'s `extensions` mechanism must go through the
shared up-front collection (`all_ext_args` in bash, `$allExtensions` in ps1) — adding a
per-entry `find`/`Get-ChildItem` call back in would silently reintroduce the O(entries ×
extensions) tree-walk cost this fix removed. [[feedback_ps1_bash_parity]]
