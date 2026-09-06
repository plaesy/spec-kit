#!/bin/bash

# AI-Specific Header Injection Script
# Injects platform-specific headers into prompt files based on chosen AI platform

set -euo pipefail

# Script configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
HEADERS_DIR="$REPO_ROOT/templates/ai-headers"
DEFAULT_HEADER="manual.header.yaml"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" >&2
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" >&2
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" >&2
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
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
    --merge               Merge header keys into existing front-matter (add missing keys)
    --help                Show this help message

HEADER FILES:
    Headers are now stored as YAML files in templates/ai-headers/
    - <platform>.prompts.yaml    - Headers for prompt files
    - <platform>.chatmodes.yaml  - Headers for chatmode files
    - <platform>.instructions.yaml - Headers for instruction files
    - <platform>.header.yaml     - Generic headers for the platform

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

# Determine header file for platform and file type (prompts, chatmodes, instructions, etc.)
# Fallback order:
# 1. $HEADERS_DIR/<platform>.<type>.yaml
# 2. $HEADERS_DIR/<platform>.header.yaml
# 3. $HEADERS_DIR/$DEFAULT_HEADER
get_header_file() {
    local platform="$1"
    local ftype="${2:-}"

    # candidate 1: platform + type
    if [[ -n "$ftype" ]]; then
        local candidate="$HEADERS_DIR/${platform}.${ftype}.yaml"
        if [[ -f "$candidate" ]]; then
            echo "$candidate"
            return 0
        fi
    fi

    # candidate 2: platform header
    local platform_header="$HEADERS_DIR/${platform}.header.yaml"
    if [[ -f "$platform_header" ]]; then
        echo "$platform_header"
        return 0
    fi

    # fallback: default header
    if [[ -n "$ftype" ]]; then
        log_warning "No specific header for ${platform} (type: $ftype), using default header"
    else
        log_warning "No specific header for ${platform}, using default header"
    fi
    echo "$HEADERS_DIR/$DEFAULT_HEADER"
}

# Find prompt files
find_prompt_files() {
    local target_dir="$1"
    shift || true
    local user_patterns=()
    local user_excludes=()
    # collect positional args as patterns until a special token
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --exclude)
                shift
                user_excludes+=("$1")
                shift
                ;;
            *)
                user_patterns+=("$1")
                shift
                ;;
        esac
    done

    if [[ ! -d "$target_dir" ]]; then
        log_error "Target directory does not exist: $target_dir"
        return 1
    fi

    # default safe patterns
    if [[ ${#user_patterns[@]} -eq 0 ]]; then
        user_patterns=("*.prompt.md" "*.chatmode.md" "*.instructions.md")
    fi

    local expr=""
    for p in "${user_patterns[@]}"; do
        if [[ -n "$expr" ]]; then expr="$expr -o "; fi
        expr="$expr -name \"$p\""
    done

    # build exclude expression
    local excl_expr=""
    for e in "${user_excludes[@]}"; do
        excl_expr="$excl_expr -not -path \"*/$e\""
    done

    # shell-eval the find command safely
    eval "find \"$target_dir\" -type f \( $expr \) $excl_expr"
}

# Check if file already has header
has_header() {
    local file="$1"

    # Check for YAML front-matter: first line must be '---' and there must be a closing '---' within first 50 lines
    if sed -n '1,1p' "$file" | grep -q '^---$'; then
        # extract the frontmatter block
        fm=$(sed -n '1,50p' "$file" | awk '/^---$/ { if (++c==1) {next} else {exit}} c==1 {print}') || true
        # check for explicit keys platform: or type:
        if printf '%s' "$fm" | grep -qiE '^[[:space:]]*(platform|type)[[:space:]]*:' ; then
            return 0
        fi
        # if no explicit keys, still accept front-matter presence
        if sed -n '2,50p' "$file" | grep -q '^---$'; then
            return 0
        fi
    fi

    # Fallback: look for common header indicators in first 20 lines
    if head -20 "$file" | grep -qE "(Constitutional.*Framework|AI.*Configuration|# .*Header)" ; then
        return 0
    fi

    return 1
}

# Extract a key from YAML front-matter (returns empty if not present)
get_front_matter_key() {
    local file="$1"
    local key="$2"
    if sed -n '1,1p' "$file" | grep -q '^---$'; then
        sed -n '1,50p' "$file" | awk '/^---$/ { if (++c==1) {next} else {exit}} c==1 {print}' | sed -n "s/^[[:space:]]*${key}:[[:space:]]*\(.*\)/\1/pI" | sed -n '1p' || true
    fi
}

# Extract description from target file's YAML front-matter
get_file_description() {
    local target_file="$1"
    local description=""

    # Check if file exists and has YAML front-matter
    if [[ -f "$target_file" ]] && sed -n '1,1p' "$target_file" | grep -q '^---$'; then
        # Extract description from front-matter
        description=$(sed -n '1,50p' "$target_file" | awk '/^---$/ { if (++c==1) {next} else {exit}} c==1 {print}' | sed -n 's/^[[:space:]]*description:[[:space:]]*"\?\([^"]*\)"\?[[:space:]]*$/\1/pI' | sed -n '1p' || true)
    fi

    # If no description found, provide a default based on filename
    if [[ -z "$description" ]]; then
        local filename=$(basename "$target_file" | sed 's/\.[^.]*$//')
        case "$target_file" in
            *.prompt.md) description="Prompt: $filename" ;;
            *.chatmode.md) description="Chat mode: $filename" ;;
            *.instruction.md) description="Instructions: $filename" ;;
            *) description="Configuration: $filename" ;;
        esac
    fi

    echo "$description"
}

