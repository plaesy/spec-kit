#!/usr/bin/env bash
# Simple common functions for plaesy scripts

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly PURPLE='\033[0;35m'
readonly CYAN='\033[0;36m'
readonly WHITE='\033[1;37m'
readonly NC='\033[0m' # No Color

# Print functions for user feedback
print_step() {
    echo -e "[STEP] $1"
}

print_success() {
    echo -e "[SUCCESS] $1"
}

print_warning() {
    echo -e "[WARNING] $1" >&2
}

print_error() {
    echo -e "[ERROR] $1" >&2
    exit 1
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
        print_error "This script should not be run as root for security reasons."
    fi
}

validate_disk_space() {
    local required_mb="$1"
    local available_space
    available_space=$(df "$HOME" | awk 'NR==2 {print $4}')
    local required_kb=$((required_mb * 1024))
    
    if [[ $available_space -lt $required_kb ]]; then
        print_error "Insufficient disk space. At least ${required_mb}MB required."
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
