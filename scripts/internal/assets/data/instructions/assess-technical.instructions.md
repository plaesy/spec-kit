---
description: "Technical assessment workflow - code quality, architecture, tests, security, performance"
applyTo: "**/*"
---

# Technical Assessment Instructions

**Use with**: `/assess` when dimension is **Technical**

**Referenced from**: `/assess` (orchestrator)

---

**Reference checklists**: `.plaesy/checklists/sa.checklist.md` (architecture, project
setup, dependencies, tech choices, security, performance, deployment) and
`.plaesy/checklists/qa.checklist.md` (quality gates, testing strategy, security/OWASP,
deployment validation, sign-off) — use alongside this file's scoring dimensions.

**Delegates to**: `.plaesy/roles/dev.md`, `.plaesy/roles/devsecops.md` unconditionally.
Also, only when the trigger holds: `.plaesy/roles/qa.md` (test strategy/coverage is a
finding), `.plaesy/roles/tw.md` (documentation accuracy is a finding),
`.plaesy/roles/data-engineer.md` (pipelines, ETL/ELT, warehouse, or streaming),
`.plaesy/roles/ai-architect.md` (AI/ML in the design or in production),
`.plaesy/roles/mlops.md` (models trained, served, versioned, or monitored). A role
that loads and finds nothing in scope reports "no findings" — it does not
manufacture a finding to justify having been called.

## Assessment Workflow

### Step 1: Project Analysis

- Detect language + framework (from file patterns)
- Identify technology stack
- Map project structure
- Detect frontend/UI presence: Check for React, Vue, Angular, Flutter, mobile UI, or Figma files

### Step 2: Test Execution

- Run full test suite
- Capture coverage metrics
- Record failure messages (if any)

### Step 3: Code Quality Review

- Check for: unused code, complexity hotspots, naming clarity
- Identify: n+1 queries, missing error handling, security issues (OWASP Top 10 High/Critical only)
- Score findings against the relevant `*-design-principles.instructions.md` file(s)
  for the code being reviewed (see `instructions/mapping.json` for the full set —
  e.g. [[software-design-principles]] for SOLID/coupling/duplication findings,
  [[database-design-principles]] for schema/query findings,
  [[concurrency-design-principles]] for race/lock issues,
  [[security-design-principles]] for architectural findings vs [[security-and-owasp]]
  for implementation-level ones) — name which principle a finding violates, not just
  that something looks wrong
- Deployment, release, environment parity, rollback, IaC, and observability findings
  are **out of scope here** — they belong to the `operations` dimension (see
  `.plaesy/instructions/assess-operations.md` and `.plaesy/instructions/dimension-mapping.md`). Report them as a
  handoff, never as a scored finding: a finding is counted once, in the dimension
  that owns it.
- Rate: 0-100 based on findings

### Step 4: Hand Off to Operations

Technical assessment does not score infrastructure or deployment. When the audit
finds any of the following, record it under a `## Handoff: operations` heading in
the report and continue with the remaining steps — do not rate it, and do not
fold it into another category's score:

- Pipeline stages, promotion between environments, environment parity
- Deployment reliability, rollback capability, release and deprecation process
- Monitoring, alerting, observability coverage
- Disaster recovery, backup/restore, high-availability topology
- Infrastructure as code, environment reproducibility, secrets management in
  the deployed environment

Each handoff line names the finding, its evidence, and its severity. The
`operations` assessment consumes them; it does not re-derive them.

### Step 5: Security Assessment

- Flag OWASP Top 10 High/Critical vulnerabilities only (not Low/Medium)
- No speculative security concerns
- Source: actual code patterns + known vulnerabilities

### Step 6: Documentation Review

