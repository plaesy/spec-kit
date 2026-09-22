---
description: "Generate complete implementation from specifications with TDD enforcement"
subagent: true
---

# `/implement` command instructions

⚡ **Run with**: `ultracode` (multi-agent parallel execution)

## Usage Format

```bash
/implement                        # Implement per constitution's active dimension(s)
/implement:technical              # Code generation, TDD-enforced
/implement:design                 # Design/component implementation
/implement:business               # Business case / model drafting
/implement:legal                  # Contract / policy drafting
/implement:marketing              # Campaign / content drafting

# Narrow to one sub-area within the technical dimension
/implement:technical --focus tdd      # Expand test suite specifically
/implement:technical --focus wcag     # Fix accessibility issues specifically
```

Same 8-dimension scope set as `/assess:{scope}` (see `/assess`) — `--focus`
narrows to a specific sub-area within the chosen dimension; it never selects the
dimension itself.

## Objective
Universal implementation orchestrator: Execute specialized implementation based on
the active dimension(s) declared in `.plaesy/memory/constitution.md` (see
`/start` Phase 0):
- **Software dimension**: Code generation (Go, Java, TypeScript, Python, Rust, etc.),
  design implementation (Figma design specs, component generation), full-stack
  implementation (code + design + tools) — TDD-enforced per below
- **Other dimensions** (business, legal, marketing, product, financial, management):
  Draft the deliverable (business case, contract clause set, campaign copy,
  roadmap, etc.) per that dimension's `assess-{dimension}.md` spine — Draft-Review-
  Revise cycle instead of TDD (see "Non-Code Deliverable Cycle" below)

**Revising an existing standalone artifact, no constitution/spec involved** (touch
up one already-written document/deck/design file)? That's not implementation from
a spec — use `/improve` (Artifact Mode) instead.

## Protocol
**Smart orchestration**: Detect active dimension(s) → for `software`, load
technology/tool instructions and execute with TDD; for other dimensions, load
`assess-{dimension}.md` and execute the Draft-Review-Revise cycle

## Implementation Orchestration Process
1. **Project Analysis** - Detect project type, technology stack, requirements
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
- ✅ **COVERAGE** - Minimum test coverage enforced (default 90% if constitution unset)
- ✅ **MODERN TECH ONLY** - Context7 validated technologies
- ✅ **SECURITY-FIRST** - OWASP compliance built-in

## Context7 Protocol (Required)
**Before ANY implementation**:
1. `mcp__context7__resolve-library-id` for ALL technologies
2. `mcp__context7__get-library-docs` (tokens=500)
3. **Citation**: "Implementation based on Context7 (/library/id) - Retrieved {{CURRENT_DATE}}"
4. **Error Recovery**: Alternative sources if Context7 fails

## Chain-of-Thought Reasoning (BEFORE Writing Code)

**SHOW your thinking step-by-step BEFORE implementation**:

### 1. Parse Specification & Requirements
```
What this requires:
- Core features: [list from spec]
- Non-functional requirements: [performance/security/scaling targets]
- Dependencies: [what already exists, what's new]
- Constraints: [time/resources/platform limitations]
```

### 2. Design Decision Points
```
For [architectural decision], options are:
A. [Option] - Pros: [...] Cons: [...]
B. [Option] - Pros: [...] Cons: [...]
C. [Option] - Pros: [...] Cons: [...]

Choosing [A] because: [specific rationale]
If [assumption changes], reconsider [alternative]
```

### 3. Test Strategy (BEFORE Writing Code)
```
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
```
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
`management`, skip TDD entirely and use this cycle instead:

1. **DRAFT** - Write the section per its `assess-{dimension}.md` spine (Components →
   Tokens → States & Edge Cases, per `/assess` Design Spine table),
   guided by the same chain-of-thought reasoning as code (parse requirements → weigh
   options with rationale → identify what needs verification → sequence by dependency)
2. **REVIEW** - Run Mode 1 Ambiguity Resolution (`/assess`) on just the
   drafted section: cite every claim/number/clause, flag unresolved forks
3. **REVISE** - Fix flagged gaps; re-run the dimension's mandatory audit row
4. **COMMIT** - Each section with a clear commit message

### Commit Pattern (Non-Code)
```
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
```
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

Do not rely on a hardcoded technology list and do not re-parse `mapping.json`
yourself — `mapping.json` lives only in the spec-kit source (`instructions/mapping.json`);
it is **not** copied into `.plaesy/`, so a per-project session can't read it directly.
Detection is already implemented once, in the `plaesy detect-stack` command
(`scripts/internal/detectstack/`) — reuse it instead of duplicating its logic:

1. Run `plaesy detect-stack <project-dir>` — it scans manifests, imports, and
   spec/context files against `mapping.json` and prints one
   `*.instructions.md` filename per line (always-load files plus every matched
   framework/language)
2. **Name transform**: each printed name has the source-repo suffix
   (`go.instructions.md`); the per-project copy under `.plaesy/instructions/` has
   it stripped (`go.md`) — this is the same rename `plaesy-init`'s `copy_instructions`
   does (`basename "$file" .instructions.md` → `$basename.md`). Strip the suffix
   before checking `.plaesy/instructions/`
3. For each detected instruction: if `.plaesy/instructions/<name>.md` already exists,
   load it as-is; if it's missing (tech added after the project's last `plaesy init`),
   copy it from the spec-kit source with the same rename, then load it — don't skip
   it just because it wasn't there yet
4. Multiple matches are expected and fine — load all of them (e.g. a project that
   both serves a web app and generates PPTX reports loads both `nextjs.md` and
   `powerpoint.md`)
5. If something the task clearly needs isn't in the mapping registry at all (e.g.
   Figma design work → `.plaesy/roles/designer.md` + the Figma MCP server, if
   connected), load it directly — the mapping registry covers per-technology
   instructions, not every tool reference
