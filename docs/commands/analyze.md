# `analyze`

```bash

plaesy analyze [project_path] [flags]
```

Generate AI-optimized project analysis under `.plaesy/analysis/`.

It also refreshes the search indexes `plaesy search` queries (the semantic symbol index and the document index), so a
search after an analyze run is not refused as stale.

| Flag | Default | Description |
|---|---|---|
| `--no-graph` | `false` | Skip dependency graph build entirely |
| `--no-index` | `false` | Skip the semantic-search and document index builds (`plaesy analyze` still builds it on demand) |
| `--force` | `false` | Force full regeneration (bypass fingerprint check) |
| `--if-changed` | `false` | Hidden no-op — fingerprint fast path is always on |
| `--watch` | `false` | Re-run analyze automatically when source files change |
| `--watch-interval` | `3` | Poll interval (seconds) for `--watch` |

**Example**: `plaesy analyze` — analyzes the current directory.

See [CLI Reference](../reference.md).
