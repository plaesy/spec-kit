# Spec templates

Copy the template for each file into `.plaesy/spec/<topic>/`. Replace every
bracketed placeholder. Delete a section only when it does not apply, and say so in
one line rather than leaving it empty.

## spec.md

```markdown
# Spec: <topic title>

Status: draft
Topic: <topic>
Tasks: <task folder names that use this topic>

## Purpose

<one paragraph: the problem and who has it>

## Requirements

| ID | Requirement | Acceptance criterion (Given / When / Then) | Status |
| --- | --- | --- | --- |
| R1 | <what must be possible> | Given <state>, when <action>, then <observable result> | open |

## Non-goals

- <what this topic will not do, even though a reader may expect it>

## Constraints

- <limits the work must respect: performance, accessibility level, browsers, legal, data>

## Open questions

- UNKNOWN: <question>. Owner: <person or role>.
- ASSUMED — <default>: <why this default was chosen>.

## References

- <path to a file that exists, with a one-line reason>
```

## design.md (any user interface)

```markdown
# Design: <topic title>

Status: draft
Spec: spec.md
Mock: references/<page>.html (one per page)

## Screens

| Screen | Purpose | Entry point | Requirements |
| --- | --- | --- | --- |
| <name> | <what the user does here> | <how they get here> | R1, R3 |

## States per screen

For each screen, list every state the user can see:

- Empty: <what is shown and what the user can do>
- Loading: <what is shown while waiting>
- Error: <message wording and recovery action>
- Success: <what confirms the action>

## Layout and responsiveness

- Breakpoints: <for example 360px, 768px, 1280px>
- Behaviour below the smallest breakpoint: <stack, scroll, or hide>

## Content

- Labels, button text, and error messages in the order the user reads them.

## Accessibility

- Target level: <for example WCAG 2.2 AA>
- Keyboard order: <for each screen>
- Focus and status messages: <where they appear>
- Colour contrast: <checked or to be checked>

## Visual system

- Colours: <tokens or hex values, with their use>
- Type: <font family, size scale>
- Spacing: <base unit>
- Components: <buttons, fields, lists, and their states>

## Decisions and open questions

- <decision, with the requirement it serves>
- UNKNOWN: <question>.
```

## references/<page>.html (one per page, any user interface)

One file per page, named after the page: `references/login.html`,
`references/dashboard.html`. It opens from disk with no network access, no build
step, and no external images or fonts. Start from `../assets/mock.template.html`.
Each state of that page from `design.md` appears as a section.

## flow.md (any user interface with more than one screen)

```markdown
# Flow: <topic title>

Status: draft
Spec: spec.md

## Entry points

- <where the user starts: first screen, link, notification, or deep link>

## Steps

| Step | Screen | User action | Next step |
| --- | --- | --- | --- |
| 1 | <screen> | <action> | 2 |

## Decision points

- At <step>: <what decides the path, such as a permission, a setting, or data>
  - Option A: <path>, goes to <step>
  - Option B: <path>, goes to <step>

## Exit points

- Success: <screen and what the user sees>
- Abandon: <what happens to any entered data>
- Error end: <error and where the user ends up>

## Error and empty states

- <state>: <what the user sees and the way out>
```

## openapi.yaml (any HTTP API)

Write the machine-readable contract for every HTTP endpoint in `api.md`. Keep the two
in step: every endpoint in `api.md` appears here, with the same method, path, and
status codes. Use OpenAPI 3.1.

```yaml
openapi: 3.1.0
info:
  title: <topic title>
  version: 0.1.0
paths:
  /v1/<resource>:
    post:
      summary: <what the endpoint does>
      operationId: <verbNoun>
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [<field>]
              properties:
                <field>:
                  type: string
      responses:
        "200":
          description: <success>
        "400":
          description: <when this happens, and what the caller does>
components:
  schemas: {}
```

## api.md (any interface other systems call)

```markdown
# API: <topic title>

Status: draft
Spec: spec.md

## Endpoints or events

| Name | Method or type | Path or topic | Requirements |
| --- | --- | --- | --- |
| <name> | POST | /v1/<resource> | R2 |

## Request

- <fields, types, which are required, limits>

## Response

- Success: <status and body>
- Errors: <each error code, when it happens, and what the caller should do>

## Compatibility

- Versioning: <rule>
- Breaking changes: <what requires a new version>

## Examples

- <one request and one response, as they would appear on the wire>
```

## data.md (stored data or migration)

```markdown
# Data: <topic title>

Status: draft
Spec: spec.md

## Entities

| Entity | Fields | Required | Requirements |
| --- | --- | --- | --- |

## Relationships (ERD)

| From | To | Cardinality | Participation | Enforced by |
| --- | --- | --- | --- | --- |
| <entity> | <entity> | one-to-many | from: mandatory; to: optional | <constraint, trigger, or application code> |

Cardinality is the number of related rows allowed on each side. Participation says
whether a row must have a related row (mandatory) or may have none (optional). Write
where each rule is enforced, because a database does not always enforce cardinality
by itself.

## Rules

- <validation, uniqueness, retention, deletion>

## Migration

- Existing data: <what happens to it>
- Rollback: <how the change is undone>
```

## security.md (security or personal data)

```markdown
# Security: <topic title>

Status: draft
Spec: spec.md

## Data classified

| Data | Classification | Where it is stored | Who can read it |
| --- | --- | --- | --- |

## Threats considered

- <threat, and the requirement or control that addresses it>

## Authentication and authorization

- <who may do what, and how that is checked>

## Logging

- <what is logged, and what must never be logged>
```

## Checklist before status `draft` is reported

- [ ] Every requirement has an ID, a Given/When/Then criterion, and a status.
- [ ] Every criterion names a precondition, a trigger, and an observable result.
- [ ] Non-goals are written.
- [ ] Every gap is labeled `UNKNOWN` or `ASSUMED`.
- [ ] Every file the work needs exists (design.md, flow.md, one mock per page in references/, api.md and openapi.yaml for an HTTP API, data.md, security.md).
- [ ] Every page in design.md has its own file in references/, with each of its states.
- [ ] Every reference points to a file that exists.
- [ ] No plan, code, or contract was written.
