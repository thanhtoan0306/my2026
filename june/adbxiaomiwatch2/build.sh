#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
SDK="${ANDROID_HOME:-${ANDROID_SDK_ROOT:-/opt/homebrew/share/android-commandlinetools}}"

if [ ! -f "$ROOT/local.properties" ]; then
  echo "Run ./setup.sh first"
  exit 1
fi

java_home_17() {
  if /usr/libexec/java_home -v 17 >/dev/null 2>&1; then
    /usr/libexec/java_home -v 17
  elif [ -d "/opt/homebrew/opt/openjdk@17/libexec/openjdk.jdk/Contents/Home" ]; then
    echo "/opt/homebrew/opt/openjdk@17/libexec/openjdk.jdk/Contents/Home"
  elif [ -d "/usr/local/opt/openjdk@17/libexec/openjdk.jdk/Contents/Home" ]; then
    echo "/usr/local/opt/openjdk@17/libexec/openjdk.jdk/Contents/Home"
  fi
}

if java_home_17 >/dev/null 2>&1; then
  export JAVA_HOME="$(java_home_17)"
  export PATH="$JAVA_HOME/bin:$PATH"
fi

export ANDROID_HOME="${ANDROID_HOME:-${ANDROID_SDK_ROOT:-$SDK}}"
export PATH="$SDK/platform-tools:$PATH"

chmod +x "$ROOT/gradlew"
cd "$ROOT"
./gradlew assembleDebug

APK="$ROOT/app/build/outputs/apk/debug/app-debug.apk"
echo ""
echo "Built: $APK"
ls -lh "$APK"
