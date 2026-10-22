---
name: success-contract
description: The task checklist rule. Every task keeps a contracts.json checklist in its own folder. plaesy tasks validate runs each item and moves the task to done when every item is passed. Load it before starting any task and before claiming completion.
---

# Success Contract

Every task has a checklist: the conditions that say the work is done. The checklist is `contracts.json`
in the task's folder, and `plaesy tasks validate` runs it.

## 1. Where it lives

```text
.plaesy/tasks/<status>/<task>/
├── spec.md          what the task is for
└── contracts.json   the checklist
```

The checklist sits in the task folder, next to the task. `<status>` is the folder the task sits in:
`backlog`, `todo`, `doing`, `done`, or `blocked`. `blocked` is only for a task waiting on something
outside the work (another task, an external service); an unfinished task stays in `doing`.

## 2. The file

```json
{
  "id": "implement-login-form",
  "checklist": [
    {
      "id": "C1",
      "check": "login tests pass",
      "run": "cd app && go test ./login -count=1",
      "expected": "pass",
      "status": "failed",
      "histories": [
        {"at": "2026-10-22T10:15:00Z", "actual": "exit code 1", "run": "cd app && go test ./login -count=1"}
      ]
    }
  ]
}
```

- `checklist`: the items of the task, in order.
- Each item has an `id`, a `check` (what is proven), a `run` (a shell command, run from the project root),
  and an `expected` (what the command's last output line must reach).
- `status` is one of:
  - `pending`: never run to a pass yet.
  - `passed`: the last run passed.
  - `failed`: the last run failed. The reason is in `histories`.
- `histories`: every failed run, with `at` (time), `actual` (why it failed), and `run` (the command that ran).
  A pass does not erase earlier failures.

## 3. Expected values

| Expected | The run passes when |
|---|---|
| `pass` | the command exits 0 |
| `zero`, `none` | the output's last line is `0` or `none` |
| `yes`, `present`, `absent` | the last line is the matching word |
| `all`, `every` | the last line is `N of N` |
| `>= 9`, `<= 2 s`, `>= 80%` | the last line is a number that meets the comparison, with the same unit |

A run must print its measurement as its last line, for example `wc -l < file`, not `wc -l file`.

## 4. Running it

```bash
plaesy tasks validate                         # every task with a contracts.json
plaesy tasks validate implement-login-form    # one task
```

Each item's `run` is executed. A passing run sets the item to `passed`. A failing run sets it to `failed`
and adds a history entry. When every item is `passed`, the task moves to `done`. A task that is not finished
is never moved to `done`.

## 5. Rules

1. Write the checklist before the work, so the expected values are set before the result is known. Change an expected value
   only by writing down why, in the task's `spec.md`.
2. Do not report a task complete while any item is not `passed`. This is constitution NN-08.
3. An item is passed only by its `run`. Do not set `status` by hand.
4. Do not delete `histories`. They are the record of what failed and when.
