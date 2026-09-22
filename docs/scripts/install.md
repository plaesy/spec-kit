# plaesy install / status / uninstall / repair / upgrade

**Binary installation and self-management commands.**

**Priority:** HIGH - Required for framework installation

Source: `scripts/cmd/plaesy/install.go` + `scripts/cmd/plaesy/status.go` +
`scripts/internal/installer/installer.go`.

## Purpose

`plaesy` is a single self-contained Go binary. There is no installer script
to download and no `git clone` of the whole repository into a
`PLAESY_HOME` directory anymore — the running binary itself is the
deliverable. These commands manage *that* binary's presence on the system.

> **v1 model — read before relying on `repair`/`upgrade`**: `plaesy install`
> copies the currently-running binary to a well-known per-OS directory.
> `plaesy repair` and `plaesy upgrade` are **not yet implemented** — the old
> bash/PowerShell self-update flow depended on git-cloning the repo fresh,
> which no longer applies, and a proper replacement needs a release/download
> pipeline that doesn't exist yet. Both commands print a clear
> "not yet implemented" message rather than attempting a half-working
> self-update. To upgrade today: build a new binary from source and run
> `plaesy install` again.

## Usage

There is no published release yet, so start by building from source:

```bash
git clone https://github.com/plaesy/spec-kit.git
cd spec-kit/scripts
go build -o plaesy ./cmd/plaesy
```

Then:

```bash
# Install: copy the running binary into a well-known bin directory
./plaesy install

# Status: report install location, whether it's on PATH, and version
plaesy status

# Uninstall: remove the installed binary (prompts y/N)
plaesy uninstall
```

## Install location

| OS | Directory |
|----|-----------|
| Linux / macOS | `$HOME/.local/bin` |
| Windows | `%LOCALAPPDATA%\Plaesy\bin` |

`plaesy install` copies itself there. If that directory isn't already on
`PATH`, the command prints instructions for adding it — it does not modify
shell rc files or the Windows registry for you.

## `plaesy status`

Reports:
- The resolved install directory and whether the binary exists there
- Whether that directory is currently on `PATH`
- The current version (`internal/common.Version`, set at build time via
  `-ldflags`, falls back to `0.0.0` when built without it)
- Whether the current working directory is inside a git repository (a
  leftover general-purpose check from the shared `internal/common` package,
  useful when running `plaesy status` to sanity-check your environment
  generally, not specific to installation)

## `plaesy uninstall`

Removes the installed binary after a confirmation prompt (`y/N`, default
`N`, read from stdin). Does not remove any project's `.plaesy/` directory —
only the `plaesy` binary itself.

## `plaesy repair` / `plaesy upgrade`

Print a "not yet implemented — re-run `plaesy install` with the latest
release binary" message. See the v1 model note above for why.

## Troubleshooting

**`plaesy: command not found` after `plaesy install`**
```bash
plaesy status   # confirms the install dir and whether it's on PATH
```
If it reports the install directory is not on PATH, add it yourself (e.g.
append `export PATH="$HOME/.local/bin:$PATH"` to your shell rc file on
Linux/macOS, or add `%LOCALAPPDATA%\Plaesy\bin` to your Windows PATH
environment variable), then restart your shell.

**Need to upgrade**
```bash
cd spec-kit/scripts && git pull && go build -o plaesy ./cmd/plaesy && ./plaesy install
```

## CI/CD Integration

```yaml
- name: Build and install plaesy
  run: |
    cd scripts && go build -o plaesy ./cmd/plaesy
    ./plaesy install
```

## Migration note

The old `scripts/bash/install.sh` / `scripts/powershell/install.ps1` did
substantially more: environment/dependency validation (`git`, `curl`, `jq`),
`git clone --depth 1` of the whole repo into `~/.plaesy`, a generated shell
wrapper at `~/.local/bin/plaesy` that dispatched subcommands to
`~/.plaesy/scripts/bash/*.sh`, and self-repair/upgrade by re-downloading and
re-running `install.sh`. None of that infrastructure is needed anymore: the
Go binary has every command built in, so "installing" is just "put this one
file somewhere on PATH." If you relied on the old `INSTALL_DIR`, `VERBOSE`,
`SKIP_VALIDATION`, `DEV_MODE`, or `FORCE` environment variables, they have no
equivalent in the Go port — say so if you need one back.

## Related Commands

- **`plaesy init`** — sets up a project's `.plaesy/` structure once the binary is installed (see [plaesy-init.md](./plaesy-init.md))
- **`plaesy config detect-platform`** — used internally during setup (see [config-manager.md](./config-manager.md))
