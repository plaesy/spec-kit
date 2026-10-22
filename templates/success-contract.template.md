# Task checklist: {{TASK}}

Save this as `contracts.json` in the task folder, next to `spec.md`:
`.plaesy/specs/{{SPEC}}/tasks/{{STATUS}}/{{TASK}}/`. Run `plaesy tasks validate {{TASK}}` to record the
results. The task is done when every item is `passed`.

- `checklist`: the items of the task. Each has a `run` command (run from the project root) and an `expected`
  (what the command's last output line must reach).
- `status`: `pending` until its first pass, `passed` while its last run passed, `failed` while its last run failed.
- `histories`: every failed run, with `at`, `actual` and `run`. Failures are kept.

```json
{
  "id": "{{TASK}}",
  "checklist": [
    {"id": "C1", "check": "{{CHECK}}", "run": "{{RUN}}", "expected": "{{EXPECTED}}", "status": "pending", "histories": []}
  ]
}
```

Expected values: `pass` (exit 0), `zero`, `yes`, `present`, `absent`, `all` (last line `N of N`), or a number with
a comparison and optional unit (`>= 9`, `<= 2 s`). The rule is in `.plaesy/instructions/success-contract.md`.
