# Plaesy Constitution Kit Installer for Windows PowerShell

# Color functions
function Write-Header { Write-Host $args -ForegroundColor Cyan }
function Write-Success { Write-Host "✓ $args" -ForegroundColor Green }
function Write-Warning { Write-Host "⚠️  $args" -ForegroundColor Yellow }
function Write-Error_ { Write-Host "❌ $args" -ForegroundColor Red }
function Write-Info { Write-Host "ℹ️  $args" -ForegroundColor Cyan }

# Display header
Write-Host ""
Write-Host "╔════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║   Plaesy Constitution Kit Installer    ║" -ForegroundColor Cyan
Write-Host "║            For Windows PowerShell      ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

# Detect platform
function Get-Platform {
    Write-Header "Step 1: Detecting platform..."

    $arch = if ([System.Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }

    # Check if ARM64 (Windows on ARM)
    if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
        $arch = "arm64"
    }

    Write-Success "Detected: windows/$arch"
    return $arch
}

# Get latest release version
function Get-LatestVersion {
    Write-Header "Step 2: Fetching latest release..."

    try {
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/plaesy/spec-kit/releases/latest" -ErrorAction Stop
        $version = $response.tag_name

        if (-not $version) {
            Write-Error_ "Failed to parse release information"
            exit 1
        }

        Write-Success "Latest version: $version"
        return $version
    }
    catch {
        Write-Error_ "Failed to fetch latest release. Make sure you have internet connection."
        Write-Info "Error: $_"
        exit 1
    }
}

# Download binary
function Download-Binary {
    param(
        [string]$Architecture,
        [string]$Version
    )

    Write-Header "Step 3: Downloading binary..."

    $binaryName = "plaesy-windows-$Architecture.exe"
    $downloadUrl = "https://github.com/plaesy/spec-kit/releases/download/$Version/$binaryName"
    $tempDir = New-TemporaryDirectory
    $binaryPath = Join-Path $tempDir "plaesy.exe"

    Write-Host "   URL: $downloadUrl"

    try {
        $ProgressPreference = 'SilentlyContinue'
        Invoke-WebRequest -Uri $downloadUrl -OutFile $binaryPath -ErrorAction Stop
        $ProgressPreference = 'Continue'

        Write-Success "Binary downloaded successfully"
        return $binaryPath
    }
    catch {
        Write-Error_ "Failed to download binary"
        Write-Info "Make sure the release is published and your internet is working"
        Write-Info "Error: $_"
        exit 1
    }
}

# Helper: Create temp directory (PowerShell 5.0+ compatible)
function New-TemporaryDirectory {
    $parent = [System.IO.Path]::GetTempPath()
    $name = [System.IO.Path]::GetRandomFileName()
    New-Item -ItemType Directory -Path (Join-Path $parent $name)
}

# Determine install directory
function Get-InstallDirectory {
    Write-Header "Step 4: Preparing installation..."

    $installDir = $null

    # Check if Program Files is writable (preferred)
    if (Test-Path "$env:ProgramFiles\Plaesy" -PathType Container) {
        $installDir = "$env:ProgramFiles\Plaesy"
    }
    elseif ([Security.Principal.WindowsIdentity]::GetCurrent().Groups -contains "S-1-5-32-544") {
        # User is admin - use Program Files
        $installDir = "$env:ProgramFiles\Plaesy"
        if (-not (Test-Path $installDir)) {
            New-Item -ItemType Directory -Path $installDir -Force | Out-Null
        }
    }
    else {
        # User is not admin - use LocalAppData
        $installDir = "$env:LOCALAPPDATA\Plaesy\bin"
        if (-not (Test-Path $installDir)) {
            New-Item -ItemType Directory -Path $installDir -Force | Out-Null
        }
    }

    Write-Success "Install directory: $installDir"
    return $installDir
}

# Install binary
function Install-Binary {
    param(
        [string]$BinaryPath,
        [string]$InstallDir
    )

    Write-Header "Step 5: Installing binary..."

    try {
        if (-not (Test-Path $InstallDir)) {
            New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        }

        $targetPath = Join-Path $InstallDir "plaesy.exe"
        Copy-Item -Path $BinaryPath -Destination $targetPath -Force

        # Verify installation
        $version = & $targetPath --version 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Binary installed"
            return $targetPath
        }
        else {
            Write-Error_ "Failed to verify installation"
            exit 1
        }
    }
    catch {
        Write-Error_ "Failed to install binary: $_"
        exit 1
    }
}

# Check and update PATH
function Update-PathIfNeeded {
    param([string]$InstallDir)

    Write-Header "Step 6: Verifying PATH..."

    $pathArray = $env:PATH -split ";"
    $isInPath = $pathArray -contains $InstallDir

    if ($isInPath) {
        Write-Success "$InstallDir is in PATH"
        return $true
    }
    else {
        Write-Warning "$InstallDir is NOT in PATH"
        Write-Host ""
        Write-Info "To add to PATH for current session, run:"
        Write-Host "`$env:PATH = `"$InstallDir;`$env:PATH`"" -ForegroundColor Blue
        Write-Host ""
        Write-Info "To add permanently, run this in PowerShell (as Admin):"
        Write-Host "[Environment]::SetEnvironmentVariable('PATH', `"$InstallDir;`$env:PATH`", 'User')" -ForegroundColor Blue
        return $false
    }
}

# Main installation flow
function Install-Plaesy {
    try {
        # Step 1: Detect platform
        $architecture = Get-Platform
        Write-Host ""

        # Step 2: Get latest version
        $version = Get-LatestVersion
        Write-Host ""

        # Step 3: Download binary
        $binaryPath = Download-Binary -Architecture $architecture -Version $version
        Write-Host ""

        # Step 4: Get install directory
        $installDir = Get-InstallDirectory
        Write-Host ""

        # Step 5: Install binary
        $installedPath = Install-Binary -BinaryPath $binaryPath -InstallDir $installDir
        Write-Host ""

        # Step 6: Check PATH
        $pathOk = Update-PathIfNeeded -InstallDir $installDir
        Write-Host ""

        # Cleanup
        Remove-Item -Path (Split-Path $binaryPath) -Recurse -Force -ErrorAction SilentlyContinue

        # Success message
        Write-Host "╔════════════════════════════════════════╗" -ForegroundColor Green
        Write-Host "║   ✓ Installation Successful!           ║" -ForegroundColor Green
        Write-Host "╚════════════════════════════════════════╝" -ForegroundColor Green
        Write-Host ""

        Write-Host "Next steps:" -ForegroundColor Yellow
        Write-Host "  1. Verify installation:" -ForegroundColor Yellow
        if (-not $pathOk) {
            Write-Host "     `$env:PATH = `"$installDir;`$env:PATH`"" -ForegroundColor Blue
        }
        Write-Host "     plaesy --version" -ForegroundColor Blue
        Write-Host ""
        Write-Host "  2. Initialize a new project:" -ForegroundColor Yellow
        Write-Host "     plaesy init my-awesome-app" -ForegroundColor Blue
        Write-Host ""
        Write-Host "  3. Or analyze an existing project:" -ForegroundColor Yellow
        Write-Host "     plaesy analyze" -ForegroundColor Blue
        Write-Host ""
        Write-Host "Documentation:" -ForegroundColor Yellow
        Write-Host "  https://github.com/plaesy/spec-kit#readme" -ForegroundColor Blue
    }
    catch {
        Write-Error_ "Installation failed: $_"
        exit 1
    }
}

# Run installation
Install-Plaesy
