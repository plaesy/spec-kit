---
description: "Automated validation gates shared by /implement, /assess, /optimize, /fix"
applyTo: "**/*"
---

# Quality Gates

**Referenced by**: `/implement`, `/assess`, `/optimize`, `/fix`, `/doc`, `/save` —
each loads this file as `.plaesy/instructions/quality-gates.md`.

## Purpose

A single automated checkpoint every phase runs before declaring output "done", so
standards aren't re-implemented (and re-drifted) per prompt. Values below are
**defaults** — a project's `.plaesy/memory/constitution.md`, when present, overrides them.

## Quality Thresholds

**The two thresholds above are defined here and nowhere else.** A prompt that
needs one writes `{{QUALITY_REMEDIATION_FLOOR}}` or `{{QUALITY_SKIP_OPTIMIZE}}`
and may append the default in parentheses for readability — never the bare
number on its own, because a bare number is a second definition waiting to
drift. Four commands were previously
routing on four different numbers (80, 85, 85, 90) that no file reconciled, and
a project scoring 82 satisfied `/optimize`'s precondition while failing
`/continue`'s optimize branch — eligible and ineligible at the same time.

There are two numbers, not three, and they have different jobs:

| Threshold | Default | Job |
|---|---|---|
| `{{QUALITY_REMEDIATION_FLOOR}}` | **85** | Below this the work is **remediation**. Route to `/fix` or `/loop`. `/optimize` refuses to start. |
| `{{QUALITY_SKIP_OPTIMIZE}}` | **90** | At or above this *and* the dimension's mandatory audit passing, optimization has nothing left to add. |

**Rules that follow from the relationship between them:**

1. These gate on the **dimension score `/assess` Mode 2 reports**, not on a
   metric the command measures for itself. A command that cannot cite a Mode 2
   score does not apply the threshold.
2. `/optimize` uses `{{QUALITY_REMEDIATION_FLOOR}}` as its precondition, not a
   lower number of its own. Optimizing rubble is fixing, and a lower optimize
   threshold is how a project ends up eligible for and ineligible for `/optimize`
   in the same run.
3. `/loop --quality-target` defaults to `{{QUALITY_REMEDIATION_FLOOR}}`. The loop
   widens its axis when it reaches the target — the same rule as `/continue`'s
   routing, not a second, quieter threshold.
4. The two numbers are ordered. A project cannot set a skip threshold below the
   remediation floor; if it does, the floor governs and the skip threshold is
   ignored.

### Grade bands

The 0–100 scale has the same band edges in every dimension. Only the
*description* of each band is dimension-specific, so the edges live here and
`assess-{dimension}.md` supplies its own column. Nine files previously restated
these edges, which is how `assess-technical.md` came to demand "score ≥80" while
`/continue` routed on 85 for the same number.

| Score | Grade | General meaning |
|---|---|---|
| 95–100 | A+ | Nothing known to be wrong; thresholds met with margin. |
| 90–94 | A | Works well; only minor improvements remain. |
| 85–89 | B+ | Viable, with some assumptions or risks left open. |
| 80–84 | B | Works, but with concerns to address next iteration. |
| 70–79 | C | Acceptable; significant validation or remediation needed. |
| <70 | D | Needs work; blocker-class issues are present. |

Two bands are load-bearing beyond their letter:

- **90 is `{{QUALITY_SKIP_OPTIMIZE}}`** — the first band where optimization has
  nothing to add.
- **85 is `{{QUALITY_REMEDIATION_FLOOR}}`** — the first band where the work is
  optimization rather than remediation.

A dimension scores **0** when a blocking gate fails, and is capped at **49** when
its mandatory audit fails. Both override the band table: a project with a red
build is not a B, it is a zero.

## Placeholders

`{{NAME|default}}` marks a value the agent supplies at emit time. The `|default`
is not decoration — it is the value to use when nothing else determines it, and
it is written so the site is readable without this file.

**Substitute every placeholder before emitting the output.** A literal
`{{CURRENT_DATE}}` reaching a user is a defect: a citation reading
`retrieved {{CURRENT_DATE}}` is not a citation, and a gate reporting
`{{TEST_COVERAGE_MIN|90}}%` cannot be failed or passed. `plaesy validate
constitution` treats a surviving `{{...}}` as a hard stop, and the same rule
applies to anything you write.

