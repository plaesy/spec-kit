---
description: "Technical assessment workflow - code quality, architecture, tests, security, infrastructure, deployment"
---

# Technical Assessment Instructions

**Use with**: `/assess` when dimension is **Technical**

**Referenced from**: `.plaesy/memory/assess.md` (orchestrator)

---

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
- Rate: 0-100 based on findings

### Step 4: Infrastructure & Deployment Assessment

- Audit current infrastructure (on-prem, cloud, hybrid) and scalability
- Verify CI/CD pipeline setup, automation, and deployment procedures
- Assess deployment reliability, rollback capability, and environment parity
- Review monitoring, alerting, and observability coverage
- Evaluate disaster recovery, backup procedures, and high-availability setup
- Assess access control, secrets management, and security patching procedures

**Rate**: 0-100 based on infrastructure and deployment maturity

### Step 5: Security Assessment

- Flag OWASP Top 10 High/Critical vulnerabilities only (not Low/Medium)
- No speculative security concerns
- Source: actual code patterns + known vulnerabilities

### Step 6: Documentation Review

- Check: README completeness, API docs accuracy, code comments clarity
- Flag: outdated/missing sections only (don't invent missing content)

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

- Overall score (0-100)
- Per-category scores (Quality, Tests, Infrastructure, Security, Docs, Performance, [+ Design if frontend])
- Top 5 findings (ordered by impact)

**Next Phase Decision Tree**:
- **If design quality score <80 (UI projects)** → Load `.plaesy/memory/assess-design.md`
- **If security/critical bugs found** → Run `/fix` first, then `/optimize`
- **If infrastructure/deployment issues found** → Address before scaling
- **If only performance issues** → Run `/optimize`
- **If all scores ≥80** → Can proceed to `/optimize`

---

## Quality Scoring

### Backend/API Projects

```
code_quality: 25%           # Structure, readability, patterns
test_coverage: 20%          # Coverage %, pass rate
infrastructure: 15%         # Scalability, redundancy, disaster recovery
security: 20%               # Vulnerabilities (High/Critical only)
documentation: 12%          # Completeness + accuracy
performance: 8%             # Response time, resource usage
```

### Frontend/UI Projects

```
code_quality: 22%           # Structure, readability, patterns
test_coverage: 18%          # Coverage %, pass rate
infrastructure: 12%         # Deployment automation, CI/CD, environment parity
security: 15%               # Vulnerabilities (High/Critical only)
documentation: 12%          # Completeness + accuracy
design_quality: 13%         # Component consistency, accessibility, design tokens, dark mode
performance: 8%             # Response time, resource usage, animation performance
```

### Design Quality Breakdown (if frontend/UI exists)

```
component_consistency: 25%   # Naming conventions, variants, documentation
accessibility: 35%           # WCAG 2.1 AA compliance, contrast, keyboard nav, ARIA
design_tokens: 20%          # Token usage, consistency, light/dark themes
design_system_adoption: 20% # Component reuse rate, library usage
```

### Grade Mapping

```
95-100: A+ (Exceptional - ship as-is)
90-94:  A  (Excellent - minor improvements suggested)
85-89:  B+ (Very Good - should fix before shipping)
80-84:  B  (Good - address issues in next iteration)
70-79:  C  (Acceptable - significant improvements needed)
<70:    D  (Needs Work - blocker issues present)
```

### Success Criteria (Hard Stops)

- Must have: Score ≥80 OR documented exceptions
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

## Anti-Patterns (NEVER Do These)

- ❌ **Never report unverified issues** - Every finding must come from actual analysis
- ❌ **Never speculate on architectural design** - Don't invent problems
- ❌ **Never include Low/Medium security findings** - Focus on High/Critical only
- ❌ **Never change the score manually** - Use the formula strictly
- ❌ **Never blame external factors** - Assess only what's in the codebase
- ❌ **Never recommend features** - Only assess what exists
- ❌ **Never assess design on backend-only projects** - Design quality only for frontend/UI
- ❌ **Never speculate on design quality** - Only audit actual design files, components, UI code
- ❌ **Never skip WCAG audit on frontend** - Accessibility is non-negotiable for UI projects

---

## Pre-Assessment Validation

Run these checks BEFORE attempting assessment:

```
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

## Success Criteria

Assessment is complete when:

- ✅ All applicable workflow steps executed (Steps 1-6 always; Step 7 if frontend/UI)
- ✅ Infrastructure & deployment assessment completed (Step 4)
- ✅ Score calculated using formula (not estimated)
- ✅ Top 5 findings listed with file:line or Figma URL locations
- ✅ Design quality assessed (if frontend/UI project detected)
- ✅ Infrastructure/deployment risks identified
- ✅ Next phase recommended clearly
- ✅ Console output delivered to user
