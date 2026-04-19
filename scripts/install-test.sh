#!/bin/sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)

tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT

helper_home="$tmpdir/home"
HOME="$helper_home"
SHELL=/bin/zsh
export HOME SHELL

EVMGO_INSTALLER_TESTING=1 . "$script_dir/install.sh"

assert_eq() {
  expected=$1
  actual=$2

  if [ "$expected" != "$actual" ]; then
    printf 'expected %s, got %s\n' "$expected" "$actual" >&2
    exit 1
  fi
}

asset=$(asset_name linux amd64 v1.2.3)
assert_eq "evmgo_1.2.3_linux_amd64.tar.gz" "$asset"

checksums_file="$tmpdir/checksums.txt"
cat > "$checksums_file" <<'EOF'
0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef  evmgo_1.2.3_linux_amd64.tar.gz
fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210  evmgo_1.2.3_darwin_arm64.tar.gz
EOF

checksum=$(checksum_for_asset "$checksums_file" "evmgo_1.2.3_linux_amd64.tar.gz")
assert_eq "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" "$checksum"

(
  HOME="$helper_home"
  SHELL=/bin/zsh
  export HOME SHELL

  mkdir -p "$HOME"
  : > "$HOME/.zshrc"

  profile=$(profile_file_for_shell)
  assert_eq "$HOME/.zshrc" "$profile"

  version=$(latest_version_from_json '{"tag_name":"v0.1.0","draft":false,"prerelease":false}')
  assert_eq "v0.1.0" "$version"

  append_path_line_if_missing "$HOME/.zshrc" "$DEFAULT_UNIX_BIN_DIR"
  append_path_line_if_missing "$HOME/.zshrc" "$DEFAULT_UNIX_BIN_DIR"

  expected_line=$(printf 'export PATH="%s:$PATH"' "$helper_home/.local/bin")
  actual_line=$(grep -F '.local/bin' "$HOME/.zshrc")
  assert_eq "$expected_line" "$actual_line"

  path_lines=$(grep -c '.local/bin' "$HOME/.zshrc")
  assert_eq "1" "$path_lines"
)

(
  HOME="$tmpdir/bash-profile-home"
  SHELL=/bin/bash
  export HOME SHELL

  mkdir -p "$HOME"
  : > "$HOME/.bash_profile"

  profile=$(profile_file_for_shell)
  assert_eq "$HOME/.bash_profile" "$profile"
)

(
  HOME="$tmpdir/bash-profile-fallback-home"
  SHELL=/bin/bash
  export HOME SHELL

  mkdir -p "$HOME"
  : > "$HOME/.profile"

  profile=$(profile_file_for_shell)
  assert_eq "$HOME/.profile" "$profile"
)

(
  HOME="$helper_home"
  SHELL=/bin/bash
  export HOME SHELL

  version=$(latest_version_from_json '[
    {"tag_name":"release-2026-04-01","draft":false,"prerelease":false},
    {"tag_name":"v0.2.0-rc1","draft":false,"prerelease":true},
    {"tag_name":"v0.1.0","draft":false,"prerelease":false}
  ]')
  assert_eq "v0.1.0" "$version"
)

fixture_root="$tmpdir/fixture"
api_root="$fixture_root/api"
download_root="$fixture_root/downloads"
archive_stage="$fixture_root/stage/evmgo_0.1.0_linux_amd64"

mkdir -p "$api_root" "$download_root/v0.1.0" "$archive_stage"

cat > "$api_root/releases" <<'EOF'
[
  {"tag_name":"release-2026-04-01","draft":false,"prerelease":false},
  {"tag_name":"v0.2.0-rc1","draft":false,"prerelease":true},
  {"tag_name":"v0.1.0","draft":false,"prerelease":false}
]
EOF

cat > "$archive_stage/evmgo" <<'EOF'
#!/bin/sh
printf 'fixture evmgo\n'
EOF
chmod 755 "$archive_stage/evmgo"

archive_path="$download_root/v0.1.0/evmgo_0.1.0_linux_amd64.tar.gz"
tar -czf "$archive_path" -C "$fixture_root/stage" "evmgo_0.1.0_linux_amd64"

checksum_value=$(sha256_file "$archive_path")
cat > "$download_root/v0.1.0/checksums.txt" <<EOF
$checksum_value  evmgo_0.1.0_linux_amd64.tar.gz
EOF

e2e_home="$tmpdir/e2e-home"
mkdir -p "$e2e_home"

e2e_bash_home="$tmpdir/e2e-bash-home"
mkdir -p "$e2e_bash_home"
: > "$e2e_bash_home/.bash_profile"

HOME="$e2e_bash_home" \
SHELL=/bin/bash \
PATH="$e2e_bash_home/.local/bin:$PATH" \
GITHUB_API_ROOT="file://$api_root" \
RELEASE_DOWNLOAD_ROOT="file://$download_root" \
sh "$script_dir/install.sh"

HOME="$e2e_home" \
SHELL=/bin/sh \
PATH="$e2e_home/.local/bin:$PATH" \
GITHUB_API_ROOT="file://$api_root" \
RELEASE_DOWNLOAD_ROOT="file://$download_root" \
sh "$script_dir/install.sh"

HOME="$e2e_home" \
SHELL=/bin/sh \
PATH="$e2e_home/.local/bin:$PATH" \
GITHUB_API_ROOT="file://$api_root" \
RELEASE_DOWNLOAD_ROOT="file://$download_root" \
sh "$script_dir/install.sh"

test -x "$e2e_bash_home/.local/bin/evmgo"
test -x "$e2e_home/.local/bin/evmgo"

bash_profile_lines=$(grep -c '.local/bin' "$e2e_bash_home/.bash_profile")
assert_eq "1" "$bash_profile_lines"

if [ -f "$e2e_bash_home/.bashrc" ]; then
  bashrc_lines=$(grep -c '.local/bin' "$e2e_bash_home/.bashrc" || true)
  assert_eq "0" "$bashrc_lines"
fi

profile_lines=$(grep -c '.local/bin' "$e2e_home/.profile")
assert_eq "1" "$profile_lines"
