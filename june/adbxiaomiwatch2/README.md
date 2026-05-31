# HelloWatch — Xiaomi Watch 2

Minimal Wear OS Hello World app. Builds from the command line (no Android Studio).

## Quick start

```bash
cd june/adbxiaomiwatch2
./setup.sh    # once: JDK 17, Android SDK, Gradle wrapper
./build.sh    # build app-debug.apk
./install.sh  # connect + install on watch (default 192.168.1.5:45741)
```

Custom watch address:

```bash
./install.sh 192.168.1.5:45741
# or
ADB_HOST=192.168.1.5:45741 ./install.sh
```

## Watch setup

1. Settings → Developer options → **Wireless debugging** ON
2. Note IP address and port (e.g. `192.168.1.5:45741`)
3. Mac and watch on the same Wi‑Fi

## Manual ADB

```bash
adb connect 192.168.1.5:45741
adb devices
adb install -r app/build/outputs/apk/debug/app-debug.apk
adb shell monkey -p com.example.hellowatch -c android.intent.category.LAUNCHER 1
```

## Project

| Item | Value |
|------|-------|
| Package | `com.example.hellowatch` |
| Min SDK | 30 (Wear OS 3+) |
| Output | `app/build/outputs/apk/debug/app-debug.apk` |
