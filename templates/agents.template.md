---
description: "Short description of the agent's role and purpose"
---

# {{ROLE_TITLE}} Agent

## Role Definition (RACE Framework)

**Role**: Brief role description.

**Action**: Primary actions the agent should take.

**Context**: Constraints and the constitutional context.

**Execute**: Expected deliverables and behavior.

## Project Rules & Constitutional Precedence

The articles that bind every role are defined once in
`.plaesy/memory/constitution.md`, which overrides anything written here.
Nothing below is a constitutional article — it is this role's working scope;
a rule not in the constitution does not get to borrow its authority.

- List of project rules this role applies (name any that the constitution
  also governs as such, don't restate the article text)

## Response Style & Behavior

- Communication and approach guidance

## Key Capabilities

- Bullet list of capabilities

## Boundaries & Escalation

- **Owns**: The decisions this role makes without escalating
- **Defers to**: The roles this role hands work to, and what it hands over
- **Escalate to @{{DEFERRED_ROLE}} when**: A single binary condition that triggers the handoff

## Memory Protocol

Read `.plaesy/memory/roles/[this-role].md` (if it exists) at the start of a
turn; append one dated entry at the end before handing off. Full protocol:
`.plaesy/instructions/plaesy.md` → Memory Hierarchy → Role Memory. Identical
across all role files by design — don't rewrite it per role.

## Subagent Invocation Contract

When an orchestrator (`/implement`, `/loop`, or another role) dispatches this
role as a subagent instead of a human reading it directly:

- **Tools granted**: whatever the dispatching orchestrator's task grants —
  this role assumes no access beyond what it was handed, and names what it
  actually used in its summary.
- **Input shape**: a task brief — objective, relevant files/context, expected
  output format, and any constraint from Boundaries & Escalation above that
  narrows scope for this invocation.
- **Fresh context per invocation**: each dispatch starts from this file plus
  the brief, not cumulative history from a prior invocation — Memory Protocol
  above is the one exception, read explicitly, not inherited automatically.
- **Output contract**: a structured summary — done / not done / blocked /
  needs escalation (per Boundaries & Escalation) — not freeform narration.
  The orchestrator, not this role, talks to the user (see
  `.plaesy/instructions/multi-agent-patterns.md` → Single Response
  Principle).

## Example Use Cases

- Short scenario description

## Example

- Input: "{{EXAMPLE_USER_PROMPT}}"
- Expected Output Format: `{{{PREFERRED_RESPONSE_FORMAT}}}`
- Output: "{{EXAMPLE_AGENT_REPLY}}"
