---
name: spec-marketing
description: >-
  Write the marketing spec for a topic: audience, message, channels, campaign goals, and measures, with requirements in R1..Rn and Given/When/Then criteria. Use when the user asks for a marketing spec, a campaign brief, or go-to-market requirements.
---

# spec-marketing

A marketing spec says what must be true for the marketing side of a topic. It does not say how
to achieve it; that is the plan. The spec is written to `.plaesy/spec/<topic>/spec.md`.
It sits alongside the software spec (`spec`) when a topic has both.

## Rules

- Write only from the request and the files it names. Do not invent a requirement.
- Every requirement has an ID (R1, R2, ...) and a Given/When/Then criterion: a
  precondition, a trigger, and an observable result.
- Unknowns are labeled `UNKNOWN` (no answer yet) or `ASSUMED — <default>` (a default
  applied, to be confirmed). They go under Open questions, never as numbered
  requirements.
- Every spec has a `## Non-goals` heading.
- Claims about markets, prices, laws, or competitors need a source URL and the date
  accessed. Without both, the claim is `ASSUMED`.
- A value nobody has given (an owner, a number, a date) is written as `UNKNOWN`, never as
  `TBD` or a guess.
- Status is `draft`. Only a person changes it to `reviewed` or `final`.

## Sections

- Audience
- Message and positioning
- Channels and budget
- Campaign goals and measures
- Claims that need evidence

## Template

```markdown
# Spec: <topic title>

Status: draft
Topic: <topic>

## Purpose

<one paragraph>

## Requirements

| ID | Requirement | Acceptance criterion (Given / When / Then) | Status |
| --- | --- | --- | --- |
| R1 | <requirement> | Given <state>, when <action>, then <observable result> | open |

## Non-goals

- <what this topic will not do>

## Open questions

- UNKNOWN: <question>. Owner: <person or role>.
- ASSUMED — <default>: <why>.

## References

- <URL>, accessed <YYYY-MM-DD>
```

## Workflow

1. Read the request and any existing spec for the topic.
2. Write the requirements, one row each, under the sections above.
3. Write each acceptance criterion in Given/When/Then and check it has a precondition,
   a trigger, and an observable result.
4. Record non-goals and open questions.
5. Check the list against the rules, then report the path, the number of requirements,
   and the open questions. Do not write a plan or change code.
