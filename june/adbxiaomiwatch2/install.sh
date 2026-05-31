#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
APK="$ROOT/app/build/outputs/apk/debug/app-debug.apk"
HOST="${1:-${ADB_HOST:-192.168.1.5:45741}}"

if [ ! -f "$APK" ]; then
  echo "APK not found. Run ./build.sh first"
  exit 1
fi

echo "==> Connecting to $HOST"
adb connect "$HOST"

echo "==> Installing HelloWatch"
adb -s "$HOST" install -r "$APK"

echo "==> Launching app"
adb -s "$HOST" shell monkey -p com.example.hellowatch -c android.intent.category.LAUNCHER 1

echo "Done. Open HelloWatch on your Xiaomi Watch 2."
