# Build Dlyven on Windows (primary platform).
#
#   powershell -ExecutionPolicy Bypass -File scripts\build.ps1             # portable Dlyven.exe
#   powershell -ExecutionPolicy Bypass -File scripts\build.ps1 -Installer  # + NSIS installer (needs makensis on PATH)
#   powershell -ExecutionPolicy Bypass -File scripts\build.ps1 -Arch arm64
param(
  [switch]$Installer,
  [ValidateSet('amd64', 'arm64')][string]$Arch = 'amd64',
  [string]$Version = ''
)
$ErrorActionPreference = 'Stop'
Set-Location (Join-Path $PSScriptRoot '..')

if (-not (Get-Command wails -ErrorAction SilentlyContinue)) {
  throw 'wails CLI not found. Install it with: go install github.com/wailsapp/wails/v2/cmd/wails@v2.10.2'
}
if (-not $Version) { $Version = (Get-Content wails.json -Raw | ConvertFrom-Json).info.productVersion }
$ldflags = "-s -w -X github.com/amirreza-azadmehr/dlyven/internal/appinfo.Version=$Version"

Write-Host '-> Running Go tests'
go test ./internal/...
if ($LASTEXITCODE -ne 0) { throw 'tests failed' }

# -webview2 embed bundles the WebView2 bootstrapper so the app also runs on machines without it.
$wailsArgs = @('build', '-clean', '-trimpath', '-platform', "windows/$Arch", '-webview2', 'embed', '-ldflags', $ldflags)
if ($Installer) { $wailsArgs += '-nsis' }
& wails @wailsArgs
if ($LASTEXITCODE -ne 0) { throw 'wails build failed' }
& (Join-Path $PSScriptRoot 'sign-windows.ps1') -Exe 'build\bin\Dlyven.exe'
if ($Installer) {
  Get-ChildItem 'build\bin\*.exe' | ForEach-Object {
    if ($_.Name -ne 'Dlyven.exe') {
      & (Join-Path $PSScriptRoot 'sign-windows.ps1') -Exe $_.FullName
    }
  }
}
Write-Host "OK: build\bin\Dlyven.exe$(if ($Installer) { ' (+ NSIS installer)' })"
