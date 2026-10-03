---
name: Cross-Platform CI
description: Patterns that make a test pass on Windows but fail on Linux/macOS (or vice versa), found getting CI green for the first time since September
updatedAt: "2026-10-02T21:50:00.000Z"
---

# Cross-Platform CI

CI (`.github/workflows/ci.yml`, matrix `[ubuntu-latest, windows-latest,
macos-latest]`) had never passed since its first run in September — every run
failed on at least one OS. 2026-10-02's session got it green by fixing one
real bug and several OS-only test assumptions. The bug classes below are
worth recognizing fast if CI goes red on one OS again.

## `os.Getwd()` after `os.Chdir()` is not always the string you Chdir'd with

A test that does `os.Chdir(dir)` then later compares a result against the
literal `dir` string is comparing against the wrong thing on at least two
platforms:

- **macOS**: `t.TempDir()` lands under `/var/folders/...`, and `/var` is a
  symlink to `/private/var`. `os.Getwd()` after a `Chdir` into that path
  returns the kernel's resolved, symlink-free view —
  `/private/var/folders/...` — not the unresolved string `t.TempDir()`
  returned.
- **Windows (GitHub-hosted runners specifically)**: `GetCurrentDirectory` can
  resolve a long path segment (`runneradmin`) to its 8.3 short form
  (`RUNNER~1`). Worse: this was observed to be *inconsistent within the same
  test run* — production code's error named the long form while a second,
  independent `os.Getwd()` call made for comparison in the test returned the
  short form. Two `Getwd()` calls after the same `Chdir` are not guaranteed to
  agree with each other, let alone with the pre-Chdir string.

**Fix, in order of preference:**

1. Don't compare absolute paths across a Chdir boundary at all — assert on the
   OS-stable *relative suffix* instead (`strings.HasSuffix(err.Error(),
   filepath.Join(".plaesy", "specs", branch, "plan.md"))`), if the thing under
   test is a relative location, not an absolute identity.
2. If an absolute comparison is unavoidable, derive both sides from a fresh
   `os.Getwd()` call made at the same point in the same function — not from a
   pre-Chdir string, and not from two independent Getwd() calls hoping they
   agree.

Fixed under this pattern: `internal/graph/build_test.go`
(`TestResolvePaths`'s relative-path subtest), `internal/featurepath/featurepath_test.go`
(`TestGetPathsReport`'s git-failure fallback cases),
`cmd/plaesy/trim_cmd_test.go` (`inTrimRepo` now returns the resolved value),
`internal/agentcontext/update_exported_test.go` (`TestUpdateNoPlanFound`,
which needed the suffix-only form after the Getwd-vs-Getwd approach itself
proved unreliable).

## No `.gitattributes` means line endings are whatever `core.autocrlf` says

This repo has no `.gitattributes`. `windows-latest`'s default
`core.autocrlf=true` means a Windows checkout writes `README.md` (and every
other text file) with CRLF. Any regex or string match written assuming `\n`
breaks silently — not with a parse error, with "found 0 matches" — on that one
platform only. `fencedBlocks()` in `cmd/plaesy/docs_drift_test.go` required
` ``` ` + language + a bare `\n`; fixed to `\r?\n`. If a Windows-only test
failure says "found 0 of X" for something that definitely exists in the file,
suspect CRLF before suspecting the parser logic.

## `filepath.ToSlash` is not "normalize backslashes to forward slashes"

`filepath.ToSlash` only converts the *current OS's* `filepath.Separator`. On
Linux/macOS that separator is already `/`, so it is a no-op — a backslash that
reached the string some other way (recorded on a different OS, present in
test data) passes through unconverted. If the intent is "canonicalize this
path-shaped string regardless of host OS," use
`strings.ReplaceAll(s, `\`, "/")` instead. Found in `internal/graph/graph.go`'s
`toSlash()`.

## A function with a `runtime.GOOS` guard needs a test that knows about the guard

`rootViaCwd()` in `internal/common/gitpaths.go` is deliberately a no-op off
Windows (it repairs a Git-Bash-on-Windows MSYS path-mangling bug that does not
exist elsewhere). The test asserting it "works" unconditionally was the bug,
not the function — fixed with `if runtime.GOOS != "windows" { t.Skip(...) }`.
Before changing a GOOS-gated function because its test fails on another OS,
check whether the gate is deliberate.

## One coverage baseline cannot satisfy two OSes if the package forks on GOOS

`internal/cleaner`, `internal/common`, `internal/installer`, `internal/graph`
and `internal/mdlint` each have a real `if runtime.GOOS == "windows"` branch,
so the *statements reachable* — and therefore the measured percentage — differ
by platform, not by test quality. `coverage-baseline.json` has one number per
package with no OS dimension, and `TestCoverageDoesNotRegress` also fails a
baseline recorded more than 2 points *below* the measured number (to stop the
baseline from going stale-low and silently hiding real coverage loss). Those
two rules together mean there is no single number for an OS-forked package
that passes on both Windows and Linux: tight enough to catch a Linux
regression, and it is already "recorded too low" on Windows.

**Resolution**: `TestCoverageDoesNotRegress` now skips on every `GOOS` except
`linux`. `ubuntu-latest` is canonical, matching the reasoning `ci.yml`'s
"corpus" job already used to run on `ubuntu-latest` only rather than across
the matrix — one designated platform beats a race to the lowest common
denominator. See `[[coverage-ratchet-single-os]]` for the full record. A
per-OS-keyed baseline would be the complete fix; it was explicitly deferred as
future work, not attempted this session.

## Debugging workflow that got CI from 20+ failures to 0

1. `gh run view <id> --job <id> --log > ci-log.txt`, then
   `grep -n "FAIL:" ci-log.txt | grep -v "0 FAIL"` to list every failing test
   name without the noise.
2. Before touching anything, check whether the failure is new: `git stash`
   (include untracked), re-run the exact same test, compare. Several failures
   this session were pre-existing on the unmodified tree — fixing those is
   still worth doing, but don't attribute them to the change in front of you.
3. Fix one class of failure, push, re-watch the next run. Each of the 7 pushes
   in this session's remediation fixed one cause and revealed the next one
   underneath (fail-fast cancels the other matrix jobs before they run, so you
   only see the first OS's failures until it goes green). Expect several
   iterations, not one.
