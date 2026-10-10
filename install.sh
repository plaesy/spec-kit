#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   Plaesy Constitution Kit Installer    ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"
echo ""

# Detect OS and Architecture
detect_platform() {
    local os=""
    local arch=""

    case "$(uname -s)" in
        Linux*)     os="linux";;
        Darwin*)    os="darwin";;
        MINGW*|MSYS*|CYGWIN*) os="windows";;
        *)          os="unknown";;
    esac

    case "$(uname -m)" in
        x86_64)     arch="amd64";;
        arm64|aarch64) arch="arm64";;
        *)          arch="unknown";;
    esac

    if [ "$os" = "unknown" ] || [ "$arch" = "unknown" ]; then
        echo -e "${RED}❌ Unsupported platform: $(uname -s) $(uname -m)${NC}" >&2
        exit 1
    fi

    echo -e "${GREEN}✓${NC} Detected: $os/$arch" >&2
    echo "$os-$arch"
}

# Get latest release version
get_latest_version() {
    echo -e "${YELLOW}⏳ Fetching latest release...${NC}" >&2

    local version=$(curl -s https://api.github.com/repos/plaesy/spec-kit/releases/latest \
        | grep '"tag_name"' \
        | head -1 \
        | cut -d'"' -f4)

    if [ -z "$version" ]; then
        echo -e "${RED}❌ Failed to fetch latest release. Make sure you have internet connection.${NC}" >&2
        exit 1
    fi

    echo -e "${GREEN}✓${NC} Latest version: $version" >&2
    echo "$version"
}

# Download binary
download_binary() {
    local platform=$1
    local version=$2
    local binary_name="plaesy-${platform}"
    local binary_filename="plaesy"
    if [[ "$platform" == windows-* ]]; then
        binary_name="${binary_name}.exe"
        binary_filename="plaesy.exe"
    fi
    local download_url="https://github.com/plaesy/spec-kit/releases/download/${version}/${binary_name}"
    local temp_dir=$(mktemp -d)
    local binary_path="${temp_dir}/${binary_filename}"

    echo -e "${YELLOW}⏳ Downloading binary...${NC}" >&2
    echo "   URL: $download_url" >&2

    if ! curl -fsSL -o "$binary_path" "$download_url"; then
        echo -e "${RED}❌ Failed to download binary from: $download_url${NC}" >&2
        echo -e "${YELLOW}💡 Tip: Make sure the release is published and your internet is working.${NC}" >&2
        rm -rf "$temp_dir"
        exit 1
    fi

    chmod +x "$binary_path"
    echo -e "${GREEN}✓${NC} Binary downloaded successfully" >&2
    echo "$binary_path"
}

# Determine install directory
get_install_dir() {
    local platform=$1
    local install_dir=""

    # Windows (Git Bash/MSYS/Cygwin): no sudo, no /usr/local/bin convention.
    if [[ "$platform" == windows-* ]]; then
        install_dir="$HOME/bin"
        echo "$install_dir"
        return
    fi

    # Check if /usr/local/bin exists and is writable
    if [ -d "/usr/local/bin" ] && [ -w "/usr/local/bin" ]; then
        install_dir="/usr/local/bin"
    # Check if ~/.local/bin exists and is writable
    elif [ -d "$HOME/.local/bin" ] && [ -w "$HOME/.local/bin" ]; then
        install_dir="$HOME/.local/bin"
    # Fall back to /usr/local/bin and ask for sudo
    elif [ -d "/usr/local/bin" ]; then
        install_dir="/usr/local/bin"
        echo -e "${YELLOW}⚠️  /usr/local/bin requires sudo access${NC}" >&2
    else
        install_dir="$HOME/.local/bin"
    fi

    echo "$install_dir"
}

# Install binary
install_binary() {
    local binary_path=$1
    local install_dir=$2
    local platform=$3
    local binary_filename="plaesy"
    if [[ "$platform" == windows-* ]]; then
        binary_filename="plaesy.exe"
    fi
    local target_path="$install_dir/$binary_filename"

    echo -e "${YELLOW}⏳ Installing to $install_dir...${NC}"

    # Create directory if it doesn't exist
    if [ ! -d "$install_dir" ]; then
        mkdir -p "$install_dir"
    fi

    # Windows has no sudo; just copy directly.
    if [[ "$platform" == windows-* ]]; then
        if ! cp "$binary_path" "$target_path"; then
            echo -e "${RED}❌ Failed to install binary${NC}"
            exit 1
        fi
        echo -e "${GREEN}✓${NC} Binary installed"
    # Determine if we need sudo
    elif [ ! -w "$install_dir" ]; then
        if ! sudo cp "$binary_path" "$target_path"; then
            echo -e "${RED}❌ Failed to install binary (permission denied)${NC}"
            exit 1
        fi
        echo -e "${GREEN}✓${NC} Binary installed with sudo"
    else
        if ! cp "$binary_path" "$target_path"; then
            echo -e "${RED}❌ Failed to install binary${NC}"
            exit 1
        fi
        echo -e "${GREEN}✓${NC} Binary installed"
    fi

    # Verify installation (the binary has no --version flag; just confirm it runs)
    if ! "$target_path" --help >/dev/null 2>&1; then
        echo -e "${RED}❌ Failed to verify installation${NC}"
        exit 1
    fi
}

# Check and update PATH
check_path() {
    local install_dir=$1

    if echo "$PATH" | grep -q "$install_dir"; then
        echo -e "${GREEN}✓${NC} $install_dir is in PATH"
        return 0
    else
        echo -e "${YELLOW}⚠️  $install_dir is NOT in PATH${NC}"
        return 1
    fi
}

# Main installation flow
main() {
    # Step 1: Detect platform
    echo -e "${BLUE}Step 1: Detecting platform...${NC}"
    platform=$(detect_platform)
    echo ""

    # Step 2: Get latest version
    echo -e "${BLUE}Step 2: Fetching latest release...${NC}"
    version=$(get_latest_version)
    echo ""

    # Step 3: Download binary
    echo -e "${BLUE}Step 3: Downloading binary...${NC}"
    binary_path=$(download_binary "$platform" "$version")
    echo ""

    # Step 4: Get install directory
    echo -e "${BLUE}Step 4: Preparing installation...${NC}"
    install_dir=$(get_install_dir "$platform")
    echo -e "${GREEN}✓${NC} Install directory: $install_dir"
    echo ""

    # Step 5: Install binary
    echo -e "${BLUE}Step 5: Installing binary...${NC}"
    install_binary "$binary_path" "$install_dir" "$platform"
    echo ""

    # Step 6: Verify PATH
    echo -e "${BLUE}Step 6: Verifying PATH...${NC}"
    if ! check_path "$install_dir"; then
        echo -e "${YELLOW}ℹ️  Add this line to your shell config (~/.bashrc, ~/.zshrc, etc.):${NC}"
        echo "   export PATH=\"$install_dir:\$PATH\""
    fi
    echo ""

    # Cleanup
    rm -rf "$(dirname "$binary_path")"

    # Success message
    echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║   ✓ Installation Successful!           ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${YELLOW}Next steps:${NC}"
    echo "  1. Verify installation:"
    echo -e "     ${BLUE}plaesy --help${NC}"
    echo ""
    echo "  2. Initialize a new project:"
    echo -e "     ${BLUE}plaesy init my-awesome-app${NC}"
    echo ""
    echo "  3. Or analyze an existing project:"
    echo -e "     ${BLUE}plaesy analyze${NC}"
    echo ""
    echo -e "${YELLOW}Documentation:${NC}"
    echo "  https://github.com/plaesy/spec-kit#readme"
}

main
