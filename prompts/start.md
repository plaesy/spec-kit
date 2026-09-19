---
description: "Orchestrate complete project workflow automation from idea to production"
subagent: true
---

# `/start` command instructions

⚡ **Run with**: `ultracode` (multi-agent parallel execution)

## Objective
Transform project description → production-ready deliverable through modular
orchestration. The deliverable is not always code — it can be a business plan, a
legal document set, a marketing campaign, a product spec, or a mix. Detect which
before generating a constitution, since a software constitution and a business-plan
constitution enforce different things.

## Why This Phase Matters

**Purpose**: Set project direction clearly BEFORE expensive implementation work

**Why Now**: Starting strong saves debugging later (clarity at start → 30% faster delivery)

**What Breaks If Skipped**:
- Vague specifications → constant rework cycles
- Missing non-functional requirements → performance fails later
- Unclear architecture → wrong technology choices → rewrite risk

**Success Enables**:
- Confident `/implement` phase (everyone knows requirements)
- Better assessment results (clear baseline)
- Faster optimization (correct priorities identified upfront)

## Protocol
**Execute project automation by coordinating specialized prompts**
- Load memory state from `.plaesy/memory/`
- Execute phases using dedicated prompts
- Track progress and handle state transitions

### Phase 0: Constitution (One-Time, Before Anything Else)

**Detect project type first** — read the project description for the primary
deliverable: `software` (has code/app/service/API), `business` (plan, model, pitch),
`legal` (contracts, compliance, policy), `marketing` (campaign, positioning, content),
`design` (UI/UX, brand, product design), `product` (roadmap, feature spec), or `mixed`
(a software product with business/legal/marketing components alongside the code —
the common case for a real product, not an edge case). Mixed projects get one
constitution with a section per active dimension, not multiple separate files.

If `.plaesy/memory/constitution.md` does not exist yet:
1. Generate it from `.plaesy/templates/constitution.template.md`, filling defaults **per
   detected dimension** unless the project description specifies otherwise:
   - `software`: 90% coverage, <200ms, zero known vulnerabilities
   - `business`: unit-economics reviewed, assumptions logged, go/no-go criteria defined
   - `legal`: every clause/claim traceable to a source, compliance owner assigned
   - `marketing`: every claim sourced, brand voice/positioning defined
   - `design`: WCAG 2.1 AA, design tokens defined (no hardcoded values)
   - `product`: roadmap prioritization criteria, competitive baseline defined
   A `mixed` project keeps the software defaults only for the parts that are actually
   software, and adds the matching non-software defaults for the rest — never applies
   coverage/latency targets to a business plan or a legal document.
2. Ask the user only if a default would be a real judgment call (e.g. approved tech
   stack, non-negotiables) — otherwise fill defaults and let `/save` record amendments later.
3. Every later phase reads this file first; it overrides any hardcoded default elsewhere
   in these prompts. A phase must read which dimension(s) the constitution declares
   active and apply only the matching defaults — not assume `software` when the
   constitution says otherwise.

If it already exists, load it and skip regeneration.

### Phase 0.5: Ambiguity Resolution (Before Spec Generation)

Run `/assess` (Mode 1, autonomous) on the project description before writing the spec —
it covers both external research and internal ambiguity resolution. Bounded to ≤5
questions; safe defaults are assumed and labeled, hard forks with no safe default block
and ask the user once. See `/assess` Mode 1.

### Initialize Task System

Create project task infrastructure:

**Create directory structure**:
```
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
Defined once in `.plaesy/memory/constitution.md` (Phase 0), per the dimension(s)
detected there. For the `software` dimension, defaults if unset: Test Coverage ≥90%,
Performance <200ms, Security: zero known vulnerabilities. Every dimension shares one
non-negotiable regardless of type: documentation/rationale complete for all
deliverables. See `.plaesy/instructions/quality-gates.md` for the full automated gate
checklist (software-focused; other dimensions use their `assess-{dimension}.md`
mandatory-audit row instead — see `/assess` Design Spine).

### Error Handling
- Follow `.plaesy/instructions/error-recovery.md` protocol
- Retry with alternatives on failures
- Document all issues and solutions

## 📊 Progress Tracking

Format:
```
Phase: [current/11] | Quality: [score/100]
Status: [working/completed/failed] | Next: [next phase]
```

---

**Execute this orchestrator with `/start` command!**
