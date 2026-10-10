---
description: "Financial assessment workflow - cost structure, pricing, profitability, funding, unit economics"
applyTo: "**/*"
---

# Financial Assessment Instructions

**Use with**: `/assess:financial`

**Referenced from**: `/assess` (orchestrator), `.plaesy/instructions/dimension-mapping.md`

**Delegates to**: `.plaesy/roles/ba.md`, `.plaesy/roles/pm.md`, `.plaesy/roles/bo.md`

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

**Report sections**: canonical names per `.plaesy/instructions/dimension-mapping.md`
→ *Assessment Report Section Registry*; the bullets below go *inside* those
sections and do not rename them.

- Overall financial health score (0-100)
- Per-category scores (Cost Structure, Pricing, Profitability, Cash Flow/Funding, Risk)
- Top 5 findings (ordered by financial impact, $ or % where possible)
- Ranked options with tradeoffs (e.g., "raise prices 15%" vs "cut infra cost 20%")

---

## Quality Scoring

```text

cost_structure:        20%   # Cost breakdown clarity, scaling behavior, ownership
pricing:                25%   # Price-to-value fit, elasticity risk, competitive position
profitability:          25%   # Margins, unit economics (CAC/LTV), break-even clarity
cash_flow_funding:      20%   # Burn rate, runway, funding needs, revenue concentration risk
financial_risk:         10%   # Assumption sensitivity, compliance exposure, FX/vendor risk
```

### Grade Mapping

| Grade | What it means here |
|---|---|
| A+ | Exceptional - Strong margins, healthy unit economics, ample runway |
| A | Excellent - Sound financials, minor optimization opportunities |
| B+ | Very Good - Viable, some cost/pricing inefficiencies |
| B | Good - Viable but margin/runway pressure exists, address soon |
| C | Acceptable - Unit economics marginal, validation/optimization needed |
| D | Needs Work - Negative unit economics or runway <6 months, urgent action |

The score edges behind these letters, and the general meaning of each band,
are defined once in `.plaesy/instructions/quality-gates.md` →
**Grade bands**. Do not restate them here.

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

## Finding Standards

- **State the figure, or state that the data is unavailable** — a silent estimate
  reads as a fact and gets planned against
- **Report top-line and margin separately** in every figure
- **Quantify revenue concentration** — one large customer is a going-concern risk
- **State the elasticity assumption** behind any pricing recommendation
- **Record funding need as a binding constraint**, at the weight it deserves
- **Surface outlier cost drivers individually** — one catastrophic driver matters more
  than a smooth average

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

```text

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

Load `.plaesy/instructions/uncertainty-surfacing.md`; do not restate it here.
Only this dimension's specifics follow.

**Confidence basis for this dimension**:

- HIGH: actual billing or finance data with verifiable figures
- MEDIUM: partial data plus industry benchmark extrapolation
- LOW: assumptions with no supporting data

**Worked example**:

```text
I assumed a 4% monthly churn because no retention data was provided.
If churn is off by +/-20%, the 18-month viability conclusion changes to breakeven
at month 26, which fails the bar in section 1.
```

---

## Assessment Completion Checklist

Assessment is complete when:

- ✅ Cost structure broken into fixed/variable with drivers identified
- ✅ Pricing strategy validated against value and competitive position (or flagged as unavailable)
- ✅ Unit economics calculated (LTV:CAC, payback period) or explicitly marked unavailable
- ✅ Runway/burn rate assessed if cash data exists
- ✅ Financial risks identified and severity-ranked
- ✅ Scores calculated per category
- ✅ Top 5 findings with $ /% impact and data source listed
- ✅ Next phase recommended (`/optimize --cost`, `/optimize --pricing`, or `/assess:financial — Mode 1 if data is insufficient)
- ✅ Console output delivered to user

## Persona Protocol

Loaded by every `each family:financial` persona prompt for this dimension — `/assess`,
`/implement`, `/fix`, `/optimize`, `/loop` and `/improve` all point their
`Loads:` line here. The rules below are dimension-neutral and are written here
once rather than restated in 54 persona files, where they had already drifted
apart from the parents they point at.

- **Follow the parent protocol.** The persona fixes the *dimension*; the family
  prompt fixes the *verb*. Follow the family prompt's full protocol. Where the
  parent and this file disagree, the parent governs the behaviour.
- **This file is the *assessment* criteria for `financial`** — workflow, weights,
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
