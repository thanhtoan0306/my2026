#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
export ANDROID_HOME="${ANDROID_HOME:-$HOME/Library/Android/sdk}"
export JAVA_HOME="${JAVA_HOME:-/Applications/Android Studio.app/Contents/jbr/Contents/Home}"
export PATH="$JAVA_HOME/bin:$ANDROID_HOME/platform-tools:$PATH"

cd "$ROOT"

if [[ ! -x "$ROOT/gradlew" ]]; then
  echo "==> bootstrap Gradle wrapper"
  if command -v gradle >/dev/null 2>&1; then
    gradle wrapper --gradle-version 8.2
  else
    GRADLE_ZIP="/tmp/gradle-8.2-bin.zip"
    GRADLE_DIR="/tmp/gradle-8.2"
    if [[ ! -x "$GRADLE_DIR/bin/gradle" ]]; then
      curl -fsSL -o "$GRADLE_ZIP" https://services.gradle.org/distributions/gradle-8.2-bin.zip
      rm -rf "$GRADLE_DIR"
      unzip -q "$GRADLE_ZIP" -d /tmp
    fi
    "$GRADLE_DIR/bin/gradle" wrapper --gradle-version 8.2
  fi
  chmod +x gradlew
fi

if [[ ! -d "$ANDROID_HOME/platforms/android-36" ]] && [[ ! -d "$ANDROID_HOME/platforms/android-34" ]]; then
  echo "Install Android SDK platform (34+): Android Studio → SDK Manager"
  exit 1
fi

echo "==> assemble debug APK"
./gradlew assembleDebug --no-daemon

APK="$ROOT/app/build/outputs/apk/debug/app-debug.apk"
OUT="$ROOT/build"
mkdir -p "$OUT"
cp "$APK" "$OUT/SSRApk-1.0.0-debug.apk"

echo "==> done: $OUT/SSRApk-1.0.0-debug.apk"
echo "Install on TV box / phone:"
echo "  adb install -r \"$OUT/SSRApk-1.0.0-debug.apk\""
