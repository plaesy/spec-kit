#!/bin/bash

# Plaesy Project Initialization Script
# Constitutional Development Framework

set -euo pipefail  # Exit on error, undefined vars, pipe failures
IFS=$'\n\t'       # Secure Internal Field Separator

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PLAESY_HOME="$HOME/.plaesy"

# Source common functions and constants
# shellcheck source=common.sh
if [ -f "$SCRIPT_DIR/../common.sh" ]; then
    source "$SCRIPT_DIR/common.sh"
elif [ -f "$SCRIPT_DIR/common.sh" ]; then
    source "$SCRIPT_DIR/common.sh"
else
    echo "❌ Cannot find common.sh script"
    exit 1
fi

readonly TARGET_DIR="${1:-.}"

# Cleanup function for trap
cleanup() {
    local exit_code=$?
    if [ $exit_code -ne 0 ]; then
        echo -e "${RED}❌ Initialization failed. Check the error above.${NC}" >&2
    fi
}

# Set up error handling
trap cleanup EXIT

# Validate environment
validate_init_environment() {
    if [[ ! -d "$PLAESY_HOME" ]]; then
        print_error "Plaesy framework not found. Please run the installer first."
    fi
    
    if [[ ! -r "$TARGET_DIR" ]]; then
        print_error "Target directory '$TARGET_DIR' is not readable."
    fi
    
    if [[ ! -w "$TARGET_DIR" ]]; then
        print_error "Target directory '$TARGET_DIR' is not writable."
    fi
}

select_ai_assistant() {
    echo -e "${PURPLE}🤖 Select your AI Assistant:${NC}"
    echo "1. GitHub Copilot"
    echo "2. Cursor AI"
    echo "3. Windsurf AI"
    echo "4. Claude Code (Anthropic)"
    echo "5. Trae.ai"
    echo "6. ChatGPT (OpenAI)"
    echo "7. Gemini (Google)"
    echo "8. Qwen Code"
    echo "9. Kilo Code"
    echo "10. Codex CLI"
    echo "11. Codex Web"
    echo "12. OpenCode CLI"
    echo "13. Qoder AI"
    echo "14. Local AI (Ollama/LM Studio)"
    echo "15. No AI (Manual development)"
    echo ""
    
    # Check for --ai parameter
    local ai_found=false
    for arg in "$@"; do
        if [ "$ai_found" = true ]; then
            case $arg in
                copilot|github)
                    AI_CHOICE="copilot"
                    return
                    ;;
                cursor)
                    AI_CHOICE="cursor"
                    return
                    ;;
                windsurf)
                    AI_CHOICE="windsurf"
                    return
                    ;;
                claude)
                    AI_CHOICE="claude"
                    return
                    ;;
                trae|trae-ai)
                    AI_CHOICE="trae-ai"
                    return
                    ;;
                chatgpt|openai)
                    AI_CHOICE="chatgpt"
                    return
                    ;;
                gemini|google)
                    AI_CHOICE="gemini"
                    return
                    ;;
                qwen|qwen-code)
                    AI_CHOICE="qwen-code"
                    return
                    ;;
                kilo|kilo-code)
                    AI_CHOICE="kilo-code"
                    return
                    ;;
                codex|codex-cli)
                    AI_CHOICE="codex-cli"
                    return
                    ;;
                codex-web)
                    AI_CHOICE="codex-web"
                    return
                    ;;
                opencode|opencode-cli)
                    AI_CHOICE="opencode-cli"
                    return
                    ;;
                qoder|qoder-ai)
                    AI_CHOICE="qoder-ai"
                    return
                    ;;
                local|ollama)
                    AI_CHOICE="local"
                    return
                    ;;
                none|manual)
                    AI_CHOICE="none"
                    return
                    ;;
            esac
            ai_found=false
        fi
        
        case $arg in
            --ai=copilot|--ai=github)
                AI_CHOICE="copilot"
                return
                ;;
            --ai=cursor)
                AI_CHOICE="cursor"
                return
                ;;
            --ai=windsurf)
                AI_CHOICE="windsurf"
                return
                ;;
            --ai=claude)
                AI_CHOICE="claude"
                return
                ;;
            --ai=trae|--ai=trae-ai)
                AI_CHOICE="trae-ai"
                return
                ;;
            --ai=chatgpt|--ai=openai)
                AI_CHOICE="chatgpt"
                return
                ;;
            --ai=gemini|--ai=google)
                AI_CHOICE="gemini"
                return
                ;;
            --ai=qwen|--ai=qwen-code)
                AI_CHOICE="qwen-code"
                return
                ;;
            --ai=kilo|--ai=kilo-code)
                AI_CHOICE="kilo-code"
                return
                ;;
            --ai=codex|--ai=codex-cli)
                AI_CHOICE="codex-cli"
                return
                ;;
            --ai=codex-web)
                AI_CHOICE="codex-web"
                return
                ;;
            --ai=opencode|--ai=opencode-cli)
                AI_CHOICE="opencode-cli"
                return
                ;;
            --ai=qoder|--ai=qoder-ai)
                AI_CHOICE="qoder-ai"
                return
                ;;
            --ai=local|--ai=ollama)
                AI_CHOICE="local"
                return
                ;;
            --ai=none|--ai=manual)
                AI_CHOICE="none"
                return
                ;;
            --ai)
                ai_found=true
                ;;
        esac
    done
    
    # Interactive selection if no --ai parameter
    while true; do
        read -p "Choose your AI assistant (1-15): " choice
        case $choice in
            1) AI_CHOICE="copilot"; break ;;
            2) AI_CHOICE="cursor"; break ;;
            3) AI_CHOICE="windsurf"; break ;;
            4) AI_CHOICE="claude"; break ;;
            5) AI_CHOICE="trae-ai"; break ;;
            6) AI_CHOICE="chatgpt"; break ;;
            7) AI_CHOICE="gemini"; break ;;
            8) AI_CHOICE="qwen-code"; break ;;
            9) AI_CHOICE="kilo-code"; break ;;
            10) AI_CHOICE="codex-cli"; break ;;
            11) AI_CHOICE="codex-web"; break ;;
            12) AI_CHOICE="opencode-cli"; break ;;
            13) AI_CHOICE="qoder-ai"; break ;;
            14) AI_CHOICE="local"; break ;;
            15) AI_CHOICE="none"; break ;;
            *) echo "Please enter a number between 1-15" ;;
        esac
    done
}

