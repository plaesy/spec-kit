---
title: "Coverage ratchet (TestCoverageDoesNotRegress) runs on ubuntu-latest only"
updatedAt: "2026-10-02T21:50:00.000Z"
---

# Coverage ratchet is single-OS (linux)

**Decision**: `internal/quality/coverage_ratchet_test.go`'s
`TestCoverageDoesNotRegress` now skips immediately on every `runtime.GOOS`
except `"linux"`. `scripts/coverage-baseline.json`'s recorded percentages are
ubuntu-latest's measured numbers.

**Why**: `internal/cleaner`, `internal/common`, `internal/installer`,
`internal/graph` and `internal/mdlint` each have a real
`if runtime.GOOS == "windows"` fork, so the statements actually reachable by a
test run — and therefore the measured coverage percentage — differ by
platform, not by test quality. Measured directly: cleaner 93.3%(win)/90.5%(linux),
common 89.3/83.2, installer 71.6/70.1, graph 98.4/98.5, mdlint 87.9/88.1. The
ratchet test enforces two rules at once — never regress, and never let the
baseline sit more than 2.0 points below the current measurement (so a stale
number can't quietly hide real coverage loss) — and for every one of those
five packages, no single number satisfies both rules on both Windows and
Linux: tight enough to catch a real Linux regression, and it already reads as
"recorded too low" the moment someone runs it on Windows and gets a higher
number. Lowering the entries to the smaller of the two measurements just
converted a "regression" failure on one OS into a "recorded too low" failure
on the other — tried, and rejected as a dead end during this session.

**Rejected alternative**: a per-package epsilon wide enough to cover the
observed OS variance (~3-7 points depending on the package). Rejected because
it defeats the ratchet's actual purpose for every *other* package: a real
single-line regression in, say, `internal/validate` would then also fall
inside that same wide tolerance and go undetected.

**Precedent this follows**: `.github/workflows/ci.yml`'s "corpus" job already
runs `plaesy validate markdown` / `plaesy validate memory` on `ubuntu-latest`
only, not across the matrix — because those checks don't need multiple OSes to
mean something, and picking one avoids exactly this class of problem. The
coverage ratchet is the same shape of check.

**Deferred, not solved**: the complete fix is a baseline keyed by `GOOS`
(`{"linux": {...}, "windows": {...}, "darwin": {...}}` or three separate
files), which would let the ratchet actually enforce "no regression" on every
platform instead of silently trusting Windows/macOS test runs not to regress.
That is a real gap — explicitly left as future work rather than attempted in
this session, which was already mid-way through getting CI green for the
first time since the workflow's introduction in September.

**Verified**: `go test ./internal/quality -run TestCoverageDoesNotRegress -v`
on Windows now prints `--- SKIP` with the reason, not a false pass; the
`ubuntu-latest` CI job's own run of the same test is what gates the baseline
for real.
