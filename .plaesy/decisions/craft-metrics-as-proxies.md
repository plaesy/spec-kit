---
title: "Decision: Treated craft metrics as proxies"
description: "Ratchet heading presence, body length, and duplication — not whether an example is good"
updatedAt: "2026-09-30T22:10:00Z"
decided: "2026-09-30"
status: "Accepted"
---

# Decision: Treated craft metrics as proxies

**Status**: Accepted
**Decided**: 2026-09-30
**Deciders**: session agent
**Related**: decisions/accepted-mdlint-coverage-dip.md

## Context and Problem Statement

Prompt craft went from unmeasured to a ratchet: 10 prompts measured on objective
coverage, output-contract coverage, and worked-example coverage. The ratchet
raised the floor to 10/10 on each. But the metrics check heading presence, body
length, and duplication — they cannot judge whether an example is actually good.

## Decision Drivers

- Reporting 10/10 as quality would have been a false claim; the instrument
  measures structure, not craft.
- Recording the limitation beats reporting a perfect score, because the next
  author knows what the number does and does not mean.

## Decision Outcome

**Chosen option**: Treat the craft metrics as proxies, not as the goal.

**Rationale**: they check heading presence, body length and duplication — they
cannot judge whether an example is good. Recording that limitation in memory
beats reporting 10/10 as quality.

## Constitutional Compliance & Quality Gates

- [ ] Maintainability / testability / scalability / reliability acceptable

## References

- `scripts/internal/quality/prompt_craft_ratchet_test.go` — the ratchet.
- `scripts/internal/quality/prompt-craft-baseline.json` — recorded floor.

---

**Document Control**: Treated craft metrics as proxies · v1.0 · Created 2026-09-30

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 1.0 | 2026-09-30 | session agent | Initial version |