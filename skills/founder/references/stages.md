# Stage templates

Each template is the starting point for one file in `.plaesy/ideas/<topic>/`. Replace the bracketed
text. Delete any line the user has no information for, and write "unknown" rather
than guessing.

## .plaesy/ideas/<topic>/idea.md

```markdown
# <working name>

- What it is: <one sentence>
- Who it is for: <the target customer, as specific as the user can make it>
- What it sells, at what price: <product and price>
- Where and how: <city or online; how customers find it; how it is delivered>
- Budget and constraints: <money, time, skills, decisions already made>
```

## .plaesy/ideas/<candidate-topic>/idea.md (candidate)

Used during idea discovery. One folder per candidate. Replace `candidate` with
`chosen` in the status line when the user picks it.

```markdown
# <working name>

status: candidate
Found: <YYYY-MM-DD>
Interests and constraints given: <interests, skills, budget, time, where to sell>

- Problem: <the problem, in the customer's words>
- Who pays: <the buyer, and how many there may be, with a source or "unknown">
- Why now: <what changed, with a source URL and the date accessed>
- Existing sellers: <names and links>
- Unknown: <what must be tested>
```

## .plaesy/ideas.md (index line)

One line per idea, flat list, no sub-bullets. Link only to files that exist.

```markdown
- [<working name>](ideas/<topic>/plan.md) — <status: candidate | chosen | verdict: Profitable | Not yet | Incomplete> · <YYYY-MM-DD>
```

Link to `idea.md` instead of `plan.md` until the plan exists.

## .plaesy/ideas/<topic>/competitors.md

```markdown
| Alternative | Price (as published) | Promise | Source | Accessed |
| --- | --- | --- | --- | --- |
| Do nothing | 0 | <what the customer lives with> | - | - |
| <name> | <price, or "not published"> | <promise> | <link> | <YYYY-MM-DD> |
```

## .plaesy/ideas/<topic>/buyers.md

```markdown
| Date | Person (role, not name, if private) | Question asked | Answer | Would pay? |
| --- | --- | --- | --- | --- |
```

Count only "yes, and would pay at this price" as a buyer signal, and say how many
conversations there were in total.

## .plaesy/ideas/<topic>/pricing.md

```markdown
- Proposed price: <P>
- Compared with: <the competitors and their prices>
- Buyer evidence: <what the conversations said about price>
- Recommendation: <keep, raise, or lower, with the reason>
```

## .plaesy/ideas/<topic>/offer.md

```markdown
- Outcome: <what the customer gets, in their words>
- Guarantee: <what happens if it does not work>
- Included: <the list>
- Excluded: <what is not included>
```

## .plaesy/ideas/<topic>/marketing.md

```markdown
- Starting channel: <one channel>
- Why this channel: <evidence, or "untested">
- Message: <one sentence>
- First test: <what to post, send, or ask, and how many>
```

## .plaesy/ideas/<topic>/brand.md

```markdown
- Name: <name, and whether it is free to use, with a source>
- Tone: <three words>
- What it must not sound like: <one line>
```

## .plaesy/ideas/<topic>/ops.md

```markdown
- What must be true to deliver the offer: <list>
- Who does it: <role or person>
- Capacity: <how many sales per month this can handle>
```

## .plaesy/ideas/<topic>/launch.md

```markdown
- First 30 days, in order: <steps>
- Smallest test: <what it is, and what result would show demand exists>
- Stop rule: <the result that means do not continue>
```

## .plaesy/ideas/<topic>/plan.md

```markdown
# <working name>: business plan

**Verdict:** <Profitable | Not yet | Incomplete>
**Numbers behind it:** <margin per sale, break-even per month, expected sales per month, and the basis for expected sales>
**Board votes:** <one line per lens>

## Summary (one page)

<what it is, who it is for, the offer, the price, the verdict in plain words>

## Stages

<the contents of each stage file, in order, with its heading>
```
