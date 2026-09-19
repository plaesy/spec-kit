# plaesy analyze: overview.md Replaces context.md/memory.md Appending (2026-09-16)

## Problem
`plaesy analyze` was appending a dated `## Analysis Update` section to both `.plaesy/context.md`
and `.plaesy/memory.md` on every run (a behavior introduced earlier the same day, see
[[plaesy-analyze-path-fix-2026-09-16]]). User found this noisy — running `analyze` repeatedly grew
both files with duplicate boilerplate, polluting files meant for manual/curated notes.

## Fix
In both `scripts/powershell/plaesy-analyze.ps1` and `scripts/bash/plaesy-analyze.sh`:
- Removed `New-ContextMd`/`New-OverviewMd` (PS1) and `generate_context_md`/`generate_overview_md`
  (bash) — these wrote/appended to `context.md`/`memory.md`.
- Added a single replacement function — `New-AnalysisOverviewMd` (PS1) /
  `generate_analysis_overview_md` (bash) — that always **overwrites**
  `.plaesy/analysis/overview.md` with a fresh snapshot (tech stack, file counts, recommendations,
  timestamp). No append, no existence check.
- `context.md` / `memory.md` are no longer touched by `analyze` at all — they're pure
  manual/append-preserved notes now.
- Internal cross-references inside the generated snapshot were changed from
  `analysis/project.json` → `project.json` (no `analysis/` prefix) because `overview.md` now
  *lives inside* `.plaesy/analysis/`, so siblings are referenced with plain relative paths — the
  old prefix was correct when context.md/memory.md (at `.plaesy/` root) referenced files one
  level down, but would be wrong (double `analysis/analysis/`) from inside the same folder.
- `project.json`'s `analysis_files.overview` field (bash only; PS1's project.json has no such
  block — pre-existing parity gap, not touched) updated from `"../memory.md"` → `"overview.md"`.
- Stale `## Analysis Update (...)` blocks already appended to `context.md`/`memory.md` earlier
  that same session were manually trimmed out (dead weight now that analyze won't add more).

## Rule going forward
`.plaesy/analysis/overview.md` is the disposable, always-regenerated analysis snapshot.
`.plaesy/context.md` / `.plaesy/memory.md` are manual/curated and `plaesy analyze` must never
write to them again. This supersedes the "append-dated-section on rerun" rule recorded in
[[plaesy-analyze-path-fix-2026-09-16]] for context.md/memory.md specifically.
