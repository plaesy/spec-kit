# Plaesy Project Initialization Script (PowerShell)
# Constitutional Development Framework

param(
  [string]$TargetDir = ".",
  [string]$AI = $null,
  [string[]]$Arguments = @()
)

$ErrorActionPreference = "Stop"

# Colors
function Write-Step { 
  param([string]$Message)
  Write-Host "[STEP] $Message" -ForegroundColor Blue
}

function Write-Success { 
  param([string]$Message)
  Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Write-Warning { 
  param([string]$Message)
  Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

function Write-Error-Custom { 
  param([string]$Message)
  Write-Host "[ERROR] $Message" -ForegroundColor Red
  exit 1
}

$PleasyHome = "$env:USERPROFILE\.plaesy"

# Validate environment
function Test-InitEnvironment {
  if (!(Test-Path $PleasyHome)) {
    Write-Error-Custom "Plaesy framework not found. Please run the installer first."
  }
    
  if (!(Test-Path $TargetDir)) {
    Write-Error-Custom "Target directory '$TargetDir' does not exist."
  }
    
  try {
    $testFile = Join-Path $TargetDir "test-write-$(Get-Random).tmp"
    "test" | Out-File -FilePath $testFile -ErrorAction Stop
    Remove-Item $testFile -ErrorAction SilentlyContinue
  }
  catch {
    Write-Error-Custom "Target directory '$TargetDir' is not writable."
  }
}

if (!(Test-Path $PleasyHome)) {
  Write-Error-Custom "Plaesy framework not found. Please run the installer first."
}

function Select-AIAssistant {
  # Check for --ai parameter in various formats
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
      { $_ -in @("qoder", "qoder-ai") } { return "qoder-ai" }
      { $_ -in @("local", "ollama") } { return "local" }
      { $_ -in @("none", "manual") } { return "none" }
      default {
        Write-Warning "Unknown AI assistant: $AI. Falling back to interactive selection."
      }
    }
  }

  # Check command line arguments for --ai parameter
  foreach ($arg in $Arguments) {
    if ($arg -match "^--ai=(.+)$") {
      $aiValue = $matches[1].ToLower()
      switch ($aiValue) {
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
        { $_ -in @("qoder", "qoder-ai") } { return "qoder-ai" }
        { $_ -in @("local", "ollama") } { return "local" }
        { $_ -in @("none", "manual") } { return "none" }
        default {
          Write-Warning "Unknown AI assistant: $aiValue. Falling back to interactive selection."
        }
      }
    }
  }
    
  Write-Host "Select your AI Assistant:" -ForegroundColor Magenta
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
  Write-Host "13. Qoder AI"
  Write-Host "14. Local AI (Ollama/LM Studio)"
  Write-Host "15. No AI (Manual development)"
  Write-Host ""
    
  do {
    $choice = Read-Host "Choose your AI assistant (1-15)"
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
      "13" { return "qoder-ai" }
      "14" { return "local" }
      "15" { return "none" }
      default { Write-Host "Please enter a number between 1-15" -ForegroundColor Red }
    }
  } while ($true)
}

