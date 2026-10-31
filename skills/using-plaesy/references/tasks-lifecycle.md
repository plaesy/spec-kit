
# Task Management Instructions

A task folder holds `task.md` (status, topic, title) and `contracts.json` (checklist). The spec and plan for a task live with its topic: `.plaesy/spec/<topic>/spec.md` (written by the `spec` skill) and `.plaesy/plan/<topic>/plan.md` (written by the `plan` skill). Templates: `../assets/contract.json`.

## Overview

Tasks in Plaesy are organized by status in the `.plaesy/tasks/` directory structure. Each task moves through lifecycle stages from conception to completion. The system enforces consistent workflow through status transitions, quality gates, and validation checkpoints.

## Task Management Rules

### Create Tasks

- Run `plaesy tasks list`, if exist continue with that tasks.
- If not exists run `plaesy tasks create <priority> <category> <title>`. The topic is `<category>-<title>`, and the folder is `<priority>_<topic>`.
- Write the topic spec with the `spec` skill (`.plaesy/spec/<topic>/`), including design.md and mock.html for any UI, and api.md, data.md, or security.md when they apply.
- Write the topic plan with the `plan` skill (`.plaesy/plan/<topic>/plan.md`), and the checklist (contracts.json).

### Moving Tasks

- Move the task folder between status directories as work progresses
- Use the CLI for every move. It rewrites the `status:` line in `task.md` to match the folder
  - `plaesy tasks next` — `backlog/` → `todo/` (the most urgent backlog task of the spec)
  - `plaesy tasks start <task>` — `todo/` → `doing/`
  - `plaesy tasks complete <task>` — `doing/` → `done/`
  - `plaesy tasks block <task> [reason]` — into `blocked/`; appends a `## Blocked Reason` section to `task.md`
  - `plaesy tasks unblock <task>` — `blocked/` → `todo/`
  - `plaesy tasks move <task> <from_status> <to_status>` — any other ad-hoc move
- A plain `mv` moves the folder but skips the `status:` rewrite and the Blocked Reason section
- Don't delete tasks — move them to the right status (`.plaesy/tasks/<status>/<task>/`)

## Task Organization

### Directory Structure

```text

.plaesy/tasks/
├── backlog/       → Ideas, features, bugs (unscheduled)
├── todo/          → Ready to start (next in queue)
├── doing/         → In active work
├── done/          → Completed and validated
└── blocked/       → Waiting on dependency (on hold)
```

### Work-in-Progress (WIP) Limits

Kanban best practice limits how many tasks can sit in each stage so blockers surface early and teams finish before starting. Set and monitor WIP limits per status directory:

| Status | Recommended WIP Limit | Rationale |
|--------|----------------------|-----------|
| `todo/` | Team members count | Enough ready work, never idle |
| `doing/` | 1 per team member max | Forces finishing before starting |
| `blocked/` | No hard cap, but age-tracked | Blocked work ages; track and resolve quickly |

When a stage hits its WIP limit, no new task moves in until one moves out. Do not start new work while a task is blocked — swarm on the blocker or wait. Review WIP limits each cycle; mature teams run lower limits to expose bottlenecks.

Record the current WIP count in `.plaesy/context.md` after each session.

### Task Lifecycle

```text

backlog → todo → doing → [quality-gates] → done / blocked
```

**Allowed transitions:**

| From → To | CLI command | Validation |
|-----------|-------------|------------|
| `backlog/` → `todo/` | `plaesy tasks next` or `plaesy tasks run` | Task must pass Definition of Ready |
| `todo/` → `doing/` | `plaesy tasks start <task>` | Task must be assigned; topic plan approved (`.plaesy/plan/<topic>/plan.md` status `approved`) |
| `doing/` → `done/` | `plaesy tasks complete <task>` | All contracts.json checks `passed`, DoD met |
| `doing/` → `blocked/` | `plaesy tasks block <task> [reason]` | Reason required; `## Blocked Reason` appended to task.md |
| `blocked/` → `todo/` | `plaesy tasks unblock <task>` | Dependency resolved; reason removed from task.md |
| Any → any | `plaesy tasks move <task> <from> <to>` | Ad-hoc; status line in task.md rewritten |

A plain `mv` moves the folder but skips the `status:` rewrite and the Blocked Reason section — never use `mv` for task status changes.