copy_core_framework() {
    local ai_choice="$1"
    print_step "Setting up constitutional framework for $ai_choice..."

    # Create target directory if it doesn't exist
    if ! mkdir -p "$TARGET_DIR"; then
        print_error "Failed to create target directory: $TARGET_DIR"
    fi

    if ! cd "$TARGET_DIR"; then
        print_error "Failed to change to target directory: $TARGET_DIR"
    fi

    # Create .plaesy directory (shared framework files)
    if ! mkdir -p .plaesy; then
        print_error "Failed to create .plaesy directory"
    fi

    # Copy essential framework files to .plaesy with error checking
    local core_dirs=("memory" "templates" "chatmodes" "instructions" "checklists" "prompts")
    local failed_copies=()

    for dir in "${core_dirs[@]}"; do
        if [[ -d "$PLAESY_HOME/$dir" ]]; then
            if ! cp -r "$PLAESY_HOME/$dir" .plaesy/; then
                failed_copies+=("$dir")
            fi
        else
            print_warning "Source directory $PLAESY_HOME/$dir not found"
        fi
    done

    if [[ ${#failed_copies[@]} -gt 0 ]]; then
        print_warning "Failed to copy some directories: ${failed_copies[*]}"
    fi

    # Inject AI-specific headers into prompt files
    inject_ai_headers "$ai_choice"
    
    # Copy scripts based on AI choice
    case "$ai_choice" in
        "copilot"|"cursor"|"windsurf"|"claude"|"trae-ai"|"chatgpt"|"gemini"|"qwen-code"|"kilo-code"|"codex-cli"|"codex-web"|"opencode-cli"|"qoder-ai"|"local")
            # All AI types need these scripts
            mkdir -p .plaesy/scripts
            local script_files=(
                "common.sh"
                "check-task-prerequisites.sh"
                "create-new-feature.sh"
                "create-new-idea.sh"
                "create-new-plan.sh"
                "detect-secrets-local.sh"
                "get-feature-paths.sh"
                "update-agent-context.sh"
                "validate-llms.sh"
            )
            for script in "${script_files[@]}"; do
                if [[ -f "$PLAESY_HOME/scripts/bash/$script" ]]; then
                    mkdir -p .plaesy/scripts/bash
                    cp "$PLAESY_HOME/scripts/bash/$script" .plaesy/scripts/bash/ 2>/dev/null || true
                fi
            done
            ;;
        "none")
            # Manual development needs all scripts for reference
            cp -r "$PLAESY_HOME/scripts" .plaesy/ 2>/dev/null || true
            ;;
    esac
    
    print_success "Core framework files copied for $ai_choice"
}

# Function to inject AI-specific headers into prompt and chatmode files
inject_ai_headers() {
    local ai_platform="$1"
    print_step "Injecting $ai_platform-optimized headers..."

    # Map AI platform to header file
    local header_file
    case "$ai_platform" in
        "copilot") header_file="copilot.header.md" ;;
        "cursor") header_file="cursor.header.md" ;;
        "windsurf") header_file="windsurf.header.md" ;;
        "claude") header_file="claude.header.md" ;;
        "trae-ai") header_file="trae-ai.header.md" ;;
        "chatgpt") header_file="chatgpt.header.md" ;;
        "gemini") header_file="gemini.header.md" ;;
        "qwen-code") header_file="qwen-code.header.md" ;;
        "kilo-code"|"codex-cli"|"codex-web") header_file="codex-cli.header.md" ;;
        "opencode-cli"|"qoder-ai") header_file="opencode-cli.header.md" ;;
        "local") header_file="local-ai.header.md" ;;
        "none") header_file="manual.header.md" ;;
        *) header_file="manual.header.md" ;;
    esac

    # Check if header file exists
    local header_path="$PLAESY_HOME/templates/ai-headers/$header_file"
    if [[ ! -f "$header_path" ]]; then
        print_warning "Header file not found: $header_file, using manual header"
        header_path="$PLAESY_HOME/templates/ai-headers/manual.header.md"
    fi

    if [[ ! -f "$header_path" ]]; then
        print_warning "No header files found, skipping header injection"
        return 0
    fi

    # Load header content
    local header_content
    header_content=$(cat "$header_path") || {
        print_warning "Failed to read header file, skipping injection"
        return 0
    }

    # Find and inject headers into prompt files
    local files_processed=0
    if [ -d ".plaesy/prompts" ]; then
        for file in .plaesy/prompts/*.prompt.md; do
            # Skip if no files match the pattern
            [[ -f "$file" ]] || continue

            # Create temporary file with header + original content
            {
                echo "$header_content"
                echo ""
                echo ""
                cat "$file"
            } > "${file}.tmp" && {
                # Replace original file only if temp file creation succeeded
                mv "${file}.tmp" "$file" && files_processed=$((files_processed + 1))
            }
        done
    fi

    # Find and inject headers into chatmode files
    if [ -d ".plaesy/chatmodes" ]; then
        for file in .plaesy/chatmodes/*.chatmode.md; do
            # Skip if no files match the pattern
            [[ -f "$file" ]] || continue

            # Create temporary file with header + original content
            {
                echo "$header_content"
                echo ""
                echo ""
                cat "$file"
            } > "${file}.tmp" && {
                # Replace original file only if temp file creation succeeded
                mv "${file}.tmp" "$file" && files_processed=$((files_processed + 1))
            }
        done
    fi

    if [[ $files_processed -gt 0 ]]; then
        print_success "Injected $ai_platform headers into $files_processed files"

        # Create configuration file with error handling
        local timestamp
        timestamp=$(date -u +%Y-%m-%dT%H:%M:%SZ) || timestamp="$(date)"

        {
            echo "{"
            echo "  \"ai_platform\": \"$ai_platform\","
            echo "  \"injection_timestamp\": \"$timestamp\","
            echo "  \"processed_files\": $files_processed,"
            echo "  \"header_file\": \"$header_file\","
            echo "  \"constitutional_framework_version\": \"3.0.0\""
            echo "}"
        } > ".plaesy-headers.json" || print_warning "Failed to create configuration file"
    else
        print_warning "No prompt or chatmode files found for header injection"
    fi

    return 0
}

setup_ai_specific_config() {
    print_step "Configuring for $AI_CHOICE..."
    
    case "$AI_CHOICE" in
        "copilot")
            # GitHub Copilot configuration
            mkdir -p ".github/chatmodes"
            mkdir -p ".github/prompts"
            mkdir -p ".github/instructions"

            # Create copilot-instructions.md (repository-wide custom instructions)
            if [ -f ".plaesy/instructions/plaesy.instructions.md" ]; then
                cp ".plaesy/instructions/plaesy.instructions.md" ".github/copilot-instructions.md"
            fi

            # Copy chatmodes as path-specific instructions (standardized .chatmode.md)
            if [ -d ".plaesy/chatmodes" ]; then
                for chatmode in ".plaesy/chatmodes/"*.chatmode.md; do
                    if [ -f "$chatmode" ]; then
                        filename=$(basename "$chatmode" .chatmode.md)
                        cp "$chatmode" ".github/chatmodes/$filename.chatmode.md"
                    fi
                done
            fi

            # Copy prompts as path-specific instructions (standardized .prompt.md)
            if [ -d ".plaesy/prompts" ]; then
                for prompt in ".plaesy/prompts/"*.prompt.md; do
                    if [ -f "$prompt" ]; then
                        filename=$(basename "$prompt" .prompt.md)
                        cp "$prompt" ".github/prompts/$filename.prompt.md"
                    fi
                done
            fi
            ;;
            
        "cursor")
            # Cursor AI configuration (2025 conventions - MDC format only)
            mkdir -p ".cursor/rules"

            # Create main project rules in MDC format
            if [ -f ".plaesy/instructions/plaesy.instructions.md" ]; then
                # Convert to MDC format with frontmatter
                cat > ".cursor/rules/plaesy.mdc" << 'MDC_HEADER'
---
description: "Plaesy Constitutional Framework Rules"
globs: ["**/*"]
alwaysApply: true
---
MDC_HEADER
                cat ".plaesy/instructions/plaesy.instructions.md" >> ".cursor/rules/plaesy.mdc"
            fi

            # Copy chatmodes as scoped MDC rules (standardized .chatmode.md)
            if [ -d ".plaesy/chatmodes" ]; then
                for chatmode in ".plaesy/chatmodes/"*.chatmode.md; do
                    if [ -f "$chatmode" ]; then
                        filename=$(basename "$chatmode" .chatmode.md)
                        # Create MDC rule with scope
                        cat > ".cursor/rules/$filename.mdc" << MDC_ROLE_HEADER
