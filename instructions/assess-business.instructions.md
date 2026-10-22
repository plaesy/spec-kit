---
description: "Business assessment workflow - market fit, business model, financial viability"
applyTo: "**/*"
---

# Business Assessment Instructions

**Use with**: `/assess:business`

**See also**: `.plaesy/instructions/assess-financial.md` (Financial dimension),
`.plaesy/instructions/assess-marketing.md` (Marketing dimension) — split out as
dedicated files; this file retains its own Marketing/Financial
component-scoring tables below for backward reference only.

**Referenced from**: `/assess` (orchestrator)

---

## Assessment Workflow

### Step 1: Market & Competitive Analysis

- Identify target market and customer segments
- Analyze competitive landscape and positioning
- Validate market size and growth potential
- Assess customer validation and validation methods
- Identify market differentiation and competitive advantages

### Step 2: Business Model Assessment

- Evaluate business model viability and sustainability
- Identify revenue streams and revenue model
- Assess unit economics and scalability assumptions
- Review pricing strategy and market positioning
- Validate assumptions with market data

### Step 3: Financial Assessment

- Validate cost structure and cost assumptions
- Analyze profitability scenarios and financial projections
- Calculate funding needs and burn rate (if applicable)
- Review pricing viability and market fit
- Assess financial runway and sustainability

### Step 4: Product-Market Fit Analysis

- Validate feature completeness vs market needs
- Assess user need satisfaction and value proposition
- Identify product roadmap alignment with strategy
- Review MVP scope and market positioning
- Measure customer validation signals

### Step 5: Go-to-Market Readiness

- Assess positioning clarity and messaging effectiveness
- Evaluate target audience fit and reach strategy
- Review go-to-market timeline and milestones
- Validate brand consistency and positioning
- Identify marketing and sales readiness

### Step 6: Risk Assessment

- Identify key business risks and mitigations
- Assess market adoption risks
- Review financial sustainability risks
- Evaluate competitive response risks
- Identify regulatory or compliance risks

### Step 7: Generate Report & Recommendations

**Report sections**: canonical names per `.plaesy/instructions/dimension-mapping.md`
→ *Assessment Report Section Registry*; the bullets below go *inside* those
sections and do not rename them.

- Overall viability score (0-100)
- Per-category scores (Market Fit, Business Model, Financial, GTM, Roadmap)
- Top 5 findings (ordered by impact)
- Ranked options with tradeoffs

---

## Quality Scoring

### Business Assessment Components

```text

market_fit: 25%           # Target market, customer validation, competitive advantage
business_model: 25%       # Revenue model, unit economics, scalability
financial: 20%            # Cost structure, profitability, funding needs
go_to_market: 15%         # Positioning, messaging, marketing readiness
product_roadmap: 15%      # MVP scope, feature prioritization, alignment
```

### Marketing Assessment Components

```text

positioning: 25%          # Market positioning, differentiation, messaging
audience: 25%             # Target audience definition, reach strategy
messaging: 20%            # Message clarity, value proposition
competitive: 15%          # Competitive advantage, market positioning
brand: 15%                # Brand consistency, market awareness
```

### Financial Assessment Components

```text

cost_structure: 20%       # Cost analysis, assumptions, sustainability
pricing: 25%              # Pricing viability, market fit, unit economics
profitability: 20%        # Profitability scenarios, margin analysis
funding_needs: 20%        # Funding requirements, burn rate, runway
financial_sustainability: 15% # Long-term viability, financial planning
```

### Grade Mapping

| Grade | What it means here |
|---|---|
| A+ | Exceptional - Strong viability, market fit validated |
| A | Excellent - Good viability, minor risks |
| B+ | Very Good - Viable with medium risks, some assumptions |
| B | Good - Viable but with some concerns, address in next iteration |
| C | Acceptable - Significant uncertainties, validation needed |
| D | Needs Work - Major risks, significant validation needed |

The score edges behind these letters, and the general meaning of each band,
are defined once in `.plaesy/instructions/quality-gates.md` →
**Grade bands**. Do not restate them here.

