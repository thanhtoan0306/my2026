# SSR Tauri (hello world v1.0.0)

Same **SSR hello world** as [`may/ssrdesktop`](../ssrdesktop/README.md), using **Tauri 2** (Rust + system WebView).

| | `ssrdesktop` | `ssrtauri` |
|---|--------------|------------|
| Shell | Go + `webview_go` | **Tauri 2** |
| SSR server | Go `net/http` | **Rust Axum** |
| UI | Same template + CSS | Same layout |
| Version | 1.0.0 | 1.0.0 |

## Requirements

- Rust (`rustup`)
- Node.js + npm
- macOS: Xcode CLT (for building)

## Dev (hot reload window)

```bash
cd may/ssrtauri
npm install
npm run tauri dev
```

## Build macOS `.app`

```bash
chmod +x build.sh
./build.sh
open build/SSRTauri.app
```

## How it works

1. Axum serves `127.0.0.1:<random port>/` with server-rendered HTML.
2. Tauri opens a window to that URL (same pattern as Go desktop).
3. Form `?name=` triggers SSR on each request.

## Project layout

```
may/ssrtauri/
├── src-tauri/
│   ├── src/lib.rs          # Axum SSR + Tauri window
│   ├── templates/index.html
│   └── static/style.css
├── package.json
└── build.sh
```

## Android APK (Tauri mobile)

Same Rust SSR code, built as APK for phone / **TV box**.

**Needs ~3GB free disk** for NDK + Rust Android targets.

```bash
cd may/ssrtauri
chmod +x setup-android.sh build-android.sh share-android.sh

./setup-android.sh      # once: NDK + rustup Android targets
./build-android.sh      # → build/SSRTauri-1.0.0-android-debug.apk
./share-android.sh      # → build/dist/SSRTauri-android.zip

adb install -r build/SSRTauri-1.0.0-android-debug.apk
```

| | `ssrapk` (Kotlin) | `ssrtauri` Android |
|---|-------------------|---------------------|
| Backend | Kotlin | **Same Rust as Mac** |
| Build | Gradle only | NDK + Rust + Tauri |
| One repo with Mac | No | **Yes** |

If `setup-android.sh` fails with **No space left on device**, free disk first (Trash, `~/Library/Android/sdk/.temp`, old `target/` folders).

## Compare

- **WebView (Go)** — smallest stack if you already use Go.
- **Tauri** — Rust backend, small binary, good if you want Rust + optional web UI later.
- **Electron** — Node + Chromium, heavier.
