# Checklists

Structured validation checklists for each development phase and role. Work through the relevant checklist's items, address anything unchecked, then pass its quality gates before moving to the next phase.

## Index

| Checklist | Phase / Role | Focus |
|-----------|--------------|-------|
| [story-draft.checklist.md](../checklists/story-draft.checklist.md) | Story creation | Requirements, acceptance criteria, INVEST compliance |
| [update.checklist.md](../checklists/update.checklist.md) | Change updates | Impact assessment, rollback/re-scope options |
| [story-done.checklist.md](../checklists/story-done.checklist.md) | Story completion | Definition of Done, testing, documentation |
| [pm.checklist.md](../checklists/pm.checklist.md) | Project Manager | Planning, resourcing, risk, delivery |
| [po.checklist.md](../checklists/po.checklist.md) | Product Owner | Vision, backlog, value delivery |
| [sa.checklist.md](../checklists/sa.checklist.md) | Solutions Architect | Architecture, tech choices, security, performance |
| [qa.checklist.md](../checklists/qa.checklist.md) | QA Engineer | Quality gates, testing, sign-off |

## Format

Each checklist is a Markdown file with `- [ ]` items grouped by category, plus quality-gate criteria at the end. AI agents load the checklist matching the current phase, work items top to bottom, and document evidence for anything checked or explicitly marked N/A. `[[LLM: ...]]` blocks give the agent short execution guidance for that section — they are instructions, not checklist items.

## Usage

```bash
./bash/get-feature-paths.sh          # get current context
./bash/check-task-prerequisites.sh   # validate prerequisites
# work through the relevant checklist
./bash/update-agent-context.sh       # record progress
```

## Adding a checklist

1. Confirm the phase/role isn't already covered.
2. Draft items as specific, testable actions — not vague goals.
3. Add quality-gate criteria and cross-reference related checklists.
4. Add a row to the Index table above.

## Related

- [../README.md](../README.md)
- [../templates/README.md](../templates/README.md)
- [../instructions/README.md](../instructions/README.md)
- Issues/requests: [github.com/plaesy/spec-kit/issues](https://github.com/plaesy/spec-kit/issues)
