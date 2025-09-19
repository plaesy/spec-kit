#!/bin/bash

# AI-Specific Header Injection Script
# Injects platform-specific headers into prompt files based on chosen AI platform

set -euo pipefail

# Script configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
HEADERS_DIR="$REPO_ROOT/templates/ai-headers"
DEFAULT_HEADER="manual.header.md"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Help function
show_help() {
    cat << EOF
AI Header Injection Script

USAGE:
    $0 --ai <platform> --target <directory> [OPTIONS]

ARGUMENTS:
    --ai <platform>        AI platform to configure headers for
    --target <directory>   Target directory containing prompt files

OPTIONS:
    --dry-run             Show what would be changed without making changes
    --force               Overwrite existing headers without confirmation
    --backup              Create backup of original files
    --help                Show this help message

SUPPORTED AI PLATFORMS:
    copilot              GitHub Copilot
    cursor               Cursor AI
    windsurf             Windsurf AI
    claude               Claude Code
    chatgpt              ChatGPT
    gemini               Google Gemini
    trae-ai              Trae.ai Multi-Agent
    qwen-code            Qwen Code
    codex-cli            Codex CLI
    opencode-cli         OpenCode CLI
    local-ai             Local AI Models (Ollama, LM Studio)
    manual               Manual Development (no AI)

EXAMPLES:
    $0 --ai cursor --target ./prompts
    $0 --ai claude --target . --backup --dry-run
    $0 --ai copilot --target ./my-project --force

EOF
}

# Validate AI platform
validate_ai_platform() {
    local platform="$1"
    local valid_platforms=(
        "copilot" "cursor" "windsurf" "claude" "chatgpt" "gemini"
        "trae-ai" "qwen-code" "codex-cli" "opencode-cli" "local-ai" "manual"
    )

    for valid_platform in "${valid_platforms[@]}"; do
        if [[ "$platform" == "$valid_platform" ]]; then
            return 0
        fi
    done

    return 1
}

# Check if header file exists
get_header_file() {
    local platform="$1"
    local header_file="$HEADERS_DIR/${platform}.header.md"

    if [[ -f "$header_file" ]]; then
        echo "$header_file"
    else
        log_warning "No specific header for $platform, using default header"
        echo "$HEADERS_DIR/$DEFAULT_HEADER"
    fi
}

# Find prompt files
find_prompt_files() {
    local target_dir="$1"

    if [[ ! -d "$target_dir" ]]; then
        log_error "Target directory does not exist: $target_dir"
        return 1
    fi

    find "$target_dir" -name "*.prompt.md" -type f
}

# Check if file already has header
has_header() {
    local file="$1"

    # Check for common header indicators
    if head -20 "$file" | grep -qE "(^---$|Constitutional.*Framework|AI.*Configuration|# .*Header)" ; then
        return 0
    fi

    return 1
}

# Create backup of file
create_backup() {
    local file="$1"
    local backup_file="${file}.backup.$(date +%Y%m%d_%H%M%S)"

    cp "$file" "$backup_file"
    log_info "Created backup: $backup_file"
}

# Remove existing header from file
remove_existing_header() {
    local file="$1"
    local temp_file="${file}.tmp"

    # Find the end of the header (second occurrence of ---)
    local header_end
    header_end=$(grep -n "^---$" "$file" | sed -n '2p' | cut -d: -f1)

    if [[ -n "$header_end" ]]; then
        # Skip header and keep content after it
        tail -n "+$((header_end + 2))" "$file" > "$temp_file"
        mv "$temp_file" "$file"
        log_info "Removed existing header from $(basename "$file")"
    fi
}

# Inject header into file
inject_header() {
    local file="$1"
    local header_content="$2"
    local temp_file="${file}.tmp"

    # Create new file with header + original content
    {
        echo "$header_content"
        echo ""
        echo ""
        cat "$file"
    } > "$temp_file"

    mv "$temp_file" "$file"
}

