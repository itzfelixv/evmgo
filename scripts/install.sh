#!/usr/bin/env sh
set -eu

PROJECT_OWNER="${PROJECT_OWNER:-itzfelixv}"
PROJECT_REPO="${PROJECT_REPO:-evmgo}"
DEFAULT_UNIX_BIN_DIR="${DEFAULT_UNIX_BIN_DIR:-${HOME:-/nonexistent}/.local/bin}"
GITHUB_API_ROOT="${GITHUB_API_ROOT:-https://api.github.com/repos/${PROJECT_OWNER}/${PROJECT_REPO}}"
RELEASE_DOWNLOAD_ROOT="${RELEASE_DOWNLOAD_ROOT:-https://github.com/${PROJECT_OWNER}/${PROJECT_REPO}/releases/download}"

asset_name() {
  os=$1
  arch=$2
  version=$3
  asset_version=${version#v}

  case "$os" in
    windows)
      printf '%s\n' "${PROJECT_REPO}_${asset_version}_${os}_${arch}.zip"
      ;;
    *)
      printf '%s\n' "${PROJECT_REPO}_${asset_version}_${os}_${arch}.tar.gz"
      ;;
  esac
}

checksum_for_asset() {
  checksums_file=$1
  asset=$2

  awk -v asset="$asset" '$2 == asset { print $1; exit }' "$checksums_file"
}

latest_version_from_json() {
  json=$1
  releases_file=$(mktemp)
  version=''

  printf '%s' "$json" |
    tr '\n' ' ' |
    sed 's/^[[:space:]]*\[//; s/\][[:space:]]*$//' |
    sed 's/}[[:space:]]*,[[:space:]]*{/}\n{/g' > "$releases_file"

  while IFS= read -r release || [ -n "$release" ]; do
    tag=$(
      printf '%s\n' "$release" |
        sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' |
        head -n 1
    )
    draft=$(
      printf '%s\n' "$release" |
        sed -n 's/.*"draft"[[:space:]]*:[[:space:]]*\(true\|false\).*/\1/p' |
        head -n 1
    )
    prerelease=$(
      printf '%s\n' "$release" |
        sed -n 's/.*"prerelease"[[:space:]]*:[[:space:]]*\(true\|false\).*/\1/p' |
        head -n 1
    )

    case "$tag:$draft:$prerelease" in
      v*:false:false)
        version=$tag
        break
        ;;
    esac
  done < "$releases_file"

  rm -f "$releases_file"

  if [ -z "$version" ]; then
    printf '%s\n' 'unable to find a stable v* release in release metadata' >&2
    return 1
  fi

  printf '%s\n' "$version"
}

download_to_file() {
  url=$1
  destination=$2

  if command -v curl >/dev/null 2>&1; then
    curl -fsSL --retry 3 --output "$destination" "$url"
    return 0
  fi

  if command -v wget >/dev/null 2>&1; then
    wget -qO "$destination" "$url"
    return 0
  fi

  printf '%s\n' 'curl or wget is required to download evmgo' >&2
  return 1
}

sha256_file() {
  file=$1

  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$file" | awk '{ print $1 }'
    return 0
  fi

  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$file" | awk '{ print $1 }'
    return 0
  fi

  printf '%s\n' 'sha256sum or shasum -a 256 is required to verify evmgo' >&2
  return 1
}

platform_os() {
  os_name=$(uname -s 2>/dev/null || printf '%s' 'unknown')

  case "$os_name" in
    Linux)
      printf '%s\n' 'linux'
      ;;
    Darwin)
      printf '%s\n' 'darwin'
      ;;
    *)
      printf 'unsupported operating system: %s\n' "$os_name" >&2
      return 1
      ;;
  esac
}

platform_arch() {
  arch_name=$(uname -m 2>/dev/null || printf '%s' 'unknown')

  case "$arch_name" in
    x86_64 | amd64)
      printf '%s\n' 'amd64'
      ;;
    aarch64 | arm64)
      printf '%s\n' 'arm64'
      ;;
    *)
      printf 'unsupported architecture: %s\n' "$arch_name" >&2
      return 1
      ;;
  esac
}

