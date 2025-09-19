# Plaesy Spec-Kit Constitutional Framework

## Overview
This repository uses the **Plaesy Spec-Kit Constitutional Framework** - a chat-based methodology optimized for IDE native chat interfaces. Transform ideas into structured solutions through AI-guided workflows directly in your IDE chat. Use simple commands like `/idea`, `/specify`, `/plan`, `/tasks` to create high-quality deliverables with constitutional compliance.

**✨ Optimized for**: GitHub Copilot Chat, Qoder IDE, Cursor, Claude Code, and other IDE chat interfaces

## Environment Requirements
- **Operating System**: Windows, macOS, Linux
- **Shell**: Bash (Linux/macOS) or PowerShell (Windows)
- **Permissions**: Read/write access to project directory
- **Network**: Internet access for initial setup (optional)

## Setup & Installation

**Linux/macOS:**
```bash
/bin/bash <(curl -s https://raw.githubusercontent.com/plaesy/spec-kit/main/scripts/bash/install.sh)
```

**Windows (PowerShell):**
```powershell
iwr -useb https://raw.githubusercontent.com/plaesy/spec-kit/main/scripts/powershell/install.ps1 | iex
```

```bash
# Initialize constitutional framework in your project
plaesy init .
```

## Dependencies
- **Required**: None - framework is completely self-contained
- **Optional**: Git (for version control integration)
- **Usage**: Just start chatting with `/idea`, `/specify`, `/plan`, `/tasks` commands

## Methodology & Constitutional Requirements

### Core Principles (NON-NEGOTIABLE)
- **Structured Thinking**: Follow constitutional reasoning patterns and frameworks
- **Quality-First**: Excellence and thoroughness in analysis and execution
- **Evidence-Based**: Decisions based on validated data and constitutional compliance
- **Documentation Standards**: Comprehensive documentation with clear reasoning
- **Iterative Refinement**: Continuous improvement through constitutional validation

### Chat-Based Workflow
1. **Start with an idea**: Chat `/idea I want to create an app like Google Photos`
2. **Create specification**: Chat `/specify` to transform idea into detailed requirements
3. **Plan implementation**: Chat `/plan` to create structured approach and architecture
4. **Break down tasks**: Chat `/tasks` to get actionable implementation steps
5. **Use expert roles**: Chat with `@dev`, `@qa`, `@architect` for specialized guidance

## AI Assistant Roles

### Available Expert Personas
- `dev.chatmode.md` - Technical Implementation Expert with constitutional compliance
- `qa.chatmode.md` - Quality Assurance Specialist for validation and testing
- `sa.chatmode.md` - Solution Architect for system design and planning
- `security.chatmode.md` - Security Specialist for risk assessment and compliance
- `devops.chatmode.md` - Operations Expert for process and infrastructure
- `ba.chatmode.md` - Business Analyst for requirements and stakeholder management
- `pm.chatmode.md` - Product Manager for strategy and prioritization
- `po.chatmode.md` - Product Owner for value delivery and backlog management
- `bo.chatmode.md` - Business Owner for strategic oversight and governance
- `designer.chatmode.md` - UX/UI Designer for user experience and interface design
- `sm.chatmode.md` - Scrum Master for process facilitation and team coaching

### Chat Commands
- `/idea` - Transform raw ideas into structured concepts (e.g., `/idea create app like Google Photos`)
- `/specify` - Convert ideas into detailed specifications and requirements
- `/plan` - Create implementation roadmap and technical approach
- `/tasks` - Break down plans into actionable implementation steps
- `/remember` - Save learnings and decisions to framework knowledge base

## File Structure
```
.plaesy/                            # Framework core directory
├── memory/
│   ├── constitution.md             # Non-negotiable framework rules (LOAD FIRST)
│   ├── knowledge-base.md           # Domain knowledge repository
│   └── test-priorities-matrix.md   # Quality assessment matrix
├── instructions/
│   ├── plaesy.instructions.md      # Main constitutional instructions
│   ├── context.instructions.md     # Dynamic context awareness
│   └── *.instructions.md           # Technology-specific instructions
├── chatmodes/                      # Role-based expert personas (14 available)
│   ├── dev.chatmode.md             # Technical Implementation Expert
│   ├── qa.chatmode.md              # Quality Assurance Specialist
│   ├── sa.chatmode.md              # Solution Architect
│   ├── ba.chatmode.md              # Business Analyst
│   └── *.chatmode.md               # Other expert roles
├── prompts/                        # Workflow methodology prompts
│   ├── idea.prompt.md              # Idea exploration workflow
│   ├── specify.prompt.md           # Specification creation
│   ├── plan.prompt.md              # Implementation planning
│   └── tasks.prompt.md             # Task breakdown workflow
├── templates/                      # Document templates (40+ available)
│   ├── spec.template.md            # Feature specification template
│   ├── plan.template.md            # Implementation plan template
│   └── *.template.md               # Various document templates
├── checklists/                     # Quality assurance checklists
└── scripts/                        # Automation and validation scripts
    ├── bash/                       # Linux/macOS scripts
    └── powershell/                 # Windows scripts

AGENTS.md                          # This file - AI assistant configuration
CLAUDE.md                          # Claude Code specific instructions (auto-generated)
.cursor/rules/                     # Cursor AI rules (auto-generated)
.github/copilot-instructions.md    # GitHub Copilot instructions (auto-generated)
```

