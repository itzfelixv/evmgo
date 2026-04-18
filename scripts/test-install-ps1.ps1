#!/usr/bin/env pwsh
$ErrorActionPreference = 'Stop'

$scriptDir = Split-Path -Parent $PSCommandPath
$installerPath = Join-Path $scriptDir 'install.ps1'

function Assert-Equal {
  param(
    [Parameter(Mandatory = $true)]
    [string]$Expected,
    [Parameter(Mandatory = $true)]
    [string]$Actual,
    [Parameter(Mandatory = $true)]
    [string]$Message
  )

  if ($Expected -ne $Actual) {
    throw "$Message`nexpected: $Expected`nactual:   $Actual"
  }
}

function Assert-True {
  param(
    [Parameter(Mandatory = $true)]
    [bool]$Condition,
    [Parameter(Mandatory = $true)]
    [string]$Message
  )

  if (-not $Condition) {
    throw $Message
  }
}

function Invoke-InstallerSubprocess {
  param(
    [string]$TestingValue
  )

  $pwshName = if ($IsWindows) { 'pwsh.exe' } else { 'pwsh' }
  $pwshPath = Join-Path $PSHOME $pwshName
  $originalTestingValue = $env:EVMGO_INSTALLER_TESTING

  try {
    if ($null -eq $TestingValue) {
      Remove-Item Env:EVMGO_INSTALLER_TESTING -ErrorAction SilentlyContinue
    }
    else {
      $env:EVMGO_INSTALLER_TESTING = $TestingValue
    }

    $output = & $pwshPath -NoProfile -File $installerPath 2>&1
    return @{
      ExitCode = $LASTEXITCODE
      Output = ($output | Out-String)
    }
  }
  finally {
    if ($null -eq $originalTestingValue) {
      Remove-Item Env:EVMGO_INSTALLER_TESTING -ErrorAction SilentlyContinue
    }
    else {
      $env:EVMGO_INSTALLER_TESTING = $originalTestingValue
    }
  }
}

$env:LOCALAPPDATA = Join-Path ([System.IO.Path]::GetTempPath()) 'LocalAppData'
$null = New-Item -ItemType Directory -Force -Path $env:LOCALAPPDATA
$env:EVMGO_INSTALLER_TESTING = '1'
. $installerPath

if (Get-Variable -Name GitHubApiRoot -ErrorAction SilentlyContinue) {
  throw 'GitHubApiRoot should not be defined in the Task 3 installer scaffold'
}

if (Get-Variable -Name ReleaseDownloadRoot -ErrorAction SilentlyContinue) {
  throw 'ReleaseDownloadRoot should not be defined in the Task 3 installer scaffold'
}

$linuxAssetName = Get-AssetName -Os 'linux' -Arch 'amd64' -Version 'v0.1.0'
Assert-Equal -Expected 'evmgo_v0.1.0_linux_amd64.tar.gz' -Actual $linuxAssetName -Message 'Get-AssetName returned an unexpected linux asset name'

$windowsAssetName = Get-AssetName -Os 'windows' -Arch 'amd64' -Version 'v0.1.0'
Assert-Equal -Expected 'evmgo_v0.1.0_windows_amd64.zip' -Actual $windowsAssetName -Message 'Get-AssetName returned an unexpected windows asset name'

$testingEnabledResult = Invoke-InstallerSubprocess -TestingValue '1'
Assert-Equal -Expected '0' -Actual "$($testingEnabledResult.ExitCode)" -Message 'Installer testing guard should suppress Invoke-Main only when EVMGO_INSTALLER_TESTING is exactly 1'

$testingDisabledResult = Invoke-InstallerSubprocess -TestingValue '0'
Assert-Equal -Expected '1' -Actual "$($testingDisabledResult.ExitCode)" -Message 'Installer should still invoke Invoke-Main when EVMGO_INSTALLER_TESTING is 0'
Assert-True -Condition ($testingDisabledResult.Output -match 'installer implementation not complete yet') -Message 'Installer failure output should include the scaffold placeholder message when testing is disabled'

$tmpDir = Join-Path ([System.IO.Path]::GetTempPath()) ([System.IO.Path]::GetRandomFileName())
New-Item -ItemType Directory -Path $tmpDir | Out-Null

try {
  $checksumsPath = Join-Path $tmpDir 'checksums.txt'
  @'
0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef  evmgo_v0.1.0_windows_amd64.zip
fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210  evmgo_v0.1.0_linux_amd64.tar.gz
'@ | Set-Content -NoNewline -Path $checksumsPath

  $checksum = Get-ChecksumForAsset -ChecksumsPath $checksumsPath -AssetName 'evmgo_v0.1.0_windows_amd64.zip'
  Assert-Equal -Expected '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef' -Actual $checksum -Message 'Get-ChecksumForAsset returned an unexpected checksum'
}
finally {
  Remove-Item -Recurse -Force $tmpDir
}
