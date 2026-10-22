# plaesy features

The feature workflow: create one, list them, print the current one's paths,
check it is ready for task generation, and clean up the worktrees of merged ones.

Source: `scripts/cmd/plaesy/features.go` (verb) + `create_feature.go`,
`features_clean.go` / `feature_paths.go` / `features_validate.go` (subcommands),
`scripts/internal/featurepath/create.go`, `worktree.go`, `clean.go`,
`scripts/internal/common/specs.go`.

| Subcommand | Documented in |
|---|---|
| `plaesy features` (list) | below |
| `plaesy features create` | below |
| `plaesy features clean` | below |
| `plaesy features paths` | [features-paths.md](./features-paths.md) |
| `plaesy features validate` | [features-validate.md](./features-validate.md) |

> **Renamed.** These were three top-level commands — `plaesy create-new-feature`,
> `plaesy get-feature-paths` and `plaesy check-task-prerequisites` — in three
> different shapes, none of which said they belonged together. All three
> spellings were **removed**, not aliased: they had no deprecation window, and a
> leftover alias would be invisible, since deprecated commands are hidden from
> help and from the documentation bar the drift tests enforce. Update any
> script or prompt that calls them.

## Naming

A feature has two names that differ by a timestamp:

- **Branch**: `auth-login` — the first three words of the description, lowercased
  and joined by hyphens.
- **Spec folder**: `.plaesy/specs/20261015-143022-auth-login/` —
  `{TIMESTAMP}-{BRANCH}`, where the timestamp is the UTC creation time as
  `YYYYMMDD-HHMMSS`.

The timestamp means a new feature never needs a scan for the highest number,
and two machines creating features at once cannot pick the same number. Features
created before timestamps used `NNN-slug` folders and branches. Those still
resolve, because a spec is matched to its branch by its name without any
timestamp prefix.

## `features create`

### Purpose

Given a feature description, the command derives a branch name, creates the
branch, creates the spec folder `{TIMESTAMP}-{BRANCH}` with its `tasks/` status
folders, and seeds `spec.md` from the spec template.

### Quick Start

```bash

plaesy features create "Add user authentication system"
plaesy features create --worktree "Add billing page"
plaesy features create --json "Add user authentication system"
```

### Options

| Flag | Effect |
|---|---|
| `--worktree` | Create the feature in its own git worktree, outside the repository, so several features can be in progress at once |
| `--json` | Emit the result as a single-line JSON object instead of plain text |
| `--help` / `-h` | Print usage and exit |

### Behavior

#### Without `--worktree` (default)

1. Resolves the repo root with `git rev-parse --show-toplevel`, and creates `.plaesy/specs/` if needed.
2. Runs `git checkout -b <branch>` in the current working tree. It fails if the
   branch exists or git cannot switch branches.
3. Creates `.plaesy/specs/<branch>/` with `spec.md`. Tasks live in `.plaesy/tasks/`
   and `spec.md`.

#### With `--worktree`

1. Runs `git -C <main checkout> worktree add -b <branch> <path>`. The new branch
   starts from the main checkout's current commit.
2. The worktree path is `{WorktreeHome}/{project}-{hash8}/{spec}/`. See
   [Worktree location](#worktree-location).
3. Creates the spec folder inside the worktree, on the feature branch.
4. Reads `spec.md`'s template from the main checkout, because templates are
   git-ignored and a worktree does not have them.

If step 3 fails, the worktree and its branch are removed, so a failed create
leaves nothing behind.

### Worktree location

Worktrees are kept outside every repository, so a worktree never nests in the
tree it came from.

| Platform | Location |
|---|---|
| Any, when `PLAESY_WORKTREE_HOME` is set | that directory |
| Windows | `%LOCALAPPDATA%\Plaesy\worktrees\` |
| Linux and macOS | `$XDG_DATA_HOME/plaesy/worktrees/`, or `~/.local/share/plaesy/worktrees/` |

Under that root, each project has `{project}-{hash8}`, where `hash8` is the
first eight hex digits of the SHA-256 of the main checkout's path. Two
repositories with the same folder name therefore do not share worktrees. Moving
or renaming the repository changes the hash, so its existing worktrees are no
longer listed by `features clean`. Run `git worktree list` to find them.

### Ignored spec warning

When git ignores the new `spec.md`, the command still succeeds and prints a
`WARNING:` on stderr. This happens when a global gitignore covers `.plaesy/`. An
ignored spec is never committed, and `features clean` will not remove its
worktree.

### Output Format

Plain text (default):

```text

BRANCH_NAME: auth-login
SPEC_NAME: 20261015-143022-auth-login
SPEC_FILE: /repo/.plaesy/specs/20261015-143022-auth-login/spec.md
WORKTREE_PATH: C:\Users\you\AppData\Local\Plaesy\worktrees\repo-1a2b3c4d\20261015-143022-auth-login
```

`WORKTREE_PATH` is printed only with `--worktree`.

JSON (`--json`):

```json

{"BRANCH_NAME":"auth-login","SPEC_NAME":"20261015-143022-auth-login","SPEC_FILE":"/repo/.plaesy/specs/20261015-143022-auth-login/spec.md","WORKTREE_PATH":""}
```

> **Breaking change:** the `FEATURE_NUM` key is gone. It held the `NNN` number,
> which no longer exists. Read `SPEC_NAME` instead.

## `features clean`

Lists the feature worktrees of this project. With `--apply`, it removes the
merged ones.

```bash

plaesy features clean            # dry run: lists, changes nothing
plaesy features clean --apply    # removes merged worktrees
```

A worktree is **merged** when its branch is an ancestor of `main` (or `master`).
A worktree is **removed** only when all of these hold:

- its branch is merged,
- it has no git-ignored files, because git would delete them without asking,
- `git worktree remove` accepts it. Git refuses a worktree with uncommitted or
  untracked changes, and `clean` reports that refusal rather than forcing it.

`clean` never uses `--force`, and it never deletes branches.

**Squash-merged branches are not ancestors of `main`**, so they show as not
merged and are kept. Remove them by hand with `git worktree remove` once you
have checked them.

## `features` (list)

Lists every spec folder, with the design documents present in each and `*` on
the spec of the checked-out branch. Folders are ordered by name, which puts the
timestamps in chronological order.

**Example**: `plaesy features`

## Troubleshooting (features)

**"feature description is required" error** — no description was provided; pass
one as the (only) non-flag argument.

**"feature description has no letters or digits" error** — the description has
no usable characters, so no branch name can be made. Give it some words.

**`git checkout -b` fails with "branch already exists"** — a branch with that
name already exists. Use a different description, or delete the stale branch.

**`git worktree add` fails** — the worktree path already exists, or the branch is
already checked out in another worktree. Run `git worktree list` to see which.

**Not run inside a git repository** — `git rev-parse` fails; run the command from
inside the repo.

**`WARNING: git ignores ...`** — see [Ignored spec warning](#ignored-spec-warning).

## Related Commands

- **`plaesy features paths`** — resolves paths for an existing feature (see [features-paths.md](./features-paths.md))
- **`plaesy features validate`** — validates feature setup before task generation (see [features-validate.md](./features-validate.md))
- **`plaesy tasks`** — the tasks of a spec, under `.plaesy/tasks/`
- **`plaesy context update`** — updates AI context for the new feature (see [context-update.md](./context-update.md))
- **`plaesy images create`** — generates an image asset (see [images-create.md](./images-create.md))
