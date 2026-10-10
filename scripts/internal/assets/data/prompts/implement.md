---
description: "Generate complete implementation from specifications with TDD enforcement"
subagent: true
---

# `/implement` command instructions

⚡ **Run with**: any parallel-capable mode — a host that provides
subagents, or a mode that does. Not a requirement: a single-agent host
runs this identically, just serially. `/loop` documents when fan-out is
worth its cost.

🔄 **Invocation mode**: dispatched as a background task per
`.plaesy/instructions/plaesy.md` → Long-Running Commands. Control returns to
the user immediately; completion is reported, not polled. Run synchronously
only if the user explicitly asks to wait or this conversation's next step
needs the result right away.

## Usage Format

```bash

/implement                        # Implement per constitution's active dimension(s)
/implement:technical              # Code generation, TDD-enforced
/implement:design                 # Design/component implementation
/implement:business               # Business case / model drafting
/implement:marketing              # Campaign / content drafting
/implement:legal                  # Contract / policy drafting
/implement:financial              # Cost model / pricing drafting
/implement:product                # Roadmap / feature-spec drafting
/implement:management             # Process / org structure drafting
/implement:operations             # Deployment, release, runbook, IaC implementation

# Narrow to one sub-area within the technical dimension
/implement:technical --focus tdd      # Expand test suite specifically
/implement:technical --focus wcag     # Fix accessibility issues specifically
```

Same 9-dimension scope set as `/assess:{scope}` (see `/assess`) — `technical`,
`design`, `business`, `marketing`, `legal`, `financial`, `product`, `management`,
and `operations` — `--focus` narrows to a specific sub-area within the chosen
dimension; it never selects the dimension itself.

## Objective

Universal implementation orchestrator: executes specialized implementation for
the active dimension(s) declared in `.plaesy/memory/constitution.md` (see
`/start` Phase 1, sub-step 1.0 — Constitution):

- **Software dimension**: code generation (Go, Java, TypeScript, Python, Rust, etc.),
  design implementation (Figma specs, components), full-stack work — TDD-enforced (below)
- **Other dimensions** (business, legal, marketing, product, financial, management,
  operations): draft the deliverable (business case, contract clauses, campaign copy,
  roadmap, runbook, etc.) per that dimension's `assess-{dimension}.md` spine —
  Draft-Review-Revise instead of TDD (see "Non-Code Deliverable Cycle" below)

**Touching up one already-written document/deck/design file, no constitution/spec
involved?** Not implementation from a spec — use `/improve` (Artifact Mode) instead.

## Protocol

**Smart orchestration**: Detect active dimension(s) → for `software`, load
technology/tool instructions and execute with TDD; for other dimensions, load
`assess-{dimension}.md` and execute the Draft-Review-Revise cycle

## Implementation Orchestration Process

