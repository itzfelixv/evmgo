#!/usr/bin/env pwsh
$ErrorActionPreference = 'Stop'

$ProjectOwner = 'itzfelixv'
$ProjectRepo = 'evmgo'
$GithubApiRoot = if ([string]::IsNullOrWhiteSpace($env:EVMGO_GITHUB_API_ROOT)) {
  "https://api.github.com/repos/$ProjectOwner/$ProjectRepo"
}
else {
  $env:EVMGO_GITHUB_API_ROOT
}
$ReleaseDownloadRoot = if ([string]::IsNullOrWhiteSpace($env:EVMGO_RELEASE_DOWNLOAD_ROOT)) {
  "https://github.com/$ProjectOwner/$ProjectRepo/releases/download"
}
else {
  $env:EVMGO_RELEASE_DOWNLOAD_ROOT
}
$localAppDataRoot = if ([string]::IsNullOrWhiteSpace($env:LOCALAPPDATA)) {
  [Environment]::GetFolderPath('LocalApplicationData')
}
else {
  $env:LOCALAPPDATA
}
$DefaultBinDir = if ([string]::IsNullOrWhiteSpace($env:EVMGO_DEFAULT_BIN_DIR)) {
  Join-Path $localAppDataRoot 'Programs\evmgo\bin'
}
else {
  $env:EVMGO_DEFAULT_BIN_DIR
}

function Get-FileUriLocalPath {
  param(
    [Parameter(Mandatory = $true)]
    [string]$Uri
  )

  if (-not $Uri.StartsWith('file://', [System.StringComparison]::OrdinalIgnoreCase)) {
    return $null
  }

  return ([System.Uri]$Uri).LocalPath
}

function Get-AssetName {
  param(
    [Parameter(Mandatory = $true)]
    [string]$Os,
    [Parameter(Mandatory = $true)]
    [string]$Arch,
    [Parameter(Mandatory = $true)]
    [string]$Version
  )

  if ($Os -eq 'windows') {
    return "$ProjectRepo`_$Version`_$Os`_$Arch.zip"
  }

  return "$ProjectRepo`_$Version`_$Os`_$Arch.tar.gz"
}

function Get-ChecksumForAsset {
  param(
    [Parameter(Mandatory = $true)]
    [string]$ChecksumsPath,
    [Parameter(Mandatory = $true)]
    [string]$AssetName
  )

  foreach ($line in Get-Content -Path $ChecksumsPath) {
    if ([string]::IsNullOrWhiteSpace($line)) {
      continue
    }

    if ($line -match '^([0-9A-Fa-f]+)\s+\*?(.+)$' -and $Matches[2] -eq $AssetName) {
      return $Matches[1]
    }
  }

  return $null
}

function Get-LatestVersionFromJson {
  param(
    [Parameter(Mandatory = $true)]
    [string]$Json
  )

  foreach ($release in @(ConvertFrom-Json -InputObject $Json)) {
    if ($null -eq $release) {
      continue
    }

    $tagName = [string]$release.tag_name
    if ([string]::IsNullOrWhiteSpace($tagName)) {
      continue
    }

    if (-not [bool]$release.draft -and -not [bool]$release.prerelease -and $tagName.StartsWith('v', [System.StringComparison]::Ordinal)) {
      return $tagName
    }
  }

  throw 'unable to find a stable v* release in release metadata'
}

function Add-PathEntryIfMissing {
  param(
    [AllowNull()]
    [string]$CurrentPath,
    [Parameter(Mandatory = $true)]
    [string]$Entry
  )

  $trimmedEntry = $Entry.Trim()
  if ([string]::IsNullOrWhiteSpace($trimmedEntry)) {
    throw 'path entry must not be empty'
  }

  if ([string]::IsNullOrWhiteSpace($CurrentPath)) {
    return $trimmedEntry
  }

  foreach ($existingEntry in ($CurrentPath -split ';')) {
    if ($existingEntry.Trim().Equals($trimmedEntry, [System.StringComparison]::OrdinalIgnoreCase)) {
      return $CurrentPath
    }
  }

  return "$CurrentPath;$trimmedEntry"
}

