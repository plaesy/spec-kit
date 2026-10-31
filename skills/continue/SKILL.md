---
name: continue
description: "Intelligent state detection and continuation orchestrator Use when the user types /continue."
disable-model-invocation: true
---

# `/continue` command instructions

Run with: any parallel-capable mode. A single-agent host runs it serially, which is fine. See `/loop` for when fan-out is worth its cost.

## Success Contract

Write `contracts.json` in the task folder before work starts (one `run` and `expected` per item). Run `plaesy tasks validate <task>`. Report completion only when every item is `passed` (NN-08; `.plaesy/instructions/success-contract.md`).

## Objective

Detect the current project state, run the remaining phases, and finish. This lets interrupted work resume without losing context or redoing completed phases. Detection also prevents the wrong phase from running.

## Before You Start

- [ ] Memory files exist: `.plaesy/context.md`, `.plaesy/memory.md`, `.plaesy/decisions.md`.
- [ ] The previous session's state loads.
- [ ] The project directory is accessible and the git repo is in a valid state.

## Memory

Follow `.plaesy/instructions/plaesy.md`, section Context Handling (authoritative).

1. Load `.plaesy/context.md` (previous session state).
2. Load `.plaesy/memory.md` (knowledge index).
3. Load `.plaesy/decisions.md`, and follow its links to `.plaesy/decisions/{topic}.md`.
4. Restore progress from the checkpoint.

Task state (`.plaesy/instructions/tasks.md`, Task Management Rules):

1. Scan `.plaesy/tasks/`. Count tasks in `backlog/`, `todo/`, `doing/`, `done/`, and
   `blocked/`. Parse each task's frontmatter: status, phase, acceptance criteria, dependencies.
2. Take the active phase from `.plaesy/context.md`. Filter `todo/` and `doing/` tasks for it. Flag
   `doing/` tasks with unchecked acceptance criteria as incomplete.
