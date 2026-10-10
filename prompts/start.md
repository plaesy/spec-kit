---
description: "Orchestrate complete project workflow automation from idea to production"
subagent: true
---

# `/start` command instructions

⚡ **Run with**: any parallel-capable mode — a host that provides
subagents, or a mode that does. Not a requirement: a single-agent host
runs this identically, just serially. `/loop` documents when fan-out is
worth its cost.

## Objective

Transform project description → production-ready deliverable through modular
orchestration. The deliverable isn't always code — could be a business plan, legal
document set, marketing campaign, product spec, or a mix. Detect which before
generating a constitution: a software constitution and a business-plan constitution
enforce different things.

## Why This Phase Matters

**Purpose**: Set project direction clearly BEFORE expensive implementation work

**Why Now**: Clarity at start → 30% faster delivery, less rework

**What Breaks If Skipped**: vague specs → rework cycles; missing non-functional
requirements → performance fails later; unclear architecture → wrong tech choices →
rewrite risk

**Success Enables**: confident `/implement` (requirements known), better assessment
baselines, faster optimization (priorities correct upfront)

## Usage Format

```bash

/start                          # Detect project type, create the constitution, seed tasks
/start <project description>    # Same, with the description supplied up front
```

## Protocol

### The phase ladder

`/start` is a one-time entry point. It runs sub-step 1.0 and seeds the task
system, then hands off to `/continue`, which owns routing for the rest of the
ladder. The ladder itself is **not defined here** — the canonical 9-phase model,
its commands, and which phases are mandatory live in
`.plaesy/instructions/plaesy.md` and nowhere else. Do not restate or re-number it
in this prompt; a second phase table is how `/start` and `/continue` ended up
disagreeing about what Phase 3 is.

`/start` covers these steps before handing off:

1. Load memory state from `.plaesy/memory/`
2. Sub-step 1.0 — Constitution (below)
3. Sub-step 1.1 — Ambiguity Resolution (below)
4. Seed `.plaesy/tasks/`
5. Report the phase ladder position and hand off to `/continue`

### Sub-step 1.0: Constitution (One-Time, Before Anything Else)

**Detect project type first** — read the project description for the primary
deliverable: `software` (code/app/service/API), `business` (plan, model, pitch),
`legal` (contracts, compliance, policy), `marketing` (campaign, positioning, content),
`design` (UI/UX, brand, product design), `product` (roadmap, feature spec), or `mixed`
(software product with business/legal/marketing components alongside the code — the
common case for a real product, not an edge case). Mixed projects get one
constitution with a section per active dimension, not multiple separate files.

If `.plaesy/memory/constitution.md` does not exist yet:

1. Generate it from `.plaesy/templates/constitution.template.md`. **The template's
   §1 Active Dimensions table is the single source of the per-dimension defaults** —
   fill its `Active` and threshold cells from the detected type; do not restate or
   re-derive defaults here. Detection types: `software`, `business`, `legal`,
   `marketing`, `design`, `product`, or `mixed`. Mixed projects activate several
   dimension rows in one file, not several files.
2. A threshold never travels between dimensions: the software defaults ({{TEST_COVERAGE_MIN|90}} coverage,
   <200ms, zero Sev-High vulnerabilities) apply to the Technical row only. A
   `business`, `legal`, or `marketing` project inherits the mandatory audit from its
   own row instead — never a coverage or latency target.
3. Even on a real judgment call (e.g. approved tech stack, non-negotiables),
   apply the best-practice default, label it `ASSUMED — <default>`, and let
   `/save` record it — unless it is a hard-stop category
   (`.plaesy/instructions/plaesy.md` rule 8), which blocks instead.
4. Every later phase reads this file first; it overrides any hardcoded default
   elsewhere in these prompts. A phase must read which dimension(s) the constitution
   declares active and apply only the matching defaults — never assume `software`
   when the constitution says otherwise.
5. Run `plaesy validate constitution` after writing it — an unfilled `{{...}}`
   placeholder or an `active_dimensions`/§1 mismatch is a hard stop, not a warning.

If it already exists, load it and skip regeneration.

### Sub-step 1.1: Ambiguity Resolution (Before Spec Generation)

Run `/assess` (Mode 1, autonomous) on the project description before writing the spec —
covers both external research and internal ambiguity resolution in one pass. Bounded
to ≤5 questions; safe defaults are assumed and labeled, hard forks with no safe
default apply the best-practice choice and record it too — only a hard-stop
category (`.plaesy/instructions/plaesy.md` rule 8) blocks. See `/assess` Mode 1.

### Initialize Task System

Create project task infrastructure:

**Create directory structure**:

```text

.plaesy/tasks/
├── backlog/           # Ideas, features, bugs not yet scheduled
├── todo/              # Scheduled for current phase
├── doing/             # In active work
├── done/              # Completed and validated
└── blocked/           # Blocked on external factors or dependencies
```

**Seed initial tasks from project description** (optional):

1. Extract key work items from requirements / architecture specifications
2. Create task files in `.plaesy/tasks/backlog/`
3. Each task includes YAML frontmatter defining Status, Phase, Description, Acceptance Criteria, Priority

**Integration across phases**:

- `/continue` auto-detects tasks in `todo/` and `doing/`, can run in a loop for autonomous execution
- `/save` reports task counts per status and phase to `context.md`

## 🎯 Execution Rules

