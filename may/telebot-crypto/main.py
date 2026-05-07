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
        # Strip optional quotes
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
        raise RuntimeError(f"HTTP {e.code}: {body[:400]}") from e
    except Exception as e:
        raise RuntimeError(f"HTTP request failed: {e}") from e

    try:
        return json.loads(raw.decode("utf-8"))
    except Exception as e:
        raise RuntimeError(f"Invalid JSON response: {e}") from e


def _now_utc_label() -> str:
    return datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M UTC")


@dataclass(frozen=True)
class PriceRow:
    label: str
    usd: float | None
    usd_24h_change: float | None


def fetch_prices(*, coingecko_ids: list[str]) -> list[PriceRow]:
    # CoinGecko "simple price" API (no key required).
    # Docs: https://www.coingecko.com/en/api/documentation
    ids = ",".join([x.strip() for x in coingecko_ids if x.strip()])
    if not ids:
        raise SystemExit("COINGECKO_IDS is empty.")

    params = {
        "ids": ids,
        "vs_currencies": "usd",
        "include_24hr_change": "true",
    }
    url = "https://api.coingecko.com/api/v3/simple/price?" + urllib.parse.urlencode(params)
    j = _http_json(url)

    out: list[PriceRow] = []
    for coin_id in coingecko_ids:
        item = j.get(coin_id) if isinstance(j, dict) else None
        usd = None
        chg = None
        if isinstance(item, dict):
            try:
                usd = float(item.get("usd")) if item.get("usd") is not None else None
            except Exception:
                usd = None
            try:
                chg = float(item.get("usd_24h_change")) if item.get("usd_24h_change") is not None else None
            except Exception:
                chg = None

        out.append(PriceRow(label=coin_id, usd=usd, usd_24h_change=chg))
    return out


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


def telegram_send_message(*, bot_token: str, chat_id: str, text: str) -> None:
    url = f"https://api.telegram.org/bot{bot_token}/sendMessage"
    payload = urllib.parse.urlencode(
        {
            "chat_id": chat_id,
            "text": text,
            "disable_web_page_preview": "true",
        }
    ).encode("utf-8")
    j = _http_json(url, method="POST", headers={"Content-Type": "application/x-www-form-urlencoded"}, data=payload)
    if not isinstance(j, dict) or not j.get("ok"):
        raise RuntimeError(f"Telegram sendMessage failed: {j}")


def run_once() -> None:
    bot_token = _env("TELEGRAM_BOT_TOKEN")
    chat_id = _env("TELEGRAM_CHAT_ID")
    ids_raw = _env("COINGECKO_IDS", required=False) or "bitcoin,dogecoin,bio"
    ids = [x.strip() for x in ids_raw.split(",") if x.strip()]

    print(f"Using chat_id={chat_id} ids={','.join(ids)}")
    rows = fetch_prices(coingecko_ids=ids)
    msg = format_message(rows)
    telegram_send_message(bot_token=bot_token, chat_id=chat_id, text=msg)
    print("Sent:", _now_utc_label())


def main() -> None:
    _load_dotenv_if_present()
    if "--once" in sys.argv:
        run_once()
        return

    interval_s_raw = _env("INTERVAL_SECONDS", required=False) or str(2 * 60 * 60)
    try:
        interval_s = int(interval_s_raw)
    except Exception:
        raise SystemExit("INTERVAL_SECONDS must be an integer number of seconds.")
    if interval_s < 60:
        raise SystemExit("INTERVAL_SECONDS must be >= 60.")

    # Send immediately, then every interval.
    while True:
        try:
            run_once()
        except Exception as e:
            # Keep the loop alive; send failures are common (network, rate limits, etc).
            print("ERROR:", str(e))
        time.sleep(interval_s)


if __name__ == "__main__":
    main()

