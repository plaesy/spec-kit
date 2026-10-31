# Plan: {{TASK_TITLE}}

Task folder: `.plaesy/tasks/{{STATUS}}/{{TASK}}/`
Status: {{PLAN_STATUS}} (draft | approved | superseded)
Spec: `.plaesy/spec/<topic>/spec.md` (the task's topic)
Updated: {{DATE}}

A plan is a review checkpoint. Nothing in it is implemented until it is approved, and the task moves to `doing` only after approval.

## Approach

{{APPROACH_ONE_PARAGRAPH}}

Options considered, if more than one was real:

- {{OPTION_A}}: {{TRADE_OFF_A}}
- {{OPTION_B}}: {{TRADE_OFF_B}}

Chosen: {{CHOSEN_OPTION}}, because {{REASON}}.

## Constitution check (gate)

Must pass before research/planning proceeds; re-check after design.

- [ ] {{PRINCIPLE_1}}: {{PASS_OR_FAIL}} — {{EVIDENCE}}
- [ ] {{PRINCIPLE_2}}: {{PASS_OR_FAIL}} — {{EVIDENCE}}

Complexity tracking (if any principle is violated, justify here):

| Violation | Why needed | Simpler alternative rejected |
|---|---|---|
| {{VIOLATION}} | {{REASON}} | {{ALTERNATIVE}} |

## Requirement map

Every requirement in the topic spec appears here. Each row needs at least one step and one verification. Use the same IDs as the topic spec.

| Req | Steps | Files | Verification |
|---|---|---|---|
| R1 | {{STEP_IDS}} | {{FILE_PATHS}} | {{COMMAND_OR_CHECK}} |
| R2 | {{STEP_IDS}} | {{FILE_PATHS}} | {{COMMAND_OR_CHECK}} |

## Steps

Order matters: dependencies come first. Write "(parallel)" on independent steps.

1. {{STEP_DESCRIPTION}}. Files: {{FILE_PATHS}}. Verification: {{COMMAND_OR_CHECK}}.
2. (parallel) {{STEP_DESCRIPTION}}. Files: {{FILE_PATHS}}. Verification: {{COMMAND_OR_CHECK}}.

## Files

Every path in the plan either exists in the repository or is marked `NEW`. Check each one before approval.

| Path | State | Reason |
|---|---|---|
| {{FILE_PATH}} | exists or NEW | {{REASON}} |

## Decisions

- D1: {{DECISION}}. Reason: {{REASON}}. Record: {{DECISION_LINK_OR_NA}}

## Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| {{RISK}} | {{LOW_MEDIUM_HIGH}} | {{MITIGATION}} |

## Assumptions and open questions

- ASSUMED — {{DEFAULT}} — because {{REASON}}
- OPEN — {{QUESTION}} — owner: {{OWNER_OR_UNKNOWN}}

## Subtasks (proposed, not created)

- {{SUBTASK_NAME}}: covers {{REQ_IDS}}; blocked by {{TASK_OR_NONE}}

## Revision log

- {{DATE}}: first draft.

## Evidence rules

- Every fact from outside the repo carries its source and retrieval date, or is labeled `ASSUMED` or `UNKNOWN`.
- Do not invent file paths, command names, requirement IDs, or results. A check that has not been run is not reported as passed.