| Placeholder | Default | Supplied from |
|---|---|---|
| `{{CURRENT_DATE}}`, `{{CURRENT_YEAR}}`, `{{CURRENT_MONTH}}`, `{{CURRENT_DAY}}` | see `date-system.md` | **Owned by `date-system.md`**, which defines the format and the auto-replacement rule. Do not define or re-spell them here — a second spelling is how a prompt and the file it cites emit different strings. |
| `{{TEST_COVERAGE_MIN\|90}}` | 90 | The constitution's row for the active dimension |
| `{{QUALITY_REMEDIATION_FLOOR\|85}}` | 85 | `## Quality Thresholds` above |
| `{{QUALITY_SKIP_OPTIMIZE\|90}}` | 90 | `## Quality Thresholds` above |

A constitution may override a default. It may not override the *role* of a
threshold — see rule 2 under **Quality Thresholds**.

## Closing check

These two questions apply at the end of every phase. Prompts that carry a
Self-Audit checklist reference this section rather than restating the
questions:

- **Ready for the next phase?** — the next phase's precondition is met, or the
  specific thing that is missing is named. "Mostly ready" is not an answer.
- **Would you ship this to production?** — and if not, say what blocks it.

Neither question is satisfied by having produced output. A phase that ran no
gates cannot answer the first one affirmatively, and one that cannot answer the
second must not report completion.

## Gate Checklist (run in order; stop at first blocking failure)

1. **Build** — project compiles/builds cleanly. Blocking.
2. **Tests** — full suite passes. Blocking.
3. **Coverage** — meets `{{TEST_COVERAGE_MIN|90}}%` (or constitution override). Blocking.
4. **Security** — no known vulnerability at or above the constitution's security bar
   (default: Sev-High). Blocking.
