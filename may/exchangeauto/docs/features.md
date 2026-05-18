# exchangeauto — features

Small tooling around **OKX BTC-USDT spot**: CLI quotes, a local web dashboard, boundary alerts with browser audio, and optional Telegram notifications.

---

## Market data (OKX public API)

- **Instrument:** `BTC-USDT` spot (`btc_price.py`, `FLAG="0"` live trading domain).
- **No exchange API keys** required for ticker or candles.

---

## CLI: live ticker (`btc_price.py`)

- Prints **last**, **bid**, **ask**, and **24h low–high** to stdout.
- Run: `python btc_price.py` (with deps installed).

---

## HTTP server (`server.py`)

- **Default URL:** `http://127.0.0.1:8080/`
- **Static UI:** serves `static/index.html` at `/`.
- **Telegram watcher:** background thread starts on launch when `.env` has bot credentials (see below).

### REST endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/ticker` | JSON snapshot: last, bid, ask, 24h high/low, timestamp. |
| GET | `/api/candles` | JSON closes for chart: **5m** bars, **48** points (~4h), chronological `points[].ts` / `close`. |
| GET | `/api/telegram/test-ding` | Sends Telegram message **`ding`** if configured; `{ "ok": true }` or `{ "error": "..." }`. |

### Caching headers

- JSON APIs: `Cache-Control: no-store`.
- HTML: `no-cache, no-store, must-revalidate` to reduce stale UI after deploys.

### Logging

- Access log for `/api/ticker`, `/api/candles`, `/api/telegram/test-ding` is suppressed to limit noise.

---

## Web dashboard (`static/index.html`)

- **Layout:** dark card ~**90vw × 90dvh**, flex column; chart uses remaining height.
- **Price block:** large last price (responsive font), pair label.
- **Chart:** Canvas line + area under **close** prices; **green/red** by first vs last point in window; Y-axis labels; **orange** dot on latest.
- **Bid / Ask** grid.
- **24h range** bar with gradient, marker for last within low–high.
- **Polling:** ticker every **2s**; candles every **60s**; chart redraw on window resize (re-fetch candles).
- **Errors:** shown in-page when ticker fails.

---

## Browser audio (“Bell v2”)

- Uses **Web Audio API** (multi-harmonic bell); direction uses higher vs lower base frequency for **up** vs **down** moves.
- **Autoplay:** user gesture unlocks audio (e.g. **Test ding** click).

### Boundary rules (sound only)

Priority:

1. **$1,000** steps — **10 dings** per thousand-dollar bucket crossed (when toggle on).
2. Else **$100** steps — **2 dings** per hundred-dollar bucket crossed (when toggle on).

If both boundaries fire in one jump (e.g. large spike), **only the $1k branch** runs for that poll.

### UI controls

- Toggle **$100 boundary · 2 dings** — persists as `localStorage` key `ding100`; legacy import once from `ding200` if present.
- Toggle **$1000 boundary · 10 dings** — persists as `ding1000`.
- **Test ding:** plays **2 dings** (same count as one **$100** alert), flashes price green for “up”, unlocks audio if needed.

---

## Telegram (`telegram_alert.py`)

- **Config:** `may/exchangeauto/.env` — `TELEGRAM_BOT_TOKEN`, `TELEGRAM_CHAT_ID` (see `.env.example`; trim stray slashes/quotes handled in code).
- **Watcher:** polls OKX ticker every **5s** in a daemon thread; on **$1,000** boundary cross sends a formatted message (↑/↓, level, price, bid/ask, instrument).
- **Test ding:** `send_test_ding()` sends literal **`ding`**; wired from dashboard button and `/api/telegram/test-ding`.
- **Standalone:** `python telegram_alert.py` runs watcher only (requires `.env`).
- If `.env` missing or incomplete: watcher skipped; console explains; HTTP test endpoint returns configuration error JSON.

---

## Shared Python helpers (`btc_price.py`)

- `fetch_ticker(inst_id=…)` — normalized ticker dict for UI/API/Telegram.
- `fetch_candles(inst_id=…, bar="5m", limit=48)` — `{ instId, bar, points }` for chart API.

---

## Dependencies (`requirements.txt`)

- **python-okx** — OKX REST client.
- **python-dotenv** — load `.env` for Telegram.

---

## Related docs / samples

- `docs/okx-api.md` — OKX ticker snippet reference.

---

## Repository hygiene

- `.env` is gitignored (secrets).
- `.venv/` under `may/exchangeauto/` is gitignored.
