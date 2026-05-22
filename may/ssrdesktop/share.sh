#!/usr/bin/env bash
# Package SSRDesktop.app for sharing (zip + friend instructions).
set -euo pipefail

export PATH="/opt/homebrew/bin:${PATH:-}"

ROOT="$(cd "$(dirname "$0")" && pwd)"
APP_NAME="SSRDesktop"
BUILD_DIR="$ROOT/build"
APP_BUNDLE="$BUILD_DIR/${APP_NAME}.app"
DIST_DIR="$BUILD_DIR/dist"
ZIP_NAME="${APP_NAME}-macOS.zip"
VERSION="$(/usr/libexec/PlistBuddy -c 'Print CFBundleShortVersionString' "$APP_BUNDLE/Contents/Info.plist" 2>/dev/null || echo "1.0.0")"

cd "$ROOT"

if [[ ! -d "$APP_BUNDLE" ]]; then
  echo "App not found. Run ./build.sh first."
  exit 1
fi

echo "==> ad-hoc sign (local trust only, not notarized)"
codesign --force --deep --sign - "$APP_BUNDLE" 2>/dev/null || true

echo "==> write friend instructions"
mkdir -p "$DIST_DIR"
cat >"$DIST_DIR/INSTALL.txt" <<'EOF'
SSR Desktop — macOS install (for your friend)
=============================================

Requirements
  • Mac with macOS 11 (Big Sur) or newer
  • Apple Silicon (M1/M2/M3) or Intel — this build supports both

Steps
  1. Unzip SSRDesktop-macOS.zip
  2. Drag SSRDesktop.app into Applications (or Desktop)
  3. First launch — macOS may block unsigned apps:

     Option A (easiest)
       Right-click SSRDesktop.app → Open → click Open again

     Option B (Terminal)
       xattr -cr /Applications/SSRDesktop.app
       open /Applications/SSRDesktop.app

  4. If macOS still warns: System Settings → Privacy & Security
     → allow SSRDesktop (or “Open Anyway”)

You do NOT need Go or Xcode — only the .app in the zip.

Troubleshooting
  • “App is damaged” → run: xattr -cr path/to/SSRDesktop.app
  • App bounces and quits → friend’s macOS may be older than 11.0
  • Sent via WeChat/email → unzip fully before moving the .app

Built by a friend — not signed with an Apple Developer ID.
EOF

echo "==> create zip (INSTALL.txt + SSRDesktop.app at top level)"
rm -f "$DIST_DIR/$ZIP_NAME"
(
  cd "$DIST_DIR"
  rm -rf _zip_stage
  mkdir -p _zip_stage
  cp INSTALL.txt "_zip_stage/"
  ditto "$APP_BUNDLE" "_zip_stage/${APP_NAME}.app"
  cd _zip_stage
  zip -r -y "../$ZIP_NAME" INSTALL.txt "${APP_NAME}.app"
  cd ..
  rm -rf _zip_stage
)

BYTES=$(stat -f%z "$DIST_DIR/$ZIP_NAME" 2>/dev/null || stat -c%s "$DIST_DIR/$ZIP_NAME")
MB=$(echo "scale=1; $BYTES / 1048576" | bc)

echo ""
echo "Share this file with your friend:"
echo "  $DIST_DIR/$ZIP_NAME  (~${MB} MB)"
echo ""
echo "Ways to send:"
echo "  • AirDrop (Mac ↔ Mac)"
echo "  • iCloud Drive / Google Drive / Dropbox"
echo "  • USB stick"
echo "  • Avoid email if over 25 MB — use cloud link instead"
echo ""
echo "Optional: include INSTALL.txt (also inside the zip folder layout)"
echo "  $DIST_DIR/INSTALL.txt"
