# Task Management System

**Plaesy Spec-Kit Task Management** — File-based task tracking integrated into project workflows.

## Overview

Tasks live in `.plaesy/tasks/` as markdown files, organized by status (backlog, todo, doing, done, blocked). They bridge project planning with automated workflow execution.

### Key Features

- ✅ **File-based**: Version-controlled, Git-friendly
- ✅ **Auto-detection**: `/continue` and `/evolve` auto-scan for tasks
- ✅ **Lifecycle managed**: Tasks move through phases automatically
- ✅ **Dependency aware**: Tasks can link to other tasks
- ✅ **Phase-aware**: Tasks tied to workflow phases (research, clarify, implement, etc.)

## Directory Structure

```
.plaesy/tasks/
├── backlog/      # Ideas, features, bugs — unscheduled
├── todo/         # Ready to start — next in queue
├── doing/        # In active work — phase executing now
├── done/         # Completed & validated — criteria met
└── blocked/      # Waiting on dependency — user must unblock
```

## Task Lifecycle

```
User Creates        Phase Executes      Quality Gates      Phase Complete
    ↓                    ↓                    ↓                  ↓
backlog ─→ todo ─→ doing ─→ [check criteria] ─→ done / blocked
 (idea)  (ready) (active)   (validation)    (success) (failure)
```

## Task File Format

### Naming Convention
```
{{yyyymmddhhmmss}}-{{topic}}.md

Example: 20260906143045-add-design-validation.md
         ↑                  ↑
      Timestamp          Topic (kebab-case, <30 chars)
```

### YAML Frontmatter
```yaml
---
Status: backlog|todo|doing|done|blocked
Created: {{CURRENT_DATE}}
Updated: {{CURRENT_DATE}}
Phase: research|clarify|design|flow|implement|assess|optimize|fix
Description: >
  One-sentence description of what this task accomplishes.
Acceptance Criteria:
  - [ ] Criterion 1
  - [ ] Criterion 2
Related: []
Priority: critical|high|medium|low
---
```

## Workflow Integration

### /start
- **Creates**: Initial `.plaesy/tasks/` directory structure
- **Seeds** (optional): Tasks in backlog if project plan provides them

### /continue
- **Detects**: Tasks in todo/ and doing/
- **Reports**: In-flight work summary
- **Alerts**: If blocked/ has entries

### /evolve
- **Pops**: Next task from backlog/ (by priority)
- **Moves**: todo/ → doing/ → done/ or blocked/
- **Manages**: Task lifecycle through execution phases

### /save
- **Scans**: All task directories
- **Counts**: Tasks per status and phase
- **Updates**: context.md with task summary

## Installation Integration

When you run `plaesy init . --ai claude_code`:

1. ✅ Creates `.plaesy/tasks/` with all status directories
2. ✅ Copies task templates from templates/tasks/
3. ✅ Adds task management to project setup
4. ✅ Ready for /start → /continue → /evolve workflows

---

**Status**: ✅ Production Ready  
**Framework**: Plaesy Spec-Kit Constitutional Development Framework
