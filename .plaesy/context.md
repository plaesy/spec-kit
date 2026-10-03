---
title: "Session Context"
updatedAt: "2026-10-04T04:00:00.000Z"
phase: fix
status: done
---

## Context

`plaesy search`: embedding-based semantic search over extracted symbols, wired
into the Anti-Duplication Protocol in `instructions/plaesy.instructions.md` as
step 1b (literal `find`/`grep` remains 1a). Decision trail:
`.plaesy/decisions/embedding-based-semantic-duplicate-search.md`. Durable
lessons: `.plaesy/memory/semantic-search-feature.md`,
`.plaesy/memory/testing-discipline.md`.

## Session Outcome

Feature work complete, then the blocking gate failure found by `/assess` was
fixed in `/fix`. Gates are now green except for one artifact of uncommitted
work.

Delivered across 6 subagents on disjoint files, one writer per file: one
`docsByExt` table, 24 extensions, all 15 languages `symbols.go` covers now
covered. Full detail in the decision record and the `done/` task files.

## The vet fix — it was a real bug, not a lint nit

`common.LogInfo(format, args...)` forwards to `fmt.Sprintf`, so vet infers it
as a printf wrapper. 52 call sites in 7 files passed a runtime string into the
format position. Because those sites interpolate **filenames and paths**, and
`%` is legal in filenames on all three target OSes, output was being corrupted:

```
before:  ✓ report%!f(MISSING)inal.md
after:   ✓ report%20final.md
```

Fixed by wrapping each site as `LogX("%s", expr)` — vet's prescribed fix,
output-identical where no `%` is present, and a genuine repair where one is.
`gofmt -w` was then required on all 7 files (two-argument call lists space `+`
differently).

**Why this survived since the initial commit:** CI *does* run `go vet ./...`
(`ci.yml:24-26`), but the repo has no CI history — `git log` shows only
`1d46bea Initial Commit` — and the local convention of not running the full
suite without permission meant nothing caught it either.

## Gates — after `/fix`, 2026-10-04

`go build ./...` clean · `gofmt -l` clean · `go vet ./...` **exit 0** (was 52
errors) · `go test ./...` **0 × `[build failed]`** (was 2 packages
unexecuted) · 18/19 packages ok · coverage **88.6%** (unchanged) ·
`validate markdown` 527 files, 0 violations.

Remaining failure: `TestNoBuildArtifactIsTracked` — three
`.plaesy/tasks/backlog/*.md` paths reported tracked-but-absent. These are this
project's own uncommitted `backlog/` → `done/` moves, not a defect; clears on
commit.

**Fixing the build exposed a defect it had been hiding:**
`TestReadmeCommandListMatchesTheReferenceIndex` failed — the README's two CLI
blocks never listed `plaesy search`, and the guard lives in the package that
could not compile. Both blocks now list it.

Coverage: 88.6% is under the 90% default but unenforced — no constitution, no
`ci.yml` floor. See `decisions/accepted-mdlint-coverage-dip.md`.

## Tasks

`backlog/`: 1 · `todo/`: 0 · `doing/`: 0 · `done/`: 5 · `blocked/`: 0

- **done** `high_fix-vet-and-command-bookkeeping-drift` — 52 vet errors cleared,
  `%`-corruption fixed, CHANGELOG + `docs/reference.md` + `docs/metadata.json`
  refreshed, README blocks filled, prevention test added and mutation-verified.
- **backlog** `medium_graph-include-ext-missing-hh-cxx` — `internal/graph` has
  three out-of-sync extension lists; `.hh`/`.cxx` are handled by `symbolsOf` but
  never collected, so those two `docsByExt` entries can never fire. Widening
  collection changes what `plaesy analyze` gathers for every project, so verify
  with a real graph rebuild, not just a unit test.

## Structural Gap — no constitution

`.plaesy/memory/constitution.md` does **not exist**, so every
dimension-dependent route is undefined here, and the 90% coverage default has
no project override. Adding one is an explicit decision, not auto-recovery.

## Next

- **Commit the uncommitted work** — 14 new `.go` files, the `docs/scripts/
  plaesy-search.md` edit, the bookkeeping fixes, and the `backlog/` → `done/`
  task moves. This is the last thing keeping a gate red.
- Re-run `/assess` to confirm the score is no longer pinned to 0 by a blocking
  gate, and to get a real technical number for the first time.
- Optional: extractors still carry private copies of `symbols.go`'s symbol-name
  regexes. Guarded by drift tests, but exporting one helper from `internal/graph`
  would remove the duplication.
- Older carry-over, never opened: a bloat audit of the four largest
  `*.instructions.md` files (`brandkit` 854 lines, `react-native` 556,
  `dart-n-flutter` 503, `nestjs` 404), and `prompts/` not yet spot-checked for
  `fix`, `doc`, `continue`, `optimize`, `improve`, `start`, `create`, `spec`.