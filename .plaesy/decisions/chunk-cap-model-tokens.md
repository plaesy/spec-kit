---
name: chunk-cap-model-tokens
description: "Chunks are capped at 256 tokens counted by the embedding model's tokenizer; oversized paragraphs are split by sentence, then word, then rune. MaxChunkTokens was never wired into the pipeline, so the cap was silently 0."
metadata:
  type: decision
  updatedAt: "2026-10-13T00:00:00.000Z"
---

# Decision: Chunk cap in model tokens (256), oversized text split

**Status**: Accepted — implemented
**Decided**: 2026-10-13
**Deciders**: user (asked for steps 1–2 of the retrieval plan) + session agent
**Related**: [remove-onnx-and-http-embedders](remove-onnx-and-http-embedders.md), [retrieval-eval-and-chunking](../memory/retrieval-eval-and-chunking.md)

## Context

- `rembed` (the only embedder) truncates input at the model's 512-token limit
  without a warning. A chunk longer than that was embedded from its first 512
  tokens only.
- `pipeline.go` built `chunker.Config` without `MaxChunkTokens`, so the cap was 0.
  Every paragraph became its own chunk, and single paragraphs reached 8,529 tokens
  (estimated with words × 1.3).

## Decision

1. Chunk size is counted with the embedding model's tokenizer
   (`RembedEmbedder.CountTokens`, via `rembed`'s `Tokenize`). The chunker requires
   a counter: `NewStructureAwareChunker` panics without one, and `NewIndexer`
   fails if the embedder cannot count tokens.
2. Default cap `MaxChunkTokens = 256`. Reason: the model accepts 512 tokens, and
   the cap must leave room for `[CLS]`, `[SEP]`, and the `passage: ` prefix.
3. Text over the cap is split in order: whole sentences first, then words, then
   runes for a single token longer than the cap. No chunk leaves the chunker over
   the cap, including after overlap and after merging small chunks.
4. `TargetTokens` is removed. Nothing read it.
5. `plaesy search index --max-chunk-tokens N` overrides the cap for experiments.

## Verification

- Real model, `PLAESY_TEST_REMBED=1`: a short phrase counts in single digits; text
  far past the limit counts exactly 512, which marks truncation.
- Index of `docs/` (32 files): with cap 480, max 480 and median 438; with cap 256,
  max 256 and median 235. Before the fix, max was 8,529 with median 286.
- Unit tests for the cap: oversized paragraph split, sentence boundaries, no text
  lost, rune split terminates, overlap cannot exceed the cap.

## Consequences

- Indexes built before this change are invalid. Rebuild with `--force`.
- The cap was first set to 480 from a 39-label matrix. The 101-label matrix (retrieval-prefix-and-cap-101-labels) favours 256, so the default is 256 since 2026-10-13.
