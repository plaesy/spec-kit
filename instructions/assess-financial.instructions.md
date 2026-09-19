---
description: "Financial assessment workflow - cost structure, pricing, profitability, funding, unit economics"
---

# Financial Assessment Instructions

**Use with**: `/assess:financial`

**Referenced from**: `/assess` (orchestrator), `.plaesy/instructions/dimension-mapping.md`

**Delegates to**: `.plaesy/roles/ba.md`, `.plaesy/roles/pm.md`

---

## Assessment Workflow

### Step 1: Cost Structure Analysis

- Break down fixed costs (infrastructure, salaries, tooling, licensing) vs variable costs (COGS, per-transaction fees, usage-based cloud spend)
- Identify cost drivers and their scaling behavior (linear, sub-linear, step-function)
- Flag costs with no clear owner or that scale faster than revenue
- Compare actual/projected costs against budget or prior baseline

### Step 2: Pricing Strategy Validation

- Map current or proposed pricing model (flat, tiered, usage-based, seat-based, hybrid)
- Check price-to-value alignment against competitor pricing and willingness-to-pay signals
- Assess price elasticity risk (would a 10-20% price change plausibly break the model?)
- Validate discounting/promo practices don't erode margin structurally

### Step 3: Profitability & Unit Economics

- Calculate or validate gross margin, contribution margin, and (if applicable) net margin
- Compute unit economics: CAC (Customer Acquisition Cost), LTV (Lifetime Value), LTV:CAC ratio, CAC payback period
- Identify break-even point (unit-level and company-level, if data supports it)
- Flag negative or deteriorating unit economics as a hard-stop finding, not a suggestion

### Step 4: Cash Flow & Funding Needs

- Assess burn rate and runway (months of cash remaining at current burn)
- Identify funding requirements if runway is short or growth requires capital injection
- Review cash flow timing risks (e.g., annual prepay costs vs monthly revenue recognition)
- Flag any single-point-of-failure revenue concentration (e.g., >30% revenue from one customer)

### Step 5: Financial Risk Assessment

- Identify risks to financial sustainability (rising CAC, margin compression, currency/FX exposure, vendor lock-in cost risk)
- Assess sensitivity to key assumptions (what breaks the model if growth is 50% of forecast?)
- Cross-check for regulatory/financial compliance exposure (route to `/assess:legal` if found — do not duplicate)

### Step 6: Generate Report & Recommendations

- Overall financial health score (0-100)
- Per-category scores (Cost Structure, Pricing, Profitability, Cash Flow/Funding, Risk)
- Top 5 findings (ordered by financial impact, $ or % where possible)
- Ranked options with tradeoffs (e.g., "raise prices 15%" vs "cut infra cost 20%")

---

## Quality Scoring

```
cost_structure:        20%   # Cost breakdown clarity, scaling behavior, ownership
pricing:                25%   # Price-to-value fit, elasticity risk, competitive position
profitability:          25%   # Margins, unit economics (CAC/LTV), break-even clarity
cash_flow_funding:      20%   # Burn rate, runway, funding needs, revenue concentration risk
financial_risk:         10%   # Assumption sensitivity, compliance exposure, FX/vendor risk
```

### Grade Mapping

```
95-100: A+ (Exceptional - Strong margins, healthy unit economics, ample runway)
90-94:  A  (Excellent - Sound financials, minor optimization opportunities)
85-89:  B+ (Very Good - Viable, some cost/pricing inefficiencies)
80-84:  B  (Good - Viable but margin/runway pressure exists, address soon)
70-79:  C  (Acceptable - Unit economics marginal, validation/optimization needed)
<70:    D  (Needs Work - Negative unit economics or runway <6 months, urgent action)
```

### Success Criteria (Hard Stops)

- Must have: Cost structure broken down into fixed vs variable with clear drivers
- Must have: LTV:CAC ratio calculated (or explicitly flagged as unavailable + assumption stated)
- Must have: Runway calculated if burn-rate data exists
- Must have: Any negative/deteriorating unit economics flagged as CRITICAL, not MEDIUM

---

## Critical Rules

