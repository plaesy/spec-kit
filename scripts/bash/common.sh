#!/usr/bin/env bash
# Enhanced common functions for plaesy scripts
# Version: 1.0.0
# Features: Improved logging, error handling, and debugging

# Script metadata
if [[ -z "${PLAESY_VERSION:-}" ]]; then
    # Get version from VERSION file in repository root
    get_plaesy_version() {
        local version_file=""

        # Determine script directory
        local script_dir=""
        if [[ -n "${SCRIPT_DIR:-}" ]]; then
            script_dir="$SCRIPT_DIR"
        elif [[ -n "${BASH_SOURCE[1]:-}" ]]; then
            script_dir="$(cd "$(dirname "${BASH_SOURCE[1]}")" && pwd)"
        else
            script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
        fi

        # Try to find VERSION file in multiple locations
        local version_locations=(
            "./VERSION"
            "$script_dir/../../VERSION"
            "$(cd "$script_dir/../.." 2>/dev/null && pwd)/VERSION"
            "${HOME}/.plaesy/VERSION"
        )

        for location in "${version_locations[@]}"; do
            if [[ -f "$location" ]]; then
                version_file="$location"
                break
            fi
        done

        if [[ -z "$version_file" ]]; then
            echo "0.0.0" # Fallback version
            return
        fi

        # Read and sanitize version
        local version=$(cat "$version_file" 2>/dev/null | head -1 | tr -d '[:space:]')
        if [[ -z "$version" || ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
            echo "0.0.0" # Fallback for invalid format
        else
            echo "$version"
        fi
    }

    readonly PLAESY_VERSION=$(get_plaesy_version)
fi
if [[ -z "${SCRIPT_START_TIME:-}" ]]; then
    readonly SCRIPT_START_TIME=$(date +%s)
fi

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly PURPLE='\033[0;35m'
readonly CYAN='\033[0;36m'
readonly WHITE='\033[1;37m'
readonly GRAY='\033[0;37m'
readonly NC='\033[0m' # No Color

# Logging configuration
readonly LOG_LEVEL="${PLAESY_LOG_LEVEL:-INFO}"
readonly LOG_FILE="${PLAESY_LOG_FILE:-}"
readonly DEBUG_MODE="${PLAESY_DEBUG:-false}"

# Enhanced logging functions with timestamps and levels
# Use bash builtin `printf %()T` for the timestamp (no `date` fork) and
# `${0##*/}` instead of `basename` (no fork). Each fork costs ~50-100ms on
# MSYS/Windows, and scripts emit many log lines.
LOG_SCRIPT_NAME_CACHE="${0##*/}"
log_message() {
    local level="$1"
    local message="$2"
    local color="$3"
    local timestamp
    printf -v timestamp '%(%Y-%m-%d %H:%M:%S)T' -1
    local script_name="$LOG_SCRIPT_NAME_CACHE"

    # Console output with colors
    echo -e "${color}[${timestamp}] [${level}] [${script_name}]${NC} $message"

    # File logging (if enabled)
    if [[ -n "$LOG_FILE" ]]; then
        mkdir -p "$(dirname "$LOG_FILE")"
        echo "[${timestamp}] [${level}] [${script_name}] $message" >> "$LOG_FILE"
    fi
}

log_debug() {
    # Ensure DEBUG_MODE is accessible even in error scenarios
    local debug_mode="${DEBUG_MODE:-false}"
    [[ "$debug_mode" == "true" ]] && log_message "DEBUG" "$1" "$GRAY"
}

log_info() {
    log_message "INFO" "$1" "$BLUE"
}

log_success() {
    log_message "SUCCESS" "$1" "$GREEN"
}

log_warning() {
    log_message "WARNING" "$1" "$YELLOW" >&2
}

log_error() {
    log_message "ERROR" "$1" "$RED" >&2
}

# Enhanced error handling with stack traces
handle_error() {
    local exit_code=$?
    local line_number=$1
    local bash_lineno=$2
    local last_command=$3

    # Store DEBUG_MODE locally to avoid issues in error scenarios
    local debug_mode="${DEBUG_MODE:-false}"

    log_error "Script failed with exit code ${exit_code} at line ${line_number}"
    log_error "Failed command: ${last_command}"

    # Show stack trace if debug mode is enabled
    if [[ "$debug_mode" == "true" ]]; then
        log_error "Call stack:"
        local frame=0
        while caller $frame; do
            ((frame++))
        done
    fi

    cleanup_on_error
    exit $exit_code
}

# Cleanup function for error scenarios
cleanup_on_error() {
    log_debug "Performing error cleanup..."
    # Remove temporary files
    find /tmp -name "plaesy-*" -user "$(id -u)" -mmin +60 -delete 2>/dev/null || true
}


# Performance measurement
time_execution() {
    local start_time=$(date +%s.%N)
    "$@"
    local exit_code=$?
    local end_time=$(date +%s.%N)
    local duration=$(echo "$end_time - $start_time" | bc -l 2>/dev/null || echo "0")
    log_debug "Command executed in ${duration}s with exit code $exit_code"
    return $exit_code
}

# Enhanced validation functions
validate_command_exists() {
    local cmd="$1"
    local description="${2:-command}"

    if command -v "$cmd" >/dev/null 2>&1; then
        log_debug "Found $cmd: $(command -v "$cmd")"
        return 0
    else
        log_error "Required $description not found: $cmd"
        return 1
    fi
}

validate_file_exists() {
    local file="$1"
    local description="${2:-file}"

    if [[ -f "$file" ]]; then
        log_debug "Found $description: $file"
        return 0
    else
        log_error "Required $description not found: $file"
        return 1
    fi
}

validate_directory_exists() {
    local dir="$1"
    local description="${2:-directory}"

    if [[ -d "$dir" ]]; then
        log_debug "Found $description: $dir"
        return 0
    else
        log_error "Required $description not found: $dir"
        return 1
    fi
}

# Safe file operations
safe_copy() {
    local src="$1"
    local dest="$2"
    local backup="${dest}.backup.$(date +%s)"

    if [[ -f "$dest" ]]; then
        cp "$dest" "$backup" || {
            log_error "Failed to create backup of $dest"
            return 1
        }
        log_debug "Created backup: $backup"
    fi

    cp "$src" "$dest" || {
        log_error "Failed to copy $src to $dest"
        [[ -f "$backup" ]] && mv "$backup" "$dest"
        return 1
    }

    log_success "Copied $src to $dest"
    return 0
}

# Environment validation
validate_environment() {
    log_debug "Validating execution environment..."

    # Check minimum bash version
    if [[ "${BASH_VERSION%%.*}" -lt 4 ]]; then
        log_error "Bash version 4.0 or higher required. Current: $BASH_VERSION"
        return 1
    fi

    # Check required commands
    local required_commands=("git" "curl")
    for cmd in "${required_commands[@]}"; do
        validate_command_exists "$cmd" || return 1
    done

    # Check git repository
    if ! git rev-parse --git-dir >/dev/null 2>&1; then
        log_warning "Not in a git repository"
    fi

    log_success "Environment validation passed"
    return 0
}

# Setup error handling
setup_error_handling() {
    # Set up error trap first (before strict mode)
    trap 'handle_error ${LINENO} ${BASH_LINENO} "$BASH_COMMAND"' ERR

    # Set up cleanup trap
    trap cleanup_on_error EXIT

    # Now enable strict mode
    set -euo pipefail

    # Use echo instead of log_debug to avoid circular dependency
    [[ "${DEBUG_MODE:-false}" == "true" ]] && echo "[DEBUG] Error handling configured" >&2
}

print_banner() {
    local title="$1"
    local subtitle="$2"
    echo -e "${PURPLE}"
    echo "██████╗ ██╗      █████╗ ███████╗███████╗██╗   ██╗"
    echo "██╔══██╗██║     ██╔══██╗██╔════╝██╔════╝╝██╗ ██╔╝"
    echo "██████╔╝██║     ███████║█████╗  ███████╗ ╚████╔╝ "
    echo "██╔═══╝ ██║     ██╔══██║██╔══╝  ╚════██║  ╚██╔╝  "
    echo "██║     ███████╗██║  ██║███████╗███████║   ██║   "
    echo "╚═╝     ╚══════╝╚═╝  ╚═╝╚══════╝╚══════╝   ╚═╝   "
    echo ""
    if [ -n "$title" ]; then
        echo "🏛️  $title"
    fi
    if [ -n "$subtitle" ]; then
        echo "   $subtitle"
    fi
    echo -e "${NC}"
}

# Validation functions
validate_not_root() {
    if [[ $EUID -eq 0 ]]; then
        log_error "This script should not be run as root for security reasons."
    fi
}

validate_disk_space() {
    local required_mb="$1"
    local available_space
    available_space=$(df "$HOME" | awk 'NR==2 {print $4}')
    local required_kb=$((required_mb * 1024))
    
    if [[ $available_space -lt $required_kb ]]; then
        log_error "Insufficient disk space. At least ${required_mb}MB required."
    fi
}

get_repo_root() { git rev-parse --show-toplevel; }
get_current_branch() { git rev-parse --abbrev-ref HEAD; }

check_feature_branch() {
    local branch="$1"
    if [[ ! "$branch" =~ ^[0-9]{3}- ]]; then
        echo "ERROR: Not on a feature branch. Current branch: $branch" >&2
        echo "Feature branches should be named like: 001-feature-name" >&2
        return 1
    fi; return 0
}

get_feature_dir() { echo "$1/specs/$2"; }

get_feature_paths() {
    local repo_root=$(get_repo_root)
    local current_branch=$(get_current_branch)
    local feature_dir=$(get_feature_dir "$repo_root" "$current_branch")
    cat <<EOF
REPO_ROOT='$repo_root'
CURRENT_BRANCH='$current_branch'
FEATURE_DIR='$feature_dir'
FEATURE_SPEC='$feature_dir/spec.md'
IMPL_PLAN='$feature_dir/plan.md'
TASKS='$feature_dir/tasks.md'
RESEARCH='$feature_dir/research.md'
DATA_MODEL='$feature_dir/data-model.md'
QUICKSTART='$feature_dir/quickstart.md'
CONTRACTS_DIR='$feature_dir/contracts'
EOF
}

check_file() { [[ -f "$1" ]] && echo "  ✓ $2" || echo "  ✗ $2"; }
check_dir() { [[ -d "$1" && -n $(ls -A "$1" 2>/dev/null) ]] && echo "  ✓ $2" || echo "  ✗ $2"; }
