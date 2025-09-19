# Plaesy Spec-Kit Installer for Windows PowerShell
# Constitutional Development Framework
# https://github.com/plaesy/spec-kit

param(
    [string]$InstallPath = $(if ($IsWindows -or $env:OS -eq "Windows_NT") { "$env:USERPROFILE\.plaesy" } else { "$env:HOME/.plaesy" }),
    [switch]$Force
)

# Error handling
$ErrorActionPreference = "Stop"

# Configuration
$RepoUrl = "https://github.com/plaesy/spec-kit"
$Version = "1.0.0"
$BinDir = $(if ($IsWindows -or $env:OS -eq "Windows_NT") { "$env:USERPROFILE\.local\bin" } else { "$env:HOME/.local/bin" })

# Validate critical variables
function Test-Configuration {
    if ([string]::IsNullOrWhiteSpace($InstallPath)) {
        Write-Error-Custom "InstallPath parameter is empty or not set"
    }
    
    $homeVar = if ($IsWindows -or $env:OS -eq "Windows_NT") { $env:USERPROFILE } else { $env:HOME }
    if ([string]::IsNullOrWhiteSpace($homeVar)) {
        Write-Error-Custom "Home directory environment variable is not set"
    }
    
    if ([string]::IsNullOrWhiteSpace($BinDir)) {
        Write-Error-Custom "BinDir is not properly configured"
    }
    
    # Test if we can create the parent directory
    $parentDir = Split-Path $InstallPath -Parent
    if (![string]::IsNullOrWhiteSpace($parentDir)) {
        try {
            if (!(Test-Path $parentDir)) {
                New-Item -ItemType Directory -Path $parentDir -Force -ErrorAction Stop | Out-Null
                Remove-Item -Path $parentDir -Force -ErrorAction Stop
            }
        }
        catch {
            Write-Error-Custom "Cannot access or create parent directory: $parentDir - $($_.Exception.Message)"
        }
    }
}

# Colors for output (limited support in older PowerShell versions)
function Write-Step { 
    param([string]$Message)
    Write-Host "$Message" -ForegroundColor Blue
}

function Write-Success { 
    param([string]$Message)
    Write-Host "$Message" -ForegroundColor Green
}

function Write-Warning { 
    param([string]$Message)
    Write-Host "$Message" -ForegroundColor Yellow
}

function Write-Error-Custom { 
    param([string]$Message)
    Write-Host "$Message" -ForegroundColor Red
    exit 1
}

function Write-Banner {
    Write-Host ""
    Write-Host "Plaesy Spec-Kit Installer" -ForegroundColor Magenta
    Write-Host "   Constitutional Development Framework" -ForegroundColor Magenta
    Write-Host "   Version: $Version" -ForegroundColor Magenta
    Write-Host ""
}

function Test-Prerequisites {
    Write-Step "Checking system requirements..."
    
    # Check PowerShell version
    if ($PSVersionTable.PSVersion.Major -lt 5) {
        Write-Error-Custom "PowerShell 5.0 or higher is required. Current version: $($PSVersionTable.PSVersion)"
    }
    
    # Check for git
    try {
        git --version | Out-Null
    }
    catch {
        Write-Error-Custom "Git is required but not found. Please install Git first."
    }
    
    Write-Success "System requirements satisfied"
}

function Install-Framework {
    Write-Step "Downloading Plaesy Spec-Kit framework..."
    
    # Validate InstallPath is not empty
    if ([string]::IsNullOrWhiteSpace($InstallPath)) {
        Write-Error-Custom "InstallPath is empty or not set"
    }
    
    # Remove existing installation if Force is specified
    if (Test-Path $InstallPath) {
        if ($Force) {
            Write-Warning "Removing existing installation..."
            try {
                Remove-Item -Path $InstallPath -Recurse -Force -ErrorAction Stop
            }
            catch {
                Write-Error-Custom "Failed to remove existing installation: $($_.Exception.Message)"
            }
        }
        else {
            Write-Warning "Installation already exists at $InstallPath"
            $response = Read-Host "Remove existing installation? (y/N)"
            if ($response -eq 'y' -or $response -eq 'Y') {
                try {
                    Remove-Item -Path $InstallPath -Recurse -Force -ErrorAction Stop
                }
                catch {
                    Write-Error-Custom "Failed to remove existing installation: $($_.Exception.Message)"
                }
            }
            else {
                Write-Error-Custom "Installation aborted. Use -Force to overwrite automatically."
            }
        }
    }
    
    # Create parent directory
    $parentDir = Split-Path $InstallPath -Parent
    if (![string]::IsNullOrWhiteSpace($parentDir) -and !(Test-Path $parentDir)) {
        try {
            New-Item -ItemType Directory -Path $parentDir -Force | Out-Null
        }
        catch {
            Write-Error-Custom "Failed to create parent directory: $($_.Exception.Message)"
        }
    }
    
    # Clone the repository
    try {
        git clone --depth 1 $RepoUrl $InstallPath
        
        # Remove .git directory to save space
        $gitPath = Join-Path -Path $InstallPath -ChildPath ".git"
        if (Test-Path $gitPath) {
            Remove-Item -Path $gitPath -Recurse -Force -ErrorAction SilentlyContinue
        }
        
        Write-Success "Framework downloaded successfully"
    }
    catch {
        Write-Error-Custom "Failed to download framework: $($_.Exception.Message)"
    }
}

