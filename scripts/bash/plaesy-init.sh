#!/bin/bash

# Plaesy Spec-Kit Initialization Script - Clean & Dynamic
# 100% configuration-driven platform support

set -euo pipefail
IFS=$' \n\t'

# Source common functions
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -f "$SCRIPT_DIR/common.sh" ]]; then
    source "$SCRIPT_DIR/common.sh"
else
    echo "ERROR: common.sh not found at $SCRIPT_DIR/common.sh" >&2
    exit 1
fi

# Constants
readonly CONFIG_FILE="$SCRIPT_DIR/../configs/platform.json"
readonly CONFIG_MANAGER="$SCRIPT_DIR/config-manager.sh"
readonly SPEC_KIT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Validate critical files exist
if [[ ! -f "$CONFIG_FILE" ]]; then
    log_error "Configuration file not found: $CONFIG_FILE"
fi

if [[ ! -f "$CONFIG_MANAGER" ]]; then
    log_error "Config manager script not found: $CONFIG_MANAGER"
fi

# ============================================================================
# COPY FUNCTIONS - Copy instructions, scripts, and templates to project
# ============================================================================

# Validate source directory exists and has required files
validate_source_directory() {
    local source_dir="$1"
    local dir_name="$2"

    if [[ ! -d "$source_dir" ]]; then
        log_warning "Source directory not found: $source_dir"
        return 1
    fi

    local file_count
    file_count=$(find "$source_dir" -type f | head -1 | wc -l)
    if [[ $file_count -eq 0 ]]; then
        log_warning "Source directory is empty: $source_dir ($dir_name)"
        return 1
    fi

    return 0
}

