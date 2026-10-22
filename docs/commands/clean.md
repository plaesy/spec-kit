# `clean`

```bash

plaesy clean [TARGET_DIR] [flags]
```

Remove Plaesy framework files and directories.

| Flag | Default | Description |
|---|---|---|
| `--yes`, `-y` | `false` | Auto-confirm deletion |
| `--dry-run` | `false` | Show what would be removed |
| `--level` | `safe` | Cleanup level: `safe`, `thorough`, `complete` |
| `--ai` | `""` | AI platform to clean (auto-detected if omitted) |
| `--verbose` | `false` | Show detailed progress |
| `--config` | `""` | Path to platform.json |
| `--backup` | `true` | Create backup before removal |
| `--no-backup` | `false` | Skip backup creation |

Levels decide what is actually removed:

| Level | Removes |
|---|---|
| `safe` | `.plaesy/` and the platform's mapped `core`/`instructions`/`prompts`/`agents` targets. `specs/` is kept. |
| `thorough` | The above plus `specs/` — the feature documents `create feature` writes. |
| `complete` | The above plus the `generic_ai` fallback's own files (`AI-INSTRUCTIONS.md`, `ai-config/`). |

`--backup` is on by default and covers every directory the chosen level removes.
A dry run lists what will go and what will be preserved.

**Example**: `plaesy clean --dry-run --level thorough`

See [CLI Reference](../reference.md).
