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