function Copy-CoreFramework {
  param([string]$AIChoice)
  Write-Step "Setting up constitutional framework for $AIChoice..."
    
  # Create target directory if it doesn't exist
  if (!(Test-Path $TargetDir)) {
    New-Item -ItemType Directory -Path $TargetDir -Force | Out-Null
  }
    
  Set-Location $TargetDir
    
  # Create .plaesy directory (shared framework files)
  if (!(Test-Path ".plaesy")) {
    New-Item -ItemType Directory -Path ".plaesy" -Force | Out-Null
  }
    
  # Copy essential framework files to .plaesy with error checking
  $coreDirectories = @("memory", "templates", "chatmodes", "instructions", "checklists", "prompts")
  $failedCopies = @()
    
  foreach ($dir in $coreDirectories) {
    $sourcePath = "$PleasyHome\$dir"
    $targetPath = ".plaesy\$dir"
    if (Test-Path $sourcePath) {
      try {
        Copy-Item -Path $sourcePath -Destination $targetPath -Recurse -Force -ErrorAction Stop
      }
      catch {
        $failedCopies += $dir
      }
    }
    else {
      Write-Warning "Source directory $sourcePath not found"
    }
  }
    
  if ($failedCopies.Count -gt 0) {
    Write-Warning "Failed to copy some directories: $($failedCopies -join ', ')"
  }
    
  # Copy scripts based on AI choice
  switch ($AIChoice) {
    { $_ -in @("copilot", "cursor", "windsurf", "claude", "trae-ai", "chatgpt", "gemini", "qwen-code", "kilo-code", "codex-cli", "codex-web", "opencode-cli", "qoder-ai", "local") } {
      # All AI types need these scripts
      if (!(Test-Path ".plaesy\scripts")) {
        New-Item -ItemType Directory -Path ".plaesy\scripts" -Force | Out-Null
      }
      $scriptFiles = @(
        "common.ps1",
        "check-task-prerequisites.ps1",
        "create-new-feature.ps1",
        "create-new-idea.ps1",
        "create-new-plan.ps1",
        "detect-secrets-local.ps1",
        "get-feature-paths.ps1",
        "update-agent-context.ps1",
        "validate-llms.ps1"
      )
      foreach ($script in $scriptFiles) {
        $scriptPath = "$PleasyHome\scripts\powershell\$script"
        if (Test-Path $scriptPath) {
          if (!(Test-Path ".plaesy\scripts\powershell")) {
            New-Item -ItemType Directory -Path ".plaesy\scripts\powershell" -Force | Out-Null
          }
          Copy-Item -Path $scriptPath -Destination ".plaesy\scripts\powershell\" -Force -ErrorAction SilentlyContinue
        }
      }
    }
    "none" {
      # Manual development needs all scripts for reference
      Copy-Item -Path "$PleasyHome\scripts" -Destination ".plaesy\" -Recurse -Force -ErrorAction SilentlyContinue
    }
  }
    
  # Inject AI-specific headers into prompt files
  Invoke-InjectAIHeaders -AIPlatform $AIChoice

  Write-Success "Core framework files copied for $AIChoice"
}

# Function to inject AI-specific headers into prompt and chatmode files
function Invoke-InjectAIHeaders {
  param([string]$AIPlatform)

  Write-Step "Injecting $AIPlatform-optimized headers..."

  # Map AI platform to header file
  $headerFile = switch ($AIPlatform) {
    "copilot" { "copilot.header.md" }
    "cursor" { "cursor.header.md" }
    "windsurf" { "windsurf.header.md" }
    "claude" { "claude.header.md" }
    "trae-ai" { "trae-ai.header.md" }
    "chatgpt" { "chatgpt.header.md" }
    "gemini" { "gemini.header.md" }
    "qwen-code" { "qwen-code.header.md" }
    { $_ -in @("kilo-code", "codex-cli", "codex-web") } { "codex-cli.header.md" }
    { $_ -in @("opencode-cli", "qoder-ai") } { "opencode-cli.header.md" }
    "local" { "local-ai.header.md" }
    "none" { "manual.header.md" }
    default { "manual.header.md" }
  }

  # Check if header file exists
  $headerPath = Join-Path $PleasyHome "templates\ai-headers\$headerFile"
  if (-not (Test-Path $headerPath)) {
    Write-Warning "Header file not found: $headerFile, using manual header"
    $headerPath = Join-Path $PleasyHome "templates\ai-headers\manual.header.md"
  }

  if (-not (Test-Path $headerPath)) {
    Write-Warning "No header files found, skipping header injection"
    return
  }

  # Load header content with error handling
  try {
    $headerContent = Get-Content $headerPath -Raw -ErrorAction Stop
  } catch {
    Write-Warning "Failed to read header file, skipping injection"
    return
  }

  # Find and inject headers into prompt files
  $filesProcessed = 0

  # Process prompt files
  if (Test-Path ".plaesy\prompts") {
    $promptFiles = Get-ChildItem ".plaesy\prompts\*.prompt.md" -File -ErrorAction SilentlyContinue
    foreach ($file in $promptFiles) {
      try {
        $originalContent = Get-Content $file.FullName -Raw -ErrorAction Stop
        $newContent = $headerContent + "`r`n`r`n" + $originalContent
        Set-Content -Path $file.FullName -Value $newContent -Encoding UTF8 -ErrorAction Stop
        $filesProcessed++
      } catch {
        Write-Warning "Failed to process file: $($file.Name)"
      }
    }
  }

  # Process chatmode files
  if (Test-Path ".plaesy\chatmodes") {
    $chatmodeFiles = Get-ChildItem ".plaesy\chatmodes\*.chatmode.md" -File -ErrorAction SilentlyContinue
    foreach ($file in $chatmodeFiles) {
      try {
        $originalContent = Get-Content $file.FullName -Raw -ErrorAction Stop
        $newContent = $headerContent + "`r`n`r`n" + $originalContent
        Set-Content -Path $file.FullName -Value $newContent -Encoding UTF8 -ErrorAction Stop
        $filesProcessed++
      } catch {
        Write-Warning "Failed to process file: $($file.Name)"
      }
    }
  }

  if ($filesProcessed -gt 0) {
    Write-Success "Injected $AIPlatform headers into $filesProcessed files"

    # Create configuration file with error handling
    try {
      $timestamp = Get-Date -Format "yyyy-MM-ddTHH:mm:ssZ" -ErrorAction Stop
      $config = @{
        ai_platform = $AIPlatform
        injection_timestamp = $timestamp
        processed_files = $filesProcessed
        header_file = $headerFile
        constitutional_framework_version = "3.0.0"
      } | ConvertTo-Json -Depth 3

      Set-Content -Path ".plaesy-headers.json" -Value $config -Encoding UTF8 -ErrorAction Stop
    } catch {
      Write-Warning "Failed to create configuration file: $($_.Exception.Message)"
    }
  } else {
    Write-Warning "No prompt or chatmode files found for header injection"
  }
}

