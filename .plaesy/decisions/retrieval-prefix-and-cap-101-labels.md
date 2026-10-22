---
name: retrieval-prefix-and-cap-101-labels
description: "Matrix on 101 labelled docs queries: E5 query/passage prefixes help vector and hybrid at both caps; cap 256 beats 480 for hybrid and vector with prefixes. Reverses the 39-query result. Default changed to 256 (approved)."
metadata:
  type: decision
  updatedAt: "2026-10-13T00:00:00.000Z"
---

# Decision: E5 prefixes on; cap choice pending on 101 labels

**Status**: Accepted. Prefixes kept; cap default changed to 256 with the user's approval.
**Decided**: 2026-10-13
**Deciders**: session agent, measured on `scripts/testdata/golden_docs.jsonl` (101 labels)
**Related**: [chunk-cap-model-tokens](chunk-cap-model-tokens.md), [hybrid-matches-before-parent-child](hybrid-matches-before-parent-child.md), task `high_retrieval-eval-expand-labels`

## Context

The first matrix used 39 labels and suggested no-prefix was equal or better, and that
cap 480 was best. The task then expanded the set to 101 labels (61 paraphrases, 10
Indonesian). The indexes were the same ones built after the hybrid fix; only the
queries changed.

## Results (MRR, 101 labels)

| config | vector | bm25 | hybrid |
|---|---|---|---|
| prefix, cap 480 | 0.589 | 0.633 | 0.621 |
| no prefix, cap 480 | 0.497 | 0.631 | 0.581 |
| prefix, cap 256 | **0.629** | 0.627 | **0.663** |
| no prefix, cap 256 | 0.578 | 0.626 | 0.616 |

## Decision

1. **Keep the E5 prefixes** (`query: ` / `passage: `). The model card prescribes them
   for asymmetric retrieval. They help vector search at both caps, by +0.09 at 480
   and +0.05 at 256. The difference is larger than the noise seen on 39 labels.
2. **Cap**: 256 is better for hybrid (+0.04) and vector (+0.04) with prefixes. BM25 is
   unchanged by the cap on this set. The default is now 256 (approved by the user).


## Caveats

- The labels were written by the same agent that built the search, from the same
  documents. Paraphrase labels reduce lexical bias but do not remove it. A second labeller
  would strengthen the conclusion.
- Corpus is `docs/` only (32 files). Source code retrieval is not measured here
  (`golden_code.jsonl` has no matching index).
- Differences of 0.01–0.02 remain within noise even at 101 labels.
