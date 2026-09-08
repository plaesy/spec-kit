---
description: "Business assessment workflow - market fit, business model, financial viability"
---

# Business Assessment Instructions

**Use with**: `/assess` when dimensions are **Business**, **Marketing**, or **Financial**

**Referenced from**: `.plaesy/memory/assess.md` (orchestrator)

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

- Overall viability score (0-100)
- Per-category scores (Market Fit, Business Model, Financial, GTM, Roadmap)
- Top 5 findings (ordered by impact)
- Ranked options with tradeoffs

---

## Quality Scoring

### Business Assessment Components

```
market_fit: 25%           # Target market, customer validation, competitive advantage
business_model: 25%       # Revenue model, unit economics, scalability
financial: 20%            # Cost structure, profitability, funding needs
go_to_market: 15%         # Positioning, messaging, marketing readiness
product_roadmap: 15%      # MVP scope, feature prioritization, alignment
```

### Marketing Assessment Components

```
positioning: 25%          # Market positioning, differentiation, messaging
audience: 25%             # Target audience definition, reach strategy
messaging: 20%            # Message clarity, value proposition
competitive: 15%          # Competitive advantage, market positioning
brand: 15%                # Brand consistency, market awareness
```

### Financial Assessment Components

```
cost_structure: 20%       # Cost analysis, assumptions, sustainability
pricing: 25%              # Pricing viability, market fit, unit economics
profitability: 20%        # Profitability scenarios, margin analysis
funding_needs: 20%        # Funding requirements, burn rate, runway
financial_sustainability: 15% # Long-term viability, financial planning
```

### Grade Mapping

```
95-100: A+ (Exceptional - Strong viability, market fit validated)
90-94:  A  (Excellent - Good viability, minor risks)
85-89:  B+ (Very Good - Viable with medium risks, some assumptions)
80-84:  B  (Good - Viable but with some concerns, address in next iteration)
70-79:  C  (Acceptable - Significant uncertainties, validation needed)
<70:    D  (Needs Work - Major risks, significant validation needed)
```

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

## Anti-Patterns (NEVER Do These)

- ❌ **Never report unvalidated market claims** - Every finding must be verified
- ❌ **Never assume market demand** - Validate with actual customer feedback
- ❌ **Never ignore competitive threats** - Acknowledge competitive landscape
- ❌ **Never hide financial risks** - Flag sustainability concerns clearly
- ❌ **Never blend assumptions with facts** - Distinguish clearly
- ❌ **Never recommend features** - Only assess market viability, not product direction
- ❌ **Never ignore regulatory risks** - Flag compliance/legal concerns
- ❌ **Never speculate on timing** - Use data-driven market entry analysis

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

```
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

## Success Criteria

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