1. **Project Analysis** - Detect project type, technology stack, requirements.
   If `.plaesy/analysis/` is missing or stale, run `plaesy analyze` first (it
   skips regeneration automatically when the fingerprint is unchanged, so
   it's cheap to call unconditionally). Before writing any new function,
   check `.plaesy/analysis/project.symbols.md` — the project's function/class
   index — for an existing symbol with the same name or intent; reuse or
   extend it instead of creating a duplicate.
2. **Instruction Loading** - Load appropriate instruction files:
  - **Technology-specific**: go.instructions.md, java.instructions.md, reactjs.instructions.md, etc.
  - **Tool-specific**: terraform.instructions.md, sql.instructions.md, etc.
  - **Specialty**: tdd-enforcement.instructions.md, security-and-owasp.instructions.md, etc.
3. **Context Validation** - Verify all specs, architecture, design requirements
4. **Execution** - Execute per loaded instructions (TDD, security, tooling, etc.)
5. **Quality Verification** - All quality gates + automated validation
6. **Documentation** - Per technology best practices

## Constitutional Rules (Mandatory)

Read `.plaesy/memory/constitution.md` first — its numbers override the defaults below.
Rules below the dividing line apply only when `software` is an active dimension.

- ✅ **TRACEABLE** - Every claim/clause/number in a non-code deliverable cites its
  source (per `/assess` Web Research Requirement); every code path has a
  spec/acceptance-criterion it traces back to
- ✅ **MANDATORY AUDIT PASSES** - the dimension's audit row from `/assess`
  Design Spine (WCAG for design, unit-economics for business, clause traceability
  for legal, claim sourcing for marketing) before marking complete
- ✅ **QUALITY GATES** - Automated validation mandatory, per `.plaesy/instructions/quality-gates.md`
  (software gates) or the dimension's `assess-{dimension}.md` checklist (other dimensions)

--- *software-dimension only* ---

- ✅ **TDD REQUIRED** - Red-Green-Refactor for all code
- ✅ **COVERAGE** - Minimum test coverage enforced ({{TEST_COVERAGE_MIN|90}} if the constitution is unset)
- ✅ **MODERN TECH ONLY** - Context7 validated technologies
- ✅ **SECURITY-FIRST** - OWASP compliance built-in

## Context7 Protocol (Required)

Load `.plaesy/instructions/context7-protocol.md` before writing any code that
touches a third-party library, and follow it as written — it is the single
source for the mandatory-lookup table, tool sequence, citation form, and
fallback chain; do not restate or re-derive it here.

For `/implement`: every technology added or upgraded this run gets a
`resolve-library-id` call and a citation. Lines depending on an API signature
carry `Implementation based on Context7 (/library/id) — Retrieved {{CURRENT_DATE}}`.

## Chain-of-Thought Reasoning (BEFORE Writing Code)

**SHOW your thinking step-by-step BEFORE implementation**:

### 1. Parse Specification & Requirements

```markdown

What this requires:
- Core features: [list from spec]
- Non-functional requirements: [performance/security/scaling targets]
- Dependencies: [what already exists, what's new — cross-check
  `.plaesy/analysis/project.symbols.md` before naming a new function/class]
- Constraints: [time/resources/platform limitations]
```

### 2. Design Decision Points

```markdown

For [architectural decision], options are:
A. [Option] - Pros: [...] Cons: [...]
B. [Option] - Pros: [...] Cons: [...]
C. [Option] - Pros: [...] Cons: [...]

Choosing [A] because: [specific rationale]
If [assumption changes], reconsider [alternative]
```

### 3. Test Strategy (BEFORE Writing Code)

```markdown

Behaviors needing test coverage:
- Happy path: [feature works as specified]
- Error cases: [invalid input handling]
- Edge cases: [boundary conditions]
- Performance: [targets from spec]

Integration points to verify:
- [Component A] → [Component B]
- [API endpoint] → [Database]
```

### 4. Implementation Order (Dependency-Based)

```markdown

Build in this sequence:
1. [Foundation layer - lowest dependency]
2. [Service layer - uses foundation]
3. [API layer - uses services]
4. [Integration - wires everything]

Reason for order: [explains why dependencies matter]
```

THEN write code, guided by this reasoning above.

## Non-Code Deliverable Cycle (Non-Negotiable, Other Dimensions)

When the active dimension is `business`/`legal`/`marketing`/`product`/`financial`/
`management`/`operations`, skip TDD entirely and use this cycle instead:

1. **DRAFT** - Write the section per its `assess-{dimension}.md` spine (Components →
   Tokens → States & Edge Cases, per `/assess` Design Spine table),
   guided by the same chain-of-thought reasoning as code (parse requirements → weigh
   options with rationale → identify what needs verification → sequence by dependency)
2. **REVIEW** - Run Mode 1 Ambiguity Resolution (`/assess`) on just the
   drafted section: cite every claim/number/clause, flag unresolved forks
3. **REVISE** - Fix flagged gaps; re-run the dimension's mandatory audit row
4. **COMMIT** - Each section with a clear commit message

### Commit Pattern (Non-Code)

```markdown

DRAFT: Add unit-economics section to business case
REVIEW: Source CAC/LTV figures, flag payback-period assumption
REVISE: Correct payback period per sourced figure, resolve assumption
```

## TDD Enforcement (Non-Negotiable, Software Dimension)

### Red-Green-Refactor Cycle

1. **RED** - Write failing test first (guided by chain-of-thought above)
2. **GREEN** - Make test pass with minimal code
3. **REFACTOR** - Improve code quality
4. **COMMIT** - Each phase with specific commit messages

### Commit Pattern

```markdown

RED: Add user authentication validation test
GREEN: Implement authentication service with JWT
REFACTOR: Extract validation logic to separate module
```

## Validation Checklist (Before Running)

- ✅ Specs are clear (research done, gaps found via `/assess`)
- ✅ Project type detected (language, framework, tools identified)
- ✅ Appropriate instruction files available (tech + tools)
- ✅ Design specifications available (if frontend/UI needed)
- ✅ TDD environment ready (test framework installed)
- ✅ No blocking dependencies missing
- ✅ If the spec is a user story: work through `.plaesy/checklists/story-draft.checklist.md`
  before implementing (INVEST compliance, acceptance criteria clarity)

## Definition of Done

Before marking implementation complete, work through
`.plaesy/checklists/story-done.checklist.md` (testing, documentation, Definition of Done)
— it's the completion gate this section's own criteria below summarize.

## Technology-Specific & Tool-Specific Implementation

### Step 1: Detect Project Needs & Load Instructions

Don't hardcode a technology list or re-parse `mapping.json` yourself —
`instructions/mapping.json` lives only in the spec-kit source, not `.plaesy/`.
Detection is already implemented in `plaesy stack detect`
(`scripts/internal/detectstack/`) — reuse it:

1. Run `plaesy stack detect <project-dir>` — scans manifests/imports/spec files
   against `mapping.json`, prints one `*.instructions.md` filename per line
   (always-load files + every matched framework/language)
2. **Name transform**: strip the source-repo suffix (`go.instructions.md` →
   `go.md`) before checking `.plaesy/instructions/` — same rename `plaesy init`
   does via its `scaffold` package
3. Per detected instruction: load `.plaesy/instructions/<name>.md` if it exists;
   if missing (tech added after last `plaesy init`), copy it from source with the
   same rename, then load it — never skip
4. Multiple matches are normal — load all (e.g. a project serving a web app and
   generating PPTX reports loads both `nextjs.md` and `powerpoint.md`)
5. Something the task needs but isn't in the registry (e.g. Figma work →
   `.plaesy/roles/designer.md` + the Figma MCP server)? Load it directly — the
   registry covers per-technology instructions, not every tool reference
    - UI/frontend work → read `.plaesy/memory/design.md` first if it exists
      (project-wide tokens + rationale); reuse its tokens, never hardcode new
      colors/spacing/typography. Missing? Don't invent one inline — flag it
      and recommend `/assess:design` (Design Spine, Mode 1)
    - Before creating a UI component, check for an existing one with the
      same *purpose* (not name) — read `docs/components.md` if `/doc` has
      run (past 15 components it indexes `docs/components/{name}.md`), else
      scan the component directory directly. The tree is descriptive
      (derived from code), unlike `design.md`'s prescriptive tokens — never
      hand-maintain a separate list, it drifts. Duplicate-purpose components
      are an `/assess:design` finding, not a style nit

