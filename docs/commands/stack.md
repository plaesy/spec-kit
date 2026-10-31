# `stack`

```bash

plaesy stack [subcommand] [flags]
```

Inspect a target project's technology stack.

| Subcommand | Args | Description |
|---|---|---|
| *(none)* | — | Print this list |
| `detect` | `[target-dir]` | List instruction files relevant to that stack |

`stack detect` was the bare `plaesy detect-stack`. "detect-stack" fused the two
things the rest of this tree was un-bundling: it detects a technology *stack*,
from marker *files* in a target project. Both are nouns here — `platforms` is a
target, a `stack` is a target — and `detect` is a verb, so `stack detect` says
the same thing while matching `features validate`, `images create` and `context
update`. It also read as a typo'd `detect --stack`.

`stack` currently has one subcommand, which is usually a sign a parent is
premature. It earns the wrapper because the noun is the stable part: a future
`stack explain` or `stack check` then lands without another top-level rename.

## `stack detect`

List instruction files relevant to a target project's tech stack, using
`instructions/mapping.json` as the keyword registry.

| Flag | Default | Description |
|---|---|---|
| `--plaesy-root` | `""` | Path to Plaesy repo root (`$PLAESY_ROOT`, then git root) |
| `--install` | `false` | Copy each detected instruction file into `.plaesy/instructions/` (the `.instructions` suffix is stripped). Existing files are never overwritten. |

**Examples**:
`plaesy stack detect .` — prints one instruction filename per line.
`plaesy stack detect . --install` — prints the same list and copies each file in.

See [CLI Reference](../reference.md).
