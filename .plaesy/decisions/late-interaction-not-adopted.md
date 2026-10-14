---
name: late-interaction-not-adopted
description: Late-interaction (MaxSim over per-token E5 vectors) measured against the current single-vector search and not adopted
metadata:
  type: project
---

# Decision: late-interaction retrieval is not adopted

**Status:** Decided 2026-10-13. Single-vector E5 stays.

## Context

`rembed` exposes per-token hidden states (`EmbedTokens`). A ColBERT-style MaxSim
search over those vectors was tried as a way to improve ranking without a second model.

## Measurement

Spike in scratchpad (not in the repo), 19 questions, 28 documents, 241 chunks of at most
300 tokens, E5 with the same query/passage prefixes as production.

| Method | hit@1 | hit@3 | hit@5 | MRR | Median latency | Memory per chunk |
|---|---|---|---|---|---|---|
| Single-vector (production) | 0.526 | 0.947 | 0.947 | 0.714 | 48 ms | 3,072 B |
| Late-interaction (MaxSim) | 0.632 | 0.895 | 0.947 | 0.763 | 648 ms | about 715 KB |

## Reasons

- hit@1 rose by two questions out of 19, and hit@3 fell. The gain is not consistent.
- Latency is about 14 times higher, and memory about 230 times higher per chunk. Memory
  grows linearly with the number of chunks.
- Labels were written by the agent after reading the documents, so there is possible bias.
- E5 was not trained for late interaction, so this result may understate the method.

## References

- `rembed` README, supported models and per-token API: local module
  `github.com/rostamlabs/rembed@v0.3.0` (accessed 2026-10-13)
- Measurement program and labels: scratchpad `spike/` (not versioned)