### Step 2: Execute Per Loaded Instructions

Follow the loaded instruction file(s) exactly. Multi-technology projects load multiple instructions:

```markdown

Example: React + Node.js + PostgreSQL + Docker
Load: reactjs.instructions.md + nestjs.instructions.md + sql.instructions.md + devops-core-principles.instructions.md
Execute in sequence per each instruction spec
```

### Step 3: Universal Constraints (All Technologies)

Regardless of technology chosen:

- ✅ TDD enforcement (per tdd-enforcement.instructions.md)
- ✅ Test coverage mandatory — meets `.plaesy/memory/constitution.md` ({{TEST_COVERAGE_MIN|90}})
- ✅ OWASP security compliance
- ✅ Modern tech stack (Context7 validated)
- ✅ Complete documentation

## Implementation Standards

- **Write the test first** (Red-Green-Refactor per the loaded instructions)
- **Meet the constitution's coverage minimum before committing** — new code arrives
  with tests
- **Verify a dependency is current via Context7** before adding it
- **Use only the spec'd tech stack** — introducing another technology is a spec
  change, not an implementation detail
- **Implement the OWASP controls the spec calls for** as part of the work
- **Follow the researched requirements exactly** — deviation is a spec change and
  gets made as one
- **Meet WCAG AA on any UI work** as part of the implementation
- **Execute per the technology-specific instructions that are loaded**
- **Load both instruction sets and coordinate the work** when design and code land
  in the same change

## Error Recovery

Test failures, build/compilation errors, and dependency conflicts during
`/implement` all have dedicated decision trees — not restated here:

→ `.plaesy/instructions/error-recovery.md` → "/implement Phase Recovery"

## Verifier Separation (Before Marking Complete)

Per `.plaesy/instructions/quality-gates.md`: the implementing agent/session does
not get the final say on its own work. Before reporting "Implementation Complete":

