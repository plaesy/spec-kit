# Spec: {{TASK_TITLE}}

Task folder: `.plaesy/tasks/{{STATUS}}/{{TASK}}/`
Status: {{SPEC_STATUS}} (draft | reviewed | final)
Owner: {{OWNER}}
Updated: {{DATE}}

## Goal

{{GOAL_ONE_SENTENCE}}

## Background

{{WHY_THIS_TASK_EXISTS}}

Source of the request: {{SOURCE_OR_NA}}

## Requirements

Each requirement has an ID (R1, R2, …) and an observable acceptance criterion. Do not renumber a requirement after a plan refers to it. Retire it with status `dropped` and add a new ID.

Uncertainty: use `[NEEDS CLARIFICATION: specific question]` inline in any requirement or acceptance criterion when the answer is not known and no reasonable default exists. Cap at 3 markers per spec, prioritized by impact (scope > security/privacy > UX > technical details). Resolve all markers before planning.

| ID | Requirement | Acceptance criterion (observable) | Status |
|---|---|---|---|
| R1 | {{REQUIREMENT_1}} | {{CRITERION_1}} | open |
| R2 | {{REQUIREMENT_2}} | {{CRITERION_2}} | open |

## Non-goals

- {{NON_GOAL_1}} (source of request: {{SOURCE_OR_NA}})

## Success criteria

Measurable outcomes, each with an ID (SC-1, SC-2, …).

- SC-1: {{MEASURABLE_OUTCOME_1}}
- SC-2: {{MEASURABLE_OUTCOME_2}}

## Key entities (data-bearing features only)

- {{ENTITY_1}}: {{FIELDS_AND_RELATIONSHIPS}}

## Constraints

- Thresholds that apply: {{THRESHOLDS_OR_NA}} (source: `.plaesy/memory/constitution.md`)
- {{CONSTRAINT_1}}

## References

Keep only the lines that apply. Each link must point to a file that exists.

- Research: {{RESEARCH_LINK_OR_NA}}
- Design tokens (UI work only): {{DESIGN_LINK_OR_NA}}
- Decisions: {{DECISION_LINK_OR_NA}}

## Open questions

- {{QUESTION_1}}: owner {{OWNER_OR_UNKNOWN}}, due {{DATE_OR_UNKNOWN}}

## Sign-off baseline

Version: {{BASELINE_VERSION}} (e.g. v1.0)
Review log (requirement ID, problem, decision, owner, due date):

| ID | Problem | Decision | Owner | Due |
|---|---|---|---|---|
| R1 | {{PROBLEM}} | {{DECISION}} | {{OWNER}} | {{DATE}} |

Sign-off result: {{ACCEPTED | ACCEPTED_WITH_CHANGES | REVIEW_AGAIN}}

## Evidence rules

- Every fact from outside the repo is written as `[claim] — Source: <URL>, retrieved {{DATE}}`.
- A fact that is not verified is labeled `ASSUMED — <default>` or `UNKNOWN`. Do not present it as settled.
- Do not invent requirement IDs, file paths, command names, or numbers. If a value is not known, write `UNKNOWN`.