# Copy instructions to .plaesy/instructions/
copy_instructions() {
    local target_dir="$1"
    local source_dir="$SPEC_KIT_ROOT/instructions"

    if ! validate_source_directory "$source_dir" "instructions"; then
        log_warning "Skipping instruction copy - source directory issues"
        return 1
    fi

    log_info "Copying instructions to .plaesy/instructions/..."
    mkdir -p "$target_dir/instructions"

    # Get always_load instructions from mapping.json (or use detect-stack)
    local instructions_to_copy=()
    local mapping_file="$SPEC_KIT_ROOT/instructions/mapping.json"

    if [[ "${ALL_INSTRUCTIONS:-false}" == "true" ]]; then
        # -all-instructions: install every instruction file, skip detection
        while IFS= read -r instruction; do
            [[ -n "$instruction" ]] && instructions_to_copy+=("$instruction")
        done < <(find "$source_dir" -maxdepth 1 -name "*.instructions.md" -type f -exec basename {} \; | sort)
    elif [[ -f "$mapping_file" ]]; then
        # Use detect-stack.sh if available (most comprehensive). detect-stack.sh
        # does its own JSON parsing and does not need jq, so its availability
        # must not be gated on jq being installed (jq is only needed by the
        # always_load fallback below).
        if [[ -f "$SPEC_KIT_ROOT/scripts/bash/detect-stack.sh" ]]; then
            # Scan the project root ("."), not $target_dir -- that parameter is
            # the .plaesy base directory instructions are copied INTO (e.g.
            # ".plaesy"), which is empty at this point and never contains the
            # project's own .docx/.xlsx/.pptx/manifest files.
            while IFS= read -r instruction; do
                [[ -n "$instruction" ]] && instructions_to_copy+=("$instruction")
            done < <("$SPEC_KIT_ROOT/scripts/bash/detect-stack.sh" "." 2>/dev/null)
        fi

        # Fallback: read always_load from mapping.json (requires jq)
        if [[ ${#instructions_to_copy[@]} -eq 0 ]] && command -v jq >/dev/null 2>&1; then
            while IFS= read -r instruction; do
                [[ -n "$instruction" ]] && instructions_to_copy+=("$instruction")
            done < <(jq -r '.mappings.always_load[]' "$mapping_file" 2>/dev/null)
        fi
    fi

    if [[ ${#instructions_to_copy[@]} -eq 0 ]]; then
        # Last fallback: nothing matched above (no mapping.json, detect-stack.sh
        # missing/found nothing, and jq unavailable for the always_load read)
        # Must mirror mapping.json's "always_load" array exactly
        log_info "  (using default always-load instructions)"
        instructions_to_copy=(
            "plaesy.instructions.md"
            "plaesy-trim.instructions.md"
            "plaesy-graph.instructions.md"
            "tasks.instructions.md"
            "quality-gates.instructions.md"
            "error-recovery.instructions.md"
            "date-system.instructions.md"
            "assess-technical.instructions.md"
            "assess-design.instructions.md"
            "assess-business.instructions.md"
            "assess-financial.instructions.md"
            "assess-marketing.instructions.md"
            "assess-legal.instructions.md"
            "assess-management.instructions.md"
            "assess-product.instructions.md"
            "dimension-mapping.instructions.md"
            "error-recovery-predictive.instructions.md"
            "output-validation.instructions.md"
            "universal-orchestrator.instructions.md"
        )
    fi

    # Copy detected/always-load instructions, transforming filename
    # e.g., quality-gates.instructions.md → quality-gates.md
    for instruction in "${instructions_to_copy[@]}"; do
        local file="$source_dir/$instruction"
        if [[ -f "$file" ]]; then
            local basename=$(basename "$file" .instructions.md)
            local target_file="$target_dir/instructions/$basename.md"
            if [[ -f "$target_file" ]]; then
                log_warning "  ⊘ $basename.md (already exists, skipping)"
            else
                if cp "$file" "$target_file" 2>/dev/null; then
                    log_info "  ✓ $basename.md"
                else
                    log_warning "  ✗ $basename.md (failed to copy)"
                fi
            fi
        else
            log_warning "  ? $instruction (source not found, skipping)"
        fi
    done

    # Copy memory.md from template - only if not exists
    if [[ -f "$target_dir/memory.md" ]]; then
        log_warning "  ⊘ memory.md (already exists, skipping)"
    else
        local template_file="$SPEC_KIT_ROOT/templates/memory.template.md"

        if [[ -f "$template_file" ]]; then
            cp "$template_file" "$target_dir/memory.md"
            log_info "  ✓ memory.md"
        else
            # Fallback: create empty file if template not found
            log_warning "  ! memory.template.md not found, creating empty file"
            touch "$target_dir/memory.md"
            log_info "  ✓ memory.md (empty)"
        fi
    fi

    log_success "Instructions copied"
}

# Copy templates to .plaesy/templates/ (flat, whole library — referenced by
# update-agent-context.sh, plan.template.md, status.template.md as always-present)
copy_templates() {
    local target_dir="$1"
    local source_dir="$SPEC_KIT_ROOT/templates"

    if ! validate_source_directory "$source_dir" "templates"; then
        log_warning "Skipping template copy - source directory issues"
        return 1
    fi

    log_info "Copying templates to .plaesy/templates/..."
    mkdir -p "$target_dir/templates"

    local copied_count=0
    for file in "$source_dir"/*.md "$source_dir"/*.json "$source_dir"/*.yaml; do
        if [[ -f "$file" ]]; then
            local basename=$(basename "$file")
            local target_file="$target_dir/templates/$basename"
            if [[ -f "$target_file" ]]; then
                log_warning "  ⊘ $basename (already exists, skipping)"
            else
                if cp "$file" "$target_file" 2>/dev/null; then
                    copied_count=$((copied_count + 1))
                else
                    log_warning "  ✗ $basename (failed to copy)"
                fi
            fi
        fi
    done
    [[ $copied_count -gt 0 ]] && log_info "  ✓ Templates copied ($copied_count files)"

    log_success "Templates copied"
}

# Copy checklists to .plaesy/checklists/ (flat, whole library)
copy_checklists() {
    local target_dir="$1"
    local source_dir="$SPEC_KIT_ROOT/checklists"

    if ! validate_source_directory "$source_dir" "checklists"; then
        log_warning "Skipping checklist copy - source directory issues"
        return 1
    fi

    log_info "Copying checklists to .plaesy/checklists/..."
    mkdir -p "$target_dir/checklists"

    local copied_count=0
    for file in "$source_dir"/*.md; do
        if [[ -f "$file" ]]; then
            local basename=$(basename "$file")
            local target_file="$target_dir/checklists/$basename"
            if [[ -f "$target_file" ]]; then
                log_warning "  ⊘ $basename (already exists, skipping)"
            else
                if cp "$file" "$target_file" 2>/dev/null; then
                    copied_count=$((copied_count + 1))
                else
                    log_warning "  ✗ $basename (failed to copy)"
                fi
            fi
        fi
    done
    [[ $copied_count -gt 0 ]] && log_info "  ✓ Checklists copied ($copied_count files)"

    log_success "Checklists copied"
}

# Copy scripts to .plaesy/scripts/
copy_scripts() {
    local target_dir="$1"
    local source_dir="$SPEC_KIT_ROOT/scripts"

    if ! validate_source_directory "$source_dir" "scripts"; then
        log_warning "Skipping script copy - source directory issues"
        return 1
    fi

    log_info "Copying scripts to .plaesy/scripts/..."

    # Detect current platform (bash vs powershell)
    # For bash installation, copy bash scripts and configs
    mkdir -p "$target_dir/scripts/bash"
    mkdir -p "$target_dir/scripts/configs"

    # Copy bash scripts (skip if already exist to preserve user versions)
    if [[ -d "$source_dir/bash" ]]; then
        local copied_count=0
        for file in "$source_dir/bash"/*.sh; do
            if [[ -f "$file" ]]; then
                local basename=$(basename "$file")
                local target_file="$target_dir/scripts/bash/$basename"
                if [[ -f "$target_file" ]]; then
                    log_warning "  ⊘ $basename (already exists, skipping)"
                else
                    if cp "$file" "$target_file" 2>/dev/null; then
                        copied_count=$((copied_count + 1))
                    else
                        log_warning "  ✗ $basename (failed to copy)"
                    fi
                fi
            fi
        done
        [[ $copied_count -gt 0 ]] && log_info "  ✓ Bash scripts copied ($copied_count files)"
    fi

    # Copy config files (skip if already exist)
    if [[ -d "$source_dir/configs" ]]; then
        local config_count=0
        for file in "$source_dir/configs"/*; do
            if [[ -f "$file" ]]; then
                local basename=$(basename "$file")
                local target_file="$target_dir/scripts/configs/$basename"
                if [[ -f "$target_file" ]]; then
                    log_warning "  ⊘ $basename (already exists, skipping)"
                else
                    if cp "$file" "$target_file" 2>/dev/null; then
                        config_count=$((config_count + 1))
                    else
                        log_warning "  ✗ $basename (failed to copy)"
                    fi
                fi
            fi
        done
        [[ $config_count -gt 0 ]] && log_info "  ✓ Configuration files copied ($config_count files)"
    fi

    # Also copy PowerShell scripts for cross-platform projects
    mkdir -p "$target_dir/scripts/powershell"
    if [[ -d "$source_dir/powershell" ]]; then
        for file in "$source_dir/powershell"/*.ps1; do
            if [[ -f "$file" ]]; then
                local basename=$(basename "$file")
                local target_file="$target_dir/scripts/powershell/$basename"
                if [[ -f "$target_file" ]]; then
                    log_warning "  ⊘ $basename (already exists, skipping)"
                else
                    cp "$file" "$target_file"
                fi
            fi
        done
        log_info "  ✓ PowerShell scripts copied"
    fi

    log_success "Scripts copied"
}

# Create task management structure
create_task_structure() {
    local target_dir="$1"

    log_info "Creating task management structure..."

    mkdir -p "$target_dir/tasks/backlog"
    mkdir -p "$target_dir/tasks/todo"
    mkdir -p "$target_dir/tasks/doing"
    mkdir -p "$target_dir/tasks/done"
    mkdir -p "$target_dir/tasks/blocked"

    # Create README for task structure
    cat > "$target_dir/tasks/README.md" << 'EOF'
# Task Management

Tasks are organized by status:

- **backlog/** — Ideas, features, bugs (unscheduled)
- **todo/** — Ready to start (next in queue)
- **doing/** — In active work
- **done/** — Completed and validated
- **blocked/** — Waiting on dependency

## Task Lifecycle

```
backlog → todo → doing → [quality-gates] → done / blocked
```

Move tasks between directories as status changes. Use `/continue` or `/loop` for automated workflow.

**See `.plaesy/instructions/tasks.md` for detailed task management instructions.**
EOF

    log_success "Task structure created"
}

# Create memory and analysis structure
create_memory_structure() {
    local target_dir="$1"

    log_info "Creating memory and analysis directories..."

    mkdir -p "$target_dir/memory"
    # plaesy-analyze.sh/.ps1 write their output to analysis/project.json
    # (ANALYSIS_DIR="$PROJECT_PATH/.plaesy/analysis") — flat, matching the
    # instructions/roles/scripts/tasks convention (not nested under memory/).
    mkdir -p "$target_dir/analysis"

    # Copy context.md from template - only if not exists
    if [[ -f "$target_dir/context.md" ]]; then
        log_warning "  ⊘ context.md (already exists, skipping)"
    else
        local template_file="$SPEC_KIT_ROOT/templates/context.template.md"

        if [[ -f "$template_file" ]]; then
            cp "$template_file" "$target_dir/context.md"
            log_info "  ✓ context.md"
        else
            # Fallback: create empty file if template not found
            log_warning "  ! context.template.md not found, creating empty file"
            touch "$target_dir/context.md"
            log_info "  ✓ context.md (empty)"
        fi
    fi

    log_success "Memory and analysis structure created"
}

# Create loop state for autonomous workflows
create_loop_state() {
    local target_dir="$1"

    log_info "Creating autonomous loop configuration..."

    # Only create if not exists (preserve existing configurations on re-run)
    if [[ -f "$target_dir/state.json" ]]; then
        log_warning "  ⊘ state.json (already exists, skipping)"
        return 0
    fi

    local template_file="$SPEC_KIT_ROOT/templates/state.template.json"
    local timestamp=$(date -u +%Y-%m-%dT%H:%M:%SZ)

    if [[ -f "$template_file" ]]; then
        # Copy from template and replace timestamp placeholder
        sed "s/\[TIMESTAMP\]/$timestamp/g" "$template_file" > "$target_dir/state.json"
        log_info "  ✓ state.json"
    else
        # Fallback: create empty file if template not found
        log_warning "  ! state.template.json not found, creating empty file"
        touch "$target_dir/state.json"
        log_info "  ✓ state.json (empty)"
    fi

    log_success "Loop state created"
}

# Load available platforms dynamically
load_platforms() {
    if [[ -f "$CONFIG_MANAGER" ]]; then
        "$CONFIG_MANAGER" list-platforms 2>/dev/null || _load_platforms_fallback
    else
        _load_platforms_fallback
    fi
}

# Fallback platform list when config-manager unavailable
_load_platforms_fallback() {
    echo "claude_code cursor_ai github_copilot windsurf_ai cline deepseek kilo_code qoder trae_ai continue_dev tabnine codeium codewhisperer studio_bot replit_ghostwriter llama_index ollama lm_studio generic_ai"
}

# Get platform display name dynamically
get_platform_name() {
    local platform="$1"

    if [[ -f "$CONFIG_MANAGER" ]]; then
        "$CONFIG_MANAGER" get-platform-config "$platform" "name" 2>/dev/null || _get_platform_name_fallback "$platform"
    else
        _get_platform_name_fallback "$platform"
    fi
}

# Fallback platform names when config-manager unavailable
_get_platform_name_fallback() {
    local platform="$1"
    case "$platform" in
        claude_code) echo "Claude Code" ;;
        cursor_ai) echo "Cursor AI" ;;
        github_copilot) echo "GitHub Copilot" ;;
        windsurf_ai) echo "Windsurf AI" ;;
        cline) echo "Cline" ;;
        deepseek) echo "DeepSeek" ;;
        kilo_code) echo "Kilo Code" ;;
        qoder) echo "Qoder AI" ;;
        trae_ai) echo "Trae AI" ;;
        continue_dev) echo "Continue.dev" ;;
        tabnine) echo "Tabnine" ;;
        codeium) echo "Codeium" ;;
        codewhisperer) echo "Amazon CodeWhisperer" ;;
        studio_bot) echo "Google Studio Bot" ;;
        replit_ghostwriter) echo "Replit Ghostwriter" ;;
        llama_index) echo "LlamaIndex" ;;
        ollama) echo "Ollama" ;;
        lm_studio) echo "LM Studio" ;;
        generic_ai) echo "Generic AI" ;;
        *) echo "$platform" ;;
    esac
}

# Normalize platform input
normalize_platform() {
    local input="$1"

    # Convert input to lowercase for matching
    local input_lower=$(echo "$input" | tr '[:upper:]' '[:lower:]')

    # Check partial/common name matches FIRST (pure bash, no subprocess).
    # The platform loop below calls get_platform_name per platform (~0.5s
    # each on MSYS), so common names like "claude" or "none" must be
    # resolved here without iterating the platform list.
    case "$input_lower" in
        "claude"|"claude_code"|"anthropic"|"claude-code")
            echo "claude_code"
            return
            ;;
        "cursor"|"cursor_ai"|"cursor-ai")
            echo "cursor_ai"
            return
            ;;
        "github"|"copilot"|"github_copilot"|"github-copilot"|"gh-copilot")
            echo "github_copilot"
            return
            ;;
        "cline")
            echo "cline"
            return
            ;;
        "deepseek"|"deepseek_ai"|"deepseek-ai")
            echo "deepseek"
            return
            ;;
        "kilo"|"kilo_code"|"kilo-code")
            echo "kilo_code"
            return
            ;;
        "qoder"|"qoder_ai"|"qoder-ai")
            echo "qoder"
            return
            ;;
        "trae"|"trae_ai"|"trae-ai")
            echo "trae_ai"
            return
            ;;
        "windsurf"|"windsurf_ai"|"windsurf-ai")
            echo "windsurf_ai"
            return
            ;;
        "continue"|"continue_dev"|"continue-dev"|"continue.dev")
            echo "continue_dev"
            return
            ;;
        "tabnine")
            echo "tabnine"
            return
            ;;
        "codeium")
            echo "codeium"
            return
            ;;
        "codewhisperer"|"code_whisperer"|"amazon_codewhisperer"|"whisperer")
            echo "codewhisperer"
            return
            ;;
        "studio_bot"|"studio-bot"|"google_studio_bot"|"studiobot")
            echo "studio_bot"
            return
            ;;
        "replit"|"ghostwriter"|"replit_ghostwriter"|"replit-ghostwriter")
            echo "replit_ghostwriter"
            return
            ;;
        "llamaindex"|"llama_index"|"llama-index"|"code-llama")
            echo "llama_index"
            return
            ;;
        "ollama")
            echo "ollama"
            return
            ;;
        "lmstudio"|"lm_studio"|"lm-studio")
            echo "lm_studio"
            return
            ;;
        "generic"|"generic_ai"|"generic-ai"|"none"|"manual")
            echo "generic_ai"
            return
            ;;
    esac

    # Exact match against platform IDs / display names (subprocess per call)
    local platforms
    platforms=$(load_platforms)
    for platform in $platforms; do
        if [[ "$input" == "$platform" ]]; then
            echo "$platform"
            return
        fi

        local display_name
        display_name=$(get_platform_name "$platform")
        if [[ "$input" == "$display_name" ]]; then
            echo "$platform"
            return
        fi
    done

    # Return original input if no match found
    echo "$input"
}

# Validate platform exists
validate_platform() {
    local platform="$1"
    local platforms
    platforms=$(load_platforms)
    echo "$platforms" | grep -q "^${platform}$"
}

# Get AI choice from user
get_ai_choice() {
    local platforms
    platforms=$(load_platforms)

    # Check if platforms list is empty
    if [[ -z "$platforms" ]]; then
        log_error "No platforms available to choose from."
        echo "none"
        return
    fi

    # Check if we're in an interactive environment before printing the menu
    # (non-interactive mode is the common CI/automation path; skip all the
    # per-platform subprocess calls that only feed the interactive display).
    if [[ ! -t 0 ]]; then
        local platform_count=$(echo "$platforms" | wc -w)
        if [[ $platform_count -gt 0 ]]; then
            echo "$platforms" | head -1 | tr -d '\n'
        else
            echo "none"
        fi
        return
    fi

    echo "🤖 Available AI Assistants:"
    echo ""

    echo "0. Cancel initialization"
    local index=1
    for platform in $platforms; do
        local display_name
        display_name=$(get_platform_name "$platform")
        echo "$index. $display_name"
        ((index++))
    done

    echo "$index. No AI (Manual development)"
    echo ""

    local platform_count=$(echo "$platforms" | wc -w)
    local max_choice=$((platform_count + 1))

    # Timeout for read to prevent infinite loop
    local timeout=30
    local choice=""

    # Use read with timeout to prevent infinite loop
    if read -t $timeout -p "Choose your AI assistant (0-$max_choice): " choice 2>/dev/null; then
        # User provided input within timeout
        if ! [[ "$choice" =~ ^[0-9]+$ ]]; then
            log_error "Please enter a valid number."
            echo "none"
            return
        fi

        if [[ $choice -lt 0 || $choice -gt $max_choice ]]; then
            log_error "Please enter a number between 0 and $max_choice."
            echo "none"
            return
        fi

        if [[ $choice -eq 0 ]]; then
            echo "cancel"
            return
        fi

        if [[ $choice -eq $max_choice ]]; then
            echo "none"
            return
        fi

        local index=1
        for platform in $platforms; do
            if [[ $index -eq $choice ]]; then
                echo "$platform"
                return
            fi
            ((index++))
        done

        # Fallback if something went wrong
        echo "none"
    else
        # Timeout or no input, use default
        echo "none"
    fi
}

# Setup platform configuration dynamically based on platform.json mapping (like PowerShell)
setup_platform_config() {
    local platform="$1"

    log_info "Setting up configuration for: $(get_platform_name "$platform") based on platform.json mapping..."

    # Get expanded root path (same as create_structure function)
    local expanded_root
    local plaesy_root="${SCRIPT_DIR}/../.."

    # Check for PLAESY_ROOT override
    expanded_root="${PLAESY_ROOT:-$plaesy_root}"

    # Note: Core file (agents.instructions.md → CLAUDE.md/AGENTS.md) is handled by platform-specific loop below
    # Note: instructions are copied by copy_instructions() (called from create_structure), which
    # itself applies selective install via detect-stack.sh / --all-instructions.

    # Copy chatmodes to .plaesy/roles/ (platform-agnostic, not platform-specific)
    # This is universal like instructions → .plaesy/instructions/
    log_info "Copying chatmodes to .plaesy/roles/..."
    local source_mapping
    source_mapping=$("$CONFIG_MANAGER" get-mapping-value "plaesy" "chatmodes" 2>/dev/null)

    if [[ -n "$source_mapping" && "$source_mapping" != "null" ]]; then
        local source_mapping_clean="${source_mapping%\/*}"
        local source_dir="$expanded_root/$source_mapping_clean"
        local roles_dir=".plaesy/roles"

        mkdir -p "$roles_dir"

        if [[ -d "$source_dir" ]]; then
            find "$source_dir" -maxdepth 1 -name "*.chatmode.md" -type f | while read -r file; do
                local filename="${file##*/}"
                filename="${filename%.chatmode.md}"
                local dest_file="$roles_dir/${filename}.md"

                if [[ -f "$dest_file" ]]; then
                    log_warning "  ⊘ ${filename}.md (already exists, skipping)"
                else
                    cp "$file" "$dest_file"
                    log_info "  ✓ ${filename}.md"
                fi
            done
        else
            log_warning "Source chatmodes directory not found: $source_dir"
        fi
    fi

    # Setup platform-specific files based on platform.json mapping
    local platform_mappings="prompts"

    # Also add core mapping if platform has one (for platform-specific core files like Cursor, Copilot)
    local platform_core_mapping
    platform_core_mapping=$("$CONFIG_MANAGER" get-mapping-value "$platform" "core" 2>/dev/null)
    if [[ -n "$platform_core_mapping" && "$platform_core_mapping" != "null" ]]; then
        platform_mappings="core $platform_mappings"
    fi

    for mapping_type in $platform_mappings; do
        local target_path
        target_path=$("$CONFIG_MANAGER" get-mapping-value "$platform" "$mapping_type" 2>/dev/null)

        if [[ -n "$target_path" && "$target_path" != "null" ]]; then
            # Create target directory if needed
            local target_dir
            target_dir=$(dirname "$target_path")
            if [[ "$target_dir" != "." && -n "$target_dir" ]]; then
                mkdir -p "$target_dir"
                log_success "Created directory: $target_dir"
            fi

            # Copy source files based on mapping type
            case "$mapping_type" in
                "core")
                    # Copy platform-specific core file (Cursor, Copilot, etc.)
                    local source_mapping
                    source_mapping=$("$CONFIG_MANAGER" get-mapping-value "plaesy" "core" 2>/dev/null)

                    if [[ -n "$source_mapping" && "$source_mapping" != "null" ]]; then
                        local source_file="$expanded_root/$source_mapping"
                        if [[ -f "$source_file" ]]; then
                            if [[ -f "$target_path" ]]; then
                                log_info "Platform-specific core exists, skipping: $target_path"
                            else
                                cp "$source_file" "$target_path"
                                log_success "Created platform-specific core file: $target_path"
                            fi
                        else
                            log_warning "Source core file not found: $source_file"
                        fi
                    fi
                    ;;

                "prompts")
                    # Copy prompt files: source from plaesy.prompts, destination from platform.prompts
                    local source_mapping
                    source_mapping=$("$CONFIG_MANAGER" get-mapping-value "plaesy" "prompts" 2>/dev/null)

                    if [[ -n "$source_mapping" && "$source_mapping" != "null" ]]; then
                        # Remove wildcard /* from mapping for directory path
                        local source_mapping_clean="${source_mapping%\/*}"
                        local source_dir="$expanded_root/$source_mapping_clean"

                        # Ensure target_path exists as a directory
                        if [[ ! -d "$target_path" ]]; then
                            mkdir -p "$target_path"
                            log_success "Created directory: $target_path"
                        fi

                        if [[ -d "$source_dir" ]]; then
                            find "$source_dir" -name "*.md" -type f | while read -r file; do
                                local filename="${file#"$source_dir"/}"
                                filename="${filename%.md}"
                                local extension=".md"

                                case "$platform" in
                                    "claude_code") extension=".md" ;;
                                    "cursor_ai") extension=".md" ;;
                                    "github_copilot") extension=".prompt.md" ;;
                                    *) extension=".md" ;;
                                esac

                                mkdir -p "$target_path/$(dirname "$filename")"
                                cp "$file" "$target_path/${filename}${extension}"
                                log_success "Copied prompt: $target_path/${filename}${extension}"
                            done
                        else
                            log_warning "Source prompts directory not found: $source_dir"
                        fi
                    fi
                    ;;

            esac
        fi
    done

    log_success "AI-specific configuration completed based on platform.json mapping"
}

