---
name: Testing Discipline
description: How tests are classified, mutated and trusted in this repo
updatedAt: "2026-10-03T19:45:00.000Z"
---

# Testing Discipline

- **Do not run `go test ./...` without explicit permission.** Scoped
  `go test ./cmd/plaesy -run <Name>` for a specific guard is acceptable.
- **A red test suite is a finding, not background noise.** Never record failing
  tests as "pre-existing, unrelated" on the strength of which files the current
  diff touched. That classification was accurate about the diff and wrong about
  the world: 21 failures across `clean`/`config`/`platforms` were one root cause,
  and the most serious defect in the repository was visible only as a *passing*
  test over a switch statement that could never reach its case. Scoped runs are a
  way to *narrow* a red suite, not to declare part of it out of scope.
- **A guard that is right about the fact can still be wrong about the meaning.**
  `TestNoFileDuplicatesAnotherOutsideTheKnownMirrors` correctly observed that
  `AGENTS.md` is byte-identical to `instructions/agents.instructions.md`; calling
  that a stray copy described the framework deleting correct behaviour. The
  generated target set is now read from `platform.json` instead of listed in the
  test, because two platforms already share `AGENTS.md` and a third would fail
  the guard for doing its job.
- **`git check-ignore` silently skips paths already in the index.** Any guard
  asking "is this path ignored?" about tracked files must pass `--no-index`, or it
  answers "not ignored" for exactly the set under test. Found by mutation testing:
  the first version of `TestSessionStatePathsAreTrackable` passed with the fix
  deleted.
- **This project has now been bitten four times by a green mutation result.** The
  recurring lesson is the same: a new guard is not proven until you delete the
  thing it guards and watch it fail. A test that has never been seen to fail has
  not been shown to detect anything.
- **"Don't run `go test ./...` without permission" is a rule with a real cost,
  and 2026-10-03 priced it.** A 52-error build failure sat undetected in
  `cmd/plaesy` and `internal/scaffold` because nothing ever ran the full suite
  locally — and it was suppressing every guard in those two packages. The rule
  still stands for exploratory work, but it does **not** apply when a phase
  mandates the full suite as an exit criterion: `/fix`'s Self-Audit requires
  "full test suite passes, no new failures", and `/assess` Gate 2 requires the
  same. In those phases, run it.
- **Verify end to end, not only through the suite.** The platform-id bug
  (`platform.json` keyed six platforms short while docs, `--ai` help and
  `promptExtension()` used the long form) was invisible to unit tests: it made
  `clean --ai claude_code` remove nothing **and report that it had**, and made
  Copilot prompts install as `.md` because that case could never fire. When one
  source disagrees with several, fix the one — here the 6 config keys were renamed,
  not the ~24 doc references, which also restored exact agreement with the list
  at `docs/overview.md:46`.
- **Confirm what a metric counts before reporting it.** Scoring prompts by
  `## Protocol` heading *length* concluded `loop.md` was hollow; it has a 180-line
  protocol under a different name. Write a script that actually does what its log
  line claims: one fixer printed confident per-file success while `" " + l[1:]`
  made its rewrite a no-op.
- **An automated fixer can satisfy the ratchet instead of the requirement.** A
  Markdown fixer agent renamed a heading specifically to satisfy a check that had
  been written moments earlier. Review automated corpus edits as changes, not as
  progress.

## Reading a `go test ./...` result correctly

- **`FAIL <pkg> [build failed]` means zero tests ran — it is not a test
  failure.** `go test` runs a vet subset (`printf` among them) before compiling,
  so a single vet error makes the entire package report `[build failed]` and none
  of its tests execute. Two packages here (`cmd/plaesy`, `internal/scaffold`)
  carried 52 `non-constant format string` errors and had been silently
  unexecuted. Re-run with `-vet=off` to get the real pass/fail picture before
  quoting a test count or a coverage figure — a piped `go test ./... | tail`
  hides this completely, because the line looks like any other failure.
- **A package that cannot build is not a package with a red test; it is a
  package with no signal at all — and the defects it hides surface the moment
  it builds.** Fixing the 52 vet errors immediately exposed
  `TestReadmeCommandListMatchesTheReferenceIndex` failing: the README's two CLI
  command blocks had never listed `plaesy search`, and the guard that says so
  lives in the package that could not compile. Two real defects (the missing
  README rows, the `docs/metadata.json` drift) had been sitting behind a
  build failure rather than reported by it. Never report a suite as "passing
  except known items" while a package in it is reporting `[build failed]`.
- **Coverage figures collected under `-vet=off` are complete, not
  partial.** The assumption that excluding two packages left the total
  understated was wrong — `-vet=off` runs those tests, it only suppresses the
  vet gate. Re-measured after the fix, the total was byte-identical. Check how
  a number was collected before predicting it will move.
- **"Pre-existing" needs the right check, and `git status` is not it.** An `M`
  in `git status` means "modified relative to HEAD", and HEAD is the initial
  commit here — it does not mean "changed by an earlier session". The check that
  actually establishes pre-existence is `git status --porcelain <files>` (empty
  output = untouched *now*) plus `git log --oneline -- <files>` showing nothing
  but the initial commit. An earlier note here attributed the vet failures to
  `git status` showing those files "already modified", which is not evidence of
  anything.
- **A dirty worktree makes `TestNoBuildArtifactIsTracked` fail spuriously.** It
  compares git-tracked paths against the working tree, so any uncommitted file
  *move* (e.g. `tasks/backlog/x.md` → `tasks/done/x.md`) reports the old path as
  "tracked but not present in the working tree". A commit clears it. Mid-session,
  that failure is an artifact of uncommitted work, not a tracked-build-artifact
  defect — check `git status` before diagnosing it.

## Shipping a new top-level command: three bookkeeping edits, not one

Adding a `plaesy <verb>` command starts **three** drift guards, and a feature is
not done until all three are satisfied — `plaesy search` shipped with none, so
they surfaced only at `/assess` a session later:

| Guard | File to edit |
|---|---|
| `TestEveryShippedCommandHasAChangelogLine` | a `plaesy search` line in `CHANGELOG.md`'s Unreleased section |
| `TestEveryCommandIsInTheReferenceIndex` | a row in `docs/reference.md`'s Command Index table |
| `TestMetadataCountsMatchTheTree` | `docs/metadata.json` counts |

The third is the one that gets missed for the wrong reason: it is easy to assume
it only tracks commands, so a change that adds no command looks exempt. It does
not. `metadata_drift_test.go` re-derives `goTestFiles`, `goTestFunctions`,
`goFilesCmd`, `goFilesInternal`, `internalPackages`, `cobraCommandConstructors`
and `topLevelRegisteredParents` from the tree, so **any** added or removed `.go`
file moves it. Refresh `docs/metadata.json` at the end of every feature that
touches Go files, whether or not it adds a command.
