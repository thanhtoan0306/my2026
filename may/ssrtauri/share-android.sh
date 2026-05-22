#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")" && pwd)"
APK="$ROOT/build/SSRTauri-1.0.0-android-debug.apk"
DIST="$ROOT/build/dist"
ZIP="$DIST/SSRTauri-android.zip"

if [[ ! -f "$APK" ]]; then
  echo "Run ./build-android.sh first."
  exit 1
fi

mkdir -p "$DIST"
cat >"$DIST/INSTALL-android.txt" <<'EOF'
SSRTauri — Android APK (Tauri 2 + Rust SSR)
===========================================

Same hello world as SSRTauri.app on Mac, built with Tauri Android.

Install:
  adb install -r SSRTauri-1.0.0-android-debug.apk

Or copy APK to TV box / phone and open (allow unknown apps).

Requires Android 7.0+ (API 24).
EOF

rm -f "$ZIP"
(cd "$DIST" && zip -j "$ZIP" "$APK" INSTALL-android.txt)
ls -lh "$ZIP"
echo "Share: $ZIP"
