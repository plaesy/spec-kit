# Board: three lenses

The board is three reviewers. Each reads the same brief and applies only its own
lens. Each writes a short section in `.plaesy/ideas/<topic>/board.md`. The lenses summarize
published ideas; the reviewers never speak as the people those ideas come from.

## The brief

Write `.plaesy/ideas/<topic>/board/brief.md` first, from `.plaesy/ideas/<topic>/idea.md` and anything else the
user said that a stranger would need. The reviewers cannot see the conversation,
so the brief must stand alone. Do not add your own opinion of the idea to it.

## Lenses

**Offer lens** (from the idea of a "grand slam" offer: a promise so strong the
buyer's risk feels small).

- Is the promise specific and worth paying for?
- What does the buyer get, how soon, and what do they risk?
- Would a buyer compare this with doing nothing? Does it beat that?

**Defensibility lens** (from the idea of a monopoly niche: a market small enough to
own).

- Who could copy this in a month, and why would they not?
- What is unique: location, relationship, data, skill, license, or nothing?
- What would make a rich competitor enter, and when?

**Product lens** (from the idea that a product should do one thing very well).

- Does the first version solve one problem end to end?
- What is the smallest version that could be sold this month?
- What would the user do in the first five minutes, and does it work?

## Each reviewer's output

```markdown
## <Lens name>

Score: <1 to 5>
Kill risks: <up to three, each one sentence>
Vote: fund | fund with conditions | pass
Conditions (if any): <what must be true>
```

## Combining votes

- Two or more **pass**: the verdict cannot be Profitable. Record the votes in
  `.plaesy/ideas/<topic>/plan.md`; the economics can still be computed.
- Otherwise record the votes and the conditions as they are. Do not average the
  scores into a verdict; the verdict comes from `.plaesy/ideas/<topic>/economics.md`.
