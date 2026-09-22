<div align="center" style="background-color: #f8f9fa; padding: 20px; border-radius: 10px; margin: 20px 0;">
  <img src="https://raw.githubusercontent.com/plaesy/.github/refs/heads/main/assets/img/Logo.svg" alt="Plaesy Constitution Kit - AI Development Framework" />
</div>

# Plaesy: Constitution Kit

**Version: 0.0.1**

## What is Plaesy Constitution Kit?

Plaesy Constitution Kit is an AI prompt framework built around a project **constitution** — a single governing source of truth (`.plaesy/memory/constitution.md`) that every phase (`/assess`, `/implement`, `/optimize`, `/fix`, `/improve`) reads before acting. Unlike spec-driven tools scoped to code alone, Plaesy drives all eight dimensions of a project from that constitution — technical, design, business, legal, marketing, financial, management, and product — through workflow automation from idea to production-ready deliverable, with anti-hallucination protocols enforcing evidence-backed output at each phase.

## Why Plaesy Constitution Kit (vs. other spec-driven tools)

Most spec-driven development tools — including GitHub's own `spec-kit` — scope the "spec" to code: a technical plan for an AI coding agent to implement. Plaesy Constitution Kit treats the **constitution**, not the spec, as the source of truth, and applies it beyond code: an `/assess:business` or `/assess:legal` run is governed by the same constitution as `/implement:technical`. If your project needs more than code specified — pricing model, compliance posture, go-to-market messaging — those live in the same governed workflow, not a separate tool.

## Quick Start

### Install & Initialize

Plaesy Constitution Kit ships as a single cross-platform Go binary (`plaesy`). There is no published release yet — build it from source (requires [Go](https://go.dev/dl/) 1.22+):

```bash
git clone https://github.com/plaesy/spec-kit.git
cd spec-kit/scripts
go build -o plaesy ./cmd/plaesy

# put the binary on PATH, then:
./plaesy install          # copies itself to a well-known bin dir and prints PATH instructions

# Initialize new project (interactive AI selection)
plaesy init my-awesome-app

# Or initialize in current directory with specific AI
plaesy init . --ai claude_code
```

The same binary and commands work identically on Linux, macOS, and Windows — no separate bash/PowerShell install path is needed anymore.

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

### Essential Commands
```bash
# Linux/macOS CLI
plaesy init                    # Interactive AI selection in current directory
plaesy init <directory>        # Interactive AI selection in specified directory
plaesy init <directory> --ai <platform>    # Use specific AI platform in directory
plaesy analyze                 # Analyze current project structure and generate documentation
plaesy clean                   # Clean current directory
plaesy upgrade                 # Upgrade framework
plaesy status                  # Check installation status and system information
plaesy repair                  # Fix missing components and scripts
plaesy uninstall               # Remove Plaesy Constitution Kit completely
```

```powershell
# Windows PowerShell CLI
plaesy init              # Interactive AI selection in current directory
plaesy init <directory>  # Interactive AI selection in specified directory
plaesy init <directory> -AI <platform>    # Use specific AI platform in directory
plaesy analyze           # Analyze current project structure and generate documentation
plaesy clean <directory> # Clean specified directory (default: current)
plaesy upgrade           # Upgrade framework
plaesy status            # Check installation status and system information
plaesy repair            # Fix missing components and scripts
plaesy uninstall         # Remove Plaesy Constitution Kit completely
```

### Popular AI Platforms
- `claude_code` - Claude Code by Anthropic
- `cursor_ai` - Cursor AI Assistant
- `github_copilot` - GitHub Copilot
- `windsurf_ai` - Windsurf AI

*And 6+ other platforms available - see full list in documentation*

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
```

## Documentation

Each top-level folder (`scripts/`, `prompts/`, `templates/`, `instructions/`, `chatmodes/`, `checklists/`, `testing/`) holds the actual framework content consumed by the CLI and AI assistants. The matching folder under `docs/` holds the human-readable guide for that content — start at [docs/README.md](docs/README.md) for the full hub.

| Component | Content | Documentation |
|-----------|---------|----------------|
| **Scripts** | [scripts/](scripts/) — bash + PowerShell automation | [docs/scripts/](docs/scripts/README.md) |
| **Prompts** | [prompts/](prompts/) — AI-optimized workflow prompts | [docs/prompts/](docs/prompts/README.md) |
| **Instructions** | [instructions/](instructions/) — technology-specific guides | [docs/instructions/](docs/instructions/README.md) |
| **Templates** | [templates/](templates/) — project/document templates | [docs/templates/](docs/templates/README.md) |
| **Chat Modes** | [chatmodes/](chatmodes/) — AI role configurations | [docs/chatmodes/](docs/chatmodes/README.md) |
| **Checklists** | [checklists/](checklists/) — quality gate checklists | [docs/checklists/](docs/checklists/README.md) |
| **Testing** | [testing/](testing/) — test strategy assets | [docs/testing/](docs/testing/README.md) |

---

## Version Information

**Current Version: 0.0.1**

### Getting Version Information
```bash
# Check current version
cat VERSION

# Or use Plaesy CLI (when available)
plaesy --version
```

---