#!/usr/bin/env bash
# Run adb with Android SDK on PATH (works before ~/.zshrc is reloaded).
export ANDROID_HOME="${ANDROID_HOME:-$HOME/Library/Android/sdk}"
export PATH="$ANDROID_HOME/platform-tools:$PATH"
exec adb "$@"
