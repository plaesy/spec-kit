# `platforms`

```bash

plaesy platforms [subcommand] [flags]
```

The AI platform a project targets: `claude`, `cursor_ai`,
`github_copilot`, and the rest of `configs/platform.json`.

| Subcommand | Args | Description |
|---|---|---|
| *(none)* | — | Print this list |
| `detect` | — | Which platform this project is on |
| `list` | — | Every platform plaesy knows about |
| `show` | `[platform]` | One platform in detail |
| `get` | `<platform> <key>` | A platform's configuration value |

These four were `plaesy config detect`, `config list`, `config get-platform`
and `config show`. `config` had ended up grouping commands by the **file** they
read (`configs/platform.json`) rather than by the **object** a user is
thinking about — nobody asks "what does platform.json say", they ask "which
platform am I on" and "which platforms exist". Grouping by resource is what
kubectl, `gh` and terraform do, and it completes the noun family this CLI now
has: `features`, `tasks`, `images`, `platforms`, `context`, `config`, `stack`.

The six that stayed in `config` — `get-mapping`, `get-excludes`,
`get-clean-files`, `get-clean-dirs`, `get-structure` and `validate` — genuinely
are about the file. `platforms get-structure` or `platforms validate` would
name a resource the command does not touch.

`--config <path>` overrides `platform.json` for all four.

**Example**: `plaesy platforms detect`

See [CLI Reference](../reference.md).
