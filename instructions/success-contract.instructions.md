---
name: success-contract
description: The Success Contract rule. Before any phase starts, it records measurable KPIs and a measurable checklist, frozen in git; its final gate compares results against them item by item. Load it before starting any phase and before claiming completion.
---

# Success Contract

**Mandatory for every phase, and a constitution non-negotiable (NN-08).** Before work starts, a phase
writes a *contract*: a short list of measurable items it will be judged against, committed to git. At
the end, the phase compares each item with its result, item by item. Completion is valid only when that
comparison is recorded and no blocking item is unmet or unmeasured. This applies to every phase,
including `/assess`, `/implement`, `/optimize`, `/fix`, `/improve`, `/loop`, `/create`, `/doc`, `/spec`,
`/continue`, `/save`, `/start`, and their per-dimension variants.

The contract is created **before** the work, never after. A contract written from the results is not a
contract. The validator checks this against git history where it can; section 4 says what it cannot.

## 1. When and where

1. **Before the first action** of the phase (after any routing or persona step the prompt names),
   write the contract. The template is `.plaesy/templates/success-contract.template.md`.
2. Store it at `.plaesy/contracts/<phase>-<id>.contract.md`, where `<phase>` is the command (for example
   `implement` or `assess-technical`) and `<id>` is a short slug. Contracts are tracked in git, like
   `.plaesy/tasks/`. Keep them; do not delete them. Each contract is at most 150 lines, one file per phase.
3. Fill the `frozen` snapshot (the targets as they stand before work), validate with
   `plaesy validate contract <file>`, fix every error, and **commit the file** before the first action.
4. Only then start the work. A phase with no valid, committed, frozen contract is not started.
   `quality-gates` Gate 0 enforces this.

## 2. Contract schema

The contract is a markdown file with **exactly one** fenced `json` block. The validator reads that block;
the prose around it is for people. Examples go in prose, never in a second block.

| Field | Required | Meaning |
|---|---|---|
| `id` | yes | Slug, lower case, digits and hyphens. |
| `phase` | yes | The command that owns the contract, for example `implement`, `assess:technical`. |
| `dimension` | yes | One of `technical`, `design`, `business`, `legal`, `marketing`, `financial`, `management`, `product`, `operations`, or `general`. |
| `status` | yes | `open` while work runs; `closed` only once the gate passes. A `closed` contract is checked against the gate in every mode. |
| `frozen` | yes | `{"items": [...]}`: `id`, `metric`, `target`, `method` and `blocking` for every item, as they were before work. |
| `items` | yes | One or more items (below). At least one must be blocking, unless a waiver says otherwise. |
| `amendments` | yes (may be empty) | The only way a frozen target changes (section 4). Append-only. |
| `waivers` | yes (may be empty) | Logged exceptions. The only kind is `no-blocking`, with a reason. |

Each item:

| Field | Required | Meaning |
|---|---|---|
| `id` | yes | Unique within the contract, for example `C1`. |
| `kpi` | yes | Short name of what is measured. |
| `metric` | yes | Exactly what is counted or computed. |
| `baseline` | no | The value before the work, if known. Recorded, not judged. |
| `target` | yes | A comparison with one number and an optional unit (`>= 90`, `<= 2 s`, `>= 80%`), or a binary condition (`zero`, `none`, `pass`, `yes`, `no`, `present`, `absent`, `all pages`, `every route`). |
| `method` | yes | The command, count or procedure that produces the actual value. At least 8 characters, with a procedure word such as `test`, `count`, `measure`, `run`, `inspect`. |
| `blocking` | yes | `true` or `false`. A blocking item that is unmet or unmeasured fails the phase. |
| `evidence` | at the gate | A repo-relative path that exists (optionally `:line` or `:start-end`), or `cmd:` followed by the command that was run. |
| `result` | at the gate | `met`, `not_met` or `not_measurable`. |
| `actual` | at the gate | The measured value, with its unit when the target has one (`92.4`, `1.5 s`, `0`, `5 of 5`, `pass`). |
| `reason` | at the gate, for `not_measurable` | Why it could not be measured. |

## 3. What counts as measurable

An item is measurable only if a person or a command can decide met or not met from its `method`
alone, with no judgement about quality in general.

**Accepted.** `>= 90` coverage from `go test -cover`; `zero` Sev-High findings from the named scanner;
`all 5 pages render at 390 px with no horizontal scroll`; `pass` on `make test`.