- ✅ **NUMBERS OVER NARRATIVE** — every finding backed by a number, ratio, or explicit assumption
- ✅ **SHOW THE MATH** — state the formula used (e.g., `LTV = ARPU × Gross Margin % × (1 / Churn Rate)`) so the user can verify
- ✅ **CITE DATA SOURCE** — finance/billing system, spreadsheet, or "ASSUMED — no data provided"
- ✅ **DISTINGUISH ACTUAL VS PROJECTED** — never blend historical actuals with forward projections without labeling
- ✅ **FLAG UNIT ECONOMICS FIRST** — negative unit economics is a blocking finding regardless of other scores
- ✅ **CURRENCY & PERIOD EXPLICIT** — always state currency and time period (monthly vs annual) for every figure

---

## Anti-Patterns (NEVER Do These)

- ❌ **Never invent financial figures** — if data is unavailable, say so and request it; do not estimate silently
- ❌ **Never conflate revenue with profit** — always distinguish top-line from margin
- ❌ **Never ignore revenue concentration risk** — a single large customer is a going-concern risk
- ❌ **Never recommend a price change without elasticity reasoning** — state the assumption behind any pricing recommendation
- ❌ **Never treat funding need as automatically solvable** — flag it as a real constraint, not a footnote
- ❌ **Never average away outliers** — a single catastrophic cost driver matters more than a smooth average

---

## Research Methods

**Primary Data Sources**:
- Billing/finance system exports, accounting ledgers, bank/cash statements
- Pricing page, sales contracts, discount/promo logs
- Cloud/infra billing dashboards (cost per unit, per customer, per request)

**Secondary/Benchmark Sources**:
- Industry SaaS benchmarks (e.g., median LTV:CAC, median gross margin by sector) — cite source and retrieval date, do not rely on recall
- Competitor public pricing pages

**Modeling**:
- Cost structure modeling (fixed vs variable breakdown)
- Unit economics calculation (CAC, LTV, payback period, contribution margin)
- Runway/burn-rate projection (linear extrapolation, clearly labeled as such)

---

## Pre-Assessment Validation

```
Is financial/billing data available?
  → yes: Proceed with full cost/margin/unit-economics analysis
  → partial: Flag which figures are ASSUMED, proceed with available data
  → no: Do not fabricate numbers — report "insufficient data" and request specific inputs

Is pricing information available (own + competitors)?
  → yes: Validate price-to-value and elasticity
  → no: Flag pricing assessment as SPECULATIVE, skip elasticity claims

Is burn-rate/cash data available?
  → yes: Calculate runway
  → no: Skip runway calculation explicitly rather than guessing
```

---

## Uncertainty Surfacing (Before Finalizing Assessment)

**MUST state explicitly**:

1. **Confidence Levels**:
   - HIGH: Based on actual billing/finance data with verifiable figures
   - MEDIUM: Based on partial data + industry benchmark extrapolation
   - LOW: Based on assumptions with no supporting data

2. **Assumptions Made**:
   - "I assumed [churn rate/ARPU/cost] = [X] because [no data was provided]"
   - "If [assumption] is off by ±20%, [conclusion] changes to [alternative]"

3. **Missing Information**:
   - "I could not calculate [LTV/runway/margin] because [specific data unavailable]"
   - "To improve confidence, provide [specific financial export/figure]"

4. **Speculative vs Confirmed**:
   - CONFIRMED: [Based on actual financial data]
   - LIKELY: [Based on benchmark extrapolation, not this project's actuals]
   - SPECULATIVE: [No data — needs financial input before acting]

---

## Success Criteria

Assessment is complete when:

- ✅ Cost structure broken into fixed/variable with drivers identified
- ✅ Pricing strategy validated against value and competitive position (or flagged as unavailable)
- ✅ Unit economics calculated (LTV:CAC, payback period) or explicitly marked unavailable
- ✅ Runway/burn rate assessed if cash data exists
- ✅ Financial risks identified and severity-ranked
- ✅ Scores calculated per category
- ✅ Top 5 findings with $ /% impact and data source listed
- ✅ Next phase recommended (`/optimize --cost`, `/optimize --pricing`, or `/assess:financial --research` if data is insufficient)
- ✅ Console output delivered to user
