---
name: success-contract
description: The Success Contract rule. Before any phase starts, it records measurable KPIs and a measurable checklist; its final gate compares results against them item by item. Load it before starting any phase and before claiming completion.
---

# Success Contract

**Mandatory for every phase.** Before work starts, a phase writes a *contract*: a short list of
measurable items it will be judged against. At the end, the phase compares each item with the
result, item by item. Completion is valid only when that comparison is recorded and no blocking
item is unmet or unmeasured. This applies to every phase, including `/assess`, `/implement`,
`/optimize`, `/fix`, `/improve`, `/loop`, `/create`, `/doc`, `/spec`, `/continue`, `/save`,
`/start`, and their per-dimension variants.

The contract is created **before** the work, never after. A contract written from the results
is not a contract.

## 1. When and where

1. **Before the first action** of the phase, write the contract file (the template is
   `.plaesy/templates/success-contract.template.md`).
2. Store it at `.plaesy/contracts/<phase>-<id>.contract.md`, where `<phase>` is the command
   (for example `implement` or `assess-technical`) and `<id>` is a short slug. Contracts are
   session records, like `.plaesy/tasks/`: keep them, do not delete them.
3. Validate it: `plaesy validate contract <file>`. Fix every error before doing the work.
4. **Freeze it.** Fill the `frozen` snapshot in the same file (the validator checks it against the
   items). From then on the targets are fixed (section 4).

A phase with no valid, frozen contract is not started. `quality-gates` Gate 0 enforces this.

## 2. Contract schema

The contract is a markdown file with one fenced `json` block. The block is what the validator
reads; the prose around it is for people.

| Field | Required | Meaning |
|---|---|---|
| `id` | yes | Slug, lower case, digits and hyphens. Unique per phase. |
| `phase` | yes | The command that owns the contract, for example `implement`, `assess:technical`. |
| `dimension` | yes | One of `technical`, `design`, `business`, `legal`, `marketing`, `financial`, `management`, `product`, `operations`, or `general`. |
| `status` | yes | `open` while work runs; `closed` only when the gate has recorded every result. |
| `frozen` | yes once frozen | Snapshot of `id`, `metric`, `target`, `method` and `blocking` for every item, taken before work. |
| `items` | yes | One or more items (below). |
| `amendments` | when targets change | Each change to a frozen target, with a reason (section 4). |

Each item:

| Field | Required | Meaning |
|---|---|---|
| `id` | yes | Unique within the contract, for example `C1`. |
| `kpi` | yes | Short name of what is measured. |
| `metric` | yes | Exactly what is counted or computed. |
| `baseline` | no | The value before the work, if known. Recorded, not judged. |
| `target` | yes | A threshold: a number with a comparison (`>= 90`, `<= 200`), or a binary condition (`zero`, `none`, `all pages`, `pass`). |
| `method` | yes | The command, check or procedure that produces the actual value. |
| `blocking` | yes | `true` or `false`. A blocking item that is unmet or unmeasured fails the phase. |
| `evidence` | at the gate | Where the proof lives: a file, a report line, a command output. |
| `result` | at the gate | `met`, `not_met` or `not_measurable`. |
| `actual` | at the gate | The measured value, for example `92.4`. Required for `met`. |
| `reason` | at the gate, for `not_measurable` | Why it could not be measured. |

## 3. What counts as measurable

An item is measurable only if a person or a command can decide met or not met from its
`method` alone, with no judgement about quality in general.

**Accepted.** `>= 90` coverage from `go test -cover`; `zero` Sev-High findings from the named
scanner; `all 5 pages render at 390 px with no horizontal scroll`; `pass` on `make test`.

**Rejected by the validator.** Targets or metrics that use vague words: `good`, `nice`,
`clean`, `fast`, `intuitive`, `user-friendly`, `robust`, `high quality`, `appropriate`,
`as needed`, `improve`, `better`, `acceptable`, `reasonable`, `sufficient`, `TBD`, `TODO`. A
numeric or binary target is required; a `method` that is one vague word is rejected. "Good UX"
is not measurable. "Task completion rate >= 80% in five moderated sessions (method: session notes)"
is.

## 4. Freeze and amendments (anti-hindsight)

Results may never redefine targets. The rules:

1. At freeze, `frozen.items` holds the targets as they were.
2. A change to `metric`, `target`, `method` or `blocking` after freeze is allowed only as an
   **amendment**: a new entry in `amendments` with `item`, `field`, `from`, `to` and a non-empty
   `reason`, written before the work it affects continues.
3. `blocking` may **not** change from `true` to `false` after freeze. Loosening a blocking item
   means stopping and asking the human, not editing the file.
4. An item added after freeze is an amendment with `field: "added"`; a removed item is
   `field: "removed"`. Both need a reason.
5. The validator reports every change it cannot match to an amendment. A phase that sees such a
   report stops and records the amendment before continuing.

## 5. Gate comparison (end of the phase)

The final gate fills `result` and `actual` (or `reason`) for **every** item, then runs
`plaesy validate contract <file> --gate`. The rules:

- `met` needs `actual` and `evidence`. For a numeric target the validator checks the claim: if
  the target is `>= 90` and the actual is `85`, `met` is refused.
- `not_met` and `not_measurable` need `evidence` or `reason`.
- A **blocking** item that is `not_met` or `not_measurable` fails the gate. The phase must not
  report completion. It reports the failure, with the item, and stops.
