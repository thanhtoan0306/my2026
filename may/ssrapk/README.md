# SSR Apk (Android hello world)

Same **SSR hello world** as [`may/ssrdesktop`](../ssrdesktop/README.md), packaged as an **APK** for Android phones and **TV boxes**.

| | `ssrdesktop` (Mac) | `ssrapk` (Android) |
|---|---------------------|---------------------|
| Shell | WebView (macOS WebKit) | WebView (Android System WebView) |
| Server | Go `net/http` + `html/template` | Kotlin + **NanoHTTPD** (same idea) |
| URL | `http://127.0.0.1:<port>/` | Same |
| Version | 1.0.0 | 1.0.0 |
| Ship | `.app` zip | `.apk` zip |

Go is not inside the APK (Android needs Kotlin/Java for standard APK). Logic and UI match the Mac app.

## Build

Needs **Android SDK** (Android Studio or command-line tools).

```bash
cd may/ssrapk
chmod +x build.sh share.sh
./build.sh
```

Output: `build/SSRApk-1.0.0-debug.apk`

## Install on TV box

`adb` must be on your PATH (Android SDK `platform-tools`). If you see `command not found`:

```bash
source ~/.zshrc
# or once:
export PATH="$HOME/Library/Android/sdk/platform-tools:$PATH"
```

```bash
adb connect 192.168.1.50:5555   # your box IP
adb install -r build/SSRApk-1.0.0-debug.apk
```

Or use `./adb.sh connect …` / `./adb.sh install -r …` from this folder.

Open **SSR Apk** from the launcher (may appear under Apps).

## Share with friend

```bash
./share.sh
# → build/dist/SSRApk-android.zip
```

## Project layout

```
may/ssrapk/
├── app/src/main/
│   ├── assets/templates/index.html   # SSR template
│   ├── assets/static/style.css
│   └── java/.../MainActivity.kt      # WebView + local server
├── build.sh
└── share.sh
```

## Open in Android Studio

File → Open → `may/ssrapk` → Run on device/emulator.
