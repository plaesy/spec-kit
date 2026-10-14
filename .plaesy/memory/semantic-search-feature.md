---
title: "Semantic Search Feature — Gotchas"
description: "go.mod toolchain bump via go get, assets/markdown resync after editing instructions, CLI verb-naming convention, pre-existing vet failures disable two packages' tests, parallel agents in one Go package, verify agent claims against disk"
updatedAt: "2026-10-03T19:50:00.000Z"
---

# Semantic Search Feature — Gotchas

Built while adding `plaesy search` (embedding-based semantic search over
extracted symbols, for the Anti-Duplication Protocol in
`instructions/plaesy.instructions.md`). Full decision trail:
[embedding-based-semantic-duplicate-search](decisions/embedding-based-semantic-duplicate-search.md).

## `go get` can silently bump `go.mod`'s `go` directive repo-wide

Adding `github.com/rostamlabs/rembed` forced `scripts/go.mod`'s `go`
directive from `1.20` to `1.26.1` — not scoped to the new package, it
raises the minimum Go version for the *entire module*, every contributor
and CI run. `go build ./...` succeeding after a `go get` is not proof
nothing else changed — always `git diff scripts/go.mod` right after adding
a dependency and surface the version bump explicitly before assuming it's
free. This repo's CI uses `go-version-file: scripts/go.mod`, so the bump
self-propagates to CI with no separate workflow edit needed — but a reader
needs to know that's *why* no CI edit was made, not assume it was missed.

## Editing `instructions/*.md` or `prompts/*.md` has two follow-up steps, not one

After any edit to a file under `instructions/` or `prompts/`, both of these
are required before the change is actually "done", and forgetting either
fails CI even though `go build`/`go test` on the Go code itself is green:

1. `go run ./internal/assets/gen` — resyncs the embedded copy under
   `scripts/internal/assets/data/`; `internal/assets`'s
   `TestAssetsDataMatchesSource` fails otherwise.
2. `go run ./cmd/plaesy validate markdown` — the repo's markdown ratchet is
   zero-tolerance (`internal/mdlint`'s `TestRepoMarkdownStaysWithinRatchet`),
   so a single `MD013` (line too long) or similar violation in the edited
   file fails CI.

Run both after every instructions/prompts edit, not just once at the end of
a session — cheap to run, expensive to discover failing in CI.

## CLI naming: a distinct verb gets its own top-level command

First cut of `plaesy search` was `plaesy graph --semantic-build` /
`--semantic-query` (flags on the existing `plaesy graph` builder command).
User found this unfamiliar/undiscoverable and asked for research into
alternatives; the fix was a dedicated top-level `plaesy search` command —
consistent with `gh search`, `rg <pattern>`: a "find things" verb is its
own command, not flags bolted onto an unrelated builder command's flag
namespace. Worth checking this convention *before* adding new flags to an
existing `plaesy` subcommand for a genuinely new verb, rather than after a
user correction.

## Pre-existing `go vet` failures are not yours to fix — but they are not harmless

`cmd/plaesy/reload.go` and several `internal/scaffold/*.go` files fail
`go vet` (`non-constant format string in call to ...LogInfo` etc.). They are
untouched since the initial commit: `git status --porcelain` on them is empty
and `git log --oneline -- <files>` shows only `1d46bea Initial Commit`.
(The earlier version of this note claimed `git status` showed them "already
modified" as the evidence — that was wrong reasoning, `M` just means "differs
from HEAD", and HEAD *is* the initial commit.)

So don't spend feature time fixing unrelated pre-existing breakage. **But do not
dismiss it either**, because `go test` runs vet before compiling: these 52 errors
make `cmd/plaesy` and `internal/scaffold` report `FAIL ... [build failed]` and
**none of their tests execute at all**. Any confidence in those two packages was
unfounded until a `-vet=off` run measured them. Full treatment under
[Testing Discipline](testing-discipline.md) → "Reading a `go test ./...` result
correctly".

## Parallel agents on one Go package: prefix every identifier

Adding 11 language extractors across 4 concurrent agents in a single worktree
worked, but only under three rules:

1. **One writer per file.** The extension dispatch (`docsFor`/`docsByExt` in
   `semantic.go`) and `docs/scripts/plaesy-search.md` were owned by the
   orchestrator alone; agents returned the doc text as text instead of editing
   it. Two agents editing one file in one worktree is last-write-wins, not a
   merge.
2. **Prefix every new identifier per agent** (`rust*`, `php*`, `csharp*`, …).
   Four agents adding helpers to one package collide on any generic name, and a
   duplicate declaration breaks the build for everyone. Also designate one
   explicitly-shared helper (e.g. `hashCommentLinesAbove`) or two agents will
   independently pick the same natural name for it.
3. **Tell agents to ignore failures in files they don't own**, and expect it to
   be necessary — they will otherwise try to "fix" each other.

Grouping the languages by *comment convention* (rustdoc | `#` comments |
Javadoc/KDoc | C-style block) was what made the split clean: within a group the
scanner is shared, so only the name matcher differs per language.

**Verify agents' claims against disk.** One agent's final report was a
corrupted tool-call echo, and it had in fact never written its test file — the
extractor was complete and completely untested. Two other agents reported work
green that was not (a redeclared test helper that broke the package build; a
test asserting a contract that did not exist).

## A test that cannot fail is worse than no test

The first version of the regex-drift guard asserted only that every keyed doc
name appears in `graph.Node.Symbols`. I verified it by deliberately breaking a
regex — and it **passed**. Breaking a regex makes an extractor produce *fewer*
docs, and the survivors are still validly keyed, so a subset check sails
through while enrichment has quietly rotted. That is the exact silent-failure
mode the guard existed to catch, and the test had it.

Fix: assert both directions. The fixture must name the symbols it documents and
all of them must get a doc; every keyed name must be a real graph symbol; and
an undocumented fixture symbol must stay out. Then re-run the deliberate-break
check — a guard that has never been seen to fail is unverified.

Corollary: when a test fails, check whether the *fixture* is wrong before
blaming the code. Several C-family fixtures here used `int f(void)` and bare
prototypes; `graph.Build` extracts no symbol from either, so the extractor was
correct and the fixture was not. Verify with the real `graph.Build` rather than
reasoning about the regex.
