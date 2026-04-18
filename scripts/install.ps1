#!/usr/bin/env pwsh
$ErrorActionPreference = 'Stop'

$ProjectOwner = 'itzfelixv'
$ProjectRepo = 'evmgo'
$DefaultBinDir = Join-Path $env:LOCALAPPDATA 'Programs\evmgo\bin'

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

    $parts = $line -split '\s+'
    if ($parts.Count -ge 2 -and $parts[1] -eq $AssetName) {
      return $parts[0]
    }
  }

  return $null
}

function Invoke-Main {
  throw 'installer implementation not complete yet'
}

if ($env:EVMGO_INSTALLER_TESTING -ne '1') {
  Invoke-Main
}
