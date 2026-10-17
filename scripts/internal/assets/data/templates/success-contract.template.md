# Success Contract: {{PHASE}} ({{SLUG}})

Written before the first action of the phase, and committed with its frozen snapshot before any
work starts. The rule is `.plaesy/instructions/success-contract.md`. Keep it to 2–3 items for a
small change and 5–10 for a feature. A contract over 150 lines is too large to read at the gate.

Phase: {{PHASE}}  ·  Dimension: {{DIMENSION}}  ·  Created: {{DATE}}

## How to fill it

1. Replace every `{{...}}` and every empty string. The validator refuses empty fields, placeholders
   and vague words.
2. Each `target` is a comparison with one number and, if the actual value has one, its unit
   (`>= 90`, `<= 2 s`, `>= 80%`), or a binary condition (`zero`, `none`, `pass`, `all pages`).
   Each `method` names the command, count or procedure that produces the actual value.
3. Copy each item into `frozen.items` (id, metric, target, method, blocking) **before work starts**.
   Run `plaesy validate contract <this file>`, fix every error, then commit the file. The freeze is
   checked against that first commit.
4. At the end, fill `result` (`met`, `not_met` or `not_measurable`), `actual` or `reason`, and
   `evidence` (a repo-relative path that exists, optionally `:line`, or `cmd:` followed by the command
   that was run). Set `status` to `closed` once `plaesy validate contract <this file> --gate` passes.
   Put the same results in the phase report under "Contract comparison".
5. A target changes only by an entry in `amendments`, with a reason, and the `from` value must be the
   frozen one. Blocking never goes from `true` to `false`, and a blocking item is never removed.
   Entries are only ever appended: a committed amendment must not be edited or deleted.
6. An amendment entry has this shape (write it in the `amendments` array above, not in a second block):
   `item`, `field` (`target`, `metric`, `method`, `blocking`, `added` or `removed`), `from`, `to`, `reason`.

## Contract

```json
{
  "id": "{{SLUG}}",
  "phase": "{{PHASE}}",
  "dimension": "{{DIMENSION}}",
  "status": "open",
  "frozen": {
    "items": [
      {"id": "C1", "metric": "{{METRIC}}", "target": "{{TARGET}}", "method": "{{METHOD}}", "blocking": true}
    ]
  },
  "items": [
    {"id": "C1", "kpi": "{{KPI}}", "metric": "{{METRIC}}", "target": "{{TARGET}}", "method": "{{METHOD}}", "blocking": true}
  ],
  "amendments": [],
  "waivers": []
}
```

## Amendments and waivers

Empty until a target changes or a waiver is needed. Write each entry into the `amendments` or
`waivers` array of the block above. Example of an amendment entry (illustration only, not a block):
`{"item": "C1", "field": "target", "from": "{{OLD}}", "to": "{{NEW}}", "reason": "{{WHY}}"}`.
A waiver is `{"kind": "no-blocking", "reason": "{{WHY}}"}`, for a phase with no blocking item.

## Gate comparison

| Item | KPI | Target (frozen) | Actual | Result | Blocking | Evidence or reason |
|---|---|---|---|---|---|---|
| C1 | {{KPI}} | {{TARGET}} |  |  | yes |  |
