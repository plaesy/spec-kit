#!/bin/bash

# Plaesy Config Manager
# Centralized configuration management using platform.yaml

set -euo pipefail
IFS=$' \n\t'

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PLATFORM_CONFIG="$SCRIPT_DIR/../configs/platform.json"

# Color definitions
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
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

# Simple JSON parser using bash built-ins
parse_json_value() {
    local json="$1"
    local key="$2"

    # Extract value from JSON using bash string manipulation
    # This is a simplified parser for well-formed JSON
    local pattern="\"$key\"[[:space:]]*:[[:space:]]*"
    local match

    # Find the key and extract everything after it
    match=$(echo "$json" | grep -o "\"$key\"[[:space:]]*:[[:space:]]*[^,}]*" | head -1)

    if [[ -n "$match" ]]; then
        # Extract value after colon
        local value=$(echo "$match" | sed 's/^[^:]*:[[:space:]]*//')

        # Remove quotes if it's a string
        if [[ "$value" =~ ^\".*\"$ ]]; then
            value="${value:1:-1}"
        fi

        echo "$value"
        return 0
    fi

    return 1
}

parse_json_array() {
    local json="$1"
    local key="$2"

    # Extract array from JSON
    local pattern="\"$key\"[[:space:]]*:[[:space:]]*\["
    local start_pos

    if [[ "$json" =~ $pattern ]]; then
        start_pos=${BASH_REMATCH[0]}
        # Find the position of the opening bracket
        local bracket_pos=${#start_pos}

        # Extract substring starting from the array
        local array_str="${json:$bracket_pos}"

        # Find the closing bracket
        local depth=0
        local end_pos=0
        for ((i=0; i<${#array_str}; i++)); do
            char="${array_str:$i:1}"
            case "$char" in
                "[") ((depth++)) ;;
                "]")
                    ((depth--))
                    if [[ $depth -eq 0 ]]; then
                        end_pos=$i
                        break
                    fi
                    ;;
            esac
        done

        if [[ $end_pos -gt 0 ]]; then
            local array_content="${array_str:0:$end_pos}"
            # Remove brackets and split by comma
            echo "$array_content" | sed 's/^\[//;s/\]$//' | tr ',' '\n' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//' | sed 's/^"//;s/"$//' | grep -v '^$'
            return 0
        fi
    fi

    return 1
}

# Check dependencies (no external dependencies required)
check_dependencies() {
    return 0
}

# Parse platform detection priority from JSON configuration
get_detection_priority() {
    # Dynamic priority list from JSON configuration
    if [[ -f "$PLATFORM_CONFIG" ]]; then
        list-platforms
    else
        # Return error if config file not found
        return 1
    fi
}

# Detect current AI platform
detect-platform() {
    local platforms
    platforms=$(get_detection_priority)

    for platform in $platforms; do
        if detect_platform_specific "$platform"; then
            echo "$platform"
            return 0
        fi
    done

    # No platform detected - return empty to trigger interactive selection
    return 1
}

# Platform-specific detection logic (dynamic from JSON)
detect_platform_specific() {
    local platform="$1"

    # Read detection patterns from JSON configuration
    if [[ -f "$PLATFORM_CONFIG" ]]; then
        local json_content
        json_content=$(cat "$PLATFORM_CONFIG" 2>/dev/null || return 1)

        # Extract detection array - handle empty and non-empty arrays
        local detection_patterns=""
        local platform_section
        platform_section=$(echo "$json_content" | sed -n "/\"$platform\"[[:space:]]*:[[:space:]]*{/,/}/p")

        # Check if detection array exists and is not empty
        if echo "$platform_section" | grep -q '"detection"[[:space:]]*:[[:space:]]*\[\s*\]'; then
            # Empty detection array - nothing to detect
            detection_patterns=""
        elif echo "$platform_section" | grep -q '"detection"'; then
            # Extract detection patterns (non-empty array)
            detection_patterns=$(echo "$platform_section" | \
                               sed -n '/"detection"[[:space:]]*:/,/]/p' | \
                               sed '1d;$d' | \
                               grep '"[^"]*"' | \
                               sed 's/"//g' | \
                               sed 's/^[[:space:]]*//' | \
                               sed 's/,[[:space:]]*$//' | \
                               grep -v '^$')
        fi

        if [[ -n "$detection_patterns" ]]; then
            # Check each detection pattern
            while read -r pattern; do
                if [[ -n "$pattern" ]]; then
                    # Check if pattern exists as file or directory
                    if [[ -f "$pattern" ]] || [[ -d "$pattern" ]]; then
                        return 0
                    fi
                fi
            done <<< "$detection_patterns"
            return 1
        fi
    fi

    # No fallback hardcoded logic - return error if config not available
    return 1
}

# Parse JSON array to space-separated string
parse_json_array() {
    local json_array="$1"

    # Remove brackets and quotes, then replace commas with spaces
    echo "$json_array" | sed 's/^\[//;s/\]$//' | sed 's/"//g' | tr ',' ' '
}

# Get platform configuration
get-platform-config() {
    local platform="$1"
    local key="$2"

    if [[ ! -f "$PLATFORM_CONFIG" ]]; then
        log_error "Platform configuration file not found: $PLATFORM_CONFIG"
        return 1
    fi

    # Use simpler approach with awk for more reliable parsing
    local result
    if [[ "$key" == *"."* ]]; then
        local parent_key="${key%.*}"
        local child_key="${key#*.}"

        # Use awk to extract nested values
        result=$(awk -v platform="$platform" -v parent="$parent_key" -v child="$child_key" '
        BEGIN {
            in_platform = 0;
            in_parent = 0;
            found = 0;
        }
        /"[^"]*"[[:space:]]*:[[:space:]]*{/ {
            if (index($0, "\"" platform "\"")) {
                in_platform = 1
            }
        }
        in_platform && /"[^"]*"[[:space:]]*:[[:space:]]*{/ {
            if (index($0, "\"" parent "\"")) {
                in_parent = 1
            }
        }
        in_parent && index($0, "\"" child "\"") {
            # Extract value after colon - simpler approach
            line = $0
            # Find the key position
            key_pos = index(line, "\"" child "\"")
            if (key_pos > 0) {
                # Extract everything after the key
                remainder = substr(line, key_pos + length("\"" child "\""))
                # Find the colon
                colon_pos = index(remainder, ":")
                if (colon_pos > 0) {
                    # Extract value after colon
                    value = substr(remainder, colon_pos + 1)
                    # Clean up whitespace, quotes, and trailing comma
                    gsub(/^[[:space:]]+/, "", value)
                    gsub(/,[[:space:]]*$/, "", value)
                    gsub(/^"|"$/, "", value)
                    print value
                    found = 1
                    exit
                }
            }
        }
        in_platform && /^[[:space:]]*}/ && in_parent {
            if (!found) exit
        }
        ' "$PLATFORM_CONFIG")
    else
        # Use pure bash/grep/sed approach (no Python dependency)
        result=$(grep -A 50 "\"$platform\"" "$PLATFORM_CONFIG" | grep -A 10 "\"$key\"" | head -1 | sed 's/.*:[[:space:]]*//' | sed 's/^[[:space:]]*"//;s/",$//' | tr -d ',')
    fi

    # Post-process: if result looks like array, convert to space-separated
    local final_result
    final_result=$(echo "$result" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//' 2>/dev/null || echo "")

    if [[ "$final_result" == \[*\] ]]; then
        parse_json_array "$final_result"
    else
        echo "$final_result"
    fi
}

# Get mapping value with support for new structure (value/excludes)
get-mapping-value() {
    local section="$1"
    local mapping_type="$2"

    if [[ ! -f "$PLATFORM_CONFIG" ]]; then
        log_error "Platform configuration file not found: $PLATFORM_CONFIG"
        return 1
    fi

    # First try to get the value from the new structure
    local mapping_value
    mapping_value=$(awk -v section="$section" -v type="$mapping_type" '
    BEGIN {
        in_section = 0;
        in_mapping = 0;
        found = 0;
    }
    /"[^"]*"[[:space:]]*:[[:space:]]*{/ {
        if (index($0, "\"" section "\"")) {
            in_section = 1
        }
    }
    in_section && /"[^"]*"[[:space:]]*:[[:space:]]*{/ {
        if (index($0, "\"mapping\"")) {
            in_mapping = 1
        }
    }
    in_mapping && index($0, "\"" type "\"") {
        # Look for "value" field in the object
        line = $0
        # This might be the start of the object, need to find the "value" field
        # Store current line and continue
        current_line = line
        while ((getline line) > 0) {
            if (index(line, "\"value\"")) {
                # Extract value
                colon_pos = index(line, ":")
                if (colon_pos > 0) {
                    value = substr(line, colon_pos + 1)
                    gsub(/^[[:space:]]+/, "", value)
                    gsub(/,[[:space:]]*$/, "", value)
                    gsub(/[[:space:]]+$/, "", value)
                    gsub(/^"|"$/, "", value)
                    print value
                    found = 1
                    exit
                }
            }
            if (line ~ /^[[:space:]]*}/) break
        }
    }
    ' "$PLATFORM_CONFIG")

    if [[ -n "$mapping_value" ]]; then
        echo "$mapping_value"
        return 0
    fi

    # Fallback to old structure (direct string value)
    get-platform-config "$section" "mapping.$mapping_type"
}

