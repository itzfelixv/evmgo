#!/bin/sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)

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
assert_eq "evmgo_v1.2.3_linux_amd64.tar.gz" "$asset"

tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT

checksums_file="$tmpdir/checksums.txt"
cat > "$checksums_file" <<'EOF'
0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef  evmgo_v1.2.3_linux_amd64.tar.gz
fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210  evmgo_v1.2.3_darwin_arm64.tar.gz
EOF

checksum=$(checksum_for_asset "$checksums_file" "evmgo_v1.2.3_linux_amd64.tar.gz")
assert_eq "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" "$checksum"

if output=$(sh "$script_dir/install.sh" 2>&1); then
  printf 'expected sh scripts/install.sh to fail\n' >&2
  exit 1
fi

assert_eq "installer implementation not complete yet" "$output"
