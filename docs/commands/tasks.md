# `tasks`

```bash

plaesy tasks [subcommand]
```

Task lifecycle management for `.plaesy/tasks/{backlog,todo,doing,done,blocked}/`.
Every task of the project lives here, in folders. Each task records the spec it belongs to in its `spec.md`.
The spec is the current git branch, which is also its name. `--spec <name>` selects another.

## Subcommands

| Subcommand | Args | Description |
|---|---|---|
| `create` | `<priority> <slug> <title>` | Create a task in the backlog, for the current spec (created first if missing) |
| `list` | `[--all]`, `[--sync-context]` | List tasks across every status; `--all` shows every spec with integrity problems; `--sync-context` rewrites the In-Flight section of `.plaesy/context.md` |
| `list-status` | `<status>` | List tasks in one status |
| `next` | — | Move the next backlog task into `todo` and report it |
| `start` | `<task>` | Move task to `doing` |
| `complete` | `<task>` | Move task to `done` |
| `block` | `<task>` | Move task to `blocked` |
| `unblock` | `<task>` | Unblock a task |
| `move` | `<task_file> <from> <to>` | Move between arbitrary statuses |
| `show` | `<task_file> [status]` | Display task content |
| `run` | `--chain <name>` or `--provider <name>`, `[--max N]`, `[--timeout D]`, `[--confirm]` | Claim todo tasks of the current spec (highest priority first) and send each to a delegate provider or chain, in the project root. Output goes to `<spec>/runs/`; the task stays in `doing` until reviewed, and a failed run moves it to `blocked` |
| `validate` | `[task]`, `[--spec <folder>]` | Run the checklist of each task (`contracts.json`), record each result, and move a finished task to `done` |

Valid statuses: `backlog`, `todo`, `doing`, `done`, `blocked`. `blocked` is only for a task waiting on something outside the work.

It was `plaesy task-manage`, which spelled the resource as `task` and then
repeated the idea in `-manage`: `tasks list` says everything `task-manage list`
does, one word shorter. The old spelling was removed, not aliased.

See [CLI Reference](../reference.md).
