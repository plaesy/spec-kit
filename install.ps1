# Plaesy Constitution Kit installer for Windows (Windows PowerShell 5.1 or PowerShell 7+).
#
#   Invoke-WebRequest -Uri https://raw.githubusercontent.com/plaesy/spec-kit/main/install.ps1 -OutFile install.ps1
#   .\install.ps1
#
# Downloads the release binary, verifies it against SHA256SUMS from the same
# release, runs `plaesy install`, then adds the install directory to the user
# PATH if it is not already there. Open a new terminal afterwards.
#
# Environment:
#   PLAESY_VERSION   release to install, e.g. v0.0.1 (default: latest)
#   PLAESY_BASE_URL  directory holding the release assets; overrides the GitHub
#                    URL (for testing or a mirror)

$ErrorActionPreference = 'Stop'
$RepoUrl = 'https://github.com/plaesy/spec-kit'

function Get-BaseUrl {
    if ($env:PLAESY_BASE_URL) {
        return $env:PLAESY_BASE_URL.TrimEnd('/')
    }
    $version = if ($env:PLAESY_VERSION) { $env:PLAESY_VERSION } else { 'latest' }
    if ($version -eq 'latest') {
        return "$RepoUrl/releases/latest/download"
    }
    if ($version -notmatch '^v?\d+\.\d+\.\d+$') {
        throw "PLAESY_VERSION must look like v1.2.3 or 'latest', got: $version"
    }
    if ($version -notlike 'v*') {
        $version = "v$version"
    }
    return "$RepoUrl/releases/download/$version"
}

function Invoke-Download([string]$Uri, [string]$OutFile) {
    $previous = $ProgressPreference
    $ProgressPreference = 'SilentlyContinue'
    try {
        Invoke-WebRequest -Uri $Uri -OutFile $OutFile -UseBasicParsing
    }
    finally {
        $ProgressPreference = $previous
    }
}

# Fails closed: no checksum entry for the asset means no install.
function Test-Checksum([string]$File, [string]$SumsFile, [string]$Asset) {
    $expected = $null
    foreach ($line in Get-Content -LiteralPath $SumsFile) {
        $parts = $line.Trim() -split '\s+', 2
        if ($parts.Count -eq 2 -and $parts[1].TrimStart('*') -eq $Asset) {
            $expected = $parts[0].ToLower()
        }
    }
    if (-not $expected) {
        throw "no checksum for $Asset in SHA256SUMS; refusing to install"
    }
    $actual = (Get-FileHash -LiteralPath $File -Algorithm SHA256).Hash.ToLower()
    if ($actual -ne $expected) {
        throw "checksum mismatch for ${Asset}: expected $expected, got $actual"
    }
    Write-Host 'Checksum verified (SHA-256)'
}

# Adds dir to the user PATH (not machine) if it is missing. Matches entries
# case-insensitively and ignores a trailing backslash, as Windows does.
function Add-UserPath([string]$Dir) {
    $current = [Environment]::GetEnvironmentVariable('Path', 'User')
    $entries = @($current -split ';' | Where-Object { $_ -ne '' })
    if ($entries | Where-Object { $_.TrimEnd('\') -ieq $Dir.TrimEnd('\') }) {
        Write-Host "$Dir is already on your user PATH"
        return
    }
    $updated = (@($entries) + $Dir) -join ';'
    [Environment]::SetEnvironmentVariable('Path', $updated, 'User')
    # Also update this session so the command works right away in it.
    $env:Path = "$env:Path;$Dir"
    Write-Host "Added $Dir to your user PATH. Open a new terminal to use plaesy."
}

function Main {
    # Windows PowerShell 5.1 defaults to older TLS versions that GitHub rejects.
    [Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12

    # Only an amd64 binary is published. Windows on ARM runs it under emulation.
    $asset = 'plaesy-windows-amd64.exe'
    $base = Get-BaseUrl

    $tmp = Join-Path ([IO.Path]::GetTempPath()) ('plaesy-' + [IO.Path]::GetRandomFileName())
    New-Item -ItemType Directory -Path $tmp | Out-Null
    try {
        $bin = Join-Path $tmp 'plaesy.exe'
        $sums = Join-Path $tmp 'SHA256SUMS'

        Write-Host "Downloading $asset from $base"
        try {
            Invoke-Download "$base/$asset" $bin
        }
        catch {
            throw "download failed: $base/$asset ($($_.Exception.Message))"
        }
        try {
            Invoke-Download "$base/SHA256SUMS" $sums
        }
        catch {
            throw 'could not download SHA256SUMS; refusing to install an unverified binary'
        }

        Test-Checksum $bin $sums $asset

        & $bin install
        if ($LASTEXITCODE -ne 0) {
            throw "plaesy install exited with code $LASTEXITCODE"
        }

        $installDir = Join-Path $env:LOCALAPPDATA 'Plaesy\bin'
        Add-UserPath $installDir
    }
    finally {
        Remove-Item -LiteralPath $tmp -Recurse -Force -ErrorAction SilentlyContinue
    }
}

try {
    Main
}
catch {
    [Console]::Error.WriteLine("error: $($_.Exception.Message)")
    exit 1
}
