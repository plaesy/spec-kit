# Success Contract: fix-empty-slug

Fixture: a valid, frozen, open contract. Used by `plaesy validate contract` tests.

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