**Stage skip is forbidden:** `backlog/` → `doing/`, `todo/` → `done/`, or `backlog/` → `done/` without passing through intermediate stages. If a task needs to jump stages, report it as an anomaly in `.plaesy/context.md` before proceeding.

## Task Format

A task is a folder, never a single file. Create it with the CLI, not by hand:

```text

.plaesy/tasks/<status>/<priority>_<category>-<title>/
├── task.md          status, topic, title
│  (spec and plan live in .plaesy/spec/<topic>/ and .plaesy/plan/<topic>/)
└── contracts.json   the checklist that tasks, run `plaesy tasks validate` after you already finish the job and tasks will move from `doing` into `done`
```

### INVEST Criteria for Quality Tasks

Every task should pass the INVEST criteria to ensure quality and maintainability:

- **I** *Independent*: Can be implemented and deployed independently of other tasks
- **N** *Negotiable*: Scope is flexible and can be negotiated with stakeholders
- **V** *Valuable*: Provides clear value to users or the business
- **E** *Estimable*: Team can accurately estimate effort and complexity
- **S** *Small*: Can be completed within 1-4 hours of work
- **T** *Testable*: Has clear acceptance criteria that can be objectively verified

### Naming Convention

- Folder: `<priority>_<category>-<title>`, where priority is `critical`, `high`, `medium`, or `low`, and slug is kebab-case
- Example: `critical_setup-auth-middleware/`
- Create one with `plaesy tasks create <priority> <category> <title>`. It lands in `backlog/` with a `task.md` and an empty `contracts.json`, on the topic `<category>-<title>`

### task.md

```markdown

---
spec: [spec name, the git branch the task belongs to]
title: [title]
phase: [design|implement|assess|optimize|fix]
status: [backlog|todo|doing|done|blocked]
---

## Goal
One sentence: the result this task produces.

## Scope
What this task changes, and what it does not.

## Context
Files, state, and constraints the work depends on.

## Description
Clear description of what needs to be done.

## Decision
- ... list decisions

## Acceptance Criteria
- [ ] Criterion 1
- [ ] Criterion 2
- [ ] Criterion 3

## KPI
- Measurable result, with the number or evidence that counts as success

## Verification
- Command or check that produces the KPI value (for example `go test ./pkg/ -run '^TestX$'`)

## Reference
- [topics](memory/{topics})
- ... list `.plaesy/memory/{topics}.md` for reference (see: `.plaesy/memory.md` for memory table of contents)

## Dependencies
- Other task this one waits on (if blocked by another task, ex: `.plaesy/tasks/{status}/<task>/task.md`)
- External dependency (if any)

## Notes
Implementation notes, references, or constraints.

## Next
Next todo list for this tasks
```

## Definition of Ready (DoR)

Before moving a task to `todo/`, verify it meets the Definition of Ready:

- [ ] **Independent**: Can be started without waiting on other tasks (unless explicitly blocked)
- [ ] **Negotiable**: Scope is clear but open to refinement
- [ ] **Valuable**: Purpose and business value are documented
- [ ] **Estimable**: Effort can be estimated (t-shirt size or story points)
- [ ] **Small**: Fits within 1-4 hours of work
- [ ] **Testable**: Acceptance criteria are measurable and verifiable
- [ ] **Spec.md complete**: Goal, scope, description, acceptance criteria all present
- [ ] **Dependencies known**: All blocking tasks and external dependencies listed
- [ ] **contracts.json initialized**: Checklist created with expected values
- [ ] **Topic plan approved**: every requirement in the topic spec has a step and a verification

### Blocking check: If any dependency is pending or another task is blocked

- [ ] **No pending dependency**: All task dependencies are either complete or the task is marked `blocked/` with a reason
- [ ] **Blocker not stale**: If a dependency is blocked, that blocker's age in `.plaesy/context.md` is under 7 days, or the task waits in `backlog/` until it clears

## Definition of Done (DoD)

Before moving a task from `doing/` to `done/`, verify all of:

