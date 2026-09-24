# Plaesy Constitution Kit — CLI Overview

**Version**: 0.0.1 | **Language**: Go 1.20+ | **Framework**: [Cobra](https://github.com/spf13/cobra) | **Dependencies**: stdlib only (cobra, pflag, mousetrap)

## Purpose

Plaesy Constitution Kit is an AI development framework whose enforcement is
driven by a project constitution (`.plaesy/memory/constitution.md`) — a single
governing source of truth. The `plaesy` CLI is the tooling layer that scaffolds
projects, detects AI platforms, analyzes code, builds knowledge graphs, manages
task lifecycles, and compresses context — all from one cross-platform binary.

## Scope

`plaesy` replaces the legacy `scripts/bash/*.sh` and
`scripts/powershell/*.ps1` dispatch scripts with a single Go binary
(`scripts/cmd/plaesy/`). It operates on **any** project directory, not just
spec-kit itself — it is designed to be cloned, built, and run against arbitrary
codebases to generate analysis, detect tech stacks, and inject platform-specific
AI configuration.

## Core Features

| Feature | Command(s) | Internal Package |
|---|---|---|
| Project scaffolding | `plaesy init` | `internal/scaffold` |
| AI platform detection | `plaesy config detect-platform` | `internal/config` |
| Tech-stack instruction selection | `plaesy detect-stack` | `internal/detectstack` |
| Project analysis + fingerprinting | `plaesy analyze` | `internal/analyze` |
| Knowledge graph builder + queries | `plaesy graph` | `internal/graph` |
| Binary install / status / uninstall | `plaesy install`, `plaesy status`, `plaesy uninstall` | `internal/installer` |
| Token/context compression | `plaesy trim` | `internal/trimmer` |
| Platform file cleanup | `plaesy clean` | `internal/cleaner` |
| AI header injection | `plaesy inject-ai-headers` | `internal/aiheaders` |
| Task lifecycle management | `plaesy task-manage` | `internal/taskmanage` |
| Feature branch scaffolding | `plaesy create-new-feature`, `plaesy get-feature-paths` | `internal/featurepath` |
| Agent context sync | `plaesy update-agent-context` | `internal/agentcontext` |
| File validation | `plaesy validate-*` | `internal/validate` |
| Image generation | `plaesy generate-image` | `internal/imagegen` |

## Tech Stack

- **Language**: Go (stdlib only — no external HTTP/db dependencies)
- **CLI framework**: Cobra (command registration, flag parsing)
- **Configuration**: JSON (`scripts/configs/platform.json`, `instructions/mapping.json`)
- **Supported platforms**: Linux, macOS, Windows (native; MSYS2 path normalization included)
- **Supported AI platforms** (18): claude_code, opencode, cursor_ai, github_copilot, cline, deepseek, kilo_code, qoder, trae_ai, windsurf_ai, continue_dev, tabnine, codeium, codewhisperer, studio_bot, replit_ghostwriter, llama_index, ollama, lm_studio, generic_ai

## Quick Start

```bash
# Build from source
cd scripts && go build -o plaesy ./cmd/plaesy

# Install to well-known bin directory
./plaesy install

# Scaffold a new project
plaesy init my-project --ai claude_code

# Analyze the project
plaesy analyze
```

> **Note on `repair` / `upgrade`**: These commands print "not yet implemented."
> The Go binary model has no git-clone-based self-update. To upgrade: rebuild
> from source and re-run `plaesy install` (see `docs/scripts/install.md`).

## Quick Links

| Document | Purpose |
|---|---|
| [architecture.md](./architecture.md) | High-level design + component relationships |
| [components.md](./components.md) | Per-package responsibility + implementing paths |
| [reference.md](./reference.md) | CLI command reference (usage, flags, examples) |
| [docs/scripts/README.md](./scripts/README.md) | Per-command guides (AI assistant quick-reference) |
