# `config`

```bash

plaesy config [subcommand] [flags]
```

Centralized platform configuration management using
`configs/platform.json`.

Global flag: `--config` (path to platform.json; default: auto-resolved).

## Subcommands

| Subcommand | Args | Description |
|---|---|---|
| `get-mapping` | `<section> <mapping_type>` | Get mapping value with value/excludes structure |
| `get-excludes` | `<section> <mapping_type>` | Get exclude patterns for a mapping |
| `get-clean-files` | `[platform]` | Get files to clean for platform |
| `get-clean-dirs` | `[platform]` | Get directories to clean for platform |
| `get-structure` | `<component>` | Get Plaesy structure config |
| `validate` | — | Validate platform configuration |

These six read `configs/platform.json` and ask things about **the
file**. The four platform questions — which platform am I on, which exist, what
does one look like, what is its config value — used to sit here too, which put
two different objects under one noun: the config file, and the AI platform. They
are now `plaesy platforms`; see that section. With no subcommand, `plaesy config`
prints this list.

**Example**: `plaesy config validate`

See [CLI Reference](../reference.md).
