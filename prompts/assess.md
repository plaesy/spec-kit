---
description: "Universal assessment orchestrator - delegates to specialized assessments per dimension"
subagent: true
---

# `/assess` command instructions

⚡ **Run with**: `ultracode` (multi-agent parallel execution)

## Usage Format

```bash
/assess                          # Assess all scope
/assess:technical                # Assess code quality, architecture, tests, security, infrastructure, deployment
/assess:design                   # Assess UI/UX, accessibility, design systems
/assess:business                 # Assess market fit, business model, revenue
/assess:marketing                # Assess positioning, messaging, go-to-market
/assess:legal                    # Assess regulatory compliance, data privacy, risks
/assess:financial                # Assess cost structure, pricing, profitability
/assess:management               # Assess team capacity, project management, processes
/assess:product                  # Assess feature set, roadmap, competitive advantage

# Multiple dimensions in one run
/assess:technical,design         # Technical + Design assessment
/assess:business,management      # Business + Management assessment
/assess:technical,business,design # Tech, Business, & Design assessment
```

## Objective

Universal assessment orchestrator - assesses ANY aspect of a project across THREE modes:

**Assessment Dimensions** (can assess any combination):
- 🔧 **Technical**: Code quality, architecture, performance, security, tests, infrastructure, deployment
- 🎨 **Design**: UI/UX, accessibility, design system, consistency, user experience
- 💼 **Business**: Market fit, viability, business model, revenue potential, ROI
- 📢 **Marketing**: Positioning, messaging, audience fit, competitive advantage, go-to-market
- 📊 **Product**: Feature set, roadmap, competitive analysis, user needs fit
- ⚖️ **Legal/Compliance**: Regulatory compliance, legal risks, data privacy, accessibility standards
- 💰 **Financial**: Cost structure, pricing, profitability, funding needs
- 🏢 **Management**: Team capacity, project management, process efficiency, organizational health

**Three Modes of Operation**:
1. **RESEARCH MODE** (Phase 1, upfront): Explore options, validate choices, resolve uncertainties (any dimension)
2. **ASSESSMENT MODE** (Phases 4, 6, post-build): Measure quality/viability, identify gaps (any dimension)
3. **VERIFICATION MODE** (Phase 6, post-optimize): Verify improvements worked, validate assumptions (any dimension)

---

## Protocol

### Mode 1: Universal Research, Validation & Ambiguity Resolution (Upfront)

**When to run**: At project start OR when clarity needed on any dimension. Run
`/assess` (or `/assess:{scope}`) before a spec/decision exists and it operates in
this mode. There is no separate ambiguity-resolution command — this mode covers both
external research and internal spec ambiguity in one pass, since both block the same
thing (a confident `/implement`).

**Assess any dimension** (technical, business, marketing, product, legal, financial, operations):

1. **Research** - Explore options, understand landscape, analyze competitors
2. **Validation** - Assess against criteria (viability, feasibility, market fit, resources, risks)
3. **Analysis** - Market analysis, competitive positioning, technical feasibility, compliance gaps
4. **Ambiguity Resolution** - Scan the current spec/requirements for vague statements
   ("improve performance", "be scalable") and TBD markers. Bound to **at most 5**
   highest-impact questions. Apply a safe default and label it `ASSUMED — <default>`
   when one exists (per constitution or convention); escalate only hard forks with no
   safe default (e.g. "REST vs GraphQL" with no constitution guidance) as a single
   blocking question. Write every resolution back into the spec directly, not into a
   separate notes file.
5. **Recommendation** - Provide ranked options with tradeoffs
6. **Decision Support** - Help decide direction for any project aspect

**Output**: Recommendations + validation rationale + spec updated with resolved ambiguities

#### Web Research Requirement (Anti-Hallucination — Mandatory in Mode 1)

Mode 1 exists to replace guessing with evidence. Any factual, technical, or market
claim made in Mode 1 output **must** be backed by a live source, not by model recall:

1. **Technical claims** (library versions, API behavior, framework capability,
   security advisories) → resolve via Context7 first
   (`mcp__context7__resolve-library-id` → `get-library-docs`); if Context7 has no
   coverage, fall back to `WebSearch`/`WebFetch` against official docs.
2. **Business/market/legal/financial claims** (market size, competitor features,
   pricing, regulatory requirements, compliance deadlines) → `WebSearch`/`WebFetch`
   against primary or authoritative sources. Knowledge-cutoff recall alone is not a
   valid citation for anything time-sensitive.
3. **Cite every claim inline**: `[claim] — Source: [name/URL], retrieved {{CURRENT_DATE}}`.
   A recommendation with no citation attached is a **speculative** finding (see
   Uncertainty Surfacing Protocol below), not a confirmed one — label it as such rather
   than presenting it as researched.