### Quality Standards

Defined once in `.plaesy/memory/constitution.md` (sub-step 1.0), per the
dimension(s) detected there. For the `software` dimension, defaults if unset: Test
Coverage ≥{{TEST_COVERAGE_MIN|90}}, Performance <200ms, Security: zero known vulnerabilities. Every
dimension shares one non-negotiable regardless of type: documentation/rationale
complete for all deliverables. See `.plaesy/instructions/quality-gates.md` for the
full automated gate checklist (software-focused; other dimensions use their
`assess-{dimension}.md` mandatory-audit row instead — see `/assess` Design Spine)
and for the remediation-floor and skip-optimize thresholds every command gates on.

## Error Recovery

**If the project type cannot be determined** from the description:

- Do not guess. A constitution generated for the wrong dimension enforces the
  wrong thresholds for the rest of the project, and every later phase reads it.
- State what the description left ambiguous, name the two or three candidate types
  with the evidence for each, and ask once.

**If `.plaesy/templates/constitution.template.md` is missing or a `{{...}}`
placeholder survives the fill**:

- `plaesy validate constitution` is a hard stop, not a warning. Report the
  placeholder by name and stop. A constitution with an unfilled cell silently
  supplies no threshold, and the phase that needs it proceeds as though it were
  met.

**If the constitution already exists**:

- Load it and skip regeneration. Never overwrite it — it is the accumulated
  amendment record, and `/save` appends to it. Report the active dimensions it
  declares, because every later phase reads them.

**If ambiguity resolution produces a hard fork with no safe default**:

- Block and ask once, naming the decision and what each option implies. Assume a
  safe default only where one is genuinely safe, and **label the assumption you
  made** so it is visible in the record.

**If a step fails**:

- Follow `.plaesy/instructions/error-recovery.md`
- Retry with alternatives on failures
- Document all issues and solutions

## 📊 Progress Tracking

Format:

```text
Phase: [current]/9 | Quality: [score]/100
Status: [working/completed/failed] | Next: [next phase]
```

The total is always **9**. Sub-steps 1.0 and 1.1 are part of Phase 1 and are not
counted separately (see `.plaesy/instructions/plaesy.md`).

## Success Criteria

`/start` is complete when:

- ✅ Project type detected, or the ambiguity that prevents it is reported and asked
- ✅ `.plaesy/memory/constitution.md` exists, is filled, and passes `plaesy validate constitution`
- ✅ Active dimensions reported, so the next phase reads the right thresholds
- ✅ `.plaesy/tasks/{backlog,todo,doing,done,blocked}/` exist
- ✅ Phase position reported against the canonical ladder, and `/continue` named as the handoff

## What Counts as Done

`/start` does **not** implement, assess, optimize, or document. It produces a
project that the rest of the ladder can run against, and stops there. An agent
that continues into `/implement` has skipped the handoff that `/continue` depends
on, and `/continue` will re-detect a project state that was never reported.

---

## Output Format

`/start` produces a project initialization report and hands off to `/continue`.

### Initialization Report

```text
START — Project Initialized
Project Type: [software | business | legal | marketing | design | product | mixed]
Active Dimensions: [list from constitution.md]
Constitution: [path to .plaesy/memory/constitution.md — validated]
Task System: [.plaesy/tasks/{backlog,todo,doing,done,blocked}/ created]
Phase Position: 1/9 (Sub-step 1.0 + 1.1 complete)
Handoff: /continue
```

**Files created**:

- `.plaesy/memory/constitution.md` — active dimensions + thresholds (validated by `plaesy validate constitution`)
- `.plaesy/tasks/{backlog,todo,doing,done,blocked}/` — task directory structure
- `.plaesy/context.md` — initial session state (phase, next steps)
- `.plaesy/memory.md` — knowledge index (empty, ready for entries)
- `.plaesy/decisions.md` — decisions index (empty, ready for entries)
- `.plaesy/decisions/` — decision topic files (created on first decision)

### If Project Type Ambiguous

```text
START — Ambiguous Project Type
Description provided: [user input]
Candidate types:
  1. [type] — evidence: [what in description suggests this]
  2. [type] — evidence: [what in description suggests this]
Decision required: which type? (reply with number or description)
```

---

## Worked Example — Start a Mixed Software+Business Project

**Input**: `/start "SaaS analytics dashboard for ops teams — Go backend, React frontend,
subscription billing, GDPR compliance"`

```text
START — Project Initialized
Project Type: mixed
Active Dimensions: [technical, business, legal, operations, product]
Constitution: .plaesy/memory/constitution.md — validated
Task System: .plaesy/tasks/{backlog,todo,doing,done,blocked}/ created
Phase Position: 1/9 (Sub-step 1.0 + 1.1 complete)
Handoff: /continue

Files created:
- .plaesy/memory/constitution.md (technical: coverage≥90 p95<200ms; business: CAC payback≤12mo;
  legal: GDPR audit; operations: RTO<1hr; product: roadmap traceability)
- .plaesy/tasks/ directories ready
- .plaesy/context.md (phase=1, next=/continue)
- .plaesy/memory.md (empty index)
- .plaesy/decisions.md (empty index)
```

**If ambiguous**: user asked "software vs mixed?" — evidence for both shown,
user chose "mixed", constitution generated with 5 active dimensions.

---

**Execute this orchestrator with `/start` command!**
