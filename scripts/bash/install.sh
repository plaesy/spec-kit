#!/bin/bash

# Plaesy Spec-Kit Installer
# Constitutional Development Framework
# https://github.com/plaesy/spec-kit

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=common.sh
source "$SCRIPT_DIR/common.sh" 2>/dev/null || {
    # If common.sh is not available (e.g., remote installation), define essential functions locally
    echo "⚠️  Loading common functions failed, using embedded functions for remote installation"
    
    # Colors for output
    readonly RED='\033[0;31m'
    readonly GREEN='\033[0;32m'
    readonly YELLOW='\033[1;33m'
    readonly BLUE='\033[0;34m'
    readonly PURPLE='\033[0;35m'
    readonly CYAN='\033[0;36m'
    readonly WHITE='\033[1;37m'
    readonly NC='\033[0m' # No Color

    # Log functions for user feedback
    log_info() {
        echo -e "${BLUE}[INFO]${NC} $1"
    }

    log_success() {
        echo -e "${GREEN}[SUCCESS]${NC} $1"
    }

    log_warning() {
        echo -e "${YELLOW}[WARNING]${NC} $1" >&2
    }

    log_error() {
        echo -e "${RED}[ERROR]${NC} $1" >&2
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
}

# Configuration
readonly REPO_URL="https://github.com/plaesy/spec-kit"
readonly RAW_URL="https://raw.githubusercontent.com/plaesy/spec-kit/main"
readonly INSTALL_DIR="$HOME/.plaesy"
readonly BIN_DIR="$HOME/.local/bin"

# Get version from remote VERSION file
get_remote_version() {
    local remote_version=$(curl -s "$RAW_URL/VERSION" 2>/dev/null | head -1 | tr -d '[:space:]')
    if [[ -z "$remote_version" || ! "$remote_version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo "0.0.0" # Fallback version
    else
        echo "$remote_version"
    fi
}

readonly VERSION=$(get_remote_version)

# Ensure bin directory exists and is in PATH
mkdir -p "$BIN_DIR"

# Functions

# Validation functions
validate_install_environment() {
    validate_not_root
    validate_disk_space 100
}

# Version check function
check_existing_installation() {
    if [[ -f "$INSTALL_DIR/VERSION" ]]; then
        local installed_version
        installed_version=$(cat "$INSTALL_DIR/VERSION" 2>/dev/null || echo "unknown")
        if [[ "$installed_version" == "$VERSION" ]]; then
            log_warning "Plaesy Spec-Kit $VERSION is already installed."
            echo "Use --upgrade to update or --repair to fix installation."
            exit 0
        fi
    fi
}

check_requirements() {
    log_info "Checking system requirements..."
    
    local missing_deps=()
    
    # Check for required commands
    for cmd in curl git; do
        if ! command -v "$cmd" &> /dev/null; then
            missing_deps+=("$cmd")
        fi
    done
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        log_error "Missing required dependencies: ${missing_deps[*]}. Please install them first."
    fi
    
    log_success "System requirements satisfied"
}

download_framework() {
    log_info "Downloading Plaesy Spec-Kit framework..."
    
    # Remove existing installation if it exists
    if [[ -d "$INSTALL_DIR" ]]; then
        log_warning "Removing existing installation..."
        rm -rf "$INSTALL_DIR" || log_error "Failed to remove existing installation"
    fi
    
    # Create temp directory for safer operations
    local temp_dir
    temp_dir=$(mktemp -d) || log_error "Failed to create temporary directory"
    
    # Clone the repository to temp directory first
    if ! git clone --depth 1 --quiet "$REPO_URL" "$temp_dir"; then
        rm -rf "$temp_dir"
        log_error "Failed to download framework"
    fi
    
    # Move contents to final location (not the temp directory itself)
    mkdir -p "$INSTALL_DIR"
    mv "$temp_dir"/* "$INSTALL_DIR"/ || {
        rm -rf "$temp_dir"
        log_error "Failed to install framework"
    }
    
    # Clean up temp directory
    rm -rf "$temp_dir"
    
    # Remove .git directory to save space
    rm -rf "$INSTALL_DIR/.git"
    
    log_success "Framework downloaded successfully"
}

# Platform-specific script filtering for optimized installation
# This function removes scripts that are not needed for the current platform
# to reduce installation size and improve performance.
#
# For bash installer (Linux/macOS):
# - Removes: scripts/powershell directory and all .ps1 files  
# - Keeps: scripts/bash directory and all .sh files
# - Result: ~60% smaller installation, faster script execution
filter_platform_scripts() {
    log_info "Filtering platform-specific scripts..."
    
    # Remove PowerShell scripts since this is a bash installer
    if [ -d "$INSTALL_DIR/scripts/powershell" ]; then
        rm -rf "$INSTALL_DIR/scripts/powershell"
        log_success "Removed PowerShell scripts directory (not needed for this platform)"
    fi
    
    # Remove individual .ps1 files from scripts directory
    find "$INSTALL_DIR/scripts" -name "*.ps1" -type f -delete 2>/dev/null
    if [ $? -eq 0 ]; then
        log_success "Removed individual PowerShell scripts (not needed for this platform)"
    fi
    
    # Keep only bash scripts
    if [ -d "$INSTALL_DIR/scripts/bash" ]; then
        log_success "Keeping bash scripts for this platform"
    fi
    
    log_success "Platform-specific script filtering completed"
}

create_plaesy_cli() {
    log_info "Creating plaesy CLI command..."
    
    cat > "$BIN_DIR/plaesy" << 'EOF'
#!/bin/bash

# Plaesy CLI - Constitutional Development Framework
# Generated by installer

PLAESY_HOME="$HOME/.plaesy"

if [ ! -d "$PLAESY_HOME" ] && [ "$1" != "install" ]; then
    echo "❌ Plaesy framework not found. Please run the installer first."
    echo "curl -s https://raw.githubusercontent.com/plaesy/spec-kit/main/scripts/bash/install.sh | bash"
    exit 1
fi

case "$1" in
    "init")
        # Check if plaesy-init.sh exists, if not, offer to fix it
        if [ ! -f "$PLAESY_HOME/scripts/bash/plaesy-init.sh" ]; then
            echo "❌ Initialization script missing. This can happen after an upgrade."
            echo "🔧 Fixing by recreating the initialization script..."
            echo ""
            echo "📦 Running repair installation..."
            curl -s https://raw.githubusercontent.com/plaesy/spec-kit/main/scripts/bash/install.sh | bash -s -- --repair
            exit 0
        fi

        # Default to current directory if no target specified
        init_args=("${@:2}")
        if [[ ${#init_args[@]} -eq 0 ]]; then
            init_args=(".")
        fi

        # Use config-manager for platform detection
        if [ -f "$PLAESY_HOME/scripts/bash/config-manager.sh" ]; then
            detected_platform=$("$PLAESY_HOME/scripts/bash/config-manager.sh" detect-platform 2>/dev/null)

            if [[ -n "$detected_platform" ]]; then
                echo "✅ Detected AI platform: $detected_platform"
                echo "🚀 Using platform-adapted initialization..."
                echo "💡 Use: plaesy init . --ai $detected_platform"
                "$PLAESY_HOME/scripts/bash/plaesy-init.sh" --ai "$detected_platform" "${init_args[@]}"
            else
                echo "🤖 No AI platform detected. Starting interactive platform selection..."
                echo ""
                "$PLAESY_HOME/scripts/bash/plaesy-init.sh" "${init_args[@]}"
            fi
        else
            echo "⚠️  Platform detector not available, using standard initialization"
            "$PLAESY_HOME/scripts/bash/plaesy-init.sh" "${init_args[@]}"
        fi
        ;;
    "clean")
        # Check if plaesy-clean.sh exists, if not, offer to fix it
        if [ ! -f "$PLAESY_HOME/scripts/bash/plaesy-clean.sh" ]; then
            echo "❌ Clean script missing. This can happen after an upgrade."
            echo "🔧 Fixing by recreating the clean script..."
            echo ""
            echo "📦 Running repair installation..."
            curl -s https://raw.githubusercontent.com/plaesy/spec-kit/main/scripts/bash/install.sh | bash -s -- --repair
            exit 0
        fi

        # Default to current directory if no target specified
        clean_args=("${@:2}")
        if [[ ${#clean_args[@]} -eq 0 ]]; then
            clean_args=(".")
        fi

        # Use config-manager for platform detection
        if [ -f "$PLAESY_HOME/scripts/bash/config-manager.sh" ]; then
            detected_platform=$("$PLAESY_HOME/scripts/bash/config-manager.sh" detect-platform 2>/dev/null)

            if [[ -n "$detected_platform" ]]; then
                echo "✅ Detected AI platform: $detected_platform"
                echo "🧹 Using platform-adapted cleanup..."
                echo "💡 Use: plaesy clean . --ai $detected_platform"
                "$PLAESY_HOME/scripts/bash/plaesy-clean.sh" --ai "$detected_platform" "${clean_args[@]}"
            else
                echo "🤖 No AI platform detected. Starting universal cleanup..."
                echo "💡 Use: plaesy clean . (universal cleanup)"
                "$PLAESY_HOME/scripts/bash/plaesy-clean.sh" "${clean_args[@]}"
            fi
        else
            echo "⚠️  Platform detector not available, using standard cleanup"
            "$PLAESY_HOME/scripts/bash/plaesy-clean.sh" "${clean_args[@]}"
        fi
        ;;
    "analyze")
        # Check if plaesy-analyze.sh exists, if not, offer to fix it
        if [ ! -f "$PLAESY_HOME/scripts/bash/plaesy-analyze.sh" ]; then
            echo "❌ Analysis script missing. This can happen after an upgrade."
            echo "🔧 Fixing by recreating the analysis script..."
            echo ""
            echo "📦 Running repair installation..."
            curl -s https://raw.githubusercontent.com/plaesy/spec-kit/main/scripts/bash/install.sh | bash -s -- --repair
            exit 0
        fi

        # Run project analysis (drop the leading "analyze" subcommand,
        # otherwise plaesy-analyze.sh treats it as the project path)
        analyze_args=("${@:2}")
        if [[ ${#analyze_args[@]} -eq 0 ]]; then
            analyze_args=(".")
        fi
        echo "🔍 Running project analysis..."
        "$PLAESY_HOME/scripts/bash/plaesy-analyze.sh" "${analyze_args[@]}"
        ;;
    "repair"|"fix")
        echo "🔧 Repairing Plaesy Spec-Kit installation..."

        # Create temporary file for the repair script to preserve stdin
        temp_script=$(mktemp) || {
            echo "❌ Failed to create temporary file"
            exit 1
        }

        # Download the script to temporary file
        if ! curl -s https://raw.githubusercontent.com/plaesy/spec-kit/main/scripts/bash/install.sh > "$temp_script"; then
            echo "❌ Failed to download repair script"
            rm -f "$temp_script"
            exit 1
        fi

        # Make it executable and run with proper terminal connection
        chmod +x "$temp_script"
        bash "$temp_script" --repair
        exit_code=$?

        # Clean up
        rm -f "$temp_script"

        # Exit with the same code as the repair script
        exit $exit_code
        ;;
    "upgrade"|"update")
        echo "🔄 Upgrading Plaesy Spec-Kit..."

        # Create temporary file for the upgrade script to preserve stdin
        temp_script=$(mktemp) || {
            echo "❌ Failed to create temporary file"
            exit 1
        }

        # Download the script to temporary file
        if ! curl -s https://raw.githubusercontent.com/plaesy/spec-kit/main/scripts/bash/install.sh > "$temp_script"; then
            echo "❌ Failed to download upgrade script"
            rm -f "$temp_script"
            exit 1
        fi

        # Make it executable and run with proper terminal connection
        chmod +x "$temp_script"
        bash "$temp_script" --upgrade
        exit_code=$?

        # Clean up
        rm -f "$temp_script"

        # Exit with the same code as the upgrade script
        exit $exit_code
        ;;
    "uninstall"|"remove")
        echo "🗑️  Uninstalling Plaesy Spec-Kit..."

        # Create temporary file for the uninstall script to preserve stdin
        temp_script=$(mktemp) || {
            echo "❌ Failed to create temporary file"
            exit 1
        }

        # Download the script to temporary file
        if ! curl -s https://raw.githubusercontent.com/plaesy/spec-kit/main/scripts/bash/install.sh > "$temp_script"; then
            echo "❌ Failed to download uninstall script"
            rm -f "$temp_script"
            exit 1
        fi

        # Make it executable and run with proper terminal connection
        chmod +x "$temp_script"

        # Pass along any additional arguments (like -y, --yes)
        bash "$temp_script" --uninstall "${@:2}"
        exit_code=$?

        # Clean up
        rm -f "$temp_script"

        # Exit with the same code as the uninstall script
        exit $exit_code
        ;;
    "install")
        echo "📦 Installing Plaesy Spec-Kit..."
        curl -s https://raw.githubusercontent.com/plaesy/spec-kit/main/scripts/bash/install.sh | bash
        ;;
    "status")
        echo "🔍 Checking Plaesy Spec-Kit status..."
        echo ""
        
        # Check framework directory
        if [ -d "$PLAESY_HOME" ]; then
            echo "✅ Framework directory: $PLAESY_HOME"
        else
            echo "❌ Framework directory missing: $PLAESY_HOME"
        fi
        
        # Check CLI script
        CLI_PATH="$HOME/.local/bin/plaesy"
        if [ -f "$CLI_PATH" ]; then
            echo "✅ CLI command: $CLI_PATH"
        else
            echo "❌ CLI command missing: $CLI_PATH"
        fi
        
        # Check initialization script
        INIT_SCRIPT="$PLAESY_HOME/scripts/bash/plaesy-init.sh"
        if [ -f "$INIT_SCRIPT" ]; then
            echo "✅ Initialization script: $INIT_SCRIPT"
        else
            echo "❌ Initialization script missing: $INIT_SCRIPT"
            echo "   → Run 'plaesy repair' to fix this"
        fi
        
        # Check version
        if [ -f "$PLAESY_HOME/VERSION" ]; then
            VERSION=$(cat "$PLAESY_HOME/VERSION")
            echo "✅ Version: v$VERSION"
        else
            echo "⚠️  Version file missing (using default v1.0.0)"
        fi
        
        # Check core framework files
        echo ""
        echo "🔍 Core Framework Files:"
        for dir in "instructions" "prompts" "templates" "chatmodes" "checklists"; do
            if [ -d "$PLAESY_HOME/$dir" ]; then
                echo "✅ $dir/"
            else
                echo "❌ $dir/ missing"
            fi
        done
        
        echo ""
        echo "💡 If you see issues above, run 'plaesy repair' to fix them."
        ;;
    "version"|"--version"|"-v")
        if [ -f "$PLAESY_HOME/VERSION" ]; then
            VERSION=$(cat "$PLAESY_HOME/VERSION")
            echo "Plaesy Spec-Kit v$VERSION"
        else
            echo "Plaesy Spec-Kit v1.0.0"
        fi
        echo "Constitutional Development Framework"
        ;;
    "status")
        echo "🏛️  Plaesy Spec-Kit Status"
        echo ""
        if [ -d "$PLAESY_HOME" ]; then
            if [ -f "$PLAESY_HOME/VERSION" ]; then
                VERSION=$(cat "$PLAESY_HOME/VERSION")
                echo "✅ Status: Installed (v$VERSION)"
            else
                echo "✅ Status: Installed (v1.0.0)"
            fi
            echo "📁 Location: $PLAESY_HOME"
            echo "🔧 CLI: $HOME/.local/bin/plaesy"
            
            if [ -f "$PLAESY_HOME/INSTALL_DATE" ]; then
                INSTALL_DATE=$(cat "$PLAESY_HOME/INSTALL_DATE")
                echo "📅 Installed: $INSTALL_DATE"
            fi
            
            echo ""
            echo "📋 Available Commands:"
            echo "  plaesy init          - Initialize new project"
            echo "  plaesy analyze       - Analyze current project structure and generate documentation"
            echo "  plaesy clean         - Remove plaesy directories from project"
            echo "  plaesy upgrade       - Upgrade to latest version"
            echo "  plaesy uninstall     - Remove Plaesy Spec-Kit (use -y or --yes to skip confirmation)"
            echo "  plaesy status        - Show installation status"
        else
            echo "❌ Status: Not installed"
            echo "Run 'plaesy install' to install"
        fi
        ;;
    "help"|"--help"|"-h"|"")
        echo "🏛️  Plaesy Spec-Kit - Constitutional Development Framework"
        echo ""
        echo "Usage: plaesy <command> [options]"
        echo ""
        echo "Commands:"
        echo "  init [directory]     Initialize a new project with constitutional framework"
        echo "  analyze              Analyze current project structure and generate documentation"
        echo "  clean [directory]    Remove plaesy directories from project"
        echo "  install             Install or reinstall Plaesy Spec-Kit"
        echo "  upgrade             Upgrade to the latest version"
        echo "  uninstall [-y]      Remove Plaesy Spec-Kit completely (use -y or --yes to skip confirmation)"
        echo "  repair              Fix missing scripts after upgrade"
        echo "  status              Check installation status and show system information"
        echo "  version             Show version information"
        echo "  help                Show this help message"
        echo ""
        echo "Examples:"
        echo "  plaesy init .                    # Initialize in current directory"
        echo "  plaesy init my-project           # Initialize in new directory"
        echo "  plaesy init . --ai copilot       # Initialize with GitHub Copilot"
        echo "  plaesy init . --ai cursor        # Initialize with Cursor AI"
        echo "  plaesy analyze                   # Analyze current project structure"
        echo "  plaesy clean .                   # Clean current directory"
        echo "  plaesy clean my-project          # Clean specific directory"
        echo "  plaesy upgrade                   # Upgrade to latest version"
        echo "  plaesy repair                    # Fix missing scripts"
        echo "  plaesy status                    # Check installation status"
        echo "  plaesy uninstall                 # Remove completely"
        echo "  plaesy uninstall -y              # Remove without confirmation"
        echo ""
        echo "Learn more: https://github.com/plaesy/spec-kit"
        ;;
    *)
        echo "❌ Unknown command: $1"
        echo "Run 'plaesy help' for usage information"
        exit 1
        ;;
esac
EOF

    chmod +x "$BIN_DIR/plaesy"
    log_success "Plaesy CLI created successfully"
}

create_init_script() {
    log_info "Creating initialization script..."

    mkdir -p "$INSTALL_DIR/scripts"
    mkdir -p "$INSTALL_DIR/scripts/bash"
    mkdir -p "$INSTALL_DIR/scripts/powershell"

    # Scripts are already organized in platform-specific folders
    log_success "Platform-specific scripts organized in scripts/{bash,powershell}/ folders"

    # Make all bash scripts executable
    if [ -d "$INSTALL_DIR/scripts/bash" ]; then
        local bash_count
        bash_count=$(find "$INSTALL_DIR/scripts/bash" -name "*.sh" -type f | wc -l)
        if [ "$bash_count" -gt 0 ]; then
            find "$INSTALL_DIR/scripts/bash" -name "*.sh" -type f -exec chmod +x {} \;
            log_success "Bash scripts ready ($bash_count files)"
        else
            log_warning "No bash scripts found (optional)"
        fi
    fi

    # PowerShell scripts are ready (no chmod needed on Unix)
    if [ -d "$INSTALL_DIR/scripts/powershell" ]; then
        local ps_count
        ps_count=$(find "$INSTALL_DIR/scripts/powershell" -name "*.ps1" -type f | wc -l)
        if [ "$ps_count" -gt 0 ]; then
            log_success "PowerShell scripts ready ($ps_count files)"
        else
            log_warning "No PowerShell scripts found (optional)"
        fi
    fi
}

copy_platform_detector() {
    log_info "Platform detector removed - using config-manager instead"
    # No longer needed - config-manager handles all platform detection
    return 0
}

create_project_structure() {
    log_info "Creating project structure..."
    
    # Create .gitignore if it doesn't exist
    if [ ! -f ".gitignore" ]; then
        cat > ".gitignore" << 'GITIGNORE_EOF'
# Dependencies
node_modules/
vendor/
target/
dist/
build/

# IDE
.vscode/
.idea/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Logs
*.log
logs/

# Environment
.env
.env.local
.env.*.local

# Temporary files
tmp/
temp/
*.tmp

# Plaesy framework (keep template but ignore generated files)
.plaesy/generated/
GITIGNORE_EOF
    fi
    
    log_success "Project structure created"
}

setup_shell_integration() {
    log_info "Setting up shell integration..."

    # Determine the user's actual login shell rc file(s), not the shell
    # currently running this installer (install.sh always runs under bash,
    # so $BASH_VERSION is set even when the user's login shell is zsh).
    local rc_files=()
    case "$(basename "${SHELL:-}")" in
        zsh)
            rc_files=("$HOME/.zshrc")
            ;;
        bash)
            rc_files=("$HOME/.bashrc")
            ;;
        *)
            rc_files=("$HOME/.profile")
            ;;
    esac

    # Also update whichever rc file(s) actually exist for bash/zsh, in case
    # the user switches shells or $SHELL isn't set correctly (e.g. some
    # MSYS2/CI environments).
    for candidate in "$HOME/.bashrc" "$HOME/.zshrc"; do
        if [ -f "$candidate" ]; then
            local already_listed=false
            for rc in "${rc_files[@]}"; do
                [ "$rc" = "$candidate" ] && already_listed=true
            done
            [ "$already_listed" = false ] && rc_files+=("$candidate")
        fi
    done

    for shell_rc in "${rc_files[@]}"; do
        # Create the rc file if missing so the PATH export isn't silently skipped
        [ -f "$shell_rc" ] || touch "$shell_rc"

        if ! grep -q "$BIN_DIR" "$shell_rc"; then
            echo "" >> "$shell_rc"
            echo "# Plaesy Spec-Kit" >> "$shell_rc"
            echo "export PATH=\"$BIN_DIR:\$PATH\"" >> "$shell_rc"
            log_success "Added $BIN_DIR to PATH in $shell_rc"
        fi
    done

    # Reload the current shell's configuration automatically, if possible
    current_rc="$HOME/.$(basename "${SHELL:-bash}")rc"
    if [ -f "$current_rc" ]; then
        log_info "Reloading shell configuration..."
        if source "$current_rc" 2>/dev/null; then
            log_success "Shell configuration reloaded successfully"
        else
            log_warning "Could not reload shell configuration automatically"
            echo "Please run: source $current_rc"
        fi
    else
        log_warning "Could not detect current shell rc file automatically"
        echo "Please run: export PATH=\"$BIN_DIR:\$PATH\" or restart your terminal"
    fi

    log_success "Shell integration configured"
}

print_installation_complete() {
    echo ""
    echo -e "${GREEN}🎉 Plaesy Spec-Kit installation completed successfully!${NC}"
    echo ""
    echo -e "${CYAN}📦 Installation details:${NC}"
    echo "• Framework: $INSTALL_DIR"
    echo "• CLI: $BIN_DIR/plaesy"
    echo "• Version: $VERSION"
    echo ""
    echo -e "${YELLOW}🚀 Quick Start:${NC}"
    echo "1. Create a new project: plaesy init my-project"
    echo "2. Or initialize current directory: plaesy init ."
    echo ""
    echo -e "${BLUE}💡 Usage Examples:${NC}"
    echo "• plaesy init . --ai copilot    # Initialize with GitHub Copilot"
    echo "• plaesy init . --ai cursor     # Initialize with Cursor AI"
    echo "• plaesy init . --ai claude     # Initialize with Claude"
    echo "• plaesy help                   # Show all commands"
    echo ""
    echo -e "${PURPLE}🏛️ Constitutional Development Framework${NC}"
    echo "   Discipline through automation. Quality through structure."
    echo ""
    echo "Learn more: https://github.com/plaesy/spec-kit"
}

# Upgrade functionality
upgrade_framework() {
    print_banner "SPEC-KIT UPGRADER" "Version: $VERSION"
    echo -e "${YELLOW}🔄 Upgrading Plaesy Spec-Kit to latest version...${NC}"
    echo ""

    # Check if already installed
    if [ ! -d "$INSTALL_DIR" ]; then
        log_error "Plaesy Spec-Kit is not installed. Run installation first."
    fi

    # Get current version
    local current_version="1.0.0"
    if [ -f "$INSTALL_DIR/VERSION" ]; then
        current_version=$(cat "$INSTALL_DIR/VERSION")
    fi

    log_info "Current version: v$current_version"
    log_info "Checking for updates..."

    # Create backup before upgrade
    local backup_dir="$HOME/.plaesy-backup-$(date +%Y%m%d-%H%M%S)"
    log_info "Creating backup at $backup_dir..."
    cp -r "$INSTALL_DIR" "$backup_dir" 2>/dev/null || true

    # Preserve user configurations
    local temp_configs="$HOME/.plaesy-configs-temp"
    mkdir -p "$temp_configs"

    # Save any user customizations
    if [ -f "$INSTALL_DIR/user-config.yaml" ]; then
        cp "$INSTALL_DIR/user-config.yaml" "$temp_configs/"
    fi

    # Download latest version
    log_info "Downloading latest framework..."
    rm -rf "$INSTALL_DIR"
    git clone --depth 1 "$REPO_URL" "$INSTALL_DIR" || log_error "Failed to download latest version"
    rm -rf "$INSTALL_DIR/.git"

    # Keep both bash and PowerShell scripts for cross-platform support
    # filter_platform_scripts

    # Update version tracking
    echo "$VERSION" > "$INSTALL_DIR/VERSION"
    echo "$(date)" > "$INSTALL_DIR/UPGRADE_DATE"

    # Restore user configurations
    if [ -f "$temp_configs/user-config.yaml" ]; then
        cp "$temp_configs/user-config.yaml" "$INSTALL_DIR/"
    fi
    rm -rf "$temp_configs"

    # Update CLI and recreate initialization script
    create_plaesy_cli
    create_init_script
    copy_platform_detector

    log_success "Upgrade completed successfully!"
    echo ""
    echo -e "${GREEN}📋 Upgrade Summary:${NC}"
    echo "• Previous version: v$current_version"
    echo "• Current version: v$VERSION"
    echo "• Backup location: $backup_dir"
    echo "• Framework: $INSTALL_DIR"
    echo ""
    echo -e "${YELLOW}💡 Next Steps:${NC}"
    echo "1. Test the upgrade: plaesy version"
    echo "2. Initialize a test project: plaesy init test-project"
    echo "3. If issues occur, restore from backup"
    echo ""
}

# Uninstall functionality
uninstall_framework() {
    local auto_uninstall=false

    # Check for yes flags in all arguments
    if [[ "$*" == *"-y"* ]] || [[ "$*" == *"--yes"* ]]; then
        auto_uninstall=true
    fi

    print_banner "SPEC-KIT UNINSTALLER" "Version: $VERSION"
    echo -e "${RED}🗑️  Uninstalling Plaesy Spec-Kit...${NC}"
    echo ""

    # Check if installed
    if [ ! -d "$INSTALL_DIR" ]; then
        log_warning "Plaesy Spec-Kit is not installed."
        exit 0
    fi

    # Get version info
    local version="1.0.0"
    if [ -f "$INSTALL_DIR/VERSION" ]; then
        version=$(cat "$INSTALL_DIR/VERSION")
    fi

    log_info "Found Plaesy Spec-Kit v$version"
    log_info "Installation directory: $INSTALL_DIR"
    log_info "CLI location: $BIN_DIR/plaesy"

    # Confirmation prompt (skip if auto_uninstall is true)
    if [ "$auto_uninstall" = true ]; then
        log_info "Auto-uninstall mode: proceeding without confirmation..."
    else
        echo ""
        echo -e "${YELLOW}⚠️  This will completely remove Plaesy Spec-Kit from your system.${NC}"
        echo ""
        echo "The following will be removed:"
        echo "• Framework directory: $INSTALL_DIR"
        echo "• CLI command: $BIN_DIR/plaesy"
        echo "• Shell PATH modifications"
        echo ""

        # Check if input is being piped (non-interactive)
        if [ -t 0 ]; then
            # Interactive mode - keep asking until valid input
            while true; do
                read -p "Are you sure you want to continue? [Y/n]: " confirmation
                case "$confirmation" in
                    ""|y|Y|yes|YES)
                        log_info "Proceeding with uninstall..."
                        break
                        ;;
                    n|N|no|NO)
                        echo "❌ Uninstall cancelled."
                        exit 0
                        ;;
                    *)
                        echo "❌ Invalid input. Please enter 'y' or 'n'."
                        echo "   For auto-uninstall, use: plaesy uninstall -y"
                        echo ""
                        ;;
                esac
            done
        else
            # Non-interactive mode (piped input) - check first character of input
            first_char=$(head -c 1)
            case "$first_char" in
                ""|y|Y)
                    log_info "Proceeding with uninstall..."
                    ;;
                n|N)
                    echo "❌ Uninstall cancelled."
                    exit 0
                    ;;
                *)
                    echo "❌ Invalid input. Please enter 'y' or 'n'."
                    echo "   For auto-uninstall, use: plaesy uninstall -y"
                    echo "❌ Uninstall cancelled."
                    exit 1
                    ;;
            esac
        fi
    fi
    
    # Create final backup
    local backup_dir="$HOME/.plaesy-final-backup-$(date +%Y%m%d-%H%M%S)"
    log_info "Creating final backup at $backup_dir..."
    cp -r "$INSTALL_DIR" "$backup_dir" 2>/dev/null || true
    
    # Remove framework directory
    log_info "Removing framework directory..."
    rm -rf "$INSTALL_DIR"
    
    # Remove CLI command
    log_info "Removing CLI command..."
    rm -f "$BIN_DIR/plaesy"
    rm -f "$BIN_DIR/plaesy.bat"  # Windows batch file if exists
    
    # Clean up shell configuration
    log_info "Cleaning up shell configuration..."
    local shell_files=("$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.profile")
    
    for shell_file in "${shell_files[@]}"; do
        if [ -f "$shell_file" ]; then
            # Remove Plaesy Spec-Kit PATH entries
            sed -i '/# Plaesy Spec-Kit/d' "$shell_file" 2>/dev/null || true
            sed -i "\|$BIN_DIR|d" "$shell_file" 2>/dev/null || true
        fi
    done
    
    # Remove from current session PATH
    export PATH=$(echo "$PATH" | sed "s|$BIN_DIR:||g" | sed "s|:$BIN_DIR||g" | sed "s|$BIN_DIR||g")
    
    log_success "Plaesy Spec-Kit uninstalled successfully!"
    echo ""
    echo -e "${GREEN}📋 Uninstall Summary:${NC}"
    echo "• Framework removed from: $INSTALL_DIR"
    echo "• CLI command removed"
    echo "• Shell PATH cleaned up"
    echo "• Final backup saved: $backup_dir"
    echo ""
    echo -e "${CYAN}📝 Note:${NC}"
    echo "• Restart your terminal to complete PATH cleanup"
    echo "• Your project files remain unchanged"
    echo "• Backup is available if you need to restore"
    echo ""
    echo -e "${YELLOW}💡 To reinstall:${NC}"
    echo "curl -s https://raw.githubusercontent.com/plaesy/spec-kit/main/scripts/bash/install.sh | bash"
    echo ""
}

# Main installation flow
main() {
    # Check for upgrade or uninstall flags
    case "$1" in
        "--upgrade"|"--update")
            upgrade_framework
            return
            ;;
        "--uninstall"|"--remove")
            uninstall_framework "$@"
            return
            ;;
        "--repair"|"--fix")
            echo "🔧 Repairing Plaesy Spec-Kit installation..."
            echo ""
            
            # Check if framework directory exists
            if [ ! -d "$INSTALL_DIR" ]; then
                echo "❌ Framework directory missing. Running full installation..."
                # Continue with normal installation flow
            else
                echo "✅ Framework directory found. Recreating missing scripts..."
                
                # Ensure scripts directory exists
                mkdir -p "$INSTALL_DIR/scripts"
                
                # Recreate the initialization script
                create_init_script
                
                # Recreate the CLI command if it's missing
                if [ ! -f "$HOME/.local/bin/plaesy" ]; then
                    create_plaesy_cli
                fi
                
                echo ""
                echo "🎉 Repair completed successfully!"
                echo ""
                echo "You can now use 'plaesy init .' to initialize projects."
                return
            fi
            ;;
    esac
    
    print_banner "SPEC-KIT INSTALLER - Constitutional Development Framework" "Version: $VERSION"
    validate_install_environment
    check_existing_installation
    check_requirements
    download_framework
    # Keep both bash and PowerShell scripts for cross-platform support
    # filter_platform_scripts

    # Save version and install date
    echo "$VERSION" > "$INSTALL_DIR/VERSION"
    echo "$(date)" > "$INSTALL_DIR/INSTALL_DATE"
    
    create_plaesy_cli
    create_init_script
    copy_platform_detector
    setup_shell_integration
    print_installation_complete
}

# Run main installation
main "$@"