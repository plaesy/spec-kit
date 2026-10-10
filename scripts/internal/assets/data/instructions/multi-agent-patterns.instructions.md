---
description: "Subagent communication protocol, specialization patterns, and instruction template for multi-agent orchestration"
applyTo: "**/*"
---

# Multi-Agent Orchestration Patterns

## Subagent Communication Protocol

- **Single Response Principle**: Only the parent agent communicates with the user; subagents return findings only
- **Explicit Role Declaration**: Every subagent instruction must include: "You are a subagent. Do NOT reply to the user directly. Return findings to the parent agent."
- **Data Handoff**: Explicitly define what context passes from parent to subagent
- **Tool-First Design**: Prefer pure function tool calls over MCP for reliability

## Agent Specialization Patterns

| Pattern | Use When | Example |
|---------|----------|---------|
| Sequential Pipeline | Linear dependencies, clear phases | Research → Plan → Implement → Verify |
| Parallel Fan-Out | Independent sub-tasks | Security audit + Performance profile + Code review |
| Handoff/Routing | Domain-specific expertise needed | Frontend agent → Backend agent → Database agent |
| Group Consensus | High-stakes decisions | Architecture review with 3+ specialist agents |

## Background Task Git Safety (Single-Committer)

When more than one background-dispatched command (see
`.plaesy/instructions/plaesy.md` → Long-Running Commands) is active at once,
treat shared-state files as a single-writer resource, same principle as a
database row lock, enforced by convention since there is no database:

- **One writer at a time** — only one active command may run `git commit` or
  edit a shared index file (`.plaesy/memory.md`, `.plaesy/decisions.md`,
  `.plaesy/instructions.md`, any `tasks/*` move) at a time. Pick the writer by
  whichever command reaches that step first; it keeps the role until its
  current write completes.
- **Everyone else proposes, doesn't commit** — a non-writer command finishes
  its own work and stages the result in its own file (its task file, a
  `decisions/[topic].md` draft, a diff not yet applied) instead of touching
  shared state directly. The writer folds it in on its own next pass.
- **Isolate the working tree when both touch source code** — if two
  background commands both edit application source (not just `.plaesy/`),
  give each its own `git worktree` (`git worktree add ../<task-id> -b
  task/<task-id>`) so neither's uncommitted edits collide with the other's;
  merge back through the single writer once each finishes.
- **Stale lock recovery** — if a command crashes mid-write, the next command
  that needs to write checks whether the previous writer's process is still
  alive before waiting on it; a dead writer's claim is reclaimed, not
  honored indefinitely.

## Subagent Instruction Template

```markdown
<instructions>
<role>You are a [specialist role] subagent</role>
<constraint>Do NOT reply to the user directly. Return findings to parent agent only.</constraint>
<task>[Specific task]</task>
<context>[Relevant context from parent]</context>
<output_schema>[Exact JSON/XML schema for response]</output_schema>
<examples>
<example>Input: ... Output: ...</example>
</examples>
</instructions>
```
