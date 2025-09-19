#!/bin/bash

# Plaesy Clean Script
# Remove directories created by plaesy init

set -uo pipefail  # Exit on undefined vars, pipe failures (removed -e)
IFS=$'\n\t'       # Secure Internal Field Separator

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Global variables
AUTO_CONFIRM=false
TARGET_DIR=""

# Parse arguments
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -y|--yes)
                AUTO_CONFIRM=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            -*)
                echo "Unknown option: $1" >&2
                echo "Run 'plaesy clean --help' for usage information" >&2
                exit 1
                ;;
            *)
                if [[ -z "$TARGET_DIR" ]]; then
                    TARGET_DIR="$1"
                else
                    echo "Error: Multiple directories specified. Only one directory is allowed." >&2
                    exit 1
                fi
                shift
                ;;
        esac
    done
    
    # Set default target directory if not specified
    if [[ -z "$TARGET_DIR" ]]; then
        TARGET_DIR="."
    fi
}

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

# Cleanup function for trap
cleanup() {
    local exit_code=$?
    if [ $exit_code -ne 0 ]; then
        echo -e "${RED}❌ Clean operation failed. Check the error above.${NC}" >&2
    fi
}

# Set up error handling
trap cleanup EXIT

# Validate environment
validate_clean_environment() {
    if [[ ! -d "$TARGET_DIR" ]]; then
        print_error "Target directory '$TARGET_DIR' does not exist."
    fi
    
    if [[ ! -r "$TARGET_DIR" ]]; then
        print_error "Target directory '$TARGET_DIR' is not readable."
    fi
    
    if [[ ! -w "$TARGET_DIR" ]]; then
        print_error "Target directory '$TARGET_DIR' is not writable."
    fi
}

