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

> **v1 model**: `plaesy install` copies the currently-running binary to a
> well-known per-OS directory. `plaesy repair` and `plaesy upgrade` copy the
> running binary over the installed one when they differ, and re-extract the
> bundled assets. No network is involved, so there is no release pipeline to
> depend on: the source is the build you run.

## Usage

The release binaries are published on GitHub Releases, and the installer
scripts below are the easiest way to get them. To build from source instead:

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

### Nushell (curl only)

`install.nu` runs under nushell on Windows, Linux and macOS, so it does not
need `bash` on PATH:

```nu
nu -c (curl -fsSL https://raw.githubusercontent.com/plaesy/spec-kit/main/install.nu)
```

### Installer scripts

`install.sh` (Linux, macOS, Git Bash), `install.ps1` (Windows PowerShell 5.1
and 7+) and `install.nu` (nushell) do the same three steps:

1. Download the release binary and `SHA256SUMS` from the same release.
2. Verify the SHA-256 checksum. A missing or mismatched checksum stops the
   install before anything is copied.
3. Run `plaesy install`, which owns the install directory and PATH advice.

Environment variables, honored by all three:

- `PLAESY_VERSION`: release tag to install, e.g. `v0.0.1`. Default: latest.
- `PLAESY_BASE_URL`: directory holding the assets. Overrides the GitHub URL,
  for testing or a mirror.

`install.nu` is the one-liner for nushell (no `bash` needed):

```nu
nu -c (curl -fsSL https://raw.githubusercontent.com/plaesy/spec-kit/main/install.nu)
```

`install.ps1` adds the install directory to the Windows user PATH if it is not
already there. `install.sh` appends an `export PATH` line to `~/.bashrc`, `~/.bash_profile`
on macOS, or `~/.zshrc` if the directory is not already on PATH. Other shells
get the line printed instead. Open a new terminal afterwards. The install
directory is the one in the table above.

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

Removes the installed binary, the Plaesy home (`AppDataDir()`, which contains
the `.plaesy/` assets `plaesy install` extracted), and the PATH entry the
installers added for the install directory, after a confirmation prompt (`y/N`,
default `N`, read from stdin).

- **Windows:** the entry is removed from the user PATH (`HKCU\Environment`).
  The machine PATH is never touched.
- **macOS/Linux:** only the two-line block `install.sh` wrote is removed from
  `~/.bashrc`, `~/.bash_profile` (macOS) and `~/.zshrc`. Other lines are left as
  they are, including an export the user wrote by hand.

Project-level `.plaesy/` directories in your repositories are never touched.

## `plaesy repair` / `plaesy upgrade`

- **`plaesy repair`** makes the installed binary and the Plaesy home match the
  running build. Use it when the installed copy is missing, or a bundled asset
  was deleted or edited. It does not change which version is installed beyond
  matching the running build.
- **`plaesy upgrade`** replaces the installed binary with the running build and
  refreshes the Plaesy home. Run it from a newer build (for example
  `scripts\plaesy.exe upgrade`). Run from the installed copy itself, it prints
  that there is no newer build to upgrade from.

Both report "already up to date" when nothing changed.

## Troubleshooting

### `plaesy: command not found` after `plaesy install`

```bash

plaesy status   # confirms the install dir and whether it's on PATH
```

If it reports the install directory is not on PATH, add it yourself (e.g.
append `export PATH="$HOME/.local/bin:$PATH"` to your shell rc file on
Linux/macOS, or add `%LOCALAPPDATA%\Plaesy\bin` to your Windows PATH
environment variable), then restart your shell.

### Need to upgrade

```bash

cd spec-kit/scripts && git pull && go build -o plaesy ./cmd/plaesy && ./plaesy upgrade
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
- **`plaesy platforms detect`** — used internally during setup (see [config-manager.md](./config-manager.md))
