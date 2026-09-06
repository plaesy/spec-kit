---
description: "Task management and workflow organization"
applyTo: ".plaesy/tasks/**"
---

# Task Management Instructions

## Overview

Tasks in Plaesy are organized by status in the `.plaesy/tasks/` directory structure. Each task moves through lifecycle stages from conception to completion.

## Task Organization

### Directory Structure

```
.plaesy/tasks/
├── backlog/       → Ideas, features, bugs (unscheduled)
├── todo/          → Ready to start (next in queue)
├── doing/         → In active work
├── done/          → Completed and validated
└── blocked/       → Waiting on dependency
```

### Task Lifecycle

```
backlog → todo → doing → [quality-gates] → done / blocked
```

**Flow:**
1. **Backlog** — Initial ideas, feature requests, bug reports (not scheduled)
2. **Todo** — Prioritized and ready to start (next in queue)
3. **Doing** — Currently being worked on
4. **Done** — Completed and validated (passes quality gates)
5. **Blocked** — Waiting on external dependency or blocker resolution

## Task File Format

### Naming Convention
- Format: `{priority}_{title}.md`
- Example: `critical_setup-auth-middleware.md`

### Minimal Task File
```markdown
---
title: [title]
phase: [research|clarify|design|flow|implement|assess|optimize|fix]
status: [backlog|todo|doing|done|blocked] 
createdAt: ["2006-01-02T15:04:05.999Z"]
updatedAt: ["2006-01-02T15:04:05.999Z"]
---

## Description
Clear description of what needs to be done.

## Decision
- ... list decisions

## Acceptance Criteria
- [ ] Criterion 1
- [ ] Criterion 2
- [ ] Criterion 3

## Reference
- [topics](memory/{topics})
- ... list `.plaesy/memory/{topics}.md` for reference

## Dependencies
- path file another tasks (if blocked by another task, ex: `.plaesy/tasks/{status}/*.md`)
- External dependency (if any)

## Notes
Implementation notes, references, or constraints.

## Next
Next todo list for this tasks
```

## Task Management Rules

### Moving Tasks
- Move task file between status directories as work progresses
- Update file timestamp on move (git will track this)
- Don't delete tasks — move to appropriate directory status (`.plaesy/tasks/{status}/{topcis}.md`)

### Priority Guidelines
1. **CRITICAL** — Blocks other work or required for MVP
2. **HIGH** — Important but not blocking
3. **MEDIUM** — Nice-to-have, improves quality
4. **LOW** — Polish, optimization, tech debt
5. **BACKLOG** — Future consideration, no timeline

### Parallel Execution
- Mark parallel-safe tasks with `[P]` prefix if possible to run in parallel
- Same file/component = sequential (prevent conflicts)
- Different files = usually parallel-safe

## Autonomous Execution

Use `/continue` with task management for autonomous workflow:

**Single iteration**:
```bash
/continue
```

**Autonomous loop** (via `/schedule` or `/loop`):
```bash
/loop /continue
```

This workflow:
1. Reads tasks from `.plaesy/tasks/todo/` directory (if any exist)
2. Auto-detects project state and executes remaining phases
3. Completes current task and moves to `.plaesy/tasks/done/*.md`
4. Blocked items move to `.plaesy/tasks/blocked/*.md` with reason
5. Reports progress to `.plaesy/context.md`
6. Can loop autonomously via `/loop` or `/schedule` for repeated execution

## Best Practices

### Task Decomposition
- Break large features into 1-4 hour tasks
- One task = one clear outcome
- Include acceptance criteria (not just description)

### Documentation
- Link to relevant spec documents in task
- Reference related tasks (dependencies, blocked-by)
- Keep notes of blockers or decisions made during execution

### Quality Gates
Tasks in `doing` → `done` must pass:
- Code review (if applicable)
- Tests passing (if applicable)
- Documentation complete (if applicable)
- Acceptance criteria met
- Constitutional compliance (per `/start` principles)

---

**Related**: [[shared-protocols.instructions.md]], [[flow.instructions.md]]
