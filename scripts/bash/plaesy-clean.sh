#!/bin/bash

# Plaesy Clean Script
# Configurable cleanup using centralized configuration

set -uo pipefail  # Exit on undefined vars, pipe failures (removed -e)
IFS=$'\n\t'       # Secure Internal Field Separator

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly CONFIG_MANAGER="$SCRIPT_DIR/config-manager.sh"

# Global variables
AUTO_CONFIRM=false
TARGET_DIR=""
AI_CHOICE=""
CLEANUP_LEVEL="safe"
DRY_RUN=false
BACKUP=true
VERBOSE=false

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
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --backup)
                BACKUP=true
                shift
                ;;
            --no-backup)
                BACKUP=false
                shift
                ;;
            --level=*)
                CLEANUP_LEVEL="${1#*=}"
                case $CLEANUP_LEVEL in
                    safe|thorough|complete)
                        # Valid cleanup level
                        ;;
                    *)
                        log_error "Invalid cleanup level: $CLEANUP_LEVEL"
                        echo "Available levels: safe, thorough, complete"
                        exit 1
                        ;;
                esac
                shift
                ;;
            --level)
                if [[ -n ${2:-} ]]; then
                    CLEANUP_LEVEL="$2"
                    case $CLEANUP_LEVEL in
                        safe|thorough|complete)
                            # Valid cleanup level
                            ;;
                        *)
                            log_error "Invalid cleanup level: $CLEANUP_LEVEL"
                            echo "Available levels: safe, thorough, complete"
                            exit 1
                            ;;
                    esac
                    shift 2
                else
                    log_error "--level requires a value"
                    exit 1
                fi
                ;;
            --verbose)
                VERBOSE=true
                shift
                ;;
            --ai=*)
                AI_CHOICE="${1#*=}"
                # Validate AI choice using config manager
                if [[ -f "$CONFIG_MANAGER" ]]; then
                    if ! "$CONFIG_MANAGER" detect-platform | grep -q "^${AI_CHOICE}$"; then
                        log_error "Invalid AI assistant: $AI_CHOICE"
                        echo "Available assistants: claude, copilot, cursor, windsurf, generic"
                        exit 1
                    fi
                else
                    # Fallback validation
                    case $AI_CHOICE in
                        claude|cursor|windsurf|copilot|generic)
                            # Valid AI choice
                            ;;
                        *)
                            log_error "Invalid AI assistant: $AI_CHOICE"
                            echo "Available assistants: claude, cursor, windsurf, copilot, generic"
                            exit 1
                            ;;
                    esac
                fi
                shift
                ;;
            --ai)
                if [[ -n ${2:-} ]]; then
                    AI_CHOICE="$2"
                    # Validate AI choice
                    if [[ -f "$CONFIG_MANAGER" ]]; then
                        if ! "$CONFIG_MANAGER" detect-platform | grep -q "^${AI_CHOICE}$"; then
                            log_error "Invalid AI assistant: $AI_CHOICE"
                            echo "Available assistants: claude, cursor, windsurf, copilot, generic"
                            exit 1
                        fi
                    else
                        # Fallback validation
                        case $AI_CHOICE in
                            claude|cursor|windsurf|copilot|generic)
                                # Valid AI choice
                                ;;
                            *)
                                log_error "Invalid AI assistant: $AI_CHOICE"
                                echo "Available assistants: claude, cursor, windsurf, copilot, generic"
                                exit 1
                                ;;
                        esac
                    fi
                    shift 2
                else
                    log_error "--ai requires a value"
                    exit 1
                fi
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
        log_error "Target directory '$TARGET_DIR' does not exist."
    fi
    
    if [[ ! -r "$TARGET_DIR" ]]; then
        log_error "Target directory '$TARGET_DIR' is not readable."
    fi
    
    if [[ ! -w "$TARGET_DIR" ]]; then
        log_error "Target directory '$TARGET_DIR' is not writable."
    fi
}

