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
    [string]$TestingValue,
    [hashtable]$EnvironmentOverrides
  )

  $pwshName = if ($PSVersionTable.PSEdition -eq 'Desktop') {
    'powershell.exe'
  }
  elseif ($env:OS -eq 'Windows_NT') {
    'pwsh.exe'
  }
  else {
    'pwsh'
  }
  $pwshPath = Join-Path $PSHOME $pwshName
  $originalTestingValue = $env:EVMGO_INSTALLER_TESTING
  $originalOverrideValues = @{}

  try {
    if ($null -eq $TestingValue) {
      Remove-Item Env:EVMGO_INSTALLER_TESTING -ErrorAction SilentlyContinue
    }
    else {
      $env:EVMGO_INSTALLER_TESTING = $TestingValue
    }

    if ($null -ne $EnvironmentOverrides) {
      foreach ($key in $EnvironmentOverrides.Keys) {
        $originalOverrideValues[$key] = [Environment]::GetEnvironmentVariable($key, 'Process')
        $value = $EnvironmentOverrides[$key]
        if ($null -eq $value) {
          Remove-Item "Env:$key" -ErrorAction SilentlyContinue
        }
        else {
          [Environment]::SetEnvironmentVariable($key, [string]$value, 'Process')
        }
      }
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

    foreach ($key in $originalOverrideValues.Keys) {
      $value = $originalOverrideValues[$key]
      if ($null -eq $value) {
        Remove-Item "Env:$key" -ErrorAction SilentlyContinue
      }
      else {
        [Environment]::SetEnvironmentVariable($key, $value, 'Process')
      }
    }
  }
}

$env:LOCALAPPDATA = Join-Path ([System.IO.Path]::GetTempPath()) 'LocalAppData'
$null = New-Item -ItemType Directory -Force -Path $env:LOCALAPPDATA
$env:EVMGO_INSTALLER_TESTING = '1'
. $installerPath

$linuxAssetName = Get-AssetName -Os 'linux' -Arch 'amd64' -Version 'v0.1.0'
Assert-Equal -Expected 'evmgo_v0.1.0_linux_amd64.tar.gz' -Actual $linuxAssetName -Message 'Get-AssetName returned an unexpected linux asset name'

$windowsAssetName = Get-AssetName -Os 'windows' -Arch 'amd64' -Version 'v0.1.0'
Assert-Equal -Expected 'evmgo_v0.1.0_windows_amd64.zip' -Actual $windowsAssetName -Message 'Get-AssetName returned an unexpected windows asset name'

$latestVersion = Get-LatestVersionFromJson -Json @'
[
  {"tag_name":"release-2026-04-01","draft":false,"prerelease":false},
  {"tag_name":"v0.2.0-rc1","draft":false,"prerelease":true},
  {"tag_name":"v0.1.0","draft":false,"prerelease":false}
]
'@
Assert-Equal -Expected 'v0.1.0' -Actual $latestVersion -Message 'Get-LatestVersionFromJson should return the first stable v* release'

$updatedPath = Add-PathEntryIfMissing -CurrentPath 'C:\Windows\System32' -Entry 'C:\Users\me\AppData\Local\Programs\evmgo\bin'
Assert-Equal -Expected 'C:\Windows\System32;C:\Users\me\AppData\Local\Programs\evmgo\bin' -Actual $updatedPath -Message 'Add-PathEntryIfMissing should append a missing entry once'

$idempotentPath = Add-PathEntryIfMissing -CurrentPath $updatedPath -Entry 'C:\Users\me\AppData\Local\Programs\evmgo\bin'
Assert-Equal -Expected 'C:\Windows\System32;C:\Users\me\AppData\Local\Programs\evmgo\bin' -Actual $idempotentPath -Message 'Add-PathEntryIfMissing should be idempotent for existing entries'

if (-not (Get-Variable -Name ProjectOwner -ErrorAction SilentlyContinue)) {
  throw 'ProjectOwner should be defined'
}

if (-not (Get-Variable -Name ProjectRepo -ErrorAction SilentlyContinue)) {
  throw 'ProjectRepo should be defined'
}

if (-not (Get-Variable -Name GithubApiRoot -ErrorAction SilentlyContinue)) {
  throw 'GithubApiRoot should be defined'
}

if (-not (Get-Variable -Name ReleaseDownloadRoot -ErrorAction SilentlyContinue)) {
  throw 'ReleaseDownloadRoot should be defined'
}

if (-not (Get-Variable -Name DefaultBinDir -ErrorAction SilentlyContinue)) {
  throw 'DefaultBinDir should be defined'
}

$testingEnabledResult = Invoke-InstallerSubprocess -TestingValue '1'
Assert-Equal -Expected '0' -Actual "$($testingEnabledResult.ExitCode)" -Message 'Installer testing guard should suppress Invoke-Main only when EVMGO_INSTALLER_TESTING is exactly 1'

