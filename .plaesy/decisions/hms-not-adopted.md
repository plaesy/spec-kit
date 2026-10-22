---
title: "HMS not adopted"
description: "Holographic Memory System (writerslogic/holographic-memory) rejected for plaesy memory"
updatedAt: "2026-10-18T00:00:00.000Z"
---

# Decision: HMS not adopted

**Date:** 2026-10-18
**Status:** decided by the project owner ("abaikan hms", 2026-10-18)

## Decision

Do not adopt the Holographic Memory System (HMS) as a memory or retrieval backend for plaesy.

## Reasons

- **Scale does not justify it.** `.plaesy/memory/` holds 17 files and `decisions.md` is about 100 lines.
  BM25 or plain search covers this size.
- **Capacity limits (repo's own claims).** False positives reach 4% at 500 items and 32% at 1,000 items in
  a single bundle (D=16,384). Wrong recalled memories are a risk for agent context.
- **Weak sparse recall (repo's own claim).** 0.25 against FAISS HNSW 0.83 on glove-100.
- **Does not address the framework's gaps.** HMS does not provide lifecycle, supersession, provenance,
  or the raw vs curated split.
- **Operational cost.** AGPL-3.0 license (or commercial). Writable stores take an exclusive process lock,
  which conflicts with parallel agents. The Python SDK needs a separate server not in the repo. Native Rust
  build adds a dependency to a Go binary distribution.
- **Security.** Lossy vectors are not encryption. Ingest stores source passages by default.

## Confidence

Based on README and WebFetch summary only. No benchmark was run locally.
Re-open only if the memory store grows to thousands of items and BM25 recall is measured as insufficient.

## References

- [writerslogic/holographic-memory](https://github.com/writerslogic/holographic-memory) · accessed 2026-10-18
- Related research: [research-memory-systems-2026-10-18](../memory/research-memory-systems-2026-10-18.md)
