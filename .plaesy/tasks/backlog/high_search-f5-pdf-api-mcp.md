# Task: Search F5 - PDF Parser (Docling Sidecar) + HTTP API + MCP Server

## Objective
Add PDF parsing via Docling HTTP sidecar, REST API for search, and MCP server for agent integration.

## Scope
- PDF parser: call Docling sidecar (HTTP), extract Markdown + structure
- HTTP API: `/search`, `/index`, `/stats`, `/health` (JSON, auth optional)
- MCP server: stdio transport, tool `search_docs(query, top_k, filters, source)`
- Config: `docling.url`, `api.port`, `mcp.enabled`

## Design Rationale (ADR)
**Docling Sidecar Pattern:** All PDF parsing in separate process (Docker container) because:
- Docling requires Python + heavy ML deps (torch, transformers) - not suitable for Go CGO
- Sidecar isolates crashes, memory leaks, GPU usage from main process
- HTTP API enables remote Docling (GPU server) or local Docker
- Async request/response with timeout/retry handles variable PDF complexity

**HTTP API Framework:** `chi` router chosen for:
- Minimal deps, stdlib-compatible
- Built-in middleware (recovery, timeout, logging)
- OpenAPI spec generation via `swag` or manual spec

**MCP Server:** Stdio transport for:
- Native integration with Claude Code, other MCP clients
- No network port management
- Tool schema: `search_docs(query: string, top_k?: int, filters?: object, source?: string)`

**Auth Strategy:** Optional API key via header `X-API-Key`. Configurable:
- `api.auth.enabled: true/false`
- `api.auth.keys: ["key1", "key2"]` or external validator URL
- MCP server inherits same auth if exposed via HTTP bridge

**Rate Limiting:** Token bucket per client IP/key:
- `api.rate_limit.requests: 100`
- `api.rate_limit.window: "1m"`
- `api.rate_limit.burst: 20`

## Deliverables
1. `internal/search/parser/pdf_sidecar.go` - Docling HTTP client (async, timeout, retry)
2. `internal/search/api/rest.go` - HTTP server (chi), endpoints, OpenAPI spec
3. `internal/search/api/mcp.go` - MCP server (stdio), tool registration
4. `cmd/plaesy/search_serve.go` - `plaesy search serve --port=8080 --mcp`
5. Docling sidecar setup: `Dockerfile.docling`, `docker-compose.docling.yml` for local dev
6. Tests: PDF parsed correctly, API returns results, MCP tool call works
7. Docs: `docs/search-api.md`, `docs/mcp-integration.md`
8. OpenAPI spec: `docs/openapi/search-api.yaml`

## Acceptance Criteria
- `plaesy search index ./docs` processes PDFs via Docling (tested with sample PDFs)
- `curl localhost:8080/search?q=golang&top=5` returns JSON results with citations
- MCP client (Claude Code, etc.) can call `search_docs` tool (verified with MCP inspector)
- No PDF parsing code in Go (all via sidecar - enforced by build tag check)
- All previous phase tests pass: `go test ./internal/search/...`
- Health endpoint: `GET /health` returns `{status: "ok", index_ready: true, docling_ready: true}`
- Readiness endpoint: `GET /ready` for Kubernetes (checks index + docling connectivity)
- API versioning: `Accept: application/vnd.plaesy.v1+json`
- Rate limiting enforced with `429 Too Many Requests` + `Retry-After` header

## Test Plan
```bash
# Unit tests
go test ./internal/search/parser/... -v -run TestPDF
go test ./internal/search/api/... -v

# Integration tests (requires Docling sidecar)
docker compose -f docker-compose.docling.yml up -d
plaesy search index ./testdata/pdfs --incremental

# API tests
plaesy search serve --port=8080 &
SERVER_PID=$!
sleep 3
curl -s http://localhost:8080/health | jq .
curl -s "http://localhost:8080/search?q=golang&top=5" | jq .
curl -s "http://localhost:8080/stats" | jq .
kill $SERVER_PID

# MCP test (stdio)
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"search_docs","arguments":{"query":"golang","top_k":5}}}' | plaesy search serve --mcp | jq .

# Regression
go test ./internal/search/... -count=1
docker compose -f docker-compose.docling.yml down
```

## Dependencies
- F1-F4 complete
- Docling running locally (Docker) or remote URL
- Go HTTP framework: `github.com/go-chi/chi/v5` (minimal)
- MCP SDK: `github.com/mark3labs/mcp-go` (stdio transport)

## Config Schema Addition
```json
{
  "search": {
    "docling": {
      "url": "http://localhost:5001",
      "timeout": "30s",
      "max_retries": 3,
      "retry_backoff": "1s"
    },
    "api": {
      "enabled": false,
      "port": 8080,
      "auth": {
        "enabled": false,
        "keys": []
      },
      "rate_limit": {
        "requests": 100,
        "window": "1m",
        "burst": 20
      }
    },
    "mcp": {
      "enabled": false
    }
  }
}
```

## Config Migration
- New `DoclingConfig`, `APIConfig`, `MCPConfig` structs with defaults
- Feature flags default to `false` for backward compatibility
- `ConfigManager.MigrateFrom()` adds missing sections

## Operations
**Docling Sidecar:**
- Docker image: `ghcr.io/docling-project/docling:latest` (or pinned tag)
- Health check: `GET /health` every 10s
- Resource limits: 4GB RAM, 2 CPU (configurable)
- GPU support: `--gpus all` in docker-compose for faster processing

**API Server:**
- Graceful shutdown: SIGTERM drains connections (30s timeout)
- Metrics: Prometheus `/metrics` endpoint (requests, latency, errors)
- Logging: structured JSON, request ID correlation
- TLS: optional via `api.tls.cert_file`, `api.tls.key_file`

**MCP Server:**
- Stdio transport: no port, managed by client
- Tool schema validation on startup
- Error responses follow MCP spec (structured content)

## Observability
- Distributed tracing: OpenTelemetry (optional, via `OTEL_EXPORTER_OTLP_ENDPOINT`)
- Metrics: request count, latency p50/p95/p99, error rate, Docling queue depth
- Alerts: Docling unhealthy > 1m, API error rate > 5%, index not ready

## Security
- No PDF parsing in main process (sidecar isolation)
- Input validation: max PDF size 100MB, max pages 500
- Sandbox: Docling container runs non-root, read-only FS
- API auth optional but recommended for production

## Estimated Effort
Large (3-5 days)

## Definition of Done
- All deliverables implemented and tested
- Docling sidecar verified with 10+ sample PDFs (various layouts, languages)
- MCP integration tested with Claude Code / MCP inspector
- OpenAPI spec published at `docs/openapi/search-api.yaml`
- Documentation: `docs/search-api.md`, `docs/mcp-integration.md`, `docs/docling-setup.md`
- Docker compose for full stack: `docker-compose.yml` (plaesy + docling)