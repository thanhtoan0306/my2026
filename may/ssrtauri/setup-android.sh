#!/usr/bin/env bash
# Install Android NDK + Rust targets for Tauri Android builds.
set -euo pipefail

export PATH="/opt/homebrew/bin:${PATH:-}"
export ANDROID_HOME="${ANDROID_HOME:-$HOME/Library/Android/sdk}"
export JAVA_HOME="${JAVA_HOME:-/Applications/Android Studio.app/Contents/jbr/Contents/Home}"
export PATH="$JAVA_HOME/bin:$ANDROID_HOME/platform-tools:$PATH"

CMDLINE_ZIP="/tmp/commandlinetools-mac.zip"
CMDLINE_DIR="$ANDROID_HOME/cmdline-tools/latest"
SDKMANAGER="$CMDLINE_DIR/bin/sdkmanager"

echo "==> ANDROID_HOME=$ANDROID_HOME"

FREE_MB=$(df -m / | awk 'NR==2 {print $4}')
if [[ "${FREE_MB:-0}" -lt 3000 ]]; then
  echo "ERROR: Need at least 3GB free disk (have ~${FREE_MB}MB)."
  echo "Free Trash / old SDK caches, then run ./setup-android.sh again."
  exit 1
fi

if [[ ! -x "$SDKMANAGER" ]]; then
  echo "==> install Android command-line tools"
  mkdir -p "$ANDROID_HOME/cmdline-tools"
  curl -fsSL -o "$CMDLINE_ZIP" \
    "https://dl.google.com/android/repository/commandlinetools-mac-11076708_latest.zip"
  rm -rf /tmp/cmdline-tools-unzip
  unzip -q "$CMDLINE_ZIP" -d /tmp/cmdline-tools-unzip
  rm -rf "$CMDLINE_DIR"
  mkdir -p "$ANDROID_HOME/cmdline-tools"
  mv /tmp/cmdline-tools-unzip/cmdline-tools "$ANDROID_HOME/cmdline-tools/latest"
fi

export PATH="$CMDLINE_DIR/bin:$PATH"

if [[ ! -d "$ANDROID_HOME/ndk" ]] || [[ -z "$(ls -A "$ANDROID_HOME/ndk" 2>/dev/null)" ]]; then
  echo "==> install NDK (may take a few minutes)"
  yes | sdkmanager --licenses >/dev/null || true
  sdkmanager "ndk;27.2.12479018" "platforms;android-34" "build-tools;34.0.0"
fi

export NDK_HOME="$ANDROID_HOME/ndk/$(ls -1 "$ANDROID_HOME/ndk" | sort -V | tail -1)"
echo "==> NDK_HOME=$NDK_HOME"

echo "==> Rust Android targets"
rustup target add aarch64-linux-android armv7-linux-androideabi i686-linux-android x86_64-linux-android

echo "==> Android environment ready"
