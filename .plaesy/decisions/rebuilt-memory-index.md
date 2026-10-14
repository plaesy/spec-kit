---
title: "Decision: Rebuilt memory.md as the index it is required to be"
description: "Moved ~145 lines of inline rules into four topic files, leaving memory.md as index only"
updatedAt: "2026-09-30T22:10:00Z"
decided: "2026-09-30"
status: "Accepted"
---

# Decision: Rebuilt memory.md as the index it is required to be

**Status**: Accepted
**Decided**: 2026-09-30
**Deciders**: session agent
**Related**: decisions/prompts-source-mirror-regenerated.md

## Context and Problem Statement

`.plaesy/memory.md` is specified as an index-only file: one line per topic file,
flat list, no inline rules. It held ~145 lines of rules inline, which is why
`.plaesy/memory/` was empty — the rules lived in the index and the topic files
the index was supposed to point at did not exist.

## Decision Drivers

- An index that holds content is not an index; it is a second copy of the
  knowledge, and the second copy is the one that drifts.
- The size enforcement rule ("memory.md: Keep as index only") was being
  violated by the file's own contents.

## Decision Outcome

**Chosen option**: Rebuild memory.md as the index it is required to be.

**Rationale**: It held ~145 lines of rules inline, which is why
`.plaesy/memory/` was empty. Rules moved to four topic files; nothing lost.

## Constitutional Compliance & Quality Gates

- [ ] Maintainability / testability / scalability / reliability acceptable

## References

- `.plaesy/memory/prompt-authoring.md`
- `.plaesy/memory/quality-ratchets.md`
- `.plaesy/memory/testing-discipline.md`
- `.plaesy/memory/git-and-state.md`

---

**Document Control**: Rebuilt memory.md as the index it is required to be · v1.0 · Created 2026-09-30

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 1.0 | 2026-09-30 | session agent | Initial version |