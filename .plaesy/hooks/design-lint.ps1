# plaesy design-lint: advisory design check after a file edit. Never blocks; always exits 0.
$ErrorActionPreference = 'SilentlyContinue'
$root = git rev-parse --show-toplevel 2>$null
if (-not $root) { $root = (Get-Location).Path }
$payload = [Console]::In.ReadToEnd()
$named = [regex]::Matches($payload, '"file_path"\s*:\s*"([^"]*)"') | ForEach-Object { $_.Groups[1].Value }
$patched = [regex]::Matches($payload, '\*\*\* (?:Add|Update) File: ([^\\"]+)') | ForEach-Object { $_.Groups[1].Value }
$files = @($named) + @($patched) | ForEach-Object { $_ -replace '\\\\', '/' } | Where-Object { $_ }
if (-not $files) { $files = git -C $root status --porcelain | ForEach-Object { $_.Substring(3) } }
$bin = Join-Path $root 'plaesy.exe'
if (-not (Test-Path $bin)) { $bin = Join-Path $root 'plaesy' }
if (-not (Test-Path $bin)) {
  $cmd = Get-Command plaesy -ErrorAction SilentlyContinue
  if (-not $cmd) { exit 0 }
  $bin = $cmd.Source
}
foreach ($f in ($files | Sort-Object -Unique)) {
  if ($f -notmatch '\.(html|htm|css|scss|jsx|tsx|vue|svelte)$') { continue }
  $p = if ([System.IO.Path]::IsPathRooted($f)) { $f } else { Join-Path $root $f }
  if (-not (Test-Path $p -PathType Leaf)) { continue }
  & $bin validate design $p 2>&1 | Select-String -Pattern '^(FAIL|WARN) '
}
exit 0
