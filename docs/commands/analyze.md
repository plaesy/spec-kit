# `analyze`

```bash

plaesy analyze [project_path] [flags]
```

Generate AI-optimized project analysis under `.plaesy/analysis/`.

It also refreshes the document index (Markdown, HTML, and JSON under the project).

| Flag | Default | Description |
|---|---|---|
| `--no-graph` | `false` | Skip dependency graph build entirely |
| `--no-index` | `false` | Skip the document index build (`plaesy analyze` still builds it on demand) |
| `--force` | `false` | Force full regeneration (bypass fingerprint check) |
| `--if-changed` | `false` | Hidden no-op — fingerprint fast path is always on |
| `--watch` | `false` | Re-run analyze automatically when source files change |
| `--watch-interval` | `3` | Poll interval (seconds) for `--watch` |

**Example**: `plaesy analyze` — analyzes the current directory.

See [CLI Reference](../reference.md).
