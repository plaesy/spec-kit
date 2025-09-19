#!/bin/bash

# Script: create-new-idea.sh
# Purpose: Create idea capture document from raw idea description
# Usage: ./create-new-idea.sh "idea description" [branch-name]
# Part of Plaesy Spec-Kit

set -euo pipefail

REPO_ROOT=$(git rev-parse --show-toplevel)

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

# Configuration
IDEA_TEMPLATE="$REPO_ROOT/.plaesy/templates/idea.template.md"

usage() {
    echo "Usage: $0 [--json] \"idea description\" [branch-name]"
    echo ""
    echo "Create an idea capture document from a raw idea description."
    echo ""
    echo "Arguments:"
    echo "  --json           Output results in JSON format"
    echo "  idea description  The raw idea or problem description"
    echo "  branch-name      Optional: custom branch name (auto-generated if not provided)"
    echo ""
    echo "Examples:"
    echo "  $0 \"Build a task management system for teams\""
    echo "  $0 \"User authentication system\" auth-system"
    echo "  $0 --json \"Build a CLI tool\""
    echo ""
    echo "This script will:"
    echo "  1. Create an idea capture document (idea.md)"
    echo "  2. Set up feature branch and directory structure"
    echo "  3. Output file paths for further processing"
}

validate_inputs() {
    local idea_description="$1"
    
    if [[ -z "$idea_description" ]]; then
        echo -e "${RED}Error: Idea description cannot be empty${NC}" >&2
        usage
        exit 1
    fi
    
    # Check if we're in a git repository
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        echo -e "${RED}Error: This is not a git repository${NC}" >&2
        exit 1
    fi
}

generate_branch_name() {
    local idea_description="$1"
    local branch_name="$2"
    
    if [[ -n "$branch_name" ]]; then
        echo "$branch_name"
        return
    fi
    
    # Auto-generate branch name from idea description
    local auto_name
    auto_name=$(echo "$idea_description" | \
        tr '[:upper:]' '[:lower:]' | \
        sed 's/[^a-z0-9 ]//g' | \
        tr ' ' '-' | \
        sed 's/--*/-/g' | \
        sed 's/^-\|-$//g' | \
        cut -c1-40)
    
    # Add random number to ensure uniqueness
    local number
    number=$(shuf -i 100-999 -n 1)
    echo "${number}-${auto_name}"
}

create_idea_document() {
    local idea_description="$1"
    local branch_name="$2"
    local idea_file="$3"
    
    echo -e "${BLUE}📝 Creating idea document...${NC}"
    
    # Use the general idea template
    local template_file="$IDEA_TEMPLATE"
    
    if [[ ! -f "$template_file" ]]; then
        echo -e "${RED}Error: Idea template not found at ${template_file}${NC}" >&2
        exit 1
    fi
    
    echo -e "${BLUE}Creating idea document from template...${NC}"
    
    # Create idea document from template
    cp "$template_file" "$idea_file"
    
    # Fill in template variables
    local current_date
    current_date=$(date '+%Y-%m-%d')
    
    sed -i "s/\[IDEA NAME\]/${branch_name}/g" "$idea_file" 2>/dev/null || sed -i "" "s/\[IDEA NAME\]/${branch_name}/g" "$idea_file"
    sed -i "s/\[APPLICATION NAME\]/${branch_name}/g" "$idea_file" 2>/dev/null || sed -i "" "s/\[APPLICATION NAME\]/${branch_name}/g" "$idea_file"
    sed -i "s/\[TASK NAME\]/${branch_name}/g" "$idea_file" 2>/dev/null || sed -i "" "s/\[TASK NAME\]/${branch_name}/g" "$idea_file"
    sed -i "s/\[TOOL NAME\]/${branch_name}/g" "$idea_file" 2>/dev/null || sed -i "" "s/\[TOOL NAME\]/${branch_name}/g" "$idea_file"
    sed -i "s/\[SERVICE NAME\]/${branch_name}/g" "$idea_file" 2>/dev/null || sed -i "" "s/\[SERVICE NAME\]/${branch_name}/g" "$idea_file"
    sed -i "s/\[RESEARCH TITLE\]/${branch_name}/g" "$idea_file" 2>/dev/null || sed -i "" "s/\[RESEARCH TITLE\]/${branch_name}/g" "$idea_file"
    sed -i "s/\[DATE\]/${current_date}/g" "$idea_file" 2>/dev/null || sed -i "" "s/\[DATE\]/${current_date}/g" "$idea_file"
    sed -i "s/\$ARGUMENTS/${idea_description}/g" "$idea_file" 2>/dev/null || sed -i "" "s/\$ARGUMENTS/${idea_description}/g" "$idea_file"
    
    echo -e "${GREEN}✓ Idea document created: ${idea_file}${NC}"
}

create_feature_structure() {
    local branch_name="$1"
    local specs_dir="$2"
    
    echo -e "${BLUE}Setting up feature structure...${NC}"
    
    # Create branch
    if ! git checkout -b "feature/${branch_name}" 2>/dev/null; then
        if git show-ref --verify --quiet "refs/heads/feature/${branch_name}"; then
            echo -e "${YELLOW}Branch feature/${branch_name} already exists, checking out...${NC}"
            git checkout "feature/${branch_name}"
        else
            echo -e "${RED}Failed to create or checkout branch feature/${branch_name}${NC}" >&2
            exit 1
        fi
    fi
    
    # Create directory structure
    mkdir -p "$specs_dir"
    
    echo -e "${GREEN}✓ Feature structure created: $specs_dir${NC}"
}

main() {
    local json_output=false
    local idea_description=""
    local branch_name=""
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --json)
                json_output=true
                shift
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            *)
                if [[ -z "$idea_description" ]]; then
                    idea_description="$1"
                elif [[ -z "$branch_name" ]]; then
                    branch_name="$1"
                fi
                shift
                ;;
        esac
    done
    
    validate_inputs "$idea_description"
    
    # Generate branch name
    branch_name=$(generate_branch_name "$idea_description" "$branch_name")
    
    # Set up paths
    local specs_dir="specs/${branch_name}"
    local idea_file="${specs_dir}/idea.md"
    
    if [[ "$json_output" == "false" ]]; then
        echo -e "${GREEN}Starting create-new-idea process...${NC}"
        echo "Idea: $idea_description"
        echo "Branch: $branch_name"
        echo ""
    fi
    
    # Create feature structure
    create_feature_structure "$branch_name" "$specs_dir"
    
    # Create idea document
    create_idea_document "$idea_description" "$branch_name" "$idea_file"
    
    # Output results
    if [[ "$json_output" == "true" ]]; then
        # Output JSON format for automation
        cat <<EOF
{
  "success": true,
  "branch_name": "$branch_name",
  "idea_file": "$(realpath "$idea_file")",
  "specs_dir": "$(realpath "$specs_dir")"
}
EOF
    else
        # Human-readable summary
        echo ""
        echo -e "${GREEN}=================================================="
        echo "           IDEA CAPTURE COMPLETE"
        echo "==================================================${NC}"
        echo ""
        echo "Files created:"
        echo "  - Idea document: $idea_file"
        echo ""
        echo "Next steps:"
        echo "  1. Review and refine the idea document"
        echo "  2. Use idea.prompt.md to perform research and exploration"
        echo "  3. Use specify.prompt.md to create the specification"
        echo ""
        echo -e "${YELLOW}Tip: The idea document is ready for AI-powered research and specification${NC}"
    fi
}

# Run main function with all arguments
main "$@"