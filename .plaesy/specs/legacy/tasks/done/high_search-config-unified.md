# Task: Search Config - Unified Config Migration to .plaesy/config.json

## Objective
Consolidate all capability configs (reach, search, analyze, graph) into single `.plaesy/config.json` with ConfigManager per domain.

## Scope
- Design unified JSON schema with namespaced sections
- Migrate `reach` config from `.plaesy/reach/config.json`
- Create `search` config section with all F1-F5 parameters
- Maintain backward compatibility (read old paths, write new)
- ConfigManager pattern shared via `internal/config/manager.go`

## Deliverables
1. `internal/config/manager.go` - Generic ConfigManager[T] with file watch/hot-reload
2. `internal/config/schema.go` - Unified schema (reach, search, analyze, graph sections)
3. `internal/reach/config.go` - Migrate to use unified ConfigManager
4. `internal/search/config.go` - New, uses unified ConfigManager
5. Migration logic: on Load(), read old paths → write unified → delete old
6. CLI: `plaesy config view`, `plaesy config edit`, `plaesy config migrate`
7. Tests: migration works, hot-reload triggers callback, backward compat

## Unified Schema (.plaesy/config.json)
```json
{
  "version": 1,
  "reach": { ...existing reach config... },
  "search": {
    "embedder": { "type": "onnx", "model": "intfloat/multilingual-e5-base", ... },
    "vector_store": { "type": "chromem", "hnsw": { ... } },
    "text_index": { "type": "bleve" },
    "chunker": { "target_tokens": 350, ... },
    "retriever": { "hybrid": { ... }, "parent_child": true },
    "reranker": { "enabled": false, ... },
    "indexer": { "workers": 4, "resume": true }
  },
  "analyze": { "auto_index": true, ... },
  "graph": { "include_extensions": [...], ... }
}
```

## Acceptance Criteria
- Single `.plaesy/config.json` holds all capability configs
- `plaesy reach configure` still works (reads/writes reach section)
- `plaesy search configure` works (reads/writes search section)
- Old config files migrated on first load, then deleted
- Hot-reload: file change → callback → components refresh
- No breaking changes to existing CLI commands

## Dependencies
- None (can start in parallel with F1)

## Estimated Effort
Medium (2 days)