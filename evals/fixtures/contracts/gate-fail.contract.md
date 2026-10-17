# Success Contract: fix-empty-slug (gate failed)

Fixture: a blocking item not met. The gate must fail, and the phase must not report completion.

```json
{
  "id": "fix-empty-slug",
  "phase": "fix",
  "dimension": "technical",
  "status": "open",
  "frozen": {
    "items": [
      {"id": "C1", "metric": "failing tests in internal/validate", "target": "zero", "method": "go test ./internal/validate/", "blocking": true}
    ]
  },
  "items": [
    {"id": "C1", "kpi": "Regression fixed", "metric": "failing tests in internal/validate", "target": "zero", "method": "go test ./internal/validate/", "blocking": true, "result": "not_met", "actual": "2", "evidence": "cmd:go test ./internal/validate/"}
  ]
}
```
