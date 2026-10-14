---
title: Clear the blocking go vet failure and the shipped-command bookkeeping drift
phase: fix
status: done
createdAt: "2026-10-03T19:50:00.000Z"
updatedAt: "2026-10-04T03:55:00.000Z"
---

## Description

`/assess` on 2026-10-03 scored **0** — a blocking gate failure sets the score to
zero per `.plaesy/instructions/quality-gates.md`. Two causes, neither of them a
defect in the semantic-search feature work.

### 1. `go vet` failed with 52 errors, which silently disabled two packages' tests

`common.LogInfo/Warning/Error/Success(format string, args ...any)` forwards to
`fmt.Sprintf`, so `go vet` infers them as printf wrappers. 52 call sites passed a
runtime string into the format position, in 7 files. Pre-existing and untouched
since `1d46bea Initial Commit`.

**Not cosmetic — it was a live output-corruption bug.** Every one of the 52 sites
interpolates a filename or path, and `%` is legal in filenames on Windows,
macOS and Linux. `fmt.Sprintf` consumed the following text as a verb:

```
before:  ✓ report%!f(MISSING)inal.md
after:   ✓ report%20final.md
```

So `plaesy init`, `plaesy reload` and the prune path printed corrupted filenames
whenever a copied file contained a percent sign.

### 2. `plaesy search` shipped with none of its three bookkeeping edits

- `TestEveryShippedCommandHasAChangelogLine` — no `plaesy search` line
- `TestEveryCommandIsInTheReferenceIndex` — no row in the Command Index table
- `TestMetadataCountsMatchTheTree` — `docs/metadata.json` counts stale

The counts track the **tree**, not the command list, so they drift on any `.go`
file added or removed — not only when a command is added.

## Resolution

Fix the callers, not the logging API: each site became `LogX("%s", expr)`. This
is what `go vet` prescribes for a non-constant format string, it is exactly
output-equivalent where no `%` is present, and it repairs the corruption where
there is. Changing `LogInfo`'s signature was rejected — it would break every
correct caller that does use format verbs, and five near-identical
non-format siblings would have been duplication.

`gofmt -w` was then required on all 7 files: with two arguments instead of one,
gofmt's canonical spacing for `+` inside a call argument list differs from the
single-argument form, so the files were no longer gofmt-clean.

## Acceptance Criteria

- [x] `go vet ./...` exits 0 — zero `non-constant format string` errors
- [x] `go test ./...` reports **no** `[build failed]` lines; `cmd/plaesy` and
      `internal/scaffold` genuinely execute their tests
- [x] `CHANGELOG.md` Unreleased contains a `plaesy search` line
- [x] `docs/reference.md` Command Index contains a `search` row
- [x] `docs/metadata.json` counts re-derived from the tree
- [x] `validate markdown` — 527 files, 0 violations
- [x] Negative control: reintroducing one non-constant format string makes
      `go test` report `[build failed]` for that package (verified)

## Test Results

```
$ cd scripts && go vet ./...            # before: 52 errors
$ go test -count=1 ./...                # before: 2 × [build failed]

$ go vet ./...                           # after: exit 0
$ gofmt -l .                             # after: empty
$ go build ./...                         # after: exit 0
$ go test -count=1 ./... | grep -c 'build failed'
0
18 of 19 packages ok
```

One failure remains: `TestNoBuildArtifactIsTracked`, which reports three
`.plaesy/tasks/backlog/*.md` paths as tracked-but-absent. Those are this
project's own uncommitted `backlog/` → `done/` moves, not a defect — it clears
on commit.

### Coverage: unchanged at 88.6%

Re-measured after the fix: **88.6%**, identical to before. This corrects the
prediction made when the task was written, which expected the number to rise
once the two unmeasured packages were included. It did not, because the original
88.6% had already been collected with `-vet=off` and therefore *had* measured
`cmd/plaesy` (83.5%) and `internal/scaffold` (90.4%). The figure was complete;
the reasoning about it was wrong.

88.6% is below the `quality-gates.md` default of 90%, but nothing enforces it
here: there is no `.plaesy/memory/constitution.md`, and `ci.yml` sets no
coverage floor. `decisions/accepted-mdlint-coverage-dip.md` records a related
accepted dip.

## Prevention

- **The defect class is prevented structurally, not by a new test.** `go test`
  runs the `printf` vet check before compiling, so a new bad call site fails its
  own package's build rather than passing quietly. Verified by mutation: the old
  broken form produces `non-constant format string` at compile time.
- `TestLogRuntimeStringWithPercentIsNotCorrupted` in
  `internal/common/logging_test.go` covers the case vet *cannot* see — a future
  refactor that breaks argument substitution inside the helper. Mutation-verified
  both ways: reverting one subtest to the broken form fails at build; removing
  `Sprintf` from `LogInfo` fails the assertion with
  `stdout: "... \x1b[0m %s\n"`.
- `ci.yml:24-26` already runs `go vet ./...`. This gate was already wired and
  still did not catch the defect, because the repo has no CI history — `git log`
  shows only `1d46bea Initial Commit`.

## Reference

- [Testing Discipline](../../memory/testing-discipline.md) → "Reading a `go test
  ./...` result correctly" and "Shipping a new top-level command"
- `scripts/internal/common/logging.go:59-71` — the inferred printf wrappers
- `scripts/cmd/plaesy/metadata_drift_test.go:152-172` — the seven count fields
- `.github/workflows/ci.yml:24-26` — the existing `go vet` step

## Dependencies

None. Independent of the `medium_graph-include-ext-missing-hh-cxx` backlog task.

## Notes

Fixing the build failure immediately exposed a defect that had been hidden
behind it: `TestReadmeCommandListMatchesTheReferenceIndex` — the README's two
CLI command blocks had never been checked for `plaesy search`, because
`cmd/plaesy` had never once compiled under `go test`. Both blocks now list it.
This is the clearest argument for Gate 2: a package that cannot build is not a
package with a red test, it is a package with no signal at all.