---
title: "Decision: Full-autonomy policy + role-handoff chaining + per-role memory"
description: "Rewrote the top-priority ask-when-unsure rule into an enumerated hard-stop list, added role-to-role auto-chaining with no hop limit, and per-role memory"
updatedAt: "2026-10-02T01:00:00Z"
decided: "2026-10-02"
status: "Accepted"
---

# Decision: Full-autonomy policy + role-handoff chaining + per-role memory

**Status**: Accepted
**Decided**: 2026-10-02
**Deciders**: project owner + session agent (pressure-tested via LLM Council + web research before finalizing)
**Related**: decisions/rebuilt-memory-index.md

## Context and Problem Statement

`/assess:management` scored the project 49/100 (capped by a failed
sustainability audit — key-person risk, no cross-role handoff). The owner
wants Plaesy Spec-Kit to run like a self-running "virtual company" of AI
employees with no human monitoring. Two gaps blocked this: no automatic
role-to-role handoff (23 roles each loaded ad hoc, never chained), and no
persistent per-role memory (every role invocation stateless). A third,
higher-leverage gap was found during research: `plaesy.instructions.md`'s
top-priority rule #2 ("Ask when unsure — one clarifying question, don't
guess") directly contradicted the autonomy goal and sat above every other
routing mechanism in the framework.

## Decision Drivers

- Fixing role/memory mechanics while rule #2 stood would still produce an
  agent that stops and waits — the policy had to change first.
- An LLM Council pressure-test (5 independent advisors + peer review) and
  2026 production-agent research both converged: a single "secrets only"
  hard-stop is too narrow, and asking the agent to self-judge "is this safe to
  assume" is exactly the judgment oversight exists to catch.
- The project owner explicitly rejected a hop-limit on role chaining —
  unlimited autonomous chaining is the stated design goal, not a side effect
  to cap.

## Decision Outcome

**Chosen option**: Replace rule #2 with "default and record, don't park" +
an enumerated, non-negotiable hard-stop list in rule #8 (secrets,
destructive/irreversible actions, externally-visible actions,
money/billing/legal commitments). Added `.plaesy/instructions/role-mapping.md`
(role-to-role routing table, mirrors `dimension-mapping.md`'s shape) and wired
`/continue`, `/implement`, `/fix`, `/optimize`, `/loop` to auto-chain through
it via `state.json.active_role`/`role_handoff_log` — **no hop limit**. Added
`.plaesy/memory/roles/[role].md` as the one documented exception to memory's
"flat, no subfolders" rule, plus a shared `## Memory Protocol` section across
all 23 role files.

**Rationale**: An enumerated hard-stop list removes the self-judgment the
council flagged as the core risk, without reintroducing a human checkpoint —
it changes *what* counts as safe to assume, never *how long* an autonomous
chain may run. The `/loop` legal-escalation worked example was also
reconciled: an indemnity clause is a legal commitment, so it correctly hits
the rule-8 hard stop — this is now consistent with the policy, not an
exception to it.

**Rejected alternative**: A hop-limit/circuit-breaker on role chaining (the
council's own suggestion) — explicitly overridden by the project owner in
favor of unlimited chaining.

## Constitutional Compliance & Quality Gates

- [x] `go build/vet/test` — green on the policy/role-mapping/memory changes;
  full suite re-run after adding the validator below
- [x] `plaesy validate markdown` — 0 violations (279 files scanned)
- [x] `plaesy validate memory` — self-contained, 0 external refs
- [x] `plaesy validate assumptions` — new validator (see below), 5 unit
  tests passing (`internal/validate/assumptions_test.go`)

## Follow-up — now implemented

`plaesy validate assumptions` (`internal/validate/assumptions.go` +
`cmd/plaesy/validate.go`): scans the corpus for `ASSUMED —` tags and lists
them. Informational-only by default (`assumptions_review_days: 0` in
`.plaesy/state.json`, same opt-in convention as `checkpoint_interval`); set it
positive (or pass `--review-days`) to gate on a file's last-commit age via
`git log -1 --format=%ct`. Not wired into bare `plaesy validate` — explicit
invocation only, consistent with "opt-in, never a surprise gate."

## References

- `.plaesy/instructions/plaesy.instructions.md` (+ mirror) — rules #2, #3, #8
- `.plaesy/instructions/role-mapping.md` (new)
- `.plaesy/templates/state.template.json`, `.plaesy/state.json` —
  `active_role`, `role_handoff_log`, `assumptions_review_days`
- `scripts/internal/validate/assumptions.go`,
  `scripts/cmd/plaesy/validate.go` — the new validator
- `prompts/continue.md`, `prompts/implement.md`, `prompts/fix.md`,
  `prompts/optimize.md`, `prompts/loop.md`
- `agents/*.agents.md` + `.plaesy/roles/*.md` — `## Memory Protocol` section
- `.plaesy/memory.md` — Role Memory section

---

**Document Control**: Full-autonomy policy + role-handoff chaining + per-role memory · v1.0 · Created 2026-10-02

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 1.0 | 2026-10-02 | session agent | Initial version |
