---
title: "Installers add the plaesy bin directory to PATH (reverses 'never edit PATH')"
updatedAt: "2026-10-20T00:00:00.000Z"
---

# Installers add the plaesy bin directory to PATH

**Decision**: `install.ps1` adds `%LOCALAPPDATA%\Plaesy\bin` to the Windows user
PATH after `plaesy install` succeeds, if it is not already there. `install.sh`
appends an `export PATH="$HOME/.local/bin:$PATH"` line to `~/.bashrc` (Linux) or
`~/.bash_profile` (macOS) for bash, and to `~/.zshrc` for zsh, if the directory
is not already on PATH or already mentioned in that file. Other shells are not
modified; the line is printed instead. Both scripts skip the change when the
directory is already present.

**Why**: The previous behaviour printed PATH instructions only. A user who ran
`irm … | iex` or `curl … | bash` had to copy the printed line by hand, and the
binary was not callable right after install. The user asked for the one-step
install. This reverses the earlier rule recorded in `docs/scripts/install.md`
("The installers never edit shell profiles or the Windows user PATH").

**Scope**: The two installer scripts change. `plaesy install`
(`scripts/internal/installer/installer.go`) still only prints PATH advice, so a
binary installed by hand is not modified.

**Uninstall**: `plaesy uninstall` removes the PATH entry this change adds, so
uninstall leaves no trace. On Windows it removes the entry from the user PATH
(`HKCU\Environment`, never the machine PATH). On macOS/Linux it removes only the
two-line block `install.sh` wrote (marker line plus export), and leaves other
lines alone. Uninstall stops early, before the PATH step, when neither the
binary nor the Plaesy home exists; a PATH entry left after a manual delete of
both is not cleaned by this command.

**Verification**: `install.ps1` parses with no errors, and its `Add-UserPath`
function was run against a directory already on PATH (no change). `install.sh`
passes `bash -n`, and `add_to_path` was run against a temporary HOME for bash
and zsh (no duplicate line on a second run) and for fish (printed instruction
only). The full download-and-install path was not run, because it needs
published release assets.

**References**:
- `install.ps1` (Add-UserPath, call after `plaesy install`), `install.sh` (add_to_path), in this repo
- `docs/scripts/install.md` (Windows and Unix PATH paragraph), updated in this change
- Earlier rule replaced: `docs/scripts/install.md` before 2026-10-20, and the header comment of `install.ps1`

**Related**: [[self-provisioning-plaesy-home]]
