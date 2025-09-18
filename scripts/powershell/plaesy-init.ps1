# Plaesy Project Initialization Script (PowerShell)
# Constitutional Development Framework

param(
  [string]$TargetDir = ".",
  [string]$AI = $null
)

$ErrorActionPreference = "Stop"

# Colors
function Write-Step { 
  param([string]$Message)
  Write-Host "▶ $Message" -ForegroundColor Blue
}

function Write-Success { 
  param([string]$Message)
  Write-Host "✅ $Message" -ForegroundColor Green
}

function Write-Warning { 
  param([string]$Message)
  Write-Host "⚠️ $Message" -ForegroundColor Yellow
}

function Write-Error-Custom { 
  param([string]$Message)
  Write-Host "❌ $Message" -ForegroundColor Red
  exit 1
}

$PleasyHome = "$env:USERPROFILE\.plaesy"

if (!(Test-Path $PleasyHome)) {
  Write-Error-Custom "Plaesy framework not found. Please run the installer first."
}

function Select-AIAssistant {
  if ($AI) {
    switch ($AI.ToLower()) {
      { $_ -in @("copilot", "github") } { return "copilot" }
      "cursor" { return "cursor" }
      "windsurf" { return "windsurf" }
      "claude" { return "claude" }
      { $_ -in @("trae", "trae-ai") } { return "trae-ai" }
      { $_ -in @("chatgpt", "openai") } { return "chatgpt" }
      { $_ -in @("gemini", "google") } { return "gemini" }
      { $_ -in @("qwen", "qwen-code") } { return "qwen-code" }
      { $_ -in @("kilo", "kilo-code") } { return "kilo-code" }
      { $_ -in @("codex", "codex-cli") } { return "codex-cli" }
      "codex-web" { return "codex-web" }
      { $_ -in @("opencode", "opencode-cli") } { return "opencode-cli" }
      { $_ -in @("local", "ollama") } { return "local" }
      { $_ -in @("none", "manual") } { return "none" }
      default {
        Write-Warning "Unknown AI assistant: $AI. Falling back to interactive selection."
      }
    }
  }
    
  Write-Host "🤖 Select your AI Assistant:" -ForegroundColor Magenta
  Write-Host "1. GitHub Copilot"
  Write-Host "2. Cursor AI"
  Write-Host "3. Windsurf AI"
  Write-Host "4. Claude Code (Anthropic)"
  Write-Host "5. Trae.ai"
  Write-Host "6. ChatGPT (OpenAI)"
  Write-Host "7. Gemini (Google)"
  Write-Host "8. Qwen Code"
  Write-Host "9. Kilo Code"
  Write-Host "10. Codex CLI"
  Write-Host "11. Codex Web"
  Write-Host "12. OpenCode CLI"
  Write-Host "13. Local AI (Ollama/LM Studio)"
  Write-Host "14. No AI (Manual development)"
  Write-Host ""
    
  do {
    $choice = Read-Host "Choose your AI assistant (1-14)"
    switch ($choice) {
      "1" { return "copilot" }
      "2" { return "cursor" }
      "3" { return "windsurf" }
      "4" { return "claude" }
      "5" { return "trae-ai" }
      "6" { return "chatgpt" }
      "7" { return "gemini" }
      "8" { return "qwen-code" }
      "9" { return "kilo-code" }
      "10" { return "codex-cli" }
      "11" { return "codex-web" }
      "12" { return "opencode-cli" }
      "13" { return "local" }
      "14" { return "none" }
      default { Write-Host "Please enter a number between 1-14" -ForegroundColor Red }
    }
  } while ($true)
}

function Copy-CoreFramework {
  Write-Step "Setting up constitutional framework..."
    
  # Create target directory if it doesn't exist
  if (!(Test-Path $TargetDir)) {
    New-Item -ItemType Directory -Path $TargetDir -Force | Out-Null
  }
    
  Set-Location $TargetDir
    
  # Create .plaesy directory and copy all framework files
  if (!(Test-Path ".plaesy")) {
    New-Item -ItemType Directory -Path ".plaesy" -Force | Out-Null
  }
    
  # Copy all core framework files to .plaesy
  $coreDirectories = @("memory", "templates", "prompts", "checklists", "scripts", "chatmodes", "instructions")
  foreach ($dir in $coreDirectories) {
    $sourcePath = "$PleasyHome\$dir"
    $targetPath = ".plaesy\$dir"
    if (Test-Path $sourcePath) {
      Copy-Item -Path $sourcePath -Destination $targetPath -Recurse -Force -ErrorAction SilentlyContinue
    }
  }
    
  Write-Success "Core framework files copied"
}