5. **Lint/Static analysis** — no new errors introduced. Blocking.
6. **Performance** — meets the constitution's performance target where measurable.
   Non-blocking (flag, don't halt) unless the constitution marks it a hard stop.
7. **Accessibility** (frontend only) — WCAG 2.1 AA. Blocking for UI work. For any
   UI marked complete, this gate also requires a **click-through pass**: run/build
   it, click every interactive element (buttons, links, forms, tabs, modals,
   toggles, nav) one at a time, and confirm each produces its real effect — a
   dead control (does nothing) is a gate failure, not a cosmetic note. Exercise
   every shipped theme and the mobile breakpoints during the pass. Report the
   click-through as an evidence list, element by element (e.g. "Signup -> opens
   /signup, no console errors"), in the Output Format below — a PASS claimed
   without that list is not a PASS. If the deliverable cannot be run (a static
   mockup or chat-only output), say so explicitly and verify each element by
   code inspection instead of claiming it was clicked. (Convention: anti-slop
   R-35 "Verify Before You Deliver", github.com/miqdadbadjuber/anti-slop,
   retrieved 2026-09-29.)
8. **Documentation** — public interfaces documented per constitution's doc bar.
   Non-blocking (route to `/doc`).
9. **Cleanup** — temporary/scratch files (test scripts, throwaway helpers) created
   during iteration removed before marking complete. Non-blocking (flag, don't
   halt) unless the constitution marks it a hard stop. (Anthropic Claude
   prompting best practices, retrieved 2026-09-25.)

## Verifier Separation (Required for /implement)

The agent that writes the implementation must not be the sole judge of gate 1–5.
Before marking implementation complete, a **fresh review pass** — a new agent
invocation with no prior implementation context, given only the spec/constitution and
the diff — checks:

- Does the diff satisfy every acceptance criterion in the spec?
- Does it violate any constitution non-negotiable?
- Are there claims in the implementer's self-report not actually reflected in the diff?

Findings from this pass are treated as gate failures, not suggestions.

### Adversarial Verifier Pattern (Concrete Procedure)

The section above specifies *when* to spawn a fresh review pass before marking
implementation complete; this section specifies *how* to frame that pass as an
adversarial check rather than a neutral review that tends to rubber-stamp.

**Procedure** (spawn fresh — brief adversarially — bound inputs — cap iterations):

1. **Spawn fresh** — invoke a new agent/subagent with no prior implementation
   context. Do not carry over the implementer's conversation history,
   chain-of-thought, or self-report. Provide only the spec, the constitution,
   and the diff.
2. **Brief adversarially** — instruct the verifier to "assume this diff has at
   least one defect; find it." A neutral "please review" brief tends to
   confirmation-bias; adversarial framing does not (Anthropic, "Building
   Effective Agents": the evaluator-optimizer split outperforms same-call
   generation+evaluation, retrieved 2026-09-25).
3. **Bound inputs** — provide only: (a) the specification's acceptance
   criteria, (b) the project constitution's non-negotiables, (c) the diff.
   Re-derive claims from these sources; do not trust the implementer's
   self-report as evidence.
4. **Blocking findings** — a finding that contradicts an acceptance criterion
   or violates a constitution non-negotiable is a **gate failure**, not a
   suggestion. Do not proceed past quality gates until every blocking finding
   is resolved.
5. **Iteration cap** — allow at most one re-implement + re-verify cycle before
   escalating to `.plaesy/instructions/error-recovery.md`. Prevents infinite
   verify-fail loops where the implementer patches surface symptoms without
   addressing the root cause.
6. **One pass, not a debate loop** — the adversarial verifier performs a single
   bounded scan; it does not iterate with the implementer to hash out
   disagreements. (Per 2026-09-25 research, the evaluator-optimizer pattern
   performs best as discrete passes, not extended multi-agent debate.)

This pattern generalizes beyond `/implement` code diffs — a `/fix:business`
pricing correction or `/fix:legal` compliance amendment should also receive a
fresh adversarial re-check against the original finding before being marked
resolved (see `.plaesy/instructions/dimension-mapping.md`).

## Output Format

```text

QUALITY GATES: [PASS/BLOCKED]
1. Build:          [PASS/FAIL]
2. Tests:          [PASS/FAIL] (N passing / M failing)
3. Coverage:       [XX%] (target: YY%)
4. Security:       [PASS/FAIL] (N findings at/above bar)
5. Lint:           [PASS/FAIL]
6. Performance:    [PASS/FLAG] (measured vs target)
7. Accessibility:  [PASS/FAIL/N-A]
   Click-through:   [element -> effect, one line per element, or N/A if not UI]
8. Documentation:  [PASS/FLAG]
9. Cleanup:        [PASS/FLAG]
Verifier pass:     [PASS/FAIL] (see findings)
(first blocking failure: [gate number and name], or none — the checklist
 stops at the first blocking failure, so at most one may be reported)
```

## Deterministic Enforcement (Optional, Recommended)

The gates above are pass/fail criteria evaluated by the agent; this section
notes how to make them machine-enforced via Git pre-commit / CI hooks so they
cannot be bypassed by a careless commit or a "just make it pass" rush.

**Hookable gates** (enforce via pre-commit or CI before merge):

- **Build** (Gate 1) — project build command must exit zero.
- **Tests** (Gate 2) — full suite must exit zero.
- **Lint** (Gate 5) — linter must report no new errors (diff against baseline).
- **Security** (Gate 4) — secret-scanning hook + dependency-vulnerability scan.

**Not hookable** (remain agent-verified):

- **Coverage** (Gate 3) — requires runtime measurement against `{{TEST_COVERAGE_MIN}}`.
- **Performance** (Gate 6) — requires environment-specific baselines.
- **Accessibility** (Gate 7) — requires runtime/UI inspection.

**Recommended minimum**: any repository that wants deterministic enforcement
should wire Gates 1, 2, 4, 5 into hooks; the remaining gates (3, 6, 7) and the
Verifier Separation pass remain agent-reported per the Output Format above.

## On Failure

- Blocking failure → halt, report which gate + why, route per
  `.plaesy/instructions/error-recovery.md`.
- Non-blocking flag → continue, but surface in the phase's completion report so
  `/assess` picks it up.
