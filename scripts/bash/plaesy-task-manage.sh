#!/bin/bash

# plaesy-task-manage.sh - Task lifecycle management for autonomous workflows
# Moves task files between status directories (backlog → todo → doing → done/blocked)
# Used by /evolve and manual task management workflows

set -euo pipefail

# Colors
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m'

# Logging functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[✓]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1" >&2; exit 1; }
log_warning() { echo -e "${YELLOW}[!]${NC} $1"; }

# Get project root (find .plaesy directory)
find_project_root() {
    local current="$(pwd)"
    while [[ "$current" != "/" ]]; do
        if [[ -d "$current/.plaesy" ]]; then
            echo "$current"
            return 0
        fi
        current="$(dirname "$current")"
    done
    log_error "Project root not found (.plaesy directory not found)"
}

PROJECT_ROOT=$(find_project_root)
TASKS_DIR="$PROJECT_ROOT/.plaesy/tasks"

# Validate status
validate_status() {
    local status="$1"
    case "$status" in
        backlog|todo|doing|done|blocked) return 0 ;;
        *) log_error "Invalid status: $status (must be: backlog, todo, doing, done, blocked)" ;;
    esac
}

# Get next task from backlog
get_next_task() {
    if [[ ! -d "$TASKS_DIR/backlog" ]]; then
        log_error "Backlog directory not found: $TASKS_DIR/backlog"
    fi

    # List tasks sorted by priority (highest first) or alphabetical
    local tasks=($(ls -1 "$TASKS_DIR/backlog"/*.md 2>/dev/null | sort -r | head -1))

    if [[ ${#tasks[@]} -eq 0 ]]; then
        return 1  # No tasks in backlog
    fi

    echo "$(basename "${tasks[0]}")"
}

# Move task file between directories
move_task() {
    local task_file="$1"
    local from_status="$2"
    local to_status="$3"

    validate_status "$from_status"
    validate_status "$to_status"

    local from_path="$TASKS_DIR/$from_status/$task_file"
    local to_path="$TASKS_DIR/$to_status/$task_file"

    if [[ ! -f "$from_path" ]]; then
        log_error "Task not found: $from_path"
    fi

    mv "$from_path" "$to_path"
    log_success "Moved: $task_file ($from_status → $to_status)"
}

# Pop next task from backlog and move to todo
pop_next_task() {
    local task_file
    task_file=$(get_next_task) || {
        log_warning "No tasks in backlog"
        return 1
    }

    move_task "$task_file" "backlog" "todo"
    echo "$task_file"  # Return filename
}

# Move task to doing (in progress)
task_start() {
    local task_file="$1"
    move_task "$task_file" "todo" "doing"
}

# Move task to done (completed)
task_complete() {
    local task_file="$1"
    move_task "$task_file" "doing" "done"
    log_success "Task completed: $task_file"
}

# Move task to blocked (waiting)
task_block() {
    local task_file="$1"
    local reason="${2:-No reason provided}"
    move_task "$task_file" "doing" "blocked"

    # Append block reason to file
    if [[ -f "$TASKS_DIR/blocked/$task_file" ]]; then
        echo "" >> "$TASKS_DIR/blocked/$task_file"
        echo "## Blocked Reason" >> "$TASKS_DIR/blocked/$task_file"
        echo "$reason" >> "$TASKS_DIR/blocked/$task_file"
    fi
}

# Unblock task (move from blocked back to todo)
task_unblock() {
    local task_file="$1"
    move_task "$task_file" "blocked" "todo"
    log_success "Task unblocked: $task_file"
}

# List tasks by status
list_tasks() {
    local status="$1"
    validate_status "$status"

    if [[ ! -d "$TASKS_DIR/$status" ]]; then
        log_warning "No tasks in $status"
        return 0
    fi

    local count=$(ls -1 "$TASKS_DIR/$status"/*.md 2>/dev/null | wc -l)
    echo -e "${BLUE}$status ($count)${NC}"

    if [[ $count -gt 0 ]]; then
        ls -1 "$TASKS_DIR/$status"/*.md | xargs -I {} basename {} | sed 's/\.md$//' | sed 's/^/  - /'
    fi
}

# List all tasks
list_all_tasks() {
    echo -e "${BLUE}📋 Task Summary${NC}"
    echo ""
    for status in backlog todo doing done blocked; do
        list_tasks "$status"
        echo ""
    done
}

# Show task details
show_task() {
    local task_file="$1"
    local status="${2:-}"

    local file_path=""
    if [[ -n "$status" ]]; then
        validate_status "$status"
        file_path="$TASKS_DIR/$status/$task_file"
    else
        # Search for task in all status directories
        for s in backlog todo doing done blocked; do
            if [[ -f "$TASKS_DIR/$s/$task_file" ]]; then
                file_path="$TASKS_DIR/$s/$task_file"
                status="$s"
                break
            fi
        done
    fi

    if [[ ! -f "$file_path" ]]; then
        log_error "Task not found: $task_file"
    fi

    echo -e "${BLUE}📄 Task: $task_file${NC}"
    echo -e "${BLUE}Status: $status${NC}"
    echo ""
    cat "$file_path"
}

# Main CLI
main() {
    local cmd="${1:-list}"

    case "$cmd" in
        list)
            list_all_tasks
            ;;
        list-status)
            if [[ $# -lt 2 ]]; then
                log_error "Usage: $0 list-status {backlog|todo|doing|done|blocked}"
            fi
            list_tasks "$2"
            ;;
        next)
            pop_next_task
            ;;
        start)
            if [[ $# -lt 2 ]]; then
                log_error "Usage: $0 start <task_file>"
            fi
            task_start "$2"
            ;;
        complete)
            if [[ $# -lt 2 ]]; then
                log_error "Usage: $0 complete <task_file>"
            fi
            task_complete "$2"
            ;;
        block)
            if [[ $# -lt 2 ]]; then
                log_error "Usage: $0 block <task_file> [reason]"
            fi
            task_block "$2" "${3:-}"
            ;;
        unblock)
            if [[ $# -lt 2 ]]; then
                log_error "Usage: $0 unblock <task_file>"
            fi
            task_unblock "$2"
            ;;
        move)
            if [[ $# -lt 4 ]]; then
                log_error "Usage: $0 move <task_file> <from_status> <to_status>"
            fi
            move_task "$2" "$3" "$4"
            ;;
        show)
            if [[ $# -lt 2 ]]; then
                log_error "Usage: $0 show <task_file> [status]"
            fi
            show_task "$2" "${3:-}"
            ;;
        *)
            log_error "Usage: $0 {list|list-status|next|start|complete|block|unblock|move|show}"
            ;;
    esac
}

main "$@"
