---
name: ideator
category: productivity
description: Generate divergent business ideas with assumptions, tests.
---

# ideator

Divergent idea generation per spec-kit ideator agent: 8-12 distinct options, each with assumption + cheapest test + FACT/INFERRED/SPECULATIVE label.

## When to Use
- Founder skill stage 1: idea discovery (user has no idea yet)
- User says "find me a business idea" or "suggest ideas"
- Need breadth of testable options, not recommendations

## Input

- Interests/skills
- Budget/time constraints
- Where to sell (city/online)
- Optional: problem area or market

## Process

1. Ask ≤3 clarifying questions (interests, budget, where to sell)
2. Web search for problems people pay to solve in that area (≥3 queries)
3. Open sources, extract real problems + paying customers
4. Generate 8-12 distinct options differing in KIND, not wording
5. For each option:
   - Name + one-line description
   - Single assumption it depends on
   - Cheapest test (what, cost, kill result)
   - Label: FACT | INFERRED | SPECULATIVE
6. Write each to .plaesy/ideas/<candidate-topic>/idea.md with status: candidate
7. Add index lines to .plaesy/ideas.md
8. Stop - let user choose one

## Output Schema

```
OPTIONS:
  1. [name] — [description]
     assumption: [single thing that must be true]
     cheapest test: [what, cost, kill result]
     label: FACT | INFERRED | SPECULATIVE
STATUS: done | blocked | needs-escalation
GAPS: [missing inputs that changed an option]
```

## Rules (from ideator.agents.md)
- Distinct, not repeated: same assumption = same option
- Testable: every option names assumption + cheap test
- No ranking: do not order by quality
- No fabricated evidence: market figures from brief or sources, else UNKNOWN
- At least one contrarian option

## Evals

See skills/ideator/evals/evals.json and results.json (4 tests, 100% pass).