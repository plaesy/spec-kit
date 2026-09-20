---
title: "Feedback: verify a capability isn't implemented elsewhere before deleting its reference"
updatedAt: "2026-09-15T00:00:00.000Z"
---

# Feedback: verify before deleting a "doc says X, code doesn't do X" reference

**Rule**: when an audit finding is "the doc references path/capability X, but the
script I checked doesn't read/write X" — before deleting that reference, search
the rest of the repo for a DIFFERENT script that might implement X under a
similar-but-not-identical path. Don't conclude "X doesn't exist" from "this one
script doesn't do X."

**Why**: In the full-repo-audit session ([[full-repo-audit-2026-09-15]]),
`mapping.json` referenced `.plaesy/analysis/project.json` for stack detection.
`detect-stack.sh` didn't read it, so I deleted the reference. The user asked
"why — doesn't `plaesy analyze` create that file?" Checking `plaesy-analyze.sh`
(a script my audit's finders never even looked at) showed it does write a
`project.json` — just at `.plaesy/analysis/project.json`, one path
segment different. The real bug was a path typo plus `plaesy-init.sh` creating
the wrong empty directory — not a nonexistent capability. Deleting instead of
correcting would have silently dropped a real, valuable data source (rich
`technology_stack`/`frameworks_detected` JSON) from ever informing instruction
auto-detection.

**How to apply**: before deleting any reference that describes a *capability*
(a file some tool is supposed to produce or consume) — as opposed to a plain
dead link with no implied producer/consumer — grep the repo for the bare
filename (`project.json`, not the full path) across all scripts, not just the
one script under review. If a producer exists anywhere, fix the path; only
delete when genuinely no producer/consumer exists anywhere in the repo.
