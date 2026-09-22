# Changelog

All notable changes to Plaesy Spec-Kit are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and
this project uses [Semantic Versioning](https://semver.org/) once it reaches 1.0.0.

## [Unreleased]

- Improved `plaesy-analyze` cross-platform performance and Bash/PowerShell parity:
  cached language/framework detection, single-pass file counting, portable macOS paths,
  and removal of unused generated test-runner behavior.
- Clarified analyzer output documentation and corrected stale instruction-loading claims.
- Added `--if-changed` / `-IfChanged` fast path to `plaesy-analyze`: skips all
  regeneration when project fingerprint (file count + newest mtime + framework
  version) matches the last run. Excludes `.plaesy/` to avoid circular dependency.
- Fixed GNU-only `find -printf` in `plaesy-graph.sh` with portable macOS/BSD fallback
  via `stat -f %m`.
- PowerShell analyzer now mirrors Bash detection for development tools and build
  systems (previously hardcoded `@("Git")` / `@("Manual")`).
- Added functional analyzer smoke tests (`testing/smoke/smoke-analyze.sh` +
  `.ps1`) and CI parity job comparing Bash/PowerShell analyzer output.

## [0.0.1] - Initial release

- Initial public commit of the Plaesy Spec-Kit framework: prompts, instructions,
  chatmodes/roles, checklists, templates, and bash/PowerShell automation scripts.
