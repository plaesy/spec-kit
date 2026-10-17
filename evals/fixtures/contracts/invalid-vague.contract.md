# Success Contract: fix-empty-slug (invalid)

Fixture: vague words and a non-measurable target. The validator must reject it.

```json
{
  "id": "fix-empty-slug",
  "phase": "fix",
  "dimension": "technical",
  "status": "open",
  "frozen": {
    "items": [
      {"id": "C1", "metric": "good code quality", "target": "good", "method": "look at it", "blocking": true}
    ]
  },
  "items": [
    {"id": "C1", "kpi": "Good UX", "metric": "good code quality", "target": "good", "method": "look at it", "blocking": true}
  ]
}
```
