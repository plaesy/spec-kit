---
title: "plaesy init self-provisions its home instead of requiring a prior install/clone"
updatedAt: "2026-10-02T21:50:00.000Z"
---

# Self-provisioning Plaesy home

**Decision**: `scripts/internal/scaffold/home.go`'s `FindHome()` now has a
final fallback: when no override, `PLAESY_HOME`, or ambient spec-kit checkout
is found, it extracts the Plaesy source tree (`templates/`, `instructions/`,
`prompts/`, `agents/`, `checklists/`, `scripts/configs`) — embedded into the
binary at build time via `go:embed` (`scripts/internal/assets/`) — to
`{AppData}/Plaesy/.plaesy` (Windows) / `~/.local/share/plaesy/.plaesy`
(Unix), then uses that as home. This happens lazily, on first need, from
*any* invocation (`init`, `reload`, …), regardless of how the binary reached
the machine.

**Why**: The v1 Go installer's own doc comment says it plainly — "the running
Go binary itself is the deliverable… `install` simply copies the current
binary to a well-known bin directory." It never carried the source tree.
`plaesy init` only worked if the binary happened to be running from inside (or
below) a full spec-kit git checkout. A user who installed via
`curl … | bash` (which only places the downloaded release binary) hit `could
not locate Plaesy repo root` on the very first `plaesy init`, with no
indication that `plaesy install`'s asset-extraction step was the thing they
were missing — because `install.sh` never calls it.

**Two provisioning paths, not one**: `installer.Install()` (the `plaesy
install` subcommand) *also* extracts the same embedded assets, eagerly, every
time it runs — so a user who does run `plaesy install` always has a current
home immediately, and an upgrade refreches it. The lazy `FindHome` fallback
exists for every *other* path onto the machine (curl script, a package
manager, manual binary copy) that never calls `install`. Both call the same
`assets.Extract()`; the home directory layout is identical either way, so
nothing downstream (`copy.go`, `platform.go`) needed to change.

**Rejected alternative**: Moving the home from "$HOME/templates,
$HOME/scripts, …" (what the original ask was — "I want home under `.plaesy/`
too, like the project output") directly into `$HOME/.plaesy/` was rejected
after flagging a real collision risk: a project's own `.plaesy/` output
(`.plaesy/templates/`, `.plaesy/scripts/configs/platform.json`, …) is
structurally identical to what a "home" looks like, so `FindHome`'s walk-up
search could mistake an already-`plaesy init`'d project for a home. The chosen
location — a dedicated app-data directory the user never touches directly —
has no such collision surface.

**Verified**: built a fresh binary, set `LOCALAPPDATA` to an empty temp
directory, cleared `PLAESY_HOME`, ran `plaesy init . --ai claude` from a
directory with no ambient checkout — succeeded, and the extracted home
appeared at `{temp}/Plaesy/.plaesy/{agents,checklists,instructions,prompts,
scripts,templates}`. Confirmed again live against the user's own
previously-curl-installed binary on a separate project
(`nara/client`): `plaesy install` now prints "Plaesy home ready at …", and
`plaesy reload --ai claude` resolves without `--plaesy-home`.
