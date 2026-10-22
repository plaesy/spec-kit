# `features`

```bash

plaesy features [subcommand] [flags]
```

The feature-branch workflow: create one, list them, print the current one's
paths, and check it is ready for task generation. With no subcommand, lists the
features that exist.

| Subcommand | Args | Description |
|---|---|---|
| *(none)* | — | List every feature in `.plaesy/specs/` |
| `create` | `<feature_description>` | Branch, directories, and `spec.md` |
| `paths` | — | Shell-sourceable paths for the current feature |
| `validate` | `[--json]` | Is the current feature ready for tasks |

These were three top-level commands — `create-new-feature`, `get-feature-paths`
and `check-task-prerequisites` — in three different shapes, none of which said
they belonged together. All three spellings were **removed**, not aliased: they
had no deprecation window, and a leftover alias would be invisible, since
deprecated commands are hidden from help and from the documentation bar the
drift tests enforce.

## `features create`

```bash

plaesy features create <feature_description> [--worktree] [--json]
plaesy features clean [--apply]
```

`create` names the branch after the first three words of the description
(`auth-login`) and the spec folder `{TIMESTAMP}-{BRANCH}`
(`.plaesy/specs/20261015-143022-auth-login/`), with `tasks/` and `spec.md`.
`--worktree` makes the feature in its own git worktree outside the repository,
so several features can be in progress at once. `clean` lists the feature
worktrees and, with `--apply`, removes the merged ones. Details:
[features.md](../scripts/features.md).

| Flag | Default | Description |
|---|---|---|
| `--worktree` | `false` | Create the feature in its own worktree under the worktree home |
| `--json` | `false` | Output as JSON |

**Example**: `plaesy features create --worktree "Add dark mode toggle"`

## `features` (list)

Lists every spec folder, ordered by name (so timestamps run in chronological
order), with the design documents present in each and `*` on the spec of the
checked-out branch. On a project with no `.plaesy/specs/`, prints the command to
create one rather than an empty block.

**Example**: `plaesy features`

## `features paths`

```bash

plaesy features paths
```

Print shell-sourceable paths (`REPO_ROOT=`, `FEATURE_DIR=`, etc.) for the
current feature branch. No arguments. Reads from `internal/featurepath`.

**Example**: `eval $(plaesy features paths)`

## `features validate`

Checks the current feature has a `plan.md` and reports which design documents
are available. With no feature directory it is an error, not an empty report.

| Flag | Default | Description |
|---|---|---|
| `--json` | `false` | Output as JSON |

**Example**: `plaesy features validate --json`

See [CLI Reference](../reference.md).
