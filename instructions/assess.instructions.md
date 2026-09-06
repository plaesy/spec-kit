---
description: "Dual-mode validation: research technologies upfront OR assess implementation quality post-build"
---

# `/assess` command instructions

⚡ **Run with**: `ultracode` (multi-agent parallel execution)

## Objective

Universal validation orchestrator - assesses ANY aspect of a project across THREE modes:

**Assessment Dimensions** (can assess any combination):
- 🔧 **Technical**: Code quality, architecture, performance, security, tests
- 🎨 **Design**: UI/UX, accessibility, design system, consistency, user experience
- 💼 **Business**: Market fit, viability, business model, revenue potential, ROI
- 📢 **Marketing**: Positioning, messaging, audience fit, competitive advantage, go-to-market
- 📊 **Product**: Feature set, roadmap, competitive analysis, user needs fit
- ⚖️ **Legal/Compliance**: Regulatory compliance, legal risks, data privacy, accessibility standards
- 💰 **Financial**: Cost structure, pricing, profitability, funding needs
- 🏢 **Operations**: Process efficiency, team capacity, deployment readiness, support model

**Three Modes of Operation**:
1. **RESEARCH MODE** (Phase 1, upfront): Explore options, validate choices, resolve uncertainties (any dimension)
2. **ASSESSMENT MODE** (Phases 4, 6, post-build): Measure quality/viability, identify gaps (any dimension)
3. **VERIFICATION MODE** (Phase 6, post-optimize): Verify improvements worked, validate assumptions (any dimension)

Both modes ensure validity through research, measurement, and validation across requested dimensions.

## Protocol

### Mode 1: Universal Research & Validation (Upfront)

**When to run**: At project start OR when clarity needed on any dimension
**Assess any dimension** (technical, business, marketing, product, legal, financial, operations):

1. **Research** - Explore options, understand landscape, analyze competitors
2. **Validation** - Assess against criteria (viability, feasibility, market fit, resources, risks)
3. **Analysis** - Market analysis, competitive positioning, technical feasibility, compliance gaps
4. **Recommendation** - Provide ranked options with tradeoffs
5. **Decision Support** - Help decide direction for any project aspect

**Example Assessment Areas**:
- **Technical**: Framework selection, architecture patterns, performance requirements
- **Business**: Market viability, business model options, revenue strategies, funding needs
- **Marketing**: Positioning options, messaging strategy, audience targeting, GTM approaches
- **Product**: Feature prioritization, MVP scope, competitive advantages, user need validation
- **Legal**: Compliance requirements, regulatory landscape, risk assessment
- **Financial**: Pricing models, cost structure, profitability scenarios
- **Operations**: Team capabilities, deployment readiness, process efficiency

**Output**: Recommendations + validation rationale for chosen dimension

### Mode 2: Universal Quality & Viability Assessment (Post-Build/Development)

**When to run**: MANDATORY after `/implement` and after `/optimize` (or anytime to assess any dimension)
**Assess any dimension** to measure quality, viability, and readiness:

1. **Technical Assessment**:
   - Code quality (performance, bugs, patterns, complexity)
   - Test coverage (≥90% required)
   - Security audit (OWASP High/Critical)
   - Documentation (completeness, accuracy)

2. **Design Assessment** (if UI):
   - WCAG AA compliance, consistency, design tokens
   - User experience, interaction patterns
   - Accessibility, dark mode support

3. **Business Assessment**:
   - Market fit validation, customer validation
   - Business model viability, revenue sustainability
   - Competitive positioning, market differentiation
   - Scalability assumptions, growth potential

4. **Product Assessment**:
   - Feature completeness vs MVP scope
   - User need satisfaction
   - Competitive advantage validation
   - Roadmap alignment

5. **Marketing Assessment**:
   - Positioning clarity, messaging effectiveness
   - Target audience fit, go-to-market readiness
   - Brand consistency, messaging validation

6. **Legal/Compliance Assessment**:
   - Regulatory compliance gaps
   - Data privacy, security compliance
   - Accessibility standards (WCAG)
   - Risk identification

7. **Financial Assessment**:
   - Cost structure validation, profitability analysis
   - Pricing viability, unit economics
   - Funding runway, burn rate

8. **Operations Assessment**:
   - Team capacity, skill gaps
   - Deployment readiness, infrastructure
   - Support model, scalability

**Quality Scoring**: Generate 0-100 score per assessed dimension + top findings + research-validated recommendations

