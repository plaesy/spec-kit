# `init`

```bash

plaesy init [directory] [--ai <platform>] [--plaesy-home <path>]
```

Scaffold a new Plaesy project: `.plaesy/` structure + AI platform files.

| Flag | Default | Description |
|---|---|---|
| `--ai` | `""` | AI platform (e.g. `claude`, `cursor_ai`) |
| `--plaesy-home` | `""` | Override Plaesy repo root (`PLAESY_HOME` env var) |

Supported platforms: `claude`, `cursor_ai`, `github_copilot`,
`windsurf_ai`, `cline`, `deepseek`, `kilo`, `qoder`, `trae_ai`,
`continue_dev`, `tabnine`, `codeium`, `codewhisperer`, `studio_bot`,
`replit_ghostwriter`, `llama_index`, `ollama`, `lm_studio`, `generic_ai`.

**Examples**:

```bash

plaesy init my-project --ai claude
plaesy init . --ai cursor_ai
```

See [CLI Reference](../reference.md).
