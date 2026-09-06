---
description: "Intelligent state detection and continuation orchestrator"
---

# `/continue` command instructions

⚡ **Run with**: `ultracode` (multi-agent parallel execution)

## Objective

Auto-detect current project state → Execute remaining phases → Complete

## Protocol

**Load memory state → Detect current phase → Execute remaining workflow**

## Validation Checklist (Before Running)

- ✅ Memory files exist (.plaesy/context.md, .plaesy/memory.md)
- ✅ Previous session state loadable
- ✅ Project directory accessible
- ✅ Git repo in valid state

## Memory Integration

**Follow**: `.plaesy/memory/plaesy.md` → Context Handling (authoritative)

1. **Load `.plaesy/context.md`** - Previous session state
2. **Load `.plaesy/memory.md`** - Project knowledge index
3. **Restore Progress** - Continue from checkpoint

### Load Task State

**Follow**: `.plaesy/memory/shared-protocols.md` → Task Management Protocol

1. **Scan `.plaesy/tasks/` directories**:
   - Count tasks in: `backlog/`, `todo/`, `doing/`, `done/`, `blocked/`
   - Parse each task file's YAML frontmatter for: Status, Phase, Acceptance Criteria, Related dependencies

2. **Extract phase and status information**:
   - Load current active phase from `.plaesy/context.md`
   - Filter tasks by current phase: report all `todo/` and `doing/` tasks matching active phase
   - Identify incomplete tasks: any task in `doing/` with unmet acceptance criteria (unchecked boxes)

3. **Alert on blocked tasks**:
   - ⚠️ If `.plaesy/tasks/blocked/` has entries: Surface blocker reason and blocking task references
   - Example: "2 tasks blocked (waiting on: implementation phase, external API)"
   - User must manually unblock by resolving dependencies or removing blockers

4. **Load task summary into context**:
   - Add "## In-Flight Tasks" section to `.plaesy/context.md`
   - Keep total size of context.md ≤100 lines; if task summary exceeds capacity, extract to `.plaesy/memory/tasks-summary.md`

## State Detection Logic

### Detection Process

1. **Load task state**: Scan `.plaesy/tasks/` for active work
2. **Scan `specs/` Directory**: Identify active project folders
3. **File Existence Check**: Verify which specification files exist
4. **Content Analysis**: Check completion status of existing files
5. **Gap Identification**: Identify missing components
6. **Task assignment check**: If todo/doing tasks exist for current phase, route to appropriate phase handler

### State Categories & Workflow Mapping

#### Project States

1. **No Project Started** → Execute `/start` workflow
2. **Technology Validation Needed** → Execute `/assess` (research mode) to validate tech stack, then continue workflow
3. **Implementation Incomplete** → Execute `/implement`, then MUST run `/assess` (assessment mode) next (mandatory quality checkpoint)
4. **MANDATORY: After Implementation** → Execute `/assess` (assessment mode) to verify: performance, bugs, design quality (if UI), code quality, tests, security, docs
5. **Design Quality Low** (if UI) → If `/assess` finds design score <80 → MUST refactor then `/implement` (design focus)
6. **Quality Issues Found** → Execute `/assess` (assessment mode) to find gaps → `/assess` (research mode) to validate solutions → `/implement` fixes
8. **Quality Issues** → Execute `/assess` → [if design issues: `/design`] → `/optimize` → `/assess` (verify improvements) → `/fix`
9. **Security Issues** → Execute `/fix` → `/assess`
10. **All Phases Complete** → Generate recommendations

## Workflow Execution

### Execution Protocol

1. **Execute Detected Command**: Run appropriate prompt
2. **Monitor Progress**: Track phase completion
3. **Re-evaluate State**: Check for new state after each phase
4. **Continue Workflow**: Repeat until all phases complete
5. **Final Save**: Execute `/save` to persist final state

## Progress Tracking

Format:

```
State: [detected state]
Executing: [current phase]
Next: [next phase]
Progress: [X/12] phases completed
Status: [working/completed/failed]
Memory: [loaded/updated]
```

---

**Execute this orchestrator with `/continue` command to resume or complete your project!**
