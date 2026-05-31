#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
SDK="${ANDROID_HOME:-${ANDROID_SDK_ROOT:-$HOME/Android/Sdk}}"
BREW_SDK="/opt/homebrew/share/android-commandlinetools"
CMDLINE="$SDK/cmdline-tools/latest/bin/sdkmanager"
if [ ! -x "$CMDLINE" ] && [ -x "/opt/homebrew/bin/sdkmanager" ]; then
  CMDLINE="/opt/homebrew/bin/sdkmanager"
  SDK="$BREW_SDK"
fi

echo "==> HelloWatch setup (no Android Studio)"

if ! command -v brew >/dev/null 2>&1; then
  echo "Homebrew is required. Install from https://brew.sh"
  exit 1
fi

java_home_17() {
  if /usr/libexec/java_home -v 17 >/dev/null 2>&1; then
    /usr/libexec/java_home -v 17
  elif [ -d "/opt/homebrew/opt/openjdk@17/libexec/openjdk.jdk/Contents/Home" ]; then
    echo "/opt/homebrew/opt/openjdk@17/libexec/openjdk.jdk/Contents/Home"
  elif [ -d "/usr/local/opt/openjdk@17/libexec/openjdk.jdk/Contents/Home" ]; then
    echo "/usr/local/opt/openjdk@17/libexec/openjdk.jdk/Contents/Home"
  else
    return 1
  fi
}

if ! java_home_17 >/dev/null 2>&1; then
  echo "==> Installing OpenJDK 17"
  brew install openjdk@17
fi

export JAVA_HOME="$(java_home_17)"
export PATH="$JAVA_HOME/bin:$PATH"
echo "JAVA_HOME=$JAVA_HOME"

if ! command -v adb >/dev/null 2>&1; then
  echo "==> Installing platform-tools (adb)"
  brew install android-platform-tools
fi

mkdir -p "$SDK/cmdline-tools"

if [ ! -x "$CMDLINE" ]; then
  echo "==> Installing Android command-line tools"
  brew install --cask android-commandlinetools 2>/dev/null || true

  if [ ! -x "$CMDLINE" ]; then
    BREW_SDK="$(brew --prefix)/share/android-commandlinetools"
    if [ -d "$BREW_SDK/cmdline-tools/latest" ]; then
      ln -sfn "$BREW_SDK/cmdline-tools/latest" "$SDK/cmdline-tools/latest"
    fi
  fi
fi

if [ ! -x "$CMDLINE" ]; then
  echo "Could not find sdkmanager. Set ANDROID_HOME and install command-line tools:"
  echo "  https://developer.android.com/studio#command-line-tools-only"
  exit 1
fi

export ANDROID_HOME="$SDK"
export PATH="$SDK/cmdline-tools/latest/bin:$SDK/platform-tools:$PATH"

echo "==> Accepting SDK licenses"
yes | "$CMDLINE" --sdk_root="$SDK" --licenses >/dev/null 2>&1 || true

echo "==> Installing SDK packages"
"$CMDLINE" --sdk_root="$SDK" "platform-tools" "platforms;android-34" "build-tools;34.0.0"

cat > "$ROOT/local.properties" <<EOF
sdk.dir=$SDK
EOF

if [ ! -f "$ROOT/gradle/wrapper/gradle-wrapper.jar" ]; then
  echo "==> Downloading Gradle wrapper"
  if command -v gradle >/dev/null 2>&1; then
    (cd "$ROOT" && gradle wrapper --gradle-version 8.7)
  else
    brew install gradle
    (cd "$ROOT" && gradle wrapper --gradle-version 8.7)
  fi
fi

chmod +x "$ROOT/gradlew"
echo ""
echo "Setup complete. Run: ./build.sh"
