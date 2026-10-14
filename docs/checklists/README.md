# Checklists

Structured validation checklists for each development phase and role. Work through the relevant checklist's items, address anything unchecked, then pass its quality gates before moving to the next
phase.

## Index

| Checklist | Phase / Role | Focus |
|-----------|--------------|-------|
| [story-draft.checklist.md](../../checklists/story-draft.checklist.md) | Story creation | Requirements, acceptance criteria, INVEST compliance |
| [update.checklist.md](../../checklists/update.checklist.md) | Change updates | Impact assessment, rollback/re-scope options |
| [story-done.checklist.md](../../checklists/story-done.checklist.md) | Story completion | Definition of Done, testing, documentation |
| [pm.checklist.md](../../checklists/pm.checklist.md) | Product Manager | Problem, MVP scope, PRD, epic structure |
| [po.checklist.md](../../checklists/po.checklist.md) | Product Owner | Vision, backlog hygiene, prioritization, stories, acceptance criteria, value metrics, Definition of Ready |
| [sa.checklist.md](../../checklists/sa.checklist.md) | Solutions Architect | Architecture, tech choices, project setup, dependencies, security, performance, deployment & DR |
| [qa.checklist.md](../../checklists/qa.checklist.md) | QA Engineer | Test strategy, test types, security & OWASP coverage, quality gates, sign-off |

Ownership boundary: a checklist validates only the dimension it owns. Infrastructure,
scaffolding, dependency setup, deployment, and observability belong to `sa.checklist.md`;
backlog, prioritization, acceptance criteria, and value delivery belong to `po.checklist.md`;
test strategy and sign-off belong to `qa.checklist.md`. Do not duplicate a dimension across
checklists — cross-reference the owning checklist instead.

## Format

Each checklist is a Markdown file with `- [ ]` items grouped by category, plus quality-gate criteria at the end. AI agents load the checklist matching the current phase, work items top to bottom, and
document evidence for anything checked or explicitly marked N/A. `[[LLM: ...]]` blocks give the agent short execution guidance for that section — they are instructions, not checklist items.

A checklist is a **validator**, not a document. It does not carry a document metadata form, a
signature block, or a metrics worksheet — it holds actionable `- [ ]` items, the guidance
needed to work them, and the gate that closes the section. Numbers a reviewer needs to fill
in belong in the artifact under review, not as `[PLACEHOLDER]` fields here.

Conditional items are marked inline with a bracketed tag (`[[BROWNFIELD ONLY]]`,
`[[FRONTEND ONLY]]`). Each checklist's opening `[[LLM: ...]]` block defines what the tag
means, which project types are detected, and that non-applicable sections are skipped and
reported rather than silently passed.

## Usage

```bash

plaesy features paths          # get current context
plaesy features validate        # validate prerequisites
# work through the relevant checklist
plaesy context update           # record progress
```

## Adding a checklist

1. Confirm the phase/role isn't already covered.
2. Confirm a **command actually loads it** — an unreferenced checklist is never
   executed. Wire it into the referencing `instructions/*.instructions.md` or
   `prompts/*.md` in the same change, and add it to the Index table below.
3. Draft items as specific, testable actions — not vague goals. Every action must
   be expressible as a `- [ ]` item an agent can decide pass/fail.
4. Add an `[[LLM: ...]]` guidance block per section.
5. Add quality-gate criteria and cross-reference the related checklist that owns
   adjacent concerns.
6. Add a row to the Index table above.

## Related

- [../../README.md](../../README.md)
- [../templates/README.md](../templates/README.md)
- [../instructions/README.md](../instructions/README.md)
- Issues/requests: [github.com/plaesy/spec-kit/issues](https://github.com/plaesy/spec-kit/issues)
