#!/usr/bin/env python3
from __future__ import annotations

import json
import os
import sys
import time
import urllib.error
import urllib.parse
import urllib.request
from dataclasses import dataclass
from datetime import datetime, timezone
from pathlib import Path


class HttpError(RuntimeError):
    def __init__(self, status_code: int, body: str):
        super().__init__(f"HTTP {status_code}: {body}")
        self.status_code = status_code
        self.body = body


def _load_dotenv_if_present() -> None:
    """
    Lightweight .env loader (no dependencies).
    Prefers values in .env (overwrites existing env vars).
    """
    env_path = Path(__file__).resolve().parent / ".env"
    if not env_path.exists():
        return

    for raw_line in env_path.read_text(encoding="utf-8").splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#"):
            continue
        if "=" not in line:
            continue
        k, v = line.split("=", 1)
        k = k.strip()
        v = v.strip()
        if not k:
            continue
        if (v.startswith('"') and v.endswith('"')) or (v.startswith("'") and v.endswith("'")):
            v = v[1:-1]
        os.environ[k] = v


def _env(name: str, *, required: bool = True) -> str | None:
    v = os.getenv(name)
    if required and (v is None or not v.strip()):
        raise SystemExit(f"Missing env var: {name}")
    return v.strip() if v is not None else None


def _http_json(
    url: str,
    *,
    method: str = "GET",
    headers: dict[str, str] | None = None,
    data: bytes | None = None,
    timeout_s: float = 20.0,
) -> dict:
    hdrs = {
        "User-Agent": "telebot-crypto/1.0",
        "Accept": "application/json",
    }
    if headers:
        hdrs.update(headers)
    req = urllib.request.Request(url, headers=hdrs, method=method, data=data)
    try:
        with urllib.request.urlopen(req, timeout=timeout_s) as resp:
            raw = resp.read()
    except urllib.error.HTTPError as e:
        body = ""
        try:
            body = e.read().decode("utf-8", "replace")
        except Exception:
            body = ""
        raise HttpError(int(e.code), body[:400]) from e
    except Exception as e:
        raise RuntimeError(f"HTTP request failed: {e}") from e

    try:
        return json.loads(raw.decode("utf-8"))
    except Exception as e:
        raise RuntimeError(f"Invalid JSON response: {e}") from e


def _now_utc_label() -> str:
    return datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M UTC")


def _now_ts() -> int:
    return int(time.time())


def _state_path() -> Path:
    return Path(__file__).resolve().parent / "state.json"


def _load_state() -> dict:
    p = _state_path()
    if not p.exists():
        return {}
    try:
        return json.loads(p.read_text(encoding="utf-8"))
    except Exception:
        return {}


