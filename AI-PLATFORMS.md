# AI Platform Integration Guide

## Overview
Spec-Kit supports 12+ AI platforms with dynamically injected headers that optimize each platform for constitutional framework compliance. Each platform receives customized instructions while maintaining consistent quality standards.

## Supported AI Platforms

### 🤖 **Native AI Assistants**

#### 1. GitHub Copilot
- **Configuration**: Comment-based instructions in prompt files
- **Strengths**: Deep VS Code integration, context-aware suggestions
- **Header Format**: Instruction comments with constitutional requirements
- **Usage**: `plaesy init --ai copilot`

```markdown
---
# GitHub Copilot Instructions for Constitutional Development Framework
# Constitutional Framework Compliance
You are working within the Plaesy Spec-Kit constitutional framework...
---
```

#### 2. Cursor AI
- **Configuration**: YAML frontmatter with tool specifications
- **Strengths**: Advanced AI-assisted coding, multi-file context
- **Header Format**: Structured YAML with model and tool configurations
- **Usage**: `plaesy init --ai cursor`

```markdown
---
description: "Constitutional Development Framework agent with TDD enforcement"
model: "Claude-4"
tools: ["search", "editFiles", "runCommands", "problems", "usages"]
context: "constitutional-framework"
---
```

#### 3. Windsurf AI
- **Configuration**: Role-based YAML frontmatter
- **Strengths**: Systematic development approach, quality gates
- **Header Format**: Role definition with expertise and approach
- **Usage**: `plaesy init --ai windsurf`

```markdown
---
role: "Constitutional Development Expert"
expertise: ["TDD", "System Architecture", "API Design", "Security", "Performance"]
approach: "systematic"
quality_gates: "enforced"
compliance: "constitutional-framework"
---
```

### ☁️ **Cloud AI Platforms**

#### 4. Claude Code
- **Configuration**: System prompt with temperature and tool settings
- **Strengths**: Advanced reasoning, comprehensive analysis
- **Header Format**: System prompt with temperature and tool configurations
- **Usage**: `plaesy init --ai claude`

```markdown
---
system: |
  You are a Constitutional Development Framework specialist...
temperature: 0.1
max_tokens: 4000
tools: ["computer", "bash", "str_replace_editor"]
---
```

#### 5. ChatGPT
- **Configuration**: Context instructions and guidelines
- **Strengths**: Conversational interface, broad knowledge base
- **Header Format**: Context instructions with constitutional principles
- **Usage**: `plaesy init --ai chatgpt`

#### 6. Google Gemini
- **Configuration**: Role and context definition
- **Strengths**: Multimodal capabilities, comprehensive analysis
- **Header Format**: Detailed role definition with implementation guidelines
- **Usage**: `plaesy init --ai gemini`

### 🔧 **Specialized AI Tools**

#### 7. Trae.ai
- **Configuration**: Multi-agent coordination setup
- **Strengths**: Multi-agent workflows, specialized team roles
- **Header Format**: Agent coordination with constitutional enforcement
- **Usage**: `plaesy init --ai trae-ai`

```markdown
---
agent_type: "constitutional_development_team"
framework: "plaesy-spec-kit"
coordination: "multi-agent"
quality_enforcement: "constitutional"
---
```

#### 8. Qwen Code
- **Configuration**: Bilingual instructions (Chinese/English)
- **Strengths**: Chinese language support, code generation
- **Header Format**: Bilingual constitutional framework instructions
- **Usage**: `plaesy init --ai qwen-code`

#### 9. Codex CLI
- **Configuration**: Project memory integration
- **Strengths**: CLI integration, project memory system
- **Header Format**: Project memory and constitutional context
- **Usage**: `plaesy init --ai codex-cli`

#### 10. OpenCode CLI
- **Configuration**: JSON configuration format
- **Strengths**: Structured configuration, CLI automation
- **Header Format**: JSON-based constitutional configuration
- **Usage**: `plaesy init --ai opencode-cli`

### 🏠 **Local AI Solutions**

#### 11. Local AI (Ollama/LM Studio)
- **Configuration**: System prompt for local models
- **Strengths**: Privacy, customization, offline usage
- **Header Format**: Comprehensive system prompt for local models
- **Usage**: `plaesy init --ai local-ai`