---
description: "$filename role-based development rules"
globs: ["**/*"]
alwaysApply: false
---
MDC_ROLE_HEADER
                        cat "$chatmode" >> ".cursor/rules/$filename.mdc"
                    fi
                done
            fi
            # Note: .cursorrules is deprecated and NOT created
            ;;

        "claude")
            # Claude Code configuration (2025 conventions)
            mkdir -p ".claude/commands"

            # Create CLAUDE.md (main project instructions)
            if [ -f ".plaesy/instructions/plaesy.instructions.md" ]; then
                cp ".plaesy/instructions/plaesy.instructions.md" "CLAUDE.md"
            fi

            # Copy chatmodes as slash commands (standardized .chatmode.md)
            if [ -d ".plaesy/chatmodes" ]; then
                for chatmode in ".plaesy/chatmodes/"*.chatmode.md; do
                    if [ -f "$chatmode" ]; then
                        filename=$(basename "$chatmode")
                        cp "$chatmode" ".claude/commands/$filename"
                    fi
                done
            fi

            # Copy prompts as slash commands (standardized .prompt.md)
            if [ -d ".plaesy/prompts" ]; then
                for prompt in ".plaesy/prompts/"*.prompt.md; do
                    if [ -f "$prompt" ]; then
                        filename=$(basename "$prompt")
                        cp "$prompt" ".claude/commands/$filename"
                    fi
                done
            fi
            ;;
            
        "windsurf")
            # Windsurf AI configuration - use existing copilot instructions
            mkdir -p ".windsurf"
            
            # Copy chatmodes (standardized .chatmode.md)
            if [ -d ".plaesy/chatmodes" ]; then
                mkdir -p ".windsurf/chatmodes"
                cp ".plaesy/chatmodes/"*.chatmode.md ".windsurf/chatmodes/" 2>/dev/null || true
            fi
            
            # Use existing copilot instructions with Windsurf-specific adaptations
            if [ -f ".plaesy/instructions/plaesy.instructions.md" ]; then
                # Copy and adapt copilot instructions for Windsurf
                cp ".plaesy/instructions/plaesy.instructions.md" ".windsurf/instructions.md"
                
                # Add Windsurf-specific header
                sed -i '1i# Windsurf AI Instructions for Plaesy Spec-Kit\n\n> **Note**: This uses the same constitutional framework as GitHub Copilot with Windsurf-specific adaptations.\n' ".windsurf/instructions.md"
            else
                echo "Warning: .plaesy/instructions/plaesy.instructions.md not found"
            fi
            # Copy existing prompts for Windsurf AI
            if [ -d ".plaesy/prompts" ]; then
                mkdir -p ".windsurf/prompts"
                cp ".plaesy/prompts/"*.md ".windsurf/prompts/" 2>/dev/null || true
            fi
            ;;

        "trae-ai")
            # Trae.ai configuration - advanced multi-agent platform
            mkdir -p ".trae/agents/plaesy"
            mkdir -p ".trae/workflows"
            
            # Copy existing prompts for Trae.ai
            if [ -d ".plaesy/prompts" ]; then
                mkdir -p ".trae/prompts"
                cp ".plaesy/prompts/"*.md ".trae/prompts/" 2>/dev/null || true
            fi
            
            # Create trae configuration
            cat > ".trae/config.yaml" << 'TRAE_CONFIG_EOF'
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
TRAE_CONFIG_EOF

            cat > ".trae/constitutional-workflow.yaml" << 'TRAE_WORKFLOW_EOF'
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
TRAE_WORKFLOW_EOF

            # Create Trae.ai agent definitions using existing chatmode files
            # Copy and link to existing framework files
            cp ".plaesy/chatmodes/dev.chatmode.md" ".trae/agents/plaesy/dev.yaml" 2>/dev/null || echo "# Developer agent - see .plaesy/chatmodes/dev.chatmode.md" > ".trae/agents/plaesy/dev.yaml"
            cp ".plaesy/chatmodes/sa.chatmode.md" ".trae/agents/plaesy/architect.yaml" 2>/dev/null || echo "# Solution Architect agent - see .plaesy/chatmodes/sa.chatmode.md" > ".trae/agents/plaesy/architect.yaml"
            cp ".plaesy/chatmodes/qa.chatmode.md" ".trae/agents/plaesy/qa.yaml" 2>/dev/null || echo "# QA agent - see .plaesy/chatmodes/qa.chatmode.md" > ".trae/agents/plaesy/qa.yaml"
            cp ".plaesy/chatmodes/ba.chatmode.md" ".trae/agents/plaesy/business-analyst.yaml" 2>/dev/null || echo "# Business Analyst agent - see .plaesy/chatmodes/ba.chatmode.md" > ".trae/agents/plaesy/business-analyst.yaml"
            
            # Create agent registry pointing to existing files
            cat > ".trae/agents/plaesy/agent-registry.yaml" << 'TRAE_REGISTRY_EOF'
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
TRAE_REGISTRY_EOF
            ;;

        "qwen-code")
            # Qwen Code configuration - using existing framework files
            mkdir -p ".qwen/plaesy-method"
            
            # Copy chatmodes (standardized .chatmode.md)
            if [ -d ".plaesy/chatmodes" ]; then
                cp ".plaesy/chatmodes/"*.chatmode.md ".qwen/plaesy-method/" 2>/dev/null || true
            fi
            
            cat > ".qwen/constitutional-instructions.md" << 'QWEN_INSTRUCTIONS_EOF'
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
QWEN_INSTRUCTIONS_EOF
            ;;

        "kilo-code")
            # Kilo Code configuration - using existing framework files  
            # Copy chatmodes (standardized .chatmode.md)
            mkdir -p ".kilo/plaesy"
            if [ -d ".plaesy/chatmodes" ]; then
                cp ".plaesy/chatmodes/"*.chatmode.md ".kilo/plaesy/" 2>/dev/null || true
            fi
            
            cat > ".kilocodemodes" << 'KILO_MODES_EOF'
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
KILO_MODES_EOF

            mkdir -p ".kilo"
            cat > ".kilo/plaesy-modes.json" << 'KILO_CONFIG_EOF'
{
  "constitutional_framework": true,
  "tdd_enforcement": "mandatory",
  "interface_contracts": "required",
  "security_first": true,
  "coverage_requirement": 90
}
KILO_CONFIG_EOF
            ;;

        "codex-cli")
            # Codex CLI configuration with project memory
            cat > "AGENTS.md" << 'CODEX_AGENTS_EOF'
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
1) **TDD Mandatory**: All code must follow Red-Green-Refactor cycle
2) **Interface Contracts**: Required for all service boundaries
3) **Real Dependencies**: No mocks in integration tests
4) **Security First**: OWASP compliance and security-first design
5) **Quality Gates**: Automated validation through checklists