<#
.SYNOPSIS
    Platform-specific script filtering for optimized Windows installation

.DESCRIPTION
    This function removes scripts that are not needed for the Windows platform
    to reduce installation size and improve performance.
    
    For PowerShell installer (Windows):
    - Removes: scripts\bash directory and all .sh files
    - Keeps: scripts\powershell directory and all .ps1 files  
    - Result: ~75% smaller script footprint, optimized for Windows

.NOTES
    Called automatically during both installation and upgrade processes.
    No user intervention required.
#>
function Remove-NonPlatformScripts {
    Write-Step "Filtering platform-specific scripts..."
    
    # Validate InstallPath is not empty
    if ([string]::IsNullOrWhiteSpace($InstallPath)) {
        Write-Warning "InstallPath is empty, skipping script filtering"
        return
    }
    
    # Validate that InstallPath exists
    if (!(Test-Path $InstallPath)) {
        Write-Warning "InstallPath does not exist: $InstallPath, skipping script filtering"
        return
    }
    
    # Remove bash scripts since this is a PowerShell installer
    $bashScriptsPath = Join-Path -Path $InstallPath -ChildPath "scripts\bash"
    if (Test-Path $bashScriptsPath) {
        try {
            Remove-Item -Path $bashScriptsPath -Recurse -Force -ErrorAction Stop
            Write-Success "Removed bash scripts directory (not needed for this platform)"
        }
        catch {
            Write-Warning "Could not remove bash scripts directory: $($_.Exception.Message)"
        }
    }
    
    # Remove individual .sh files from scripts directory
    $scriptsPath = Join-Path -Path $InstallPath -ChildPath "scripts"
    if (Test-Path $scriptsPath) {
        try {
            $shFiles = Get-ChildItem -Path $scriptsPath -Filter "*.sh" -ErrorAction SilentlyContinue
            if ($shFiles) {
                foreach ($file in $shFiles) {
                    try {
                        Remove-Item -Path $file.FullName -Force -ErrorAction Stop
                    }
                    catch {
                        Write-Warning "Could not remove bash script: $($file.FullName)"
                    }
                }
                Write-Success "Removed individual bash scripts (not needed for this platform)"
            }
        }
        catch {
            Write-Warning "Could not scan for bash scripts: $($_.Exception.Message)"
        }
    }
    
    # Keep only PowerShell scripts
    $powershellScriptsPath = Join-Path -Path $InstallPath -ChildPath "scripts\powershell"
    if (Test-Path $powershellScriptsPath) {
        Write-Success "Keeping PowerShell scripts for this platform"
    }
    
    Write-Success "Platform-specific script filtering completed"
}