## Documentation Guidelines
```markdown
# Use constitutional documentation format

## Change Summary
- Describe what was accomplished with constitutional compliance
- Include reasoning and decision rationale
- Document validation and quality checks performed
- Reference constitutional principles applied

## Quality Validation
- [ ] Constitutional compliance verified
- [ ] Quality standards met
- [ ] Documentation completeness confirmed
- [ ] Stakeholder requirements satisfied

🤖 Generated with Constitutional Framework
Co-Authored-By: AI Assistant <ai@plaesy.dev>
```

## Security & Compliance
- **Data Handling**: Framework processes no sensitive data - all operations are local
- **Access Control**: Standard file system permissions apply
- **Privacy**: No data transmission to external services (self-contained)
- **Compliance**: Constitutional framework ensures quality and governance standards
- **Audit Trail**: All framework operations are logged and traceable

## Performance Considerations
- **Resource Usage**: Minimal system resources required
- **Scalability**: Framework scales with project complexity
- **Optimization**: Efficient file operations and minimal memory footprint
- **Response Time**: Instant access to templates and guidelines
- **Storage**: Approximately 50MB for complete framework

## Quality Gates
All work must pass:
- [ ] Constitutional compliance validation
- [ ] Quality standards implementation
- [ ] Thorough analysis and reasoning
- [ ] Stakeholder requirements validation
- [ ] Documentation completeness and clarity
- [ ] Evidence-based decision making
- [ ] Process adherence and methodology

## Standards & Governance
- Follow constitutional quality standards
- Implement structured thinking patterns
- Use comprehensive documentation practices
- Apply systematic versioning and change management
- Ensure stakeholder alignment and transparency

## Troubleshooting

### Common Issues
**Issue**: Framework not initializing properly
```bash
# Solution: Check permissions and run validation
chmod +x .plaesy/scripts/bash/*.sh
```

**Issue**: Templates not loading
```bash
# Solution: Verify file structure and permissions
ls -la .plaesy/templates/
find .plaesy -name "*.template.md" | wc -l  # Should show 40+ templates
```

**Issue**: AI assistant not following constitutional guidelines
```bash
# Solution: Ensure constitutional context is loaded first
cat .plaesy/memory/constitution.md
cat .plaesy/instructions/plaesy.instructions.md
```

### FAQ
**Q**: Can I customize the constitutional framework?
**A**: Yes, but maintain core principles. Modify templates and add domain-specific instructions.

**Q**: Does this work with all AI assistants?
**A**: Yes, framework generates platform-specific configurations automatically.

**Q**: How do I add new expert roles?
**A**: Create new `.chatmode.md` files following the established format and constitutional principles.

## Important Notes
- **Constitutional Priority**: Always load `.plaesy/memory/constitution.md` first
- **Evidence-Based**: All decisions must be backed by validated data and reasoning
- **Quality Focus**: Thoroughness and excellence over speed - constitutional compliance is mandatory
- **AI Orchestration**: Leverage AI assistants with constitutional guidance across any domain
- **Documentation**: Comprehensive documentation with examples, reasoning, and validation evidence
- **Universal Application**: Framework applies to any project type - business, technical, creative, or analytical
- **WSL Integration**: For WSL users, see [WSL-IDE-INTEGRATION.md](WSL-IDE-INTEGRATION.md) for optimized setup
- **IDE Chat Integration**: See [IDE-CHAT-INTEGRATION.md](IDE-CHAT-INTEGRATION.md) for comprehensive IDE chat setup

---
*This configuration follows the Plaesy Constitutional Framework standards for structured thinking and quality execution across any domain.*