# Process single file
process_file() {
    local file="$1"
    local header_content="$2"
    local force="$3"
    local backup="$4"
    local dry_run="$5"

    local filename
    filename=$(basename "$file")

    if [[ "$dry_run" == "true" ]]; then
        if has_header "$file" && [[ "$force" != "true" ]]; then
            log_info "[DRY RUN] Would skip $filename (already has header, use --force to overwrite)"
        else
            log_info "[DRY RUN] Would inject header into $filename"
        fi
        return 0
    fi

    # Check if file already has header
    if has_header "$file" && [[ "$force" != "true" ]]; then
        log_warning "Skipping $filename (already has header, use --force to overwrite)"
        return 0
    fi

    # Create backup if requested
    if [[ "$backup" == "true" ]]; then
        create_backup "$file"
    fi

    # Remove existing header if forcing update
    if [[ "$force" == "true" ]] && has_header "$file"; then
        remove_existing_header "$file"
    fi

    # Inject new header
    inject_header "$file" "$header_content"
    log_success "Injected header into $filename"
}

# Main function
main() {
    local ai_platform=""
    local target_dir=""
    local dry_run=false
    local force=false
    local backup=false

    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --ai)
                ai_platform="$2"
                shift 2
                ;;
            --target)
                target_dir="$2"
                shift 2
                ;;
            --dry-run)
                dry_run=true
                shift
                ;;
            --force)
                force=true
                shift
                ;;
            --backup)
                backup=true
                shift
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done

    # Validate required arguments
    if [[ -z "$ai_platform" ]]; then
        log_error "AI platform is required. Use --ai <platform>"
        show_help
        exit 1
    fi

    if [[ -z "$target_dir" ]]; then
        log_error "Target directory is required. Use --target <directory>"
        show_help
        exit 1
    fi

    # Validate AI platform
    if ! validate_ai_platform "$ai_platform"; then
        log_error "Invalid AI platform: $ai_platform"
        show_help
        exit 1
    fi

    log_info "Starting header injection for AI platform: $ai_platform"

    # Get header file
    local header_file
    header_file=$(get_header_file "$ai_platform")

    if [[ ! -f "$header_file" ]]; then
        log_error "Header file not found: $header_file"
        exit 1
    fi

    # Load header content
    local header_content
    header_content=$(cat "$header_file")

    log_info "Using header file: $(basename "$header_file")"

    # Find prompt files
    local prompt_files
    mapfile -t prompt_files < <(find_prompt_files "$target_dir")

    if [[ ${#prompt_files[@]} -eq 0 ]]; then
        log_warning "No prompt files found in $target_dir"
        exit 0
    fi

    log_info "Found ${#prompt_files[@]} prompt file(s)"

    # Process each file
    local processed=0
    local skipped=0

    for file in "${prompt_files[@]}"; do
        if process_file "$file" "$header_content" "$force" "$backup" "$dry_run"; then
            ((processed++))
        else
            ((skipped++))
        fi
    done

    # Summary
    echo ""
    if [[ "$dry_run" == "true" ]]; then
        log_info "DRY RUN COMPLETE"
        log_info "Would process: $processed files"
        log_info "Would skip: $skipped files"
    else
        log_success "HEADER INJECTION COMPLETE"
        log_success "Processed: $processed files"
        if [[ $skipped -gt 0 ]]; then
            log_info "Skipped: $skipped files"
        fi
    fi

    # Create configuration file
    if [[ "$dry_run" != "true" ]] && [[ $processed -gt 0 ]]; then
        local config_file="$target_dir/.plaesy-headers.json"
        cat > "$config_file" << EOF
{
  "ai_platform": "$ai_platform",
  "injection_timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "processed_files": $processed,
  "header_file": "$(basename "$header_file")",
  "constitutional_framework_version": "3.0.0"
}
EOF
        log_info "Created configuration: $(basename "$config_file")"
    fi
}

# Run main function with all arguments
main "$@"