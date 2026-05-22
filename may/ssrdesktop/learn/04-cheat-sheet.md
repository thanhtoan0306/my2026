# Cheat Sheet: SSR Web vs Desktop

## One line each

- **SSR web:** Server renders HTML → **browser** over network.
- **SSR desktop:** Server renders HTML → **WebView in .app** on `127.0.0.1`.

## Same in both

- Go `html/template`
- HTTP handlers, forms, query strings
- CSS / static files
- “View page source” shows server-filled values

## Different in both

| | Web | Desktop (`ssrdesktop`) |
|---|-----|------------------------|
| Open with | `https://...` | `open SSRDesktop.app` |
| Listen on | Public/LAN IP | `127.0.0.1` only |
| Client | Any browser | Embedded WebKit |
| Ship | Server deploy | `.app` + icon |
| Users | Many | One machine per run |

## This repo in 4 steps

1. Embed templates/static  
2. Start localhost HTTP server  
3. SSR `/` with `PageData`  
4. Show URL in native webview window  

## Commands

```bash
# Desktop bundle
cd may/ssrdesktop && ./build.sh && open build/SSRDesktop.app

# Dev (same SSR, same window)
go run .
```

## Compare locally

| Project | Style | Client |
|---------|--------|--------|
| `may/guide-CLI` | SSR web | You open browser to `:8080` |
| `may/ssrdesktop` | SSR + desktop shell | App opens webview for you |
| `may/helloword-go` | Inline HTML (not template SSR) | Browser to server |

## Mental model

```
SSR = WHERE html is built (server)
Web vs Desktop = HOW the user sees it (browser tab vs app window)
```

They are **orthogonal**: desktop does not mean “no SSR,” and SSR does not mean “must be public web.”
