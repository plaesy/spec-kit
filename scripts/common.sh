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
    echo -e "Step: $1"
}

print_success() {
    echo -e "Success: $1"
}

print_warning() {
    echo -e "Warning $1" >&2
}

print_error() {
    echo -e "Error: $1" >&2
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