#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
APK="$ROOT/build/SSRApk-1.0.0-debug.apk"
DIST="$ROOT/build/dist"
ZIP="$DIST/SSRApk-android.zip"

if [[ ! -f "$APK" ]]; then
  echo "Run ./build.sh first."
  exit 1
fi

mkdir -p "$DIST"
cat >"$DIST/INSTALL.txt" <<'EOF'
SSR Apk — Android / TV box install
==================================

Requirements
  • Android 7.0+ (API 24), works on phones and most TV boxes
  • Allow "Install unknown apps" for your file manager or adb

Phone / tablet
  1. Copy SSRApk-android.zip to device
  2. Unzip and open SSRApk-1.0.0-debug.apk
  3. Tap Install

TV box (ADB from Mac/PC)
  adb connect <box-ip>:5555
  adb install -r SSRApk-1.0.0-debug.apk

TV box (USB)
  adb install -r SSRApk-1.0.0-debug.apk

Same hello-world SSR as may/ssrdesktop (Mac), but APK + Android WebView.

Debug build — not for Play Store.
EOF

rm -f "$ZIP"
(cd "$DIST" && zip -j "$ZIP" "$APK" INSTALL.txt)

echo "Share: $ZIP"
ls -lh "$ZIP"