create_structure() {
    local ai_platform="$1"

    log_info "Creating Plaesy structure..."

    # Get structure configuration from platform.yaml
    local base_dir
    local core_dirs
    local project_dirs

    # Read structure directly from JSON
    if [[ -f "$CONFIG_FILE" ]]; then
        # Use config-manager for reliable parsing
        if [[ -f "$CONFIG_MANAGER" ]]; then
            # Get plaesy configuration using config-manager
            base_dir=$("$CONFIG_MANAGER" get-plaesy-structure "base_directory" 2>/dev/null || echo ".plaesy")

            # Get core and project directories using config-manager
            core_dirs=$("$CONFIG_MANAGER" get-plaesy-structure "core_directories" 2>/dev/null || echo "memory instructions")
            project_dirs=$("$CONFIG_MANAGER" get-plaesy-structure "project_directories" 2>/dev/null || echo "docs specs")
        else
            # Fallback: Parse plaesy section directly from JSON
            local json_content
            json_content=$(cat "$CONFIG_FILE" 2>/dev/null)

            # Extract plaesy section
            local structure_section
            structure_section=$(echo "$json_content" | sed -n '/"plaesy"[[:space:]]*:[[:space:]]*{/,/}/p' | head -1)

            # Extract values using more robust parsing
            base_dir=$(echo "$structure_section" | grep '"base_directory"' | sed 's/^[[:space:]]*"base_directory"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/')

            # Extract arrays using precise JSON parsing
            # Extract core_directories from the structure section only
            core_dirs=$(echo "$structure_section" | sed -n '/"core_directories"[[:space:]]*:/,/]/p' | sed '1d;$d' | sed 's/^[[:space:]]*"\([^"]*\)",*[[:space:]]*$/\1/' | tr '\n' ' ')
            core_dirs=${core_dirs% } # Remove trailing space

            # Extract project_directories from the structure section only
            project_dirs=$(echo "$structure_section" | sed -n '/"project_directories"[[:space:]]*:/,/]/p' | sed '1d;$d' | sed 's/^[[:space:]]*"\([^"]*\)",*[[:space:]]*$/\1/' | tr '\n' ' ')
            project_dirs=${project_dirs% } # Remove trailing space

            # Fallbacks if parsing fails - match current JSON configuration
            base_dir=${base_dir:-".plaesy"}
            core_dirs=${core_dirs:-"memory instructions"}
            project_dirs=${project_dirs:-"docs specs"}
        fi
    else
        # Default fallback - match current JSON configuration
        base_dir=".plaesy"
        core_dirs="memory instructions"
        project_dirs="docs specs"
    fi

    # Create base directory
    mkdir -p "$base_dir"

    # Create core directories
    for dir in $core_dirs; do
        if [[ -n "$dir" ]]; then
            log_info "Creating core directory: '$base_dir/$dir'"
            mkdir -p "$base_dir/$dir"
        fi
    done

    # Call copy functions to populate project with framework files
    copy_instructions "$base_dir"
    copy_templates "$base_dir"
    copy_checklists "$base_dir"
    copy_scripts "$base_dir"
    create_task_structure "$base_dir"
    create_memory_structure "$base_dir"
    create_loop_state "$base_dir"

    # Create project directories
    for dir in $project_dirs; do
        if [[ -n "$dir" ]]; then
            log_info "Creating project directory: '$dir'"
            mkdir -p "$dir"
        fi
    done

    log_success "Plaesy structure created with dynamic configuration"
}

