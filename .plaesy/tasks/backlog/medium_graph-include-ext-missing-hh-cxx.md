---
title: internal/graph has three out-of-sync extension lists; .hh and .cxx are never collected
phase: implement
status: backlog
createdAt: "2026-10-03T14:38:00.000Z"
updatedAt: "2026-10-03T14:38:00.000Z"
---

## Description

`scripts/internal/graph` keeps the set of "interesting" file extensions in
**three separate places**, and they no longer agree:

1. `graph.go` `includeExtRE` — which files are collected into the graph:
   `md|ps1|sh|js|jsx|ts|tsx|py|go|dart|java|kt|kts|swift|c|h|cc|cpp|hpp|cs|rs|rb|php`
2. `graph.go` `classify()` → `"source"` — which files are labelled source:
   `js|jsx|ts|tsx|py|go|dart|java|kt|kts|swift|c|h|cc|cpp|hpp|cs|rs|rb|php`
3. `symbols.go` `symbolsOf()` — which extensions get symbol extraction,
   including `.c .h .cpp .hpp .cc .hh .cxx` for C-family.

**Concrete bug:** `.hh` and `.cxx` are handled by `symbolsOf` but are absent
from `includeExtRE`, so those files are never collected. No graph node can
exist for them, so that branch of `symbolsOf` is unreachable dead code, and
`semantic.docsByExt` can never fire for them either.

Secondary: `md`, `sh`, and `ps1` are collected and *do* get symbol extraction
(`mdSymbols`, `shSymbols`, and the `.ps1` case) but `classify()` labels them
`"other"` rather than `"source"`. That may well be intentional — a Markdown or
shell file arguably isn't "source" — but it is currently an undocumented
divergence rather than a decision anyone recorded.

## Why it matters

Three hand-maintained lists of the same concept drift apart silently, and the
symptom is missing data rather than an error: a repo full of `.hh` files
produces a graph that silently omits them, with nothing reporting why. This is
the same failure class as the doc-comment regex drift guarded by
`scripts/internal/semantic/drift_test.go` — a lookup that quietly finds nothing.

## Notes

- **Fix belongs in `internal/graph`, not in `internal/semantic`.** The
  `semantic` side already handles `.hh`/`.cxx` correctly and waits on this.
- The blast radius is wider than it looks: widening `includeExtRE` changes what
  `plaesy analyze` collects for every project, so this should be verified with
  a graph rebuild, not just a unit test.
- Consider collapsing the three lists to one exported set
  (e.g. `graph.SourceExtensions`) and deriving `includeExtRE` and `classify()`
  from it, so a new language is added in one place.
- Found while wiring the doc-comment dispatch; not caused by that work.