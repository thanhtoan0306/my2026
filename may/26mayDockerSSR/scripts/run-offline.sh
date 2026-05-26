#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

OFFLINE_IMAGE="${OFFLINE_IMAGE:-docker-ssr-hello:offline}"
TAR="$ROOT/output/docker-ssr-hello.tar"
CONTAINER="${CONTAINER_NAME:-ssr-hello-offline}"
PORT="${PORT:-8080}"

if [[ ! -f "$TAR" ]]; then
  echo "Thiếu file: $TAR" >&2
  echo "Trên máy có mạng, chạy: ./scripts/export-offline.sh" >&2
  exit 1
fi

echo "==> Load image từ $TAR"
docker load -i "$TAR"

docker rm -f "$CONTAINER" 2>/dev/null || true

echo "==> Run $OFFLINE_IMAGE on :$PORT"
docker run -d -p "${PORT}:8080" --name "$CONTAINER" "$OFFLINE_IMAGE"

echo "OK → http://localhost:${PORT}"
