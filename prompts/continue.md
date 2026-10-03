---
description: "Intelligent state detection and continuation orchestrator"
subagent: true
---

# `/continue` command instructions

⚡ **Run with**: any parallel-capable mode — a host that provides
subagents, or a mode that does. Not a requirement: a single-agent host
runs this identically, just serially. `/loop` documents when fan-out is
worth its cost.

## Objective

Auto-detect current project state → Execute remaining phases → Complete

## Why This Phase Matters

**Purpose**: Resume interrupted work without losing context (async-friendly workflow)

**Why Now**: Enables seamless handoff between sessions — detects progress, identifies
gaps, orchestrates remaining phases so you never restart from scratch

**What Breaks If Skipped**: work context lost between sessions; completed phases
redone; wrong phase executes because current state is unknown; state detection
falls back to manual bookkeeping

**Success Enables**: confidence resuming multi-phase work, automated phase detection
instead of manual triage, clean session handoff, clear visibility on what remains

## Protocol

Load memory state → Detect current phase → Execute remaining workflow

## Validation Checklist (Before Running)

- ✅ Memory files exist (.plaesy/context.md, .plaesy/memory.md, .plaesy/decisions.md)
- ✅ Previous session state loadable
- ✅ Project directory accessible
- ✅ Git repo in valid state

## Memory Integration

**Follow**: `.plaesy/instructions/plaesy.md` → Context Handling (authoritative)

1. **Load `.plaesy/context.md`** - Previous session state
2. **Load `.plaesy/memory.md`** - Project knowledge index
3. **Load `.plaesy/decisions.md`** - Full decision history, follow links to `.plaesy/decisions/{topic}.md`
4. **Restore Progress** - Continue from checkpoint

### Load Task State

**Follow**: `.plaesy/instructions/tasks.md` → Task Management Rules

1. **Scan `.plaesy/tasks/`**: count tasks in `backlog/`, `todo/`, `doing/`, `done/`,
   `blocked/`; parse each task's YAML frontmatter for Status, Phase, Acceptance
   Criteria, Related dependencies
2. **Extract phase/status**: load current active phase from `.plaesy/context.md`;
   filter `todo/`+`doing/` tasks matching it; flag `doing/` tasks with unmet
   acceptance criteria (unchecked boxes) as incomplete
3. **Alert on blocked tasks**: if `.plaesy/tasks/blocked/` has entries, surface the
   blocker reason and blocking task references (e.g. "2 tasks blocked (waiting on:
   implementation phase, external API)") — user must resolve manually
4. **Load task summary into context**: add "## In-Flight Tasks" to
   `.plaesy/context.md`; keep it ≤100 lines, extracting overflow to
   `.plaesy/memory/tasks-summary.md`

## State Detection Logic

### Detection Process

1. **Load task state**: Scan `.plaesy/tasks/` for active work
2. **Scan `.plaesy/specs/` Directory**: Identify active project folders
3. **File Existence Check**: Verify which specification files exist
4. **Content Analysis**: Check completion status of existing files
5. **Gap Identification**: Identify missing components
6. **Task assignment check**: If todo/doing tasks exist for current phase, route to appropriate phase handler
7. **Analysis freshness check**: If `.plaesy/analysis/overview.md` exists and state
   detection is about to rely on it (e.g. routing decisions that cite tech stack,
   file counts, or detected tools), run `plaesy analyze` first — regeneration is
   skipped automatically (default behavior) when the project fingerprint hasn't
   changed since the last run, so this is cheap to call unconditionally; it only
   does real work when the project has actually drifted. Never route off
   `overview.md` content without this call; a stale analysis silently produces
   wrong routing decisions (missing new dependencies, deleted files, changed stack).

### Project States

Read the active dimension(s) from `.plaesy/memory/constitution.md` (see `/start`
Phase 1, sub-step 1.0 — Constitution) before routing — a
`business`/`legal`/`marketing` deliverable never routes through `/implement`'s
TDD cycle; it routes through `/assess:{dimension}` (Mode 1 research → draft →
Mode 1 ambiguity resolution) and `/doc` for the handoff artifact. Which command a
given dimension's findings land in is decided by
`.plaesy/instructions/dimension-mapping.md` (the canonical routing table). Full
state → next-phase mapping is the pseudo-code under "Phase Detection +
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