function Set-AISpecificConfig {
  param([string]$AIChoice)
    
  Write-Step "Configuring for $AIChoice..."
    
  switch ($AIChoice) {
    "copilot" {
      # GitHub Copilot configuration
      if (!(Test-Path ".github")) {
        New-Item -ItemType Directory -Path ".github" -Force | Out-Null
      }
            
      # Copy chatmodes for GitHub Copilot
      if (Test-Path ".plaesy\chatmodes") {
        Copy-Item -Path ".plaesy\chatmodes" -Destination ".github\" -Recurse -Force
      }

      # Copy prompts for GitHub Copilot
      if (Test-Path ".plaesy\prompts") {
        Copy-Item -Path ".plaesy\prompts" -Destination ".github\" -Recurse -Force
      }
            
      # Copy copilot instructions from source file instead of embedding
      if (Test-Path ".plaesy\instructions\copilot-instructions.md") {
        Copy-Item -Path ".plaesy\instructions\copilot-instructions.md" -Destination ".github\copilot-instructions.md" -Force
      }
      else {
        Write-Warning "Source file .plaesy\instructions\copilot-instructions.md not found, creating minimal instructions"
        $minimalInstructions = @'
# GitHub Copilot Instructions for Plaesy Spec-Kit

## 🏛️ Constitutional Priority (LOAD FIRST)
**MANDATORY**: Always load `.plaesy\memory\constitution.md` first for non-negotiable framework rules.

## Framework Structure
This project uses existing chatmodes, instructions, templates from the `.plaesy\` directory.

For complete instructions, see: `.plaesy\instructions\copilot-instructions.md`
'@
        $minimalInstructions | Out-File -FilePath ".github\copilot-instructions.md" -Encoding UTF8
      }
    }
        
    "cursor" {
      # Cursor AI configuration - use existing copilot instructions
      if (!(Test-Path ".cursor")) {
        New-Item -ItemType Directory -Path ".cursor" -Force | Out-Null
      }
            
      # Copy existing chatmodes for Cursor AI
      if (Test-Path ".plaesy\chatmodes") {
        Copy-Item -Path ".plaesy\chatmodes" -Destination ".cursor\" -Recurse -Force
      }
            
      # Copy existing prompts for Cursor AI
      if (Test-Path ".plaesy\prompts") {
        if (!(Test-Path ".cursor\prompts")) {
          New-Item -ItemType Directory -Path ".cursor\prompts" -Force | Out-Null
        }
        Copy-Item -Path ".plaesy\prompts\*.md" -Destination ".cursor\prompts\" -Force
      }
            
      # Use existing copilot instructions with Cursor-specific adaptations
      if (Test-Path ".plaesy\instructions\copilot-instructions.md") {
        # Copy and adapt copilot instructions for Cursor
        Copy-Item -Path ".plaesy\instructions\copilot-instructions.md" -Destination ".cursor\instructions.md" -Force
                
        # Add Cursor-specific header
        $cursorHeader = @'
# Cursor AI Instructions for Plaesy Spec-Kit

> **Note**: This uses the same constitutional framework as GitHub Copilot with Cursor-specific adaptations.

'@
        $existingContent = Get-Content ".cursor\instructions.md" -Raw
        $cursorHeader + $existingContent | Out-File -FilePath ".cursor\instructions.md" -Encoding UTF8
      }
      else {
        Write-Warning "Source file .plaesy\instructions\copilot-instructions.md not found"
      }
    }
        
    "windsurf" {
      # Windsurf AI configuration - use existing copilot instructions
      if (!(Test-Path ".windsurf")) { 
        New-Item -ItemType Directory -Path ".windsurf" -Force | Out-Null
      }
            
      # Copy existing chatmodes for Windsurf AI
      if (Test-Path ".plaesy\chatmodes") {
        Copy-Item -Path ".plaesy\chatmodes" -Destination ".windsurf\" -Recurse -Force
      }
            
      # Copy existing prompts for Windsurf AI
      if (Test-Path ".plaesy\prompts") {
        if (!(Test-Path ".windsurf\prompts")) {
          New-Item -ItemType Directory -Path ".windsurf\prompts" -Force | Out-Null
        }
        Copy-Item -Path ".plaesy\prompts\*.md" -Destination ".windsurf\prompts\" -Force
      }
            
      # Use existing copilot instructions with Windsurf-specific adaptations
      if (Test-Path ".plaesy\instructions\copilot-instructions.md") {
        # Copy and adapt copilot instructions for Windsurf
        Copy-Item -Path ".plaesy\instructions\copilot-instructions.md" -Destination ".windsurf\instructions.md" -Force
                
        # Add Windsurf-specific header
        $windsurfHeader = @'
# Windsurf AI Instructions for Plaesy Spec-Kit

> **Note**: This uses the same constitutional framework as GitHub Copilot with Windsurf-specific adaptations.

'@
        $existingContent = Get-Content ".windsurf\instructions.md" -Raw
        $windsurfHeader + $existingContent | Out-File -FilePath ".windsurf\instructions.md" -Encoding UTF8
      }
      else {
        Write-Warning "Source file .plaesy\instructions\copilot-instructions.md not found"
      }
    }

    "trae-ai" {
      # Trae.ai configuration - advanced multi-agent platform
      if (!(Test-Path ".trae")) {
        New-Item -ItemType Directory -Path ".trae\agents\plaesy" -Force | Out-Null
        New-Item -ItemType Directory -Path ".trae\workflows" -Force | Out-Null
      }
            
      $traeConfig = @'
# Trae.ai Configuration for Plaesy Constitutional Framework
# Uses existing chatmodes, instructions, templates, and memory

plaesy_framework:
  version: "1.0.0"
  constitutional_mode: true
  framework_base: ".plaesy"
  
agents:
  dev:
    role: "Full Stack Developer"
    chatmode: ".plaesy/chatmodes/dev.chatmode.md"
    instructions: ".plaesy/instructions/tdd-enforcement.instructions.md"
    constitutional_rules: true
    tdd_enforcement: mandatory
    
  architect:
    role: "Solution Architect"
    chatmode: ".plaesy/chatmodes/sa.chatmode.md"
    instructions: ".plaesy/instructions/security-and-owasp.instructions.md"
    interface_contracts: mandatory
    security_first: true
    
  qa:
    role: "Quality Assurance"
    chatmode: ".plaesy/chatmodes/qa.chatmode.md"
    checklist: ".plaesy/checklists/qa.checklist.md"
    coverage_requirement: 90
    real_dependencies: true
    
  business_analyst:
    role: "Business Analyst"
    chatmode: ".plaesy/chatmodes/ba.chatmode.md"
    templates: ".plaesy/templates/spec-template.md"
    
  business_owner:
    role: "Business Owner"
    chatmode: ".plaesy/chatmodes/bo.chatmode.md"
    templates: ".plaesy/templates/idea.template.md"
    
  product_manager:
    role: "Product Manager" 
    chatmode: ".plaesy/chatmodes/pm.chatmode.md"
    checklist: ".plaesy/checklists/pm.checklist.md"
    
  devops:
    role: "DevOps Engineer"
    chatmode: ".plaesy/chatmodes/devops.chatmode.md"
    instructions: ".plaesy/instructions/devops-core-principles.instructions.md"
    
  security:
    role: "Security Specialist"
    chatmode: ".plaesy/chatmodes/security.chatmode.md"
    instructions: ".plaesy/instructions/security-and-owasp.instructions.md"

workflows:
  constitutional_development:
    stages:
      - specify
      - plan  
      - tasks
      - implement
    quality_gates: true
    constitution: ".plaesy/memory/constitution.md"
    knowledge_base: ".plaesy/memory/knowledge-base.md"
'@
      $traeConfig | Out-File -FilePath ".trae\config.yaml" -Encoding UTF8

      $traeWorkflow = @'
# Constitutional Development Workflow for Trae.ai
# Uses existing templates and framework components

name: "Plaesy Constitutional Development"
description: "Automated constitutional development workflow with quality gates"

framework_references:
  constitution: ".plaesy/memory/constitution.md"
  knowledge_base: ".plaesy/memory/knowledge-base.md"
  test_priorities: ".plaesy/memory/test-priorities-matrix.md"

stages:
  specify:
    agent: "business_analyst"
    chatmode: ".plaesy/chatmodes/ba.chatmode.md"
    input: "feature_description"
    output: "specification"
    template: ".plaesy/templates/spec-template.md"
    prompt: ".plaesy/prompts/specify.prompt.md"
    
  plan:
    agent: "solution_architect"
    chatmode: ".plaesy/chatmodes/sa.chatmode.md"
    input: "specification"
    output: "technical_plan"
    template: ".plaesy/templates/plan-template.md"
    prompt: ".plaesy/prompts/plan.prompt.md"
    instructions: ".plaesy/instructions/security-and-owasp.instructions.md"
    dependencies: ["specify"]
    
  tasks:
    agent: "technical_lead"
    chatmode: ".plaesy/chatmodes/dev.chatmode.md"
    input: "technical_plan"
    output: "implementation_tasks"
    template: ".plaesy/templates/tasks-template.md"
    prompt: ".plaesy/prompts/tasks.prompt.md"
    instructions: ".plaesy/instructions/tdd-enforcement.instructions.md"
    dependencies: ["plan"]
    
  implement:
    agent: "developer"
    chatmode: ".plaesy/chatmodes/dev.chatmode.md"
    input: "implementation_tasks"
    output: "code"
    constitutional_compliance: true
    tdd_enforcement: true
    instructions: ".plaesy/instructions/tdd-enforcement.instructions.md"
    dependencies: ["tasks"]

quality_gates:
  - constitutional_compliance_check:
      constitution: ".plaesy/memory/constitution.md"
  - tdd_validation:
      instructions: ".plaesy/instructions/tdd-enforcement.instructions.md"
  - qa_validation:
      checklist: ".plaesy/checklists/qa.checklist.md"
      chatmode: ".plaesy/chatmodes/qa.chatmode.md"
  - security_review:
      instructions: ".plaesy/instructions/security-and-owasp.instructions.md"
      chatmode: ".plaesy/chatmodes/security.chatmode.md"
  - story_completion:
      checklist: ".plaesy/checklists/story-done.checklist.md"
'@
      $traeWorkflow | Out-File -FilePath ".trae\constitutional-workflow.yaml" -Encoding UTF8

      # Create Trae.ai agent definitions using existing chatmode files
      # Copy and link to existing framework files
      if (Test-Path ".plaesy\chatmodes\dev.chatmode.md") {
        Copy-Item ".plaesy\chatmodes\dev.chatmode.md" ".trae\agents\plaesy\dev.yaml"
      }
      else {
        "# Developer agent - see .plaesy/chatmodes/dev.chatmode.md" | Out-File -FilePath ".trae\agents\plaesy\dev.yaml" -Encoding UTF8
      }
            
      if (Test-Path ".plaesy\chatmodes\sa.chatmode.md") {
        Copy-Item ".plaesy\chatmodes\sa.chatmode.md" ".trae\agents\plaesy\architect.yaml"
      }
      else {
        "# Solution Architect agent - see .plaesy/chatmodes/sa.chatmode.md" | Out-File -FilePath ".trae\agents\plaesy\architect.yaml" -Encoding UTF8
      }
            
      if (Test-Path ".plaesy\chatmodes\qa.chatmode.md") {
        Copy-Item ".plaesy\chatmodes\qa.chatmode.md" ".trae\agents\plaesy\qa.yaml"
      }
      else {
        "# QA agent - see .plaesy/chatmodes/qa.chatmode.md" | Out-File -FilePath ".trae\agents\plaesy\qa.yaml" -Encoding UTF8
      }
            
      if (Test-Path ".plaesy\chatmodes\ba.chatmode.md") {
        Copy-Item ".plaesy\chatmodes\ba.chatmode.md" ".trae\agents\plaesy\business-analyst.yaml"
      }
      else {
        "# Business Analyst agent - see .plaesy/chatmodes/ba.chatmode.md" | Out-File -FilePath ".trae\agents\plaesy\business-analyst.yaml" -Encoding UTF8
      }
            
      # Create agent registry pointing to existing files
      $traeRegistry = @'
# Trae.ai Agent Registry for Plaesy Constitutional Framework
# References existing chatmode files and framework components

framework_base: ".plaesy"

agents:
  developer:
    chatmode: ".plaesy/chatmodes/dev.chatmode.md"
    instructions: ".plaesy/instructions/tdd-enforcement.instructions.md"
    templates: 
      - ".plaesy/templates/tasks-template.md"
      - ".plaesy/templates/application.template.md"
    checklists:
      - ".plaesy/checklists/story-done.checklist.md"
    
  solution_architect:
    chatmode: ".plaesy/chatmodes/sa.chatmode.md"
    instructions: 
      - ".plaesy/instructions/security-and-owasp.instructions.md"
      - ".plaesy/instructions/performance-optimization.instructions.md"
    templates:
      - ".plaesy/templates/plan-template.md"
      - ".plaesy/templates/api-documentation.template.md"
    
  qa_engineer:
    chatmode: ".plaesy/chatmodes/qa.chatmode.md"
    instructions: ".plaesy/instructions/tdd-enforcement.instructions.md"
    templates:
      - ".plaesy/templates/test-plan.template.md"
    checklists:
      - ".plaesy/checklists/qa.checklist.md"
      - ".plaesy/checklists/story-done.checklist.md"
    memory:
      - ".plaesy/memory/test-priorities-matrix.md"
  
  business_analyst:
    chatmode: ".plaesy/chatmodes/ba.chatmode.md"
    templates:
      - ".plaesy/templates/spec-template.md"
      - ".plaesy/templates/research.template.md"
    prompts:
      - ".plaesy/prompts/specify.prompt.md"
      - ".plaesy/prompts/idea.prompt.md"
    
  business_owner:
    chatmode: ".plaesy/chatmodes/bo.chatmode.md"
    templates:
      - ".plaesy/templates/idea.template.md"
      - ".plaesy/templates/brainstorming.template.md"
    prompts:
      - ".plaesy/prompts/idea.prompt.md"
      - ".plaesy/prompts/plan.prompt.md"
      
  product_manager:
    chatmode: ".plaesy/chatmodes/pm.chatmode.md"
    templates:
      - ".plaesy/templates/spec-template.md"
      - ".plaesy/templates/user-documentation.template.md"
    checklists:
      - ".plaesy/checklists/pm.checklist.md"
  
  devops_engineer:
    chatmode: ".plaesy/chatmodes/devops.chatmode.md"
    instructions:
      - ".plaesy/instructions/devops-core-principles.instructions.md"
      - ".plaesy/instructions/kubernetes-deployment-best-practices.instructions.md"
      - ".plaesy/instructions/terraform.instructions.md"
    templates:
      - ".plaesy/templates/deployment-guide.template.md"
  
  security_specialist:
    chatmode: ".plaesy/chatmodes/security.chatmode.md"
    instructions: ".plaesy/instructions/security-and-owasp.instructions.md"
    templates:
      - ".plaesy/templates/security-assessment.template.md"
  
  technical_writer:
    chatmode: ".plaesy/chatmodes/tw.chatmode.md"
    templates:
      - ".plaesy/templates/user-documentation.template.md"
      - ".plaesy/templates/api-documentation.template.md"
      - ".plaesy/templates/wiki.template.md"

constitutional_framework:
  constitution: ".plaesy/memory/constitution.md"
  knowledge_base: ".plaesy/memory/knowledge-base.md"
  test_framework: ".plaesy/memory/test-priorities-matrix.md"
'@
      $traeRegistry | Out-File -FilePath ".trae\agents\plaesy\agent-registry.yaml" -Encoding UTF8
    }

    "qwen-code" {
      # Qwen Code configuration
      if (!(Test-Path ".qwen")) {
        New-Item -ItemType Directory -Path ".qwen\plaesy-method" -Force | Out-Null
      }
            
      $qwenInstructions = @'
# Qwen Code Instructions for Plaesy Spec-Kit

## Constitutional Development Framework
This project uses the Plaesy Spec-Kit constitutional development framework.

## Agent Commands
Use these commands to activate constitutional roles:
- `/plaesy:dev` - Full Stack Developer with TDD enforcement
- `/plaesy:architect` - Solution Architect with interface contracts
- `/plaesy:qa` - Quality Assurance with 90%+ coverage requirement
- `/plaesy:devops` - DevOps with security-first deployment

## Constitutional Framework Rules
1. TDD mandatory for all code implementation
2. Interface contracts required for service boundaries  
3. Real dependencies in integration tests (no mocks)
4. Security-first design patterns
5. 90%+ test coverage requirement

## Templates Available
- Specification: `.plaesy/templates/spec-template.md`
- Planning: `.plaesy/templates/plan-template.md`
- Tasks: `.plaesy/templates/tasks-template.md`

Constitutional framework loaded from `.qwen/plaesy-method/` directory.
'@
      $qwenInstructions | Out-File -FilePath ".qwen\constitutional-instructions.md" -Encoding UTF8
    }

    "kilo-code" {
      # Kilo Code configuration with custom modes
      $kiloModes = @'
{
  "modes": [
    {
      "name": "plaesy-dev",
      "description": "Plaesy Constitutional Developer",
      "systemPrompt": "You are a constitutional full-stack developer following Plaesy Spec-Kit framework. Always enforce TDD Red-Green-Refactor cycle, implement interface contracts for all services, and use real dependencies in integration tests. Load constitutional rules from .plaesy/memory/constitution.md",
      "temperature": 0.7
    },
    {
      "name": "plaesy-architect", 
      "description": "Plaesy Constitutional Architect",
      "systemPrompt": "You are a solution architect following Plaesy constitutional principles. Design secure, scalable architectures with interface contracts. Enforce security-first design patterns and ensure all services have well-defined boundaries. Reference .plaesy/templates/ for structure.",
      "temperature": 0.8
    },
    {
      "name": "plaesy-qa",
      "description": "Plaesy Constitutional QA",
      "systemPrompt": "You are a QA engineer enforcing Plaesy constitutional quality standards. Require 90%+ test coverage, validate TDD compliance, ensure real dependencies in integration tests. Use .plaesy/checklists/ for validation.",
      "temperature": 0.6
    }
  ]
}
'@
      $kiloModes | Out-File -FilePath ".kilocodemodes" -Encoding UTF8

      if (!(Test-Path ".kilo")) {
        New-Item -ItemType Directory -Path ".kilo" -Force | Out-Null
      }
      $kiloConfig = @'
{
  "constitutional_framework": true,
  "tdd_enforcement": "mandatory",
  "interface_contracts": "required",
  "security_first": true,
  "coverage_requirement": 90
}
'@
      $kiloConfig | Out-File -FilePath ".kilo\plaesy-modes.json" -Encoding UTF8
    }

    "codex-cli" {
      # Codex CLI configuration with project memory
      $codexAgents = @'
# Constitutional Development Agents - Plaesy Spec-Kit

## Framework Overview
This project uses the Plaesy Spec-Kit constitutional development framework with strict quality gates and automated enforcement.

## Available Agents

### Developer Agent
**Role**: Full Stack Developer with TDD enforcement
**Activation**: "As dev, please..."
**Capabilities**:
- Implement features using TDD Red-Green-Refactor cycle
- Create interface contracts for all services  
- Use real dependencies in integration tests
- Follow security-first design patterns
- Maintain 90%+ test coverage

### Solution Architect Agent  
**Role**: Technical Architecture Designer
**Activation**: "As architect, please..."
**Capabilities**:
- Design scalable, secure system architectures
- Define service boundaries and interface contracts
- Enforce constitutional design principles
- Plan technology stack and integration patterns

### QA Agent
**Role**: Quality Assurance Engineer
**Activation**: "As qa, please..."  
**Capabilities**:
- Validate TDD compliance and test coverage
- Perform constitutional framework compliance checks
- Execute quality gate validations
- Review security and performance requirements

## Constitutional Framework Rules
1. **TDD Mandatory**: All code must follow Red-Green-Refactor cycle
2. **Interface Contracts**: Required for all service boundaries
3. **Real Dependencies**: No mocks in integration tests
4. **Security First**: OWASP compliance and security-first design
5. **Quality Gates**: Automated validation through checklists

Reference `.plaesy/memory/constitution.md` for complete framework details.
'@
      $codexAgents | Out-File -FilePath "AGENTS.md" -Encoding UTF8

      if (!(Test-Path ".codex")) {
        New-Item -ItemType Directory -Path ".codex" -Force | Out-Null
      }
      $codexConfig = @'
[plaesy_framework]
version = "1.0.0"
constitutional_mode = true

[enforcement]
tdd = "mandatory"
interface_contracts = "required"  
real_dependencies = true
security_first = true
coverage_minimum = 90

[templates]
spec = ".plaesy/templates/spec-template.md"
plan = ".plaesy/templates/plan-template.md" 
tasks = ".plaesy/templates/tasks-template.md"
'@
      $codexConfig | Out-File -FilePath ".codex\plaesy-config.toml" -Encoding UTF8
    }

    "codex-web" {
      # Codex Web configuration - ensure files are committed to git
      $codexWebAgents = @'
# Constitutional Development Agents - Plaesy Spec-Kit (Web)

## Framework Overview
This project uses the Plaesy Spec-Kit constitutional development framework accessible through Codex Web interface.

## Available Agents

### Developer Agent
**Role**: Full Stack Developer with constitutional enforcement
**Usage**: "As dev, implement user authentication with TDD..."
**Constitutional Rules**:
- TDD Red-Green-Refactor cycle mandatory
- Interface contracts for all services
- Real dependencies in integration tests
- Security-first design patterns

### Architect Agent  
**Role**: Solution Architecture with constitutional compliance
**Usage**: "As architect, design microservices architecture..."
**Constitutional Rules**:
- Service boundary definition required
- Interface contracts mandatory
- Security-first architectural patterns
- Scalability and maintainability focus

### QA Agent
**Role**: Quality Assurance with 90%+ coverage
**Usage**: "As qa, validate test coverage and compliance..."  
**Constitutional Rules**:
- 90%+ test coverage validation
- TDD compliance verification
- Constitutional framework adherence
- Quality gate enforcement

## Web Integration Notes
- All `.plaesy-core/` files are committed to repository
- Constitutional framework accessible through cloud interface
- Run installer to refresh when core framework updates
- Reference agents naturally in prompts for automatic activation

Framework documentation: `.plaesy-core/memory/constitution.md`
'@
      $codexWebAgents | Out-File -FilePath "AGENTS.md" -Encoding UTF8

      # Create .gitignore modification to ensure core is committed
      if (Test-Path ".gitignore") {
        # Remove any existing .plaesy entries and add comment
        $gitignoreContent = Get-Content ".gitignore" | Where-Object { $_ -notmatch "\.plaesy" }
        $gitignoreContent += @(
          "",
          "# Plaesy Constitutional Framework - Core committed for Codex Web",
          "# .plaesy-core/"
        )
        $gitignoreContent | Out-File -FilePath ".gitignore" -Encoding UTF8
      }
    }

    "opencode-cli" {
      # OpenCode CLI configuration
      $opencodeConfig = @'
{
  // OpenCode Configuration for Plaesy Constitutional Framework
  "instructions": [
    "./.plaesy-core/constitutional-config.yaml"
  ],
  "constitutional_framework": {
    "version": "1.0.0",
    "tdd_enforcement": true,
    "interface_contracts": "mandatory",
    "security_first": true,
    "coverage_requirement": 90
  },
  "agents": {
    "dev": {
      "role": "Constitutional Developer",
      "tdd_enforced": true,
      "interface_contracts": true
    },
    "architect": {
      "role": "Solution Architect", 
      "security_first": true,
      "service_boundaries": "required"
    },
    "qa": {
      "role": "Quality Assurance",
      "coverage_minimum": 90,
      "real_dependencies": true
    }
  }
}
'@
      $opencodeConfig | Out-File -FilePath "opencode.jsonc" -Encoding UTF8

      if (!(Test-Path ".opencode")) {
        New-Item -ItemType Directory -Path ".opencode" -Force | Out-Null
      }
      $opencodeInstructions = @'
# OpenCode CLI Instructions for Plaesy Constitutional Framework

constitutional_framework:
  name: "Plaesy Spec-Kit"
  version: "1.0.0"
  
instructions:
  - "Load constitutional rules from .plaesy/memory/constitution.md"
  - "Enforce TDD Red-Green-Refactor cycle for all code"
  - "Require interface contracts for service boundaries"
  - "Use real dependencies in integration tests (no mocks)"
  - "Apply security-first design patterns"
  - "Maintain 90%+ test coverage"

templates:
  specification: ".plaesy/templates/spec-template.md"
  planning: ".plaesy/templates/plan-template.md"
  tasks: ".plaesy/templates/tasks-template.md"

quality_gates:
  - "Constitutional compliance check"
  - "TDD validation"
  - "Interface contract verification"
  - "Security review"
  - "Test coverage validation"
'@
      $opencodeInstructions | Out-File -FilePath ".opencode\plaesy-instructions.yaml" -Encoding UTF8
    }
        
    { $_ -in @("claude", "chatgpt", "gemini", "local") } {
      # Generic AI configuration - use existing copilot instructions
      if (!(Test-Path ".ai-config")) {
        New-Item -ItemType Directory -Path ".ai-config" -Force | Out-Null
      }
            
      # Copy existing chatmodes for generic AI
      if (Test-Path ".plaesy\chatmodes") {
        Copy-Item -Path ".plaesy\chatmodes" -Destination ".ai-config\" -Recurse -Force
      }
            
      # Copy existing prompts for generic AI
      if (Test-Path ".plaesy\prompts") {
        if (!(Test-Path ".ai-config\prompts")) {
          New-Item -ItemType Directory -Path ".ai-config\prompts" -Force | Out-Null
        }
        Copy-Item -Path ".plaesy\prompts\*.md" -Destination ".ai-config\prompts\" -Force
      }
            
      # Use existing copilot instructions with generic AI adaptations
      if (Test-Path ".plaesy\instructions\copilot-instructions.md") {
        # Copy and adapt copilot instructions for generic AI
        Copy-Item -Path ".plaesy\instructions\copilot-instructions.md" -Destination ".ai-config\instructions.md" -Force
                
        # Add generic AI header
        $aiHeader = @'
# AI Assistant Instructions for Plaesy Spec-Kit

> **Note**: This uses the same constitutional framework as GitHub Copilot, adapted for generic AI assistants.

'@
        $existingContent = Get-Content ".ai-config\instructions.md" -Raw
        $aiHeader + $existingContent | Out-File -FilePath ".ai-config\instructions.md" -Encoding UTF8
      }
      else {
        Write-Warning "Source file .plaesy\instructions\copilot-instructions.md not found"
      }
    }
        
    "none" {
      # Manual development setup
      $manualGuide = @'
# Plaesy Spec-Kit Framework Guide

## Constitutional Development
This project uses constitutional development principles from `memory/constitution.md`.

## Development Workflow
1. **Ideation** (`prompts/idea.prompt.md`) - Capture ideas
2. **Specification** (`prompts/specify.prompt.md`) - Define requirements
3. **Planning** (`prompts/plan.prompt.md`) - Technical architecture
4. **Tasks** (`prompts/tasks.prompt.md`) - Implementation breakdown

## Quality Gates
Check `checklists/` directory for validation steps:
- `story-done.checklist.md` - Completion validation
- `qa.checklist.md` - Quality assurance
- `constitutional.checklist.md` - Framework compliance

## Templates
Use templates from `templates/` directory for consistent deliverables.

## Constitutional Principles
- Test-Driven Development mandatory
- Interface contracts required
- Real dependencies for integration tests
- Security-first design patterns

Refer to `memory/constitution.md` for complete constitutional framework.
'@
      $manualGuide | Out-File -FilePath "FRAMEWORK-GUIDE.md" -Encoding UTF8
    }
  }
    
  Write-Success "AI-specific configuration completed"
}

function New-ProjectStructure {
  Write-Step "Creating project structure..."
    
  # Create basic project structure
  $directories = @("src", "tests", "docs")
  foreach ($dir in $directories) {
    if (!(Test-Path $dir)) {
      New-Item -ItemType Directory -Path $dir -Force | Out-Null
    }
  }
    
  # Create .gitignore if it doesn't exist
  if (!(Test-Path ".gitignore")) {
    $gitignoreContent = @'
# Dependencies
node_modules/
vendor/
target/
dist/
build/

# IDE
.vscode/
.idea/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Logs
*.log
logs/

# Environment
.env
.env.local
.env.*.local

# Temporary files
tmp/
temp/
*.tmp
'@
    $gitignoreContent | Out-File -FilePath ".gitignore" -Encoding ASCII
  }
    
  Write-Success "Project structure created"
}

function Write-Completion {
  param([string]$AIChoice)
    
  Write-Host ""
  Write-Host "🎉 Plaesy Spec-Kit initialization completed!" -ForegroundColor Green
  Write-Host ""
  Write-Host "📁 Project initialized in: $(Get-Location)" -ForegroundColor Cyan
  Write-Host "🤖 AI Assistant: $AIChoice" -ForegroundColor Cyan
  Write-Host ""
    
  switch ($AIChoice) {
    "copilot" {
      Write-Host "GitHub Copilot users:" -ForegroundColor Blue
      Write-Host "• Framework instructions loaded in .github/"
      Write-Host "• Use commands like /idea, /specify, /plan, /tasks"
      Write-Host "• Use @dev, @sa, @qa for role-based assistance"
    }
    "cursor" {
      Write-Host "Cursor AI users:" -ForegroundColor Blue
      Write-Host "• Framework instructions loaded in .cursor/"
      Write-Host "• Load chatmodes from .cursor/chatmodes/ as needed"
    }
    { $_ -in @("claude", "chatgpt", "gemini", "local") } {
      Write-Host "AI Assistant users:" -ForegroundColor Blue
      Write-Host "• Framework instructions in .ai-config/"
      Write-Host "• Load constitution.md and relevant chatmodes"
    }
    "none" {
      Write-Host "Manual development:" -ForegroundColor Blue
      Write-Host "• Follow FRAMEWORK-GUIDE.md"
      Write-Host "• Use templates and checklists as reference"
    }
  }
    
  Write-Host ""
  Write-Host "🏛️ Constitutional Development Framework Active" -ForegroundColor Magenta
  Write-Host "   Quality through discipline. Excellence through automation."
}

# Main execution
Write-Host ""
Write-Host "🏛️  Plaesy Spec-Kit Initialization" -ForegroundColor Magenta
Write-Host "   Constitutional Development Framework" -ForegroundColor Magenta
Write-Host ""
Write-Host "Initializing constitutional development framework..." -ForegroundColor Cyan
Write-Host ""

$AIChoice = Select-AIAssistant
Copy-CoreFramework
Set-AISpecificConfig -AIChoice $AIChoice
New-ProjectStructure
Write-Completion -AIChoice $AIChoice