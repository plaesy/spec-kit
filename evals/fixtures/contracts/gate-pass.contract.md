# Success Contract: fix-empty-slug (gate passed)

Fixture: closed contract with every result recorded and the blocking item met.

```json
{
  "id": "fix-empty-slug",
  "phase": "fix",
  "dimension": "technical",
  "status": "closed",
  "frozen": {
    "items": [
      {"id": "C1", "metric": "failing tests in internal/validate", "target": "zero", "method": "go test ./internal/validate/", "blocking": true},
      {"id": "C2", "metric": "statement coverage of the changed function", "target": ">= 90", "method": "go test -cover ./internal/validate/", "blocking": false}
    ]
  },
  "items": [
    {"id": "C1", "kpi": "Regression fixed", "metric": "failing tests in internal/validate", "target": "zero", "method": "go test ./internal/validate/", "blocking": true, "result": "met", "actual": "0", "evidence": "cmd:go test ./internal/validate/ (run 2026-10-17)"},
    {"id": "C2", "kpi": "Changed function covered", "metric": "statement coverage of the changed function", "target": ">= 90", "method": "go test -cover ./internal/validate/", "blocking": false, "result": "met", "actual": "92.4", "evidence": "cmd:go test -cover ./internal/validate/"}
  ]
}
```
