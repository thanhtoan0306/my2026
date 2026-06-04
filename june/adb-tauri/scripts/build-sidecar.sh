#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SIDECAR_DIR="$ROOT/sidecar"
BIN_DIR="$ROOT/src-tauri/binaries"
mkdir -p "$BIN_DIR"

cd "$SIDECAR_DIR"

build_one() {
  local goos=$1 goarch=$2 triple=$3
  echo "==> building adb-sidecar for $triple"
  GOOS=$goos GOARCH=$goarch CGO_ENABLED=0 go build -ldflags="-s -w" -o "$BIN_DIR/adb-sidecar-$triple" .
}

HOST="$(uname -s)"
ARCH="$(uname -m)"
case "$HOST-$ARCH" in
  Darwin-arm64|Darwin-aarch64)
    build_one darwin arm64 aarch64-apple-darwin
    ;;
  Darwin-x86_64)
    build_one darwin amd64 x86_64-apple-darwin
    ;;
  Linux-x86_64)
    build_one linux amd64 x86_64-unknown-linux-gnu
    ;;
  Linux-aarch64|Linux-arm64)
    build_one linux arm64 aarch64-unknown-linux-gnu
    ;;
  *)
    echo "Unsupported host: $HOST $ARCH — build manually with GOOS/GOARCH"
    build_one darwin arm64 aarch64-apple-darwin
    ;;
esac

echo "==> sidecar binaries in $BIN_DIR"
