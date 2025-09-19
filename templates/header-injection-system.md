# AI-Specific Header Injection System

## Overview
This document describes how the Spec-Kit framework dynamically injects AI-specific headers into prompt files based on the chosen AI platform during initialization.

## Header Injection Process

### 1. Platform Detection
During `plaesy init --ai <platform>`, the system:
- Identifies the target AI platform
- Loads the corresponding header template
- Prepends the header to all prompt files

### 2. Available AI Platform Headers

| AI Platform | Header File | Configuration Type |
|-------------|-------------|-------------------|
| GitHub Copilot | `copilot.header.md` | Instruction comments |
| Cursor AI | `cursor.header.md` | YAML frontmatter + instructions |
| Windsurf AI | `windsurf.header.md` | YAML frontmatter + role definition |
| Claude Code | `claude.header.md` | System prompt + temperature settings |
| ChatGPT | `chatgpt.header.md` | Context instructions |
| Gemini | `gemini.header.md` | Role and context definition |
| Trae.ai | `trae-ai.header.md` | Multi-agent coordination |
| Qwen Code | `qwen-code.header.md` | Bilingual instructions (Chinese/English) |
| Codex CLI | `codex-cli.header.md` | Project memory integration |
| OpenCode CLI | `opencode-cli.header.md` | JSON configuration format |
| Local AI | `local-ai.header.md` | System prompt for local models |
| Manual | `manual.header.md` | Human-readable guidelines |

### 3. Header Injection Algorithm

```bash
# Pseudocode for header injection
function inject_headers(ai_platform, target_directory) {
    header_file = "templates/ai-headers/${ai_platform}.header.md"

    if (!file_exists(header_file)) {
        log_warning("No specific header for ${ai_platform}, using default")
        header_file = "templates/ai-headers/manual.header.md"
    }

    header_content = read_file(header_file)

    # Find all prompt files
    prompt_files = find_files(target_directory, "*.prompt.md")

    for prompt_file in prompt_files {
        original_content = read_file(prompt_file)

        # Inject header at the beginning
        new_content = header_content + "\n\n" + original_content

        write_file(prompt_file, new_content)

        log_info("Injected ${ai_platform} header into ${prompt_file}")
    }
}
```

### 4. Platform-Specific Configurations

#### GitHub Copilot
```markdown
---
# GitHub Copilot Instructions for Constitutional Development Framework
# Constitutional Framework Compliance
You are working within the Plaesy Spec-Kit constitutional framework...
---
```

#### Cursor AI
```markdown
---
description: "Constitutional Development Framework agent with TDD enforcement"
model: "Claude-4"
tools: ["search", "editFiles", "runCommands"]
context: "constitutional-framework"
---
```

#### Windsurf AI
```markdown
---
role: "Constitutional Development Expert"
expertise: ["TDD", "System Architecture", "API Design", "Security"]
approach: "systematic"
quality_gates: "enforced"
---
```

#### Claude Code
```markdown
---
system: |
  You are a Constitutional Development Framework specialist...
temperature: 0.1
max_tokens: 4000
tools: ["computer", "bash", "str_replace_editor"]
---
```

### 5. Header Validation

#### Required Elements
All headers must include:
- Constitutional framework context
- TDD enforcement requirements
- Real dependency testing mandate
- Interface contract requirements
- Observability and security guidelines
- Platform-specific configuration syntax

#### Validation Checklist
- [ ] Constitutional principles clearly stated
- [ ] TDD requirements emphasized
- [ ] Real dependency testing mandate
- [ ] Security and OWASP compliance
- [ ] Observability requirements
- [ ] Platform-specific syntax correct
- [ ] Error handling guidelines
- [ ] Quality gates defined

### 6. Custom Header Creation

#### For New AI Platforms
1. Create new header file: `templates/ai-headers/{platform}.header.md`
2. Follow platform's specific configuration format
3. Include all constitutional requirements
4. Test with sample prompts
5. Add to platform mapping in scripts

#### Header Template Structure
```markdown
# Platform-Specific Configuration Section
[Platform-specific configuration syntax]

# Constitutional Framework Context
[Standard constitutional requirements]

# Response Guidelines
[Platform-specific response formatting]

# Quality Standards
[Constitutional compliance requirements]

# Context Integration
[Memory and architecture integration]
```

