---
applyTo: "{{SCOPE_GLOB}}"
description: "{{ONE_SENTENCE_WHAT_THIS_GOVERNS_AND_WHEN_RELEVANT}}"
---

# {{TOPIC_TITLE}}

{{ONE_TO_TWO_SENTENCE_OVERVIEW}}

<!--
Fill this file top to bottom, then delete every HTML comment before committing.

Strategy check first: does this concern need ONE file, or several scoped by
area/topic ("nested")? If {{TOPIC_TITLE}} would push this file past ~300-500
lines, split by concern now instead of writing a long single file.

Frontmatter:

- applyTo: a glob, or comma-separated globs (e.g. "src/**/*.ts,src/**/*.tsx").
  Narrow it to exactly the files this concerns — an overly broad glob makes
  the agent apply the wrong rules to unrelated files.
- description: one sentence, what it covers AND when it's relevant. Not a
  restatement of the title.
-->

## Guidelines

- Use {{X}}. — {{WHY: the reason this beats the default, e.g. a past bug it
  prevents, an invariant it protects}}
- Never {{Y}}. — {{WHY}}

<!--

- Every rule needs its own "why" clause. Reasoning is what lets the agent
  generalize correctly to a case you didn't enumerate; a rule with no reason
  is just a demand.
- Write rules as direct imperatives ("Use X", "Never do Y"), not narrated
  background ("This project uses X").
- Only include what the agent would otherwise get wrong or have to guess —
  exact commands, non-default conventions, gotchas, required ordering. Skip
  anything a competent model, or this project's linter/formatter, already
  gets right.
- Prefer a flat "Never do Y" over a soft "try to avoid Y" for hard rules —
  explicit prohibitions are followed more reliably than open-ended advice.
- Match how strict a rule reads to how fragile the thing it governs is:
  heuristic wording for tasks with many valid approaches, an exact
  command/snippet with zero room to deviate for fragile or destructive ones
  (e.g. "run exactly `{{{COMMAND}}}`, do not add flags").
- One consistent term per concept throughout — don't alternate synonyms.

-->

## Example

```{{LANGUAGE}}
{{ONE_CONCRETE_WORKED_EXAMPLE_MATCHING_A_GUIDELINE_ABOVE}}
```

<!--

- Prefer one real example taken from this repo over an invented one.
- Add a contrasting "Bad:" example only when the wrong pattern is genuinely
  easy to reach for (e.g. the obvious/default API call is the one to avoid).
- If output *style* matters more than correctness (e.g. commit message
  format), give an input -> output pair instead of a bare snippet.
-->

<!--

## Old patterns (optional — delete unless something is deprecated)

Keep deprecated guidance out of the main Guidelines section so it can't be
mistaken for current advice, but don't delete it outright if agents might
still encounter old code following it.

**Old:** {{DEPRECATED_APPROACH}} — no longer used because {{REASON}}.
**Current:** see Guidelines above.
-->

<!--
Before finishing:

- Re-read the applyTo glob against 2-3 real file paths in this repo; confirm
  it matches only the intended set and doesn't overlap/contradict another
  existing *.instructions.md file.
- Cut anything that didn't need its own "why" — if you can't justify a rule,
  it's probably not worth the tokens.
- If this repo has an instructions index (e.g. instructions/mapping.json),
  add an entry for this file.
See instructions/how-to-create-instructionsmd.instructions.md for full guidance.
-->
