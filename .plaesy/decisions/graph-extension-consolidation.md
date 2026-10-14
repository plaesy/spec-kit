---
title: "Graph extension lists consolidated; .hh and .cxx collected"
status: accepted
updatedAt: "2026-10-13T00:00:00.000Z"
---

# Graph extension lists consolidated

## Context

`scripts/internal/graph` kept its extension lists in three places that drifted.
`.hh` and `.cxx` had symbol extraction but were missing from the collection regex,
so no graph node could carry them.

## Decision

- `scripts/internal/graph/extensions.go` holds `SourceExtensions`, the one list
  of source-code extensions.
- `includeExtRE` is derived from `SourceExtensions` plus the non-source
  `.md .sh .ps1`.
- `nodeTypeOf` returns `"source"` through `isSourceExt`, which reads the same list.
- `.hh` and `.cxx` are added to the list, so they are collected and labelled source.

Classification stays `"source"` or `"other"`, with directory labels taking
precedence as before. The `"doc"`, `"script"` labels proposed in the backlog task
are not adopted: they would change the labels of existing nodes, and nothing reads
them yet. Revisit if a consumer needs them.

## Rejected

- Keeping three lists and only adding `.hh`/`.cxx` to the regex. That repeats the
  drift this decision fixes.

## Consequences

- Projects with C++ headers or `.cxx` sources get those files in the graph on the
  next `plaesy analyze --force`. Existing fingerprints do not change, so a plain
  `plaesy analyze` may skip the rebuild until sources change.

## References

- Task: `.plaesy/tasks/done/medium_graph-include-ext-missing-hh-cxx.md`
- Accessed: 2026-10-13