**Rejected by the validator.** In a KPI, metric or target, these words are rejected: `good`, `nice`, `clean`, `fast`, `intuitive`, `user-friendly`, `user friendly`, `robust`, `high
quality`, `appropriate`, `as needed`, `improve`, `better`, `acceptable`, `reasonable`, `sufficient`, `tbd`, `todo`, `approx`, `approximately`, `about`, `roughly`, `around`,
`several`, `accessible`, `secure`, `adequate`, `enough`, `significant`, `usable`, `simple`, `easy`, `quick`, `modern`, `stable`, `scalable`.

Inside a code span or after a dash (a flag such as `--fast`), a word is not counted. A target must be a comparison with one number, or a binary condition; `90` alone, `about 90`,
`12 files`, and `no more than 3` are rejected.

"Good UX" is not measurable. "Task completion rate >= 80% in five moderated sessions (method: session notes)" is.

## 4. Freeze and amendments (anti-hindsight)

Results may never redefine targets. The rules:

1. **The freeze is the first committed version.** When the contract file is in git, the validator takes its
   first committed version that holds a `frozen` snapshot. The snapshot must (a) be identical to the current
   `frozen`, (b) hold no results, and (c) every amendment ever committed must still be present, unchanged,
   in the current file. Changing `frozen` to match an edit is therefore caught.
2. A change to `metric`, `target`, `method` or `blocking` after freeze is allowed only as an **amendment**:
   an `amendments` entry with `item`, `field`, `from` (the frozen value), `to` (the new value) and a non-empty
   `reason`. Write it before the work it affects continues, and commit it.
3. `blocking` may **not** change from `true` to `false` after freeze. A blocking item is **never** removed.
   Loosening a blocking item, or removing one, means stopping and asking the human, not editing the file.
   These are errors even with an amendment.
4. An item added after freeze is an amendment with `field: "added"`; a non-blocking item removed after freeze
   is `field: "removed"`. Both need a reason.
5. Amendments are append-only. Editing or deleting a committed amendment fails the anchor check.

**What is not verified, stated plainly.** The git check proves the snapshot and the amendment log match the
first committed version. It does not prove that the first commit came before the work (git records when a
commit was made, not whether the work came first). It does not prove that the contract was not replaced by a
new file under another name. Outside git, or before the contract is committed, the freeze is reported as
**NOT VERIFIED**: a warning in plain mode, an error at the gate, and an error with `--require-git`. The phase
report must say which case applied.

## 5. Gate comparison (end of the phase)

The final gate fills `result` and `actual` (or `reason`) and `evidence` for **every** item, sets `status` to
`closed` once it passes, and runs `plaesy validate contract <file> --gate`. The rules:

- `met` needs `actual` and `evidence`. The validator checks the claim against the target:
  - a numeric target: the number and the unit must match (time units convert: `ms`, `s`, `min`, `h`); a
    claim that does not satisfy the target is refused, and so is an actual that is not a readable number;
  - a binary target: `zero` or `none`; `pass`, `yes`, `true` or `present`; `absent`, `no`, `false`, `none` or
    `0`; `all X` or `every X` takes `pass` or `n of n` with n above zero. Anything else is refused.
- `evidence` must be a path that exists under the repository, optionally with `:line`, or `cmd:` and the
  command. The validator checks that the path exists. It does not read the file or re-run the command.
- `not_met` needs evidence. `not_measurable` needs a `reason`.
- A **blocking** item that is `not_met` or `not_measurable` fails the gate. The phase must not report completion.
- A non-blocking `not_met` is a warning, provided the contract has a blocking item that can fail the phase.
- A contract whose items are **all** non-blocking fails the gate when any item is `not_met`, unless a
  `no-blocking` waiver is recorded (then it is a warning).

The result is also written into the phase's report, in the "Contract comparison" table of `quality-gates`.

## 6. Proportionality (avoid ceremony)

- A tiny fix: **2 to 3 items**, for example the failing test passes, the regression is covered, the touched
  file stays under its line budget.
- A feature: 5 to 10 items.
- A whole-project assessment: one contract per dimension, 3 to 8 items each.
- A status or state phase (`/save`, `/continue`, a routing step): **one** blocking item, for example
  "state file written and validated with `plaesy validate memory`". The block and the gate still apply.
- A contract over 150 lines is an error: it is too large to read at the gate. Cut items or split the phase.
- Reuse what exists. Thresholds already in the constitution are cited, not restated: use the constitution's
  value in `target` and point to the constitution row in `evidence`.

## 7. Defaults from the constitution

