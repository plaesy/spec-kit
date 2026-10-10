# Task: Search F3 - Reranker + Parent-Child Retrieval + Metadata Filter

## Objective
Add cross-encoder reranker for precision, parent-child retrieval for context, and metadata filtering.

## Scope
- Reranker: jina-reranker-v2-base-multilingual (ONNX), applied to top-N from hybrid
- Parent-child: retrieve child chunks, return parent document/section for context
- Metadata filter: source, date, path prefix, file type (applied pre-scoring)

## Design Rationale (ADR)
**Cross-Encoder Model:** `jina-reranker-v2-base-multilingual` chosen because:
- Multilingual support matches embedder (intfloat/multilingual-e5-base)
- ONNX export available, runs locally via onnxruntime-go (Linux/macOS) or HTTP sidecar (Windows)
- 278M params, ~1GB model, ~50ms latency on CPU for 10 candidates
- Outperforms bi-encoder on nDCG@10 by 8-12% in BEIR benchmarks

**Reranker Application:** Applied to top-50 from hybrid (configurable `reranker.top_n`), not all results. This bounds latency: 50 * 50ms = 2.5s worst case, typically <1s with batching.

**Parent-Child Retrieval:** Child chunks (200-500 tokens) retrieved for precision; parent sections returned for context. Implemented as post-processing wrapper: `ParentChildRetriever` wraps `HybridRetriever`, expands child hits to parent sections, deduplicates.

**Metadata Filter Pipeline:** Applied pre-scoring (before vector/BM25) to:
- Reduce candidate set early (performance)
- Prevent cross-tenant leakage (security)
- Support: `source`, `date_range`, `path_prefix`, `file_type`
- Filter syntax: `--filter=source:github,date:>2024-01-01,path:internal/search`

## Deliverables
1. `internal/search/reranker/interface.go` - Reranker interface
2. `internal/search/reranker/onnx_cross_encoder.go` - ONNX cross-encoder implementation
3. `internal/search/reranker/http_sidecar.go` - HTTP fallback (Windows, remote)
4. `internal/search/retriever/parent_child.go` - ParentChildRetriever wrapper
5. `internal/search/retriever/hybrid.go` - Add metadata filter pipeline (pre-scoring)
6. `internal/search/store/document.go` - Extend with `parent_id`, `section_path`, `metadata` fields
7. CLI flags: `--rerank`, `--parent-child`, `--filter=source:github,date:>2024-01-01`
8. Config: `reranker.enabled`, `reranker.top_n`, `reranker.model`, `parent_child.enabled`
5. Tests: reranker improves precision@5, parent-child returns correct parent, filters isolate

## Acceptance Criteria
- `--rerank` improves nDCG@10 on golden set vs hybrid alone (target: +5% relative)
- `--parent-child` returns parent section with child chunk citations (verified by test)
- Metadata filter applied before scoring (no cross-tenant leakage - integration test)
- Configurable via `.plaesy/config.json` and CLI flags
- All F1/F2 tests still pass: `go test ./internal/search/...`
- Windows: reranker works via HTTP sidecar (TEI/Ollama compatible)
- Memory budget: reranker model < 2GB RAM, < 1s p95 for top-50

## Test Plan
```bash
# Unit tests
go test ./internal/search/reranker/... -v
go test ./internal/search/retriever/... -v -run TestParentChild
go test ./internal/search/retriever/... -v -run TestMetadataFilter

# Integration tests
plaesy search index .plaesy/memory
plaesy search query "golang" --rerank --top=10 --json
plaesy search query "golang" --parent-child --top=10 --json
plaesy search query "golang" --filter=source:markdown --top=10 --json

# Regression
go test ./internal/search/... -count=1
plaesy search eval --golden=.plaesy/eval/golden.jsonl --modes=hybrid,hybrid_rerank
```

## Dependencies
- F2 complete (hybrid retriever working)
- ONNX cross-encoder model (downloaded on first use, cached at `~/.cache/onnx_models/jina-reranker-v2-base-multilingual.onnx`)

## Config Schema Addition
```json
{
  "search": {
    "reranker": {
      "enabled": false,
      "top_n": 50,
      "model": "jina-reranker-v2-base-multilingual",
      "model_path": "",
      "http_endpoint": ""
    },
    "parent_child": {
      "enabled": false
    }
  }
}
```

## Config Migration
- New `RerankerConfig` and `ParentChildConfig` structs with defaults
- `ConfigManager.MigrateFrom()` adds missing sections with defaults
- Feature flags default to `false` for backward compatibility

## Operations
- Reranker model download on first use (with progress log)
- HTTP sidecar health check: `GET /health` before rerank requests
- Latency budget: reranker adds < 500ms p95 for top-50
- Memory: model loaded once, shared across requests

## Estimated Effort
Medium (2-3 days)

## Definition of Done
- All deliverables implemented and tested
- Reranker improves nDCG@10 on golden set (documented in `docs/eval/reranker-<date>.md`)
- Windows HTTP fallback verified
- Documentation: `docs/search-reranker.md`, `docs/search-parent-child.md`