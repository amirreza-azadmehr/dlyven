# Authenticode-sign Dlyven.exe so SmartScreen stops treating it as an unknown publisher.
#
# Env:
#   WIN_CERTIFICATE_PFX        base64-encoded .pfx
#   WIN_CERTIFICATE_PASSWORD
param(
  [string]$Exe = 'build\bin\Dlyven.exe'
)
$ErrorActionPreference = 'Stop'
Set-Location (Join-Path $PSScriptRoot '..')

if (-not (Test-Path $Exe)) { throw "no exe at $Exe" }
if (-not $env:WIN_CERTIFICATE_PFX) {
  Write-Host 'No WIN_CERTIFICATE_PFX — SmartScreen will still warn on downloads.'
  exit 0
}

$pfx = Join-Path $env:RUNNER_TEMP 'dlyven-code-sign.pfx'
if (-not $pfx) { $pfx = Join-Path $env:TEMP 'dlyven-code-sign.pfx' }
[IO.File]::WriteAllBytes($pfx, [Convert]::FromBase64String($env:WIN_CERTIFICATE_PFX))

$signtool = Get-ChildItem "${env:ProgramFiles(x86)}\Windows Kits\10\bin\*\x64\signtool.exe" -ErrorAction SilentlyContinue |
  Sort-Object FullName -Descending | Select-Object -First 1
if (-not $signtool) { throw 'signtool.exe not found (Windows 10 SDK)' }

Write-Host "-> signtool $($signtool.FullName)"
& $signtool.FullName sign /f $pfx /p $env:WIN_CERTIFICATE_PASSWORD /tr http://timestamp.digicert.com /td sha256 /fd sha256 /d Dlyven $Exe
if ($LASTEXITCODE -ne 0) { throw 'signtool failed' }
& $signtool.FullName verify /pa $Exe
Remove-Item $pfx -Force -ErrorAction SilentlyContinue
Write-Host "OK: signed $Exe"
