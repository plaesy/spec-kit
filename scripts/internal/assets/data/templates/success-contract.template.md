# Success Contract: {{PHASE}} ({{SLUG}})

Written before the first action of the phase, and frozen before any work starts. The rule is
`.plaesy/instructions/success-contract.md`. Keep it to 2–3 items for a small change and 5–10 for a
feature. A contract over 150 lines is too large to read at the gate.

Phase: {{PHASE}}  ·  Dimension: {{DIMENSION}}  ·  Created: {{DATE}}

## How to fill it

1. Replace every `{{...}}` and every empty string. The validator refuses empty fields and vague words.
2. Each `target` is a number with a comparison (`>= 90`, `<= 200`) or a binary condition (`zero`,
   `none`, `all pages`, `pass`). Each `method` names the command, count or procedure that produces the
   actual value.
3. Copy each item into `frozen.items` (id, metric, target, method, blocking) **before work starts**,
   then run `plaesy validate contract <this file>`. Fix every error.
4. At the end, fill `result` (`met`, `not_met` or `not_measurable`), `actual` or `reason`, and
   `evidence` for every item. Set `status` to `closed` only after `plaesy validate contract <this file> --gate`
   passes. Put the same results in the phase report under "Contract comparison".
5. A target changes only by an entry in `amendments`, with a reason. Blocking never goes from `true` to `false`.

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
  ]
}
```

## Amendments

Empty until a target changes. Each entry:

```json
{"item": "C1", "field": "target", "from": "{{OLD}}", "to": "{{NEW}}", "reason": "{{WHY}}"}
```

## Gate comparison

| Item | KPI | Target (frozen) | Actual | Result | Blocking | Evidence or reason |
|---|---|---|---|---|---|---|
| C1 | {{KPI}} | {{TARGET}} |  |  | yes |  |