- [ ] **Acceptance criteria met**: Every criterion in the topic spec is checked off
- [ ] **contracts.json passed**: `plaesy tasks validate <task>` passes every check
- [ ] **Tests pass**: Task-related tests pass (`pass` expected value)
- [ ] **Build clean**: Project builds/compiles without errors
- [ ] **lint clean**: No new lint errors introduced
- [ ] **Documentation updated**: Public interfaces documented if applicable
- [ ] **No temporary files**: Scratch/test files created during work removed
- [ ] **Status line updated**: `status: done` in task.md frontmatter
- [ ] **Quality gate**: Per `.plaesy/instructions/quality-gates.md`

### Blocked-task aging policy

When a task moves to `blocked/`, its `## Blocked Reason` section in task.md must include:

```markdown
## Blocked Reason

- **Reason**: [why it is blocked]
- **Blocked since**: [ISO date, e.g., 2026-10-26]
- **Depends on**: [task folder path or external dependency]
- **Escalation due**: [blocked-since + 7 days]
```

A blocked task older than **7 days** must be reviewed at the next session. Options:

1. **Unblock** — dependency resolved; move to `todo/`
2. **Reclassify** — the blocker is not coming; move to `backlog/` with a note
3. **Split** — decompose the task so unblocked parts can proceed

Record the review decision in `.plaesy/context.md` and the task's `task.md` under a new `## Blocked Review` section. A task that ages past 14 days without review is a process failure — flag it in the next `/save` or `/continue` run.

## Task Estimation

Estimate effort before moving a task to `doing/`. Record the estimate in task.md:

| Size | Range | Example |
|------|-------|---------|
| XS   | < 30 min | Fix typo, update comment |
| S    | 30 min - 2 hrs | Small feature tweak, refactor single function |
| M    | 2 hrs - 4 hrs | Add endpoint with validation, new page component |
| L    | > 4 hrs | Break into smaller tasks before starting |

Tasks larger than `L` must be decomposed before moving to `doing/`. Use reference tasks — previously completed tasks of known size — to calibrate estimates.

## Task Prioritization

Prioritize backlog tasks using one of these methods:

- **MoSCoW**: Must have / Should have / Could have / Won't have this time
- **WSJF**: Weighted Shortest Job First — Cost of Delay divided by job duration
- **RICE**: Reach × Impact × Confidence ÷ Effort

Record the priority level in the folder name (`critical`, `high`, `medium`, `low`) and the prioritization method used in task.md notes.

## Risk Assessment

Before starting a task in `doing/`, identify risks and mitigations:

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| [description] | [low/med/high] | [low/med/high] | [how to reduce or handle] |

Record risks in task.md under **Notes** or a dedicated **Risk** section.

## Dependency Management

List dependencies in task.md under **Dependencies**:

- **Task dependency**: Another task this one waits on (link to its task.md path)
- **External dependency**: Service, API, or resource required but not owned

If blocked by another task, move to `blocked/` and note the dependency. When the dependency completes, move back to `todo/` with `plaesy tasks unblock`.

**Blocked tasks age.** Track `blocked/` age in `.plaesy/context.md` (WIP count includes blocked age). A blocked task exceeding 7 days triggers a mandatory review per the **Blocked-task aging policy** in the Definition of Done.

## Task Folder

Every task is a folder, `.plaesy/tasks/<status>/<task>/`. It holds these files:

- `.plaesy/spec/<topic>/spec.md`: requirements R1..Rn with Given/When/Then acceptance criteria, non-goals, and the design, API, data, and security files the topic needs.
- `.plaesy/plan/<topic>/plan.md`: how each R is met and verified, with risks, rollback, and a go/no-go gate. Written by the `plan` skill before any code changes, and approved before the task moves to `doing`. Its status line says `draft`, `approved`, or `superseded`.
- `contracts.json`: the checklist. Each item has a `run` command and an `expected` value. `plaesy tasks validate <task>` runs them and moves the task to `done` only when every blocking item passes (`.plaesy/instructions/success-contract.md`).

Keep one source of truth for each fact. Visual tokens belong in `.plaesy/memory/design.md`, research in `.plaesy/memory/`, and architectural decisions in `.plaesy/decisions/`. The task links to them and does not copy them. Every link must resolve to a file that exists.

## Task Contract (`contracts.json`)

Every task folder holds a `contracts.json` with one entry per check:

```json
{
  "id": "<task-folder-name>",
  "checklist": [
    {"id": "tests", "check": "The project test suite passes", "run": "<command>", "expected": "pass", "status": "pending", "histories": []}
  ]
}
```