### 7. Implementation Scripts

#### Bash Implementation
```bash
#!/bin/bash
# scripts/bash/inject-ai-headers.sh

AI_PLATFORM="$1"
TARGET_DIR="$2"

HEADER_FILE="templates/ai-headers/${AI_PLATFORM}.header.md"

if [[ ! -f "$HEADER_FILE" ]]; then
    echo "Warning: No header for $AI_PLATFORM, using manual header"
    HEADER_FILE="templates/ai-headers/manual.header.md"
fi

HEADER_CONTENT=$(cat "$HEADER_FILE")

find "$TARGET_DIR" -name "*.prompt.md" | while read -r prompt_file; do
    echo "Injecting $AI_PLATFORM header into $prompt_file"

    # Create temporary file with header + original content
    {
        echo "$HEADER_CONTENT"
        echo ""
        echo ""
        cat "$prompt_file"
    } > "$prompt_file.tmp"

    # Replace original file
    mv "$prompt_file.tmp" "$prompt_file"
done

echo "Header injection complete for $AI_PLATFORM"
```

#### PowerShell Implementation
```powershell
# scripts/powershell/inject-ai-headers.ps1

param(
    [Parameter(Mandatory=$true)]
    [string]$AiPlatform,

    [Parameter(Mandatory=$true)]
    [string]$TargetDirectory
)

$HeaderFile = "templates/ai-headers/$AiPlatform.header.md"

if (-not (Test-Path $HeaderFile)) {
    Write-Warning "No header for $AiPlatform, using manual header"
    $HeaderFile = "templates/ai-headers/manual.header.md"
}

$HeaderContent = Get-Content $HeaderFile -Raw

Get-ChildItem -Path $TargetDirectory -Filter "*.prompt.md" -Recurse | ForEach-Object {
    Write-Host "Injecting $AiPlatform header into $($_.FullName)"

    $OriginalContent = Get-Content $_.FullName -Raw
    $NewContent = $HeaderContent + "`n`n" + $OriginalContent

    Set-Content -Path $_.FullName -Value $NewContent
}

Write-Host "Header injection complete for $AiPlatform"
```

### 8. Integration with plaesy init

#### Modified Initialization Process
```bash
# Enhanced plaesy init command
plaesy init my-project --ai cursor

# Process:
# 1. Create project structure
# 2. Copy base templates
# 3. Inject AI-specific headers
# 4. Configure platform-specific files
# 5. Validate configuration
```

#### Configuration File Update
```json
{
  "project": "my-project",
  "ai_platform": "cursor",
  "constitutional_framework": "3.0.0",
  "header_injected": true,
  "injection_timestamp": "2024-01-15T10:30:00Z",
  "prompt_files": [
    "prompts/idea.prompt.md",
    "prompts/specify.prompt.md",
    "prompts/plan.prompt.md",
    "prompts/tasks.prompt.md"
  ]
}
```

### 9. Header Management

#### Update Headers
```bash
# Update headers for existing project
plaesy update-headers --ai new-platform

# Re-inject headers after framework update
plaesy refresh-headers
```

#### Remove Headers
```bash
# Remove platform-specific headers
plaesy clean-headers

# Restore original prompt files
plaesy restore-prompts
```

### 10. Quality Assurance

#### Header Testing
- Test each header with sample prompts
- Validate platform-specific syntax
- Ensure constitutional compliance
- Test with different AI models

#### Continuous Validation
- Automated header validation in CI/CD
- Regular testing with AI platforms
- Community feedback integration
- Performance impact measurement

## Benefits

### For Users
- **Platform Optimization**: Headers optimized for each AI platform
- **Consistent Quality**: Constitutional compliance across all platforms
- **Easy Switching**: Simple platform migration with header updates
- **Better Results**: AI responses aligned with framework principles

### For Framework
- **Scalability**: Easy addition of new AI platforms
- **Maintainability**: Centralized header management
- **Flexibility**: Platform-specific customizations
- **Quality Control**: Consistent constitutional enforcement

---