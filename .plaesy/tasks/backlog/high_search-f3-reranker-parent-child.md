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
## Status (background fork)
- Blocked, not started. The reranker is an ONNX cross-encoder (jina-reranker-v2, ~1GB model download, cached under ~/.cache). The repo decision remove-onnx-and-http-embedders.md drops onnxruntime and cgo, and this task proposes an ONNX reranker and an HTTP sidecar. Resolving that needs a decision first: which runtime, and whether a sidecar service is acceptable.

## Status: F3 part 1 (parent-child, metadata filter, CLI flags)

Implemented but unverified. No test has been run. Compile was not confirmed at the end
of this pass: `internal/search/chunker` was mid-edit by another agent (undefined
`normalizeNewlines`, `splitUnits` in `structure_aware.go`), and `retriever` imports it.
`gofmt -e` (syntax) passes on all new and changed files.

**Built**
- `retriever/filter.go`: `ParseFilter` for `source`, `path` (whole-segment, matched
  anywhere in the absolute stored path), `file_type`, and `date` (`>`, `>=`, `<`, `<=`, `=`).
  A chunk without the `modified` date fails every date term (fail closed).
- `retriever/hybrid.go`: `QueryOptions{Filters, Where}` and `RetrieveWith`. `Where`
  predicates run on the vector and text candidates before RRF fusion. `Retrieve`
  keeps its signature, so the eval adapter and docindex are unchanged.
- `retriever/parentchild.go`: `ParentChildRetriever`. A parent section is every chunk
  of a document with the same `SectionPath`. It returns the section text and the
  matched children as citations with offsets. This is separate from the sibling
  expansion in `expandParentChild`, which is kept as it was.

**Acceptance criteria (status)**
- Reranker `--rerank` and deliverables 1-3: NOT implemented (licence and runtime
  decision pending: jina-reranker-v2 is CC-BY-NC-4.0 and deprecated; rembed runs XLM-R
  encoders but the reranker head is not built in).
- `--parent-child` returns parent section with child citations: implemented, untested.
- Metadata filter applied before scoring, no cross-tenant leakage: predicates run before
  fusion, implemented, untested. Exact-match filters are still applied by the stores.
- Configurable via config and CLI flags: NOT done. The CLI flags `--parent-child` and
  `--filter` are not wired. `docindex.Query` calls `retr.Retrieve(ctx, text, nil)` and
  does not take options. Wiring needs a `QueryWith(ctx, ..., QueryOptions)` addition in
  `docindex`, which was out of scope for this pass.
- Existing F1/F2 tests still pass: not run.
- Windows HTTP sidecar reranker: NOT done (blocked on the decision above).
- Memory and latency budget: not measured.

**Known gaps**
- Date filter needs chunk metadata key `modified` (YYYY-MM-DD). `indexer/pipeline.go`
  does not write it yet. Until it does, every date term matches nothing.
- Path filter uses whole-segment contains on the absolute path, so `path:internal/search`
  works, but a relative-path prefix would need the document root.
- Predicates run after the store's top-K, so a selective filter can return fewer than K.
  Raising the store's K for filtered queries is not done.

**Tests written, not run**: `retriever/filter_test.go`, `retriever/parentchild_test.go`.

## Status (wiring pass)

Wired, compiled (`go build ./...`, `go vet` on touched packages), **not tested** (test run deferred by request):

- `retriever/filter.go`: `ParseFilter` for `source`, `path`, `file_type`, `date`. Applied before RRF fusion through `QueryOptions.Where`.
- `retriever/parentchild.go`: `ParentChildRetriever` returns whole sections with `Path`, `StartLine`/`EndLine` (span of the section's chunks) and matched children.
- `docindex.QueryWith(..., opts, parentChild)`: new entry point. `Query` keeps its signature and calls it with zero options, so eval and existing callers are unchanged.
- `indexer/pipeline.go`: writes `modified` (YYYY-MM-DD from `doc.ModTime`, omitted when zero). `source` was already written.
- `cmd/plaesy/search.go`: `--filter` (doc results only; a bad spec fails before any index is read) and `--parent-child` (doc results only).

Not done:
- Reranker and `--rerank`: blocked on the licence and runtime decision (see the reranker section above).
- Acceptance criteria 1 (nDCG gain) and the integration test for cross-tenant isolation: need a test run and a golden set.
