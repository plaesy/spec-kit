# `reach`

```bash

plaesy reach [platform] [query] [flags]
plaesy reach configure <platform>
plaesy reach list
```

Access external data sources (web, GitHub, YouTube, etc.) across multiple
platforms, with response caching and rate limiting.

| Flag | Default | Description |
|---|---|---|
| `--path` | `.` | Project directory |
| `--outdir` | `.plaesy/analysis` | Output directory |
| `--action` | `""` | Action for platform (`search`, `view`, etc.) |
| `--limit` | `10` | Max results for search actions |
| `--index` | `false` | Rebuild semantic index (internal use) |
| `--timeout` | `""` | Query timeout (e.g. `30s`, `1m`) |
| `--cache-ttl` | `""` | Cache TTL (e.g. `5m`, `1h`) |
| `--no-cache` | `false` | Disable caching for this query |
| `--verbose` | `false` | Verbose output |
| `--json` | `false` | Output as JSON |

## `reach configure`

Store credentials for a platform (e.g. an API token) in the reach config.

## `reach list`

List every platform `reach` can target.

**Examples**:

- `plaesy reach github "spec-kit issues"` — query a platform directly
- `plaesy reach configure github` — store credentials for a platform
- `plaesy reach list` — list available platforms

See [CLI Reference](../reference.md).
