<div align="center" style="background-color: #f8f9fa; padding: 20px; border-radius: 10px; margin: 20px 0;">
  <img src="https://raw.githubusercontent.com/plaesy/.github/refs/heads/main/assets/img/Logo.svg" alt="Plaesy Constitution Kit - AI Development Framework" />
</div>

# Plaesy: Constitution Kit

**Version:** 0.0.1

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE) [![Go](https://img.shields.io/badge/built%20with-Go-00ADD8.svg)](https://go.dev/)
[![Status: Early](https://img.shields.io/badge/status-early%20(0.0.x)-orange.svg)](VERSION)

## What is Plaesy Constitution Kit?

Plaesy Constitution Kit is an AI prompt framework built around a project **constitution** — a single governing source of truth (`.plaesy/memory/constitution.md`) that every phase (`/assess`,
`/implement`, `/optimize`, `/fix`, `/improve`) reads before acting. Unlike spec-driven tools scoped to code alone, Plaesy drives all eight dimensions of a project from that constitution — technical,
design, business, legal, marketing, financial, management, and product — through workflow automation from idea to production-ready deliverable, with anti-hallucination protocols enforcing
evidence-backed output at each phase.

## Why Plaesy Constitution Kit (vs. other spec-driven tools)

Most spec-driven development tools — including GitHub's own `spec-kit` — scope the "spec" to code: a technical plan for an AI coding agent to implement. Plaesy Constitution Kit treats the
**constitution**, not the spec, as the source of truth, and applies it beyond code: an `/assess:business` or `/assess:legal` run is governed by the same constitution as `/implement:technical`. If your
project needs more than code specified — pricing model, compliance posture, go-to-market messaging — those live in the same governed workflow, not a separate tool.

Agentic Agile frameworks like [BMAD-METHOD](https://github.com/bmad-code-org/BMAD-METHOD) solve an adjacent problem well: role-based agents (Analyst, PM, Architect) collaborating through a structured
planning-to-build loop, with human-in-the-loop checkpoints. That agent-role discipline is one we share (see `.plaesy/roles/`). Where Plaesy differs is scope of governance: BMAD's loop is sized to a
single software delivery effort, while Plaesy's constitution governs every dimension of a project — technical, design, business, legal, marketing, financial, management, product — under one
evidence-backed source of truth, so a `/loop` or `/improve` pass on the business model is held to the same rigor as one on the codebase.

## Quick Start

### Install & Initialize

Plaesy Constitution Kit ships as a single cross-platform Go binary (`plaesy`).

#### Option 1: Automated Installation (Recommended) ⚡

**Linux / macOS:**

```bash

curl -sSL https://raw.githubusercontent.com/plaesy/spec-kit/main/install.sh | bash
```

**Windows PowerShell:**

```powershell

Invoke-WebRequest -Uri "https://raw.githubusercontent.com/plaesy/spec-kit/main/install.ps1" -OutFile install.ps1 -UseBasicParsing; .\install.ps1
```

The install script automatically:

- ✅ Detects your platform (Linux, macOS, Windows) and architecture
- ✅ Downloads the latest binary from GitHub Releases
- ✅ Installs to a well-known bin directory
- ✅ Adds to PATH if needed
- ✅ Verifies installation

#### Option 2: Build from Source (requires [Go](https://go.dev/dl/) 1.20+)

```bash

git clone https://github.com/plaesy/spec-kit.git
cd spec-kit/scripts
go build -o ../plaesy ./cmd/plaesy

# put the binary on PATH, then:
./plaesy init my-awesome-app

# Or initialize in current directory with specific AI
./plaesy init . --ai claude
```

Using any AI assistant with optimized prompts:

```markdown

# Complete automation from idea to production-ready code
/start Build a privacy-first photo organizer that automatically groups images by event, location, and people

# AI will automatically execute ALL phases:
# 1. Project Analysis → 2. Technical Research → 3. Specification Generation
# 4. Implementation → 5. Quality Review → 6. Performance Optimization
# 7. Bug Resolution → 8. Completion Report with next steps

# Resume work anytime - executes remaining phases automatically
/continue
```

### For Existing Projects: Documentation & Assessment

Before adding features to existing codebases, assess and document the project:

```markdown

# Comprehensive project assessment
Chat: "/assess @tw ./legacy-application"

# Security-focused assessment
Chat: "/assess @security ./production-system"

# Architecture assessment for new feature planning
Chat: "/assess @architecture ./microservices-platform"
```

The `/assess` command generates complete project documentation, analysis, and AI-powered insights with autonomous implementation capabilities.

### Commands

Every subcommand, matching `docs/reference.md`. Flags and examples for each one live
there; this list is the index.

```bash

# Linux/macOS CLI
plaesy init                    # Scaffold a new Plaesy project
plaesy init <directory>        # Interactive AI selection in specified directory
plaesy init <directory> --ai <platform>    # Use specific AI platform in directory
plaesy reload                 # Refresh generated .plaesy/ files, keeping memory and context
plaesy analyze                 # Analyze current project structure and generate documentation
plaesy graph                   # Build or query a knowledge graph of the project
plaesy search                  # Embedding-based semantic search over extracted symbols
plaesy index                   # Build or update the search index (top-level alias for `search index`)
plaesy query                   # Search the local index directly (top-level alias for `search query`)
plaesy eval                    # Run the search evaluation harness against a golden set
plaesy doctor                  # Check health of reach platforms, search index, and config
plaesy reach                   # Access external data sources (web, github, youtube, etc.)
plaesy stats                   # Show search index and reach cache statistics
plaesy validate [target]       # Check a constitution, memory, Markdown, or OOXML file
plaesy config                  # Read and validate scripts/configs/platform.json
plaesy platforms               # AI platforms plaesy can target
plaesy features                # List every feature in .plaesy/specs/
plaesy features create "<desc>"  # Create feature branch, directories, and spec.md
plaesy features paths          # Print current feature paths without creating anything
plaesy features validate       # Verify the current feature has a plan.md
plaesy tasks                   # Task lifecycle under .plaesy/tasks/
plaesy context update          # Sync an existing agent context file with plan.md
plaesy stack detect            # List instruction files for a target tech stack
plaesy trim                    # Token and context compression
plaesy images create           # Generate an image asset via a configured provider API
plaesy install                 # Install the plaesy binary to a well-known bin directory
plaesy uninstall               # Remove the installed plaesy binary
plaesy status                  # Check installation status and system information
plaesy clean                   # Clean current directory
plaesy clean <directory>       # Clean specified directory (default: current)
plaesy upgrade                 # Stub — prints "not yet implemented"
plaesy repair                  # Stub — prints "not yet implemented"
```

```powershell

# Windows PowerShell CLI
plaesy init              # Scaffold a new Plaesy project
plaesy init <directory>  # Interactive AI selection in specified directory
plaesy init <directory> -AI <platform>    # Use specific AI platform in directory
plaesy reload               # Refresh generated .plaesy/ files, keeping memory and context
plaesy analyze           # Analyze current project structure and generate documentation
plaesy graph             # Build or query a knowledge graph of the project
plaesy search            # Embedding-based semantic search over extracted symbols
plaesy index              # Build or update the search index (top-level alias for `search index`)
plaesy query              # Search the local index directly (top-level alias for `search query`)
plaesy eval               # Run the search evaluation harness against a golden set
plaesy doctor             # Check health of reach platforms, search index, and config
plaesy reach              # Access external data sources (web, github, youtube, etc.)
plaesy stats              # Show search index and reach cache statistics
plaesy validate          # Run the project-wide checks
plaesy config            # Read and validate scripts/configs/platform.json
plaesy platforms         # AI platforms plaesy can target
plaesy features          # List every feature in .plaesy/specs/
plaesy features create "<desc>"  # Create feature branch, directories, and spec.md
plaesy features paths    # Print current feature paths without creating anything
plaesy features validate # Verify the current feature has a plan.md
plaesy tasks             # Task lifecycle under .plaesy/tasks/
plaesy context update    # Sync an existing agent context file with plan.md
plaesy stack detect      # List instruction files for a target tech stack
plaesy trim              # Token and context compression
plaesy images create     # Generate an image asset via a configured provider API
plaesy install           # Install the plaesy binary to a well-known bin directory
plaesy uninstall         # Remove the installed plaesy binary
plaesy status            # Check installation status and system information
plaesy clean              # Clean current directory
plaesy clean <directory> # Clean specified directory (default: current)
plaesy upgrade           # Stub — prints "not yet implemented"
plaesy repair            # Stub — prints "not yet implemented"
```

### Popular AI Platforms

- `claude` - Claude Code by Anthropic
- `cursor_ai` - Cursor AI Assistant
- `github_copilot` - GitHub Copilot
- `windsurf_ai` - Windsurf AI

And 6+ other platforms available - see full list in documentation

### AI Assistant Workflow Commands

```markdown

/start <description>     # Begin new development workflow (generates project constitution first)
/assess [role] [path]    # Unified project assessment and deep research
/continue                # Resume and complete remaining phases
/implement               # Begin implementation phase with TDD (independent verifier pass before done)
/optimize                # Performance optimization and code refactoring
/improve                 # Best-practice gate for every active dimension -> routes defects/known gaps to /assess, /optimize, /fix
/loop                    # Autonomous assess -> fix -> verify -> repeat, no user input
/fix                     # Bug resolution and error recovery
/doc                     # Generate comprehensive project documentation
/save                    # Save current context and new knowledge
/create:images          # Generate image assets from text prompt or design spec (router for /create:{scope})
```

## Documentation

Each top-level folder (`scripts/`, `prompts/`, `templates/`, `instructions/`, `agents/`, `checklists/`) holds the actual framework content consumed by the CLI and AI assistants. The matching folder
under `docs/` holds the human-readable guide for that content — start at [docs/README.md](docs/README.md) for the full hub.

| Component | Content | Documentation |
|-----------|---------|----------------|
| **Scripts** | [scripts/](scripts/) — bash + PowerShell automation | [docs/scripts/](docs/scripts/README.md) |
| **Prompts** | [prompts/](prompts/) — AI-optimized workflow prompts | [docs/prompts/](docs/prompts/README.md) |
| **Instructions** | [instructions/](instructions/) — technology-specific guides | [docs/instructions/](docs/instructions/README.md) |
| **Templates** | [templates/](templates/) — project/document templates | [docs/templates/](docs/templates/README.md) |
| **Agents** | [agents/](agents/) — AI role configurations | [docs/agents/](docs/agents/README.md) |
| **Checklists** | [checklists/](checklists/) — quality gate checklists | [docs/checklists/](docs/checklists/README.md) |
| **Testing** | [docs/testing/](docs/testing/README.md) — test strategy assets |  [docs/testing/](docs/testing/README.md) |

---

## Version Information

**Current Version:** 0.0.1

### Getting Version Information

```bash

# Check current version
cat VERSION

# Or use Plaesy CLI (when available)
plaesy --version
```

---
