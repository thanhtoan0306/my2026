# SSR Desktop (Go Hello World)

A minimal **server-side rendered** Go app shown in a native macOS window (WebKit via `webview_go`), packaged as `SSRDesktop.app` with a custom icon.

## How it works

1. Go embeds `templates/` and `static/` and serves them on `127.0.0.1` with a random free port.
2. `html/template` renders the hello page on each request (SSR).
3. A native webview window loads that local URL.
4. `build.sh` produces `build/SSRDesktop.app` with `AppIcon.icns`.

## Requirements

- macOS 11+
- Xcode Command Line Tools (`xcode-select --install`)
- Go 1.22+

## Build

```bash
cd may/ssrdesktop
chmod +x build.sh
./build.sh
open build/SSRDesktop.app
```

## Share with a friend

```bash
./build.sh   # universal Mac binary (Apple Silicon + Intel)
./share.sh   # creates build/dist/SSRDesktop-macOS.zip
```

Send **`build/dist/SSRDesktop-macOS.zip`**. Your friend unzips, drags `SSRDesktop.app` to Applications, then **right-click → Open** the first time (unsigned app). Details: [`learn/05-sharing-with-a-friend.md`](learn/05-sharing-with-a-friend.md).

## Run without packaging

```bash
go run .
```

## Learn

See **[`learn/`](learn/README.md)** for a short course on **SSR web vs SSR desktop** (diagrams, tables, how this repo fits).

## Project layout

```
may/ssrdesktop/
├── assets/icon.png      # App icon source (1024×1024 PNG)
├── build.sh             # Build binary + .app + .icns
├── learn/               # SSR web vs desktop explained
├── main.go
├── static/style.css
└── templates/index.html
```
