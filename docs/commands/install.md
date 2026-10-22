# `install`

```bash

plaesy install
```

Install the `plaesy` CLI to a well-known bin directory. Copies the currently-
running binary to `%LOCALAPPDATA%\Plaesy\bin\plaesy.exe` (Windows) or
`$HOME/.local/bin/plaesy` (Unix). It also extracts the embedded Plaesy home
(`templates/`, `instructions/`, `prompts/`, `agents/`, `checklists/`,
`configs`) to `%LOCALAPPDATA%\Plaesy\.plaesy` (Windows) or
`$HOME/.local/share/plaesy/.plaesy` (Unix), so `plaesy init` works without a
spec-kit checkout of its own — `FindHome` falls back to this location when no
override, `PLAESY_HOME`, or ambient checkout is found. If the bin directory is
not on PATH, prints instructions for adding it (does not modify shell rc or
registry).

**No flags.** Reads its own executable path via `os.Executable()`.

See [CLI Reference](../reference.md).