# Extract header content from YAML file and replace placeholders
get_yaml_header_content() {
    local yaml_file="$1"
    local target_file="${2:-}"

    # Check if file exists and is readable
    if [[ ! -f "$yaml_file" || ! -r "$yaml_file" ]]; then
        log_error "YAML header file not found or not readable: $yaml_file"
        return 1
    fi

    # Use awk to extract the header_content value, handling multi-line YAML strings
    local content
    content=$(awk '
    BEGIN { in_header_content = 0; content = "" }
    /^header_content:[[:space:]]*\|[[:space:]]*$/ {
        in_header_content = 1
        next
    }
    in_header_content && /^[[:space:]]+/ {
        # Remove the leading spaces that indicate YAML indentation
        line = $0
        gsub(/^[[:space:]]+/, "", line)
        if (content != "") content = content "\n"
        content = content line
        next
    }
    in_header_content && /^[^[:space:]]/ {
        in_header_content = 0
    }
    END {
        if (content != "") print content
    }' "$yaml_file")

    # Replace placeholders if target file is provided
    if [[ -n "$target_file" && -n "$content" ]]; then
        local file_description
        file_description=$(get_file_description "$target_file")

        # Replace {{DESCRIPTION}} placeholder with actual description
        content=$(echo "$content" | sed "s/{{DESCRIPTION}}/$file_description/g")
    fi

    echo "$content"
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
    local tmp

    tmp=$(mktemp "${file}.tmp.XXXXXX") || { log_error "Failed to create temp file"; return 1; }
    trap 'rm -f "$tmp"' RETURN

    {
        printf '%s\n\n\n' "$header_content"
        cat "$file"
    } > "$tmp"

    # preserve permissions and ownership where possible
    if command -v stat >/dev/null 2>&1; then
        perms=$(stat -c '%a' "$file" 2>/dev/null || true)
        owner=$(stat -c '%u' "$file" 2>/dev/null || true)
        group=$(stat -c '%g' "$file" 2>/dev/null || true)
        if [[ -n "$perms" ]]; then
            chmod "$perms" "$tmp" 2>/dev/null || true
        fi
        if [[ -n "$owner" && -n "$group" ]]; then
            chown "$owner":"$group" "$tmp" 2>/dev/null || true
        fi
    fi

    mv "$tmp" "$file"
    trap - RETURN
}

# Merge header front-matter keys into existing file front-matter.
# Only adds keys that do not exist in the file's front-matter. Does not overwrite.
merge_front_matter() {
    local file="$1"
    local header_content="$2"
    local force="$3"
    local backup="$4"
    local dry_run="$5"

    local filename
    filename=$(basename "$file")

    # Extract existing front-matter (without the --- markers)
    existing_fm=$(sed -n '1,200p' "$file" | awk '/^---$/ { if (++c==1) {next} else {exit}} c==1 {print}') || true

    # Extract header front-matter from header_content
    header_fm=$(printf '%s\n' "$header_content" | sed -n '1,200p' | awk '/^---$/ { if (++c==1) {next} else {exit}} c==1 {print}') || true

    if [[ -z "$existing_fm" || -z "$header_fm" ]]; then
        log_warning "Cannot merge front-matter for $filename: missing existing or header front-matter"
        return 1
    fi

    # Build merged front-matter: keep existing keys, then append header keys that do not exist
    merged_fm=$(awk -F': ' '
        NR==FNR { key=$1; sub(/^[ \t]+/,"",key); vals[key]=substr($0,index($0,": ")+2); seen[key]=1; order[++n]=key; next }
        { key=$1; sub(/^[ \t]+/,"",key); if(!(key in seen)){ vals[key]=substr($0,index($0,": ")+2); extra[++m]=key } }
        END {
            for(i=1;i<=n;i++){ k=order[i]; print k": " vals[k] }
            for(i=1;i<=m;i++){ k=extra[i]; print k": " vals[k] }
        }' <(printf '%s\n' "$existing_fm") <(printf '%s\n' "$header_fm") ) || true

    if [[ -z "$merged_fm" ]]; then
        log_warning "Merged front-matter is empty for $filename"
        return 1
    fi

    if [[ "$dry_run" == "true" ]]; then
        log_info "[DRY RUN] Would merge front-matter into $filename:\n$merged_fm"
        return 0
    fi

    # Create backup if requested
    if [[ "$backup" == "true" ]]; then
        create_backup "$file"
    fi

    # Rebuild the file: new front-matter block + rest of original content after existing fm
    local header_end
    header_end=$(grep -n "^---$" "$file" | sed -n '2p' | cut -d: -f1)
    if [[ -z "$header_end" ]]; then
        log_error "Failed to find end of front-matter for $filename"
        return 1
    fi

    tmp=$(mktemp "${file}.tmp.XXXXXX") || { log_error "Failed to create temp file"; return 1; }
    trap 'rm -f "$tmp"' RETURN

    {
        echo '---'
        printf '%s\n' "$merged_fm"
        echo '---'
        tail -n "+$((header_end + 1))" "$file"
    } > "$tmp"

    mv "$tmp" "$file"
    trap - RETURN
    log_success "Merged front-matter into $filename"
    return 0
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
    local merge=false
    local backup=false
    local list_only=false
    local patterns=()
    local excludes=()

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
            --list-only)
                list_only=true
                shift
                ;;
            --pattern)
                patterns+=("$2")
                shift 2
                ;;
            --exclude)
                excludes+=("$2")
                shift 2
                ;;
            --force)
                force=true
                shift
                ;;
            --merge)
                merge=true
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

    # Load header content (for main header file, no specific target file yet)
    local header_content
    if [[ "$header_file" == *.yaml ]]; then
        header_content=$(get_yaml_header_content "$header_file")
    else
        header_content=$(cat "$header_file")
    fi

    log_info "Using header file: $(basename "$header_file")"

    # Find files that may need headers
    local prompt_files
    if [[ ${#patterns[@]} -gt 0 ]]; then
        mapfile -t prompt_files < <(find_prompt_files "$target_dir" "${patterns[@]}")
    else
        mapfile -t prompt_files < <(find_prompt_files "$target_dir")
    fi

    if [[ ${#prompt_files[@]} -eq 0 ]]; then
        log_warning "No prompt files found in $target_dir"
        exit 0
    fi

    log_info "Found ${#prompt_files[@]} prompt file(s)"

    # If list-only, print mapping and exit
    if [[ "$list_only" == "true" ]]; then
        for file in "${prompt_files[@]}"; do
            relpath=$(realpath --relative-to="$target_dir" "$file")
            if [[ "$relpath" =~ (^|/)prompts?(/|$) ]] || [[ "$file" == *.prompt.md ]]; then
                ftype="prompts"
            elif [[ "$relpath" =~ (^|/)chatmodes?(/|$) ]] || [[ "$file" == *.chatmode.md ]] || [[ "$file" == *chatmode* ]]; then
                ftype="chatmodes"
            elif [[ "$relpath" =~ (^|/)instructions?(/|$) ]] || [[ "$file" == *.instruction.md ]] || [[ "$file" == *instructions* ]]; then
                ftype="instructions"
            else
                ftype="generic"
            fi
            header_file_for_type=$(get_header_file "$ai_platform" "$ftype")
            echo "$file -> type=$ftype -> header=$(basename "$header_file_for_type")"
        done
        exit 0
    fi

    # Process each file
    local processed=0
    local skipped=0

    for file in "${prompt_files[@]}"; do
        # derive file type from path/filename
        local relpath
        relpath=$(realpath --relative-to="$target_dir" "$file")
        local ftype=""

        # heuristics: check directory name or filename patterns
        if [[ "$relpath" =~ (^|/)prompts?(/|$) ]] || [[ "$file" == *.prompt.md ]]; then
            ftype="prompts"
        elif [[ "$relpath" =~ (^|/)chatmodes?(/|$) ]] || [[ "$file" == *.chatmode.md ]] || [[ "$file" == *chatmode* ]]; then
            ftype="chatmodes"
        elif [[ "$relpath" =~ (^|/)instructions?(/|$) ]] || [[ "$file" == *.instruction.md ]] || [[ "$file" == *instructions* ]]; then
            ftype="instructions"
        else
            # fallback to generic
            ftype="generic"
        fi

        # pick header for this file type
        local header_file_for_type
        header_file_for_type=$(get_header_file "$ai_platform" "$ftype")
        local header_content_for_file
        if [[ "$header_file_for_type" == *.yaml ]]; then
            # Pass the target file to enable dynamic description replacement
            header_content_for_file=$(get_yaml_header_content "$header_file_for_type" "$file")
        else
            header_content_for_file=$(cat "$header_file_for_type")
        fi

        # Skip if header content is empty
        if [[ -z "$(echo "$header_content_for_file" | tr -d '[:space:]')" ]]; then
            log_info "Skipping $(basename "$file") (header file $(basename "$header_file_for_type") is empty)"
            skipped=$((skipped + 1))
            continue
        fi

        # skip if file has explicit front-matter platform that doesn't match target platform
        fm_platform=$(get_front_matter_key "$file" platform || true)
        if [[ -n "$fm_platform" && "$fm_platform" != "$ai_platform" && "$force" != "true" ]]; then
            log_info "Skipping $(basename "$file") (front-matter platform: $fm_platform does not match target: $ai_platform)"
            skipped=$((skipped + 1))
            continue
        fi

        # skip if file already has the same platform configured (no merge needed)
        if [[ -n "$fm_platform" && "$fm_platform" == "$ai_platform" && "$force" != "true" ]]; then
            log_info "Skipping $(basename "$file") (platform $ai_platform already configured)"
            skipped=$((skipped + 1))
            continue
        fi

        # If merge requested and file already has front-matter, attempt a merge first
        if [[ "$merge" == "true" ]] && has_header "$file" && [[ "$force" != "true" ]]; then
            if merge_front_matter "$file" "$header_content_for_file" "$force" "$backup" "$dry_run"; then
                processed=$((processed + 1))
                continue
            else
                skipped=$((skipped + 1))
                continue
            fi
        fi

        if process_file "$file" "$header_content_for_file" "$force" "$backup" "$dry_run"; then
            processed=$((processed + 1))
        else
            skipped=$((skipped + 1))
        fi
    done

    # Summary
    echo ""
    if [[ "$dry_run" == "true" ]]; then
        log_info "DRY RUN COMPLETE"
        log_info "Would process: $processed files"
        log_info "Would skip: $skipped files"
        exit 0
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