#### 12. Manual Development
- **Configuration**: Human-readable guidelines
- **Strengths**: No AI dependency, full control
- **Header Format**: Detailed manual implementation guide
- **Usage**: `plaesy init --ai manual`

## Platform Comparison Matrix

| Platform | Integration | Strengths | Constitutional Support | Best For |
|----------|-------------|-----------|----------------------|----------|
| **GitHub Copilot** | VS Code Native | Context-aware, Fast | ✅ Full | VS Code users |
| **Cursor AI** | Editor Native | Multi-file, Advanced | ✅ Full | Professional developers |
| **Windsurf AI** | Editor Integration | Systematic, Quality gates | ✅ Full | Quality-focused teams |
| **Claude Code** | Web/API | Advanced reasoning | ✅ Full | Complex problem solving |
| **ChatGPT** | Web/API | Conversational | ✅ Full | Iterative development |
| **Gemini** | Web/API | Multimodal | ✅ Full | Comprehensive analysis |
| **Trae.ai** | Platform | Multi-agent | ✅ Full | Team coordination |
| **Qwen Code** | Various | Bilingual | ✅ Full | Chinese developers |
| **Codex CLI** | CLI | Project memory | ✅ Full | CLI-focused workflows |
| **OpenCode CLI** | CLI | Structured config | ✅ Full | Automated workflows |
| **Local AI** | Local | Privacy, Offline | ✅ Full | Privacy-conscious teams |
| **Manual** | Human | Full control | ✅ Full | No AI preference |

## Header Injection System

### Automatic Header Injection
When you run `plaesy init --ai <platform>`, the system automatically:

1. **Detects Platform**: Identifies the chosen AI platform
2. **Loads Header**: Retrieves platform-specific header template
3. **Injects Headers**: Prepends headers to all prompt files
4. **Configures Project**: Sets up platform-specific configurations

### Manual Header Management

#### Inject Headers
```bash
# Linux/macOS
./scripts/bash/inject-ai-headers.sh --ai cursor --target ./prompts

# Windows
.\scripts\powershell\inject-ai-headers.ps1 -AiPlatform cursor -TargetDirectory .\prompts
```

#### Switch Platforms
```bash
# Remove old headers and inject new ones
./scripts/bash/inject-ai-headers.sh --ai claude --target ./prompts --force

# With backup
./scripts/bash/inject-ai-headers.sh --ai claude --target ./prompts --force --backup
```

#### Dry Run (Preview Changes)
```bash
# See what would change without making changes
./scripts/bash/inject-ai-headers.sh --ai windsurf --target ./prompts --dry-run
```

## Constitutional Framework Enforcement

### Common Across All Platforms
Every AI platform header includes these constitutional requirements:

#### 1. Test-Driven Development (TDD)
- **Red-Green-Refactor cycle** mandatory
- **90%+ test coverage** requirement
- **Real dependencies** in integration tests (no mocks)
- **Test hierarchy** validation

#### 2. Interface-First Design
- **OpenAPI specifications** before implementation
- **Comprehensive error handling** with meaningful messages
- **Input validation** with clear feedback
- **Semantic versioning** with backward compatibility

#### 3. Built-in Observability
- **Structured logging** with correlation IDs
- **Health check endpoints** for monitoring
- **Metrics collection** for performance and business KPIs
- **Alerting configuration** for proactive monitoring

#### 4. Security by Design
- **OWASP compliance** from design phase
- **Authentication and authorization** implementation
- **Data protection** with encryption at rest and in transit
- **Security testing** integration in CI/CD

#### 5. Platform Agnostic Architecture
- **Containerization** with Docker
- **Infrastructure as Code** for reproducible environments
- **Environment parity** between dev and production
- **Scalability** for horizontal and vertical growth

## Platform-Specific Optimizations

### GitHub Copilot Optimizations
- **Context-aware suggestions** aligned with constitutional principles
- **VS Code integration** for seamless development
- **Inline suggestions** following TDD patterns
- **Comment-based guidance** for constitutional compliance

### Cursor AI Optimizations
- **Multi-file context** for architectural consistency
- **Advanced refactoring** following constitutional patterns
- **Tool integration** for testing and validation
- **Real-time constitutional compliance checking**

