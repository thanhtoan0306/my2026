## telebot-crypto

Telegram bot that sends BTC/DOGE/BIO price notifications every 2 hours (configurable).

### 1) Create Telegram bot token

- Talk to `@BotFather` in Telegram
- Run `/newbot`
- Copy the token → set it as `TELEGRAM_BOT_TOKEN`

### 2) Get your `TELEGRAM_CHAT_ID`

- Open your bot in Telegram and send it a message (e.g. `hi`)
- Then run:

```bash
cd may/telebot-crypto
export TELEGRAM_BOT_TOKEN="...token from BotFather..."
python3 get_chat_id.py
```

- Copy one of the printed chat ids → set as `TELEGRAM_CHAT_ID`

### 3) Configure env

```bash
cd may/telebot-crypto
cp .env.example .env
```

Edit `.env` with your values.

Notes:
- This bot now uses **Binance public API** (no key) to fetch prices.
- `COINS` (or legacy `COINGECKO_IDS`) values are mapped to Binance symbols by appending `USDT` (e.g. `BTC` → `BTCUSDT`).
- For common names, `bitcoin` → `BTC`, `dogecoin` → `DOGE`.
- If `BIOUSDT` is not listed on Binance, `BIO` will show `(not found)` — in that case tell me which BIO you mean (exchange + pair) and I’ll map it.

### 4) Run

This sends immediately once, then every 2 hours:

```bash
cd may/telebot-crypto
set -a
source .env
set +a
python3 main.py
```

