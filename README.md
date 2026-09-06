<div align="center" style="background-color: #f8f9fa; padding: 20px; border-radius: 10px; margin: 20px 0;">
  <img src="https://raw.githubusercontent.com/plaesy/.github/refs/heads/main/assets/img/Logo.svg" alt="Plaesy Logo - AI Development Framework" />
</div>

# Spec-Kit: Constitutional Development Framework

**Version: 0.0.1**

## What is Plaesy Spec-Kit?

Plaesy Spec-Kit is a **state-of-the-art AI prompt framework** featuring **true workflow automation** from idea to production-ready code. It provides zero-ambiguity prompts with anti-hallucination protocols that enforce disciplined development through mandatory full automation—ensuring quality, security, and maintainability without manual intervention.

## Quick Start

### Install & Initialize

**Linux/macOS:**
```bash
# Install Plaesy Spec-Kit
curl -s https://raw.githubusercontent.com/plaesy/spec-kit/main/scripts/bash/install.sh | bash

# Initialize new project (interactive AI selection)
plaesy init my-awesome-app

# Or initialize in current directory with specific AI
plaesy init . --ai claude_code
```

**Windows (PowerShell):**
```powershell
# Install Plaesy Spec-Kit
iwr -useb https://raw.githubusercontent.com/plaesy/spec-kit/main/scripts/powershell/install.ps1 | iex

# Initialize new project (interactive AI selection)
plaesy init my-awesome-app

# Or initialize in current directory with specific AI
plaesy init . -AI claude_code
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
plaesy uninstall               # Remove Plaesy Spec-Kit completely
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
plaesy uninstall         # Remove Plaesy Spec-Kit completely
```

### Popular AI Platforms
- `claude_code` - Claude Code by Anthropic
- `cursor_ai` - Cursor AI Assistant
- `github_copilot` - GitHub Copilot
- `windsurf_ai` - Windsurf AI

*And 6+ other platforms available - see full list in documentation*

### AI Assistant Workflow Commands
```markdown
/start <description>     # Begin new development workflow
/assess [role] [path]    # Unified project assessment
/continue                # Resume and complete remaining phases
/research                # Deep research to resolve technical uncertainties
/clarify                 # Requirements clarification and ambiguity resolution
/design                  # UI/UX design generation with component systems
/flow                    # Architecture and workflow diagram generation
/implement               # Begin implementation phase with TDD
/optimize                # Performance optimization and code refactoring
/fix                     # Bug resolution and error recovery
/doc                     # Generate comprehensive project documentation
/evolve                  # Autonomous loop: idea → implementation → re-evaluation
/save                    # Save current context and new knowledge
```

## Documentation

Each top-level folder (`scripts/`, `prompts/`, `templates/`, `instructions/`, `chatmodes/`, `checklists/`, `testing/`) holds the actual framework content consumed by the CLI and AI assistants. The matching folder under `docs/` holds the human-readable guide for that content — start at [docs/README.md](docs/README.md) for the full hub.

| Component | Content | Documentation |
|-----------|---------|----------------|
| **Scripts** | [scripts/](scripts/) — bash + PowerShell automation | [docs/scripts/](docs/scripts/) |
| **Prompts** | [prompts/](prompts/) — AI-optimized workflow prompts | [docs/prompts/](docs/prompts/) |
| **Instructions** | [instructions/](instructions/) — technology-specific guides | [docs/instructions/](docs/instructions/) |
| **Templates** | [templates/](templates/) — project/document templates | [docs/templates/](docs/templates/) |
| **Chat Modes** | [chatmodes/](chatmodes/) — AI role configurations | [docs/chatmodes/](docs/chatmodes/) |
| **Checklists** | [checklists/](checklists/) — quality gate checklists | [docs/checklists/](docs/checklists/) |
| **Testing** | [testing/](testing/) — test strategy assets | [docs/testing/](docs/testing/) |

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