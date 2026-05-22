#!/usr/bin/env bash
# Build SSR Tauri as Android APK (debug).
set -euo pipefail

export PATH="/opt/homebrew/bin:${PATH:-}"
ROOT="$(cd "$(dirname "$0")" && pwd)"
export ANDROID_HOME="${ANDROID_HOME:-$HOME/Library/Android/sdk}"
export JAVA_HOME="${JAVA_HOME:-/Applications/Android Studio.app/Contents/jbr/Contents/Home}"

cd "$ROOT"

# Need ~3GB free for NDK + Rust Android std
FREE_MB=$(df -m / | awk 'NR==2 {print $4}')
if [[ "${FREE_MB:-0}" -lt 3000 ]]; then
  echo "ERROR: Need at least 3GB free disk space (have ~${FREE_MB}MB)."
  echo "Free space: empty Trash, remove old Xcode/Android SDK caches, then retry."
  exit 1
fi

if [[ ! -d "$ANDROID_HOME/ndk" ]] || [[ -z "$(ls -A "$ANDROID_HOME/ndk" 2>/dev/null)" ]]; then
  echo "==> setup Android NDK (first time)"
  ./setup-android.sh
fi

export NDK_HOME="${NDK_HOME:-$ANDROID_HOME/ndk/$(ls -1 "$ANDROID_HOME/ndk" | sort -V | tail -1)}"
export PATH="$JAVA_HOME/bin:$ANDROID_HOME/platform-tools:$ANDROID_HOME/cmdline-tools/latest/bin:$PATH"

echo "NDK_HOME=$NDK_HOME"

if [[ ! -d node_modules ]]; then
  npm install
fi

if [[ ! -d src-tauri/gen/android ]]; then
  echo "==> tauri android init"
  npx tauri android init --ci
fi

echo "==> tauri android build (debug APK)"
npx tauri android build --debug

APK_SRC=$(find src-tauri/gen/android -name "*.apk" -path "*/build/outputs/apk/*" 2>/dev/null | head -1)
OUT="$ROOT/build"
mkdir -p "$OUT"

if [[ -n "$APK_SRC" && -f "$APK_SRC" ]]; then
  cp "$APK_SRC" "$OUT/SSRTauri-1.0.0-android-debug.apk"
  echo "==> done: $OUT/SSRTauri-1.0.0-android-debug.apk"
  echo "Install: adb install -r \"$OUT/SSRTauri-1.0.0-android-debug.apk\""
else
  echo "APK built; search under src-tauri/gen/android/app/build/outputs/apk/"
  find src-tauri/gen/android -name "*.apk" 2>/dev/null || true
fi