function Set-AISpecificConfig {
  param([string]$AIChoice)

  Write-Step "Configuring for $AIChoice..."
    
  switch ($AIChoice) {
    "copilot" {
      # GitHub Copilot configuration (2025 conventions)
      if (!(Test-Path ".github")) {
        New-Item -ItemType Directory -Path ".github" -Force | Out-Null
      }
      if (!(Test-Path ".github\instructions")) {
        New-Item -ItemType Directory -Path ".github\instructions" -Force | Out-Null
      }
      if (!(Test-Path ".github\chatmodes")) {
        New-Item -ItemType Directory -Path ".github\chatmodes" -Force | Out-Null
      }
      if (!(Test-Path ".github\prompts")) {
        New-Item -ItemType Directory -Path ".github\prompts" -Force | Out-Null
      }

      # Create copilot-instructions.md (repository-wide custom instructions)
      if (Test-Path ".plaesy\instructions\plaesy.instructions.md") {
        Copy-Item -Path ".plaesy\instructions\plaesy.instructions.md" -Destination ".github\copilot-instructions.md" -Force
      }  

      # Copy chatmodes as path-specific instructions (standardized .chatmode.md)
      if (Test-Path ".plaesy\chatmodes") {
        Get-ChildItem ".plaesy\chatmodes\*.chatmode.md" | ForEach-Object {
          $filename = $_.BaseName -replace '\.chatmode$', ''
          Copy-Item $_.FullName ".github\chatmodes\$filename.chatmode.md" -Force
        }
      }

      # Copy prompts as path-specific instructions (standardized .prompt.md)
      if (Test-Path ".plaesy\prompts") {
        Get-ChildItem ".plaesy\prompts\*.prompt.md" | ForEach-Object {
          $filename = $_.BaseName -replace '\.prompt$', ''
          Copy-Item $_.FullName ".github\prompts\$filename.prompt.md" -Force
        }
      }
    }
        
    "cursor" {
      # Cursor AI configuration (2025 conventions - MDC format only)
      if (!(Test-Path ".cursor")) {
        New-Item -ItemType Directory -Path ".cursor" -Force | Out-Null
      }
      if (!(Test-Path ".cursor\rules")) {
        New-Item -ItemType Directory -Path ".cursor\rules" -Force | Out-Null
      }

      # Create main project rules in MDC format
      if (Test-Path ".plaesy\instructions\plaesy.instructions.md") {
        # Convert to MDC format with frontmatter
        $mdcHeader = @'
---
description: "Plaesy Constitutional Framework Rules"
globs: ["**/*"]
alwaysApply: true
---
'@
        $existingContent = Get-Content ".plaesy\instructions\plaesy.instructions.md" -Raw
        ($mdcHeader + $existingContent) | Out-File -FilePath ".cursor\rules\plaesy.mdc" -Encoding UTF8
      }

      # Copy chatmodes as scoped MDC rules (standardized .chatmode.md)
      if (Test-Path ".plaesy\chatmodes") {
        Get-ChildItem ".plaesy\chatmodes\*.chatmode.md" | ForEach-Object {
          $filename = $_.BaseName -replace '\.chatmode$', ''
          # Create MDC rule with scope
          $mdcRoleHeader = @"
---
description: "$filename role-based development rules"
globs: ["**/*"]
alwaysApply: false
---
"@
          $chatmodeContent = Get-Content $_.FullName -Raw
          ($mdcRoleHeader + $chatmodeContent) | Out-File -FilePath ".cursor\rules\$filename.mdc" -Encoding UTF8
        }
      }
      # Note: .cursorrules is deprecated and NOT created
    }
        
    "claude" {
      # Claude Code configuration (2025 conventions)
      if (!(Test-Path ".claude")) {
        New-Item -ItemType Directory -Path ".claude" -Force | Out-Null
      }
      if (!(Test-Path ".claude\commands")) {
        New-Item -ItemType Directory -Path ".claude\commands" -Force | Out-Null
      }

      # Create CLAUDE.md (main project instructions)
      if (Test-Path ".plaesy\instructions\plaesy.instructions.md") {
        Copy-Item -Path ".plaesy\instructions\plaesy.instructions.md" -Destination "CLAUDE.md" -Force
      }

      # Copy chatmodes as slash commands (standardized .chatmode.md)
      if (Test-Path ".plaesy\chatmodes") {
        Get-ChildItem ".plaesy\chatmodes\*.chatmode.md" | ForEach-Object {
          $filename = $_.Name
          Copy-Item $_.FullName ".claude\commands\$filename" -Force
        }
      }

      # Copy prompts as slash commands (standardized .prompt.md)
      if (Test-Path ".plaesy\prompts") {
        Get-ChildItem ".plaesy\prompts\*.prompt.md" | ForEach-Object {
          $filename = $_.Name
          Copy-Item $_.FullName ".claude\commands\$filename" -Force
        }
      }
    }
    "windsurf" {
      # Windsurf AI configuration - use existing copilot instructions
      if (!(Test-Path ".windsurf")) { 
        New-Item -ItemType Directory -Path ".windsurf" -Force | Out-Null
      }
            
      # Copy chatmodes (standardized .chatmode.md)
      if (Test-Path ".plaesy\chatmodes") {
        if (!(Test-Path ".windsurf\chatmodes")) {
          New-Item -ItemType Directory -Path ".windsurf\chatmodes" -Force | Out-Null
        }
        Copy-Item -Path ".plaesy\chatmodes\*.chatmode.md" -Destination ".windsurf\chatmodes\" -Force -ErrorAction SilentlyContinue
      }
            
      # Use existing copilot instructions with Windsurf-specific adaptations
      if (Test-Path ".plaesy\instructions\plaesy.instructions.md") {
        # Copy and adapt copilot instructions for Windsurf
        Copy-Item -Path ".plaesy\instructions\plaesy.instructions.md" -Destination ".windsurf\instructions.md" -Force

        # Add Windsurf-specific header
        $windsurfHeader = @'
# Windsurf AI Instructions for Plaesy Spec-Kit

> **Note**: This uses the same constitutional framework as GitHub Copilot with Windsurf-specific adaptations.

'@
        $existingContent = Get-Content ".windsurf\instructions.md" -Raw
        $windsurfHeader + $existingContent | Out-File -FilePath ".windsurf\instructions.md" -Encoding UTF8
      }
      else {
        Write-Warning "Source file .plaesy\instructions\plaesy.instructions.md not found"
      }
      
      # Copy existing prompts for Windsurf AI
      if (Test-Path ".plaesy\prompts") {
        if (!(Test-Path ".windsurf\prompts")) {
          New-Item -ItemType Directory -Path ".windsurf\prompts" -Force | Out-Null
        }
        Copy-Item -Path ".plaesy\prompts\*.md" -Destination ".windsurf\prompts\" -Force -ErrorAction SilentlyContinue
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
      # Qwen Code configuration - using existing framework files
      if (!(Test-Path ".qwen")) {
        New-Item -ItemType Directory -Path ".qwen\plaesy-method" -Force | Out-Null
      }
            
      # Copy chatmodes (standardized .chatmode.md)
      if (Test-Path ".plaesy\chatmodes") {
        Copy-Item -Path ".plaesy\chatmodes\*.chatmode.md" -Destination ".qwen\plaesy-method\" -Force -ErrorAction SilentlyContinue
      }
            
      $qwenInstructions = @'
# Qwen Code Instructions for Plaesy Spec-Kit

## Constitutional Development Framework
This project uses the Plaesy Spec-Kit constitutional development framework.

## Framework Structure
All core files are in `.plaesy/` directory:

### Chatmodes (Role-Based Prompts)
- `.plaesy/chatmodes/dev.chatmode.md` - Full Stack Developer (@dev)
- `.plaesy/chatmodes/sa.chatmode.md` - Solution Architect (@sa)  
- `.plaesy/chatmodes/qa.chatmode.md` - Quality Assurance (@qa)
- `.plaesy/chatmodes/ba.chatmode.md` - Business Analyst (@ba)
- `.plaesy/chatmodes/devops.chatmode.md` - DevOps Engineer (@devops)
- `.plaesy/chatmodes/security.chatmode.md` - Security Specialist (@security)

### Templates & Instructions
- `.plaesy/templates/spec-template.md` - Specification template
- `.plaesy/templates/plan-template.md` - Planning template
- `.plaesy/templates/tasks-template.md` - Tasks template
- `.plaesy/instructions/tdd-enforcement.instructions.md` - TDD guidelines
- `.plaesy/instructions/security-and-owasp.instructions.md` - Security guidelines

### Quality Gates & Memory
- `.plaesy/checklists/qa.checklist.md` - QA validation
- `.plaesy/checklists/story-done.checklist.md` - Completion checklist
- `.plaesy/memory/constitution.md` - Constitutional framework
- `.plaesy/memory/knowledge-base.md` - Knowledge base

## Usage Patterns
1) **Load relevant chatmode**: Copy content from `.plaesy/chatmodes/[role].chatmode.md`
2) **Use templates**: Reference `.plaesy/templates/` for structured content
3) **Follow instructions**: Apply `.plaesy/instructions/` for technology-specific guidance
4) **Validate with checklists**: Use `.plaesy/checklists/` for quality gates

