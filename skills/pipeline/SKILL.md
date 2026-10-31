---
name: pipeline
category: productivity
description: Run full pipeline from niche to make money verdict.
---

# pipeline

Master orchestrator for the complete business validation pipeline.

## When to Use
- User wants end-to-end: niche to idea to validation to business plan with verdict
- Single command to run all stages with proper handoffs

## Pipeline Stages

```
1. ideator          → Idea discovery (if no idea)
2. founder (idea.md)         → Define the chosen idea
3. founder (board.md)        → 3-lens board review
4. founder (competitors.md)  → Competitor research + pricing
5. buyer-discovery  → Real buyer interviews (5+)
6. founder (pricing.md)      → Pricing vs competitors + buyer evidence
7. founder (offer.md)        → Outcome, guarantee, included/excluded
8. compliance-health → Validate health claims (if health niche)
9. economics-engine → Margin, break-even, verdict
10. founder (marketing.md)   → Channel, message, first test
11. founder (brand.md)       → Name, tone, anti-brand
12. founder (ops.md)         → Delivery requirements, capacity
13. founder (launch.md)      → 30-day plan, smallest test, stop rules
14. plan-compiler   → Compile plan.md with verdict
15. assess          → Optional: audit business, marketing, financial
```

## Usage

```
/pipeline --niche health --idea "AI influencer for health coaches"
# or
/pipeline --topic ai-health-influencer --resume-from economics
```

## Handoff Rules
- Each stage writes to .plaesy/ideas/<topic>/<stage>.md
- Next stage reads previous stage files
- If any stage fails loudly (missing input), ask up to 3 questions, record UNKNOWN, continue
- Economics stage produces verdict: Incomplete | Not yet | Profitable
- Plan compiler runs automatically after economics
- Assessment runs optionally after plan

## Pivot Decision Gate (After Stage 9)

**If verdict = Not yet**:
- Run segment analysis (economics-engine)
- Run segment validation (buyer-discovery)
- Present pivot options:
  a) Narrow to viable segment(s)
  b) Reduce fixed costs (automate compliance, cut legal)
  c) Raise prices (reposition premium)
  d) Channel partnership (lower CAC)
- **REQUIRE USER DECISION** before proceeding to stages 10-13
- Document pivot rationale in plan.md → `## Pivot Analysis` with options table, chosen pivot, validation required

*Simulation learning: Pivot gate is the ONLY mandatory human decision in the pipeline. Economics verdict Not yet = pipeline halts until decision recorded. No auto-proceed.*

## Output
- Full .plaesy/ideas/<topic>/ directory with all stage files
- plan.md with verdict at top
- .plaesy/ideas.md index updated
- Console: verdict, 1-2 driving numbers, path to plan.md

## Evals

See skills/pipeline/evals/evals.json and results.json (6 tests, 100% pass).