The constitution (`.plaesy/memory/constitution.md`, section 1) sets binding thresholds per active dimension.
A contract for an active dimension should carry them as **blocking** items unless the human overrides the cell
in the constitution. Current defaults: technical 90% coverage, p95 under 200 ms, zero known Sev-High
vulnerabilities; design WCAG 2.2 AA; marketing every published claim sourced; legal every clause traceable;
financial every cost sourced; business unit economics reviewed; product prioritization applied; management no
single point of failure; operations as set by the constitution's operations row.

The validator does not yet require those threshold rows to be present. That needs a mapping from each
constitution cell to the item that measures it, which is a human decision (see the review notes).

## 8. Starter KPI sets

Starting points, not a mandate. Pick the items that fit the phase, set the target from the constitution or the
brief, and keep it measurable. Each line gives the target and the method in code spans; a test validates every
one of them against the validator's rules.

**technical** (`dimension: technical`)

- Line coverage of changed packages. Target `>= 90`. Method `go test -cover` on those packages.
- p95 latency of the changed endpoint. Target `<= 200 ms`. Method `load run with the named tool`.
- Known Sev-High vulnerabilities. Target `zero`. Method `run the dependency and secret scanners`.
- Tests added for the change. Target `all new behaviour covered`. Method `diff review against the acceptance criteria`.

**design** (`dimension: design`)

- Text contrast, normal text (WCAG 2.2 AA). Target `>= 4.5`. Method `contrast check on every colour pair the page uses`.
- Text contrast, large text (WCAG 2.2 AA). Target `>= 3`. Method `contrast check on large headings`.
- Keyboard-only path to each primary action. Target `pass`. Method `manual tab walk recorded in evidence`.
- Layout at 1440, 768 and 390 px. Target `pass`. Method `render each page at the three widths and inspect for clipped text`.

**business** (`dimension: business`)

- Contribution margin per unit. Target `>= 0`. Method `compute the unit-economics sheet, base case`.
- Payback period on CAC. Target `<= 12 months`. Method `calculate CAC divided by monthly contribution`.
- Assumptions logged with a source. Target `all assumptions`. Method `review the assumption register`.

**legal** (`dimension: legal`)

- Clauses traceable to a source. Target `all clauses`. Method `review the clause-to-source table`.
- Compliance owner named. Target `present`. Method `review the policy or contract header`.
- Open blocking compliance findings. Target `zero`. Method `run the compliance review checklist`.

**marketing** (`dimension: marketing`)

- Published claims carrying a source. Target `all published claims`. Method `review the claim inventory against the source list`.
- Brand voice rules applied. Target `all pieces`. Method `review each piece against the voice checklist`.

**financial** (`dimension: financial`)

- Every cost and price input sourced. Target `all cost inputs`. Method `review the input register`.
- Unit economics under the pessimistic case. Target `>= 0`. Method `read the lowest scenario in the sensitivity table`.

**management** (`dimension: management`)

- Roles named for every deliverable. Target `all deliverables`. Method `review the RACI table`.
- Single points of failure. Target `zero`. Method `review the team and ownership map`.

**product** (`dimension: product`)

- Roadmap items with a recorded priority score. Target `all roadmap items`. Method `review the roadmap table`.
- Competitive baseline sourced. Target `>= 3`. Method `count the sources in the source list`.

**operations** (`dimension: operations`)

- Deployment is repeatable. Target `pass`. Method `run the named deploy command in a clean environment`.
- Rollback tested. Target `pass`. Method `run the named rollback command`.
- Alerting on the stated SLO. Target `present`. Method `review the alert definition`.

## 9. Worked example

A small fix to a validator bug. Two items, both measurable, one blocking. Committed with this frozen snapshot
before the first action, then closed by the gate with results.

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
  ],
  "amendments": [],
  "waivers": []
}
```

The validator accepts this file as open. It passes the gate once both results are recorded, with evidence that
exists and actual values that satisfy the targets. The full template is `templates/success-contract.template.md`.

## 10. Relation to other rules

- `quality-gates` Gate 0 refuses a phase with no valid, committed, frozen contract, and the contract comparison is
  part of the gate output. Completion claims without it are invalid.
- Constitution NN-08 makes this rule non-negotiable: no phase reports completion while a blocking item is unmet
  or unmeasured.
- `tasks` acceptance criteria are the story-level form of the same idea. A story's criteria may be copied into a
  contract as items; they are not a substitute for one.
- The constitution sets the thresholds. This file sets how a phase records them. A phase may never weaken a
  constitution threshold; it may only record a measured value against it.