Constitutional framework loaded from `.plaesy/` directory.
'@
      $qwenInstructions | Out-File -FilePath ".qwen\constitutional-instructions.md" -Encoding UTF8
    }

    "kilo-code" {
      # Kilo Code configuration - using existing framework files  
      # Copy chatmodes (standardized .chatmode.md)
      if (!(Test-Path ".kilo")) {
        New-Item -ItemType Directory -Path ".kilo\plaesy" -Force | Out-Null
      }
      if (Test-Path ".plaesy\chatmodes") {
        Copy-Item -Path ".plaesy\chatmodes\*.chatmode.md" -Destination ".kilo\plaesy\" -Force -ErrorAction SilentlyContinue
      }
            
      $kiloModes = @'
{
  "modes": [
    {
      "name": "plaesy-dev",
      "description": "Plaesy Constitutional Developer",
      "systemPrompt": "You are a constitutional full-stack developer following Plaesy Spec-Kit framework. Load chatmode from .plaesy/chatmodes/dev.chatmode.md and constitutional rules from .plaesy/memory/constitution.md. Follow TDD Red-Green-Refactor cycle strictly.",
      "temperature": 0.7
    },
    {
      "name": "plaesy-architect", 
      "description": "Plaesy Constitutional Architect",
      "systemPrompt": "You are a solution architect following Plaesy constitutional principles. Load chatmode from .plaesy/chatmodes/sa.chatmode.md and reference .plaesy/templates/ for structure. Design secure, scalable architectures with interface contracts.",
      "temperature": 0.8
    },
    {
      "name": "plaesy-qa",
      "description": "Plaesy Constitutional QA",
      "systemPrompt": "You are a QA engineer following Plaesy constitutional standards. Load chatmode from .plaesy/chatmodes/qa.chatmode.md and use .plaesy/checklists/ for validation. Require 90%+ test coverage.",
      "temperature": 0.6
    }
  ]
}
'@
      $kiloModes | Out-File -FilePath ".kilocodemodes" -Encoding UTF8

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
        
    "qoder-ai" {
      # Qoder AI configuration - using existing framework files
      if (!(Test-Path ".qoder\rules")) {
        New-Item -ItemType Directory -Path ".qoder\rules" -Force | Out-Null
      }
            
      # Copy chatmodes (standardized .chatmode.md)
      if (Test-Path ".plaesy\chatmodes") {
        Copy-Item -Path ".plaesy\chatmodes\*.chatmode.md" -Destination ".qoder\rules\" -Force -ErrorAction SilentlyContinue
      }

      if (Test-Path ".plaesy\instructions\plaesy.instructions.md") {
        Copy-Item -Path ".plaesy\instructions\plaesy.instructions.md" -Destination ".qoder\constitutional-instructions.md" -Force
      }
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
      # Use existing plaesy instructions with generic AI adaptations
      if (Test-Path ".plaesy\instructions\plaesy.instructions.md") {
        # Copy and adapt plaesy instructions for generic AI
        Copy-Item -Path ".plaesy\instructions\plaesy.instructions.md" -Destination ".ai-config\instructions.md" -Force

        # Add generic AI header
        $aiHeader = @'
# AI Assistant Instructions for Plaesy Spec-Kit

> **Note**: This uses the same constitutional framework as Plaesy, adapted for generic AI assistants.

'@
        $existingContent = Get-Content ".ai-config\instructions.md" -Raw
        $aiHeader + $existingContent | Out-File -FilePath ".ai-config\instructions.md" -Encoding UTF8
      }
      else {
        Write-Warning "Source file .plaesy\instructions\plaesy.instructions.md not found"
      }
    }
        
    "none" {
      # Manual development setup
      $manualGuide = @'
# Plaesy Spec-Kit Framework Guide

## Constitutional Development
This project uses constitutional development principles from `.plaesy/memory/constitution.md`.

## Development Workflow
1) **Specify** - Use `.plaesy/templates/spec-template.md` for requirements
2) **Plan** - Use `.plaesy/templates/plan-template.md` for technical architecture
3) **Tasks** - Use `.plaesy/templates/tasks-template.md` for implementation breakdown