profile_file_for_shell() {
  shell_path=${SHELL:-}
  shell_name=${shell_path##*/}
  home_dir=${HOME:-/nonexistent}

  case "$shell_name" in
    zsh)
      printf '%s\n' "$home_dir/.zshrc"
      ;;
    bash)
      for candidate in "$home_dir/.bash_profile" "$home_dir/.bash_login" "$home_dir/.profile"; do
        if [ -f "$candidate" ]; then
          printf '%s\n' "$candidate"
          return 0
        fi
      done

      printf '%s\n' "$home_dir/.bashrc"
      ;;
    *)
      printf '%s\n' "$home_dir/.profile"
      ;;
  esac
}

append_path_line_if_missing() {
  profile=$1
  bin_dir=$2
  line=$(printf 'export PATH="%s:$PATH"' "$bin_dir")

  mkdir -p "$(dirname "$profile")"
  if [ ! -f "$profile" ]; then
    : > "$profile"
  fi

  if grep -Fqx "$line" "$profile" 2>/dev/null; then
    return 0
  fi

  if [ -s "$profile" ]; then
    printf '\n' >> "$profile"
  fi

  printf '%s\n' "$line" >> "$profile"
}

resolve_version() {
  if [ -n "${VERSION:-}" ]; then
    printf '%s\n' "$VERSION"
    return 0
  fi

  metadata_file=$(mktemp)
  if ! download_to_file "${GITHUB_API_ROOT}/releases" "$metadata_file"; then
    rm -f "$metadata_file"
    return 1
  fi

  metadata=$(cat "$metadata_file")
  rm -f "$metadata_file"

  latest_version_from_json "$metadata"
}

main() {
  version=$(resolve_version)
  os=$(platform_os)
  arch=$(platform_arch)
  archive=$(asset_name "$os" "$arch" "$version")

  tmpdir=$(mktemp -d)
  trap 'rm -rf "$tmpdir"' EXIT HUP INT TERM

  archive_path="$tmpdir/$archive"
  checksums_path="$tmpdir/checksums.txt"
  extract_dir="$tmpdir/extracted"
  target_path="${DEFAULT_UNIX_BIN_DIR}/evmgo"

  mkdir -p "$extract_dir" "$DEFAULT_UNIX_BIN_DIR"

  download_to_file "${RELEASE_DOWNLOAD_ROOT}/${version}/checksums.txt" "$checksums_path"
  download_to_file "${RELEASE_DOWNLOAD_ROOT}/${version}/${archive}" "$archive_path"

  expected_checksum=$(checksum_for_asset "$checksums_path" "$archive")
  if [ -z "$expected_checksum" ]; then
    printf 'checksum not found for asset %s\n' "$archive" >&2
    exit 1
  fi

  actual_checksum=$(sha256_file "$archive_path")
  if [ "$expected_checksum" != "$actual_checksum" ]; then
    printf 'checksum mismatch for %s\n' "$archive" >&2
    printf 'expected %s\n' "$expected_checksum" >&2
    printf 'actual   %s\n' "$actual_checksum" >&2
    exit 1
  fi

  if command -v tar >/dev/null 2>&1; then
    tar -xzf "$archive_path" -C "$extract_dir"
  else
    printf '%s\n' 'tar is required to extract the release archive' >&2
    exit 1
  fi

  binary_path=$(find "$extract_dir" -type f -name evmgo | head -n 1)
  if [ -z "$binary_path" ]; then
    printf '%s\n' 'evmgo binary not found in extracted archive' >&2
    exit 1
  fi

  cp "$binary_path" "$target_path"
  chmod 755 "$target_path"

  profile=$(profile_file_for_shell)
  path_line=$(printf 'export PATH="%s:$PATH"' "$DEFAULT_UNIX_BIN_DIR")
  path_updated=0
  if grep -Fqx "$path_line" "$profile" 2>/dev/null; then
    path_updated=0
  else
    append_path_line_if_missing "$profile" "$DEFAULT_UNIX_BIN_DIR"
    path_updated=1
  fi

  printf 'Installed %s %s to %s\n' "$PROJECT_REPO" "$version" "$target_path"
  if [ "$path_updated" -eq 1 ]; then
    printf 'Added %s to PATH in %s\n' "$DEFAULT_UNIX_BIN_DIR" "$profile"
    printf 'Restart your shell or run: . "%s"\n' "$profile"
  else
    printf 'Restart your shell if %s is not yet available on PATH.\n' "$target_path"
  fi
}

if [ "${EVMGO_INSTALLER_TESTING:-0}" != "1" ]; then
  main "$@"
fi