# Show help
show_help() {
    echo "Plaesy Spec-Kit Initialization Script"
    echo ""
    echo "USAGE:"
    echo "    plaesy-init [DIRECTORY] [OPTIONS]"
    echo ""
    echo "ARGUMENTS:"
    echo "    DIRECTORY       Target directory (default: current directory)"
    echo ""
    echo "OPTIONS:"
    echo "    -h, --help      Show this help message"
    echo "    -v, --version   Show version information"
    echo "    --ai <platform> Specify AI platform (will prompt if not provided)"
    echo "    --target <dir>  Specify target directory (alternative to positional argument)"
    echo "    --all-instructions  Install every instruction file instead of only the ones"
    echo "                        detected as relevant to this project's stack"
    echo ""
    echo "AVAILABLE AI PLATFORMS:"
    echo "    (Dynamically loaded from platform.json)"
    echo ""
    echo "EXAMPLES:"
    echo "    plaesy init                                    # Interactive mode in current directory"
    echo "    plaesy init ./my-project                        # Interactive mode in specified directory"
    echo "    plaesy init . --ai claude_code                   # Use Claude Code in current directory"
    echo "    plaesy init ./my-project --ai claude_code       # Use Claude Code in specified directory"
    echo "    plaesy init --ai claude_code --target ./project # Long form options"
    echo "    plaesy init --ai=claude_code --target=./project # Combined format"
}

