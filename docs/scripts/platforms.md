# plaesy platforms

The AI platform a project targets: `claude`, `cursor_ai`,
`github_copilot`, and the rest of `scripts/configs/platform.json`.

Source: `scripts/cmd/plaesy/platforms.go` (verb) + `platforms_detect.go` /
`platforms_list.go` / `platforms_show.go` / `platforms_get.go`
(subcommands), `scripts/internal/config`.

| Subcommand | Args | Description |
|---|---|---|
| *(none)* | — | Print this list |
| `detect` | — | Which platform this project is on |
| `list` | — | Every platform plaesy knows about |
| `show` | `[platform]` | One platform in detail |
| `get` | `<platform> <key>` | A platform's configuration value |

These four were `plaesy config detect`, `config list`, `config get-platform`
and `config show`. `config` had ended up grouping commands by the **file** they
read (`scripts/configs/platform.json`) rather than by the **object** a user is
thinking about — nobody asks "what does platform.json say", they ask "which
platform am I on" and "which platforms exist". Grouping by resource is what
kubectl, `gh` and terraform do, and it completes the noun family this CLI now
has: `features`, `tasks`, `images`, `platforms`, `context`, `config`, `stack`.

The six that stayed in `config` — `get-mapping`, `get-excludes`,
`get-clean-files`, `get-clean-dirs`, `get-structure` and `validate` — genuinely
are about the file. `platforms get-structure` or `platforms validate` would
name a resource the command does not touch.

`--config <path>` overrides `platform.json` for all four subcommands.

**Example**: `plaesy platforms detect`

## `platforms detect`

```bash

plaesy platforms detect
```

Which platform this project is on, from `platforms.<name>.detection_patterns`.
No arguments. Prints the platform id (e.g. `claude`) and exits non-zero
if no platform is detected.

## `platforms list`

```bash

plaesy platforms list
```

Every platform plaesy knows about, one id per line.

## `platforms show`

```bash

plaesy platforms show [platform]
```

One platform in detail: name, provider, category. With no argument, shows the
detected platform.

## `platforms get`

```bash

plaesy platforms get <platform> <key>
```

A platform's configuration value, supporting dotted keys like `mapping.core`.
Shorthand names are resolved to the ids `platform.json` declares
(`claude_code` → `claude`, `copilot` → `github_copilot`, `cursor` →
`cursor_ai`, `windsurf` → `windsurf_ai`, `generic` → `generic_ai`).

## Troubleshooting

**"no platform detected"** — no detection pattern matched; check the project
has one of the platform marker files (e.g. `CLAUDE.md`, `.claude/`).

**"unknown platform"** — pass a declared id, display name, or shorthand.

**"missing key"** — verify the key exists in `platform.json` for that platform.

## Related Commands

- **`plaesy config`** — reads the same file for mapping/structure/cleanup questions (see [config-manager.md](./config-manager.md))
- **`plaesy init`** — scaffolds a project for a chosen platform (see [plaesy-init.md](./plaesy-init.md))
- **`plaesy clean`** — removes a platform's files (see [plaesy-clean.md](./plaesy-clean.md))
