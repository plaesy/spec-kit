# `doctor`

```bash

plaesy doctor [platform] [flags]
```

Run health checks across all plaesy capabilities: reach platforms (backend
availability, credentials, rate limits), the search index (freshness,
document count, vector/text index status), and configuration (validity,
migration status).

| Flag | Default | Description |
|---|---|---|
| `--platform` | `""` | Specific platform to check (`web`, `github`, `youtube`, etc.) |
| `--verbose` | `false` | Show rate limit and cache stats |

**Examples**:

- `plaesy doctor` — full system health check
- `plaesy doctor --platform web` — check one reach platform
- `plaesy doctor --verbose` — include rate limits and cache stats

See [CLI Reference](../reference.md).
