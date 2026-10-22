---
title: "Decision: Nara-first routing and no hard stops"
description: "Every request routes through Nara, who decides all questions and rule-8 categories under the Constitution's NN articles; no run stops for the user"
updatedAt: "2026-10-14T06:00:00Z"
decided: "2026-10-14"
status: "Accepted"
supersedes-part: "full-autonomy-role-handoff (rule 8 hard-stop list)"
---

# Decision: Nara-first routing and no hard stops

**Status**: Accepted
**Decided**: 2026-10-14
**Deciders**: project owner (explicit choice of option A), session agent
**Related**: [full-autonomy-role-handoff](full-autonomy-role-handoff.md), [revert-without-history-reset](revert-without-history-reset.md)

## Context and Problem Statement

A project that ran `plaesy init` still stopped and asked the user for decisions. Three causes
were found:

1. `instructions/plaesy.instructions.md` (the always-loaded file) never mentioned Nara. Routing
   rules lived in `role-mapping.md`, which is scope-loaded, so nothing pointed a session at Nara.
2. The `/start` prompt still said "ask once" and "Decision required" (lines 152 and 245 before this change).
3. Rule 8 listed hard stops that block the run. The owner wanted the run never to stop.

## Decision Drivers

- The owner wants every issue handled by Nara, including asking roles such as `security` and
  researching best practice on the internet, with no question to the user.
- Removing the stop changes who decides, not the safety limits. The Constitution's
  Non-Negotiables (NN-01..07) are kept as the only limit, because they are constraints that
  Nara can satisfy without stopping.

**Refined 2026-10-17** by `spec-kit-open-decisions-2026-10-17.md` (decision 1): an irreversible or
externally visible action with no reversible form is held as blocked (`awaiting human approval`)
instead of performed on Nara's decision alone. Everything reversible keeps option A.

## Decision Outcome

**Chosen option (A)**: remove the hard-stop behavior from rule 8. Nara decides every item,
including secrets, destructive or irreversible actions, publishing, money/legal commitments,
and background budgets. Each decision must satisfy NN-01..07, and Nara records the NN check.
An item that cannot satisfy an NN article is recorded as blocked with its reason; the run
continues.

Routing rules added:

- Every request goes to Nara first (`.plaesy/roles/nara.md`).
- Agents never ask the user. Nara decides and records each choice as `ASSUMED — <default>`.
- If Nara lacks knowledge, it researches best practice on the internet (with source and retrieval
  date), asks the relevant roles through the Council Protocol, and then decides.

**Rejected alternatives** (recorded at decision time):

- **B**: keep hard stops for everything except background budget. Not chosen by the owner.
- **C**: keep hard stops only for irreversible or externally visible actions. Recommended by the
  session agent, not chosen by the owner.

## Consequences

- Nothing halts the run. The only limit on an action is NN-01..07 and the decision record.
- Actions that were hard stops (publish, delete, payments) can now run without a pause. This is
  the risk the owner accepted explicitly.
- `templates/constitution.template.md` and `.plaesy/memory/constitution.md` were not changed. Their
  section title still reads "Non-Negotiables (Hard Stops)".
- Gate-level blocks stay: `plaesy validate constitution` failures and the `/assess` mandatory
  audit are quality gates, not agent decisions.

## Constitutional Compliance & Quality Gates

- [x] `go test ./internal/assets/ -run '^TestAssetsDataMatchesSource$'` — passes after regeneration
- [x] `go test ./internal/validate/ -run 'Constitution'` — passes
- [ ] Full `go test ./...` not run (Test Scope rule: targeted tests only)

## References

- `instructions/plaesy.instructions.md` — Assistant Behavior rules 2 and 8, "Decision Routing (Nara First)" (accessed 2026-10-14)
- `instructions/role-mapping.instructions.md` — "Priority: Role Chain vs. Nara" (accessed 2026-10-14)
- `agents/nara.agents.md` — Project Rules, Boundaries, Example (accessed 2026-10-14)
- `prompts/start.md`, `prompts/loop.md`, `prompts/continue.md` — ask and halt lines removed (accessed 2026-10-14)

---

**Document Control**: Nara-first routing and no hard stops · v1.0 · Created 2026-10-14

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 1.0 | 2026-10-14 | session agent | Initial version |
