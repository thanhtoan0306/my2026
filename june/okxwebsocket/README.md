# OKX BTC Ticker

Go server that listens to the [OKX public WebSocket](https://www.okx.com/docs-v5/en/#websocket-api-public-channel) for **BTC-USDT** spot price and renders it with **SSR HTML + HTMX** (no client-side WebSocket).

## Quick start

```bash
cd june/okxwebsocket
go mod tidy
go run .
```

Open [http://127.0.0.1:8092](http://127.0.0.1:8092).

Custom port:

```bash
PORT=3000 go run .
```

## How it works

| Layer | Role |
|-------|------|
| **OKX WebSocket** | Background goroutine subscribes to `tickers` for `BTC-USDT` |
| **In-memory store** | Latest price, bid/ask, 24h stats |
| **SSR** | Full HTML page rendered on `GET /` |
| **HTMX** | Polls `/fragment/price` every 1s and `/fragment/status` every 2s for partial updates |

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Full page (SSR) |
| GET | `/fragment/price` | Price panel partial (HTMX) |
| GET | `/fragment/status` | Connection status partial (HTMX) |

## Stack

- Go 1.22 (`net/http` routing)
- [gorilla/websocket](https://github.com/gorilla/websocket) — OKX client
- [HTMX 2](https://htmx.org/) — live DOM swaps without JS framework
- Embedded `html/template`
