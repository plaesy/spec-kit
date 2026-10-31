---
name: plan-compiler
category: productivity
description: Compile founder stage files into plan.md with verdict.
---

# plan-compiler

Auto-generate plan.md from all stage files in .plaesy/ideas/<topic>/.

## When to Use
- Founder skill stage 10: compile plan.md
- Want consistent format across all ideas
- Need to update plan when any stage file changes

## Input

Directory: .plaesy/ideas/<topic>/
Required files: idea.md, board.md, competitors.md, buyers.md, pricing.md, offer.md, economics.md, marketing.md, brand.md, ops.md, launch.md

## Process

1. Verify all required files exist
2. Read economics.md for verdict + driving numbers
3. Read board.md for board votes summary
4. Concatenate all stage files in order with headers
5. Write plan.md with:
   - Verdict + numbers at top
   - One-page summary (auto-generated from idea + offer + economics)
   - All stage files in order
6. Update .plaesy/ideas.md index line with verdict + date

## One-Page Summary Template

```markdown
# <topic>: business plan

**Verdict:** <Profitable|Not yet|Incomplete>
**Numbers:** margin $X (Y%), break-even Z sales/mo, expected S sales/mo, MoS W%
**Board:** <votes summary>

## Summary
<what it is, who for, offer, price, verdict in plain words>
```

## Pivot Documentation (If Applicable)

If economics verdict = Not yet, append:

```markdown
## Pivot Analysis

**Original Model:** <segments, pricing, F, S, B, verdict>
**Failing Number:** <S < B | m <= 0 | missing input>
**Pivot Options Evaluated:**
| Option | Target Segments | P | F | Target S | B | MoS | Verdict |
|---|---|---|---|---|---|---|---|
| <name> | <segments> | $X | $Y | N | M | Z% | <verdict> |

**Chosen Pivot:** <option name>
**Rationale:** <why this option>
**Validation Required:** <what to test before proceeding>
```

## Integration
- Called automatically after economics.md written
- Can be run standalone to refresh plan.md

## Evals

See skills/plan-compiler/evals/evals.json and results.json (5 tests, 100% pass).