# Detect all platforms present in the project using platform.json detection patterns
detect_all_platforms() {
    local config_file="$SCRIPT_DIR/../configs/platform.json"

    if [[ ! -f "$config_file" ]]; then
        echo "generic_ai"
        return
    fi

    # Use awk to parse JSON and detect platforms
    # awk emits "platform<TAB>pattern" lines for every detection pattern;
    # existence checks happen in bash with builtins (-f/-d, no fork per
    # pattern). The old version called awk's system("[ -f ... ]") for each
    # of ~49 patterns, ~6s on MSYS/Windows.
    local detected_platforms
    detected_platforms=$(awk '
        BEGIN { in_platforms=0; brace_depth=0; in_detection=0; platform=""; found=0 }

        # Look for the "platforms" section
        /"platforms"[[:space:]]*:[[:space:]]*\{/ {
            in_platforms=1
            brace_depth=1
            next
        }

        # Track brace depth to properly handle nested objects
        in_platforms {
            if (/\{/) brace_depth++
            if (/\}/) brace_depth--

            if (brace_depth <= 0) {
                in_platforms=0
                next
            }
        }

        # Within platforms section, look for platform objects (but skip "mapping")
        in_platforms && brace_depth==2 && /^[[:space:]]*"[^"]+"[[:space:]]*:[[:space:]]*\{/ {
            gsub(/^[[:space:]]*"/, ""); gsub(/".*$/, "")
            if ($0 != "mapping") {
                platform = $0
                in_detection=0
            }
            next
        }

        # Look for detection array within a platform (only if we have a valid platform)
        in_platforms && platform != "" && /"detection"[[:space:]]*:[[:space:]]*\[/ {
            in_detection=1
            next
        }

        # End of detection array
        in_detection && /\][[:space:]]*$/ {
            in_detection=0
            next
        }

        # Emit detection patterns (existence check done in bash)
        in_detection && /^[[:space:]]*"[^"]+"[[:space:]]*[,]?[[:space:]]*$/ {
            gsub(/^[[:space:]]*"/, ""); gsub(/".*$/, "")
            print platform "\t" $0
        }
    ' "$config_file")

    # Check pattern existence in bash (builtins, no fork per pattern)
    local found_platforms=""
    local first=1
    while IFS=$'\t' read -r platform pattern; do
        [[ -n "$platform" && -n "$pattern" ]] || continue
        if [[ -e "$pattern" || -d "$pattern" ]]; then
            if [[ $first -eq 1 ]]; then
                found_platforms="$platform"
                first=0
            else
                found_platforms="$found_platforms
$platform"
            fi
        fi
    done <<< "$detected_platforms"

    # Return results
    if [[ -n "$found_platforms" ]]; then
        echo "$found_platforms"
    else
        echo "generic_ai"
    fi
}