# Show version
show_version() {
    echo "Plaesy Spec-Kit v$(get_plaesy_version)"
}

# Parse arguments
parse_args() {
    local ai_choice=""
    local target_dir="."

    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -v|--version)
                show_version
                exit 0
                ;;
            --ai)
                shift
                ai_choice="$1"
                ;;
            --target)
                shift
                target_dir="$1"
                ;;
            --ai=*)
                ai_choice="${1#--ai=}"
                ;;
            --target=*)
                target_dir="${1#--target=}"
                ;;
            --all-instructions)
                # Handled separately in main() since parse_args runs in a subshell
                ;;
            -*)
                log_error "Unknown option: $1"
                echo "Use --help for usage information"
                exit 1
                ;;
            *)
                # First positional argument is target directory
                if [[ -z "$target_dir" || "$target_dir" == "." ]]; then
                    target_dir="$1"
                else
                    log_error "Too many arguments. Use: plaesy-init [directory] [--ai <platform>]"
                    exit 1
                fi
                ;;
        esac
        shift
    done

    echo "$ai_choice|$target_dir"
}

# Validate target directory
validate_target_dir() {
    local target_dir="$1"

    # Check if directory exists
    if [[ ! -d "$target_dir" ]]; then
        log_error "Target directory does not exist: $target_dir"
        echo "💡 Create directory first: mkdir -p '$target_dir'"
        return 1
    fi

    # Check if writable
    if [[ ! -w "$target_dir" ]]; then
        log_error "Target directory is not writable: $target_dir"
        echo "💡 Check permissions: ls -ld '$target_dir'"
        return 1
    fi

    # Check if already has .plaesy (warn but allow)
    if [[ -d "$target_dir/.plaesy" ]]; then
        log_warning "Target directory already has .plaesy/ (will be overwritten)"
    fi

    return 0
}