- Check: README completeness, API docs accuracy, code comments clarity
- Flag: outdated/missing sections only (don't invent missing content)

### Uncertainty Surfacing (Before Finalizing Assessment)

Load `.plaesy/instructions/uncertainty-surfacing.md`; do not restate it here.
Only this dimension's specifics follow.

**Confidence basis for this dimension**:

- HIGH: source read, test run, or metric measured — the code or metric itself was inspected
- MEDIUM: pattern match, or inference from a comparable module rather than this one
- LOW: could not examine the area at all — no access, no test, no specification

**Worked example**:

```text
I could not assess the async retry path because the worker deployment
manifest is not in the repository.
To improve confidence, provide the deployed worker manifest, or a run of the
retry integration test.
```

### Step 7: Design Quality Review (Conditional — only if frontend/UI detected)

**When to run**: Project has React, Vue, Angular, Flutter, web UI, mobile UI, or Figma files

- **Component Consistency**: Check if UI components follow naming conventions, variants documented
- **Accessibility**: Audit WCAG 2.1 AA compliance (contrast ratios, keyboard navigation, ARIA labels)
- **Design Tokens**: Verify colors, spacing, typography defined as tokens (not hardcoded)
- **Dark Mode Support**: Check if light/dark theme colors are consistent and accessible
- **Design System Adoption**: Measure component reuse rate from design library
- **Design Documentation**: Check if design spec exists and is current
- **Figma Organization**: If using Figma, check file structure, component library, Code Connect mappings

**Rate**: 0-100 based on findings (applied ONLY if frontend/UI exists)

### Step 8: Generate Report & Next Phase Decision

**Report sections**: canonical names per `.plaesy/instructions/dimension-mapping.md`
→ *Assessment Report Section Registry*; the bullets below go *inside* those
sections and do not rename them.

- Overall score (0-100)
- Per-category scores (Quality, Tests, Security, Docs, Performance, [+ Design if frontend])
- Top 5 findings (ordered by impact)
- `## Handoff: operations` section, if any operations findings were identified

**Next Phase Decision Tree**:

- **If design quality is in the `C` or `D` band (UI projects)** → Load `.plaesy/instructions/assess-design.md`
- **If security/critical bugs found** → Run `/fix` first, then `/optimize`
- **If operations findings were handed off** → Run `/assess:operations` before scaling
- **If only performance issues** → Run `/optimize`
- **If all scores ≥ {{QUALITY_REMEDIATION_FLOOR}}** → Can proceed to `/optimize`

---

## Quality Scoring

### Backend/API Projects

```text

code_quality: 30%           # Structure, readability, patterns
test_coverage: 24%          # Coverage %, pass rate
security: 23%               # Vulnerabilities (High/Critical only)
documentation: 13%          # Completeness + accuracy
performance: 10%            # Response time, resource usage
```

Weights exclude infrastructure and deployment; those are scored by the
`operations` dimension.

### Frontend/UI Projects

```text

code_quality: 25%           # Structure, readability, patterns
test_coverage: 21%          # Coverage %, pass rate
security: 17%               # Vulnerabilities (High/Critical only)
documentation: 14%          # Completeness + accuracy
design_quality: 14%         # Component consistency, accessibility, design tokens, dark mode
performance: 9%             # Response time, resource usage, animation performance
```

### Design Quality Breakdown (if frontend/UI exists)

```text

component_consistency: 25%   # Naming conventions, variants, documentation
accessibility: 35%           # WCAG 2.1 AA compliance, contrast, keyboard nav, ARIA
design_tokens: 20%          # Token usage, consistency, light/dark themes
design_system_adoption: 20% # Component reuse rate, library usage
```

### Grade Mapping

| Grade | What it means here |
|---|---|
| A+ | Exceptional - ship as-is |
| A | Excellent - minor improvements suggested |
| B+ | Very Good - should fix before shipping |
| B | Good - address issues in next iteration |
| C | Acceptable - significant improvements needed |
| D | Needs Work - blocker issues present |

The score edges behind these letters, and the general meaning of each band,
are defined once in `.plaesy/instructions/quality-gates.md` →
**Grade bands**. Do not restate them here.

### Success Criteria (Hard Stops)

- Must have: Score computed from the published weights, with an evidence line for every criterion
- Must have: No OWASP High/Critical vulns unfixed
- Must have: Tests pass (or known failures documented)
- Must have (if frontend): WCAG 2.1 AA compliance or documented accessibility exceptions

---

## Critical Rules

- ✅ **NO SPECULATION** - Only report what code/design actually does
- ✅ **SECURITY: High/Critical Only** - Ignore OWASP Low/Medium findings
- ✅ **CITE WITH LOCATION** - Every finding includes file:line reference
- ✅ **MEASURE DON'T ESTIMATE** - Use actual test results, coverage %, metrics
- ✅ **ACCURACY > COMPLETENESS** - Report fewer findings with high confidence
- ✅ **DESIGN ONLY IF FRONTEND** - Only assess design if project has UI

---

## Finding Standards

- **Ground every finding in actual analysis** of this codebase
- **Report only architectural problems you can point at** in the code
- **Scope the security finding set to High and Critical** — Low and Medium are out
  of scope for this assessment
- **Compute the score from the formula**, with no manual adjustment
- **Attribute findings to code in this repository**, not to external factors
- **Assess what exists**; recommendations belong to `/optimize:technical`
- **Assess design only where a frontend/UI surface exists**
- **Audit the actual design files, components and UI code** that exist
- **Run the WCAG audit on any frontend** — accessibility is non-negotiable for UI
  work

---

## Pre-Assessment Validation

Run these checks BEFORE attempting assessment:

```text

Does code compile?
  → success: Continue
  → fails: Stop, recommend /fix

Do tests exist?
  → found: Continue
  → missing: Stop, report "No tests found"

Can you read source?
  → readable: Continue
  → denied: Stop, report permission issue

Is this frontend/UI?
  → yes (React/Vue/Angular/Flutter/web-UI): Will run Design Quality Review
  → no (backend-only): Skip design review, run only Steps 1-5
```

---

## Assessment Completion Checklist

All the **Hard Stops** above apply to this run as preconditions — if one fails,
the run is not a passing assessment. It is complete when:

- ✅ Hard Stops above all satisfied (or their documented-exception path taken)
- ✅ All applicable workflow steps executed (Steps 1-6 always; Step 7 if frontend/UI)
- ✅ Operations hand-off produced or confirmed empty (Step 4)
- ✅ Score calculated using formula (not estimated)
- ✅ Top 5 findings listed with file:line or Figma URL locations
- ✅ Design quality assessed (if frontend/UI project detected)
- ✅ Every operations finding handed off, none scored here
- ✅ Next phase recommended clearly
- ✅ Console output delivered to user

## Persona Protocol

Loaded by every dimension-scoped persona prompt for this dimension — `/assess`,
`/implement`, `/fix`, `/optimize`, `/loop` and `/improve` all point their
`Loads:` line here. The rules below are dimension-neutral and are written here
once rather than restated in 54 persona files, where they had already drifted
apart from the parents they point at.

**What this file is, and is not.** It is the *assessment* criteria for this
dimension: the workflow, the weights, the grade bands and the mandatory audit.
Under `/assess` it is the criteria the run is scored against. Under the other
five families it is the only per-dimension reference that exists — the parent
family prompt supplies the verb (what to build, repair, optimize, iterate or
audit for currency) and this file supplies the dimension's subject matter. If
you need criteria the parent does not define, say so in your report rather than
borrowing the assessment workflow; an `/optimize` run is not an `/assess` run
with a different verb.

- **Follow the parent protocol.** The persona fixes the *dimension*; the family
  prompt fixes the *verb*. Follow the family prompt's full protocol, using this
  file for the dimension's criteria, weights and mandatory audit. Where the
  parent and this file disagree, the parent governs the behaviour and this file
  governs the criteria.
- **`$ARGUMENTS` is scoping detail, never a selector.** It narrows work *within*
  the dimension. It cannot select a different dimension, and it cannot be used to
  broaden scope to another dimension — that requires the user to invoke the
  other family explicitly (`/fix:technical,design`). A literal-minded reading of
  `$ARGUMENTS` as a command name is the failure this rule exists to prevent.
- **Do not manufacture work to justify being loaded.** A persona that finds
  nothing in its dimension reports "no findings in this dimension". Loading every
  persona for every run is what produces filler findings, not smaller honest ones.
- **Report the dimension you actually assessed.** Say which dimension ran. A
  finding filed under a neighbouring dimension's name is routed by that
  dimension's rules and will be closed by the wrong reviewer.