# Get exclude patterns for mapping
get-mapping-excludes() {
    local section="$1"
    local mapping_type="$2"

    if [[ ! -f "$PLATFORM_CONFIG" ]]; then
        log_error "Platform configuration file not found: $PLATFORM_CONFIG"
        return 1
    fi

    awk -v section="$section" -v type="$mapping_type" '
    BEGIN {
        in_section = 0;
        in_mapping = 0;
        in_type = 0;
        found_excludes = 0;
    }
    /"[^"]*"[[:space:]]*:[[:space:]]*{/ {
        if (index($0, "\"" section "\"")) {
            in_section = 1
        }
    }
    in_section && /"[^"]*"[[:space:]]*:[[:space:]]*{/ {
        if (index($0, "\"mapping\"")) {
            in_mapping = 1
        }
    }
    in_mapping && /"[^"]*"[[:space:]]*:[[:space:]]*{/ {
        if (index($0, "\"" type "\"")) {
            in_type = 1
        }
    }
    in_type && index($0, "\"excludes\"") {
        # Extract excludes array
        in_excludes = 1
        next
    }
    in_excludes && /]/ {
        in_excludes = 0
        in_type = 0
    }
    in_excludes {
        # Extract each exclude pattern
        if (match($0, /"([^"]+)"/, arr)) {
            print arr[1]
        }
    }
    ' "$PLATFORM_CONFIG"
}

