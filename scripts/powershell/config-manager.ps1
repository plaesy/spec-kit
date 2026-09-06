# Plaesy Config Manager (PowerShell)
# Centralized configuration management using platform.json

param(
    [Parameter(Mandatory=$false)]
    [string]$Command = "",

    [Parameter(Mandatory=$false)]
    [string]$Platform = "",

    [Parameter(Mandatory=$false)]
    [string]$Key = ""
)

$ErrorActionPreference = "Stop"

# Get script directory
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$PlatformConfig = Join-Path $ScriptDir "..\configs\platform.json"

# Logging functions
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

function Write-Error-Custom {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
    exit 1
}

# Read platform configuration
function Get-PlatformConfig {
    if (-not (Test-Path $PlatformConfig)) {
        Write-Error-Custom "Platform configuration file not found: $PlatformConfig"
    }

    try {
        $content = Get-Content $PlatformConfig -Raw
        return $content | ConvertFrom-Json
    }
    catch {
        Write-Error-Custom "Failed to parse platform configuration: $_"
    }
}

# List all available platforms
function Show-AvailablePlatforms {
    $config = Get-PlatformConfig
    $platforms = $config.platforms.PSObject.Properties.Name

    foreach ($platform in $platforms) {
        Write-Output $platform
    }
}

# Detect current AI platform
function Find-AIPlatform {
    $config = Get-PlatformConfig
    $platforms = $config.platforms.PSObject.Properties.Name

    foreach ($platform in $platforms) {
        $platformInfo = $config.platforms.$platform
        if ($platformInfo.detection -and $platformInfo.detection.Count -gt 0) {
            foreach ($pattern in $platformInfo.detection) {
                if (Test-Path $pattern) {
                    Write-Output $platform
                    return
                }
            }
        }
    }

    # No platform detected - return nothing to trigger interactive selection
    return
}

# Get platform configuration value
function Get-PlatformConfigValue {
    param(
        [string]$PlatformName,
        [string]$ConfigKey
    )

    $config = Get-PlatformConfig

    # Check if it's the plaesy section or a platform section
    $sectionInfo = $null
    if ($PlatformName -eq "plaesy") {
        $sectionInfo = $config.plaesy
    } elseif ($config.platforms.PSObject.Properties.Name -contains $PlatformName) {
        $sectionInfo = $config.platforms.$PlatformName
    } else {
        Write-Error-Custom "Section '$PlatformName' not found in configuration"
        return $null
    }

    switch ($ConfigKey) {
        "name" { return $sectionInfo.name }
        "provider" { return $sectionInfo.provider }
        "category" { return $sectionInfo.category }
        "base_directory" { return $sectionInfo.base_directory }
        "core_directories" {
            # Convert array to space-separated string
            if ($sectionInfo.core_directories -is [array]) {
                return $sectionInfo.core_directories -join " "
            } else {
                return $sectionInfo.core_directories
            }
        }
        "project_directories" {
            # Convert array to space-separated string
            if ($sectionInfo.project_directories -is [array]) {
                return $sectionInfo.project_directories -join " "
            } else {
                return $sectionInfo.project_directories
            }
        }
        "mapping.core" { return $sectionInfo.mapping.core }
        "mapping.instructions" { return $sectionInfo.mapping.instructions }
        "mapping.prompts" { return $sectionInfo.mapping.prompts }
        "mapping.chatmodes" { return $sectionInfo.mapping.chatmodes }
        default {
            if ($sectionInfo.PSObject.Properties.Name -contains $ConfigKey) {
                return $sectionInfo.$ConfigKey
            } else {
                Write-Warning "Configuration key '$ConfigKey' not found for section '$PlatformName'"
                return $null
            }
        }
    }
}