function New-PleasyCLI {
    Write-Step "Creating plaesy CLI command..."
    
    # Ensure bin directory exists
    if (!(Test-Path $BinDir)) {
        New-Item -ItemType Directory -Path $BinDir -Force | Out-Null
    }
    
    # Create batch file for Windows
    $batchContent = @"
@echo off
setlocal

set "PLAESY_HOME=%USERPROFILE%\.plaesy"

if not exist "%PLAESY_HOME%" if not "%1"=="install" (
    echo Plaesy framework not found. Please run the installer first.
    echo iwr -useb https://raw.githubusercontent.com/plaesy/spec-kit/main/scripts/powershell/install.ps1 ^| iex
    exit /b 1
)

if "%1"=="init" (
    powershell -ExecutionPolicy Bypass -File "%PLAESY_HOME%\scripts\powershell\plaesy-init.ps1" %*
    goto :eof
)

if "%1"=="clean" (
    powershell -ExecutionPolicy Bypass -File "%PLAESY_HOME%\scripts\powershell\plaesy-clean.ps1" %*
    goto :eof
)

if "%1"=="upgrade" goto :upgrade
if "%1"=="update" goto :upgrade

if "%1"=="uninstall" goto :uninstall
if "%1"=="remove" goto :uninstall

if "%1"=="install" (
    powershell -ExecutionPolicy Bypass -Command "iwr -useb https://raw.githubusercontent.com/plaesy/spec-kit/main/scripts/powershell/install.ps1 | iex"
    goto :eof
)

if "%1"=="version" goto :version
if "%1"=="--version" goto :version
if "%1"=="-v" goto :version

if "%1"=="status" goto :status

if "%1"=="diagnose" goto :diagnose
if "%1"=="doctor" goto :diagnose

if "%1"=="repair" goto :repair
if "%1"=="fix" goto :repair

if "%1"=="help" goto :help
if "%1"=="--help" goto :help
if "%1"=="-h" goto :help
if "%1"=="" goto :help

echo Unknown command: %1
echo Run 'plaesy help' for usage information
exit /b 1

:upgrade
echo Upgrading Plaesy Spec-Kit...
powershell -ExecutionPolicy Bypass -Command "iwr -useb https://raw.githubusercontent.com/plaesy/spec-kit/main/scripts/powershell/install.ps1 | iex"
goto :eof

:uninstall
echo Uninstalling Plaesy Spec-Kit...
powershell -ExecutionPolicy Bypass -File "%PLAESY_HOME%\scripts\powershell\install.ps1" -uninstall
goto :eof

:version
if exist "%PLAESY_HOME%\VERSION" (
    set /p VERSION=<"%PLAESY_HOME%\VERSION"
    echo Plaesy Spec-Kit v%VERSION%
) else (
    echo Plaesy Spec-Kit v1.0.0
)
echo Constitutional Development Framework
goto :eof

:status
echo Plaesy Spec-Kit Status
echo.
if exist "%PLAESY_HOME%" (
    if exist "%PLAESY_HOME%\VERSION" (
        set /p VERSION=<"%PLAESY_HOME%\VERSION"
        echo Status: Installed (v%VERSION%)
    ) else (
        echo Status: Installed (v1.0.0)
    )
    echo Location: %PLAESY_HOME%
    echo CLI: %USERPROFILE%\.local\bin\plaesy.bat
    
    if exist "%PLAESY_HOME%\INSTALL_DATE" (
        set /p INSTALL_DATE=<"%PLAESY_HOME%\INSTALL_DATE"
        echo Installed: %INSTALL_DATE%
    )
    
    echo.
    echo Available Commands:
    echo   plaesy init          - Initialize new project
    echo   plaesy upgrade       - Upgrade to latest version
    echo   plaesy repair        - Fix missing scripts after upgrade
    echo   plaesy uninstall     - Remove Plaesy Spec-Kit
    echo   plaesy status        - Show installation status
    echo   plaesy diagnose      - Check installation and diagnose issues
) else (
    echo Status: Not installed
    echo Run 'plaesy install' to install
)
goto :eof

:diagnose
echo Plaesy Spec-Kit Diagnostic
echo.
echo === Environment Check ===
where git >nul 2>&1 && echo Git: Available || echo Git: Not found
where powershell >nul 2>&1 && echo PowerShell: Available || echo PowerShell: Not found
echo.

echo === Installation Check ===
if exist "%PLAESY_HOME%" (
    echo Framework: Installed at %PLAESY_HOME%
    
    if exist "%PLAESY_HOME%\scripts\powershell\plaesy-init.ps1" (
        echo Init Script: Available
    ) else (
        echo Init Script: Missing
    )
    
    if exist "%PLAESY_HOME%\templates\" (
        echo Templates: Available
    ) else (
        echo Templates: Missing
    )
    
    if exist "%PLAESY_HOME%\prompts\" (
        echo Prompts: Available
    ) else (
        echo Prompts: Missing
    )
    
    if exist "%PLAESY_HOME%\instructions\" (
        echo Instructions: Available
    ) else (
        echo Instructions: Missing
    )
    
    if exist "%PLAESY_HOME%\chatmodes\" (
        echo Chatmodes: Available
    ) else (
        echo Chatmodes: Missing
    )
) else (
    echo Framework: Not installed
)
echo.

