# Rebuild the NSIS installer around the current build\bin\Dlyven.exe.
#
# `wails build -nsis` packs the exe into the installer straight away, before it
# can be signed. Run this after signing the exe so the installer ships the signed
# copy, then sign the installer itself. Needs a prior `wails build -nsis`, which
# writes build\windows\installer\wails_tools.nsh.
param(
  [string]$Exe = 'build\bin\Dlyven.exe',
  [ValidateSet('amd64', 'arm64')][string]$Arch = 'amd64'
)
$ErrorActionPreference = 'Stop'
Set-Location (Join-Path $PSScriptRoot '..')

$installerDir = 'build\windows\installer'
if (-not (Test-Path (Join-Path $installerDir 'wails_tools.nsh'))) {
  throw "$installerDir\wails_tools.nsh missing - run 'wails build -nsis' first"
}
if (-not (Get-Command makensis -ErrorAction SilentlyContinue)) { throw 'makensis not found on PATH' }

$binary = (Resolve-Path $Exe).Path
$define = if ($Arch -eq 'arm64') { "-DARG_WAILS_ARM64_BINARY=$binary" } else { "-DARG_WAILS_AMD64_BINARY=$binary" }
Push-Location $installerDir
try {
  & makensis $define project.nsi
  if ($LASTEXITCODE -ne 0) { throw 'makensis failed' }
} finally {
  Pop-Location
}
Write-Host "OK: build\bin\Dlyven-$Arch-installer.exe (packs $Exe)"
