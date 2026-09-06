# Plaesy Helper for Windows Batch Files
# Handles platform detection and command execution

param(
    [string]$Command,
    [string]$TargetDir = "",
    [string]$AiChoice = ""
)

# Import platform detector
$detectorScript = "$env:USERPROFILE\.plaesy\scripts\powershell\platform-detector.ps1"

function Get-PlatformCommandWithDetection {
    param(
        [string]$Command
    )

    # Import platform detector
    try {
        . $detectorScript | Out-Null
        $detectedPlatform = Get-DetectedAiPlatform

        if ($detectedPlatform -ne "none") {
            Write-Host "[SUCCESS] Detected AI platform: $detectedPlatform" -ForegroundColor Green
            $commandPattern = Get-PlatformCommand -Platform $detectedPlatform -Command $Command
            Write-Host "[INFO] Use: $commandPattern in your AI assistant" -ForegroundColor Cyan
            return @{
                Platform = $detectedPlatform
                Command = $commandPattern
                Args = @("-ai", $detectedPlatform)
            }
        } else {
            Write-Host "[WARNING] No AI platform detected. Starting universal..." -ForegroundColor Yellow
            return @{
                Platform = ""
                Command = $Command
                Args = @()
            }
        }
    }
    catch {
        Write-Host "[WARNING] Platform detector not available, using standard" -ForegroundColor Yellow
        return @{
            Platform = ""
            Command = $Command
            Args = @()
        }
    }
}

# Execute command with platform detection
switch ($Command) {
    "init" {
        $platformInfo = Get-PlatformCommandWithDetection -Command "init"

        $initArgs = @()
        if ($AiChoice) { $initArgs += "--ai", $AiChoice }
        elseif ($platformInfo.Args) { $initArgs += $platformInfo.Args }
        if ($TargetDir) { $initArgs += $TargetDir }

        & "$env:USERPROFILE\.plaesy\scripts\powershell\plaesy-init.ps1" @initArgs
    }

    "clean" {
        $platformInfo = Get-PlatformCommandWithDetection -Command "clean"

        $cleanArgs = @()
        if ($AiChoice) { $cleanArgs += "--ai", $AiChoice }
        elseif ($platformInfo.Args) { $cleanArgs += $platformInfo.Args }
        if ($TargetDir) { $cleanArgs += $TargetDir }

        & "$env:USERPROFILE\.plaesy\scripts\powershell\plaesy-clean.ps1" @cleanArgs
    }

    default {
        Write-Error "Unknown command: $Command"
        exit 1
    }
}