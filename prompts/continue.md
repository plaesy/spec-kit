---
description: "Intelligent state detection and continuation orchestrator"
subagent: true
---

# `/continue` command instructions

⚡ **Run with**: `ultracode` (multi-agent parallel execution)

## Objective

Auto-detect current project state → Execute remaining phases → Complete

## Why This Phase Matters

**Purpose**: Resume interrupted work without losing context (async-friendly workflow)

**Why Now**: Enables seamless handoff between work sessions - detects progress, identifies gaps, and orchestrates remaining phases so you never restart from scratch

**What Breaks If Skipped**: Work context is lost; duplicate effort on completed phases; unknown current state leads to wrong phase execution; productivity drops from manual state detection

**Success Enables**: Confidence in resuming multi-phase work; automated phase detection; clean handoff between sessions; clear visibility on what remains

## Protocol

**Load memory state → Detect current phase → Execute remaining workflow**

## Validation Checklist (Before Running)

- ✅ Memory files exist (.plaesy/context.md, .plaesy/memory.md)
- ✅ Previous session state loadable
- ✅ Project directory accessible
- ✅ Git repo in valid state

## Memory Integration

**Follow**: `.plaesy/instructions/plaesy.md` → Context Handling (authoritative)

1. **Load `.plaesy/context.md`** - Previous session state
2. **Load `.plaesy/memory.md`** - Project knowledge index
3. **Restore Progress** - Continue from checkpoint

### Load Task State

**Follow**: `.plaesy/instructions/tasks.md` → Task Management Rules

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

### Project States

Read the active dimension(s) from `.plaesy/memory/constitution.md` (see `/start`
Phase 0) before routing — a `business`/`legal`/`marketing` deliverable never routes
through `/implement`'s TDD cycle; it routes through `/assess:{dimension}` (Mode 1
research → draft → Mode 1 ambiguity resolution) and `/doc` for the handoff artifact.
Full state → next-phase mapping is the pseudo-code under "Phase Detection +
Auto-Routing Logic" below — this section doesn't restate it separately.

## Workflow Execution

### Execution Protocol

1. **Execute Detected Command**: Run appropriate prompt
2. **Monitor Progress**: Track phase completion
3. **Re-evaluate State**: Check for new state after each phase
4. **Continue Workflow**: Repeat until all phases complete
5. **Final Save**: Execute `/save` to persist final state

## 🤖 Intelligent Phase Routing (NO User Input Required)

**AUTONOMY ENHANCEMENT**: Framework auto-decides next phase based on state, no questions asked.

### Phase Detection + Auto-Routing Logic

**After each phase execution**, framework auto-analyzes state:

```
IF project has no deliverable/specs yet
  → Auto-route to `/start` (full workflow from idea)

ELSE IF deliverable dimension detected but assessment never run
  → Auto-route to `/assess:{dimension}` (research mode - validate choices)
  → dimension = whatever `.plaesy/memory/constitution.md` declares active
    (technical for software, business/legal/marketing/design/product otherwise)

ELSE IF deliverable incomplete OR specs changed
  → software dimension: Auto-route to `/implement` (full TDD cycle)
  → non-software dimension: Auto-route to drafting/revision per that dimension's
    `assess-{dimension}.md` spine (no TDD cycle — draft → review → revise)

ELSE IF deliverable complete AND assess never run
  → Auto-route to `/assess:{dimension}` (assessment mode - mandatory quality check)
  → BLOCK: Cannot proceed without assess results

ELSE IF assess found issues AND <85 quality score
  → IF mandatory-audit failure (WCAG for design, unit-economics for business,
    clause traceability for legal, claim sourcing for marketing, failure-modes
    for architecture)
    → Auto-route to `/optimize` (redesign/refactor per Design Spine)
  → ELSE IF security/compliance issues (score <50)
    → Auto-route to `/fix` (security- or compliance-focused)
  → ELSE IF performance issues (software only)
    → Auto-route to `/optimize` (auto-improvement loop)
  → ELSE
    → Auto-route to `/loop` (autonomous fixes)

ELSE IF quality score ≥85 AND improvements found
  → Auto-route to `/optimize` (auto-improve, dimension-appropriate)
  → Auto-run `/assess` verification after

ELSE IF optimization complete
  → Auto-route to `/assess:{dimension}` (verify improvements)

ELSE IF all phases complete AND no blockers
  → Auto-route to `/save` (persist state)

ELSE IF blocked tasks exist
  → Alert user on blockers
  → Pause workflow (user must resolve)
```

### Autonomous Decisions During Routing

**No user input needed for**:

| Scenario | Auto-Decision |
|----------|---|
| **Which assessment scope?** | Detect from constitution's active dimension(s) — software→design+technical, API→technical, business plan→business+financial, contract set→legal, campaign→marketing |
| **Skip `/optimize`?** | Only if quality ≥90 AND the dimension's mandatory audit passes |
| **Run `/loop` or manual `/fix`?** | If issues are auto-fixable + not security/compliance-critical → `/loop` |
| **Next phase after `/fix`?** | Always `/assess` (verify fixes worked) |
| **When to commit?** | Software: after each sub-phase (RED, GREEN, REFACTOR) auto-commits. Non-software: after each drafted section passes its Mode 1 ambiguity resolution |
| **Parallel or sequential?** | Independent tasks run parallel; dependent tasks sequential |

### Real-Time Progress Display

```
🔄 CONTINUING PROJECT...
├─ Detected State: "Implementation incomplete, needs assessment"
├─ Current Phase: IMPLEMENT (80% complete, 3 components left)
├─ Quality Score: 68/100 (below 85 threshold)
├─ Next Phase: /assess (after implement completes)
│  └─ Why: Quality validation mandatory before optimization
├─ Blockers: None
└─ ETA: Complete remaining phases in ~2 hours
```

---

## Success Criteria (VERIFY Before Completing)

✅ **Clarity**:
- [ ] All specifications clear and documented
- [ ] No ambiguities remain

✅ **Quality**:
- [ ] Tests passing/metrics validated
- [ ] No regressions detected

✅ **Self-Audit**:
- [ ] Ready for next phase?
- [ ] Would ship this to production?

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
