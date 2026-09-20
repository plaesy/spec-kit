# Generate an image asset via a configured provider API and save it to disk.
# Used by prompts/generate/images.md (/generate:images).
#
# Usage:
#   ./generate-image.ps1 -Prompt "<text>" [-Provider openai|gemini] [-Size 1024x1024] -Out <path>

param(
    [Parameter(Mandatory = $true)][string]$Prompt,
    [string]$Provider = $(if ($env:PLAESY_IMAGE_PROVIDER) { $env:PLAESY_IMAGE_PROVIDER } else { "openai" }),
    [string]$Size = "1024x1024",
    [Parameter(Mandatory = $true)][string]$Out
)

$ErrorActionPreference = "Stop"

$outDir = Split-Path -Parent $Out
if ($outDir -and -not (Test-Path $outDir)) {
    New-Item -ItemType Directory -Force -Path $outDir | Out-Null
}

switch ($Provider) {
    "openai" {
        if (-not $env:OPENAI_API_KEY) {
            Write-Error "OPENAI_API_KEY is not set. Set it first, e.g.:`n  `$env:OPENAI_API_KEY = 'sk-...'"
            exit 2
        }
        $body = @{ model = "gpt-image-1"; prompt = $Prompt; size = $Size; n = 1 } | ConvertTo-Json
        try {
            $response = Invoke-RestMethod -Uri "https://api.openai.com/v1/images/generations" `
                -Method Post `
                -Headers @{ Authorization = "Bearer $($env:OPENAI_API_KEY)" } `
                -ContentType "application/json" `
                -Body $body
        } catch {
            Write-Error "OpenAI image API request failed: $($_.Exception.Message)"
            exit 3
        }
        $b64 = $response.data[0].b64_json
        [IO.File]::WriteAllBytes($Out, [Convert]::FromBase64String($b64))
    }
    "gemini" {
        if (-not $env:GEMINI_API_KEY) {
            Write-Error "GEMINI_API_KEY is not set. Set it first, e.g.:`n  `$env:GEMINI_API_KEY = '...'"
            exit 2
        }
        $body = @{ instances = @(@{ prompt = $Prompt }); parameters = @{ sampleCount = 1 } } | ConvertTo-Json -Depth 5
        $uri = "https://generativelanguage.googleapis.com/v1beta/models/imagen-3.0-generate-002:predict?key=$($env:GEMINI_API_KEY)"
        try {
            $response = Invoke-RestMethod -Uri $uri -Method Post -ContentType "application/json" -Body $body
        } catch {
            Write-Error "Gemini image API request failed: $($_.Exception.Message)"
            exit 3
        }
        $b64 = $response.predictions[0].bytesBase64Encoded
        [IO.File]::WriteAllBytes($Out, [Convert]::FromBase64String($b64))
    }
    default {
        Write-Error "Unknown provider '$Provider' (supported: openai, gemini)"
        exit 4
    }
}

if (-not (Test-Path $Out) -or (Get-Item $Out).Length -eq 0) {
    Write-Error "$Out was not written or is empty"
    exit 5
}

Write-Output $Out