**Output**: Multi-dimensional assessment report + findings + improvement recommendations

**Blocks progression**: Cannot run `/optimize`, `/fix`, or `/doc` without assessment completion

---

## Specialized Assessment Delegation

For deeper expertise on specific dimensions, `/assess` can delegate to specialized chatmodes:

### Assessment Dimension → Chatmode Delegation

| Dimension | Chatmode | Expertise |
|-----------|----------|-----------|
| **Design** | `.plaesy/roles/designer.md` | UI/UX, accessibility, design systems |
| **Business** | `.plaesy/roles/ba.md` | Market fit, business models, strategy |
| **Product** | `.plaesy/roles/pm.md` | Feature strategy, roadmap, user needs |
| **Marketing** | `.plaesy/roles/market-research-analyst.md` | Positioning, messaging, market analysis |
| **Legal** | `.plaesy/roles/privacy-legal.md` | Compliance, data privacy, regulations |
| **Accessibility** | `.plaesy/roles/accessibility.md` | WCAG compliance, a11y standards |
| **Compliance** | `.plaesy/roles/compliance.md` | Regulatory requirements, risk |
| **Operations/DevOps** | `.plaesy/roles/devops.md` | Deployment, infrastructure, operations |
| **Dev** | `.plaesy/roles/dev.md` | Code quality, architecture, patterns |
| **QA** | `.plaesy/roles/qa.md` | Testing strategy, test coverage |
| **Security** | `.plaesy/roles/devsecops.md` | Security assessment, threat modeling |
| **AI/ML** | `.plaesy/roles/mlops.md` | ML infrastructure, model validation |
| **Data** | `.plaesy/roles/data-engineer.md` | Data architecture, data quality |
| **Architecture** | `.plaesy/roles/sa.md` | System architecture, technical design |

**How to use**: When `/assess` needs expert evaluation on a dimension, reference the chatmode file from `.plaesy/roles/` for deeper analysis

---

## Validation Checklist (Before Running)

- ✅ Project compiles/builds successfully
- ✅ Tests exist and can run (test suite configured)
- ✅ Source code readable (no syntax errors blocking analysis)
- ✅ Documentation files present (README, code comments)
- ✅ Project structure understood (frameworks detected)

If any unchecked: run `/fix` first to resolve blockers.

---

## Assessment Workflow

### Step 1: Project Analysis

- Detect language + framework (from file patterns, not assumed accuracy)
- Identify technology stack
- Map project structure
- **Detect frontend/UI presence**: Check for React, Vue, Angular, Flutter, mobile UI, or Figma files → triggers optional Step 6 (Design Quality Review)

### Step 2: Test Execution

- Run full test suite
- Capture coverage metrics
- Record failure messages (if any)

### Step 3: Code Quality Review

- Check for: unused code, complexity hotspots, naming clarity
- Identify: n+1 queries, missing error handling, security issues (OWASP Top 10 High/Critical only)
- Rate: 0-100 based on findings

### Step 4: Security Assessment

- Flag OWASP Top 10 High/Critical vulnerabilities only (not Low/Medium)
- No speculative security concerns
- Source: actual code patterns + known vulnerabilities

### Step 5: Documentation Review

