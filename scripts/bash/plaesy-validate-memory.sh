#!/bin/bash
# Plaesy Memory Self-Containment Validator
# Scans .plaesy/memory/ for external paths and validates self-containment rule

set -euo pipefail
IFS=$' \n\t'

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

# Logging
log_info() { echo -e "${CYAN}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[✓]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[!]${NC} $1"; }
log_error() { echo -e "${RED}[✗]${NC} $1"; }

# Get memory directory
MEMORY_DIR=".plaesy/memory"
if [[ ! -d "$MEMORY_DIR" ]]; then
    log_error "Memory directory not found: $MEMORY_DIR"
    exit 1
fi

log_info "Scanning $MEMORY_DIR for external references..."
echo ""

# Track errors
ERRORS_FOUND=0
FILES_CHECKED=0

# Patterns to check for (external paths)
PATTERNS=(
    "~/.claude"
    "~/.claude/projects"
    "~/.claude/plans"
    "/tmp/"
    "C:\\Users\\.*\\.claude"
    "/Users/.*/.claude"
    "CLAUDE_CODE_DIR"
)

# Scan all markdown files in .plaesy/memory/
for file in $(find "$MEMORY_DIR" -type f -name "*.md"); do
    ((FILES_CHECKED++))
    ISSUES=()

    # Check each pattern
    for pattern in "${PATTERNS[@]}"; do
        if grep -q "$pattern" "$file" 2>/dev/null; then
            ISSUES+=("$pattern")
        fi
    done

    # Report if issues found
    if [[ ${#ISSUES[@]} -gt 0 ]]; then
        ((ERRORS_FOUND++))
        log_error "External references in: $file"

        # Show specific lines with issues
        for pattern in "${ISSUES[@]}"; do
            echo "  Lines with '$pattern':"
            grep -n "$pattern" "$file" 2>/dev/null | sed 's/^/    /' || true
        done
        echo ""
    fi
done

echo ""
log_info "Scan complete:"
echo "  Files checked: $FILES_CHECKED"
echo "  External refs found: $ERRORS_FOUND"
echo ""

if [[ $ERRORS_FOUND -eq 0 ]]; then
    log_success "Self-containment validated ✓"
    log_success "Memory is project-local and git-safe"
    exit 0
else
    log_warn "Self-containment issues detected"
    echo ""
    echo "To fix:"
    echo "1. For content in ~/.claude/projects: Copy it into .plaesy/memory/"
    echo "2. For external links: Replace with internal .plaesy/memory/ references (flat structure, no subfolders)"
    echo "3. Re-run this script after fixes"
    exit 1
fi
