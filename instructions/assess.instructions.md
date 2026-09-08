---
description: "Universal assessment orchestrator - delegates to specialized assessments per dimension"
---

# `/assess` command instructions

⚡ **Run with**: `ultracode` (multi-agent parallel execution)

## Objective

Universal assessment orchestrator - assesses ANY aspect of a project across THREE modes:

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

---

## Protocol

### Mode 1: Universal Research & Validation (Upfront)

**When to run**: At project start OR when clarity needed on any dimension
**Assess any dimension** (technical, business, marketing, product, legal, financial, operations):

1. **Research** - Explore options, understand landscape, analyze competitors
2. **Validation** - Assess against criteria (viability, feasibility, market fit, resources, risks)
3. **Analysis** - Market analysis, competitive positioning, technical feasibility, compliance gaps
4. **Recommendation** - Provide ranked options with tradeoffs
5. **Decision Support** - Help decide direction for any project aspect

**Output**: Recommendations + validation rationale

### Mode 2: Universal Quality & Viability Assessment (Post-Build)

**When to run**: MANDATORY after `/implement` and after `/optimize`
**Assess any dimension** to measure quality, viability, and readiness

**Output**: Multi-dimensional scores (0-100 each) + findings + recommendations

**Blocks progression**: Cannot run `/optimize`, `/fix`, or `/doc` without assessment completion

### Mode 3: Verification Mode (Post-Optimize)

**When to run**: MANDATORY immediately after `/optimize`
**Verify**: Improvements worked + no regressions

**Output**: Verification report + confidence score

---

## Specialized Assessment Delegation

For deeper expertise on specific dimensions, load specialized instruction files:

| Dimension | Load File | When Needed |
|-----------|-----------|------------|
| **Technical** | `.plaesy/memory/assess-technical.md` | Code quality, architecture, tests, security |
| **Design** | `.plaesy/memory/assess-design.md` | UI/UX, accessibility, design systems |
| **Business** | `.plaesy/memory/assess-business.md` | Market fit, business models, revenue |
| **Marketing** | `.plaesy/memory/assess-business.md` (section) | Positioning, messaging, GTM |
| **Legal/Compliance** | Ask `.plaesy/roles/privacy-legal.md` | Regulatory, compliance, data privacy |
| **Financial** | `.plaesy/memory/assess-business.md` (section) | Pricing, cost structure, profitability |
| **Operations** | Ask `.plaesy/roles/devops.md` | Deployment, infrastructure, team capacity |

**How to use**: Load relevant instruction file based on assessment dimension needed

---

## Pre-Assessment Checks

Before running assessment:

- ✅ Project structure understood
- ✅ Technology stack identified
- ✅ Required tools/dependencies available
- ✅ Project accessible and readable

---

## Output Format

**Console output**: Scores per dimension (0-100) + top 5 findings + next phase recommendation

**Files** (if user requests with `--report`):
- `assessment.json` — Structured findings
- `report.md` — Full report

---

## Error Recovery

**If assessment blocked**:
- Report blocker clearly
- Recommend `/fix` to resolve
- Suggest re-run after fix

**If assessment hangs**:
- Set timeout to 5 minutes
- Report partial results with completion status

---

## Success Criteria

Assessment is complete when:
- ✅ All applicable dimensions assessed
- ✅ Scores calculated per dimension
- ✅ Top findings listed with locations
- ✅ Next phase recommended clearly
- ✅ Console output delivered

---

**Follow shared protocols**: `.plaesy/memory/quality-gates.md` → `.plaesy/memory/error-recovery.md`
