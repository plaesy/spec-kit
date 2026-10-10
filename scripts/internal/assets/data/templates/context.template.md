---
title: "Session Context"
updatedAt: "{{UPDATED_AT|<ISO-8601 timestamp, e.g. 2026-01-31T09:00:00Z>}}"
# phase: exactly one of start | assess | implement | optimize | fix | doc | save
phase: "{{PHASE}}"
# status: exactly one of backlog | todo | doing | done | blocked
status: "{{STATUS}}"
---

## Context

- ... list current context

## Checklists

- [x] ... current checklists

## Doing

- [{{TASKS_TOPICS}}](tasks/{{STATUS}}/{{TASKS_TOPICS}}.md)
- ... other tasks

## Next

- ... Next todo list