4. **Conflicting sources** → surface both with confidence levels (see error recovery:
   `.plaesy/instructions/error-recovery.md` → Assess Research Mode Recovery), let the user
   choose rather than silently picking one.
5. **Skip web research only when**: the claim is purely about *this* project's own
   code/spec (nothing external to verify), or the constitution already settled it.

#### Document & Source Ingestion (Any Dimension)

Evidence for Mode 1 doesn't only come from live web search — it can arrive as a file
the user hands you (contract, financial statement, competitor deck, spec) or as a
dense web page (long-form article, documentation, report). Applies to every
dimension, not just technical — a `legal` review reads contracts, `financial` reads
statements, `marketing`/`business` read market reports, `product` reads competitor
specs.

1. **Office/PDF documents** (DOCX, PPTX, XLSX, PDF) with structure worth preserving
   (tables, nested lists, embedded figures) → read fully before extracting claims;
   don't skim or summarize from the filename/first page. If structure is complex
   enough that a plain read risks losing table/figure relationships, say so explicitly
   rather than silently dropping data.
2. **Web pages as source documents** (articles, docs, competitor pages) → `WebFetch`
   already isolates the relevant content from page chrome via its summarization step;
   no separate cleanup pass is needed. Still apply the same citation rule as any other
   web source (item 3 above).
3. **Cite every claim from a document the same way as a web source**:
   `[claim] — Source: [file name or URL], retrieved {{CURRENT_DATE}}` (use the file
   name for user-supplied documents, no URL needed).
4. **Ambiguous or unreadable documents** (scanned PDFs with no text layer, corrupted
   files) → flag as a Missing Information item (see Uncertainty Surfacing Protocol)
   rather than guessing at content.

#### Design Spine (Production — When the Dimension Needs a New Design)

When Mode 1 output is a **design** rather than a decision — UI/UX, architecture,
business model, org structure, or process — every dimension follows the same spine,
instantiated differently:

| Spine Step | UI/UX | Architecture | Business Model | Org Structure | Process |
|---|---|---|---|---|---|
| **Components** | screens, widgets, components | services, modules, subsystems | value propositions, revenue streams, customer segments | teams, roles, positions, decision-making bodies | steps, handoffs, approval gates, escalation paths |
| **Tokens (no hardcoding)** | colors, spacing, typography, shadows | interfaces, contracts, schemas, data types | pricing models, unit economics, assumptions, CAC/LTV | role definitions, decision rights, accountability, compensation bands | SLAs, response times, escalation criteria, entry/exit conditions |
| **States & Edge Cases** | hover, focus, active, disabled, loading, empty, error | degraded mode, failover, cold start, circuit breaker, retry, backpressure | churn scenarios, downturn/recession, competitor entry, regulatory change | growth phases (startup → scale → enterprise), attrition, reorg, merger | exception cases, rollback procedures, deadlock/priority scenarios |
| **Mandatory Audit** (hard stop before handoff) | WCAG 2.1 AA (contrast, keyboard nav, screen readers) | Failure modes (latency p99, availability %, data loss risks) | Viability (unit economics breakeven, CAC payback period) | Sustainability (no single point of failure, span of control ≤6, attrition ≤10%) | Completeness (no unowned step, unbounded queue, or missing SLA) |
| **Handoff Spec** | `.plaesy/memory/design.md` (tokens + rationale, Google DESIGN.md convention) + Code Connect + impl guide + Figma file | ADRs + interface specs, deployment playbook, monitoring thresholds | Model canvas + assumption register, business case, go/no-go criteria | Org chart + RACI matrix, onboarding guide, decision framework | Runbook + owner map, KPI dashboard, escalation procedures |

**Process**: detect dimension from context → analyze current state (new/refactor/build
system) → generate spec against the spine above → define tokens (no hardcoded
values in any dimension) → run the mandatory audit (hard stop, cannot hand off if it
fails) → produce the handoff spec. UI/UX detail (accessibility audit steps, WCAG
checklist) lives in `assess-design.instructions.md`; the other four dimensions use the
table above directly since no dimension-specific instructions file exists yet for them.

**Output**: design spec/tokens/handoff doc for the dimension in scope, with the
mandatory audit result attached — pass required before `/implement` consumes it.

### Mode 2: Universal Quality & Viability Assessment (Post-Build)

**When to run**: MANDATORY after `/implement` and after `/optimize`
**Assess any dimension** to measure quality, viability, and readiness

**Output**: Multi-dimensional scores (0-100 each) + findings + recommendations