3. If `blocked/` has entries, surface each blocker and what it waits on (for example "2 tasks blocked:
   waiting on implementation, external API"). The user must resolve these.
4. Add an "## In-Flight Tasks" section to `.plaesy/context.md`. Keep that file at 100 lines or fewer.
   Move overflow to `.plaesy/memory/tasks-summary.md`.

## Detection

1. Load task state from `.plaesy/tasks/`.
2. Check which specification files exist.
3. Check the completion status of each file that exists.
4. Identify the missing components.
5. If todo or doing tasks exist for the current phase, route to that phase's handler.
6. Analysis freshness. If you are about to route on `.plaesy/analysis/overview.md` (stack, file
   counts, tools), run `plaesy analyze` first. It skips regeneration when the fingerprint is unchanged,
   so calling it is cheap. A stale analysis gives wrong routes: missing dependencies, deleted files, changed stack.

## Dimensions

Read the active dimension(s) in `.plaesy/memory/constitution.md` before routing. A `business`, `legal`,
or `marketing` deliverable never goes through `/implement`'s TDD cycle. It goes through
`/assess:{dimension}` (research, draft, ambiguity resolution), then `/doc` for the handoff. The
command for each finding comes from `.plaesy/instructions/dimension-mapping.md`. The routing
cascade below is the full state-to-phase mapping.

## Execution

1. Run the detected command.
2. Track phase completion.
3. Re-check state after each phase.
4. Repeat until every phase is complete.
5. Run `/save` to persist the final state.

## Phase Routing (No User Input)

After each phase, detect the state and route:

```text
IF state.json.active_role is set AND that role's task is Status: done
  → Look up active_role's "Typical next role" in .plaesy/instructions/role-mapping.md
  → IF a next role exists
      → Append {from, to, reason, at} to state.json.role_handoff_log; set active_role = next_role
      → Route to that role (load .plaesy/roles/{next_role}.md; resume the command family that started the chain)
      → No hop limit. Repeat every time a role completes. Only "no next role" ends the chain
        (emit "## Handoff: none").
  → ELSE (end of chain)
      → Set active_role = null, report the chain's final state, continue the cascade below

ELSE IF the project has no deliverable or specs yet
  → Route to /start (full workflow from idea)

ELSE IF a deliverable dimension is detected but assessment never ran
  → Route to /assess:{dimension} (research: validate choices)
  → dimension = the active one in the constitution

ELSE IF the deliverable is incomplete OR specs changed
  → Software: /implement (full TDD cycle)
  → Other dimensions: drafting and revision per the dimension's assess-{dimension}.md
    (draft → review → revise, no TDD)

ELSE IF the deliverable is complete AND assess never ran
  → Route to /assess:{dimension} (mandatory quality check)
  → BLOCK: nothing proceeds without assess results

ELSE IF assess found issues AND score < {{QUALITY_REMEDIATION_FLOOR}} (85)
  → IF a mandatory audit failed (WCAG for design, unit economics for business, clause
      traceability for legal, claim sourcing for marketing, failure modes for architecture)
      → Route to /fix (redesign or refactor per the Design Spine), or /loop to auto-remediate.
        A failed audit caps the dimension at 49, and /optimize refuses to start below the floor.
  → ELSE IF security or compliance issues (a failed blocking gate, or band D)
      → Route to /fix (security or compliance focus)
  → ELSE IF performance issues (software only)
      → Route to /loop. Use /optimize only once the score reaches the floor.
  → ELSE
      → Route to /loop (autonomous fixes)

ELSE IF score ≥ {{QUALITY_REMEDIATION_FLOOR}} (85) AND improvements found
  → Route to /optimize (dimension-appropriate), then run /assess to verify

ELSE IF optimization is complete
  → Route to /assess:{dimension} (verify improvements)

ELSE IF all phases are complete AND no blockers
  → IF the current phase is /save (already terminal)
      → Report "Project complete. Nothing remaining." Stop routing.
  → ELSE (tasks just completed, or /improve has not run since the last save)
      → Route to /improve (improvement gate before final save)
      → /improve sets phase=/improve. When it finds nothing to do, it routes to /save,
        which sets phase=/save and ends the loop.

ELSE IF blocked tasks exist
  → Alert the user on the blockers. Pause (the user must resolve them).

ELSE (no state above matched)
  → Do NOT guess a phase. An unmatched state means this cascade does not describe the run.
    Usually: a context.md phase that no phase owns, or a project with no constitution.
  → Report four inputs: the phase in context.md, task counts per status in .plaesy/tasks/,
    whether /assess ran and its score, and whether /optimize ran.
  → If .plaesy/memory/constitution.md is missing → route to /start.
  → Otherwise set status: blocked in context.md and stop. Guessing a phase is how a run skips a mandatory gate.
```

`/improve` and `/save` are routing targets here, not phases. `/improve` is an out-of-phase gate that
can run at any point. `/save` is the terminal step of Phase 9.

## Autonomous Decisions

No user input is needed for these:

- Assessment scope. Software: technical (add design when there is a UI). API: technical. Business plan: business and financial. Contract set: legal. Campaign: marketing.
- Skip `/optimize`? Only if the score is at or above `{{QUALITY_SKIP_OPTIMIZE}}` (default 90) and the dimension's mandatory audit passes. Both thresholds are in `quality-gates.md`.
- `/loop` or manual `/fix`? If the issues are auto-fixable and not security or compliance-critical, use `/loop`.
- Next phase after `/fix`. Always `/assess`, to verify the fix.
- When to commit. Only when the user asked. Software: each RED, GREEN, and REFACTOR step. Other dimensions: each drafted section after its ambiguity resolution.
- Parallel or sequential. Independent tasks run in parallel. Dependent tasks run in sequence.
- Chain to the next role. Always, per `role-mapping.md`. Nara decides each step, including the rule 8 categories (`plaesy.md`), under the NN articles.

## Progress Display

```text
🔄 CONTINUING PROJECT...
├─ Detected State: "Implementation incomplete, needs assessment"
├─ Current Phase: Implementation (Phase 2), 80% complete, 3 components left
├─ Quality Score: 68/100 (below {{QUALITY_REMEDIATION_FLOOR|85}})
├─ Next Phase: /assess after implement completes. Quality validation comes before optimization.
├─ Blockers: none
└─ ETA: ~2 hours
```

## Success Criteria

- [ ] Every Success Contract item is compared with its result at the gate. Blocking items are met.
- [ ] Clarity: specifications are clear and documented, with no open ambiguity.
- [ ] Quality: tests pass and metrics are validated. No regressions.
- [ ] Closing check applied (`.plaesy/instructions/quality-gates.md`, section Closing check): ready for the next phase, and would this ship?

## Progress Tracking

```text
State: [detected state]
Executing: [current phase]
Next: [next phase]
Progress: [X/9] phases completed
Status: [working | completed | failed]
Memory: [loaded | updated]
```

The `9` is the canonical phase count (`.plaesy/instructions/workflow-phases.md`). Phase 3 is an
automatic gate with no command, so clearing it does not advance the count. `/continue` and `/loop`
are out-of-phase drivers. They are never counted.

## Output Format

Each phase transition produces a report. The run ends with a completion report.

```text
CONTINUE — Phase Transition
Detected State: [state]
Phase Completed: [phase]
Quality Score: [NN]/100 (dimension: [dimension])
Next Phase: [next command]
Blockers: [none | description]
ETA: [remaining time]
Memory: [loaded | updated]
```

```text
CONTINUE — Project Complete
Phases Executed: [list]
Final Quality: [dimension: NN/100, ...]
Deliverables: [artifacts with paths]
Memory Checkpoints: [context.md lines, memory.md entries]
Next: [none, or /improve if improvements remain]
```

Files written (only when state changes): `.plaesy/context.md` (100 lines or fewer),
`.plaesy/memory.md`, `.plaesy/decisions.md`, and `.plaesy/tasks/` status changes.

---

## Worked Example — Resume an Interrupted Implementation

Scenario: the last session stopped in `/implement:technical` with 3 of 5 components done.

```text
CONTINUE — Phase Transition
Detected State: Implementation incomplete, 3/5 components done
Phase Completed: Implementation (Phase 2), 60%
Quality Score: N/A (assessment not yet run)
Next Phase: /assess:technical (mandatory quality gate)
ETA: ~45 min (2 components + assess)

CONTINUE — Phase Transition
Detected State: Implementation complete, assessment pending
Phase Completed: Assessment (Phase 4), technical
Quality Score: 78/100 (band: C)
Next Phase: /optimize:technical --focus performance (score below 85)
ETA: ~30 min

CONTINUE — Project Complete
Phases Executed: [2, 3, 4, 5, 6, 8, 9]
Final Quality: technical 87/100
Deliverables: [src/api/, src/service/, tests/, docs/]
Memory Checkpoints: context.md (42 lines), memory.md (12 entries)
Next: none, project complete
```

---

Run `/continue` to resume or finish your project.
