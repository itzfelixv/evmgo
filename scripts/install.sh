#!/usr/bin/env sh
set -eu

PROJECT_OWNER="itzfelixv"
PROJECT_REPO="evmgo"
DEFAULT_UNIX_BIN_DIR="${HOME:-/nonexistent}/.local/bin"

asset_name() {
  os=$1
  arch=$2
  version=$3

  case "$os" in
    windows)
      printf '%s\n' "${PROJECT_REPO}_${version}_${os}_${arch}.zip"
      ;;
    *)
      printf '%s\n' "${PROJECT_REPO}_${version}_${os}_${arch}.tar.gz"
      ;;
  esac
}

checksum_for_asset() {
  checksums_file=$1
  asset=$2

  awk -v asset="$asset" '$2 == asset { print $1; exit }' "$checksums_file"
}

main() {
  printf '%s\n' 'installer implementation not complete yet' >&2
  exit 1
}

if [ "${EVMGO_INSTALLER_TESTING:-0}" != "1" ]; then
  main "$@"
fi
