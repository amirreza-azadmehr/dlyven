# Authenticode-sign an exe with a .pfx so SmartScreen sees a known publisher.
# Always timestamped, so the signature stays valid after the certificate expires.
#
# Env:
#   WINDOWS_CERTIFICATE           base64-encoded .pfx   (legacy name: WIN_CERTIFICATE_PFX)
#   WINDOWS_CERTIFICATE_PASSWORD                         (legacy name: WIN_CERTIFICATE_PASSWORD)
param(
  [string]$Exe = 'build\bin\Dlyven.exe'
)
$ErrorActionPreference = 'Stop'
Set-Location (Join-Path $PSScriptRoot '..')

$certB64 = if ($env:WINDOWS_CERTIFICATE) { $env:WINDOWS_CERTIFICATE } else { $env:WIN_CERTIFICATE_PFX }
$certPassword = if ($env:WINDOWS_CERTIFICATE_PASSWORD) { $env:WINDOWS_CERTIFICATE_PASSWORD } else { $env:WIN_CERTIFICATE_PASSWORD }

if (-not (Test-Path $Exe)) { throw "no exe at $Exe" }
if (-not $certB64) {
  Write-Host "No WINDOWS_CERTIFICATE - $Exe stays unsigned; SmartScreen will warn on downloads."
  exit 0
}

$tempDir = if ($env:RUNNER_TEMP) { $env:RUNNER_TEMP } else { $env:TEMP }
$pfx = Join-Path $tempDir 'dlyven-code-sign.pfx'
try {
  [IO.File]::WriteAllBytes($pfx, [Convert]::FromBase64String($certB64))

  $signtool = Get-ChildItem "${env:ProgramFiles(x86)}\Windows Kits\10\bin\*\x64\signtool.exe" -ErrorAction SilentlyContinue |
    Sort-Object FullName -Descending | Select-Object -First 1
  if (-not $signtool) { throw 'signtool.exe not found (Windows 10 SDK)' }

  Write-Host "-> signtool $($signtool.FullName) $Exe"
  & $signtool.FullName sign /f $pfx /p $certPassword /tr http://timestamp.digicert.com /td sha256 /fd sha256 /d Dlyven $Exe
  if ($LASTEXITCODE -ne 0) { throw 'signtool sign failed' }
  & $signtool.FullName verify /pa $Exe
  if ($LASTEXITCODE -ne 0) { throw 'signtool verify failed' }
} finally {
  Remove-Item $pfx -Force -ErrorAction SilentlyContinue
}
Write-Host "OK: signed $Exe"
