# Plaesy Platform Detector for PowerShell
# Auto-detects AI platforms using platform.json configuration

function Get-DetectedAiPlatform {
    <#
    .SYNOPSIS
        Detects the current AI platform using platform.json configuration

    .DESCRIPTION
        Automatically detects which AI platform is being used by reading
        detection patterns from platform.json and checking for matching files.

    .OUTPUTS
        String representing the detected platform (claude_code, cursor_ai, etc.)
        Returns $null if no platform is detected
    #>

    # Get script directory and locate platform.json
    $ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
    $PlatformConfig = Join-Path $ScriptDir "..\configs\platform.json"

    # Check if platform.json exists
    if (-not (Test-Path $PlatformConfig)) {
        Write-Warning "Platform configuration file not found: $PlatformConfig"
        return $null
    }

    try {
        # Read and parse platform.json
        $configContent = Get-Content $PlatformConfig -Raw | ConvertFrom-Json

        # Get platforms in order from JSON
        $platforms = $configContent.platforms.PSObject.Properties.Name

        # Check each platform for detection patterns
        foreach ($platform in $platforms) {
            $platformInfo = $configContent.platforms.$platform

            if ($platformInfo.detection -and $platformInfo.detection.Count -gt 0) {
                # Check each detection pattern
                foreach ($pattern in $platformInfo.detection) {
                    # Handle wildcards and directory patterns
                    if ($pattern.Contains("*")) {
                        # Convert PowerShell-style wildcards to actual file checks
                        $matches = Get-ChildItem -Path $pattern -ErrorAction SilentlyContinue
                        if ($matches -and $matches.Count -gt 0) {
                            return $platform
                        }
                    } elseif (Test-Path $pattern) {
                        return $platform
                    }
                }
            }
        }

        # No platform detected
        return $null
    }
    catch {
        Write-Warning "Failed to parse platform configuration: $_"
        return $null
    }
}

function Get-PlatformCommand {
    <#
    .SYNOPSIS
        Gets the command pattern for a specific platform

    .DESCRIPTION
        Returns the appropriate command prefix/suffix for the detected platform

    .PARAMETER Platform
        The platform name (claude_code, cursor_ai, etc.)

    .PARAMETER Command
        The command to execute (init, clean, analyze, etc.)

    .OUTPUTS
        String representing the full command pattern for the platform
    #>

    param(
        [string]$Platform,
        [string]$Command
    )

    switch ($Platform.ToLower()) {
        "claude_code" { return "/$Command" }
        "cursor_ai" { return "--$Command" }
        "github_copilot" { return "/$Command" }
        "windsurf_ai" { return "/$Command" }
        "cline" { return "/$Command" }
        "deepseek" { return "/$Command" }
        "kilo_code" { return "/$Command" }
        "qoder" { return "/$Command" }
        "trae_ai" { return "/$Command" }
        default { return $Command }
    }
}

function Show-DetectionResult {
    <#
    .SYNOPSIS
        Shows the platform detection result with helpful information

    .DESCRIPTION
        Displays the detected platform and provides usage instructions
    #>

    $platform = Get-DetectedAiPlatform

    if ($platform -ne "none") {
        Write-Host "✅ Detected AI platform: $platform" -ForegroundColor Green

        # Show platform-specific command examples
        switch ($platform) {
            "claude_code" {
                Write-Host "💡 Use: /init, /analyze, /continue in your AI assistant" -ForegroundColor Cyan
            }
            "cursor_ai" {
                Write-Host "💡 Use: --init, --analyze, --continue in your AI assistant" -ForegroundColor Cyan
            }
            default {
                Write-Host "💡 Use: /$command in your AI assistant" -ForegroundColor Cyan
            }
        }
    } else {
        Write-Host "🤖 No AI platform detected" -ForegroundColor Yellow
        Write-Host "💡 Use: plaesy init to start interactive platform selection" -ForegroundColor Cyan
    }
}

# Export functions if module is imported
Export-ModuleMember -Function Get-DetectedAiPlatform, Get-PlatformCommand, Show-DetectionResult