$testingDisabledResult = Invoke-InstallerSubprocess -TestingValue '0'
if (-not (Test-IsWindowsPlatform)) {
  Assert-Equal -Expected '1' -Actual "$($testingDisabledResult.ExitCode)" -Message 'Installer should still invoke Invoke-Main when EVMGO_INSTALLER_TESTING is 0'
  Assert-True -Condition ($testingDisabledResult.Output -match 'unsupported operating system') -Message 'Installer failure output should report the unsupported operating system when testing is disabled on a non-Windows host'
}

$subprocessSyntheticRoot = Join-Path ([System.IO.Path]::GetTempPath()) ([System.IO.Path]::GetRandomFileName())
New-Item -ItemType Directory -Path $subprocessSyntheticRoot | Out-Null

try {
  $subprocessPayloadDir = Join-Path $subprocessSyntheticRoot 'payload'
  $subprocessArchiveSourceDir = Join-Path $subprocessPayloadDir 'evmgo_v0.1.0_windows_amd64'
  $subprocessApiRoot = Join-Path $subprocessSyntheticRoot 'api'
  $subprocessReleaseRoot = Join-Path $subprocessSyntheticRoot 'downloads'
  $subprocessVersionRoot = Join-Path $subprocessReleaseRoot 'v0.1.0'
  $subprocessArchiveName = 'evmgo_v0.1.0_windows_amd64.zip'
  $subprocessArchivePath = Join-Path $subprocessVersionRoot $subprocessArchiveName
  $subprocessChecksumsPath = Join-Path $subprocessVersionRoot 'checksums.txt'
  $subprocessDefaultBinDir = Join-Path $subprocessSyntheticRoot 'custom-bin'
  $subprocessExpectedInstallPath = Join-Path $subprocessDefaultBinDir 'evmgo.exe'
  $subprocessLocalAppData = Join-Path $subprocessSyntheticRoot 'LocalAppData'

  New-Item -ItemType Directory -Force -Path $subprocessArchiveSourceDir | Out-Null
  New-Item -ItemType Directory -Force -Path $subprocessVersionRoot | Out-Null
  New-Item -ItemType Directory -Force -Path $subprocessApiRoot | Out-Null
  'synthetic-subprocess-evmgo-binary' | Set-Content -NoNewline -Path (Join-Path $subprocessArchiveSourceDir 'evmgo.exe')

  [System.IO.Compression.ZipFile]::CreateFromDirectory($subprocessArchiveSourceDir, $subprocessArchivePath)
  $subprocessArchiveChecksum = (Get-FileHash -Algorithm SHA256 -Path $subprocessArchivePath).Hash.ToLowerInvariant()
  "$subprocessArchiveChecksum  $subprocessArchiveName" | Set-Content -NoNewline -Path $subprocessChecksumsPath
  @'
[
  {"tag_name":"release-2026-04-01","draft":false,"prerelease":false},
  {"tag_name":"v0.2.0-rc1","draft":false,"prerelease":true},
  {"tag_name":"v0.1.0","draft":false,"prerelease":false}
]
'@ | Set-Content -NoNewline -Path (Join-Path $subprocessApiRoot 'releases')

  $subprocessWindowsGuardResult = Invoke-InstallerSubprocess -TestingValue '0' -EnvironmentOverrides @{
    OS = 'Windows_NT'
    PROCESSOR_ARCHITECTURE = 'AMD64'
    LOCALAPPDATA = $subprocessLocalAppData
    EVMGO_GITHUB_API_ROOT = "file://$subprocessApiRoot"
    EVMGO_RELEASE_DOWNLOAD_ROOT = "file://$subprocessReleaseRoot"
    EVMGO_DEFAULT_BIN_DIR = $subprocessDefaultBinDir
  }

  Assert-Equal -Expected '0' -Actual "$($subprocessWindowsGuardResult.ExitCode)" -Message 'Installer should still invoke Invoke-Main when EVMGO_INSTALLER_TESTING is 0 on the Windows path'
  Assert-True -Condition (Test-Path -LiteralPath $subprocessExpectedInstallPath) -Message 'Installer subprocess should install into the overridden Windows bin directory when testing is disabled'
  Assert-Equal -Expected 'synthetic-subprocess-evmgo-binary' -Actual ([System.IO.File]::ReadAllText($subprocessExpectedInstallPath)) -Message 'Installer subprocess should preserve binary contents during the synthetic Windows install'
}
finally {
  Remove-Item -Recurse -Force $subprocessSyntheticRoot
}

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

$syntheticRoot = Join-Path ([System.IO.Path]::GetTempPath()) ([System.IO.Path]::GetRandomFileName())
New-Item -ItemType Directory -Path $syntheticRoot | Out-Null
Add-Type -AssemblyName System.IO.Compression.FileSystem -ErrorAction SilentlyContinue