echo === CLI Check ===
if exist "%USERPROFILE%\.local\bin\plaesy.bat" (
    echo CLI Command: Available
) else (
    echo CLI Command: Missing
)
echo.

echo === Quick Fix ===
echo If issues found, try: plaesy upgrade
goto :eof

:repair
echo Repairing Plaesy Spec-Kit installation...
powershell -ExecutionPolicy Bypass -Command "iwr -useb https://raw.githubusercontent.com/plaesy/spec-kit/main/scripts/powershell/install.ps1 | iex"
goto :eof

:help
echo Plaesy Spec-Kit - Constitutional Development Framework
echo.
echo Usage: plaesy ^<command^> [options]
echo.
echo Commands:
echo   init [directory]     Initialize a new project with constitutional framework
echo   install             Install or reinstall Plaesy Spec-Kit
echo   upgrade             Upgrade to the latest version
echo   repair              Fix missing scripts after upgrade
echo   uninstall           Remove Plaesy Spec-Kit completely
echo   status              Show installation status and information
echo   diagnose            Check installation status and diagnose issues
echo   version             Show version information
echo   help                Show this help message
echo.
echo Examples:
echo   plaesy init .                    # Initialize in current directory
echo   plaesy init my-project           # Initialize in new directory
echo   plaesy init . -ai copilot        # Initialize with GitHub Copilot
echo   plaesy init . -ai cursor         # Initialize with Cursor AI
echo   plaesy upgrade                   # Upgrade to latest version
echo   plaesy repair                    # Fix missing scripts
echo   plaesy diagnose                  # Check for issues
echo   plaesy uninstall                 # Remove completely
goto :eof
"@

    $batchPath = "$BinDir\plaesy.bat"
    $batchContent | Out-File -FilePath $batchPath -Encoding ASCII
    
    Write-Success "Plaesy CLI created successfully"
}

function New-InitScript {
    Write-Step "Creating initialization script..."
    
    # Validate InstallPath is not empty
    if ([string]::IsNullOrWhiteSpace($InstallPath)) {
        Write-Error-Custom "InstallPath is empty or not set"
    }
    
    # The script is already in the correct location within the framework
    # CLI batch file will call it directly from the framework location
    $sourceInitScript = Join-Path -Path $InstallPath -ChildPath "scripts\powershell\plaesy-init.ps1"
    
    if (Test-Path $sourceInitScript) {
        Write-Success "Initialization script available in framework"
    }
    else {
        Write-Error-Custom "plaesy-init.ps1 not found in framework at: $sourceInitScript"
    }
}

