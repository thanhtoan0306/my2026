# curl → JSON Desktop (Go + WebView)

Desktop app wrapping the Go SSR version: local HTTP server + native window (WebKit webview).

## Dev run

```bash
cd may/29mayjsoncurl-desktop
go mod tidy
go run .
```

## Build macOS app (with icon)

```bash
chmod +x build.sh
./build.sh
open build/JsonCurl.app
```

Output: `build/JsonCurl.app` with custom icon from `assets/icon.png`.

## How it works

1. Starts embedded Go HTTP server on a random localhost port
2. Opens a native window loading that URL
3. Same curl validation, JSON beautify, and copy UI as `may/29mayjsoncurl-go`

Based on the pattern in `may/ssrdesktop`.