- A non-blocking `not_met` is reported as a warning and does not fail the gate.
- The status becomes `closed` only when the gate passes.

The result is also written into the phase's report using the table in `quality-gates`
("Contract comparison").

## 6. Proportionality (avoid ceremony)

- A tiny fix: **2 to 3 items**, for example the failing test passes, the regression is covered,
  the touched file stays under its line budget.
- A feature: 5 to 10 items.
- A whole-project assessment: one contract per dimension, 3 to 8 items each.
- A contract over 150 lines is too large to read at the gate. The validator warns at 150 lines.
- Reuse what exists. Thresholds already in the constitution are cited, not restated: use the
  constitution's value in `target` and point to the constitution row in `evidence`.

## 7. Defaults from the constitution

The constitution (`.plaesy/memory/constitution.md`, section 1) sets binding thresholds per active
dimension. A contract for an active dimension must include them unless the human overrides the
cell in the constitution. Current defaults: technical 90% coverage, p95 under 200 ms, zero known
Sev-High vulnerabilities; design WCAG 2.2 AA; marketing every published claim sourced; legal
every clause traceable; financial every cost sourced; business unit economics reviewed; product
prioritization applied; management no single point of failure; operations as set by the
constitution's operations row.

## 8. Starter KPI sets

Starting points, not a mandate. Pick the items that fit the phase, set the target from the
constitution or the brief, and keep it measurable.

**technical** (`dimension: technical`)

- Line coverage of changed packages, `>= 90`, method: `go test -cover` on those packages.
- p95 latency of the changed endpoint, `< 200 ms`, method: load run with the named tool.
- Known Sev-High vulnerabilities, `zero`, method: the project's dependency and secret scanners.
- Tests added for the change, `all new behaviour covered`, method: diff review against the story's acceptance criteria.

**design** (`dimension: design`)

- Text contrast, `>= 4.5:1` normal text and `>= 3:1` large text (WCAG 2.2 AA), method: contrast check on every colour pair the page uses.
- Keyboard-only path to each primary action, `all pass`, method: manual tab walk recorded in evidence.
- Layout at 1440, 768 and 390 px, `no horizontal scroll and no clipped text`, method: renders at all three widths.

**business** (`dimension: business`)

- Contribution margin per unit, `>= target set in the plan`, method: the model's unit-economics sheet.
- Payback period, `<= months in the plan`, method: CAC divided by monthly contribution.
- Assumptions logged with a source, `all`, method: assumption register review.

**legal** (`dimension: legal`)

- Clauses traceable to a source, `all`, method: clause-to-source table.
- Compliance owner named, `yes`, method: policy or contract header.
- Open compliance findings at the gate, `zero blocking`, method: the compliance review checklist.

**marketing** (`dimension: marketing`)

- Published claims carrying a source, `100%`, method: claim inventory against source list.
- Brand voice rules applied, `all pieces reviewed`, method: review checklist per piece.

**financial** (`dimension: financial`)

- Every cost and price input sourced, `all`, method: input register.
- Unit economics under the pessimistic case, `>= 0`, method: the sensitivity table's lowest scenario.

**management** (`dimension: management`)

- Roles named for every deliverable, `all`, method: RACI review.
- Single points of failure, `zero blocking`, method: team and ownership map review.

**product** (`dimension: product`)

- Roadmap items with a recorded priority score, `all`, method: roadmap table.
- Competitive baseline sourced, `>= 3 sources`, method: the source list.

**operations** (`dimension: operations`)

- Deployment is repeatable, `pass`, method: the named deploy command run in a clean environment.
- Rollback tested, `pass`, method: the named rollback command.
- Alerting on the stated SLO, `present`, method: alert definition review.

## 9. Worked example

A small fix to a validator bug. Two items, both measurable, one blocking:

```json
{
  "id": "fix-empty-slug",
  "phase": "fix",
  "dimension": "technical",
  "status": "open",
  "frozen": {
    "items": [
      {"id": "C1", "metric": "failing tests in internal/validate", "target": "zero", "method": "go test ./internal/validate/", "blocking": true},
      {"id": "C2", "metric": "statement coverage of the changed function", "target": ">= 90", "method": "go test -cover ./internal/validate/", "blocking": false}
    ]
  },
  "items": [
    {"id": "C1", "kpi": "Regression fixed", "metric": "failing tests in internal/validate", "target": "zero", "method": "go test ./internal/validate/", "blocking": true},
    {"id": "C2", "kpi": "Changed function covered", "metric": "statement coverage of the changed function", "target": ">= 90", "method": "go test -cover ./internal/validate/", "blocking": false}
  ]
}
```

The validator accepts this file, and it passes the gate once both results are recorded with
an actual value that satisfies the target. The full template is
`templates/success-contract.template.md`.

## 10. Relation to other rules

- `quality-gates` Gate 0 refuses a phase with no valid, frozen contract, and the contract
  comparison is part of the gate output. Completion claims without it are invalid.
- `tasks` acceptance criteria are the story-level form of the same idea. A story's criteria may
  be copied into a contract as items; they are not a substitute for one.
- The constitution sets the thresholds. This file sets how a phase records them. A phase may
  never weaken a constitution threshold; it may only record a measured value against it.
