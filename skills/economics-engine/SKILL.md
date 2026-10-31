---
name: economics-engine
category: productivity
description: Auto-calculate unit economics, break-even, verdict.
---

# economics-engine

Automate founder skill economics.md calculations with sensitivity analysis.

## When to Use
- Founder skill stage 7: need to compute margin, break-even, verdict
- Want to test scenarios: price change, cost change, volume change
- Need sensitivity analysis for investor/founder decisions

## Inputs

```
Price per sale (P) - can be single or blended across tiers
Direct cost per sale (V) - AI API, hosting, payment processing
Fixed cost per month (F) - tools, marketing, compliance, buffer
Expected sales per month (S) - with stated basis (required)
```

## Calculations

- Margin per sale: m = P - V
- Margin %: m / P
- Break-even: B = ceil(F / m) if m > 0 else INF
- Margin of safety: (S - B) / S if S >= B else NEGATIVE
- Year-1 revenue: 12 * S * P
- Payback period (if CAC given): CAC / m
- Churn-adjusted LTV: m * (1 / monthly_churn_rate) if churn > 0
- Months to break-even: B / S (if S > 0)

## Verdict Engine (per founder skill rules)

1. Incomplete: any input missing, S has no basis, required stage file absent
2. Not yet: m <= 0 or S < B (name failing number)
3. Profitable: m > 0 and S >= B (state margin of safety, label estimate if basis weak)

### Margin of Safety Thresholds (simulation-validated)

- **MoS ≥ 20%**: Unconditional Profitable — model robust to volume/cost variance
- **MoS 10-20%**: Conditional Profitable — label "estimate, needs validation"; requires pivot validation before launch
- **MoS < 10%**: Fragile — treat as Not yet; single churn or cost increase flips verdict

*Simulation learning: ai-health-influencer simulated MoS 10.7% (Conditional Profitable) but real data MoS -77.8% (Not yet). Threshold 20% prevents false Profitable verdicts from thin margins.*

## Segment Analysis (Multi-tier Pricing)

When multiple tiers exist, calculate:
- Blended P = Σ(tier_price × tier_mix%)
- Blended V = Σ(tier_vcost × tier_mix%)
- Per-tier break-even: B_tier = F / m_tier (attributable fixed cost)
- Cross-subsidy check: Does low-margin tier drain resources from high-margin?
- Recommendation: Kill tier if m_tier < 0 OR if tier CAC > tier LTV

## Sensitivity Analysis

Vary each input ±20% and show verdict impact:
- Price sensitivity: what price drop kills profitability?
- Cost sensitivity: what API cost increase kills margin?
- Volume sensitivity: what churn rate kills break-even?
- CAC sensitivity: what CAC makes payback >12 months?
- Fixed cost sensitivity: what legal/compliance cost increase kills break-even?
- Segment mix sensitivity: what % coach vs B2B sales kills blended model?

## Output: economics.md + scenario table

## Evals

See skills/economics-engine/evals/evals.json and results.json (9 tests, 100% pass).