# Show what will be removed
show_cleanup_plan() {
    print_step "Analyzing directories and files to remove..."
    
    local directories_to_remove=(
        ".plaesy"
        ".github/chatmodes"
        ".github/prompts"
        ".cursor"
        ".claude"
        ".windsurf"
        ".trae"
        ".qwen"
        ".kilo"
        ".codex"
        ".opencode"
        ".ai-config"
    )
    
    local files_to_remove=(
        ".github/copilot-instructions.md"
        ".cursor/instructions.md"
        ".claude/instructions.md"
        ".windsurf/instructions.md"
        ".ai-config/instructions.md"
    )
    
    local found_directories=()
    local found_files=()
    
    for dir in "${directories_to_remove[@]}"; do
        if [[ -d "$TARGET_DIR/$dir" ]]; then
            found_directories+=("$dir")
        fi
    done
    
    for file in "${files_to_remove[@]}"; do
        if [[ -f "$TARGET_DIR/$file" ]]; then
            found_files+=("$file")
        fi
    done
    
    if [[ ${#found_directories[@]} -eq 0 && ${#found_files[@]} -eq 0 ]]; then
        print_success "No plaesy directories or files found to remove."
        exit 0
    fi
    
    if [[ ${#found_directories[@]} -gt 0 ]]; then
        echo -e "${CYAN}The following directories will be removed:${NC}"
        for dir in "${found_directories[@]}"; do
            echo "  - $dir"
        done
    fi
    
    if [[ ${#found_files[@]} -gt 0 ]]; then
        echo -e "${CYAN}The following files will be removed:${NC}"
        for file in "${found_files[@]}"; do
            echo "  - $file"
        done
    fi
    echo ""
}

# Confirm removal with user
confirm_removal() {
    if [[ "$AUTO_CONFIRM" == true ]]; then
        echo -e "${YELLOW}⚠️  Auto-confirm mode: proceeding with deletion...${NC}"
        return 0
    fi
    
    local confirmation
    echo -e "${YELLOW}⚠️  This will permanently delete the directories and files listed above.${NC}"
    read -p "Are you sure you want to continue? (y/N): " confirmation
    
    case "$confirmation" in
        [yY]|[yY][eE][sS])
            return 0
            ;;
        *)
            echo "Clean operation cancelled."
            exit 0
            ;;
    esac
}

# Remove plaesy directories
remove_plaesy_directories() {
    print_step "Removing plaesy directories..."
    
    local directories_to_remove=(
        ".plaesy"
        ".github/chatmodes"
        ".github/prompts"
        ".cursor"
        ".claude"
        ".windsurf"
        ".trae"
        ".qwen"
        ".kilo"
        ".codex"
        ".opencode"
        ".ai-config"
    )
    
    local files_to_remove=(
        ".github/copilot-instructions.md"
        ".cursor/instructions.md"
        ".claude/instructions.md"
        ".windsurf/instructions.md"
        ".ai-config/instructions.md"
    )
    
    local removed_count=0
    local failed_removals=()
    
    # Change to target directory
    if ! cd "$TARGET_DIR"; then
        print_error "Failed to change to target directory: $TARGET_DIR"
    fi
    
    for dir in "${directories_to_remove[@]}"; do
        if [[ -d "$dir" ]]; then
            if rm -rf "$dir" 2>/dev/null; then
                echo "  ✅ Removed: $dir"
                ((removed_count++))
            else
                failed_removals+=("$dir")
                echo "  ❌ Failed to remove: $dir"
            fi
        fi
    done
    
    # Remove individual plaesy files
    for file in "${files_to_remove[@]}"; do
        if [[ -f "$file" ]]; then
            if rm -f "$file" 2>/dev/null; then
                echo "  ✅ Removed: $file"
                ((removed_count++))
            else
                failed_removals+=("$file")
                echo "  ❌ Failed to remove: $file"
            fi
        fi
    done
    
    # Clean up empty .github directory if it exists and is empty
    if [[ -d ".github" ]] && [[ -z "$(ls -A .github 2>/dev/null)" ]]; then
        if rmdir ".github" 2>/dev/null; then
            echo "  ✅ Removed empty: .github"
            ((removed_count++))
        fi
    fi
    
    if [[ ${#failed_removals[@]} -gt 0 ]]; then
        print_warning "Failed to remove some items: ${failed_removals[*]}"
    fi
    
    if [[ $removed_count -gt 0 ]]; then
        print_success "Successfully removed $removed_count items"
    else
        print_warning "No items were removed"
    fi
}

# Show help
show_help() {
    cat << 'HELP_EOF'
Usage: plaesy clean [OPTIONS] [TARGET_DIR]

Remove directories created by plaesy init command.

Arguments:
  TARGET_DIR    Directory to clean (default: current directory)

Options:
  -y, --yes     Auto-confirm deletion (skip confirmation prompt)
  -h, --help    Show this help message

Directories that will be removed:
  - .plaesy/                (Core framework files)
  - .github/chatmodes/      (GitHub Copilot chat modes)
  - .github/prompts/        (GitHub Copilot prompts)
  - .cursor/                (Cursor AI configuration)
  - .claude/                (Claude AI configuration)
  - .windsurf/              (Windsurf AI configuration)
  - .trae/                  (Trae.ai configuration)
  - .qwen/                  (Qwen Code configuration)
  - .kilo/                  (Kilo Code configuration)
  - .codex/                 (Codex configuration)
  - .opencode/              (OpenCode configuration)
  - .ai-config/             (Generic AI configuration)

Files that will be removed:
  - .github/copilot-instructions.md  (GitHub Copilot instructions)
  - .cursor/instructions.md           (Cursor AI instructions)
  - .claude/instructions.md           (Claude AI instructions)
  - .windsurf/instructions.md         (Windsurf AI instructions)
  - .ai-config/instructions.md        (Generic AI instructions)

Examples:
  plaesy clean                    # Clean current directory
  plaesy clean /path/to/project   # Clean specific project directory
  plaesy clean -y                 # Clean current directory without confirmation
  plaesy clean -y /path/to/project # Clean specific directory without confirmation

Note: This operation cannot be undone. Use with caution.
HELP_EOF
}

# Main function
main() {
    # Parse command line arguments
    parse_arguments "$@"
    
    print_banner "Plaesy Clean" "Remove plaesy framework directories"
    
    # Validate environment
    validate_clean_environment
    
    # Show what will be cleaned
    show_cleanup_plan
    
    # Confirm with user (or auto-confirm)
    confirm_removal
    
    # Perform cleanup
    remove_plaesy_directories
    
    echo ""
    print_success "🧹 Plaesy clean completed!"
    echo ""
    echo -e "${CYAN}📁 Cleaned directory:${NC} $(cd "$TARGET_DIR" && pwd)"
    echo ""
    echo -e "${YELLOW}💡 To reinitialize the project:${NC}"
    echo "   plaesy-init.sh [--ai=your-choice]"
    echo ""
}

# Run main function
main "$@"