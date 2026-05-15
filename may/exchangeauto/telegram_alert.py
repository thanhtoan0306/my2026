#!/usr/bin/env python3
"""Send Telegram alerts when BTC crosses $1000 price boundaries."""

import json
import os
import threading
import time
import urllib.error
import urllib.parse
import urllib.request
from pathlib import Path
from typing import List, Optional, Tuple

from dotenv import load_dotenv

from btc_price import INST_ID, fetch_ticker

POLL_SEC = 5
THOUSAND = 1000
_ENV_PATH = Path(__file__).resolve().parent / ".env"


def load_config() -> tuple[str, str]:
    load_dotenv(_ENV_PATH)
    token = os.getenv("TELEGRAM_BOT_TOKEN", "").strip().strip("/")
    chat_id = os.getenv("TELEGRAM_CHAT_ID", "").strip().strip("\"'")
    return token, chat_id


def thousand_bucket(price: float) -> int:
    return int(price // THOUSAND)


def boundary_events(
    last: Optional[int], bucket: int
) -> Tuple[List[Tuple[bool, int]], int]:
    if last is None:
        return [], bucket
    if bucket == last:
        return [], last

    events: List[Tuple[bool, int]] = []
    if bucket > last:
        for b in range(last + 1, bucket + 1):
            events.append((True, b * THOUSAND))
    else:
        for b in range(last, bucket, -1):
            events.append((False, b * THOUSAND))
    return events, bucket


def send_telegram(token: str, chat_id: str, text: str) -> None:
    url = f"https://api.telegram.org/bot{token}/sendMessage"
    body = urllib.parse.urlencode(
        {"chat_id": chat_id, "text": text, "disable_web_page_preview": "true"}
    ).encode()
    req = urllib.request.Request(
        url,
        data=body,
        method="POST",
        headers={"Content-Type": "application/x-www-form-urlencoded"},
    )
    try:
        with urllib.request.urlopen(req, timeout=15) as resp:
            data = json.loads(resp.read().decode())
    except urllib.error.HTTPError as e:
        err = e.read().decode()
        raise RuntimeError(err or str(e)) from e

    if not data.get("ok"):
        raise RuntimeError(data.get("description", "Telegram API error"))


def format_alert(up: bool, level: int, ticker: dict) -> str:
    arrow = "↑" if up else "↓"
    last = ticker["last"]
    return (
        f"BTC/USDT {arrow} crossed ${level:,}\n"
        f"Price: ${last:,.2f}\n"
        f"Bid: ${ticker['bid']:,.2f}  Ask: ${ticker['ask']:,.2f}\n"
        f"{INST_ID} · OKX"
    )


def run_watcher(stop_event: threading.Event, token: str, chat_id: str) -> None:
    last_bucket: Optional[int] = None

    while not stop_event.is_set():
        try:
            ticker = fetch_ticker()
            bucket = thousand_bucket(ticker["last"])
            events, last_bucket = boundary_events(last_bucket, bucket)

            for up, level in events:
                msg = format_alert(up, level, ticker)
                send_telegram(token, chat_id, msg)
                print(f"Telegram: ${level:,} {'up' if up else 'down'}")
                time.sleep(0.3)
        except Exception as e:
            print(f"Telegram watcher: {e}")

        stop_event.wait(POLL_SEC)


def send_test_ding() -> None:
    token, chat_id = load_config()
    if not token or not chat_id:
        raise RuntimeError(
            "Telegram not configured — set TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID in .env"
        )
    send_telegram(token, chat_id, "ding")


def start_telegram_watcher() -> Optional[threading.Event]:
    token, chat_id = load_config()
    if not token or not chat_id:
        print(
            "Telegram: off — set TELEGRAM_BOT_TOKEN and "
            "TELEGRAM_CHAT_ID in may/exchangeauto/.env"
        )
        return None

    stop = threading.Event()
    thread = threading.Thread(
        target=run_watcher,
        args=(stop, token, chat_id),
        name="telegram-watcher",
        daemon=True,
    )
    thread.start()
    print("Telegram: watching $1000 boundaries (poll every 5s)")
    return stop


if __name__ == "__main__":
    token, chat_id = load_config()
    if not token or not chat_id:
        raise SystemExit("Missing TELEGRAM_BOT_TOKEN or TELEGRAM_CHAT_ID in .env")

    stop = threading.Event()
    try:
        run_watcher(stop, token, chat_id)
    except KeyboardInterrupt:
        print("\nStopped.")
