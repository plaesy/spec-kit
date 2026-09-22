#!/usr/bin/env bash
#
# One-line install script for Plaesy Constitution Kit:
#   curl -fsSL https://raw.githubusercontent.com/plaesy/spec-kit/main/scripts/install.sh | sh
#
# Downloads the latest release binary and installs it to a well-known bin dir.
# Requires no Go toolchain, no git — just curl and sh.
set -euo pipefail

REPO="plaesy/spec-kit"
INSTALL_DIR=""
BINARY_NAME="plaesy"

# Detect OS
detect_os() {
  local os
  os="$(uname -s 2>/dev/null || echo "unknown")"
  case "$os" in
    Linux*)   echo "linux" ;;
    Darwin*)  echo "darwin" ;;
    MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
    *)        echo "unknown" ;;
  esac
}

# Detect architecture
detect_arch() {
  local arch
  arch="$(uname -m 2>/dev/null || echo "unknown")"
  case "$arch" in
    x86_64|amd64) echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *)            echo "unknown" ;;
  esac
}

# Determine install directory (matches Go installer's InstallDir())
detect_install_dir() {
  local os="$1"
  if [ "$os" = "windows" ]; then
    if [ -n "${LOCALAPPDATA:-}" ]; then
      echo "$LOCALAPPDATA/Plaesy/bin"
    else
      echo "$HOME/AppData/Local/Plaesy/bin"
    fi
  else
    echo "$HOME/.local/bin"
  fi
}

# Check if a directory is in PATH
on_path() {
  local dir="$1"
  case ":${PATH}:" in
    *":$dir:"*) return 0 ;;
    *)          return 1 ;;
  esac
}

main() {
  local os arch tag url suffix dest

  os="$(detect_os)"
  arch="$(detect_arch)"

  if [ "$os" = "unknown" ] || [ "$arch" = "unknown" ]; then
    echo "ERROR: Unsupported OS ($os) or architecture ($arch)" >&2
    echo "Supported: Linux/macOS/Windows on amd64 and arm64" >&2
    exit 1
  fi

  # Get latest release tag
  echo "Fetching latest release..."
  tag="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep -o '"tag_name": *"[^"]*"' \
    | head -1 \
    | sed 's/.*"v\(.*\)"/\1/')"

  if [ -z "$tag" ]; then
    echo "ERROR: Could not find latest release" >&2
    echo "Check: https://github.com/${REPO}/releases" >&2
    exit 1
  fi

  # Build download URL
  suffix=""
  if [ "$os" = "windows" ]; then
    suffix=".exe"
  fi
  url="https://github.com/${REPO}/releases/download/v${tag}/plaesy-${os}-${arch}${suffix}"

  # Determine install directory
  INSTALL_DIR="${INSTALL_DIR:-$(detect_install_dir "$os")}"
  dest="${INSTALL_DIR}/${BINARY_NAME}${suffix}"

  echo "Installing plaesy v${tag} (${os}/${arch})..."
  echo "  Binary: $url"
  echo "  Install: $dest"

  mkdir -p "$INSTALL_DIR"

  # Download
  if ! curl -fsSL "$url" -o "$dest"; then
    echo "ERROR: Download failed" >&2
    echo "  URL: $url" >&2
    echo "  Check: https://github.com/${REPO}/releases" >&2
    exit 1
  fi

  if [ "$os" != "windows" ]; then
    chmod +x "$dest"
  fi

  # Verify
  if [ ! -f "$dest" ]; then
    echo "ERROR: Installation failed — file not found" >&2
    exit 1
  fi

  echo ""
  echo "✅ plaesy v${tag} installed to $dest"
  echo ""

  # Check PATH
  if on_path "$INSTALL_DIR"; then
    echo "Run 'plaesy' to verify the installation."
  else
    echo "⚠️  $INSTALL_DIR is not on your PATH."
    if [ "$os" = "windows" ]; then
      echo "Add it to PATH in PowerShell:"
      echo "  [Environment]::SetEnvironmentVariable('Path', \$env:Path + ';${INSTALL_DIR}', 'User')"
      echo "Then open a new terminal."
    else
      echo "Add it to your shell config:"
      echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc"
      echo "Then run: source ~/.bashrc"
    fi
  fi

  echo ""
  echo "Next steps:"
  echo "  plaesy init my-project   # Set up Plaesy for a new project"
}

main "$@"
