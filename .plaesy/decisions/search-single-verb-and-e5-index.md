---
name: search-single-verb-and-e5-index
description: "One search verb, plaesy search \"<text>\". plaesy analyze builds every index it reads, and all indexes use the E5 model. Graph command removed; its query flags moved to analyze. Watch moved to analyze."
metadata:
  type: decision
  updatedAt: "2026-10-13T00:00:00.000Z"
---

# Decision: one search verb; analyze builds the indexes; E5 everywhere

**Status**: Accepted — partly implemented
**Decided**: 2026-10-13
**Deciders**: user (chose the direction) + session agent
**Related**: [remove-onnx-and-http-embedders](remove-onnx-and-http-embedders.md), [retrieval-prefix-and-cap-101-labels](retrieval-prefix-and-cap-101-labels.md)

## Context

`plaesy search` had two meanings: symbol search (from `plaesy analyze`) and document
search (its own `index` and `query`). The session first built a `search symbols` /
`search docs` tree on the agent's own proposal. The user rejected it: they want one
verb, `plaesy search`, and the data for it is indexed when `plaesy analyze` runs,
because the data depends on the project (code, second-brain notes, business documents).

## Decision

1. `plaesy search "<text>"` is the only search command. It reads the indexes that
   `plaesy analyze` wrote, and refuses an index that is missing or stale.
2. `plaesy analyze` builds every index search reads. Symbols are indexed now. The
   document index is to be wired in (task `high_doc-index-in-analyze`).
3. One embedding model for all indexes: E5 (`intfloat/multilingual-e5-base`, 768
   dimensions). The symbol index moved from MiniLM. Queries use `EmbedQuery`,
   stored passages use `EmbedPassages`.
4. `plaesy graph` is removed. Its query and export flags are `plaesy analyze` flags.
5. `plaesy analyze --watch` replaces `plaesy graph --watch`.
6. No deprecation aliases: the commands had not been released.

## Result format

- One ranked list. Each row carries a `kind` label: `symbol` or `doc`.
- The two indexes are merged with Reciprocal Rank Fusion, not raw scores. Cosine
  scores from the two indexes are on different scales, so a raw merge would favour
  one kind.
- `--kind symbol|doc` filters the list.
- Row format: `score  kind  ext  path:start-end  name`. Lines are 1-based and inclusive
  (see the line-range decision in the memory topic).

## Consequences

- `reference.md` no longer lists `graph`, `index`, or `query`. The search row was
  updated; the graph section was removed.
- Symbol index results are still file-level. Per-declaration units with line ranges
  are the next step: [high_symbol-units-with-line-ranges](../tasks/backlog/high_symbol-units-with-line-ranges.md).
- `--watch` still re-runs the whole analyze. Incremental per-file re-index is
  [high_search-f4-incremental-reindex](../tasks/backlog/high_search-f4-incremental-reindex.md).
