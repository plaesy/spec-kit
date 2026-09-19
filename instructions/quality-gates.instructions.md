---
description: "Automated validation gates shared by /implement, /assess, /optimize, /fix"
---

# Quality Gates

**Referenced by**: `/implement`, `/assess`, `/optimize`, `/fix`, `/doc`, `/save` —
each loads this file as `.plaesy/instructions/quality-gates.md`.

## Purpose

A single automated checkpoint every phase runs before declaring output "done", so
standards aren't re-implemented (and re-drifted) per prompt. Values below are
**defaults** — a project's `.plaesy/memory/constitution.md`, when present, overrides them.

## Gate Checklist (run in order; stop at first blocking failure)

1. **Build** — project compiles/builds cleanly. Blocking.
2. **Tests** — full suite passes. Blocking.
3. **Coverage** — meets `{{TEST_COVERAGE_MIN|90}}%` (or constitution override). Blocking.
4. **Security** — no known vulnerability at or above the constitution's security bar
   (default: Sev-High). Blocking.
5. **Lint/Static analysis** — no new errors introduced. Blocking.
6. **Performance** — meets the constitution's performance target where measurable.
   Non-blocking (flag, don't halt) unless the constitution marks it a hard stop.
7. **Accessibility** (frontend only) — WCAG 2.1 AA. Blocking for UI work.
8. **Documentation** — public interfaces documented per constitution's doc bar.
   Non-blocking (route to `/doc`).

## Verifier Separation (Required for /implement)

The agent that writes the implementation must not be the sole judge of gate 1–5.
Before marking implementation complete, a **fresh review pass** — a new agent
invocation with no prior implementation context, given only the spec/constitution and
the diff — checks:
- Does the diff satisfy every acceptance criterion in the spec?
- Does it violate any constitution non-negotiable?
- Are there claims in the implementer's self-report not actually reflected in the diff?

Findings from this pass are treated as gate failures, not suggestions.

## Output Format

```
QUALITY GATES: [PASS/BLOCKED]
1. Build:          [PASS/FAIL]
2. Tests:          [PASS/FAIL] (N passing / M failing)
3. Coverage:       [XX%] (target: YY%)
4. Security:       [PASS/FAIL] (N findings at/above bar)
5. Lint:           [PASS/FAIL]
6. Performance:    [PASS/FLAG] (measured vs target)
7. Accessibility:  [PASS/FAIL/N-A]
8. Documentation:  [PASS/FLAG]
Verifier pass:     [PASS/FAIL] (see findings)
```

## On Failure

- Blocking failure → halt, report which gate + why, route per
  `.plaesy/instructions/error-recovery.md`.
- Non-blocking flag → continue, but surface in the phase's completion report so
  `/assess` picks it up.
