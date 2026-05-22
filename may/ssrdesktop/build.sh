#!/usr/bin/env bash
set -euo pipefail

export PATH="/opt/homebrew/bin:${PATH:-}"

ROOT="$(cd "$(dirname "$0")" && pwd)"
APP_NAME="SSRDesktop"
BUILD_DIR="$ROOT/build"
APP_BUNDLE="$BUILD_DIR/${APP_NAME}.app"
ICON_SRC="$ROOT/assets/icon.png"
ICONSET="$BUILD_DIR/icon.iconset"

cd "$ROOT"

echo "==> tidy modules"
go mod tidy

echo "==> build binary (Universal: Apple Silicon + Intel)"
mkdir -p "$BUILD_DIR"
LDFLAGS="-s -w"
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags="$LDFLAGS" -o "$BUILD_DIR/${APP_NAME}_arm64" .
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags="$LDFLAGS" -o "$BUILD_DIR/${APP_NAME}_amd64" .
lipo -create -output "$BUILD_DIR/$APP_NAME" \
  "$BUILD_DIR/${APP_NAME}_arm64" "$BUILD_DIR/${APP_NAME}_amd64"
rm -f "$BUILD_DIR/${APP_NAME}_arm64" "$BUILD_DIR/${APP_NAME}_amd64"
file "$BUILD_DIR/$APP_NAME"

echo "==> prepare app icon"
if [[ ! -f "$ICON_SRC" ]]; then
  echo "Missing $ICON_SRC — add a 1024x1024 PNG or run from repo with assets/icon.png"
  exit 1
fi

rm -rf "$ICONSET"
mkdir -p "$ICONSET"
for size in 16 32 128 256 512; do
  sips -z $size $size "$ICON_SRC" --out "$ICONSET/icon_${size}x${size}.png" >/dev/null
  double=$((size * 2))
  sips -z $double $double "$ICON_SRC" --out "$ICONSET/icon_${size}x${size}@2x.png" >/dev/null
done
iconutil -c icns "$ICONSET" -o "$BUILD_DIR/AppIcon.icns"

echo "==> assemble ${APP_NAME}.app"
rm -rf "$APP_BUNDLE"
mkdir -p "$APP_BUNDLE/Contents/MacOS"
mkdir -p "$APP_BUNDLE/Contents/Resources"

cp "$BUILD_DIR/$APP_NAME" "$APP_BUNDLE/Contents/MacOS/"
cp "$BUILD_DIR/AppIcon.icns" "$APP_BUNDLE/Contents/Resources/AppIcon.icns"

cat >"$APP_BUNDLE/Contents/Info.plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleDevelopmentRegion</key>
  <string>en</string>
  <key>CFBundleExecutable</key>
  <string>${APP_NAME}</string>
  <key>CFBundleIconFile</key>
  <string>AppIcon</string>
  <key>CFBundleIdentifier</key>
  <string>com.my2026.ssrdesktop</string>
  <key>CFBundleInfoDictionaryVersion</key>
  <string>6.0</string>
  <key>CFBundleName</key>
  <string>${APP_NAME}</string>
  <key>CFBundlePackageType</key>
  <string>APPL</string>
  <key>CFBundleShortVersionString</key>
  <string>1.0.0</string>
  <key>CFBundleVersion</key>
  <string>1</string>
  <key>LSMinimumSystemVersion</key>
  <string>11.0</string>
  <key>NSHighResolutionCapable</key>
  <true/>
  <key>NSPrincipalClass</key>
  <string>NSApplication</string>
</dict>
</plist>
EOF

echo "==> done: $APP_BUNDLE"
echo "Open with: open \"$APP_BUNDLE\""
