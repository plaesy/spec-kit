---
name: plan
description: >-
  Write the complete implementation plan for a topic: how each requirement in its
  spec will be met and verified, with ordered steps, dependencies, risks, a
  rollback path, and a go/no-go gate. Use when the user asks for a plan, an
  implementation approach, or a rollout, after the spec exists and before any code
  changes.
---

# plan

A plan says how the spec's requirements will be met and how each one will be
verified. It is a review checkpoint: nothing is implemented, and the plan is
approved by a person before a task moves to `doing`. The plan is written to
`.plaesy/plan/<topic>/plan.md`.

## Rules

- The plan starts from `.plaesy/spec/<topic>/spec.md`. Every requirement R1..Rn in
  the spec appears in the requirement map with at least one step and one
  verification. A requirement with no verification is not planned.
- Every step has an ID, the files it changes, the requirements it serves, and the
  command or check that proves it. A file the plan creates is marked `NEW`.
- Dependencies are written as step IDs. Reject an unknown ID, a duplicate ID, a step
  that depends on itself, and a cycle. For a cycle, name the steps in it and ask which
  order is correct; do not ask about anything else.
- A checked box does not prove readiness. Readiness is shown by evidence: a command
  that passed, or a reviewer's recorded decision.
- Every external claim has a source URL and the date it was accessed. An unsourced
  claim is labeled `ASSUMED` or `UNKNOWN`.
- If there is a real choice between approaches, list each with its trade-off and
  record the one chosen and why.
- The status is `draft` until a person approves it, then `approved`.

## Workflow

1. **Read the spec.** Read `spec.md` in full, and every other spec file in the
   topic folder (`design.md`, `mock.html`, `api.md`, `data.md`, `security.md`).
2. **Read the code the plan touches.** Confirm each existing file path. Mark the
   files that do not exist yet as `NEW`.
3. **Research what the plan needs.** Only what decides an approach or a risk. Use
   the research format in `references/templates.md`.
4. **Choose the approach.** One paragraph, plus alternatives if they were real.
5. **Build the requirement map.** One row per requirement: its steps, its files, its
   verification.
6. **Write the steps.** In dependency order. Mark independent steps `[P]`. Give each
   step one verification.
7. **Plan the UI work when the spec has a design.** Map each screen and state in
   `design.md` to the components and tests that implement it. Each state in
   `mock.html` needs a test.
8. **Write the verification and quality plan.** Unit, integration, end-to-end, and
   accessibility checks, with the command for each.
9. **Write risks, rollback, and the go/no-go gate.** Each risk gets impact,
   probability, an owner, and a mitigation. The rollback says how to undo the change
   and how long it takes.
10. **Write the rollout and communication**, when the change reaches users.
11. **Write the post-implementation validation**: what is checked after release, and
    who signs off.
12. **Report** the plan path, the open questions, and the risks with high impact.
    Do not change code, the spec, or the task folder.

## Sources

The plan structure follows the common implementation-plan templates: a step table
with owners, dependencies, and verification criteria; a risk section with impact,
probability, and mitigation; a rollback plan with go/no-go criteria; and post-release
validation. Examples: the [ClickUp implementation plan template](https://clickup.com/learn/?p=70788)
and the [Playcode implementation plan guide](https://playcode.io/blog/implementation-plan-template),
which requires prerequisite IDs, evidence-bound acceptance criteria, and rejection of
dependency cycles. Public template searches were generic; keep the sections the topic
needs and drop the rest.
