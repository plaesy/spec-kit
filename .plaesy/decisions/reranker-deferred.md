---
name: reranker-deferred
description: Cross-encoder reranker (F3) deferred. Candidate bge-reranker-v2-m3 is Apache-2.0 but needs a rembed change and does not justify the cost for duplicate checks
metadata:
  type: project
---

# Decision: reranker deferred

**Status:** Decided 2026-10-13. No reranker is implemented. F3 keeps its parent-child
and filter work, which is wired.

## Context

Task F3 asked for a cross-encoder reranker (`jina-reranker-v2-base-multilingual`, ONNX)
on top of the hybrid search. The repo had already removed ONNX (see
[remove-onnx-and-http-embedders](remove-onnx-and-http-embedders.md)).

## Findings

- **jina-reranker-v2-base-multilingual** is CC-BY-NC-4.0 (non-commercial). Source:
  https://www.promptlayer.com/models/jina-reranker-v2-base-multilingual (accessed 2026-10-13).
  This is a secondary source; the license was not re-checked on Jina's own page.
- **BAAI/bge-reranker-v2-m3** is listed as apache-2.0 on its Hugging Face page, which was
  fetched on 2026-10-13: https://huggingface.co/BAAI/bge-reranker-v2-m3 . Its architecture is
  XLM-RoBERTa with a sequence-classification head (`classifier.dense`, `classifier.out_proj`).
- A spike loaded the checkpoint through `rembed` only after a local patch to `rembed`
  (a token-id method for `</s>` pairs) and added pooling config in a scratch copy. That
  patch is not in the upstream module, so adopting it would mean maintaining a fork.
- The checkpoint is about 2.3 GB (F32, 0.6B parameters).
- The spike was stopped before any score was computed. No quality numbers exist.

## Reason for deferring

The reranker would only improve the top rank (hit@1 was 0.526 against hit@3 of 0.947).
Duplicate checks read the top results anyway, so the gain does not justify a 2.3 GB model,
a patched runtime, and extra latency on every search.

## Reopen when

Search is used for document questions where the top answer matters, or a permissively
licensed reranker runs in `rembed` without a patch.

## References

- See [late-interaction-not-adopted](late-interaction-not-adopted.md) for the other
  option measured in the same session.
