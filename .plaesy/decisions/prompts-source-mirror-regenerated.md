---
title: "Decision: Kept prompts/ as source and regenerated the mirror"
description: "The runtime loads .kilo/commands/, so an un-regenerated edit changes nothing an agent can see"
updatedAt: "2026-09-30T22:10:00Z"
decided: "2026-09-30"
status: "Accepted"
---

# Decision: Kept prompts/ as source and regenerated the mirror

**Status**: Accepted
**Decided**: 2026-09-30
**Deciders**: session agent
**Related**: none

## Context and Problem Statement

The corpus ships two trees: `prompts/` as the authoring source, and
`.kilo/commands/` as the copy the agent runtime actually loads. Nothing
connected them, so every prompt edited over two sessions was mirrored stale —
and the mirror still carried `design-spine.md`, a file that has never existed,
months after the source was corrected.

## Decision Drivers

- The runtime loads the mirror, so an edit verified against the source tree
  alone is unverified where it counts.
- A mirror that drifts from its source is worse than no mirror, because it
  reports success while the change is not in effect.

## Decision Outcome

**Chosen option**: Keep `prompts/` as source, regenerate the mirror.

**Rationale**: Verified byte-identical: the runtime loads `.kilo/commands/`, so
an un-regenerated edit changes nothing an agent can see.

## Constitutional Compliance & Quality Gates

- [ ] Maintainability / testability / scalability / reliability acceptable

## References

- `scripts/cmd/plaesy/mirror_parity_test.go` — `TestPromptMirrorMatchesSource`
  guards the two trees as one contract.
- `scripts/cmd/plaesy/reload.go` — `plaesy reload --ai <platform>` regenerates
  the mirror.

---

**Document Control**: Kept prompts/ as source and regenerated the mirror · v1.0 · Created 2026-09-30

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 1.0 | 2026-09-30 | session agent | Initial version |