# Get mapping value with support for new structure (value/excludes)
function Get-MappingValue {
    param(
        [string]$Section,
        [string]$MappingType
    )

    $config = Get-PlatformConfig

    # Check if section exists in platforms or plaesy (platform-specific takes priority)
    $sectionInfo = $null
    if ($config.platforms -and $config.platforms.PSObject.Properties.Name -contains $Section) {
        $platformInfo = $config.platforms.$Section
        if ($platformInfo.mapping -and $platformInfo.mapping.PSObject.Properties.Name -contains $MappingType) {
            $sectionInfo = $platformInfo.mapping.$MappingType
        }
    } elseif ($config.plaesy -and $config.plaesy.mapping -and $config.plaesy.mapping.PSObject.Properties.Name -contains $MappingType) {
        $sectionInfo = $config.plaesy.mapping.$MappingType
    }

    if ($sectionInfo) {
        # Check if it's the new structure (object with value/excludes)
        if ($sectionInfo -is [PSCustomObject] -and $sectionInfo.PSObject.Properties.Name -contains "value") {
            return $sectionInfo.value
        } else {
            # Old structure (direct string value)
            return $sectionInfo
        }
    }

    return $null
}

# Get exclude patterns for mapping
function Get-MappingExcludes {
    param(
        [string]$Section,
        [string]$MappingType
    )

    $config = Get-PlatformConfig

    # Check if section exists in platforms or plaesy (platform-specific takes priority)
    $sectionInfo = $null
    if ($config.platforms -and $config.platforms.PSObject.Properties.Name -contains $Section) {
        $platformInfo = $config.platforms.$Section
        if ($platformInfo.mapping -and $platformInfo.mapping.PSObject.Properties.Name -contains $MappingType) {
            $sectionInfo = $platformInfo.mapping.$MappingType
        }
    } elseif ($config.plaesy -and $config.plaesy.mapping -and $config.plaesy.mapping.PSObject.Properties.Name -contains $MappingType) {
        $sectionInfo = $config.plaesy.mapping.$MappingType
    }

    if ($sectionInfo -and $sectionInfo -is [PSCustomObject] -and $sectionInfo.PSObject.Properties.Name -contains "excludes") {
        return $sectionInfo.excludes
    }

    return @()
}

# Get Plaesy structure configuration
function Get-PlaesyStructure {
    param([string]$Component)

    if (-not (Test-Path $PlatformConfig)) {
        Write-Error-Custom "Platform configuration file not found: $PlatformConfig"
        return $null
    }

    try {
        $config = Get-Content $PlatformConfig -Raw | ConvertFrom-Json

        if ($config.plaesy) {
            $plaesy = $config.plaesy
            switch ($Component) {
                "base_directory" {
                    return $plaesy.base_directory
                }
                "core_directories" {
                    if ($plaesy.core_directories -is [array]) {
                        return $plaesy.core_directories -join " "
                    } else {
                        return $plaesy.core_directories
                    }
                }
                "project_directories" {
                    if ($plaesy.project_directories -is [array]) {
                        return $plaesy.project_directories -join " "
                    } else {
                        return $plaesy.project_directories
                    }
                }
                "plaesy_root" {
                    return $plaesy.plaesy_root
                }
                default {
                    if ($plaesy.PSObject.Properties.Name -contains $Component) {
                        return $plaesy.$Component
                    } else {
                        Write-Warning "Plaesy structure component '$Component' not found"
                        return $null
                    }
                }
            }
        } else {
            # Fallback defaults
            switch ($Component) {
                "base_directory" { return ".plaesy" }
                "core_directories" { return "memory" }
                "memory_subdirectories" { return "knowledge" }
                "project_directories" { return "docs specs" }
                "plaesy_root" { return "`${SCRIPT_DIR}/../.." }
                default { return $null }
            }
        }
    }
    catch {
        Write-Warning "Failed to parse Plaesy structure: $_"
        # Return fallback defaults
        switch ($Component) {
            "base_directory" { return ".plaesy" }
            "core_directories" { return "memory" }
            "memory_subdirectories" { return "knowledge" }
            "project_directories" { return "docs specs" }
            "plaesy_root" { return "`${SCRIPT_DIR}/../.." }
            default { return $null }
        }
    }
}

