# Plaesy Constitution Kit Installer for Windows PowerShell
# Mirrors scripts/install.sh: downloads the latest release binary
# (no git clone, no Go toolchain required).
#
# Usage:
#   iwr -useb https://raw.githubusercontent.com/plaesy/spec-kit/main/scripts/install.ps1 | iex
#
# Or save and run:
#   Invoke-WebRequest -Uri https://raw.githubusercontent.com/plaesy/spec-kit/main/scripts/install.ps1 -OutFile install.ps1
#   powershell -ExecutionPolicy Bypass -File install.ps1

param(
    [string]$InstallDir = "",
    [switch]$Force
)

$ErrorActionPreference = "Stop"

$Repo = "plaesy/spec-kit"
$BinaryName = "plaesy"

function Write-Step  { param([string]$Msg) Write-Host "$Msg" -ForegroundColor Blue }
function Write-Success { param([string]$Msg) Write-Host "$Msg" -ForegroundColor Green }
function Write-Warn  { param([string]$Msg) Write-Host "$Msg" -ForegroundColor Yellow }
function Write-Err   { param([string]$Msg) Write-Host "$Msg" -ForegroundColor Red; exit 1 }

function Detect-OS {
    if ($IsWindows -or $env:OS -eq "Windows_NT") { return "windows" }
    if ($IsMacOS) { return "darwin" }
    if ($IsLinux) { return "linux" }
    return "unknown"
}

function Detect-Arch {
    switch ($env:PROCESSOR_ARCHITECTURE) {
        "AMD64"  { return "amd64" }
        "ARM64"  { return "arm64" }
    }
    if ($IsMacOS -or $IsLinux) {
        switch ((uname -m 2>/dev/null)) {
            "x86_64" { return "amd64" }
            "aarch64" { return "arm64" }
        }
    }
    return "unknown"
}

function Get-InstallDir {
    param([string]$OS)
    if ($InstallDir) { return $InstallDir }
    if ($OS -eq "windows") {
        $base = $env:LOCALAPPDATA
        if (-not $base) {
            $base = "$env:USERPROFILE\AppData\Local"
        }
        return "$base\Plaesy\bin"
    }
    # macOS / Linux
    return "$env:HOME/.local/bin"
}

function Test-OnPath {
    param([string]$Dir)
    $target = (Resolve-Path $Dir -ErrorAction SilentlyContinue)
    if (-not $target) {
        $target = $Dir
    }
    foreach ($entry in $env:PATH -split [System.IO.Path]::PathSeparator) {
        if ([string]::IsNullOrWhiteSpace($entry)) { continue }
        try {
            $cleanEntry = (Resolve-Path $entry -ErrorAction SilentlyContinue).Path
            if (-not $cleanEntry) { $cleanEntry = $entry }
            if ($cleanEntry -eq $target) { return $true }
        } catch { }
    }
    return $false
}

function Get-LatestTag {
    Write-Step "Fetching latest release..."
    $releasesUrl = "https://api.github.com/repos/$Repo/releases/latest"
    try {
        $response = Invoke-RestMethod -Uri $releasesUrl -UseBasicParsing -ErrorAction Stop
        $tag = $response.tag_name
        if (-not $tag) {
            Write-Err "ERROR: Could not find latest release. Check: https://github.com/$Repo/releases"
        }
        # Strip leading 'v' if present
        $tag = $tag -replace '^v', ''
        return $tag
    } catch {
        Write-Err "ERROR: Failed to fetch latest release. $_"
    }
}

function Invoke-Install {
    $os = Detect-OS
    $arch = Detect-Arch

    if ($os -eq "unknown" -or $arch -eq "unknown") {
        Write-Err "ERROR: Unsupported OS ($os) or architecture ($arch)"
        Write-Err "Supported: Linux/macOS/Windows on amd64 and arm64"
    }

    $tag = Get-LatestTag

    $suffix = ""
    if ($os -eq "windows") { $suffix = ".exe" }

    $url = "https://github.com/$Repo/releases/download/v$tag/$BinaryName-$os-$arch$suffix"
    $dest = Join-Path (Get-InstallDir $os) "$BinaryName$suffix"

    Write-Step "Installing plaesy v$tag ($os/$arch)..."
    Write-Host "  Binary: $url"
    Write-Host "  Install: $dest"

    # Create install directory
    $destDir = Split-Path $dest -Parent
    if (-not (Test-Path $destDir)) {
        New-Item -ItemType Directory -Path $destDir -Force | Out-Null
    }

    # Download
    Write-Step "Downloading..."
    try {
        Invoke-WebRequest -Uri $url -OutFile $dest -UseBasicParsing -ErrorAction Stop
    } catch {
        Write-Err "ERROR: Download failed"
        Write-Err "  URL: $url"
        Write-Err "  Check: https://github.com/$Repo/releases"
    }

    # Verify
    if (-not (Test-Path $dest)) {
        Write-Err "ERROR: Installation failed — file not found"
    }

    $destClean = (Get-InstallDir $os)
    $destClean = $destClean -replace '\\\\', '\'

    Write-Host ""
    Write-Success "✅ plaesy v$tag installed to $dest"
    Write-Host ""

    # Check PATH
    if (Test-OnPath $destClean) {
        Write-Host "Run 'plaesy' to verify the installation."
    } else {
        Write-Warn "⚠️  $destClean is not on your PATH."
        if ($os -eq "windows") {
            Write-Host "Add it to PATH in PowerShell:"
            Write-Host "  [Environment]::SetEnvironmentVariable('Path', `$env:Path + ';$destClean', 'User')"
            Write-Host "Then open a new terminal."
        } else {
            Write-Host "Add it to your shell config:"
            Write-Host "  echo 'export PATH=`"$HOME/.local/bin:$PATH`"' >> ~/.bashrc"
            Write-Host "Then run: source ~/.bashrc"
        }
    }

    Write-Host ""
    Write-Host "Next steps:"
    Write-Host "  plaesy init my-project   # Set up Plaesy for a new project"
}

Invoke-Install