# List all available platforms
list-platforms() {
    if [[ ! -f "$PLATFORM_CONFIG" ]]; then
        return 1
    fi

    # Read JSON file
    local json_content
    json_content=$(cat "$PLATFORM_CONFIG" 2>/dev/null || echo "")

    # Extract platforms section and find platform names (fully dynamic with brace counting)
    local platforms_section
    platforms_section=$(echo "$json_content" | awk '
        /"platforms"[[:space:]]*:[[:space:]]*{/ {
            in_platforms = 1
            brace_count = 1
            next
        }
        in_platforms {
            if (/\{/) brace_count++
            if (/\}/) brace_count--

            if (brace_count <= 0) {
                in_platforms = 0
                exit
            }

            if (/^[[:space:]]*"[^"]+"[[:space:]]*:[[:space:]]*\{$/ && !/"mapping"/) {
                gsub(/^[[:space:]]*"/, "")
                gsub(/".*/, "")
                if (length($0) > 0) print
            }
        }
    ')

    if [[ -z "$platforms_section" ]]; then
        return 1
    else
        echo "$platforms_section"
    fi
}

# Get cleanup files for platform
get-clean-files() {
    local platform="${1:-$(detect-platform)}"

    # Fallback hardcoded files (cleanup configuration not in current JSON)
    echo "CLAUDE.md .cursorrules .github/copilot-instructions.md"
}

# Get source file processing rules for plaesy-init
get-source-rules() {
    local file_type="$1"

    # Fallback hardcoded rules (source_files configuration not in current JSON)
    case "$file_type" in
        "core")
            echo "instructions/plaesy.instructions.md"
            ;;
        "instructions")
            echo "instructions/*.instructions.md plaesy.instructions.md"
            ;;
        "prompts")
            echo "prompts/*.prompt.md"
            ;;
        "chatmodes")
            echo "chatmodes/*.chatmode.md"
            ;;
    esac
}