# Show what will be removed using platform.json mapping
show_cleanup_plan() {
    log_info "Analyzing directories and files to remove based on platform.json mapping..."

    # Validate config manager exists
    if [[ ! -f "$CONFIG_MANAGER" ]]; then
        log_error "Configuration manager not found: $CONFIG_MANAGER"
        exit 1
    fi

    # Change to target directory for file checks
    if ! cd "$TARGET_DIR"; then
        log_error "Cannot change to target directory: $TARGET_DIR"
        exit 1
    fi

    # Determine platform(s) to clean
    local detected_platforms
    if [[ -n "$AI_CHOICE" ]]; then
        detected_platforms=("$AI_CHOICE")
    else
        # Auto-detect all platforms present in the project
        detected_platforms=($(detect_all_platforms))

        if [[ ${#detected_platforms[@]} -eq 0 ]]; then
            detected_platforms=("generic")
        fi
    fi

    if [[ "$DRY_RUN" == "true" ]]; then
        echo -e "${BLUE}[DRY RUN MODE]${NC} No files will be actually removed."
        echo ""
    fi

    # Build platform display string
    local platform_display
    if [[ ${#detected_platforms[@]} -eq 1 ]]; then
        platform_display="${detected_platforms[0]}"
    else
        platform_display=$(IFS=", "; echo "${detected_platforms[*]}")
    fi

    echo -e "${CYAN}Cleanup Plan (Level: $CLEANUP_LEVEL, Platforms: $platform_display)${NC}"
    echo "═══════════════════════════════════════════════════════════════"

    # Show cleanup level details
    case "$CLEANUP_LEVEL" in
        "safe")
            echo -e "${YELLOW}Safe cleanup:${NC} Framework files only, preserve user code"
            ;;
        "thorough")
            echo -e "${YELLOW}Thorough cleanup:${NC} Framework + specs, preserve user code"
            ;;
        "complete")
            echo -e "${RED}Complete cleanup:${NC} Everything Plaesy-related (DANGEROUS)"
            ;;
    esac

    # Show Plaesy framework directories
    echo -e "\n${YELLOW}Plaesy Framework Directories:${NC}"
    # Note: docs/ is not removed as it may contain important project documentation
    local plaesy_dirs=(".plaesy" "specs")
    local found_plaesy_dirs=()

    for dir in "${plaesy_dirs[@]}"; do
        if [[ -d "$dir" ]]; then
            echo "  - $dir/"
            found_plaesy_dirs+=("$dir")
        fi
    done

    # Show platform-specific files based on mapping for each detected platform
    local found_any_platform=false
    for detected_platform in "${detected_platforms[@]}"; do
        if [[ "$detected_platform" != "generic" ]]; then
            echo -e "\n${YELLOW}Platform-Specific Files ($detected_platform):${NC}"
            echo "  Based on platform.json mapping for $detected_platform:"

            local mapping_types=("core" "instructions" "prompts" "chatmodes")
            local found_any=false

            for mapping_type in "${mapping_types[@]}"; do
                local target_path
                target_path=$("$CONFIG_MANAGER" get-mapping-value "$detected_platform" "$mapping_type" 2>/dev/null)

                if [[ -n "$target_path" && "$target_path" != "null" ]]; then
                    if [[ -e "$target_path" ]]; then
                        if [[ -d "$target_path" ]]; then
                            echo "  - $target_path/ (directory)"
                        else
                            echo "  - $target_path (file)"
                        fi
                        found_any=true
                        found_any_platform=true
                    fi
                fi
            done

            if [[ "$found_any" == false ]]; then
                echo "  - No platform-specific files found for $detected_platform"
            fi
        fi
    done

    if [[ "$found_any_platform" == false ]]; then
        echo -e "\n${YELLOW}Platform-Specific Files:${NC}"
        echo "  - Generic platform: No specific files to remove"
    fi

    # Show what will be preserved
    echo -e "\n${GREEN}What will be preserved:${NC}"
    echo "  - Your source code (src/, lib/, components/, etc.)"
    echo "  - Git repository (.git/)"
    echo "  - User-generated files not in Plaesy directories"
    echo "  - Configuration files (.env, .gitignore, etc.)"

    if [[ "$BACKUP" == "true" && "$DRY_RUN" != "true" ]]; then
        echo "  - Backup will be created before removal"
    fi

    # Check if anything found to remove
    if [[ ${#found_plaesy_dirs[@]} -eq 0 && "$found_any_platform" == false ]]; then
        log_success "No Plaesy directories or files found to remove."
        exit 0
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

# Remove plaesy directories based on platform.json mapping
remove_plaesy_directories() {
    log_info "Removing plaesy directories based on platform.json mapping..."

    local removed_count=0
    local failed_removals=()

    # Change to target directory
    if ! cd "$TARGET_DIR"; then
        log_error "Failed to change to target directory: $TARGET_DIR"
        return 1
    fi

    # Determine platforms to clean
    local platforms_to_clean
    if [[ -n "$AI_CHOICE" ]]; then
        platforms_to_clean=("$AI_CHOICE")
    else
        # Auto-detect all platforms present in the project
        platforms_to_clean=($(detect_all_platforms))

        if [[ ${#platforms_to_clean[@]} -eq 0 ]]; then
            platforms_to_clean=("generic")
        fi
    fi

    # Always remove Plaesy framework directories
    # Note: docs/ is not removed as it may contain important project documentation
    local plaesy_dirs=(
        ".plaesy"
        "specs"
    )

    log_info "Removing Plaesy framework directories..."
    for dir in "${plaesy_dirs[@]}"; do
        if [[ -d "$dir" ]]; then
            if rm -rf "$dir" 2>/dev/null; then
                echo "  ✅ Removed Plaesy framework: $dir"
                ((removed_count++))
            else
                failed_removals+=("$dir")
                echo "  ❌ Failed to remove: $dir"
            fi
        fi
    done

    # Remove platform-specific files for each detected platform
    for platform_to_clean in "${platforms_to_clean[@]}"; do
        if [[ "$platform_to_clean" != "generic" ]]; then
            log_info "Removing platform-specific files for: $platform_to_clean"

            # Get platform mappings from config manager
            local mapping_types=("core" "instructions" "prompts" "chatmodes")

            for mapping_type in "${mapping_types[@]}"; do
                local target_path
                target_path=$("$CONFIG_MANAGER" get-mapping-value "$platform_to_clean" "$mapping_type" 2>/dev/null)

                if [[ -n "$target_path" && "$target_path" != "null" ]]; then
                    # Check if target is a directory or file
                    if [[ "$target_path" == */ ]]; then
                        # It's a directory
                        local dir_path="${target_path%/}"
                        if [[ -d "$dir_path" ]]; then
                            if rm -rf "$dir_path" 2>/dev/null; then
                                echo "  ✅ Removed platform $platform_to_clean $mapping_type: $dir_path"
                                ((removed_count++))
                            else
                                failed_removals+=("$dir_path")
                                echo "  ❌ Failed to remove: $dir_path"
                            fi
                        fi
                    else
                        # It's a file or directory (need to check)
                        if [[ -f "$target_path" ]]; then
                            if rm -f "$target_path" 2>/dev/null; then
                                echo "  ✅ Removed platform $platform_to_clean $mapping_type: $target_path"
                                ((removed_count++))
                            else
                                failed_removals+=("$target_path")
                                echo "  ❌ Failed to remove: $target_path"
                            fi
                        elif [[ -d "$target_path" ]]; then
                            if rm -rf "$target_path" 2>/dev/null; then
                                echo "  ✅ Removed platform $platform_to_clean $mapping_type: $target_path"
                                ((removed_count++))
                            else
                                failed_removals+=("$target_path")
                                echo "  ❌ Failed to remove: $target_path"
                            fi
                        fi
                    fi
                fi
            done

            # Clean up empty parent directories
            local parent_dirs=()
            for mapping_type in "${mapping_types[@]}"; do
                local target_path
                target_path=$("$CONFIG_MANAGER" get-mapping-value "$platform_to_clean" "$mapping_type" 2>/dev/null)

                if [[ -n "$target_path" && "$target_path" != "null" ]]; then
                    local parent_dir
                    parent_dir=$(dirname "$target_path")
                    if [[ "$parent_dir" != "." ]]; then
                        parent_dirs+=("$parent_dir")
                    fi
                fi
            done

            # Remove duplicate parent directories
            local unique_parents=($(printf "%s\n" "${parent_dirs[@]}" | sort -u))

            for parent_dir in "${unique_parents[@]}"; do
                if [[ -d "$parent_dir" ]] && [[ -z "$(ls -A "$parent_dir" 2>/dev/null)" ]]; then
                    if rmdir "$parent_dir" 2>/dev/null; then
                        echo "  ✅ Removed empty parent directory: $parent_dir"
                        ((removed_count++))
                    fi
                fi
            done
        fi
    done

    if [[ ${#failed_removals[@]} -gt 0 ]]; then
        log_warning "Failed to remove some items: ${failed_removals[*]}"
    fi

    if [[ $removed_count -gt 0 ]]; then
        log_success "Successfully removed $removed_count items"
    else
        log_warning "No items were removed"
    fi
}

# Show help
show_help() {
    cat << 'HELP_EOF'
Usage: plaesy clean [OPTIONS] [TARGET_DIR]

Remove Plaesy Spec-Kit framework files and directories with safety checks.

ARGUMENTS:
  TARGET_DIR    Directory to clean (default: current directory)

CLEANUP LEVELS:
  safe        Remove framework files only, preserve user code (default)
  thorough    Remove framework + specs, preserve user code
  complete    Remove everything Plaesy-related (DANGEROUS)

OPTIONS:
  -y, --yes                 Auto-confirm deletion (skip confirmation prompt)
  -h, --help                Show this help message
  --dry-run                 Show what would be removed without actually removing
  --backup                  Create backup before removal (enabled by default)
  --no-backup               Skip backup creation
  --level <LEVEL>           Specify cleanup level (safe|thorough|complete)
  --ai <PLATFORM>           Specify AI platform to clean (auto-detected if not specified)
  --verbose                 Show detailed progress

AVAILABLE AI PLATFORMS:
  claude      Claude Code configuration
  copilot     GitHub Copilot configuration
  cursor      Cursor AI configuration
  windsurf    Windsurf AI configuration
  generic     Generic AI configuration

CONFIGURATION:
  This script uses centralized configuration from:
  - scripts/configs/framework-config.json
  - scripts/configs/ai-platforms.json
  - scripts/configs/clean-config.json

EXAMPLES:
  plaesy clean                         # Safe cleanup with backup
  plaesy clean --dry-run               # Preview what would be removed
  plaesy clean thorough --force        # Thorough cleanup without confirmation
  plaesy clean --level complete --no-backup  # Complete cleanup without backup
  plaesy clean --ai claude             # Clean Claude Code configuration only
  plaesy clean --dry-run --level thorough  # Preview thorough cleanup

SAFETY FEATURES:
  - Configuration validation before cleanup
  - Dry-run mode to preview changes
  - Automatic backup creation (default)
  - Platform-specific cleanup detection
  - Clear confirmation prompts

Note: This operation removes Plaesy framework files but preserves your source code.
Always use --dry-run first to preview what will be removed.
HELP_EOF
}

# Main function
main() {
    # Parse command line arguments
    parse_arguments "$@"

    # Show configuration if verbose
    if [[ "$VERBOSE" == "true" ]]; then
        echo -e "${BLUE}Configuration:${NC}"
        echo "  Cleanup Level: $CLEANUP_LEVEL"
        echo "  Dry Run: $DRY_RUN"
        echo "  Backup: $BACKUP"
        echo "  Platform: ${AI_CHOICE:-auto-detect}"
        echo "  Target Directory: $TARGET_DIR"
        echo ""
    fi

    # AI platform detection if not specified
    if [[ -z "$AI_CHOICE" ]]; then
        echo "🔍 Detecting AI platforms for cleanup..."

        # Change to target directory for detection
        cd "$TARGET_DIR" || {
            log_error "Cannot change to target directory: $TARGET_DIR"
            exit 1
        }

        # Detect all platforms present in the project
        local detected_platforms
        detected_platforms=($(detect_all_platforms))

        if [[ ${#detected_platforms[@]} -eq 1 ]]; then
            echo "✅ Detected AI platform: ${detected_platforms[0]}"
            echo "🧹 Using platform-adapted cleanup..."
            AI_CHOICE="${detected_platforms[0]}"
        elif [[ ${#detected_platforms[@]} -gt 1 ]]; then
            echo "✅ Detected multiple AI platforms: $(IFS=", "; echo "${detected_platforms[*]}")"
            echo "🧹 Using multi-platform cleanup..."
            AI_CHOICE="${detected_platforms[*]}"
        else
            echo "🤖 No specific AI platform detected. Using generic cleanup..."
            AI_CHOICE="generic"
        fi
    else
        echo "✅ Using specified AI: $AI_CHOICE"
    fi

    # Validate configuration files
    if ! "$CONFIG_MANAGER" validate >/dev/null 2>&1; then
        log_error "Configuration validation failed"
        exit 1
    fi

    print_banner "Plaesy Clean" "Remove plaesy framework directories"

    # Validate environment
    validate_clean_environment

    # Show what will be cleaned
    show_cleanup_plan

    # Confirm with user (or auto-confirm)
    confirm_removal

    # Create backup if needed
    if [[ "$BACKUP" == "true" && "$DRY_RUN" != "true" ]]; then
        create_backup
    fi

    # Perform cleanup
    remove_plaesy_directories

    echo ""
    log_success "🧹 Plaesy clean completed!"

    # Show platform-specific guidance
    if [[ -n "$AI_CHOICE" ]]; then
        echo ""
        echo "🎯 Platform-specific cleanup guidance:"
        case "$AI_CHOICE" in
            "claude")
                echo "💡 For Claude Code: Use /clean command in future"
                ;;
            "cursor")
                echo "💡 For Cursor AI: Use --clean command in future"
                ;;
            "windsurf")
                echo "💡 For Windsurf AI: Use clean command in future"
                ;;
            "copilot")
                echo "💡 For GitHub Copilot: Continue using plaesy clean command"
                ;;
            *)
                echo "💡 For other platforms: Use plaesy clean command"
                ;;
        esac
    fi
    echo ""
    echo -e "${CYAN}📁 Cleaned directory:${NC} $(cd "$TARGET_DIR" && pwd)"
    echo ""
    echo -e "${YELLOW}💡 To reinitialize the project:${NC}"
    echo "   plaesy init [--ai=your-choice]"
    echo ""
}

# Create backup function
create_backup() {
    if [[ "$BACKUP" != "true" ]]; then
        return 0
    fi

    local backup_dir
    backup_dir="$TARGET_DIR/.plaesy-backup-$(date +%Y%m%d-%H%M%S)"

    log_info "Creating backup: $backup_dir"

    mkdir -p "$backup_dir"

    # Backup Plaesy directory if it exists
    if [[ -d "$TARGET_DIR/.plaesy" ]]; then
        cp -r "$TARGET_DIR/.plaesy" "$backup_dir/"
    fi

    # Determine platforms to backup
    local platforms_to_backup
    if [[ -n "$AI_CHOICE" ]]; then
        platforms_to_backup=("$AI_CHOICE")
    else
        # Auto-detect all platforms present in the project
        platforms_to_backup=($(detect_all_platforms))

        if [[ ${#platforms_to_backup[@]} -eq 0 ]]; then
            platforms_to_backup=("generic")
        fi
    fi

    # Backup platform-specific files based on mapping for each detected platform
    for platform in "${platforms_to_backup[@]}"; do
        if [[ "$platform" != "generic" ]]; then
            local mapping_types=("core" "instructions" "prompts" "chatmodes")

            for mapping_type in "${mapping_types[@]}"; do
                local target_path
                target_path=$("$CONFIG_MANAGER" get-mapping-value "$platform" "$mapping_type" 2>/dev/null)

                if [[ -n "$target_path" && "$target_path" != "null" ]]; then
                    if [[ -e "$TARGET_DIR/$target_path" ]]; then
                        if [[ -d "$TARGET_DIR/$target_path" ]]; then
                            cp -r "$TARGET_DIR/$target_path" "$backup_dir/"
                            log_info "Backed up platform $platform directory: $target_path"
                        else
                            cp "$TARGET_DIR/$target_path" "$backup_dir/"
                            log_info "Backed up platform $platform file: $target_path"
                        fi
                    fi
                fi
            done
        fi
    done

    log_success "Backup created: $backup_dir"
}

# Run main function
main "$@"