Reference `.plaesy/memory/constitution.md` for complete framework details.
CODEX_AGENTS_EOF

            mkdir -p ".codex"
            cat > ".codex/plaesy-config.toml" << 'CODEX_CONFIG_EOF'
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
CODEX_CONFIG_EOF
            ;;

        "codex-web")
            # Codex Web configuration - ensure files are committed to git
            cat > "AGENTS.md" << 'CODEX_WEB_AGENTS_EOF'
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
CODEX_WEB_AGENTS_EOF

            # Create .gitignore modification to ensure core is committed
            if [ -f ".gitignore" ]; then
                # Remove any existing .plaesy entries and add comment
                sed -i '/\.plaesy/d' .gitignore
                echo "" >> .gitignore
                echo "# Plaesy Constitutional Framework - Core committed for Codex Web" >> .gitignore
                echo "# .plaesy-core/" >> .gitignore
            fi
            ;;

        "opencode-cli")
            # OpenCode CLI configuration
            cat > "opencode.jsonc" << 'OPENCODE_CONFIG_EOF'
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
OPENCODE_CONFIG_EOF

            mkdir -p ".opencode"
            cp ".plaesy/instructions/plaesy.instructions.md" ".opencode/plaesy-instructions.md"            
            ;;
            
        "qoder-ai")
            # Qoder AI configuration - using existing framework files
            mkdir -p ".qoder/rules"
            
            # Copy chatmodes (standardized .chatmode.md)
            if [ -d ".plaesy/chatmodes" ]; then
                cp ".plaesy/chatmodes/"*.chatmode.md ".qoder/rules/" 2>/dev/null || true
            fi

            cp ".plaesy/instructions/plaesy.instructions.md" ".qoder/constitutional-instructions.md"            
            ;;
            
        "chatgpt"|"gemini"|"local")
            # Generic AI configuration - use existing copilot instructions
            mkdir -p ".ai-config"
            
            # Copy chatmodes (standardized .chatmode.md)
            if [ -d ".plaesy/chatmodes" ]; then
                mkdir -p ".ai-config/chatmodes"
                cp ".plaesy/chatmodes/"*.chatmode.md ".ai-config/chatmodes/" 2>/dev/null || true
            fi
            
            # Use existing copilot instructions with generic AI adaptations
            if [ -f ".plaesy/instructions/plaesy.instructions.md" ]; then
                # Copy and adapt copilot instructions for generic AI
                cp ".plaesy/instructions/plaesy.instructions.md" ".ai-config/instructions.md"
                
                # Add generic AI header
                sed -i '1i# AI Assistant Instructions for Plaesy Spec-Kit\n\n> **Note**: This uses the same constitutional framework as GitHub Copilot, adapted for generic AI assistants.\n' ".ai-config/instructions.md"
            else
                echo "Warning: .plaesy/instructions/plaesy.instructions.md not found"
            fi
            ;;
            
        "none")
            # Manual development setup
            cat > "FRAMEWORK-GUIDE.md" << 'MANUAL_EOF'
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
MANUAL_EOF
            ;;
    esac
    
    print_success "AI-specific configuration completed"
}