## Constitutional Principles
- Test-Driven Development mandatory
- Interface contracts required
- Real dependencies for integration tests
- Security-first design patterns

Refer to `.plaesy/memory/constitution.md` for complete constitutional framework.
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

# Plaesy framework (keep template but ignore generated files)
.plaesy/generated/
'@
    $gitignoreContent | Out-File -FilePath ".gitignore" -Encoding ASCII
  }
    
  Write-Success "Project structure created"
}

function Write-Completion {
  param([string]$AIChoice)
    
  Write-Host ""
  Write-Host "Plaesy Spec-Kit initialization completed!" -ForegroundColor Green
  Write-Host ""
  Write-Host "Project initialized in: $(Get-Location)" -ForegroundColor Cyan
  Write-Host "AI Assistant: $AIChoice" -ForegroundColor Cyan
  Write-Host ""
  Write-Host "Framework Structure:" -ForegroundColor Yellow
  Write-Host "• .plaesy/ - Constitutional framework files"
    
  switch ($AIChoice) {
    "copilot" {
      Write-Host "• .github/copilot-instructions.md - Repository-wide instructions"
      Write-Host "• AGENTS.md - Agent instructions"
      Write-Host "• .github/instructions/ - Path-specific instructions"
      Write-Host ""
      Write-Host "GitHub Copilot users:" -ForegroundColor Blue
      Write-Host "• Repository instructions auto-loaded from .github/copilot-instructions.md"
      Write-Host "• Agent mode uses AGENTS.md"
      Write-Host "• Path-specific rules in .github/instructions/"
    }
    "cursor" {
      Write-Host "• .cursor/rules/ - Modern MDC rules"
      Write-Host ""
      Write-Host "Cursor AI users:" -ForegroundColor Blue
      Write-Host "• Rules auto-loaded from .cursor/rules/*.mdc"
      Write-Host "• Use 'New Cursor Rule' command to add more"
      Write-Host "• Scoped rules for different file types"
    }
    "claude" {
      Write-Host "• CLAUDE.md - Project instructions (auto-loaded)"
      Write-Host "• .claude/commands/ - Slash commands"
      Write-Host ""
      Write-Host "Claude Code users:" -ForegroundColor Blue
      Write-Host "• CLAUDE.md auto-loaded on startup"
      Write-Host "• Use slash commands: /dev, /qa, /idea, /specify, etc."
      Write-Host "• Run 'claude /help' to see available commands"
    }
    "windsurf" {
      Write-Host "Windsurf AI users:" -ForegroundColor Blue
      Write-Host "Framework instructions loaded in .windsurf/"
      Write-Host "Constitutional framework with role-based chatmodes"
    }
    "trae-ai" {
      Write-Host "Trae.ai users:" -ForegroundColor Blue
      Write-Host "Multi-agent configuration in .trae/"
      Write-Host "Constitutional workflow with quality gates"
    }
    "qoder-ai" {
      Write-Host "• .qoder/rules/ - Constitutional framework rules"
      Write-Host "• .qoder/constitutional-instructions.md - Framework guide"
      Write-Host ""
      Write-Host "Qoder AI users:" -ForegroundColor Blue
      Write-Host "• Constitutional framework loaded in .qoder/rules/"
      Write-Host "• Use @dev-expert, @qa-expert, @sa-expert, @ba-expert in chat"
      Write-Host "• Constitutional rules apply automatically to all requests"
      Write-Host "• Framework files available in .plaesy/ for reference"
    }
    { $_ -in @("chatgpt", "gemini", "local") } {
      Write-Host "• .ai-config/ - Generic AI configuration"
      Write-Host ""
      Write-Host "AI Assistant users:" -ForegroundColor Blue
      Write-Host "• Follow instructions in .ai-config/setup-instructions.md"
      Write-Host "• Framework rules in .plaesy/memory/constitution.md"
      Write-Host "• Templates in .plaesy/templates/"
    }
    "none" {
      Write-Host ""
      Write-Host "Manual development:" -ForegroundColor Blue
      Write-Host "• Follow FRAMEWORK-GUIDE.md"
      Write-Host "• Use scripts in .plaesy/scripts/"
      Write-Host "• Templates in .plaesy/templates/"
    }
  }
    
  Write-Host ""
  Write-Host "Next Steps:" -ForegroundColor Yellow
  Write-Host "1. Start your idea with \`/idea your idea description\` command"
  Write-Host "2. Use the \`/specify\` command to define your features"
  Write-Host "3. Follow the development workflow: idea → specify → plan → tasks → implement"
  Write-Host ""
  Write-Host "Constitutional Development Framework Active" -ForegroundColor Magenta
  Write-Host "   Quality through discipline. Excellence through automation."
}

# Main execution
Write-Host ""
Write-Host "Plaesy Spec-Kit Initialization" -ForegroundColor Magenta
Write-Host "   Constitutional Development Framework" -ForegroundColor Magenta
Write-Host ""
Write-Host "Initializing constitutional development framework..."
Write-Host ""

# Validate environment first
Test-InitEnvironment

$AIChoice = Select-AIAssistant
Copy-CoreFramework -AIChoice $AIChoice
Set-AISpecificConfig -AIChoice $AIChoice
New-ProjectStructure
Write-Completion -AIChoice $AIChoice