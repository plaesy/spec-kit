---
name: tasks
description: >-
  Create, start, complete, block, unblock, and check Plaesy tasks. Each task is a
  folder under .plaesy/tasks/<status>/ with a task.md and a contracts.json, and
  belongs to a topic whose spec and plan live under .plaesy/spec/ and
  .plaesy/plan/. Use when the user asks to create a task, start or finish work on
  one, block or unblock it, check whether it is valid, or see what is in progress.
---

# tasks

A task is one unit of work. Its folder moves between statuses as the work moves.
The spec and plan are not in the task folder; they belong to the task's topic.

## Layout

```text
.plaesy/tasks/<status>/<priority>_<topic>/
  task.md          status, topic, title
  contracts.json   the checklist that says when the task is done

.plaesy/spec/<topic>/spec.md   requirements (written by the spec skill)
.plaesy/plan/<topic>/plan.md   how the requirements are met (written by the plan skill)
```

Statuses, in order: `backlog`, `todo`, `doing`, `done`, `blocked`.

## Rules

- Every task has a topic. A task with no topic is invalid and cannot move.
- Create a task with the CLI, never by hand:
  `plaesy tasks create <priority> <category> <title>`. The topic is
  `<category>-<title>` in kebab-case, and the folder is `<priority>_<topic>`.
- Move a task with the CLI, never with `mv`. A plain move skips the `status:` line
  in `task.md` and the blocked reason.
- Stages are not skipped. Backlog goes to todo, todo to doing, doing to done.
  Use `plaesy tasks move` only for an ad-hoc correction, and say why.
- A task moves to `done` only when every item in `contracts.json` has passed.
- A task is not started until its topic plan has status `approved`.

## Commands

| Action | Command | Needs |
| --- | --- | --- |
| Create | `plaesy tasks create <priority> <category> <title>` | priority: critical, high, medium, or low |
| Next task | `plaesy tasks next` | a task in backlog |
| Start | `plaesy tasks start <task>` | todo; topic plan approved |
| Complete | `plaesy tasks complete <task>` | doing; every contract item passed |
| Block | `plaesy tasks block <task> [reason]` | doing; appends `## Blocked Reason` to `task.md` |
| Unblock | `plaesy tasks unblock <task>` | blocked; the dependency is resolved |
| Check | `plaesy tasks check <task>` | the topic's spec and plan exist and their links resolve |
| List | `plaesy tasks list` | none |

## Workflow

1. **Find the task.** Run `plaesy tasks list`. If the task exists, continue it. If not,
   create it.
2. **Make sure its topic has a spec and a plan.** If not, use the `spec` skill, then
   the `plan` skill. Do not write them in the task folder.
3. **Check the task.** Run `plaesy tasks check <task>`. Fix every missing reference
   before moving on.
4. **Move it forward** with the CLI as the work moves. Fill `contracts.json` with
   one item per requirement. Each item needs a command that proves it, not a
   description.
5. **Block or unblock** when a dependency stops or resumes. Write the reason in
   `task.md`. Review a blocked task within 14 days.
6. **Complete** the task only when `plaesy tasks complete` accepts it. Report the
   result from the CLI output.

## Output

Report the task name, the status before and after, and the command output. If a
move was refused, say which contract item or precondition caused it.
