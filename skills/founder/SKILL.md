---
name: founder
description: >-
  Find or test a business idea before launch. Suggests candidate ideas from
  researched market gaps when the user has none, then walks the chosen idea
  through a board review, competitor and buyer research, pricing and offer, unit
  economics, marketing, brand, operations and launch, and compiles one business
  plan whose verdict is computed from the numbers. Use when the user says "find me
  a business idea", "suggest ideas", "test my business idea", "is this worth
  building", "write my business plan", "pressure-test this idea", or asks whether
  a business is profitable before they start.
argument-hint: "<the business idea, or leave blank to use the conversation>"
---

# founder

Take one business idea from a rough question to a plan with a verdict. Each idea
gets its own folder, `.plaesy/ideas/<topic>/`, in the user's project. Each stage
writes one file into that folder, and later stages read the earlier files.
Templates for each stage are in `references/`.

`<topic>` is the idea's name in kebab-case, for example `meal-prep-office`. Every
idea is listed in `.plaesy/ideas.md`, one line per idea, so all ideas can be found
in one place. Create that file with a `# Ideas` heading if it does not exist.

The user's idea is in `$ARGUMENTS`. If that is blank, take it from the
conversation.

## Rules

- Do not invent a budget, market size, price, or competitor. If the user has not
  given a number, ask for it or mark the figure as an estimate with its basis.
- Cite a source link for every competitor and price. A figure without a source is
  an estimate and must say so.
- A simulated buyer is not a customer. Never present simulated or survey-style
  answers as real demand.
  When the user asks for simulated buyers to show demand, decline that use in the first
  sentence, say that a simulation is not demand, and offer real conversations instead.
  Do not generate the simulation for that purpose. In that reply, state the number
  of real conversations recorded so far, as 0 when none exist. Do not ask the user for it.
- Name the lens or the numbers, never a real person, when giving a view. Write
  "the offer lens says" or "the numbers say", not "X would endorse this".
- When input is missing, do not stop. Ask the questions in the reply, record the
  missing item as `UNKNOWN`, finish every stage that does not need it, and give the
  verdict. A missing fixed cost or sales basis makes the verdict `Incomplete`.
- When web search is not available, say so in the reply and mark every market claim
  `UNKNOWN`. Do not refuse the task.
- Each stage fails loudly on missing input. Ask for the missing item; do not guess.
- Every market claim from the web (size, trend, gap, price) needs a source URL and
  the date it was accessed. A claim without both is written as an assumption.
- Suggested ideas are candidates, not recommendations. The user chooses one.

## Workflow

Run the stages in order. Skip a stage only when the user has already done it, and
say so in `.plaesy/ideas/<topic>/plan.md`.

1. **Idea discovery** (only when the user has no idea yet; template in
   `references/stages.md`). Ask at most three questions in one message:
   interests or skills, budget and time, and where they can sell (city or online).
   Then search the web for problems people pay to solve in that area: at least
   three queries, and open the sources rather than relying on search snippets.
   Write three candidate ideas, each in its own folder
   `.plaesy/ideas/<candidate-topic>/idea.md` with `status: candidate`. For each,
   give the problem, who pays, why it may be possible now (with sources), who
   already sells something similar, and what is still unknown. Add one index line per
   candidate to `.plaesy/ideas.md`. Stop and let the user choose one. Then set
   that candidate's `status` to `chosen`, update its index line, and continue to
   step 2 with it. Leave the other candidates as they are.
2. **Idea file** (`.plaesy/ideas/<topic>/idea.md`, template in `references/stages.md`). Ask for
   missing items, at most three questions in one message: name, customer, product
   and price, delivery, budget and constraints.
3. **Board** (`.plaesy/ideas/<topic>/board.md`, see `references/board.md`). Three reviewers each
   apply one lens to the idea: offer, defensibility, product. Each names the risks
   that would kill the idea and votes **fund**, **fund with conditions**, or
   **pass**. Write the brief first so a reviewer can read it alone.
4. **Competitors** (`.plaesy/ideas/<topic>/competitors.md`, template in `references/stages.md`).
   Research before writing:
  - Search the web for the problem and for "<product type> alternatives" and
     "<product type> price", at least three queries.
  - Open each alternative's own pricing or product page. Record the price as shown
     there, with its link and the date accessed.
  - If the price is not published, write "not published" and do not estimate it.
  - Include "do nothing" as the first row.
5. **Buyers** (`.plaesy/ideas/<topic>/buyers.md`). Record at least five real conversations in the
   target group, with date, question asked, and answer. If there are none, write
   "no real interviews yet" and continue; the verdict will reflect it. State the number of real
   conversations so far in buyers.md and in the reply, as 0 when there are none. Simulated
   buyers, if written at all, go in a separate section labeled "simulated, not demand".
6. **Pricing and offer** (`.plaesy/ideas/<topic>/pricing.md`, `.plaesy/ideas/<topic>/offer.md`). Compare the
   price with the competitors and the buyer answers. Write the offer as outcome,
   guarantee, and what is included.
7. **Economics** (`.plaesy/ideas/<topic>/economics.md`, method in `references/economics.md`).
   Compute contribution margin, break-even volume, and the expected sales estimate
   with its basis. Show the arithmetic.
8. **Marketing, brand, operations** (`.plaesy/ideas/<topic>/marketing.md`, `.plaesy/ideas/<topic>/brand.md`,
   `.plaesy/ideas/<topic>/ops.md`). One starting channel, the message, the name and tone, and
   what must be true to deliver the offer.
9. **Launch** (`.plaesy/ideas/<topic>/launch.md`). The first 30 days: the smallest test that could
   show demand exists or does not.
10. **Plan** (`.plaesy/ideas/<topic>/plan.md`, see `references/economics.md` for the verdict rule).
   Put the verdict and the numbers that drive it at the top, then a one-page
   summary, then the stage files in order.

## Verdict

The verdict comes from `.plaesy/ideas/<topic>/economics.md`, never from opinion. The rule:

- **Incomplete**: any stage file in the list above is missing, or the expected
  sales estimate has no stated basis.
- **Not yet**: contribution margin is zero or negative, or expected monthly sales
  are below break-even volume. Name the failing number.
- **Profitable**: contribution margin is positive and expected monthly sales are at
  least break-even volume. State the margin of safety.

## Output

Tell the user the verdict, the one or two numbers behind it, and the path to
`.plaesy/ideas/<topic>/plan.md`. Update the idea's line in `.plaesy/ideas.md` with
its verdict and date. Do not repeat the whole plan in chat.
