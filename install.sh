#!/usr/bin/env bash
# Plaesy Constitution Kit installer for Linux, macOS, and Git Bash on Windows.
#
#   curl -fsSL https://raw.githubusercontent.com/plaesy/spec-kit/main/install.sh | bash
#
# Downloads the release binary, verifies it against SHA256SUMS from the same
# release, and runs `plaesy install`, which picks the install directory and
# adds the install directory to the bash or zsh profile if it is not already
# there. Other shells get printed instructions.
#
# Environment:
#   PLAESY_VERSION   release to install, e.g. v0.0.1 (default: latest)
#   PLAESY_BASE_URL  directory holding the release assets; overrides the GitHub
#                    URL (for testing or a mirror)
set -euo pipefail

REPO_URL="https://github.com/plaesy/spec-kit"

die() {
    echo "error: $*" >&2
    exit 1
}

# Print the release asset name for this machine.
detect_asset() {
    local os arch
    case "$(uname -s)" in
        Linux*) os="linux" ;;
        Darwin*) os="darwin" ;;
        MINGW*|MSYS*|CYGWIN*) os="windows" ;;
        *) die "unsupported OS: $(uname -s)" ;;
    esac
    case "$(uname -m)" in
        x86_64|amd64) arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        *) die "unsupported architecture: $(uname -m)" ;;
    esac

    case "$os-$arch" in
        linux-amd64|linux-arm64|darwin-arm64) echo "plaesy-$os-$arch" ;;
        windows-amd64) echo "plaesy-windows-amd64.exe" ;;
        darwin-amd64) die "no release binary for Intel Macs; build from source (see README)" ;;
        *) die "no release binary for $os/$arch" ;;
    esac
}

# Print the URL directory the assets are downloaded from.
resolve_base_url() {
    if [ -n "${PLAESY_BASE_URL:-}" ]; then
        echo "${PLAESY_BASE_URL%/}"
        return
    fi
    local version="${PLAESY_VERSION:-latest}"
    if [ "$version" = "latest" ]; then
        echo "$REPO_URL/releases/latest/download"
        return
    fi
    if [[ ! "$version" =~ ^v?[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        die "PLAESY_VERSION must look like v1.2.3 or 'latest', got: $version"
    fi
    [[ "$version" == v* ]] || version="v$version"
    echo "$REPO_URL/releases/download/$version"
}

sha256_of() {
    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum "$1" | cut -d' ' -f1
    elif command -v shasum >/dev/null 2>&1; then
        shasum -a 256 "$1" | cut -d' ' -f1
    else
        die "need sha256sum or shasum to verify the download"
    fi
}

# Fails closed: no checksum entry for the asset means no install.
verify_checksum() {
    local file="$1" sums="$2" asset="$3" expected actual
    expected=$(awk -v name="$asset" '{ sub(/^\*/, "", $2) } $2 == name { print $1 }' "$sums")
    [ -n "$expected" ] || die "no checksum for $asset in SHA256SUMS; refusing to install"
    actual=$(sha256_of "$file")
    [ "$actual" = "$expected" ] || die "checksum mismatch for $asset: expected $expected, got $actual"
    echo "Checksum verified (SHA-256)"
}

# Global so the EXIT trap can still see it after main returns.
TMP_DIR=""
cleanup() {
    if [ -n "$TMP_DIR" ]; then
        rm -rf "$TMP_DIR"
    fi
}

# Appends an export line for dir to the profile of bash or zsh, unless dir is
# already on PATH or already mentioned in that profile. Other shells are left
# alone and get a manual instruction instead.
add_to_path() {
    local dir="$1" rc
    case ":$PATH:" in
        *":$dir:"*) echo "$dir is already on your PATH"; return ;;
    esac
    case "${SHELL##*/}" in
        zsh) rc="$HOME/.zshrc" ;;
        bash)
            if [ "$(uname -s)" = "Darwin" ]; then rc="$HOME/.bash_profile"; else rc="$HOME/.bashrc"; fi
            ;;
        *)
            echo "Add $dir to your PATH manually (unsupported shell: ${SHELL:-unknown})"
            return
            ;;
    esac
    if grep -qF "$dir" "$rc" 2>/dev/null; then
        echo "$dir is already referenced in $rc; open a new terminal to use plaesy"
        return
    fi
    printf '\n# Added by the plaesy installer\nexport PATH="%s:$PATH"\n' "$dir" >> "$rc" \
        || die "could not write to $rc"
    echo "Added $dir to $rc. Open a new terminal to use plaesy."
}

main() {
    local asset base bin
    asset=$(detect_asset)
    base=$(resolve_base_url)

    TMP_DIR=$(mktemp -d)
    trap cleanup EXIT
    bin="$TMP_DIR/${asset}"

    echo "Downloading $asset from $base"
    curl -fsSL --tlsv1.2 -o "$bin" "$base/$asset" || die "download failed: $base/$asset"
    curl -fsSL --tlsv1.2 -o "$TMP_DIR/SHA256SUMS" "$base/SHA256SUMS" \
        || die "could not download SHA256SUMS; refusing to install an unverified binary"
    verify_checksum "$bin" "$TMP_DIR/SHA256SUMS" "$asset"

    chmod +x "$bin"
    "$bin" install
    add_to_path "$HOME/.local/bin"
}

main "$@"