# Get cleanup directories for platform
get-clean-dirs() {
    local platform="${1:-$(detect-platform)}"

    # Fallback hardcoded directories (compute from JSON mapping structure)
    local dirs=""

    # Get core mapping and extract directory
    local core_target
    core_target=$(get-platform-config "$platform" "mapping.core" 2>/dev/null)
    if [[ -n "$core_target" ]]; then
        local core_dir=$(dirname "$core_target")
        if [[ "$core_dir" != "." ]]; then
            dirs="$dirs $core_dir"
        fi
    fi

    # Get prompts mapping and extract directory
    local prompts_target
    prompts_target=$(get-platform-config "$platform" "mapping.prompts" 2>/dev/null)
    if [[ -n "$prompts_target" && "$prompts_target" != "null" ]]; then
        local prompt_dir=$(dirname "$prompts_target")
        if [[ "$prompt_dir" != "." && "$dirs" != *"$prompt_dir"* ]]; then
            dirs="$dirs $prompt_dir"
        fi
    fi

    # Get chatmodes mapping and extract directory
    local chatmodes_target
    chatmodes_target=$(get-platform-config "$platform" "mapping.chatmodes" 2>/dev/null)
    if [[ -n "$chatmodes_target" && "$chatmodes_target" != "null" ]]; then
        local chatmode_dir=$(dirname "$chatmodes_target")
        if [[ "$chatmode_dir" != "." && "$dirs" != *"$chatmode_dir"* ]]; then
            dirs="$dirs $chatmode_dir"
        fi
    fi

    # Default directories if none found - compute from available platforms
    if [[ -z "$dirs" ]]; then
        local platforms
        platforms=$(list-platforms 2>/dev/null)
        if [[ -n "$platforms" ]]; then
            while read -r platform; do
                local core_target
                core_target=$(get-platform-config "$platform" "mapping.core" 2>/dev/null)
                if [[ -n "$core_target" ]]; then
                    local core_dir=$(dirname "$core_target")
                    if [[ "$core_dir" != "." && "$dirs" != *"$core_dir"* ]]; then
                        dirs="$dirs $core_dir"
                    fi
                fi
            done <<< "$platforms"
        fi

        # Return error if no directories found
        if [[ -z "$dirs" ]]; then
            return 1
        else
            echo "$dirs"
        fi
    else
        echo "$dirs"
    fi
}

# Validate configuration
validate() {
    if [[ ! -f "$PLATFORM_CONFIG" ]]; then
        log_error "Platform configuration not found: $PLATFORM_CONFIG"
        return 1
    fi

    # Check if python3 is available for JSON validation
    if ! command -v python3 > /dev/null 2>&1; then
        log_warning "python3 not available - skipping JSON syntax validation"
        return 0
    fi

    # Validate JSON syntax
    if ! python3 -m json.tool "$PLATFORM_CONFIG" > /dev/null 2>&1; then
        log_error "Invalid JSON syntax in platform configuration"
        return 1
    fi

    log_success "Platform configuration is valid"
    return 0
}

# Show platform information
show-platform-info() {
    local platform="${1:-$(detect-platform)}"

    echo "Platform Information for: $platform"
    echo "================================"

    echo "Name: $(get-platform-config "$platform" "name" || echo "Unknown")"
    echo "Provider: $(get-platform-config "$platform" "provider" || echo "Unknown")"
    echo "Category: $(get-platform-config "$platform" "category" || echo "Unknown")"
}

