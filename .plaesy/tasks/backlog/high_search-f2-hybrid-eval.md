# Task: Search F2 - Hybrid Search (BM25 + Vector RRF) + Eval Harness

## Objective
Implement hybrid retrieval with Reciprocal Rank Fusion (RRF) and evaluation harness with golden set for measurable quality.

## Scope
- HybridRetriever combining VectorStore + TextIndex with RRF
- Eval harness: golden set (JSONL), recall@5/10, MRR, nDCG, latency p50/p95
- Baseline comparison: FTS-only vs Vector-only vs Hybrid vs Hybrid+Rerank

## Design Rationale (ADR)
**RRF Weight Defaults:** `vector_weight=0.6, bm25_weight=0.4` based on empirical finding that dense vectors capture semantic similarity better for technical queries, while BM25 excels at exact keyword matches. RRF k=60 follows standard literature (Cormack et al. 2009) and provides good balance between rank positions. These are tunable via config.

**Fusion Strategy:** RRF chosen over linear score combination because:
- Scale-invariant (BM25 scores ~0-10, cosine ~0-1)
- No calibration needed across different score distributions
- Proven effective in TREC benchmarks

**Golden Set Construction:** 50+ queries sourced from:
- 20 from existing codebase FAQ/comments (real developer questions)
- 15 from GitHub issues/PRs (feature requests, bug reports)
- 15 manually curated covering edge cases (ambiguous terms, multi-hop)

## Deliverables
1. `internal/search/retriever/hybrid.go` - HybridRetriever with RRF (configurable k, weights)
2. `internal/search/eval/golden.go` - Golden set loader (query → expected chunk/doc IDs)
3. `internal/search/eval/metrics.go` - recall@k, MRR, nDCG, latency percentiles
4. `internal/search/eval/harness.go` - Runner: compare modes, output markdown table
5. `cmd/plaesy/search_eval.go` - `plaesy search eval --golden=golden.jsonl --modes=fts,vector,hybrid,hybrid_rerank`
6. Golden set bootstrap: 50+ queries from existing codebase + manual curation (stored at `.plaesy/eval/golden.jsonl`)
7. Baseline report in `docs/eval/baseline-<date>.md`
8. CI integration: `.github/workflows/search-eval.yml` runs eval on PR, fails if hybrid < best single-mode

## Acceptance Criteria
- Hybrid RRF outperforms single-mode baselines on golden set (statistically significant p<0.05)
- Eval command runs in < 60s, outputs comparison table with 95% CI
- Hybrid config: vector_weight=0.6, bm25_weight=0.4, rrf_k=60 (tunable via `.plaesy/config.json` and CLI flags)
- No regression in F1 functionality: `go test ./internal/search/...` passes
- Golden set schema documented: JSONL with fields {query, expected_ids[], relevance_scores?, metadata?}
- Eval harness supports `--output=json|markdown|csv` for CI consumption

## Test Plan
```bash
# Unit tests
go test ./internal/search/retriever/... -v -run TestHybrid
go test ./internal/search/eval/... -v

# Integration test (requires indexed corpus)
plaesy search index .plaesy/memory
plaesy search eval --golden=.plaesy/eval/golden.jsonl --modes=fts,vector,hybrid

# Regression test
go test ./internal/search/... -count=1
```

## Dependencies
- F1 complete (vectorstore + textindex + embedder working)
- bleve BM25 index built alongside vector index

## Config Schema Addition
```json
{
  "search": {
    "retriever": {
      "hybrid": {
        "rrf_k": 60,
        "vector_weight": 0.6,
        "bm25_weight": 0.4
      }
    }
  }
}
```

## Config Migration
- New fields added to `config.RetrieverConfig.Hybrid` with defaults above
- `ConfigManager.MigrateFrom()` handles legacy configs without these fields
- Backward compatible: missing fields → defaults applied on load

## Operations
- Eval harness emits structured logs for observability
- Golden set versioned with corpus (hash in filename: `golden-<corpus-hash>.jsonl`)
- CI gate: PR fails if hybrid recall@10 < max(fts, vector) recall@10 - 5%

## Estimated Effort
Medium-Large (2-3 days)

## Definition of Done
- All deliverables implemented and tested
- Baseline report committed to `docs/eval/`
- CI workflow passes on main branch
- Documentation updated: `docs/search-hybrid.md`