```text

IF state.json.active_role is set AND that role's task is Status: done
  → Lookup `.plaesy/instructions/role-mapping.md`'s Routing Table for
    active_role's "Typical next role"
  → IF a next role exists AND it is not a hard-stop category
    (`.plaesy/instructions/plaesy.md` rule 8)
    → Write `{from: active_role, to: next_role, reason, at}` to
      `state.json.role_handoff_log`; set `active_role = next_role`
    → Auto-route to the next role (load `.plaesy/roles/{next_role}.md`,
      resume whichever command family started the chain)
    → No hop limit — repeat this branch every time a role completes; the
      only things that stop the chain are "no next role" (end of chain,
      emit `## Handoff: none`) or a hard-stop category
  → ELSE (end of chain, or hard-stop hit)
    → Clear `active_role` to `null`, report the chain's final state, continue
      the cascade below as normal

ELSE IF project has no deliverable/specs yet
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

ELSE IF assess found issues AND score < {{QUALITY_REMEDIATION_FLOOR}} (85)
  → IF mandatory-audit failure (WCAG for design, unit-economics for business,
    clause traceability for legal, claim sourcing for marketing, failure-modes
    for architecture)
    → Auto-route to `/fix` (redesign/refactor per Design Spine), or `/loop` to
      auto-remediate. A failed audit caps the dimension at 49, and `/optimize`
      refuses to start below the remediation floor — routing here would send the
      work to the one command that must decline it.
  → ELSE IF security/compliance issues (a failed blocking gate, or the `D` band)
    → Auto-route to `/fix` (security- or compliance-focused)
  → ELSE IF performance issues (software only)
    → Auto-route to `/loop` (auto-improvement loop); `/optimize` only once the
      score reaches the remediation floor
  → ELSE
    → Auto-route to `/loop` (autonomous fixes)

ELSE IF score ≥ {{QUALITY_REMEDIATION_FLOOR}} (85) AND improvements found
  → Auto-route to `/optimize` (auto-improve, dimension-appropriate)
  → Auto-run `/assess` verification after

ELSE IF optimization complete
  → Auto-route to `/assess:{dimension}` (verify improvements)

ELSE IF all phases complete AND no blockers
  → IF current phase == /save (already terminal — improve+save both ran)
    → Report "Project complete. Nothing remaining." and stop routing.
  → ELSE (tasks just completed, or improve has not run since last save)
    → Auto-route to /improve (improvement gate before final save)
    → /improve sets phase=/improve; when it finds nothing to do it routes to /save,
      which sets phase=/save and breaks the loop.

ELSE IF blocked tasks exist
  → Alert user on blockers
  → Pause workflow (user must resolve)

ELSE  (no state above matched)
  → Do NOT guess a phase. "Nothing matched" means the run is in a state this
    cascade does not describe — usually a `.plaesy/memory/context.md` phase
    value that no phase owns, or a project with no constitution.
  → Report the four inputs that decide the branch: the `phase` recorded in
    context.md, the task counts per status in `.plaesy/tasks/`, whether
    `/assess` has run and its score, and whether `/optimize` has run.
  → If `.plaesy/memory/constitution.md` is missing → route to `/start` (the
    constitution sets the dimensions every other decision depends on).
  → Otherwise set `status: blocked` in context.md and stop. An undefined
    fall-through that picks a phase is how a run skips a mandatory gate.
