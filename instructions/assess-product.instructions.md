---
description: "Product assessment workflow - feature set, roadmap, competitive positioning"
---

# Product Assessment Instructions

**Use with**: `/assess:product` when dimension is **Product**

**Referenced from**: `.plaesy/memory/assess.md` (orchestrator)

---

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

```
feature_completeness: 25% # Feature set, functionality, gaps
product_roadmap: 20%      # Vision, strategy, prioritization, alignment
competitive_position: 20% # Competitive advantage, differentiation, positioning
user_alignment: 20%       # User needs fit, feedback incorporation, validation
metrics_success: 15%      # KPIs, tracking, data-driven decisions
```

### Grade Mapping

```
95-100: A+ (Exceptional - Strong product, clear strategy, competitive advantage)
90-94:  A  (Excellent - Complete feature set, strategic roadmap, minor gaps)
85-89:  B+ (Very Good - Good product-market fit, some improvement needed)
80-84:  B  (Good - Solid features, roadmap in place, optimization opportunities)
70-79:  C  (Acceptable - Core features present, significant roadmap work needed)
<70:    D  (Needs Work - Major product gaps, strategy unclear, market risk)
```

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

## Anti-Patterns (NEVER Do These)

- ❌ **Never ignore user feedback** - User needs drive product success
- ❌ **Never assess features in isolation** - Evaluate against market needs
- ❌ **Never overlook competitive analysis** - Market context matters
- ❌ **Never chase feature creep** - Evaluate prioritization rigorously
- ❌ **Never speculate on market needs** - Validate with user research
- ❌ **Never ignore roadmap-reality gap** - Assess actual vs. planned delivery
- ❌ **Never skip product-market fit assessment** - It's the foundation

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

```
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

## Success Criteria

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