1. Spawn a **fresh review pass** — a new agent invocation with no implementation
   context, given only the spec, `.plaesy/memory/constitution.md`, and the diff
2. That pass checks the diff against every acceptance criterion, not the
   implementer's self-report
3. Any mismatch (missing criterion, violated constitution non-negotiable,
   claimed-but-absent behavior) is a gate failure: fix and re-verify, never
   downgrade to a "note"
4. Only after the verifier pass reports PASS does `/implement` report complete

Deliberately not self-verification — a second pass's value comes from not
sharing the first pass's blind spots.

## Deliverables

**Software dimension**:

- Complete source code, comprehensive tests (meets constitution coverage minimum,
  {{TEST_COVERAGE_MIN|90}}), documentation with examples, configuration files, deployment scripts
- Verifier pass result (PASS + findings addressed)

**Other dimensions**:

- Complete deliverable document (business case, contract, campaign brief, roadmap)
- Citation/source list for every claim, clause, or figure
- Mandatory-audit result for the dimension (per Design Spine)
- Verifier pass result (PASS + findings addressed)

## Completion Format

✅ Implementation Complete

- Software: Coverage [XX%] (must meet constitution minimum, default {{TEST_COVERAGE_MIN|90}}), Tests
  [all passing/[N] failing], Deliverables [code, tests, docs, config]
- Other dimensions: Mandatory audit [PASS/FAIL], Sourced claims [XX/XX], Deliverable
  [document name + location]
- Next phase: /assess

## Progress Format

```markdown

Status: Implementing...
Component: [current component]
TDD Phase: [RED/GREEN/REFACTOR]
Coverage: [XX%]
```

## Need Help Deciding?

When facing difficult decisions during implementation, call **`@nara`**:

- Architecture patterns (which pattern best?)
- Refactoring vs shipping (now or later?)
- Technology choices (library X vs Y?)

Nara will gather perspectives from dev, architect, product and make the decision.

## Autonomous Routing (After Implementation)

**See also**: [Universal Autonomous Routing](.plaesy/instructions/plaesy.md#-universal-autonomous-routing-global-mandatory)

After `/implement` completes, run the Phase 3 quality gate: it is automatic, and a
failure blocks and must be fixed before proceeding. On pass, `/assess` (assessment
mode — Phase 4) is MANDATORY and cannot be skipped; it gates every downstream
phase. Its findings then route onward.

**Dimension-based routing details**: `.plaesy/instructions/dimension-mapping.md` —
the canonical per-dimension table (which dimension routes to which `/assess`,
`/implement`, `/optimize`, `/fix`, `/loop` command) and the severity→priority
order for when several issue types surface at once. That file, not this prompt,
decides which command a given finding lands in.

**Role handoff**: once the role that drove this `/implement` run (e.g. `@dev`)
finishes its work, look up `.plaesy/instructions/role-mapping.md`'s Routing
Table for that role's "Typical next role" (`@dev` → `@qa`), write the handoff
to `state.json.role_handoff_log`, set `active_role` to the next role, and
invoke `/continue` rather than stopping — `/continue`'s Phase Detection cascade
picks the chain up from there. No hop limit; unless the next step is a
hard-stop category (`.plaesy/instructions/plaesy.md` rule 8), the chain keeps
going.

---

## Worked Example — Implement User Authentication (TDD)

**Input**: Spec requires JWT-based auth with login, refresh, logout. Constitution:
coverage ≥90, p95 <200ms. User runs `/implement:technical`.

```text
Status: Implementing...
Component: auth service
TDD Phase: RED
Coverage: 0%

Status: Implementing...
Component: auth service
TDD Phase: GREEN
Coverage: 67%

Status: Implementing...
Component: auth middleware
TDD Phase: REFACTOR
Coverage: 92%

✅ Implementation Complete
- Software: Coverage 92% (meets constitution minimum 90%), Tests all passing (184),
  Deliverables [code, tests, docs, config]
- Next phase: /assess
```

**Verifier pass**: Fresh agent reviewed diff against 12 acceptance criteria — PASS.
**Output**: `internal/auth/{service,middleware,handler}.go`, `internal/auth/*_test.go`,
`config/auth.yaml`, `docs/reference.md` (auth section).

---

**Follow shared protocols**: `.plaesy/instructions/quality-gates.md` → `.plaesy/instructions/error-recovery.md` →
[Global Routing](.plaesy/instructions/plaesy.md#-universal-autonomous-routing-global-mandatory)
