---
name: hybrid-matches-before-parent-child
description: "Hybrid retrieval truncates to top-K matches before parent-child expansion; siblings are appended after the matches and never listed twice. Previously siblings filled the top-K list and hybrid returned about 3-5 documents per query."
metadata:
  type: decision
  updatedAt: "2026-10-13T00:00:00.000Z"
---

# Decision: Matches first, parent-child siblings after

**Status**: Accepted — implemented
**Decided**: 2026-10-13
**Deciders**: session agent, after measuring the retrieval matrix (user asked to fix step 1)
**Related**: [chunk-cap-model-tokens](chunk-cap-model-tokens.md), [retrieval-eval-and-chunking](../memory/retrieval-eval-and-chunking.md)

## Context

Parent-child retrieval is on by default. For every matched chunk, all sibling
chunks of its document were added, then the whole list was cut to `TopK` chunks.
With long documents, the siblings of the top one or two documents filled the list.
Measured on the labelled docs set: hybrid returned 5.1 documents per query at cap
480 and 3.1 at cap 256, against about 16–18 for vector and BM25.

## Decision

1. `Retrieve` and `RetrieveParallel` cut the fused matches to `TopK` first.
2. `expandParentChild` returns the matches in rank order, then the siblings. A
   sibling scores half of the match it belongs to.
3. No chunk is listed twice. The old code could list a chunk as both a match and a
   sibling.

## Verification

- `TestExpandParentChild_MatchesStayFirstInRankOrder`: matches stay first in rank
  order, and no ID repeats. The old code puts a sibling in position 2, so this test
  fails on it.
- Re-run of the labelled matrix: see the context file for the result.

## Consequences

- `Retrieve` can return more than `TopK` chunks, because siblings are appended.
  Callers that need only the matches must read the first `TopK` entries.
