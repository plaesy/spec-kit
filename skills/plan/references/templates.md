# Plan template

Copy this template to `.plaesy/plan/<topic>/plan.md`. Replace every placeholder.
Delete a section only when it does not apply, and say so in one line.

```markdown
# Plan: <topic title>

Status: draft
Spec: ../../spec/<topic>/spec.md
Tasks: <task folder names that use this topic>

## Approach

<one paragraph: what will be built and why this way>

### Alternatives considered

| Option | Trade-off | Chosen |
| --- | --- | --- |
| <option> | <cost, risk, or limit> | yes / no, and why |

## Requirement map

Every requirement in the spec appears here. A row with no verification is a gap.

| Req | Steps | Files | Verification |
| --- | --- | --- | --- |
| R1 | 1, 2 | <path>, <path> (NEW) | `<command>` and what it proves |

## Steps

Ordered by dependency. `[P]` marks a step that can run in parallel with its neighbours.

| ID | Step | Files | Depends on | Verification |
| --- | --- | --- | --- | --- |
| 1 | <what changes> | <path> | none | `<command>` |
| 2 | <what changes> | <path> | 1 | `<command>` |
| 3 [P] | <what changes> | <path> | 1 | `<command>` |

Checks on this table: every ID is unique, every dependency names an existing ID,
no step depends on itself, and the dependency graph has no cycle.

## UI implementation (only when the spec has design.md)

| Screen or state (from design.md / mock.html) | Component | Test |
| --- | --- | --- |
| <screen>: empty | <component path> | <test name> |
| <screen>: error | <component path> | <test name> |

## Quality and verification

- Unit: <what, command>
- Integration: <what, command>
- End-to-end: <flow, command>
- Accessibility: <tool or manual check, target level from design.md>
- Performance: <budget from the spec, how it is measured>

## Risks

| Risk | Impact | Probability | Owner | Mitigation |
| --- | --- | --- | --- | --- |
| <what could go wrong> | high / medium / low | high / medium / low | <person or role> | <what is done about it> |

## Rollback

- Trigger: <the observation that means roll back>
- Steps: <how the change is undone, in order>
- Time to roll back: <estimate>
- Data: <what happens to data written by the change>

## Go / no-go gate

Go when every item is true:

- [ ] Every requirement's verification passed, with the command output recorded.
- [ ] Every high-impact risk has a mitigation in place.
- [ ] The rollback steps were rehearsed or reviewed by a second person.
- [ ] The approver named below has approved.

## Rollout and communication (when users are affected)

- Stages: <for example internal, then 10%, then all>
- Who is told, when, and how: <audience, channel, date>
- Support notes: <what support needs to know>

## Post-implementation validation

- What is checked after release: <metrics, logs, flows>
- Window: <how long to watch>
- Sign-off: <person or role>

## Assumptions and open questions

- ASSUMED — <default>: <why>
- UNKNOWN: <question>. Owner: <person or role>.

## Approval

- Approved by: <name or role>
- Date: <YYYY-MM-DD>

## Revision log

- <YYYY-MM-DD>: <what changed and why>
```

## Research notes

When research shapes a decision, record it under the step it affects:

```markdown
- <claim> — Source: <URL>, accessed <YYYY-MM-DD>
```

## Checklist before status `draft` is reported

- [ ] Every requirement in the spec appears in the requirement map.
- [ ] Every step has an ID, files, dependencies, and a verification.
- [ ] Every file path exists or is marked NEW.
- [ ] Every dependency names an existing step, with no cycle.
- [ ] Every risk has impact, probability, owner, and mitigation.
- [ ] Rollback and the go/no-go gate are written.
- [ ] Every external claim has a source and an access date.
- [ ] No code, spec, or task folder was changed.