function Get-InvokeWebRequestParameters {
  param(
    [Parameter(Mandatory = $true)]
    [string]$Uri
  )

  if ([enum]::GetNames([System.Net.SecurityProtocolType]) -contains 'Tls12') {
    $currentProtocols = [System.Net.ServicePointManager]::SecurityProtocol
    $tls12Protocol = [System.Net.SecurityProtocolType]::Tls12
    if (($currentProtocols -band $tls12Protocol) -eq 0) {
      [System.Net.ServicePointManager]::SecurityProtocol = $currentProtocols -bor $tls12Protocol
    }
  }

  $parameters = @{
    Uri = $Uri
    Headers = @{
      'User-Agent' = "$ProjectRepo-installer"
    }
  }

  $invokeWebRequestCommand = Get-Command -Name Invoke-WebRequest -CommandType Cmdlet
  if ($invokeWebRequestCommand.Parameters.ContainsKey('UseBasicParsing')) {
    $parameters.UseBasicParsing = $true
  }

  return $parameters
}

function Get-LatestReleaseJson {
  $uri = "$GithubApiRoot/releases"
  $localPath = Get-FileUriLocalPath -Uri $uri
  if ($null -ne $localPath) {
    return [System.IO.File]::ReadAllText($localPath)
  }

  $parameters = Get-InvokeWebRequestParameters -Uri $uri
  $response = Invoke-WebRequest @parameters
  if ($response -is [string]) {
    return $response
  }

  if ($null -ne $response.Content) {
    return [string]$response.Content
  }

  return ($response | Out-String)
}

function Invoke-DownloadFile {
  param(
    [Parameter(Mandatory = $true)]
    [string]$Uri,
    [Parameter(Mandatory = $true)]
    [string]$OutFile
  )

  $localPath = Get-FileUriLocalPath -Uri $Uri
  if ($null -ne $localPath) {
    Copy-Item -LiteralPath $localPath -Destination $OutFile -Force
    return
  }

  $parameters = Get-InvokeWebRequestParameters -Uri $Uri
  $parameters.OutFile = $OutFile
  Invoke-WebRequest @parameters | Out-Null
}

function Test-IsWindowsPlatform {
  if ($env:OS -eq 'Windows_NT') {
    return $true
  }

  try {
    return [System.Runtime.InteropServices.RuntimeInformation]::IsOSPlatform([System.Runtime.InteropServices.OSPlatform]::Windows)
  }
  catch {
    return [Environment]::OSVersion.Platform -eq [System.PlatformID]::Win32NT
  }
}

function Get-InstallerArchitecture {
  $architecture = $null

  try {
    $architecture = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLowerInvariant()
  }
  catch {
    $architecture = $null
  }

  if ([string]::IsNullOrWhiteSpace($architecture)) {
    $rawArchitecture = $env:PROCESSOR_ARCHITEW6432
    if ([string]::IsNullOrWhiteSpace($rawArchitecture)) {
      $rawArchitecture = $env:PROCESSOR_ARCHITECTURE
    }

    if (-not [string]::IsNullOrWhiteSpace($rawArchitecture)) {
      $architecture = $rawArchitecture.ToLowerInvariant()
    }
  }

  switch ($architecture) {
    'amd64' { return 'amd64' }
    'x64' { return 'amd64' }
    default { throw "unsupported architecture: $architecture" }
  }
}

function Expand-InstallerArchive {
  param(
    [Parameter(Mandatory = $true)]
    [string]$ArchivePath,
    [Parameter(Mandatory = $true)]
    [string]$DestinationPath
  )

  Expand-Archive -LiteralPath $ArchivePath -DestinationPath $DestinationPath -Force
}

function Get-UserPathValue {
  return [Environment]::GetEnvironmentVariable('Path', 'User')
}

function Set-UserPathValue {
  param(
    [AllowNull()]
    [string]$PathValue
  )

  [Environment]::SetEnvironmentVariable('Path', $PathValue, 'User')
}