try {
  $payloadDir = Join-Path $syntheticRoot 'payload'
  $archiveSourceDir = Join-Path $payloadDir 'evmgo_v0.1.0_windows_amd64'
  $releaseDir = Join-Path $syntheticRoot 'release'
  $archiveName = 'evmgo_v0.1.0_windows_amd64.zip'
  $script:SyntheticArchivePath = Join-Path $releaseDir $archiveName
  $script:SyntheticChecksumsPath = Join-Path $releaseDir 'checksums.txt'
  $fakeBinDir = Join-Path $syntheticRoot 'bin'
  $expectedInstallPath = Join-Path $fakeBinDir 'evmgo.exe'
  $script:FakeUserPath = 'C:\Windows\System32'
  $script:DownloadLog = [System.Collections.Generic.List[string]]::new()

  New-Item -ItemType Directory -Force -Path $archiveSourceDir | Out-Null
  New-Item -ItemType Directory -Force -Path $releaseDir | Out-Null
  'synthetic-evmgo-binary' | Set-Content -NoNewline -Path (Join-Path $archiveSourceDir 'evmgo.exe')

  [System.IO.Compression.ZipFile]::CreateFromDirectory($archiveSourceDir, $script:SyntheticArchivePath)
  $archiveChecksum = (Get-FileHash -Algorithm SHA256 -Path $script:SyntheticArchivePath).Hash.ToLowerInvariant()
  "$archiveChecksum  $archiveName" | Set-Content -NoNewline -Path $script:SyntheticChecksumsPath

  function Test-IsWindowsPlatform {
    return $true
  }

  function Get-InstallerArchitecture {
    return 'amd64'
  }

  function Get-LatestReleaseJson {
    return @'
[
  {"tag_name":"release-2026-04-01","draft":false,"prerelease":false},
  {"tag_name":"v0.2.0-rc1","draft":false,"prerelease":true},
  {"tag_name":"v0.1.0","draft":false,"prerelease":false}
]
'@
  }

  function Invoke-DownloadFile {
    param(
      [Parameter(Mandatory = $true)]
      [string]$Uri,
      [Parameter(Mandatory = $true)]
      [string]$OutFile
    )

    $script:DownloadLog.Add($Uri) | Out-Null
    $sourcePath = switch -Regex ($Uri) {
      'checksums\.txt$' { $script:SyntheticChecksumsPath; break }
      'windows_amd64\.zip$' { $script:SyntheticArchivePath; break }
      default { throw "unexpected download URI: $Uri" }
    }

    Copy-Item -LiteralPath $sourcePath -Destination $OutFile -Force
  }

  function Get-UserPathValue {
    return $script:FakeUserPath
  }

  function Set-UserPathValue {
    param(
      [AllowNull()]
      [string]$PathValue
    )

    $script:FakeUserPath = $PathValue
  }

  $originalVersion = $env:VERSION
  $originalPath = $env:PATH
  $originalDefaultBinDir = $DefaultBinDir

  try {
    Remove-Item Env:VERSION -ErrorAction SilentlyContinue
    $env:PATH = 'C:\Windows\System32'
    $script:DefaultBinDir = $fakeBinDir

    Invoke-Main

    Assert-True -Condition (Test-Path -LiteralPath $expectedInstallPath) -Message 'Synthetic installer flow should copy evmgo.exe into the target bin directory'
    Assert-Equal -Expected 'synthetic-evmgo-binary' -Actual ([System.IO.File]::ReadAllText($expectedInstallPath)) -Message 'Synthetic installer flow should preserve the binary contents'
    Assert-Equal -Expected '2' -Actual "$($script:DownloadLog.Count)" -Message 'Synthetic installer flow should download release metadata artifacts once per install run'
    Assert-True -Condition ($script:DownloadLog[0] -match '/v0\.1\.0/checksums\.txt$') -Message 'Synthetic installer flow should resolve the release version before downloading checksums'
    Assert-True -Condition ($script:DownloadLog[1] -match '/v0\.1\.0/evmgo_v0\.1\.0_windows_amd64\.zip$') -Message 'Synthetic installer flow should download the matching Windows archive for the resolved version'
    Assert-Equal -Expected "C:\Windows\System32;$fakeBinDir" -Actual $script:FakeUserPath -Message 'Synthetic installer flow should append the bin directory to the user PATH once'
    Assert-True -Condition ($env:PATH -match [Regex]::Escape($fakeBinDir)) -Message 'Synthetic installer flow should add the bin directory to the current process PATH'

    Invoke-Main

    Assert-Equal -Expected '4' -Actual "$($script:DownloadLog.Count)" -Message 'Synthetic installer flow should repeat downloads on a second install run'
    Assert-Equal -Expected "C:\Windows\System32;$fakeBinDir" -Actual $script:FakeUserPath -Message 'Synthetic installer flow should keep PATH mutation idempotent across repeated installs'
  }
  finally {
    if ($null -eq $originalVersion) {
      Remove-Item Env:VERSION -ErrorAction SilentlyContinue
    }
    else {
      $env:VERSION = $originalVersion
    }

    $env:PATH = $originalPath
    $script:DefaultBinDir = $originalDefaultBinDir
  }
}
finally {
  Remove-Item -Recurse -Force $syntheticRoot
}
