---
description: "Product assessment workflow - feature set, roadmap, competitive positioning"
applyTo: "**/*"
---

# Product Assessment Instructions

**Use with**: `/assess:product` when dimension is **Product**

**Referenced from**: `/assess` (orchestrator)

---

**Reference checklists**: `.plaesy/checklists/po.checklist.md` (backlog, vision, value
delivery) and `.plaesy/checklists/update.checklist.md` (change impact assessment when a
feature/spec changes mid-flight) — use alongside this file's scoring dimensions.

**Delegates to**: `.plaesy/roles/pm.md`, `.plaesy/roles/ba.md`, `.plaesy/roles/po.md`.

## Assessment Workflow

### Step 1: Feature Set & Completeness

- Audit current feature set and functionality
- Verify feature-market fit against target audience needs
- Assess feature prioritization and roadmap alignment
- Review feature quality and user satisfaction
- Identify missing critical features or gaps

**Rate**: 0-100 based on feature completeness

### Step 2: Product Roadmap & Strategy

- Evaluate roadmap clarity and strategic alignment
- Assess product vision and long-term direction
- Review prioritization criteria and decision-making process
- Evaluate feature pipeline and delivery timeline
- Assess roadmap communication to stakeholders

**Rate**: 0-100 based on roadmap maturity

### Step 3: Competitive Analysis & Positioning

- Analyze competitive landscape and key competitors
- Assess competitive advantages and differentiation
- Evaluate market positioning vs. competitors
- Review feature parity with competing products
- Identify competitive threats and opportunities

**Rate**: 0-100 based on competitive strength

### Step 4: User Needs & Requirements Management

- Validate product against user needs and pain points
- Assess user feedback collection and incorporation
- Review requirement prioritization process
- Evaluate user research and validation methods
- Assess product-market fit signals

**Rate**: 0-100 based on user alignment

### Step 5: Product Metrics & Success Criteria

- Define product success metrics and KPIs
- Assess key performance indicators and tracking
- Review user engagement and retention metrics
- Evaluate product health indicators
- Assess data-driven decision-making practices

**Rate**: 0-100 based on metrics maturity

### Step 6: Generate Report

**Report sections**: canonical names per `.plaesy/instructions/dimension-mapping.md`
→ *Assessment Report Section Registry*; the bullets below go *inside* those
sections and do not rename them.

- Overall product score (0-100)
- Per-category scores (Features, Roadmap, Competitive Position, User Alignment, Metrics)
- Top 5 findings (ordered by impact)
- Recommended product improvements and priorities

**Blocking Issues**:

- Critical features missing that block market fit = BLOCKER
- No clear product strategy or vision = BLOCKER
- Feature roadmap misaligned with market needs = BLOCKER
- No user feedback mechanism = BLOCKER

---

## Quality Scoring

### Product Assessment Components

```text

feature_completeness: 25% # Feature set, functionality, gaps
product_roadmap: 20%      # Vision, strategy, prioritization, alignment
competitive_position: 20% # Competitive advantage, differentiation, positioning
user_alignment: 20%       # User needs fit, feedback incorporation, validation
metrics_success: 15%      # KPIs, tracking, data-driven decisions
```

### Grade Mapping

| Grade | What it means here |
|---|---|
| A+ | Exceptional - Strong product, clear strategy, competitive advantage |
| A | Excellent - Complete feature set, strategic roadmap, minor gaps |
| B+ | Very Good - Good product-market fit, some improvement needed |
| B | Good - Solid features, roadmap in place, optimization opportunities |
| C | Acceptable - Core features present, significant roadmap work needed |
| D | Needs Work - Major product gaps, strategy unclear, market risk |

The score edges behind these letters, and the general meaning of each band,
are defined once in `.plaesy/instructions/quality-gates.md` →
**Grade bands**. Do not restate them here.

### Success Criteria (Hard Stops)

- MUST HAVE: Clear product strategy and vision
- MUST HAVE: Well-defined product roadmap with prioritization
- MUST HAVE: Evidence of product-market fit or path to it
- MUST HAVE: User feedback mechanism and incorporation process
- MUST HAVE: Key success metrics and tracking

---

## Critical Rules

