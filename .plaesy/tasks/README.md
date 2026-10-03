# Task Management

Tasks are organized by status:

- **backlog/** — Ideas, features, bugs (unscheduled)
- **todo/** — Ready to start (next in queue)
- **doing/** — In active work
- **done/** — Completed and validated
- **blocked/** — Waiting on dependency

## Task Lifecycle

```
backlog → todo → doing → [quality-gates] → done / blocked
```

Move tasks between directories as status changes. Use `/continue` or `/loop` for automated workflow.

**See `.plaesy/instructions/tasks.md` for detailed task management instructions.**
