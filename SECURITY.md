# Security Policy

## Supported Versions

Plaesy Spec-Kit is currently pre-1.0 (see `VERSION`). Only the `main` branch /
latest release receives security fixes.

## Reporting a Vulnerability

Please **do not** open a public GitHub issue for security vulnerabilities.

Instead, report it privately via GitHub's
[private vulnerability reporting](https://github.com/plaesy/spec-kit/security/advisories/new)
for this repository. Include:

- A description of the vulnerability and its impact
- Steps to reproduce (a minimal script/command sequence is ideal)
- Affected file(s)/script(s) and, if known, the platform (bash/PowerShell)

We will acknowledge your report as soon as possible and keep you updated as the
issue is investigated and fixed.

## Scope

This project installs and runs shell/PowerShell automation scripts, including
scripts that download further content from `raw.githubusercontent.com` at
install/repair/upgrade time. Reports about the install/repair/upgrade path
(e.g. insufficient validation of downloaded scripts, unsafe `eval`/`Invoke-Expression`
usage, unsafe temp-file handling) are especially welcome.