5a. Any task touching UI/frontend code → read `.plaesy/memory/design.md` first if
   it exists (project-wide design tokens + rationale). New components must reuse
   its tokens, not introduce new hardcoded colors/spacing/typography. If it
   doesn't exist yet, don't invent one inline — flag it and recommend
   `/assess:design` (Design Spine, Mode 1) to generate it first
5b. Before creating any new UI component, check for an existing one with the same
   *purpose* (not just the same name) — read `docs/components.md` if `/doc` has
   been run (past 15 components it's an index — follow its links to
   `docs/components/{name}.md` for detail), otherwise scan the actual component
   directory directly. The
   component tree is descriptive (derived from code), unlike `design.md`'s
   prescriptive tokens — never author a separate hand-maintained component list,
   it would just drift from the code it's supposed to describe. Duplicate-purpose
   components (same job, different name) are a design-system-adoption finding in
   `/assess:design`, not just a style nit
6. New technologies are added by editing `instructions/mapping.json` in the spec-kit
   source, never by adding another hardcoded branch here or in `detect-stack.sh`'s
   caller — this step keeps working unchanged as the registry grows

### Step 2: Execute Per Loaded Instructions

Follow the loaded instruction file(s) exactly. Multi-technology projects load multiple instructions:
```
Example: React + Node.js + PostgreSQL + Docker
Load: reactjs.instructions.md + nestjs.instructions.md + sql.instructions.md + devops-core-principles.instructions.md
Execute in sequence per each instruction spec
```

### Step 3: Universal Constraints (All Technologies)

Regardless of technology chosen:
- ✅ TDD enforcement (per tdd-enforcement.instructions.md)
- ✅ Test coverage mandatory — meets `.plaesy/memory/constitution.md` (default 90%)
- ✅ OWASP security compliance
- ✅ Modern tech stack (Context7 validated)
- ✅ Complete documentation

## Anti-Patterns (NEVER Do These)
- ❌ **Never skip TDD** - Tests written FIRST (Red-Green-Refactor per loaded instructions)
- ❌ **Never commit below the constitution's coverage minimum** - All new code must have tests
- ❌ **Never use deprecated libraries** - All tech stack verified current via Context7
- ❌ **Never mix technologies arbitrarily** - Use spec'd tech stack only
- ❌ **Never skip security implementation** - OWASP compliance mandatory
- ❌ **Never deviate from specs** - Follow research/requirements exactly
- ❌ **Never skip accessibility on UI** - WCAG AA compliance mandatory for frontend
- ❌ **Never ignore loaded instructions** - Execute per technology-specific specs
- ❌ **Never implement both design & code simultaneously without coordination** - Load both instruction sets, coordinate execution

## Error Recovery
**If tests fail:**
- Report which test failed + error message
- Do NOT skip failing tests
- Fix root cause + re-run

**If build fails:**
- Report compilation error + line number
- Fix + rebuild before proceeding

## Verifier Separation (Before Marking Complete)

Per `.plaesy/instructions/quality-gates.md`: the implementing agent/session does not get the
final say on its own work. Before reporting "Implementation Complete":

1. Spawn a **fresh review pass** — a new agent invocation with no implementation
   context, given only: the spec, `.plaesy/memory/constitution.md`, and the diff.
2. That pass checks the diff against every acceptance criterion — not against the
   implementer's self-report of what it did.
3. Any mismatch (missing criterion, constitution non-negotiable violated, claimed-but-
   absent behavior) is a gate failure: fix and re-verify, do not downgrade to a "note."
4. Only after the verifier pass reports PASS does `/implement` report complete.

This is deliberately not self-verification by the same context — the value of a second
pass comes from it not sharing the first pass's blind spots.

## Deliverables

**Software dimension**:
- Complete source code, comprehensive tests (meets constitution coverage minimum,
  default 90%), documentation with examples, configuration files, deployment scripts
- Verifier pass result (PASS + findings addressed)

**Other dimensions**:
- Complete deliverable document (business case, contract, campaign brief, roadmap)
- Citation/source list for every claim, clause, or figure
- Mandatory-audit result for the dimension (per Design Spine)
- Verifier pass result (PASS + findings addressed)

## Completion Format
✅ Implementation Complete
- Software: Coverage [XX%] (must meet constitution minimum, default ≥90%), Tests
  [all passing/[N] failing], Deliverables [code, tests, docs, config]
- Other dimensions: Mandatory audit [PASS/FAIL], Sourced claims [XX/XX], Deliverable
  [document name + location]
- Next phase: /assess

## Progress Format
```
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

**See also**: [Universal Autonomous Routing](`.plaesy/instructions/plaesy.md#-universal-autonomous-routing-global-mandatory`)

After `/implement` completes, run quality gates: a failure blocks and must be fixed
before proceeding. On pass, `/assess` (assessment mode) is MANDATORY and cannot be
skipped — it gates every downstream phase. Its findings then route further:
`/optimize` (performance/design issues), `/fix` (bugs/critical issues),
`/implement` again (gaps/new requirements), `/assess` Mode 1 (ambiguities), or
`/doc` (if ready).

**Dimension-based routing details**: `.plaesy/instructions/dimension-mapping.md` —
canonical per-dimension table. Priority when multiple issue types surface at once:
security/compliance first (immediate `/fix`), then coverage/audit gaps (`/implement`),
then unsourced claims or missing specs (`/assess:{dimension}`, Mode 1 if ambiguity).

---

**Follow shared protocols**: `.plaesy/instructions/quality-gates.md` → `.plaesy/instructions/error-recovery.md` → [Global Routing](`.plaesy/instructions/plaesy.md#-universal-autonomous-routing-global-mandatory`)