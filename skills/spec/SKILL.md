---
name: spec
description: >-
  Write the complete spec for a topic: requirements R1..Rn with Given/When/Then
  acceptance criteria, non-goals, constraints, and open questions. Adds design.md
  and one self-contained HTML mock per page for any work with a user interface, and api.md
  for any work with an interface other agents or services call. Use when the user
  asks to write a spec, define requirements, or describe what a feature must do,
  before any plan or code.
---

# spec

A spec says what must be true when the work is done. It does not say how to make
it true; that is the plan. Every task belongs to a topic, and the topic's spec is
written to `.plaesy/spec/<topic>/`.

## Rules

- Write only from the request, the existing spec, and the files it names. Do not
  invent a requirement the request did not ask for.
- Every requirement has an ID (R1, R2, ...) and one acceptance criterion in
  Given/When/Then form: a precondition, a trigger, and an observable result. If a
  tester would need to ask a question to test it, it is not finished.
- Gaps are labeled, never filled silently: `UNKNOWN` (no answer yet) or
  `ASSUMED — <default>` (a safe default applied, to be confirmed).
- An `ASSUMED` default or an `UNKNOWN` answer is recorded under Open questions only.
  It never becomes a numbered requirement (R-row). A numbered requirement must come
  from the request or from an existing spec entry.
- Every spec has a `## Non-goals` heading, even when the list says `- none stated`.
  Without it, the work grows to cover neighbours.
- A removed requirement stays in the table marked `dropped`; IDs are never reused.
- The status is `draft`. Only a person changes it to `reviewed` or `final`.

## Which files to write

Write `spec.md` always. Add another file only when the request names that need. Do not
add `api.md`, `data.md`, or `security.md` for a plain screen with no API, stored data,
or personal data:

| Work involves                                   | Files in `.plaesy/spec/<topic>/`     |
| ----------------------------------------------- | ------------------------------------ |
| Any screen, page, or visible component          | `design.md`, `flow.md`, and one `references/<page>.html` per page |
| An HTTP API that another system calls          | `api.md` (readable summary) and `openapi.yaml` (machine contract) |
| An event or other contract (not HTTP)           | `api.md`                             |
| Stored data, a schema, or a migration           | `data.md` (entities and an ERD table) |
| Anything with security or personal data         | `security.md`                        |

A frontend topic without `design.md` and one HTML mock per page is incomplete. A mock
is complete when it shows its page with every state (empty, loading, error, success),
and it opens in a browser with no network access. The mocks live in
`.plaesy/spec/<topic>/references/`, named after the page (for example `login.html`).
Start from `assets/mock.template.html` in this skill.

## Workflow

1. **Read the source.** The request, the existing `spec.md` if any, and the files
   the request names. Nothing else is a source of requirements.
2. **Write the requirements.** One row per requirement in the table in
   `references/templates.md`. Split a request with several outcomes into several
   requirements.
3. **Write the acceptance criteria** in Given/When/Then. Check each one for a
   precondition, a trigger, and an observable outcome.
4. **Record non-goals, constraints, and open questions.** Constraints are limits the
   work must respect (performance, accessibility level, supported browsers, legal).
5. **Write the design, if the work has a user interface.** Use the design template
   in `references/templates.md`. Then build one mock per page,
   `references/<page>.html`. Every page named in `design.md` must have its own file, and
   every state listed for that page (empty, loading, error, success, and any others)
   must have its own section in that file. Before reporting, compare the lists and add
   any missing file or section.
6. **Write `flow.md` for any user interface with more than one screen, and `api.md`, `openapi.yaml`, or `data.md`** when they apply, using the templates in
   `references/templates.md`.
7. **Check the spec** against the checklist at the end of `references/templates.md`.
8. **Report** the paths written, the number of requirements, and the open
   questions. Do not write a plan or change code.

## Sources

Templates and criteria follow current practice: a technical spec with requirements,
non-goals, and open questions ([spec-coding-site guide](https://spec-coding-site.pages.dev/blog/how-to-write-technical-spec-template-guide)),
and acceptance criteria written as preconditions, triggers, and observable results
in the same guide. The interface and mock sections follow the common practice of
showing each screen and state as the user sees it. Public template searches were
generic; treat the templates as a starting point and keep what the topic needs.