# Show help information
function Show-Help {
    Write-Host ""
    Write-Host "Plaesy Config Manager - Centralized Configuration Management" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "USAGE:" -ForegroundColor Yellow
    Write-Host "    .\config-manager.ps1 <command> [options]"
    Write-Host ""
    Write-Host "COMMANDS:" -ForegroundColor Yellow
    Write-Host "    detect-platform              Detect current AI platform"
    Write-Host "    list-platforms               List all available platforms"
    Write-Host "    get-platform-config <platform> <key>  Get specific platform configuration"
    Write-Host "    get-mapping-value <section> <mapping_type>  Get mapping value"
    Write-Host "    get-mapping-excludes <section> <mapping_type>  Get mapping exclude patterns"
    Write-Host "    get-clean-files [platform]   Get files to clean for platform"
    Write-Host "    get-clean-dirs [platform]    Get directories to clean for platform"
    Write-Host "    get-plaesy-structure <component> Get Plaesy structure configuration"
    Write-Host "    show-platform-info [platform] Show detailed platform information"
    Write-Host "    validate                     Validate platform configuration"
    Write-Host "    help                         Show this help message"
    Write-Host ""
    Write-Host "EXAMPLES:" -ForegroundColor Yellow
    Write-Host "    .\config-manager.ps1 detect-platform"
    Write-Host "    .\config-manager.ps1 list-platforms"
    Write-Host "    .\config-manager.ps1 get-platform-config claude_code name"
    Write-Host "    .\config-manager.ps1 get-plaesy-structure base_directory"
    Write-Host ""
    Write-Host "DEPENDENCIES:" -ForegroundColor Yellow
    Write-Host "    None required - uses PowerShell built-in JSON parsing"
    Write-Host ""
    Write-Host "CONFIGURATION:" -ForegroundColor Yellow
    Write-Host "    Platform configuration: configs/platform.json (JSON format)"
}

# Get cleanup files for platform
function Get-CleanFiles {
    param([string]$Platform = "")

    if ([string]::IsNullOrEmpty($Platform)) {
        $Platform = Find-AIPlatform
    }

    # Fallback hardcoded files (cleanup configuration not in current JSON)
    return @("CLAUDE.md", ".cursorrules", ".github/copilot-instructions.md")
}

# Get source file processing rules for plaesy-init
function Get-SourceRules {
    param([string]$FileType)

    # Fallback hardcoded rules (source_files configuration not in current JSON)
    switch ($FileType) {
        "core" {
            return @("instructions/plaesy.instructions.md")
        }
        "instructions" {
            return @("instructions/*.instructions.md", "plaesy.instructions.md")
        }
        "prompts" {
            return @("prompts/*.prompt.md")
        }
        "chatmodes" {
            return @("chatmodes/*.chatmode.md")
        }
        default {
            return @()
        }
    }
}

# Get cleanup directories for platform
function Get-CleanDirs {
    param([string]$Platform = "")

    if ([string]::IsNullOrEmpty($Platform)) {
        $Platform = Find-AIPlatform
    }

    # Fallback hardcoded directories (compute from JSON mapping structure)
    $dirs = @()

    # Get core mapping and extract directory
    $coreTarget = Get-PlatformConfigValue -PlatformName $Platform -ConfigKey "mapping.core"
    if ($coreTarget) {
        $coreDir = Split-Path -Parent $coreTarget
        if ($coreDir -ne ".") {
            $dirs += $coreDir
        }
    }

    # Get prompts mapping and extract directory
    $promptsTarget = Get-PlatformConfigValue -PlatformName $Platform -ConfigKey "mapping.prompts"
    if ($promptsTarget -and $promptsTarget -ne "null") {
        $promptDir = Split-Path -Parent $promptsTarget
        if ($promptDir -ne "." -and $dirs -notcontains $promptDir) {
            $dirs += $promptDir
        }
    }

    # Get chatmodes mapping and extract directory
    $chatmodesTarget = Get-PlatformConfigValue -PlatformName $Platform -ConfigKey "mapping.chatmodes"
    if ($chatmodesTarget -and $chatmodesTarget -ne "null") {
        $chatmodeDir = Split-Path -Parent $chatmodesTarget
        if ($chatmodeDir -ne "." -and $dirs -notcontains $chatmodeDir) {
            $dirs += $chatmodeDir
        }
    }

    # Default directories if none found
    if ($dirs.Count -eq 0) {
        $platforms = Show-AvailablePlatforms
        foreach ($plat in $platforms) {
            $coreTarget = Get-PlatformConfigValue -PlatformName $plat -ConfigKey "mapping.core"
            if ($coreTarget) {
                $coreDir = Split-Path -Parent $coreTarget
                if ($coreDir -ne "." -and $dirs -notcontains $coreDir) {
                    $dirs += $coreDir
                }
            }
        }
    }

    if ($dirs.Count -eq 0) {
        return $null
    } else {
        return $dirs -join " "
    }
}

