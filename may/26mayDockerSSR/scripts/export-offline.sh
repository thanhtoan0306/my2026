#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

OFFLINE_IMAGE="${OFFLINE_IMAGE:-docker-ssr-hello:offline}"
OUT_DIR="$ROOT/output"
TAR="$OUT_DIR/docker-ssr-hello.tar"

mkdir -p "$OUT_DIR"

echo "==> Build image..."
docker compose build

COMPOSE_IMAGE="$(docker images -q 26maydockerssr-web:latest 2>/dev/null | head -1)"
if [[ -z "$COMPOSE_IMAGE" ]]; then
  echo "Không tìm thấy image 26maydockerssr-web:latest sau build." >&2
  exit 1
fi

echo "==> Tag: $OFFLINE_IMAGE"
docker tag "$COMPOSE_IMAGE" "$OFFLINE_IMAGE"

echo "==> Export -> $TAR"
docker save -o "$TAR" "$OFFLINE_IMAGE"

ls -lh "$TAR"
echo "Done. Copy folder output/ sang máy offline rồi chạy ./scripts/run-offline.sh"