# Get Plaesy structure configuration
get-plaesy-structure() {
    local component="$1"

    if [[ ! -f "$PLATFORM_CONFIG" ]]; then
        log_error "Platform configuration file not found: $PLATFORM_CONFIG"
        return 1
    fi

    # Read JSON file
    local json_content
    json_content=$(cat "$PLATFORM_CONFIG" 2>/dev/null || return 1)

    # Parse plaesy section using pure bash
    local plaesy_section
    if [[ "$component" == "core_directories" || "$component" == "project_directories" || "$component" == "memory_subdirectories" ]]; then
        # Extract array values - more precise parsing
        grep -A 20 '"plaesy"' "$PLATFORM_CONFIG" | grep -A 10 "\"$component\"" | sed '1d' | sed '/]/q' | grep -v '^]' | tr -d '"[]' | tr ',' ' ' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//' | grep -v '^$' | tr '\n' ' '
    else
        # Extract single value
        grep -A 20 '"plaesy"' "$PLATFORM_CONFIG" | grep -A 5 "\"$component\"" | head -1 | sed 's/.*:[[:space:]]*//' | sed 's/^[[:space:]]*"//;s/",$//' | tr -d ','
    fi
}

# Show help
show-help() {
    local first_platform
    first_platform=$(list-platforms 2>/dev/null | head -1)
    local second_platform
    second_platform=$(list-platforms 2>/dev/null | sed -n '2p')

    cat << HELP_EOF
Plaesy Config Manager - Centralized Configuration Management

USAGE:
    config-manager.sh <command> [options]

COMMANDS:
    detect-platform              Detect current AI platform
    list-platforms               List all available platforms
    get-platform-config <platform> <key>  Get specific platform configuration
    get-clean-files [platform]   Get files to clean for platform
    get-clean-dirs [platform]    Get directories to clean for platform
    get-plaesy-structure <component> Get Plaesy structure configuration
    show-platform-info [platform] Show detailed platform information
    validate                     Validate platform configuration
    help                         Show this help message

EXAMPLES:
    config-manager.sh detect-platform
    config-manager.sh list-platforms
    config-manager.sh get-platform-config <platform> <key>
    config-manager.sh get-clean-files [platform]
    config-manager.sh show-platform-info [platform]

DEPENDENCIES:
    None required - uses built-in bash functionality

CONFIGURATION:
    Platform configuration: scripts/configs/platform.json (JSON format)
HELP_EOF
}

# Main execution
main() {
    case "${1:-help}" in
        "detect-platform")
            detect-platform
            ;;
        "list-platforms")
            list-platforms
            ;;
        "get-platform-config")
            if [[ $# -lt 3 ]]; then
                log_error "Usage: get-platform-config <platform> <key>"
                exit 1
            fi
            get-platform-config "$2" "$3"
            ;;
        "get-mapping-value")
            if [[ $# -lt 3 ]]; then
                log_error "Usage: get-mapping-value <section> <mapping_type>"
                exit 1
            fi
            get-mapping-value "$2" "$3"
            ;;
        "get-mapping-excludes")
            if [[ $# -lt 3 ]]; then
                log_error "Usage: get-mapping-excludes <section> <mapping_type>"
                exit 1
            fi
            get-mapping-excludes "$2" "$3"
            ;;
        "get-clean-files")
            get-clean-files "${2:-}"
            ;;
        "get-clean-dirs")
            get-clean-dirs "${2:-}"
            ;;
        "get-plaesy-structure")
            if [[ $# -lt 2 ]]; then
                log_error "Usage: get-plaesy-structure <component>"
                exit 1
            fi
            get-plaesy-structure "$2"
            ;;
        "show-platform-info")
            show-platform-info "${2:-}"
            ;;
        "validate")
            validate
            ;;
        "help"|"-h"|"--help"|"")
            show-help
            ;;
        *)
            log_error "Unknown command: $1"
            echo "Use 'help' for usage information"
            exit 1
            ;;
    esac
}

# Run main function
main "$@"