**Design audit** (any dimension that produced a design via Mode 1's spine): re-run that
dimension's Mandatory Audit row from the spine table above against the built/current
state — WCAG for UI (see `assess-design.instructions.md`, weighted into Design Quality
at 15% of the frontend score), failure-mode analysis for architecture, unit-economics
viability for business model, sustainability check for org structure, completeness
check for process. A failed mandatory audit is a hard-stop finding, not a suggestion —
route it to `/optimize` for redesign/refactor, same as any other quality gap.

**Blocks progression**: Cannot run `/optimize`, `/fix`, or `/doc` without assessment completion

### Mode 3: Verification Mode (Post-Optimize)

**When to run**: MANDATORY immediately after `/optimize`
**Verify**: Improvements worked + no regressions

**Output**: Verification report + confidence score

---

## Uncertainty Surfacing Protocol (Anti-Hallucination)

**BEFORE delivering any assessment, EXPLICITLY surface**:

### 1. Identify Missing Information
```
What I could NOT assess and why:
- [Area] → Data not available because [reason]
- [Area] → Insufficient context for confident evaluation
```

### 2. Confidence Levels for Each Finding
```
Confidence Rating (every finding must have one):
- HIGH: Based on code analysis + metrics + test results (not inferred)
- MEDIUM: Based on pattern detection + code inspection + limited testing
- LOW: Based on incomplete data, limited visibility, or specification gaps
```

### 3. State All Assumptions Made
```
Assumptions this assessment depends on:
- "I assumed [X] because [Y was not specified]"
- "If [assumption] is false, [conclusion] changes to [alternative]"
- "To improve confidence, provide [specific data]"
```

### 4. Flag Speculative Findings
```
CONFIRMED findings: [Based on specific evidence]
LIKELY issues: [Pattern matches, but not confirmed]
POSSIBLE risks: [Could happen IF [condition]]
SPECULATIVE: [Needs verification before acting]
```

---

## Scope Resolution

When scope is specified via `/assess:{scope}` format:

| Scope | Loads | Delegates To | Use Case |
|-------|-------|--------------|----------|
| **technical** | assess-technical.md | `.plaesy/roles/dev.md`, `.plaesy/roles/devsecops.md` | Code quality, tests, security, performance, infrastructure, deployment |
| **design** | assess-design.md | `.plaesy/roles/designer.md`, `.plaesy/roles/accessibility.md` | UI/UX, WCAG compliance, design systems |
| **business** | assess-business.md | `.plaesy/roles/ba.md`, `.plaesy/roles/pm.md` | Market fit, business model, viability |
| **marketing** | assess-marketing.md | `.plaesy/roles/market-research-analyst.md`, `.plaesy/roles/pm.md` | Positioning, messaging, GTM |
| **legal** | assess-legal.md | `.plaesy/roles/privacy-legal.md`, `.plaesy/roles/compliance.md` | Regulatory, compliance, data privacy, risk |
| **financial** | assess-financial.md | `.plaesy/roles/ba.md`, `.plaesy/roles/pm.md` | Pricing, cost structure, profitability, funding |
| **management** | assess-management.md | `.plaesy/roles/pm.md`, `.plaesy/roles/sm.md` | Team capacity, project management, processes, organizational health |
| **product** | assess-product.md | `.plaesy/roles/pm.md`, `.plaesy/roles/ba.md` | Feature set, roadmap, competitive advantage |

**Multiple scopes**: Comma-separated scopes run in parallel (e.g., `/assess:technical,design`)

**No scope**: Interactive orchestrator mode - choose dimension during execution

---

## Pre-Assessment Checks

Before running assessment:

- ✅ Project structure understood
- ✅ Technology stack identified
- ✅ Required tools/dependencies available
- ✅ Project accessible and readable
- ✅ **Analysis freshness**: if the assessment will cite `.plaesy/analysis/overview.md`
  (tech stack, file counts, detected tools) as evidence, run `plaesy analyze` first —
  regeneration is skipped automatically when the project fingerprint is unchanged, so
  this is cheap to call unconditionally; it only regenerates when the project has
  actually drifted since the last run. Never score or cite `overview.md` content
  without this call; a stale analysis produces findings against a project state that
  no longer exists.

---

## 🤖 Autonomous Findings Routing (NO User Input)

**Canonical routing table**: `.plaesy/instructions/dimension-mapping.md` — full
per-dimension mapping (assessment → implementation → optimization → error recovery)
for all 8 dimensions. This section states only the priority order; do not duplicate
the full table here.

**Priority order** (highest wins when multiple finding types exist at once):
1. Critical security/legal/compliance → `/fix` (immediate, blocks everything else)
2. Test coverage below constitution minimum → `/implement` (TDD)
3. Performance below constitution target → `/optimize`
4. Code quality / duplication (auto-fixable) → `/loop`
5. Documentation gaps → `/doc`

After every routed fix: re-assess (`/assess:{dimension}`) to confirm the finding
count dropped and no regression was introduced before moving to the next item.
Independent findings (e.g. missing tests + missing docs) may be fixed in parallel;
dependent ones (e.g. performance work that needs tests in place first) run in the
order above.

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

**Follow shared protocols**: `.plaesy/instructions/quality-gates.md` → `.plaesy/instructions/error-recovery.md`
