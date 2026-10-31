---
name: buyer-discovery
category: productivity
description: Track buyer outreach, interviews, validate demand.
---

# buyer-discovery

Manage the buyer validation process from outreach to signal counting.

## When to Use
- Founder skill stage 5 (buyers.md) needs real conversations
- Need to track outreach funnel: contact -> reply -> call -> would pay
- Want to count real "would pay at this price" signals

## Workflow

### 1. Outreach Tracker
- Target list: role, company, contact info, source, status
- Template messages per segment
- Follow-up schedule

### 2. Interview Recorder
- Date, person (role, not name), questions, answers, would_pay (yes/no/maybe), price_point
- Audio recording link (optional)
- Tags: pain_point, feature_request, objection

### 3. Signal Counter
- Count: total conversations, "would pay" yes, "would pay" maybe
- Weight by: decision_maker (founder=1.0, employee=0.5), budget_authority
- Minimum threshold: 5 real conversations, 3+ "would pay" for Profitable verdict

### 4. Insight Synthesis
- Cluster answers by theme
- Extract pricing willingness distribution
- Identify must-have vs nice-to-have features

### 5. Segment Validation Matrix
- For each segment: conversations_count, would_pay_count, avg_price, price_range, top_3_pain_points, top_3_features
- Segment viability: would_pay_rate > 60% AND avg_price > tier_minimum
- Kill segment if: conversations >= 3 AND would_pay_rate < 30%
- Pivot signal: >2 segments viable but with conflicting product needs

## Output: buyers.md (founder skill format)

```markdown
| Date | Person (role) | Question | Answer | Would pay? |
|---|---|---|---|---|
```

Plus summary: conversation_count, would_pay_count, price_range, top_insights

## Evals

See skills/buyer-discovery/evals/evals.json and results.json (5 tests, 100% pass).