- **Allowed item keys:** `id`, `check`, `run`, `expected`, `status`, `histories`. Any other key is rejected, so a misspelt key cannot silently drop a check.
- **`status`:** `pending` until the check runs, then `passed` or `failed`.
- **`histories`:** one entry per failed run, `{"at", "actual", "run"}`.
- **Check the file** with `plaesy tasks validate <task>` before the first run. A task is done only when every item is `passed`.

### Expected Values

`expected` is judged by `scripts/internal/expected/expected.go`. Words are case-insensitive. These are the only accepted values. Any other text, such as a sentence, is never met, and its history names these values.

| `expected` | Met when |
|---|---|
| `pass`, `yes`, `true`, `present` | the command exits 0. The output is not read. |
| `zero`, `none` | the last output line is `0` or `none` |
| `no`, `false`, `absent` | the last output line is `0`, `none`, `no`, `false` or `absent` |
| `all`, `every`, or `all <words>`, `every <words>` | the last output line is `pass`, or `N of N` with N above 0 |
| a comparison: `>=`, `>`, `<=`, `<`, `=`, `==` (or `≥`, `≤`), then a number and an optional unit, such as `>= 9`, `<= 2 s`, `>= 80%` | the last output line is a number that meets the comparison |

- A run that exits non-zero fails, whatever `expected` is.
- The last output line is the measurement. A command that prints nothing fails every value except `pass`, `yes`, `true` and `present`.
- A comparison needs the output's unit to match the expected unit. Time units convert (`ms`; `s`, `sec`, `secs`; `min`, `mins`; `h`, `hr`, `hrs`). A missing or different unit fails the item.

### Validating Memory Files

Run `plaesy memory validate`
after hand-editing `.plaesy/memory/*.md` files — it checks the self-containment
rule (no external/absolute paths inside `.plaesy/memory/`). `/save` does not run
this automatically; run it yourself when you've made a structural edit to
memory files, not for routine content updates.

### Priority Guidelines

1. **CRITICAL** — Blocks other work or required for MVP
2. **HIGH** — Important but not blocking
3. **MEDIUM** — Nice-to-have, improves quality
4. **LOW** — Polish, optimization, tech debt

BACKLOG is a status directory, not a priority — use `plaesy tasks next` to promote backlog tasks to todo.

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

**Autonomous loop** (via `/loop`):

```bash

/loop /continue
```

This workflow:

1. Reads task folders from `.plaesy/tasks/todo/` (if any exist)
2. Auto-detects project state and executes remaining phases
3. A finished task goes to `.plaesy/tasks/done/<task>/` when `plaesy tasks validate <task>` passes every blocking check
4. A task that cannot finish goes to `.plaesy/tasks/blocked/<task>/` with `plaesy tasks block <task> <reason>`
5. Reports progress to `.plaesy/context.md`
6. Can loop autonomously via `/loop` for repeated execution

## Best Practices

### Task Decomposition

- Break large features into 1-4 hour tasks, respecting INVEST criteria
- One task = one clear outcome
- Include acceptance criteria (not just description)
- Decompose tasks larger than `L` size before starting
- Use reference tasks for calibration

### Definition of Ready / Definition of Done

Follow the DoR and DoD checklists in each task folder. A task cannot move to `todo/` without DoR passing, and cannot move to `done/` without DoD passing.

### Documentation

- Link to relevant spec documents in task
- Reference related tasks (dependencies, blocked-by)
- Keep notes of blockers or decisions made during execution
- Record risks in task.md

### Quality Gates

Tasks in `doing` → `done` must pass:

- Code review (if applicable)
- Tests passing (if applicable)
- Documentation complete (if applicable)
- Acceptance criteria met
- Constitutional compliance (per `/start` principles)
- DoD checklist completed
- contracts.json every item `passed`

Flow metrics to monitor per spec in `.plaesy/memory/design.md`:

- **Cycle time**: Elapsed time from start to done
- **Work item age**: Age of in-progress items (leading indicator)
- **Throughput**: Items completed per time period
- **WIP count**: Current work in progress

---

**Related**: `.plaesy/instructions/quality-gates.md`, `.plaesy/instructions/error-recovery.md`