### Success Criteria (Hard Stops)

- Must have: Market viability validated or strong assumptions documented
- Must have: Business model sustainable or clear path to sustainability
- Must have: Financial projections realistic or assumptions documented
- Must have: Target market and customer segments clearly identified

---

## Critical Rules

- ✅ **NO SPECULATION** - Only report validated market insights, not hunches
- ✅ **DATA-DRIVEN** - Use market data, customer feedback, competitive analysis
- ✅ **CITE SOURCES** - Every finding includes source (market data, customer interview, competitor analysis)
- ✅ **VALIDATE ASSUMPTIONS** - Distinguish between facts and assumptions
- ✅ **ACCURACY > COMPLETENESS** - Better to report fewer findings with high confidence
- ✅ **MARKET REALITY** - Assess based on actual market, not wishful thinking

---

## Finding Standards

- **Verify every finding before reporting it** — an unvalidated market claim
  carries no weight for the reader deciding whether to act on it
- **Validate demand against actual customer feedback** before treating it as real
- **Acknowledge the competitive landscape**, threats included
- **Flag financial sustainability concerns clearly**, at the severity they carry
- **Keep assumptions and facts visually distinct** in every finding
- **Assess market viability, not product direction** — a feature recommendation
  belongs to `/assess:product`
- **Flag regulatory and compliance concerns** wherever they bear on viability
- **Ground timing claims in data-driven market entry analysis**

---

## Research Methods

**Primary Research**:

- Customer interviews and validation
- Market surveys and feedback
- User testing and feature validation
- Competitive analysis interviews

**Secondary Research**:

- Market sizing and TAM/SAM/SOM
- Industry reports and analysis
- Competitive landscape research
- Pricing and benchmarking analysis

**Financial Analysis**:

- Cost structure modeling
- Unit economics calculation
- Pricing strategy analysis
- Financial projections

---

## Pre-Assessment Validation

Run these checks BEFORE attempting assessment:

```text

Is there market data available?
  → yes: Continue with analysis
  → no: Flag assumptions, recommend primary research

Are there customer signals?
  → strong: Validate market fit
  → weak: Recommend customer interviews

Is financial data available?
  → yes: Analyze profitability and costs
  → partial: Flag assumptions, use available data
  → no: Request financial modeling
```

---

## Uncertainty Surfacing (Before Finalizing Assessment)

Load `.plaesy/instructions/uncertainty-surfacing.md`; do not restate it here.
Only this dimension's specifics follow.

**Confidence basis for this dimension**:

- HIGH: actual market data, customer feedback, or a validated metric from this business
- MEDIUM: competitive analysis plus partial data, with the gaps named
- LOW: no external validation; inferred from the product alone

**Worked example**:

```text
I assumed the target segment is enterprise (>$1000 ARR) because no
segmentation data was provided.
If the segment is actually self-serve, the go/no-go case changes to a
volume-led model with a much shorter payback requirement.
```

---

## Assessment Completion Checklist

Assessment is complete when:

- ✅ Market and competitive landscape analyzed
- ✅ Business model viability assessed with data
- ✅ Financial sustainability reviewed
- ✅ Go-to-market readiness evaluated
- ✅ Key risks identified and documented
- ✅ Scores calculated per category
- ✅ Top 5 findings with sources listed
- ✅ Next phase recommended (pivot, validate, execute)
- ✅ Console output delivered to user

## Persona Protocol

Loaded by every `each family:business` persona prompt for this dimension — `/assess`,
`/implement`, `/fix`, `/optimize`, `/loop` and `/improve` all point their
`Loads:` line here. The rules below are dimension-neutral and are written here
once rather than restated in 54 persona files, where they had already drifted
apart from the parents they point at.

- **Follow the parent protocol.** The persona fixes the *dimension*; the family
  prompt fixes the *verb*. Follow the family prompt's full protocol. Where the
  parent and this file disagree, the parent governs the behaviour.
- **This file is the *assessment* criteria for `business`** — workflow, weights,
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