create_project_structure() {
    print_step "Creating project structure..."
      
    # Create .gitignore if it doesn't exist
    if [ ! -f ".gitignore" ]; then
        cat > ".gitignore" << 'GITIGNORE_EOF'
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
GITIGNORE_EOF
    fi
    
    print_success "Project structure created"
}

print_completion() {
    echo ""
    echo -e "${GREEN}🎉 Plaesy Spec-Kit initialization completed!${NC}"
    echo ""
    echo -e "${CYAN}📁 Project initialized in:${NC} $(pwd)"
    echo -e "${CYAN}🤖 AI Assistant:${NC} $AI_CHOICE"
    echo ""
    echo -e "${YELLOW}📚 Framework Structure:${NC}"
    echo "• .plaesy/ - Constitutional framework files"
    
    case "$AI_CHOICE" in
        "copilot")
            echo "• .github/copilot-instructions.md - Repository-wide instructions"
            echo "• AGENTS.md - Agent instructions"
            echo "• .github/instructions/ - Path-specific instructions"
            echo ""
            echo -e "${BLUE}GitHub Copilot users:${NC}"
            echo "• Repository instructions auto-loaded from .github/copilot-instructions.md"
            echo "• Agent mode uses AGENTS.md"
            echo "• Path-specific rules in .github/instructions/"
            ;;
        "cursor")
            echo "• .cursor/rules/ - Modern MDC rules"
            echo ""
            echo -e "${BLUE}Cursor AI users:${NC}"
            echo "• Rules auto-loaded from .cursor/rules/*.mdc"
            echo "• Use 'New Cursor Rule' command to add more"
            echo "• Scoped rules for different file types"
            ;;
        "claude")
            echo "• CLAUDE.md - Project instructions (auto-loaded)"
            echo "• .claude/commands/ - Slash commands"
            echo ""
            echo -e "${BLUE}Claude Code users:${NC}"
            echo "• CLAUDE.md auto-loaded on startup"
            echo "• Use slash commands: /dev, /qa, /idea, /specify, etc."
            echo "• Run 'claude /help' to see available commands"
            ;;
        "chatgpt"|"gemini"|"local")
            echo "• .ai-config/ - Generic AI configuration"
            echo ""
            echo -e "${BLUE}AI Assistant users:${NC}"
            echo "• Follow instructions in .ai-config/setup-instructions.md"
            echo "• Framework rules in .plaesy/memory/constitution.md"
            echo "• Templates in .plaesy/templates/"
            ;;
        "none")
            echo ""
            echo -e "${BLUE}Manual development:${NC}"
            echo "• Follow FRAMEWORK-GUIDE.md"
            echo "• Use scripts in .plaesy/scripts/"
            echo "• Templates in .plaesy/templates/"
            ;;
    esac
    
    echo ""
    echo -e "${YELLOW}📚 Next Steps:${NC}"
    echo "1. Start your idea with \`/idea your idea description\` command"
    echo "2. Use the \`/specify\` command to define your features"
    echo "3. Follow the development workflow: idea → specify → plan → tasks → implement"
    echo ""
    echo -e "${PURPLE}🏛️ Constitutional Development Framework Active${NC}"
    echo "   Quality through discipline. Excellence through automation."
}

# Main execution
echo ""
echo -e "${PURPLE}🏛️  Plaesy Spec-Kit Initialization${NC}"
echo -e "${PURPLE}   Constitutional Development Framework${NC}"
echo ""
echo "Initializing constitutional development framework..."
echo ""

validate_init_environment
select_ai_assistant "$@"
copy_core_framework "$AI_CHOICE"
setup_ai_specific_config
create_project_structure
print_completion