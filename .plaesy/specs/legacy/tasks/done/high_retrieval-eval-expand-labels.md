---
title: "Expand the labelled docs golden set to about 100 queries"
phase: implement
status: done
createdAt: "2026-10-13T00:00:00.000Z"
updatedAt: "2026-10-13T00:00:00.000Z"
---

# Task: Expand the labelled docs golden set to about 100 queries

## Description

The labelled set `scripts/testdata/golden_docs.jsonl` has 39 queries. That is too
few to decide the E5 query/passage prefix or the chunk cap. A single swapped rank
moves MRR by about 0.02–0.03, and the differences measured so far are of that size.

Context: [chunk-cap-model-tokens](../../decisions/chunk-cap-model-tokens.md),
[hybrid-matches-before-parent-child](../../decisions/hybrid-matches-before-parent-child.md),
[retrieval-eval-and-chunking](../../memory/retrieval-eval-and-chunking.md).

## Objective

Reach about 100 labelled queries, so that a prefix or cap difference of 0.05 MRR
can be told apart from noise.

## Scope

- Add about 60 queries to `scripts/testdata/golden_docs.jsonl`. Keep the existing
  format and the `keyword` check.
- Mix query styles: keyword, full question, paraphrase (no words shared with the
  heading), and Indonesian. Paraphrases matter most, because the vector leg
  depends on them.
- Include some queries whose answer is in the middle of a long section, not the heading.
- Each label is written by reading the document. `TestGoldenDocsLabels` must pass.

## Acceptance Criteria

- [x] At least 100 labels in `golden_docs.jsonl`, with at least 20 paraphrases and 10 Indonesian queries.
- [x] `go test ./internal/search/eval/ -run TestGoldenDocsLabels` passes.
- [x] The four-configuration matrix (prefix/no prefix × cap 480/256) is re-run on the
      expanded set, and the results are recorded in a decision file.

## Out of Scope

- Changing the prefix or cap defaults. That follows from the matrix result.
- Code search labels (`golden_code.jsonl`).