# Detect platform automatically
detect_platform() {
    if [[ -f "$CONFIG_MANAGER" ]]; then
        local detected
        detected=$("$CONFIG_MANAGER" detect-platform 2>/dev/null || true)
        if [[ -n "$detected" ]]; then
            echo "$detected"
        fi
    fi
    # Return empty if detection fails - caller will handle
}

# Main function
main() {
    # Check for help and version flags first
    case "${1:-}" in
        -h|--help)
            show_help
            exit 0
            ;;
        -v|--version)
            show_version
            exit 0
            ;;
    esac

    # Selective instruction install is the default (like a skills marketplace);
    # --all-instructions opts back into copying every instruction file.
    ALL_INSTRUCTIONS="false"
    for arg in "$@"; do
        [[ "$arg" == "--all-instructions" ]] && ALL_INSTRUCTIONS="true"
    done

    # Parse arguments
    local parsed
    parsed=$(parse_args "$@")

    local ai_choice="${parsed%%|*}"
    local target_dir="${parsed##*|}"

    # Convert to absolute path
    target_dir="$(cd "$(dirname "$target_dir")" && pwd)/$(basename "$target_dir")"

    # Show banner
    print_banner "Plaesy Spec-Kit Initialization" "Constitutional Development Framework v$(get_plaesy_version)"

    # Validate target directory
    if ! validate_target_dir "$target_dir"; then
        exit 1
    fi

    log_info "Target directory: $target_dir"

    # Change to target directory
    cd "$target_dir" || {
        log_error "Cannot change to directory: $target_dir"
        exit 1
    }

    # Get AI choice
    if [[ -z "$ai_choice" ]]; then
        echo "🔍 Detecting AI platform..."
        local detected
        detected=$(detect_platform)

        if [[ -n "$detected" ]]; then
            echo "✅ Detected: $(get_platform_name "$detected")"

            # Check if we're in interactive mode
            if [[ -t 0 ]]; then
                read -p "Use detected platform? (Y/n): " use_detected
                if [[ "$use_detected" != "n" && "$use_detected" != "N" ]]; then
                    ai_choice="$detected"
                fi
            else
                # Non-interactive mode, use detected platform automatically
                echo "🚀 Auto-using detected platform (non-interactive mode)"
                ai_choice="$detected"
            fi
        fi

        if [[ -z "$ai_choice" ]]; then
            # Always show platform selection when no AI detected
            echo "🤖 No AI platform detected. Please choose from available platforms:"
            echo ""

            local platforms
            platforms=$(load_platforms)
            local index=1

            echo "0. Cancel initialization"
            for platform in $platforms; do
                local display_name
                display_name=$(get_platform_name "$platform")
                echo "$index. $display_name"
                ((index++))
            done

            echo "$index. No AI (Manual development)"
            echo ""

            # Read choice from user (with timeout to prevent infinite loop)
            local platform_count=$(echo "$platforms" | wc -w)
            local max_choice=$((platform_count + 1))
            local timeout=60
            local choice=""

            echo "Choose your AI assistant (0-$max_choice): "
            if read -t $timeout choice 2>/dev/null; then
                if [[ "$choice" =~ ^[0-9]+$ ]] && [[ $choice -ge 0 ]] && [[ $choice -le $max_choice ]]; then
                    if [[ $choice -eq 0 ]]; then
                        echo "❌ Initialization cancelled by user."
                        exit 0
                    elif [[ $choice -eq $max_choice ]]; then
                        ai_choice="none"
                        echo "Selected: Manual development mode"
                    else
                        local current_index=1
                        for platform in $platforms; do
                            if [[ $current_index -eq $choice ]]; then
                                ai_choice="$platform"
                                echo "Selected: $(get_platform_name "$platform")"
                                break
                            fi
                            ((current_index++))
                        done
                    fi
                else
                    echo "❌ Invalid choice. Using manual development mode."
                    ai_choice="none"
                fi
            else
                echo "⏰ Timeout. Using manual development mode."
                ai_choice="none"
            fi
        fi
    else
        # Normalize provided AI choice
        ai_choice=$(normalize_platform "$ai_choice")
        if ! validate_platform "$ai_choice" && [[ "$ai_choice" != "none" ]]; then
            log_error "Invalid AI platform: $ai_choice"
            echo "Available platforms:"
            load_platforms | while read -r platform; do
                echo "  - $(get_platform_name "$platform")"
            done
            exit 1
        fi
    fi

    # Check if user cancelled
    if [[ "$ai_choice" == "cancel" ]]; then
        echo ""
        echo -e "${YELLOW}❌ Initialization cancelled by user.${NC}"
        exit 0
    fi

    log_success "Selected: $(get_platform_name "$ai_choice")"

    # Create structure
    create_structure "$ai_choice"

    # Setup platform configuration
    if [[ "$ai_choice" != "none" ]]; then
        setup_platform_config "$ai_choice"
    fi

    # Success message with summary
    echo ""
    echo -e "${GREEN}🎉 Plaesy Spec-Kit initialization completed!${NC}"
    echo ""

    # Display summary
    echo -e "${CYAN}📊 Initialization Summary:${NC}"
    echo "  Project: $(pwd)"
    echo "  Platform: $(get_platform_name "$ai_choice")"
    echo "  Structure: .plaesy/ with memory, instructions, templates, checklists, scripts, configs, tasks"

    # Show created structure
    if [[ -d ".plaesy" ]]; then
        local file_count
        file_count=$(find .plaesy -type f 2>/dev/null | wc -l)
        echo "  Files created: $file_count"
    fi

    echo ""
    echo -e "${YELLOW}📚 Next Steps:${NC}"
    echo "1. Start: /start <description>"
    echo "2. Continue: /continue (for automated workflow)"
    echo "3. Implement: /implement (for implementation phase)"
    echo ""
    echo -e "${BLUE}📖 Documentation:${NC}"
    echo "  Instructions: .plaesy/instructions/*.md"
    echo "  Scripts: .plaesy/scripts/{bash,powershell}/"
    echo "  Config: .plaesy/scripts/configs/platform.json"
    echo ""
    echo -e "${PURPLE}🏛️  Constitutional Development Framework Active${NC}"
    echo "   Quality through discipline. Excellence through automation."
}

# Run main function with all arguments
main "$@"