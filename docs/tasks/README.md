# Task Management System

**Plaesy Spec-Kit Task Management** — File-based task tracking integrated into project workflows.

## Overview

Each spec owns its tasks. Every task is a folder under `.plaesy/tasks/<status>/`, with `spec.md` (the description) and `contracts.json` (the checklist). The folders are organized by status (backlog, todo, doing, done, blocked).

A spec is one feature on one branch, so its tasks travel with that branch. They bridge project planning with automated workflow execution.

### Key Features

- ✅ **Folder-based**: Version-controlled, Git-friendly
- ✅ **Auto-detection**: `/continue` auto-scans for tasks
- ✅ **Lifecycle managed**: Tasks move through phases automatically
- ✅ **Dependency aware**: Tasks can link to other tasks
- ✅ **Phase-aware**: Tasks tied to workflow phases (design, implement, assess, etc.)

## Directory Structure

```text

.plaesy/tasks/
├── backlog/      # Ideas, features, bugs — unscheduled
├── todo/         # Ready to start — next in queue
├── doing/        # In active work — phase executing now
├── done/         # Completed & validated — criteria met
└── blocked/      # Waiting on dependency — user must unblock
```

## Task Lifecycle

```text

User Creates        Phase Executes      Quality Gates      Phase Complete
    ↓                    ↓                    ↓                  ↓
backlog ─→ todo ─→ doing ─→ [check criteria] ─→ done / blocked
 (idea)  (ready) (active)   (validation)    (success) (failure)
```

Moves use the CLI: `plaesy tasks next` (backlog → todo), `plaesy tasks start` (todo → doing), `plaesy tasks complete` (doing → done), `plaesy tasks block` and `plaesy tasks unblock`. Each move also rewrites the `status:` line in `spec.md`.

## Task Folder Format

### Naming Convention

```text

<priority>_<category>-<title>/

Example: high_add-design-validation/
         ↑    ↑
      Priority  Slug (kebab-case)

```

Create one with `plaesy tasks create <priority> <slug> "<title>"`. It lands in `backlog/` with `spec.md` and an empty `contracts.json`.

### spec.md Frontmatter

```yaml

---
spec: [spec name, the git branch]
title: [title]
phase: design|implement|assess|optimize|fix
status: backlog|todo|doing|done|blocked
---
```

## Workflow Integration

### /start

- **Seeds** (optional): Tasks in backlog if project plan provides them

### /continue

- **Detects**: Tasks in todo/ and doing/
- **Reports**: In-flight work summary
- **Alerts**: If blocked/ has entries
- **Reads**: Task folders in todo/ and doing/ (it reports them, it does not move backlog tasks)
- **Promotion**: `plaesy tasks next` moves the next backlog task (by priority) into todo/; `plaesy tasks run` does the same when todo/ is empty
- **Moves**: todo/ → doing/ → done/ or blocked/
- **Manages**: Task lifecycle through execution phases, can loop autonomously

### /save

- **Scans**: All task directories
- **Counts**: Tasks per status and phase
- **Updates**: context.md with task summary

## Installation Integration

When you run `plaesy init . --ai claude`:

1. ✅ Copies the templates from `templates/`
2. ✅ Adds task management to project setup
3. ✅ Ready for /start → /continue workflows

---

**Status**: ✅ Production Ready
**Framework**: Plaesy Spec-Kit Constitutional Development Framework
