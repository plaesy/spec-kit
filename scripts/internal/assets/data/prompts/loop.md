---
description: "Autonomous workflow loop orchestrator across any declared dimension — assess, fix, verify, repeat without user input"
subagent: true
---

# `/loop` - Autonomous Continuous Improvement

⚡ **Run with**: any parallel-capable mode — see [When to Fan Out](#when-to-fan-out-and-when-to-stay-in-one-response)

## Objective

Execute an autonomous workflow loop without user intervention:

```text
ASSESS (find issues)
  → IMPLEMENT (fixes)
  → VERIFY (improvements)
  → REPEAT
```

**Philosophy**: The framework knows what to do — the workflow is self-directed, no
"what's next?" prompting needed.

There is exactly one agent-side stop, and it is a failure. `consecutive_failures
≥ max_consecutive_failures` means the loop is not improving and continuing would
compound the damage; a run that ends there has not finished the work, and
Iteration Step 4 requires a **Stop Report** rather than a completion banner.
The other end of a run is the user ending the session, which the loop does not
choose and does not get to report on. Continuing past a stop condition is a
failure; halting on one is correct behaviour, not a success.

**`/loop` is an out-of-phase driver, not a phase.** The canonical 9-phase model
(`.plaesy/instructions/workflow-phases.md`) describes the project ladder; this
file's steps are the *loop's own* per-iteration cycle, numbered `Iteration Step 0-4`
precisely so they can never be confused with a global phase number. `/loop` crosses
phases 4, 5 and 7 as it iterates — it is not itself a phase and is never counted when
asking "which phase am I on?".

---

## Dimension Resolution — Read This First

`/loop` is **dimension-agnostic** — the dimension is declared by the project's
constitution, not assumed by this file. Every rule below is dimension-neutral;
where a mechanism is software-specific (test suites, coverage, p95 latency, TDD),
the software instance is named explicitly as *one* variant among the nine.

| Term | Meaning |
| --- | --- |
| **Dimension** | One of `technical`, `design`, `business`, `marketing`, `legal`, `financial`, `product`, `management`, `operations` — the same set `/assess:{scope}` uses. |
| **Software instance** | The dimension `technical`. It is the most commonly walked-through instance and is used as the worked example in several places. It is **not** the default. |
| **Deliverable** | Whatever the active dimension produces: source code, a spec, a contract, a campaign, a budget, a runbook. "Fix" always means *improve the deliverable*, never "edit code". |
| **Enforcement mode** | How a change must be justified before it is accepted. Software: TDD + passing suite. Other dimensions: sourcing/traceability, Mode 1 ambiguity resolution, restating the assumption register. |

If the constitution declares no active dimension, fall back to `technical` **and say
so in the first iteration's output** — a fallback is a fact to report, not a silent
default.

---

## When to Use `/loop`

- **During production**: run in background; it repairs issues autonomously while other work proceeds
- **Continuous improvement**: keep running until the quality threshold is reached *and then widened past it*
- **Unattended execution**: walk away; the framework keeps improving
- **Post-delivery**: auto-optimize after `/implement` (engineering) or after the equivalent handover (business, legal, operations)

**Do NOT use `/loop` when**: the next move needs a human decision that a default
answer would destroy, or the user must review each phase's output.

---

## Autonomous Loop Protocol

### Iteration Step 0: Load Loop Config (`.plaesy/state.json`)

Read `.plaesy/state.json` (created by `plaesy init` from `.plaesy/templates/state.template.json`)
for this run's controls — do not re-derive these from narrative defaults if the file
has them:

- `autonomous_loop.max_iterations` (a progress readout, not a budget — reaching it
  resets the counter rather than ending the run), `.current_iteration`,
  `.consecutive_failures` / `.max_consecutive_failures` (default 2), `.status`
  (`paused`/`running`), `.checkpoint_interval` (the template ships `0` = no scheduled checkpoints; set it to a positive integer to enable them), `.last_human_checkpoint`
- `quality_gates.*` — which gates are mandatory for this project (code review, tests,
  docs, constitutional compliance, and their non-software equivalents: sourcing audit,
  traceability check, claim review)

Update `current_iteration`, `consecutive_failures`, and `status` in this file after
every iteration (see Iteration Step 4) — this is the loop's actual state, not just
a narrative log. If the file is missing or empty (pre-existing project from before this
file existed), fall back to the defaults above and note that in the first iteration's output.

### Iteration Step 1: Intelligent Assessment (No User Input)

**Execution**:

1. Load project state (`.plaesy/context.md`, `.plaesy/memory.md`, `.plaesy/decisions.md`, `.plaesy/memory/constitution.md`)
2. Read the active dimension(s) from the constitution and run `/assess:{dimension}`
  for each active one (`/loop:{dimension}` overrides this to a single dimension —
  see Loop Configuration below). Default to `technical` only when the constitution
  declares no other dimension, and report that fallback.
3. Parse findings: categorize by severity & type
4. Filter findings:
  - ✅ **Auto-fixable** — anything mechanical and provable within the active
     dimension's enforcement mode. Examples: software — code quality, simple bugs,
     tests, docs. Writing — sourcing a missing citation, filling an ASSUMED-labeled
     default, fixing a broken cross-reference, updating a stale figure. Operations — a
     runbook step naming a command that no longer exists. Financial — a figure that
     disagrees with the assumption register it cites.
  - ⚠️ **Needs review** — anything whose correction changes a committed judgment.
     Examples: software — architecture, security, UX. Business/marketing/legal —
     unit-economics assumptions, legal clause interpretation, brand positioning calls.
  - ⏸️ **Blocked** — external dependencies, user decisions, missing inputs no agent can supply

**Output**: List of fixable issues ranked by impact, each tagged with its dimension.

### Iteration Step 2: Intelligent Implementation (Parallel Execution)

**For each auto-fixable issue**:

1. Load the appropriate instruction file — the active dimension's `assess-{dimension}.md`,
  or a tool-specific instruction when the dimension is `technical`
2. Execute the fix under the dimension's enforcement mode — software: with TDD
  enforcement. Other dimensions: revise the deliverable section, then re-run Mode 1
  ambiguity resolution on just that section
3. Validate the fix does not break existing verification (software: the suite passes;
  other dimensions: no new uncited claim and no broken cross-reference)
4. Commit with a clear message that names the dimension and the finding it closes

**Parallel execution** (if multiple independent fixes):

- Fix #1: Compliance gap → Re-run traceability check → Commit
- Fix #2: Unit-economics model error → Re-run sensitivity check → Commit
- Fix #3: Uncited claim → Attach source → Commit
- (All run in parallel, merge results)

#### When to Fan Out, and When to Stay in One Response

Modern models under-fan more often than they over-fan, so both edges are stated here.

Fan out into parallel subagents when the work splits into genuinely independent
items — separate files, scopes, dimensions, or several files to read before deciding.
Start them in the same turn; sequential spawns serialise the work for no gain.

Work directly, without a subagent, when the task fits in a single response — editing
a section already in view, a question about a file already read, a mechanical rename
across a known list. Spawning costs more than it returns when the task is already in
front of you — the subagent just rebuilds context you already have.

**Order of execution** — by impact within the active dimension:

1. **Compliance, safety, or legal exposure** (highest priority — non-negotiable regardless of other priorities)
2. **Blocking failures** (software: test failures that break the build; other dimensions: an unsourced or contradictory claim the rest of the deliverable depends on)
3. **Non-negotiable thresholds declared by the constitution** (software: p95 latency <200ms; business: CAC payback period; operations: recovery-time objective)
4. **Quality and duplication** (software: complexity, repetition; writing: restated sections, contradictions)
5. **Documentation and traceability gaps**

### Iteration Step 3: Verification (Automated Validation)

**After each fix**, run the checks the active dimension's constitution row makes
mandatory:

1. The dimension's full verification suite — software: the full test suite; writing and
  legal: the sourcing and cross-reference audit; operations: the runbook's own checks
2. Confirm the mandatory threshold still holds — software: coverage ≥{{TEST_COVERAGE_MIN|90}}; other
  dimensions: the constitution's mandatory-audit row
3. Check the dimension's threshold metrics — software: p95 latency, error rate; business:
  conversion, margin; operations: time-to-recover
4. Validate against `.plaesy/instructions/quality-gates.md`
5. Compare before/after metrics

**If verification fails**:

- Rollback the fix
- Move (or create) the task file to `.plaesy/tasks/backlog/` with the failure noted in
  its frontmatter — this is the only backlog; there is no separate `backlog.md`
- Mark as "needs review"
- Continue to the next issue

**If verification passes**:

- Keep the fix
- Update metrics
- Log to the progress tracker
- Continue the loop

### Iteration Step 4: Loop Control (When to Continue vs Stop)

**After every iteration**, update `.plaesy/state.json`:

- Increment `autonomous_loop.current_iteration`
- Reset `consecutive_failures` to 0 on a successful iteration, increment it on failure
- Set `status: "running"` while looping, `"paused"` on stop (see below)
- Set `last_human_checkpoint` when `checkpoint_interval` is a number greater than
  0 **and** `current_iteration % checkpoint_interval == 0`. A `checkpoint_interval`
  of `0` is not a modulo of zero — it means "no scheduled checkpoints"; dividing by
  it is a crash, not a checkpoint.
- When a role-scoped fix/optimize finishes within this iteration, look up
  `.plaesy/instructions/role-mapping.md`'s Routing Table for that role's
  "Typical next role," write the handoff to `role_handoff_log`, and set
  `active_role` to the next role for the following iteration — unless the next
  step is a hard-stop category (`plaesy.md` rule 8), which halts the chain and
  clears `active_role` instead.

**The loop continues by default. Stopping takes a positive reason.**

Continue unless one of the stop conditions below actually holds. In particular,
none of these end a run:

- **Every assessable issue was fixed.** That is the signal to pick a *wider* axis,
  not to stop. See [When the Checklist Is Empty](#when-the-checklist-is-empty).
- **The quality target was reached.** A score that stops rising is a plateau, and a
  plateau means the current measurement stopped being informative — change what is
  being measured (see [Plateaus](#plateaus-and-stale-metrics)).
- **`current_iteration` reached `max_iterations`.** The counter is a progress
  readout, not a budget. Reset it and carry on.
- **The remaining items need a human decision.** Gather the data and the
  best-practice recommendation, apply the recommendation, and record what was
  applied — do not park the run waiting for an answer, and do not treat an
  awkward decision as a reason to end. If applying a default would be
  destructive, take the non-destructive option that still makes progress and
  record the decision as outstanding.

**Continue loop IF**:

- ✅ The user has not interrupted the session
- ✅ `consecutive_failures < max_consecutive_failures` (state.json, default 2)
- ✅ Verification is green — if a fix broke the build or the audit, that is a failure to repair before anything else, not a reason to end the run

**Stop loop IF** — this, and only this:

- ❌ `consecutive_failures ≥ max_consecutive_failures` — the loop is failing, not
  improving, and continuing would compound the damage. Report what failed.

The user stopping a loop is not a command the agent can observe. There is no
`/stop`: the only human stop is the user interrupting the session, which arrives
as the run being cut off rather than as a message to act on. That is why Step 4
writes `state.json` on **every** iteration, not only on a scheduled checkpoint —
the interruption is unannounced, so the last written state is the only record of
where the loop got to. A loop that checkpoints on a schedule has a window in
which an interruption loses everything since the last checkpoint.

When the loop stops on the failure condition, say which condition fired and
what state it left behind. Emit a **Stop Report** in the format below — never a
completion banner, because a run that ends on a stop condition has not
completed the work.

> **On running forever.** A long-running loop is the intended mode, so exhaustion
> and completion are not stop reasons. What still halts a run is a real failure
> (`consecutive_failures`) and the user ending the session. A loop that appears to
> spin is usually not looping at all — it is re-measuring the same axis and
> reporting the same number. The cure for that is a new axis ([When the Checklist is
> Empty](#when-the-checklist-is-empty)), not a smaller cap.

---

## When the Checklist Is Empty

An empty checklist means the *current* axis is exhausted, not that the work is
done. Widen in this order, and record which axis is now in use so the next
iteration does not re-widen from the same place:

1. **Re-run the mechanical axes.** Cheap, provable, catches what survives a rewrite:
  dead links, stale counts, unresolved paths, duplicated blocks, contradicting
  headings. A green board that was never re-measured is not a green board.
2. **Verify the guards still bite.** Mutate the thing each guard watches. A guard
  never seen red is an assumption, not a check.
3. **Read what no sweep can reach.** Whether a deliverable is actionable without
  guessing, whether documents agree on *meaning* not just links, whether the
  abstraction sits at the right level — needs a reader, not a script, which is why
  it's the axis most often skipped.
4. **Compare against external best practice, not internal consistency.** A corpus
  can be self-consistent and still built on a practice the field has moved past.
  Pick one recognised standard, check against it explicitly, report the gap.
5. **Audit the tooling that judges the work.** A metric that stops discriminating
  good from bad work has stopped being a measurement.

If an axis genuinely has no remaining work, move to the next one. Only stop when
the user stops.

## Plateaus and Stale Metrics

When the score stops moving for two consecutive iterations, the number has stopped
measuring quality. Three responses, in order:

- **Re-measure from the tree, not the last run.** Derive counts from the filesystem
  — or re-read the deliverable and re-run the audit — rather than incrementing a
  stored value.
- **Change the axis.** See [When the Checklist Is Empty](#when-the-checklist-is-empty).
- **Discard the metric.** A measure that can't tell good work from bad is worse than
  none, because it's trusted. Replacing it is a real improvement even when the
  headline number doesn't move.

Never ratchet a metric that has not been shown to discriminate. A ceiling set
from an uncalibrated measurement just freezes the error.

---

## Autonomous Decision Rules

**When `/loop` encounters choices**, apply these rules (NO USER INPUT NEEDED):

These rules are written to be dimension-neutral. Where the substance differs, the
dimension's variant is named under the rule — same priority order, different content.

### 1. Foundational / Settled Choices

**Default (any dimension)**: extend what the project already committed to — do not replace or relitigate it
**Software instance**: use the tech stack already in the project; extend existing, don't replace
**Other dimensions**: use the positioning, pricing model, or legal framework already committed in the spec
**Example**: Project uses React → extend React, don't suggest Vue. The plan already commits to a subscription model → extend it, don't re-open freemium-vs-subscription

### 2. Architecture / Structure

**Default (any dimension)**: preserve the existing structure — minimal reorganization, maximum compatibility with what stakeholders already reviewed
**Software instance**: preserve existing architecture — minimal refactor, maximum compatibility
**Other dimensions**: preserve the document or campaign structure already reviewed
**Example**: Project has a monolith → add features to it. The contract is already structured by clause type → add new terms to the matching clause, don't restructure

### 3. Threshold Violations vs New Content

**Default (any dimension)**: fix threshold violations first — whatever the constitution declares non-negotiable for this dimension
**Software instance**: performance first — the constitution performance target is non-negotiable (default <200ms)
**Business**: unit-economics viability gaps first (e.g. CAC payback period)
**Legal**: compliance gaps first — non-negotiable regardless of other priorities
**Example**: Feature request vs latency bug → fix latency first. New copy vs an unsourced pricing claim → fix the claim first

### 4. Quality vs Speed

**Default (any dimension)**: accuracy wins if a quick fix exists — an unverified claim or an untraceable clause compounds risk the way technical debt does
**Software instance**: code quality wins if the refactor is under a week
**Example**: Duplicate code → extract even if it slows the current phase. A claim duplicated uncited across two sections → fix both, even if it slows the current draft

### 5. Verification Strategy

**Default (any dimension)**: the constitution's mandatory-audit row is non-negotiable — verify sourcing and traceability before drafting more
**Software instance**: the coverage minimum is non-negotiable ({{TEST_COVERAGE_MIN|90}}) — tests first (TDD), then code
**Example**: Gap in tests → write the tests before fixing the bug. Gap in citations → source the claim before adding more content that depends on it

### 6. Documentation

**Default (any dimension)**: auto-generate where possible, manual review where needed — follow the `/assess` Design Spine handoff spec (business case, ADRs, runbook, decision log)
**Software instance**: README + API docs + architecture narrative
**Example**: Code changes → auto-update related docs. Business model changes → auto-update the business case's assumption register

---

## Loop Iteration Example — Software Instance (Dimension `technical`)

### Iteration 1: Assessment

```text
ASSESS Output (dimension: technical)
- Security: 1 SQL injection risk (HIGH) ✅ Auto-fixable
- Tests: 15% coverage gap (HIGH) ✅ Auto-fixable
- Performance: 800ms query (HIGH) ✅ Auto-fixable
- Design: Missing dark mode (LOW) ⚠️ Needs review
- Docs: Missing API examples (LOW) ✅ Auto-fixable

Auto-fixable count: 4/5
Proceeding to fixes...
```

### Iteration 2: Implementation (Parallel)

```text
Fix 1: SQL injection (use parameterized queries)
Fix 2: Coverage gap (write missing tests)
Fix 3: Query performance (add index + caching)
Fix 4: API docs (generate from code)

Status: All 4 fixes in progress...
```

### Iteration 3: Verification

```text
Verification Results:
✅ Fix 1: Security verified (no SQL injection)
✅ Fix 2: Coverage now 92% (from 75%)
✅ Fix 3: Query now 45ms (from 800ms)
✅ Fix 4: API docs auto-generated

Quality score: 82/100 (was 68/100) ⬆️ +14 points
Proceeding to next assessment...
```

### Iteration 4: Re-assessment

```text
ASSESS Output:
- No high-severity issues remaining
- Design improvements identified (LOW priority)
- Performance optimizations available (MEDIUM)
- Documentation complete

Quality score: 82/100
Next axis: the mechanical guards were green, so widening to the
  read-and-judge axis (what no sweep can reach).

Loop Status: RUNNING
Recommendation: none needed — continuing on the next axis.
```

## Loop Iteration Example — Legal Instance (Dimension `legal`)

### Iteration 1: Assessment

```text
ASSESS Output (dimension: legal)
- Liability: indemnity clause absent from the MSA (HIGH) ⚠️ Hard stop
  (`.plaesy/instructions/plaesy.md` rule 8 — legal commitment) — not
  auto-fixable, not "needs review as a judgment call": this category always
  blocks, independent of how confident a default would be
- Sourcing: churn claim in §4.2 has no citation (HIGH) ✅ Auto-fixable
- Cross-reference: §7.1 points to a clause removed in the v3 restructure (MEDIUM) ✅ Auto-fixable
- Consistency: notice period stated as 30d in §3 and 45d in §9 (MEDIUM) ✅ Auto-fixable
- Style: two defined terms used before definition (LOW) ✅ Auto-fixable

Auto-fixable count: 4/5
Proceeding to fixes...
```

### Iteration 2–3: Implementation and Verification

```text
Fix 1: Attach authority to the churn claim in §4.2
Fix 2: Repoint §7.1 to the surviving clause in v3
Fix 3: Reconcile notice period to 30d, with a documented reason
Fix 4: Move both defined terms above first use

Verification Results (mandatory-audit row):
✅ Fix 1: claim traceable, Mode 1 ambiguity re-run on §4.2 only
✅ Fix 2: all 41 cross-references resolve
✅ Fix 3: no remaining contradiction across the document
✅ Fix 4: glossary check clean

Quality score: 79/100 (was 71/100) ⬆️ +8 points
The indemnity clause hit a rule-8 hard stop (legal commitment) and was
recorded in `.plaesy/decisions.md` as blocked pending a human legal call —
consistent with the policy, not an exception to it. Everything else in this
dimension is default-and-record; this one category never is.
```

---

## Detecting a Stalled Loop

**A stall is a signal to change axis, not to stop.** These three patterns mean the
loop is re-deriving the same answer:

- The same issue recurs across 3+ consecutive iterations
- The quality score has not moved for 2 iterations
- The same fix has been applied twice

**Action**:

1. Name the loop explicitly in the output: `STALLED — <axis> produced no change`.
  Say which axis, so the next iteration starts somewhere else.
2. Do not re-apply the fix. Record it as attempted and ineffective.
3. Ask what made it ineffective — a wrong diagnosis, a wrong axis, or a check that
  cannot see the defect. The third is the common one and needs its own answer.
4. Move to a different axis ([When the Checklist Is
  Empty](#when-the-checklist-is-empty)) and continue.
5. Only pause if the stall is caused by real failures rather than by repetition —
  that is `consecutive_failures`, and it is a different condition.

**Example**:

```text
STALLED — sourcing axis produced no change for 2 iterations
- Issue: "uncited claims above threshold" recurring
- Cause: new sections keep landing without a source step in the draft checklist
- Diagnosis: the audit is not the bottleneck — the missing piece is a hook that runs
  the citation check as part of drafting, so an audit that only reports the gap will
  recur forever

Action: switched to the enforcement axis. Continuing.
```

---

## Progress Tracking

**Real-time output format** (emitted every iteration):

```text
🔄 LOOP ITERATION: 3  (unlimited)
├─ Dimension: legal
├─ Status: ASSESSING
├─ Axis: cross-reference resolution  (next: meaning-level consistency)
├─ Quality Score: 79/100 (⬆️ +8 from baseline)
├─ Issues Fixed: 4
├─ Verification: mandatory-audit row green
└─ Next: Verify & assess again
```

**Stop Report** — emitted only when a Stop loop IF condition actually fires, never as
a completion banner:

```text
⏹️ LOOP STOPPED
├─ Reason: consecutive_failures reached 2 (the loop was failing, not improving)
├─ Dimension: technical
├─ Iterations completed: 5
├─ Issues Fixed: 12
├─ Last failure: audit harness could not reach the network; 3 fixes unverified
├─ Left behind: 3 fixes rolled back, tasks filed in .plaesy/tasks/backlog/
└─ Resume: re-run /loop once the dependency is restored
```

---

## Output Format

`/loop` emits a real-time progress line every iteration and a Stop Report
when a stop condition fires. It does not emit a completion banner — a run that
ends on a stop condition has not completed the work.

### Iteration Progress (every iteration)

```text
🔄 LOOP ITERATION: [N]  (unlimited)
├─ Dimension: [active dimension]
├─ Status: [ASSESSING | IMPLEMENTING | VERIFYING]
├─ Axis: [current axis name]  (next: [next axis name])
├─ Quality Score: [NN]/100 (⬆️/⬇️ [delta] from baseline)
├─ Issues Fixed: [count]
├─ Verification: [mandatory-audit row green | red — what failed]
└─ Next: [Verify & assess again | rollback & continue]
```

### Stop Report (only when `consecutive_failures ≥ max_consecutive_failures`)

```text
⏹️ LOOP STOPPED
├─ Reason: consecutive_failures reached [N] (the loop was failing, not improving)
├─ Dimension: [dimension]
├─ Iterations completed: [N]
├─ Issues Fixed: [N]
├─ Last failure: [description of what failed]
├─ Left behind: [N] fixes rolled back, tasks filed in .plaesy/tasks/backlog/
└─ Resume: re-run /loop once the dependency is restored
```

**State persistence** (every iteration):

- `.plaesy/state.json` — `current_iteration`, `consecutive_failures`, `status`, `last_human_checkpoint`

---

## Loop Configuration

**Customize loop behavior** (optional — overrides `.plaesy/state.json` for this run
only; the override is **not** written back to the file):

```markdown
/loop:technical                # Loop a single dimension (technical/design/business/
                                # marketing/legal/financial/product/management/operations;
                                # default: all dimensions active in the constitution)
/loop --max-iterations 15      # Progress readout width, NOT a stop condition — the counter
                                # resets at 15 and the run continues (default: state.json,
                                # else 10; `--max-iterations 0` means unlimited, same as null)
/loop --focus <sub-area>       # Narrows within the chosen dimension, exactly as
                                # --focus does in /fix and /optimize. Stays --flag:
                                # it is a run parameter, not a scope selector
/loop --quality-target 90      # The score the loop aims at; reaching it widens the axis,
                                # it does not end the run (default: the remediation floor,
                                # {{QUALITY_REMEDIATION_FLOOR}} — 85. Both thresholds are
                                # defined once in .plaesy/instructions/quality-gates.md)
/loop --auto-commit            # Commit after each fix (default: batch at end)
/loop --parallel 4             # Run up to 4 independent fixes in parallel (default: sequential)
/loop --no-verify              # Skip verification (NOT recommended)
```

Same 9-dimension scope set as `/assess:{scope}` (see `/assess`) — `technical`,
`design`, `business`, `marketing`, `legal`, `financial`, `product`, `management`,
and `operations` — the other flags above are run parameters, not scope selectors,
so they stay `--flag`. Which command a dimension's auto-fixable findings land in is
decided by `.plaesy/instructions/dimension-mapping.md` (the canonical routing
table), not by the order-of-execution list in Iteration Step 2.

---

## Integration with Other Commands

**`/loop` workflow pattern**:

```bash
/start Define X                # Initial setup
/continue                      # Execute the work
/loop                          # Auto-improve until the user stops
/optimize                      # Manual optimization decisions
/loop                          # Auto-verify improvements
/save                          # Final checkpoint
```

**Stopping the loop**:

```bash
# The only human stop is ending the session — interrupt the run.
# There is no /stop command: the agent cannot observe a message it will
# never receive, and a command that "gracefully stops" a run the user
# has already killed is not graceful, it is a report written too late.

# Because the interruption is unannounced, the loop writes state.json every
# iteration (Step 4). Re-running /loop resumes from the last iteration it
# reached; it does not start over.

# The loop stops itself on exactly one condition:
- consecutive_failures reaches max_consecutive_failures
```

---

## Success Criteria

The loop is working when:

- ✅ The quality score improves across iterations, or the axis is widened when it plateaus
- ✅ No user prompts or questions are raised
- ✅ Verification stays green (software: tests keep passing; other dimensions: the mandatory-audit row stays satisfied)
- ✅ Threshold metrics improve or hold (software: performance; other dimensions: their own declared thresholds)
- ✅ Progress is legible every iteration — dimension, axis, score, delta
- ✅ It runs until a real stop condition or the user ends the session, and reports which one fired
- ✅ Every fix is documented and traceable to the finding it closed

---

## Error Recovery

**If the loop encounters an error**:

1. Log the full error with context
2. Roll back the last change
3. Mark the issue as "blocked"
4. Try the next issue
5. Continue the loop

**If a critical error breaks the build or the mandatory audit**:

1. Roll back all recent commits
2. Pause the loop (`status: "paused"`)
3. Emit a Stop Report naming the failed condition
4. Recommend a manual `/fix` run before resuming

---

**Follow shared protocols**: `.plaesy/instructions/quality-gates.md` → `.plaesy/instructions/error-recovery.md`

**Usage**: Run `/loop` after `/implement` to auto-improve until a real stop
condition fires or you end the session.

**The rule that overrides every other line here**: keep going. An empty checklist, a
reached quality target, a full iteration counter, and a fix that needs a human call are
all reasons to *widen the axis*, not to end the run. The run ends on exactly one agent-side
condition — `consecutive_failures ≥ max_consecutive_failures` — or when the user ends
the session. Nothing else.