function Resolve-Version {
  if (-not [string]::IsNullOrWhiteSpace($env:VERSION)) {
    return $env:VERSION
  }

  return Get-LatestVersionFromJson -Json (Get-LatestReleaseJson)
}

function Invoke-Main {
  if (-not (Test-IsWindowsPlatform)) {
    throw "unsupported operating system: $([Environment]::OSVersion.VersionString)"
  }

  $arch = Get-InstallerArchitecture
  $version = Resolve-Version
  $assetName = Get-AssetName -Os 'windows' -Arch $arch -Version $version
  $checksumsUrl = "$ReleaseDownloadRoot/$version/checksums.txt"
  $archiveUrl = "$ReleaseDownloadRoot/$version/$assetName"

  $tempRoot = Join-Path ([System.IO.Path]::GetTempPath()) ([System.IO.Path]::GetRandomFileName())
  $archivePath = Join-Path $tempRoot $assetName
  $checksumsPath = Join-Path $tempRoot 'checksums.txt'
  $extractDir = Join-Path $tempRoot 'extracted'
  $targetPath = Join-Path $DefaultBinDir 'evmgo.exe'

  New-Item -ItemType Directory -Force -Path $tempRoot | Out-Null

  try {
    New-Item -ItemType Directory -Force -Path $extractDir | Out-Null
    New-Item -ItemType Directory -Force -Path $DefaultBinDir | Out-Null

    Invoke-DownloadFile -Uri $checksumsUrl -OutFile $checksumsPath
    Invoke-DownloadFile -Uri $archiveUrl -OutFile $archivePath

    $expectedChecksum = Get-ChecksumForAsset -ChecksumsPath $checksumsPath -AssetName $assetName
    if ([string]::IsNullOrWhiteSpace($expectedChecksum)) {
      throw "checksum not found for asset $assetName"
    }

    $actualChecksum = (Get-FileHash -Algorithm SHA256 -Path $archivePath).Hash.ToLowerInvariant()
    if ($expectedChecksum.ToLowerInvariant() -ne $actualChecksum) {
      throw "checksum mismatch for $assetName`nexpected: $expectedChecksum`nactual:   $actualChecksum"
    }

    Expand-InstallerArchive -ArchivePath $archivePath -DestinationPath $extractDir

    $binaryPath = Get-ChildItem -Path $extractDir -Recurse -Filter 'evmgo.exe' -File | Select-Object -First 1 -ExpandProperty FullName
    if ([string]::IsNullOrWhiteSpace($binaryPath)) {
      throw 'evmgo.exe not found in extracted archive'
    }

    Copy-Item -LiteralPath $binaryPath -Destination $targetPath -Force

    $currentUserPath = Get-UserPathValue
    $updatedUserPath = Add-PathEntryIfMissing -CurrentPath $currentUserPath -Entry $DefaultBinDir
    $pathUpdated = $updatedUserPath -ne $currentUserPath
    if ($pathUpdated) {
      Set-UserPathValue -PathValue $updatedUserPath
    }

    $env:PATH = Add-PathEntryIfMissing -CurrentPath $env:PATH -Entry $DefaultBinDir

    Write-Host "Installed $ProjectRepo $version to $targetPath"
    if ($pathUpdated) {
      Write-Host "Added $DefaultBinDir to your user PATH."
      Write-Host 'Open a new PowerShell window, or run:'
      Write-Host "  `$env:PATH = [Environment]::GetEnvironmentVariable('Path', 'User') + ';' + [Environment]::GetEnvironmentVariable('Path', 'Machine')"
    }
    else {
      Write-Host "$DefaultBinDir is already on your user PATH."
      Write-Host 'Open a new PowerShell window if `evmgo` is not yet available.'
    }
  }
  finally {
    Remove-Item -Recurse -Force -Path $tempRoot -ErrorAction SilentlyContinue
  }
}

if ($env:EVMGO_INSTALLER_TESTING -ne '1') {
  Invoke-Main
}