# Validate configuration
function Test-Configuration {
    if (-not (Test-Path $PlatformConfig)) {
        Write-Error-Custom "Platform configuration not found: $PlatformConfig"
        return $false
    }

    # Validate JSON syntax
    try {
        $content = Get-Content $PlatformConfig -Raw | ConvertFrom-Json
        Write-Success "Platform configuration is valid"
        return $true
    }
    catch {
        Write-Error-Custom "Invalid JSON syntax in platform configuration: $($_.Exception.Message)"
        return $false
    }
}

# Show platform information
function Show-PlatformInfo {
    param([string]$Platform = "")

    if ([string]::IsNullOrEmpty($Platform)) {
        $Platform = Find-AIPlatform
    }

    Write-Host "Platform Information for: $Platform"
    Write-Host "================================"

    $name = Get-PlatformConfigValue -PlatformName $Platform -ConfigKey "name"
    $provider = Get-PlatformConfigValue -PlatformName $Platform -ConfigKey "provider"
    $category = Get-PlatformConfigValue -PlatformName $Platform -ConfigKey "category"

    $displayName = if ($name) { $name } else { 'Unknown' }
    $displayProvider = if ($provider) { $provider } else { 'Unknown' }
    $displayCategory = if ($category) { $category } else { 'Unknown' }

    Write-Host "Name: $displayName"
    Write-Host "Provider: $displayProvider"
    Write-Host "Category: $displayCategory"
}

# Main execution logic
switch ($Command.ToLower()) {
    "detect-platform" {
        Find-AIPlatform
    }

    "list-platforms" {
        Show-AvailablePlatforms
    }

    "get-detection-patterns" {
        if ([string]::IsNullOrEmpty($Platform)) {
            Write-Error-Custom "Platform name is required for get-detection-patterns command"
        }
        $config = Get-PlatformConfig
        $platformInfo = $config.platforms.$Platform
        if ($platformInfo -and $platformInfo.detection) {
            $platformInfo.detection | ForEach-Object { Write-Output $_ }
        }
    }

    "get-platform-config" {
        if ([string]::IsNullOrEmpty($Platform)) {
            Write-Error-Custom "Platform name is required for get-platform-config command"
        }
        if ([string]::IsNullOrEmpty($Key)) {
            Write-Error-Custom "Configuration key is required for get-platform-config command"
        }

        $value = Get-PlatformConfigValue -PlatformName $Platform -ConfigKey $Key
        if ($value) {
            Write-Output $value
        }
    }

    "get-mapping-value" {
        if ([string]::IsNullOrEmpty($Platform)) {
            Write-Error-Custom "Section name is required for get-mapping-value command"
        }
        if ([string]::IsNullOrEmpty($Key)) {
            Write-Error-Custom "Mapping type is required for get-mapping-value command"
        }

        $value = Get-MappingValue -Section $Platform -MappingType $Key
        if ($value) {
            Write-Output $value
        }
    }

    "get-mapping-excludes" {
        if ([string]::IsNullOrEmpty($Platform)) {
            Write-Error-Custom "Section name is required for get-mapping-excludes command"
        }
        if ([string]::IsNullOrEmpty($Key)) {
            Write-Error-Custom "Mapping type is required for get-mapping-excludes command"
        }

        $excludes = Get-MappingExcludes -Section $Platform -MappingType $Key
        $excludes | ForEach-Object { Write-Output $_ }
    }

    "get-clean-files" {
        $files = Get-CleanFiles -Platform $Platform
        $files | ForEach-Object { Write-Output $_ }
    }

    "get-clean-dirs" {
        $dirs = Get-CleanDirs -Platform $Platform
        if ($dirs) {
            Write-Output $dirs
        }
    }

    "get-plaesy-structure" {
        if ([string]::IsNullOrEmpty($Key)) {
            Write-Error-Custom "Component name is required for get-plaesy-structure command"
        }

        $value = Get-PlaesyStructure -Component $Key
        if ($value) {
            Write-Output $value
        }
    }

    "show-platform-info" {
        Show-PlatformInfo -Platform $Platform
    }

    "validate" {
        Test-Configuration
    }

    "help" {
        Show-Help
    }

    "" {
        Show-Help
    }

    default {
        Write-Error-Custom "Unknown command: $Command. Use 'help' for usage information."
    }
}