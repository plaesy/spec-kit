# Economics and the verdict

All numbers go in `.plaesy/ideas/<topic>/economics.md`. Show the arithmetic, not only the result.

## Inputs

| Input | Symbol | Source |
| --- | --- | --- |
| Price per sale | P | `.plaesy/ideas/<topic>/pricing.md` |
| Direct cost per sale | V | user, or a cited supplier price |
| Fixed cost per month | F | user, or a cited estimate |
| Expected sales per month | S | a stated basis: buyer answers, a channel's reach, or a test plan |

If any input is missing, write it as "unknown" and set the verdict to Incomplete.
Do not fill it in.

## Formulas

- Contribution margin per sale: `m = P - V`
- Margin percentage: `m / P`
- Break-even sales per month: `B = F / m`, rounded UP to a whole number of sales (only when `m > 0`).
  Example: 600 / 37 = 16.2, so B = 17.
- Margin of safety: `(S - B) / S` (only when `S >= B`)
- Year-one revenue at expected sales: `12 × S × P`

## Example

```markdown
Price P = 49. Direct cost V = 12. Margin m = 37 (76%).
Fixed cost F = 600 per month. Break-even B = 600 / 37 = 16.2, so 17 sales per month.
Expected sales S = 25 per month, based on: 5 buyer answers said they would pay
(estimate, not a count of customers); channel reach is unknown.
Margin of safety = (25 - 17) / 25 = 32%.
```

Here S has a basis, so the verdict can be Profitable. In the example, the basis is
weak, so a careful reviewer would still write "estimate, needs a test".

## Verdict rule

Apply in this order and stop at the first match.

1. **Incomplete**: any input is missing, S has no stated basis, or a required stage
   file is absent.
2. **Not yet**: `m <= 0`, or `S < B`. Name the failing number.
3. **Profitable**: `m > 0` and `S >= B`, where S has a stated basis, even when that basis
   is an estimate. Label it as an estimate and state the margin of safety.

The rule does not weigh the board's votes or the team's enthusiasm. A Profitable
verdict is a statement about the numbers only. It is not a recommendation to
launch.