### Claude Code Optimizations
- **Advanced reasoning** for complex architectural decisions
- **Comprehensive analysis** of constitutional compliance
- **Detailed explanations** of implementation choices
- **System-level thinking** for scalable solutions

### Trae.ai Multi-Agent Optimizations
- **Architect agent** for constitutional design validation
- **Developer agent** for TDD implementation
- **Tester agent** for quality assurance validation
- **Security agent** for OWASP compliance review

## Usage Examples

### Basic Project Setup
```bash
# Initialize with Cursor AI
plaesy init my-ecommerce-app --ai cursor

# Initialize with Claude Code
plaesy init my-api-service --ai claude

# Initialize with GitHub Copilot
plaesy init my-frontend-app --ai copilot
```

### Advanced Configuration
```bash
# Initialize with backup and force overwrite
plaesy init existing-project --ai windsurf --force --backup

# Initialize with specific directory
plaesy init . --ai trae-ai --target ./my-prompts

# Initialize with dry run to preview
plaesy init test-project --ai gemini --dry-run
```

### Platform Migration
```bash
# Switch from ChatGPT to Claude
./scripts/bash/inject-ai-headers.sh --ai claude --target . --force --backup

# Migrate to Cursor AI with validation
./scripts/bash/inject-ai-headers.sh --ai cursor --target ./prompts --dry-run
./scripts/bash/inject-ai-headers.sh --ai cursor --target ./prompts --force
```

## Troubleshooting

### Common Issues

#### Header Not Applied
**Issue**: AI platform not following constitutional principles
**Solution**:
```bash
# Re-inject headers
./scripts/bash/inject-ai-headers.sh --ai <platform> --target . --force

# Verify header content
head -50 prompts/idea.prompt.md
```

#### Platform Not Recognized
**Issue**: AI platform shows as unsupported
**Solution**:
```bash
# Check available platforms
./scripts/bash/inject-ai-headers.sh --help

# Use manual fallback
plaesy init --ai manual
```

#### Mixed Headers
**Issue**: Multiple AI platform headers in same file
**Solution**:
```bash
# Clean and re-inject
./scripts/bash/inject-ai-headers.sh --ai <correct-platform> --target . --force
```

### Validation Commands
```bash
# Check header injection status
cat .plaesy-headers.json

# Validate constitutional compliance
grep -r "Constitutional Framework" prompts/

# Test with specific platform
head -20 prompts/specify.prompt.md
```

## Best Practices

### Platform Selection
1. **Team Preference**: Choose platform your team already uses
2. **Integration Needs**: Consider IDE/editor integration requirements
3. **Quality Requirements**: All platforms support constitutional framework equally
4. **Privacy Concerns**: Use local AI for sensitive projects
5. **Multi-agent Needs**: Use Trae.ai for complex team coordination

### Header Management
1. **Backup Before Changes**: Use `--backup` flag when switching platforms
2. **Test with Dry Run**: Use `--dry-run` to preview changes
3. **Version Control**: Commit header changes with clear messages
4. **Team Coordination**: Ensure entire team uses same AI platform
5. **Regular Updates**: Update headers when framework versions change

### Quality Assurance
1. **Validate Headers**: Check that constitutional principles are included
2. **Test AI Responses**: Verify AI follows constitutional requirements
3. **Monitor Compliance**: Ensure TDD and real dependency testing
4. **Review Regularly**: Periodic review of AI-generated code quality
5. **Feedback Loop**: Improve headers based on AI performance

## Contributing New AI Platforms

### Adding New Platform Support
1. Create header template: `templates/ai-headers/new-platform.header.md`
2. Include all constitutional requirements
3. Add platform to validation scripts
4. Test with sample prompts
5. Update documentation
6. Submit pull request

### Header Template Requirements
- [ ] Constitutional framework context
- [ ] TDD enforcement requirements
- [ ] Real dependency testing mandate
- [ ] Security and OWASP compliance
- [ ] Observability requirements
- [ ] Platform-specific syntax
- [ ] Quality gates definition
- [ ] Error handling guidelines

---

**Framework Version**: 3.0.0
**Supported Platforms**: 12+
**Constitutional Compliance**: 100%
**Last Updated**: 2024-01-15