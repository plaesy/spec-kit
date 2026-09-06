#!/usr/bin/env bash
# Get paths for current feature branch without creating anything
# Used by commands that need to find existing feature files

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

# Get all paths with error handling
if eval $(get_feature_paths 2>/dev/null); then
    # Successfully got paths, check if on feature branch
    if ! check_feature_branch "$CURRENT_BRANCH" 2>/dev/null; then
        # Not on feature branch, but still output available info
        echo "REPO_ROOT: $REPO_ROOT"
        echo "BRANCH: $CURRENT_BRANCH"
        echo "FEATURE_DIR: $FEATURE_DIR"
        echo "FEATURE_SPEC: $FEATURE_SPEC"
        echo "IMPL_PLAN: $IMPL_PLAN"
        echo "TASKS: $TASKS"
        echo "INFO: Not on a feature branch (format: XXX-feature-name)"
        exit 0
    fi
else
    # Failed to get paths, provide fallback values
    echo "REPO_ROOT: $(pwd)"
    echo "BRANCH: $(git branch --show-current 2>/dev/null || echo 'unknown')"
    echo "FEATURE_DIR: Not available"
    echo "FEATURE_SPEC: Not available"
    echo "IMPL_PLAN: Not available"
    echo "TASKS: Not available"
    echo "INFO: Unable to determine feature paths"
    exit 0
fi

# Output paths (don't create anything)
echo "REPO_ROOT: $REPO_ROOT"
echo "BRANCH: $CURRENT_BRANCH"
echo "FEATURE_DIR: $FEATURE_DIR"
echo "FEATURE_SPEC: $FEATURE_SPEC"
echo "IMPL_PLAN: $IMPL_PLAN"
echo "TASKS: $TASKS"
