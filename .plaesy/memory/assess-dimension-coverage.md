---
title: "Assess Dimension Coverage & mapping.json Install Bug"
updatedAt: "2026-09-15T00:00:00.000Z"
---

# Assess Dimension Coverage & mapping.json Install Bug

## What happened
User asked (in Indonesian) whether spec-kit's `/assess` only covers technology.
Investigation showed 8 dimensions exist on paper (`assess-*.instructions.md` +
matching chatmodes), but Financial and Marketing had no dedicated instruction file —
both silently fell back to `assess-business.instructions.md`.

## Fix 1: New dedicated instruction files
- Created `instructions/assess-financial.instructions.md` — cost structure, pricing,
  profitability/unit economics (LTV:CAC, payback period), cash flow/runway, financial
  risk. Numbers-over-narrative rules, mandatory data-source citation.
- Created `instructions/assess-marketing.instructions.md` — positioning, messaging,
  audience segmentation, competitive messaging, GTM readiness. Mandatory live web
  research for competitor claims (no recall-based competitive analysis).
- Updated `prompts/assess.md` scope table (line ~194-197): `marketing`/`financial`
  scopes now load their own file instead of `assess-business.md`.
- Narrowed `assess-business.instructions.md` scope line to Business only, added
  cross-references to the two new files.
- Updated `instructions/dimension-mapping.instructions.md` "See Also" section.

## Fix 2 (root cause, bigger than the 2 new files): install-time gap
Traced the actual install pipeline (`copy_instructions()` in `plaesy-init.sh`, its
PowerShell twin, and `detect-stack.sh`/`.ps1`) — all driven **solely** by
`instructions/mapping.json`'s `always_load` + keyword-category lists. Confirmed the
`platform.json` → `"instructions"` copy path in `setup_platform_config()` is dead
code: `platform_mappings="prompts"` only, `"instructions"` is never in the loop.

Found **none of the 8 `assess-*.instructions.md` files were registered in
`mapping.json`** — meaning `/assess:*` would ship with zero instruction files on any
fresh `plaesy init`, for every dimension, not just Financial/Marketing.

Also found 4 more files completely orphaned (zero references anywhere in the repo,
not just missing from mapping.json):
- `dimension-mapping.instructions.md`
- `error-recovery-predictive.instructions.md` (companion to `error-recovery.instructions.md`,
  which *was* already in `always_load`)
- `output-validation.instructions.md`
- `universal-orchestrator.instructions.md`

And 2 specialized/optional files also orphaned:
- `brandkit.instructions.md`, `redesign.instructions.md`

## Fix applied
- Added all 8 `assess-*.instructions.md` + the 4 foundational orphans to
  `mapping.json`'s `always_load` (18 entries total, up from 7).
- Added `brandkit`/`redesign` as new keyword-triggered entries under `methodologies`
  (not always_load — they're opt-in specializations) so they're reachable via
  `detect-stack.sh` instead of dead weight.
- Synced the bash hardcoded fallback list in `plaesy-init.sh` (used only when `jq`
  is unavailable) to match. PowerShell needed no code change — it reads
  `mapping.json` dynamically, no hardcoded fallback list exists there.

## Verification performed (not just code reading)
- Ran `bash scripts/bash/detect-stack.sh "."` before and after — confirmed all 8
  `assess-*` files now appear in output.
- Validated `mapping.json` is well-formed via `ConvertFrom-Json` in PowerShell (no
  `jq` available in this shell).
- Diffed all `instructions/*.instructions.md` on disk against everything referenced
  in `mapping.json` — remaining gap is only tech/keyword-triggered files that don't
  match this repo's own content (expected, not a bug) plus `agents.instructions.md`
  (intentionally handled via the separate "core" path → CLAUDE.md/AGENTS.md).
- Cross-checked chatmodes referenced vs on disk, templates referenced vs on disk
  (both directions), and prompts/*.md vs README.md slash-command list — all clean,
  no other broken references found.

## Established rule for future sessions
Any new `instructions/*.instructions.md` file MUST be added to
`instructions/mapping.json` (either `always_load` if foundational/always-needed, or
a keyword-category entry if project/tech-specific) or it will silently never be
installed into any project's `.plaesy/instructions/`. There is no other live path
that copies instruction files.
