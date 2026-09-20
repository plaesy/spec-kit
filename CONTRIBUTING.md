# Contributing to Plaesy Spec-Kit

Thanks for your interest in contributing! This project is a spec-driven-development
framework distributed as markdown prompts/instructions plus bash and PowerShell
automation scripts.

## Before you start

- Search existing [issues](https://github.com/plaesy/spec-kit/issues) and
  [pull requests](https://github.com/plaesy/spec-kit/pulls) to avoid duplicate work.
- For a non-trivial change, open an issue first to discuss the approach.

## Development setup

1. Fork and clone the repository.
2. Scripts live under `scripts/bash/` (Linux/macOS) and `scripts/powershell/`
   (Windows/cross-platform PowerShell). **Any behavioral change to a script must be
   made in both the bash and PowerShell versions** — the two must stay in parity.
3. Run the test suites before opening a PR:
   - `bash testing/bash/run.sh`
   - `pwsh testing/powershell/run.ps1`
   - Smoke tests: `testing/smoke/smoke-e2e.sh` and `testing/smoke/smoke-powershell.ps1`
4. `make` (or `make reload`) re-initializes this repo's own dogfooded `.plaesy/`
   setup — its only target, `reload`, **unconditionally deletes `.claude/`,
   `.plaesy/`, and `CLAUDE.md`** before re-running `plaesy-init.sh`, with no
   confirmation prompt. Commit or back up anything under those paths first;
   don't run bare `make` expecting a build/test step.

## Conventions

- Prompts: `prompts/*.md`
- Instructions: `instructions/*.instructions.md`
- Chatmodes/roles: `chatmodes/*.chatmode.md`
- Checklists: `checklists/*.checklist.md`
- Templates: `templates/*.template.md`

Keep new files consistent with these suffix conventions so tooling and docs indexes
can discover them.

## Security-sensitive changes

Any script change touching `eval`, remote downloads (`curl`/`Invoke-WebRequest`),
file permissions, or credential handling should call this out explicitly in the PR
description. See [SECURITY.md](SECURITY.md) for how to report vulnerabilities
privately instead of via a public issue/PR.

## Pull requests

- Keep PRs focused on one change.
- Update relevant docs (`docs/`) and the root `CHANGELOG.md` under "Unreleased".
- CI must pass (bash lint/smoke, PowerShell lint/smoke).

## Code of Conduct

This project follows the [Code of Conduct](CODE_OF_CONDUCT.md). By participating you
agree to abide by it.