```

`/improve` and `/save` appear in this logic as *routing targets*, not as phases —
`/improve` is an out-of-phase gate invocable at any point and `/save` (Phase 9) is
the ladder's terminal step. Neither consumes a phase slot beyond Phase 9 itself.

### Autonomous Decisions During Routing

**No user input needed for**:

| Scenario | Auto-Decision |
|----------|---|
| **Which assessment scope?** | Detect from constitution's active dimension(s) — software→design+technical, API→technical, business plan→business+financial, contract set→legal, campaign→marketing |
| **Skip `/optimize`?** | Only if the score is at or above `{{QUALITY_SKIP_OPTIMIZE}}` (default 90) AND the dimension's mandatory audit passes — both thresholds are defined in `.plaesy/instructions/quality-gates.md` |
| **Run `/loop` or manual `/fix`?** | If issues are auto-fixable + not security/compliance-critical → `/loop` |
| **Next phase after `/fix`?** | Always `/assess` (verify fixes worked) |
| **When to commit?** | Software: after each sub-phase (RED, GREEN, REFACTOR) auto-commits. Non-software: after each drafted section passes its Mode 1 ambiguity resolution |
| **Parallel or sequential?** | Independent tasks run parallel; dependent tasks sequential |
| **Chain to the next role?** | Always, per `.plaesy/instructions/role-mapping.md`'s Routing Table — unless the next step is a hard-stop category (`.plaesy/instructions/plaesy.md` rule 8), in which case halt and report |

### Real-Time Progress Display

```text

🔄 CONTINUING PROJECT...
├─ Detected State: "Implementation incomplete, needs assessment"
├─ Current Phase: Implementation (Phase 2) — 80% complete, 3 components left
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

- [ ] Closing check applied (`.plaesy/instructions/quality-gates.md`
      → **Closing check**): ready for the next phase, and would this ship?

## Progress Tracking

Format:

```text

State: [detected state]
Executing: [current phase]
Next: [next phase]
Progress: [X/9] phases completed
Status: [working/completed/failed]
Memory: [loaded/updated]
```

The `9` in `X/9` above is the canonical phase count (`.plaesy/instructions/workflow-phases.md`:
9 phases, 1-9) — not a command. Phase 3 is an automatic gate with no command, so clearing
it never increments the counter. `/continue` and `/loop` are out-of-phase drivers
— they run a phase or cross several, and are never counted as one.

---

## Output Format

`/continue` produces a progress report every phase transition and a final
completion report when all phases are done.

### Phase Transition Report (emitted after each phase completes)

```text
CONTINUE — Phase Transition
Detected State: [state description]
Phase Completed: [phase name/number]
Quality Score: [NN]/100 (dimension: [dimension])
Next Phase: [next command]
Blockers: [none | description]
ETA: [estimated remaining time]
Memory: [loaded/updated]
```

### Final Completion Report (when all 9 phases complete)

```text
CONTINUE — Project Complete
Phases Executed: [list of phases run this session]
Final Quality: [dimension: NN/100, ...]
Deliverables: [list of produced artifacts with paths]
Memory Checkpoints: [context.md lines, memory.md entries]
Next: [none — project complete, or /improve if improvements remain]
```

**Files** (only when state changes):

- `.plaesy/context.md` — current session state (≤100 lines)
- `.plaesy/memory.md` — knowledge index updated
- `.plaesy/decisions.md` — decisions index updated
- `.plaesy/tasks/` — task status changes

---

## Worked Example — Resume Interrupted Implementation

**Scenario**: Previous session stopped mid-`/implement:technical` with 3 of 5
components complete. User runs `/continue` to finish.

```text
CONTINUE — Phase Transition
Detected State: Implementation incomplete, 3/5 components done
Phase Completed: Implementation (Phase 2) — 60% complete
Quality Score: N/A (assessment not yet run)
Next Phase: /assess:technical (mandatory quality gate)
Blockers: None
ETA: ~45 min (2 components + assess)
Memory: loaded

CONTINUE — Phase Transition
Detected State: Implementation complete, assessment pending
Phase Completed: Assessment (Phase 4) — technical
Quality Score: 78/100 (band: C)
Next Phase: /optimize:technical --focus performance (score < 85)
Blockers: None
ETA: ~30 min
Memory: updated

CONTINUE — Project Complete
Phases Executed: [2, 3, 4, 5, 6, 8, 9]
Final Quality: technical: 87/100
Deliverables: [src/api/, src/service/, tests/, docs/]
Memory Checkpoints: context.md (42 lines), memory.md (12 entries)
Next: none — project complete
```

---

**Execute this orchestrator with `/continue` command to resume or complete your project!**
