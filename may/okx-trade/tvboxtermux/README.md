# OKX Positions Monitor (Go SSR)

Server-side rendering version of [`../index.html`](../index.html) for **TV box / Termux**: OKX REST calls run on the Go server (no browser CORS, no CryptoJS).

## Features

- Same dark OKX-style UI as the static page
- HMAC-SHA256 signing on the server (`/api/v5/account/positions`)
- Close position via `POST /api/v5/trade/close-position`
- Session cookie stores API credentials for 7 days (local use)
- OKX private **WebSocket** (`positions` / SWAP) on the Go server
- REST bootstrap once, then WS push; HTMX re-renders every 1s from memory
- Demo account: select **Demo** and server sends `x-simulated-trading: 1`

## Run

```bash
cd may/okx-trade/tvboxtermux
go run .
```

Open http://127.0.0.1:8091

### Termux (`u0_a66@192.168.1.153 -p 8022`)

**Mac** (leave Termux SSH open; run in a second terminal):

```bash
cd may/okx-trade/tvboxtermux
chmod +x deploy-to-termux.sh install-termux.sh
./deploy-to-termux.sh
```

**Termux** (paste in your existing SSH session):

```bash
bash ~/okx-ssr/install-termux.sh
BIND=0.0.0.0 ~/okx-ssr/okx-ssr
```

Open from Mac browser: `http://192.168.1.153:8091`

Optional API env on Termux:

```bash
export OKX_API_KEY=...
export OKX_SECRET_KEY=...
export OKX_PASSPHRASE=...
```

Default bind is `127.0.0.1`; use `BIND=0.0.0.0` on Termux so LAN clients can reach it.

## Routes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Dashboard (SSR) |
| POST | `/connect` | Save session, fetch positions |
| POST | `/refresh` | Manual REST refresh |
| POST | `/close` | Market close position |
| POST | `/disconnect` | Clear session |

## vs static `index.html`

| | Static HTML | Go SSR |
|--|-------------|--------|
| OKX auth | Browser + CryptoJS | Server HMAC |
| CORS | Often blocked | N/A |
| Live updates | WebSocket in browser | Server poll + meta refresh |
| Credentials | In page memory | HttpOnly session cookie |
