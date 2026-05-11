#Requires -Version 7.0
$ErrorActionPreference = 'Stop'

$RootDir = $PSScriptRoot

# Bootstrap .env on first run
$EnvFile    = Join-Path $RootDir 'backend\.env'
$EnvExample = Join-Path $RootDir '.env.example'
if (-not (Test-Path $EnvFile)) {
    Write-Host "backend\.env not found — copying from .env.example"
    Copy-Item $EnvExample $EnvFile
}

# Load env vars into the current process (PORT, CLAUDE_BIN, ANTHROPIC_API_KEY, etc.)
foreach ($line in Get-Content $EnvFile) {
    if ($line -match '^\s*#' -or $line -notmatch '=') { continue }
    $key, $value = $line -split '=', 2
    [System.Environment]::SetEnvironmentVariable($key.Trim(), $value.Trim(), 'Process')
}

Write-Host 'Installing dependencies...'
Push-Location (Join-Path $RootDir 'frontend')
npm install
Pop-Location

Push-Location (Join-Path $RootDir 'backend')
npm install
Pop-Location

Write-Host 'Building...'
Push-Location (Join-Path $RootDir 'frontend')
npm run build
Pop-Location

Push-Location (Join-Path $RootDir 'backend')
npm run build
Pop-Location

$Port = if ($env:PORT) { $env:PORT } else { '9998' }
Write-Host "Starting Claude Code Dashboard → http://localhost:$Port"
node (Join-Path $RootDir 'backend\dist\server.js')
