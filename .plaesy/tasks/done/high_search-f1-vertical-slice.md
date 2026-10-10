# Task: Search F1 - Vertical Slice (MD/HTML/JSON → Chunk → Embed → Store → CLI Search)

## Objective
Implement end-to-end vertical slice for local semantic search: parse Markdown/HTML/JSON → structure-aware chunking → embed with multilingual-e5-base (ONNX) → store in chromem-go HNSW + bleve BM25 → CLI `plaesy search index` / `plaesy search query`.

## Scope
- New package: `internal/search/` with interfaces and implementations
- Config: `.plaesy/config.json` (unified, single file)
- CLI: extend existing `search.go` with new subcommands

## Deliverables
1. `internal/search/config.go` - ConfigManager for unified `.plaesy/config.json`
2. `internal/search/parser/` - Parser interface + Markdown, HTML, JSON implementations
3. `internal/search/chunker/` - Structure-aware chunker (heading/paragraph, 200-500 tokens, overlap)
4. `internal/search/embedder/` - Embedder interface + ONNX runtime (intfloat/multilingual-e5-base)
5. `internal/search/vectorstore/` - VectorStore interface wrapping chromem-go HNSW
6. `internal/search/textindex/` - TextIndex interface wrapping bleve BM25
6. `internal/search/store/` - DocumentStore (raw + processed, hash + provenance)
7. `internal/search/indexer/` - Indexing pipeline (batch, worker pool, resume)
8. `cmd/plaesy/search_index.go` - `plaesy search index [path] --incremental`
9. `cmd/plaesy/search_query.go` - `plaesy search "query" --top=10 --json --source=local`
10. Unit + integration tests

## Acceptance Criteria
- `plaesy search index ./docs` indexes all .md/.html/.json files
- `plaesy search "golang tutorial" --top=10 --json` returns ranked results with citations
- Config persisted to `.plaesy/config.json` with search section
- No external API calls (fully local)
- Tests pass: `go test ./internal/search/...`

## Dependencies
- chromem-go (existing), bleve/v2 (new), onnxruntime-go (new CGO)
- Model download on first run (cached)

## Estimated Effort
Large (3-5 days)