- ✅ **DATA-DRIVEN** - Base assessment on metrics and user feedback, not opinion
- ✅ **USER-CENTRIC** - Prioritize user needs over feature count
- ✅ **COMPETITIVE AWARENESS** - Compare against actual market alternatives
- ✅ **TREND ANALYSIS** - Consider market trends and future direction
- ✅ **CITE WITH EVIDENCE** - Reference metrics, user feedback, competitive data
- ✅ **STRATEGIC PERSPECTIVE** - Assess alignment with long-term vision

---

## Finding Standards

- **Weigh user feedback** — it drives product success
- **Evaluate every feature against market need**, never in isolation
- **Include competitive analysis** — market context changes how a finding reads
- **Apply rigorous prioritisation** against feature creep
- **Validate market need with user research**
- **Assess the roadmap-reality gap** — actual against planned delivery
- **Assess product-market fit first**; it is the foundation the rest rests on

---

## Key Metrics to Assess

**Feature & Functionality**:

- Feature count vs. competitors
- Feature quality and user satisfaction
- Feature adoption and usage rates
- Feature gap analysis
- MVP completeness

**Roadmap & Strategy**:

- Roadmap clarity and communication
- Feature delivery predictability
- Roadmap prioritization methodology
- Strategic alignment with vision
- Long-term product direction

**Competitive Position**:

- Market share and positioning
- Competitive advantage strength
- Feature parity vs. competitors
- Price-to-value positioning
- Market differentiation

**User Alignment**:

- Feature-customer fit score
- User satisfaction and NPS
- Customer feedback incorporation rate
- Churn rate and retention
- Product-market fit signals

**Metrics & Analytics**:

- Key product KPIs tracked
- User engagement metrics
- Retention and churn rates
- Feature usage metrics
- Revenue per user (if applicable)

---

## Pre-Assessment Validation

Run these checks BEFORE attempting assessment:

```text

Is product strategy documented?
  → yes: Review and assess strategy
  → no: Flag as critical gap, recommend documentation

Do product metrics exist?
  → yes: Analyze metrics and trends
  → no: Recommend metric definition

Is user feedback available?
  → yes: Incorporate into assessment
  → partial: Note feedback gaps
  → no: Recommend user research
```

---

## Uncertainty Surfacing (Before Finalizing Assessment)

Load `.plaesy/instructions/uncertainty-surfacing.md`; do not restate it here.
Only this dimension's specifics follow.

**Confidence basis for this dimension**:

- HIGH: the backlog, roadmap, or acceptance criteria was read directly
- MEDIUM: inferred from shipped behaviour or a partial backlog
- LOW: no spec or backlog system in use; the feature set is inferred from code alone

**Worked example**:

```text
I could not assess backlog hygiene because `.plaesy/tasks/` holds only the
auto-created scaffold.
To improve confidence, provide a populated backlog, or state that this project
runs without one so the check reports N/A rather than a finding.
```

---

## Assessment Completion Checklist

Assessment is complete when:

- ✅ Feature set assessed for completeness and quality
- ✅ Product roadmap reviewed and validated
- ✅ Competitive positioning analyzed
- ✅ User needs alignment assessed
- ✅ Product metrics reviewed and KPIs defined
- ✅ Scores calculated per category
- ✅ Top 5 findings listed with impact assessment
- ✅ Blocking issues (missing strategy, market risk) flagged prominently
- ✅ Competitive opportunities identified
- ✅ Product improvement priorities recommended
- ✅ Console output delivered to user

## Persona Protocol

Loaded by every `each family:product` persona prompt for this dimension — `/assess`,
`/implement`, `/fix`, `/optimize`, `/loop` and `/improve` all point their
`Loads:` line here. The rules below are dimension-neutral and are written here
once rather than restated in 54 persona files, where they had already drifted
apart from the parents they point at.

- **Follow the parent protocol.** The persona fixes the *dimension*; the family
  prompt fixes the *verb*. Follow the family prompt's full protocol. Where the
  parent and this file disagree, the parent governs the behaviour.
- **This file is the *assessment* criteria for `product`** — workflow, weights,
  grade bands, mandatory audit. Under `/assess` that is what you are scored
  against. Under `/implement`, `/fix`, `/optimize`, `/loop` and `/improve`
  it is the only per-dimension reference that exists: the parent supplies the
  verb, this file supplies the subject matter. If you need criteria the parent
  does not define, say so in your report rather than borrowing the assessment
  workflow — an `/optimize` run is not an `/assess` run with a different verb.
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
