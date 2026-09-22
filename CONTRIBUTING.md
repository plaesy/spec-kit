# Contributing to Plaesy Spec-Kit

Thanks for your interest in contributing! This project is a spec-driven-development
framework distributed as markdown prompts/instructions plus a single cross-platform
Go CLI (`plaesy`).

## Before you start

- Search existing [issues](https://github.com/plaesy/spec-kit/issues) and
  [pull requests](https://github.com/plaesy/spec-kit/pulls) to avoid duplicate work.
- For a non-trivial change, open an issue first to discuss the approach.

## Development setup

1. Fork and clone the repository.
2. Install [Go](https://go.dev/dl/) 1.22+. The CLI source lives under `scripts/`
   (module root `scripts/go.mod`): commands in `scripts/cmd/plaesy/`, implementation
   packages in `scripts/internal/`. There is one codebase for every platform —
   Linux, macOS, and Windows all build and run the same Go source, so there's no
   bash/PowerShell parity to maintain anymore.
3. Run before opening a PR (from the repo root):
   - `cd scripts && go build ./... && go vet ./... && go test ./...`
4. `make build` compiles the CLI to `./plaesy` at the repo root. `make` (or
   `make reload`) re-initializes this repo's own dogfooded `.plaesy/` setup — its
   `reload` target **unconditionally deletes `.claude/`, `.plaesy/`, and
   `CLAUDE.md`** before re-running `./plaesy init`, with no confirmation prompt.
   Commit or back up anything under those paths first; don't run bare `make`
   expecting a build/test step.

## Conventions

- Prompts: `prompts/*.md`
- Instructions: `instructions/*.instructions.md`
- Chatmodes/roles: `chatmodes/*.chatmode.md`
- Checklists: `checklists/*.checklist.md`
- Templates: `templates/*.template.md`

Keep new files consistent with these suffix conventions so tooling and docs indexes
can discover them.

## Security-sensitive changes

Any change touching remote downloads/self-update (`net/http` calls in
`scripts/internal/installer` or `scripts/internal/imagegen`), file permissions, or
credential handling (API keys, tokens) should call this out explicitly in the PR
description. See [SECURITY.md](SECURITY.md) for how to report vulnerabilities
privately instead of via a public issue/PR.

## Pull requests

- Keep PRs focused on one change.
- Update relevant docs (`docs/`) and the root `CHANGELOG.md` under "Unreleased".
- CI must pass (`go build`/`go vet`/`go test` across Linux, macOS, and Windows).

## Code of Conduct

This project follows the [Code of Conduct](CODE_OF_CONDUCT.md). By participating you
agree to abide by it.
