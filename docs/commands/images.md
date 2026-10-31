# `images`

```bash

plaesy images create --prompt <text> --out <path> [--provider <openai|gemini>] [--size <WxH>]
```

Generates an image asset via a configured provider API. With no subcommand,
prints the list below.

| Flag | Default | Description |
|---|---|---|
| `--prompt` | `""` | Image prompt text (required) |
| `--provider` | `""` | Image provider: `openai` (default) or `gemini` |
| `--size` | `""` | Image size, e.g. `1024x1024` |
| `--out` | `""` | Output file path (required) |

**Example**: `plaesy images create --prompt "a logo" --out logo.png --provider openai`

See [CLI Reference](../reference.md).
