#!/usr/bin/env bash
set -euo pipefail

export PATH="/opt/homebrew/bin:${PATH:-}"

ROOT="$(cd "$(dirname "$0")" && pwd)"
ICON_SRC="$ROOT/../ssrdesktop/assets/icon.png"
OUT="$ROOT/build"

cd "$ROOT"

if [[ ! -d node_modules ]]; then
  echo "==> npm install"
  npm install
fi

if [[ ! -f src-tauri/icons/icon.icns ]] && [[ -f "$ICON_SRC" ]]; then
  echo "==> generate icons from ssrdesktop"
  npx tauri icon "$ICON_SRC"
fi

echo "==> tauri build (release)"
npm run tauri build

APP="$ROOT/src-tauri/target/release/bundle/macos/SSRTauri.app"
if [[ -d "$APP" ]]; then
  mkdir -p "$OUT"
  rm -rf "$OUT/SSRTauri.app"
  cp -R "$APP" "$OUT/"
  echo "==> done: $OUT/SSRTauri.app"
  echo "Open: open \"$OUT/SSRTauri.app\""
else
  echo "Build finished; check src-tauri/target/release/bundle/"
fi