function Set-PathEnvironment {
    Write-Step "Setting up environment PATH..."
    
    # Get current PATH from user environment
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    
    # Check if our bin directory is already in PATH
    if ($currentPath -notlike "*$BinDir*") {
        # Add to user PATH
        $newPath = if ($currentPath) { "$currentPath;$BinDir" } else { $BinDir }
        [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
        
        # Update current session PATH immediately
        $env:PATH = "$env:PATH;$BinDir"
        
        Write-Success "PATH environment variable updated"
    }
    else {
        Write-Success "PATH already contains Plaesy bin directory"
    }
    
    # Refresh environment variables in current session
    Write-Step "Refreshing environment variables..."
    try {
        # Force refresh of environment variables for current session
        $env:PATH = [Environment]::GetEnvironmentVariable("PATH", "Machine") + ";" + [Environment]::GetEnvironmentVariable("PATH", "User")
        Write-Success "Environment variables refreshed successfully"
    }
    catch {
        Write-Warning "Could not refresh environment automatically"
    }
}

function Write-InstallationComplete {
    Write-Host ""
    Write-Host "Plaesy Spec-Kit installation completed successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Installation details:" -ForegroundColor Cyan
    Write-Host "Framework: $InstallPath"
    Write-Host "CLI: $BinDir\plaesy.bat"
    Write-Host "Version: $Version"
    Write-Host ""
    Write-Host "Quick Start:" -ForegroundColor Yellow
    Write-Host "1. Create a new project: plaesy init my-project"
    Write-Host "2. Or initialize current directory: plaesy init ."
    Write-Host ""
    Write-Host "Usage Examples:" -ForegroundColor Blue
    Write-Host "| plaesy init . -ai copilot     # Initialize with GitHub Copilot"
    Write-Host "| plaesy init . -ai cursor      # Initialize with Cursor AI"
    Write-Host "| plaesy init . -ai claude      # Initialize with Claude"
    Write-Host "| plaesy help                   # Show all commands"
    Write-Host ""
    Write-Host "Constitutional Development Framework" -ForegroundColor Magenta
    Write-Host "Discipline through automation. Quality through structure."
    Write-Host ""
    Write-Host "Learn more: https://github.com/plaesy/spec-kit "
}

# Upgrade functionality
function Start-Upgrade {
    Write-Host ""
    Write-Host "Upgrading Plaesy Spec-Kit to latest version..." -ForegroundColor Yellow
    Write-Host ""
    
    # Check if already installed
    if (!(Test-Path $InstallPath)) {
        Write-Error "Plaesy Spec-Kit is not installed. Run installation first."
    }
    
    # Get current version
    $currentVersion = "1.0.0"
    if (Test-Path "$InstallPath\VERSION") {
        $currentVersion = Get-Content "$InstallPath\VERSION" -Raw
        $currentVersion = $currentVersion.Trim()
    }
    
    Write-Step "Current version: v$currentVersion"
    Write-Step "Checking for updates..."
    
    # Create backup before upgrade
    $backupDir = "$env:USERPROFILE\.plaesy-backup-$(Get-Date -Format 'yyyyMMdd-HHmmss')"
    Write-Step "Creating backup at $backupDir..."
    try {
        Copy-Item -Path $InstallPath -Destination $backupDir -Recurse -Force
    }
    catch {
        Write-Warning "Backup creation failed, continuing with upgrade..."
    }
    
    # Preserve user configurations
    $tempConfigs = "$env:USERPROFILE\.plaesy-configs-temp"
    if (!(Test-Path $tempConfigs)) {
        New-Item -ItemType Directory -Path $tempConfigs -Force | Out-Null
    }
    
    # Save any user customizations
    if (Test-Path "$InstallPath\user-config.yaml") {
        Copy-Item -Path "$InstallPath\user-config.yaml" -Destination $tempConfigs -Force
    }
    
    # Download latest version
    Write-Step "Downloading latest framework..."
    Remove-Item -Path $InstallPath -Recurse -Force
    
    try {
        git clone --depth 1 $RepoUrl $InstallPath
        Remove-Item -Path "$InstallPath\.git" -Recurse -Force -ErrorAction SilentlyContinue
    }
    catch {
        Write-Error-Custom "Failed to download latest version: $($_.Exception.Message)"
    }
    
    # Filter platform-specific scripts
    # Remove-NonPlatformScripts
    
    # Update version tracking
    $Version | Out-File -FilePath "$InstallPath\VERSION" -Encoding UTF8
    (Get-Date) | Out-File -FilePath "$InstallPath\UPGRADE_DATE" -Encoding UTF8
    
    # Restore user configurations
    if (Test-Path "$tempConfigs\user-config.yaml") {
        Copy-Item -Path "$tempConfigs\user-config.yaml" -Destination $InstallPath -Force
    }
    Remove-Item -Path $tempConfigs -Recurse -Force -ErrorAction SilentlyContinue
    
    # Update CLI and recreate initialization script
    New-PleasyCLI
    New-InitScript
    
    Write-Success "Upgrade completed successfully!"
    Write-Host ""
    Write-Host "Upgrade Summary:" -ForegroundColor Green
    Write-Host "Previous version: v$currentVersion"
    Write-Host "Current version: v$Version"
    Write-Host "Backup location: $backupDir"
    Write-Host "Framework: $InstallPath"
    Write-Host ""
    Write-Host "Next Steps:" -ForegroundColor Yellow
    Write-Host "1. Test the upgrade: plaesy version"
    Write-Host "2. Initialize a test project: plaesy init test-project"
    Write-Host "3. If issues occur, restore from backup"
    Write-Host ""
}

# Uninstall functionality
function Start-Uninstall {
    Write-Host ""
    Write-Host "Uninstalling Plaesy Spec-Kit..." -ForegroundColor Red
    Write-Host ""
    
    # Check if installed
    if (!(Test-Path $InstallPath)) {
        Write-Warning "Plaesy Spec-Kit is not installed."
        return
    }
    
    # Get version info
    $version = "1.0.0"
    if (Test-Path "$InstallPath\VERSION") {
        $version = Get-Content "$InstallPath\VERSION" -Raw
        $version = $version.Trim()
    }
    
    Write-Step "Found Plaesy Spec-Kit v$version"
    Write-Step "Installation directory: $InstallPath"
    Write-Step "CLI location: $BinDir\plaesy.bat"
    
    # Confirmation prompt
    Write-Host ""
    Write-Host "This will completely remove Plaesy Spec-Kit from your system." -ForegroundColor Yellow
    Write-Host ""
    Write-Host "The following will be removed:"
    Write-Host "Framework directory: $InstallPath"
    Write-Host "CLI command: $BinDir\plaesy.bat"
    Write-Host "Environment PATH modifications"
    Write-Host ""
    
    $confirmation = Read-Host "Are you sure you want to continue? (yes/no)"
    
    if ($confirmation -notmatch '^(yes|YES|y|Y)$') {
        Write-Host "Uninstall cancelled." -ForegroundColor Red
        return
    }
    
    Write-Step "Proceeding with uninstall..."
    
    # Create final backup
    $backupDir = "$env:USERPROFILE\.plaesy-final-backup-$(Get-Date -Format 'yyyyMMdd-HHmmss')"
    Write-Step "Creating final backup at $backupDir..."
    try {
        Copy-Item -Path $InstallPath -Destination $backupDir -Recurse -Force
    }
    catch {
        Write-Warning "Backup creation failed, continuing with uninstall..."
    }
    
    # Remove framework directory
    Write-Step "Removing framework directory..."
    Remove-Item -Path $InstallPath -Recurse -Force -ErrorAction SilentlyContinue
    
    # Remove CLI command
    Write-Step "Removing CLI command..."
    Remove-Item -Path "$BinDir\plaesy.bat" -Force -ErrorAction SilentlyContinue
    
    # Clean up environment PATH
    Write-Step "Cleaning up environment PATH..."
    try {
        $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
        if ($currentPath -like "*$BinDir*") {
            $newPath = $currentPath -replace [regex]::Escape("$BinDir;"), ""
            $newPath = $newPath -replace [regex]::Escape(";$BinDir"), ""
            $newPath = $newPath -replace [regex]::Escape("$BinDir"), ""
            [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
        }
    }
    catch {
        Write-Warning "PATH cleanup failed, you may need to manually remove $BinDir from your PATH"
    }
    
    Write-Success "Plaesy Spec-Kit uninstalled successfully!"
    Write-Host ""
    Write-Host "Uninstall Summary:" -ForegroundColor Green
    Write-Host "Framework removed from: $InstallPath"
    Write-Host "CLI command removed"
    Write-Host "Environment PATH cleaned up"
    Write-Host "Final backup saved: $backupDir"
    Write-Host ""
    Write-Host "Note:" -ForegroundColor Cyan
    Write-Host "Restart your terminal to complete PATH cleanup"
    Write-Host "Your project files remain unchanged"
    Write-Host "Backup is available if you need to restore"
    Write-Host ""
    Write-Host "To reinstall:" -ForegroundColor Yellow
    Write-Host "iwr -useb https://raw.githubusercontent.com/plaesy/spec-kit/main/scripts/powershell/install.ps1 | iex"
    Write-Host ""
}

# Main installation function
function Start-Installation {
    Write-Banner
    Test-Configuration
    Test-Prerequisites
    Install-Framework
    # Remove-NonPlatformScripts
    
    # Save version and install date
    $Version | Out-File -FilePath "$InstallPath\VERSION" -Encoding UTF8
    (Get-Date) | Out-File -FilePath "$InstallPath\INSTALL_DATE" -Encoding UTF8
    
    New-PleasyCLI
    New-InitScript
    Set-PathEnvironment
    Write-InstallationComplete
}

# Execute based on parameters
try {
    # Check for upgrade or uninstall parameters
    if ($args -contains "-upgrade" -or $args -contains "--upgrade") {
        Start-Upgrade
    }
    elseif ($args -contains "-uninstall" -or $args -contains "--uninstall") {
        Start-Uninstall
    }
    else {
        Start-Installation
    }
}
catch {
    Write-Host ""
    Write-Host "Operation failed: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Please try again or report the issue at: https://github.com/plaesy/spec-kit/issues"
    exit 1
}