- Check: README completeness, API docs accuracy, code comments clarity
- Flag: outdated/missing sections only (don't invent missing content)

### Step 6: Design Quality Review (Conditional — only if frontend/UI detected)

**When to run**: Project has React, Vue, Angular, Flutter, web UI, mobile UI, or Figma files

- **Component Consistency**: Check if UI components follow naming conventions, variants documented
- **Accessibility**: Audit WCAG 2.1 AA compliance (contrast ratios, keyboard navigation, ARIA labels)
- **Design Tokens**: Verify colors, spacing, typography defined as tokens (not hardcoded)
- **Dark Mode Support**: Check if light/dark theme colors are consistent and accessible
- **Design System Adoption**: Measure component reuse rate from design library
- **Design Documentation**: Check if design spec exists and is current
- **Figma Organization**: If using Figma, check file structure, component library, Code Connect mappings

**Rate**: 0-100 based on findings (applied ONLY if frontend/UI exists)

### Step 7: Generate Report & Next Phase Decision

- Overall score (0-100)
- Per-category scores (Quality, Tests, Security, Docs, Performance, [+ Design if frontend])
- Top 5 findings (ordered by impact)

**Next Phase Decision Tree**:
- **If design quality score <80 (UI projects)** → MUST run `/design` to refactor/improve
- **If security/critical bugs found** → Run `/fix` first, then `/optimize`
- **If only performance issues** → Run `/optimize`
- **If all scores ≥80** → Can proceed to `/optimize`

**IMPORTANT**: Design issues MUST be fixed before optimization. Cannot skip `/design` if design quality is low.

---

## Quality Scoring

```yaml
# Score Components (Backend/API Projects)
code_quality: 30%       # Structure, readability, patterns
test_coverage: 25%      # Coverage %, pass rate
security: 20%           # Vulnerabilities (High/Critical only)
documentation: 15%      # Completeness + accuracy
performance: 10%        # Response time, resource usage

# Score Components (Frontend/UI Projects) - OPTIONAL design_quality added
code_quality: 25%       # Structure, readability, patterns
test_coverage: 20%      # Coverage %, pass rate
security: 15%           # Vulnerabilities (High/Critical only)
documentation: 15%      # Completeness + accuracy
design_quality: 15%     # Component consistency, accessibility, design tokens, dark mode
performance: 10%        # Response time, resource usage, animation performance

# Design Quality Breakdown (if frontend/UI exists)
design_quality:
  component_consistency: 25%   # Naming conventions, variants, documentation
  accessibility: 35%           # WCAG 2.1 AA compliance, contrast, keyboard nav, ARIA
  design_tokens: 20%          # Token usage, consistency, light/dark themes
  design_system_adoption: 20% # Component reuse rate, library usage

# Grade Mapping
95-100: A+ (Exceptional - ship as-is)
90-94:  A  (Excellent - minor improvements suggested)
85-89:  B+ (Very Good - should fix before shipping)
80-84:  B  (Good - address issues in next iteration)
70-79:  C  (Acceptable - significant improvements needed)
<70:    D  (Needs Work - blocker issues present)

# Success Criteria (Hard Stops)
- Must have: Score ≥80 OR documented exceptions
- Must have: No OWASP High/Critical vulns unfixed
- Must have: Tests pass (or known failures documented)
- Must have (if frontend): WCAG 2.1 AA compliance or documented accessibility exceptions
```

---

## Critical Rules

- ✅ **NO SPECULATION** - Only report what code/design actually does, not what it might do
- ✅ **SECURITY: High/Critical Only** - Ignore OWASP Low/Medium findings (too noisy)
- ✅ **CITE WITH LOCATION** - Every finding includes file:line reference (or Figma URL for design)
- ✅ **MEASURE DON'T ESTIMATE** - Use actual test results, coverage %, accessibility metrics, not "appears to"
- ✅ **ACCURACY > COMPLETENESS** - Report fewer findings with high confidence vs guesses
- ✅ **DESIGN ONLY IF FRONTEND** - Only assess design quality if project has UI (React, Vue, Flutter, web UI, etc.)

---

## Anti-Patterns (NEVER Do These)

- ❌ **Never report unverified issues** - Every finding must come from actual code/design analysis
- ❌ **Never speculate on architectural design** - Don't invent problems not in the code
- ❌ **Never include Low/Medium security findings** - Focus on High/Critical only
- ❌ **Never change the score manually** - Use the formula above strictly
- ❌ **Never blame external factors** - Assess only what's in the codebase
- ❌ **Never recommend features** - Only assess what exists, don't suggest new code
- ❌ **Never assess design on backend-only projects** - Design quality only applies to frontend/UI projects
- ❌ **Never speculate on design quality** - Only audit actual design files, components, and UI code
- ❌ **Never skip WCAG audit on frontend** - Accessibility is non-negotiable for UI projects

---

## Need Help Prioritizing Fixes?

When assessment score is <80 or findings are complex, consult relevant specialized chatmodes:

**For multi-dimensional prioritization:**
- **Product/Business priorities** → `.plaesy/roles/pm.md` or `.plaesy/roles/ba.md`
  - Which features block shipping vs can iterate?
  - Business impact vs technical debt?
  - Timeline vs scope tradeoffs?
  
- **Technical prioritization** → `.plaesy/roles/sa.md` (architecture) or `.plaesy/roles/dev.md` (code quality)
  - Which technical issues block shipping?
  - Refactoring ROI vs new features?
  - Performance vs code quality prioritization?

- **Design/UX issues** → `.plaesy/roles/designer.md`
  - Accessibility compliance blockers
  - Design system adoption vs one-off fixes

**Example:**
```
Assessment score 76/100 with:
- 3 WCAG accessibility issues (blocker)
- 2 performance bottlenecks (nice-to-have)
- Tech debt in auth module (medium)
- Timeline: 2 weeks to ship

Consult .plaesy/roles/accessibility.md for WCAG compliance priority,
.plaesy/roles/sa.md for auth module prioritization
```

---

## Output (Console-First, Minimal Files)

**Default**: Report to stdout only. Files created **only if** user explicitly requests with `--report` flag.

**Console Output Format:**

```
✅ Assessment Complete (2 min)

📊 Scores (Backend/API):
- Overall: 92/100 (Grade A)
- Code Quality: 90/100
- Test Coverage: 88/100
- Security: 95/100
- Documentation: 85/100
- Performance: 98/100

📊 Scores (Frontend/UI):
- Overall: 84/100 (Grade B)
- Code Quality: 85/100
- Test Coverage: 82/100
- Security: 88/100
- Documentation: 80/100
- Design Quality: 78/100  ← Design assessment included
  - Component Consistency: 75/100
  - Accessibility: 72/100 (WCAG 2.1 AA issues)
  - Design Tokens: 85/100
  - Design System Adoption: 80/100
- Performance: 90/100

🔍 Findings (Top 5):
1. [src/Button.tsx:12] Insufficient color contrast in button states (WCAG AA fail)
2. [file:line] Missing keyboard navigation support in modal dialog
3. [Figma: Components] Component naming inconsistent (Button vs btn_primary)
4. [file:line] N+1 database queries in product listing (Medium impact)
5. [file:line] Design tokens not exported, colors hardcoded in components

💡 Next Steps:
- Score ≥80: Ready for /optimize
- Score <80: Run /fix then re-assess
- Design quality <80: Consider /design phase for refactoring UI/components
- Accessibility issues: Must fix WCAG violations before shipping
- Need to prioritize: Call @anara to decide what to fix first
```

**Optional Files** (if user asks with `--report`):

- `assessment.json` - Structured findings (machine-readable)
- `report.md` - Full human-readable report with details

---

## Error Recovery

**If tests fail to run:**

- Report which tool failed (pytest, jest, go test, etc.)
- List first 5 error messages
- Recommend: run `/fix` to resolve blockers

**If project doesn't compile:**

- Stop assessment
- Report: "Code does not compile. Run `/fix` first."
- Do NOT attempt code-level assessment

**If assessment hangs:**

- Set timeout to 5 minutes
- Report partial results: "Assessment interrupted at [step]"
- List what was completed before timeout

---

## Pre-Assessment Validation

Run these checks BEFORE attempting assessment:

```bash
# Does code compile?
build_cmd success? → Continue
build_cmd fails? → Stop, recommend /fix

# Do tests exist?
test_files found? → Continue
test_files missing? → Stop, report: "No tests found. Use /implement to add tests."

# Can you read source?
source_files readable? → Continue
permissions denied? → Stop, report permission issue

# Is this a frontend/UI project?
React/Vue/Angular/Flutter/Dart/Android/IOS/web-UI files found? → Will run Design Quality Review (Step 6)
Figma file found? → Will include Figma design assessment
Backend-only (no UI)? → Skip Step 6, run only Steps 1-5
```

---

## Success Criteria

Assessment is complete when:

- ✅ All applicable workflow steps executed (Steps 1-5 always; Step 6 if frontend/UI detected)
- ✅ Score calculated using formula (not estimated)
- ✅ Top 5 findings listed with file:line (or Figma URL) locations
- ✅ Design quality assessed (if frontend/UI project detected)
- ✅ Next phase recommended (optimize if ≥80, fix if <80, design if design quality low)
- ✅ Console output delivered to user

---

## Completion Format

```
✅ Assessment Complete

Quality Scores:
- Code Quality: [X/100]
- Test Coverage: [X/100]
- Security: [X/100]
- Documentation: [X/100]
- Performance: [X/100]
- Design Quality: [X/100] (if UI project)

Overall Score: [X/100] (Grade [A-D])
Findings: [N] issues identified

NEXT PHASE:
- If design score <80 (UI projects): → /design (MANDATORY - refactor UI/UX/flow)
- Else if critical bugs: → /fix (then /optimize)
- Else: → /optimize

Estimated remediation time: [N hours]
```

---

**Follow shared protocols:** `.plaesy/memory/quality-gates.md` → `.plaesy/memory/error-recovery.md`