def _save_state(state: dict) -> None:
    _state_path().write_text(json.dumps(state, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


@dataclass(frozen=True)
class PriceRow:
    label: str
    usd: float | None
    usd_24h_change: float | None


def _coins_from_env() -> list[str]:
    # Prefer COINS (symbols/queries). Backwards compat: COINGECKO_IDS.
    raw = _env("COINS", required=False) or _env("COINGECKO_IDS", required=False) or "BTC,DOGE,BIO"
    return [x.strip() for x in raw.split(",") if x.strip()]


# -------- Price provider: Binance (no key) --------

def _binance_symbols_from_coins(coins: list[str]) -> list[str]:
    # Default mapping: BTC -> BTCUSDT, DOGE -> DOGEUSDT, ...
    name_map = {
        "bitcoin": "BTC",
        "btc": "BTC",
        "dogecoin": "DOGE",
        "doge": "DOGE",
    }

    out: list[str] = []
    for c in coins:
        raw = c.strip()
        if not raw:
            continue
        sym = name_map.get(raw.lower(), raw.upper())
        out.append(f"{sym}USDT")
    return out


def fetch_prices_binance(*, coins: list[str]) -> list[PriceRow]:
    symbols = _binance_symbols_from_coins(coins)
    if not symbols:
        return []

    # Batch request: /api/v3/ticker/24hr?symbols=["BTCUSDT","DOGEUSDT"]
    symbols_json = json.dumps(symbols, separators=(",", ":"))
    # Binance is picky about percent-encoding; keep JSON punctuation unescaped.
    symbols_q = urllib.parse.quote(symbols_json, safe='[],"')
    url = "https://api.binance.com/api/v3/ticker/24hr?symbols=" + symbols_q
    j = _http_json(url)
    if not isinstance(j, list):
        raise RuntimeError(f"Unexpected Binance response: {j}")

    by_symbol: dict[str, dict] = {}
    for item in j:
        if isinstance(item, dict) and isinstance(item.get("symbol"), str):
            by_symbol[item["symbol"]] = item

    rows: list[PriceRow] = []
    for coin, sym in zip(coins, symbols):
        item = by_symbol.get(sym)
        if not isinstance(item, dict):
            rows.append(PriceRow(label=coin.upper(), usd=None, usd_24h_change=None))
            continue
        try:
            usd = float(item.get("lastPrice")) if item.get("lastPrice") is not None else None
        except Exception:
            usd = None
        try:
            chg = float(item.get("priceChangePercent")) if item.get("priceChangePercent") is not None else None
        except Exception:
            chg = None
        rows.append(PriceRow(label=coin.upper(), usd=usd, usd_24h_change=chg))
    return rows


def format_message(rows: list[PriceRow]) -> str:
    lines: list[str] = [f"Crypto prices ({_now_utc_label()}):"]
    for r in rows:
        if r.usd is None:
            lines.append(f"- {r.label}: (not found)")
            continue
        if r.usd >= 1:
            price = f"${r.usd:,.2f}"
        elif r.usd >= 0.01:
            price = f"${r.usd:,.4f}"
        else:
            price = f"${r.usd:.8f}"

        if r.usd_24h_change is None:
            lines.append(f"- {r.label}: {price}")
        else:
            lines.append(f"- {r.label}: {price} ({r.usd_24h_change:+.2f}% / 24h)")
    return "\n".join(lines)


# -------- Telegram helpers --------

def telegram_send_message(*, bot_token: str, chat_id: str, text: str) -> None:
    url = f"https://api.telegram.org/bot{bot_token}/sendMessage"
    payload = urllib.parse.urlencode(
        {"chat_id": chat_id, "text": text, "disable_web_page_preview": "true"}
    ).encode("utf-8")
    j = _http_json(url, method="POST", headers={"Content-Type": "application/x-www-form-urlencoded"}, data=payload)
    if not isinstance(j, dict) or not j.get("ok"):
        raise RuntimeError(f"Telegram sendMessage failed: {j}")


def telegram_get_updates(*, bot_token: str, offset: int | None) -> dict:
    params: dict[str, str] = {"timeout": "0"}
    if offset is not None:
        params["offset"] = str(offset)
    url = f"https://api.telegram.org/bot{bot_token}/getUpdates?" + urllib.parse.urlencode(params)
    j = _http_json(url, method="GET")
    if not isinstance(j, dict) or not j.get("ok"):
        raise RuntimeError(f"Telegram getUpdates failed: {j}")
    return j


def _extract_text_and_chat_id(update: dict) -> tuple[str | None, str | None]:
    msg = update.get("message") or update.get("edited_message") or update.get("channel_post")
    if not isinstance(msg, dict):
        return None, None
    chat = msg.get("chat")
    if not isinstance(chat, dict):
        return None, None
    chat_id = chat.get("id")
    text = msg.get("text")
    return (text if isinstance(text, str) else None), (str(chat_id) if chat_id is not None else None)


def _parse_interval_command(text: str) -> int | None:
    t = text.strip()
    parts = t.split()
    if len(parts) < 2:
        return None
    cmd = parts[0].lower()
    if cmd not in ("/interval", "interval"):
        return None
    try:
        return int(parts[1])
    except Exception:
        return None


def poll_and_apply_commands(*, bot_token: str, expected_chat_id: str, state: dict) -> int | None:
    offset = state.get("telegram_offset")
    try:
        offset_i = int(offset) if offset is not None else None
    except Exception:
        offset_i = None

    j = telegram_get_updates(bot_token=bot_token, offset=offset_i)
    result = j.get("result")
    if not isinstance(result, list) or not result:
        return None

    new_offset = offset_i
    interval_change: int | None = None
    for upd in result:
        if not isinstance(upd, dict):
            continue
        upd_id = upd.get("update_id")
        try:
            upd_id_i = int(upd_id)
        except Exception:
            upd_id_i = None
        if upd_id_i is not None:
            new_offset = max(new_offset or 0, upd_id_i + 1)

        text, chat_id = _extract_text_and_chat_id(upd)
        if chat_id != expected_chat_id or not text:
            continue

        t = text.strip().lower()
        if t in ("/help", "help"):
            telegram_send_message(
                bot_token=bot_token,
                chat_id=expected_chat_id,
                text="\n".join(
                    [
                        "Commands:",
                        "- /interval <seconds>  (example: /interval 7200)",
                        "- /status",
                        "- /help",
                    ]
                ),
            )
            continue

        if t in ("/status", "status"):
            telegram_send_message(
                bot_token=bot_token,
                chat_id=expected_chat_id,
                text=f"Status: interval_seconds={state.get('interval_seconds')}",
            )
            continue

        new_interval = _parse_interval_command(text)
        if new_interval is None:
            continue
        if new_interval < 1:
            telegram_send_message(bot_token=bot_token, chat_id=expected_chat_id, text="Interval must be >= 1")
            continue
        interval_change = new_interval
        state["interval_seconds"] = new_interval
        telegram_send_message(
            bot_token=bot_token,
            chat_id=expected_chat_id,
            text=f"OK. Updated interval_seconds={new_interval}",
        )

    if new_offset is not None and new_offset != offset_i:
        state["telegram_offset"] = new_offset
    return interval_change


def _should_fetch_prices(state: dict, *, min_fetch_seconds: int) -> bool:
    now = _now_ts()
    next_at = state.get("next_fetch_at")
    if isinstance(next_at, int) and now < next_at:
        return False
    last = state.get("last_fetch_ts")
    if isinstance(last, int) and now - last < min_fetch_seconds:
        return False
    return True


def run_once(*, state: dict) -> None:
    bot_token = _env("TELEGRAM_BOT_TOKEN")
    chat_id = _env("TELEGRAM_CHAT_ID")
    coins = _coins_from_env()

    print(f"Using chat_id={chat_id} coins={','.join(coins)} provider=binance")
    rows = fetch_prices_binance(coins=coins)

    msg = format_message(rows)
    telegram_send_message(bot_token=bot_token, chat_id=chat_id, text=msg)
    print("Sent:", _now_utc_label())


def main() -> None:
    _load_dotenv_if_present()

    state = _load_state()

    bot_token = _env("TELEGRAM_BOT_TOKEN")
    chat_id = _env("TELEGRAM_CHAT_ID")

    if "--once" in sys.argv:
        run_once(state=state)
        return

    interval_s_raw = _env("INTERVAL_SECONDS", required=False) or str(2 * 60 * 60)
    try:
        interval_s = int(interval_s_raw)
    except Exception:
        raise SystemExit("INTERVAL_SECONDS must be an integer.")
    if interval_s < 1:
        raise SystemExit("INTERVAL_SECONDS must be >= 1.")

    # Chat-set interval persists.
    try:
        interval_s = int(state.get("interval_seconds", interval_s))
    except Exception:
        pass
    state["interval_seconds"] = interval_s

    min_fetch_s_raw = _env("MIN_PRICE_FETCH_SECONDS", required=False) or "60"
    try:
        min_fetch_s = int(min_fetch_s_raw)
    except Exception:
        min_fetch_s = 60
    if min_fetch_s < 1:
        min_fetch_s = 1

    while True:
        try:
            changed = poll_and_apply_commands(bot_token=bot_token, expected_chat_id=chat_id, state=state)
            if changed is not None:
                interval_s = changed
            _save_state(state)

            if _should_fetch_prices(state, min_fetch_seconds=min_fetch_s):
                run_once(state=state)
                state["last_fetch_ts"] = _now_ts()
                state.pop("next_fetch_at", None)
                _save_state(state)
        except HttpError as e:
            # Backoff on any provider rate limit.
            if e.status_code in (429, 418):
                state["next_fetch_at"] = _now_ts() + max(120, min_fetch_s)
                _save_state(state)
            print("ERROR:", str(e))
        except Exception as e:
            print("ERROR:", str(e))

        time.sleep(interval_s)


if __name__ == "__main__":
    main()

