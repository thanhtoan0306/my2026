#!/usr/bin/env python3
from __future__ import annotations

import json
import os
import urllib.parse
import urllib.request


def _env(name: str) -> str:
    v = os.getenv(name)
    if v is None or not v.strip():
        raise SystemExit(f"Missing env var: {name}")
    return v.strip()


def main() -> None:
    token = _env("TELEGRAM_BOT_TOKEN")
    url = f"https://api.telegram.org/bot{token}/getUpdates?" + urllib.parse.urlencode({"timeout": "0"})
    req = urllib.request.Request(url, headers={"Accept": "application/json"}, method="GET")
    with urllib.request.urlopen(req, timeout=20) as resp:
        raw = resp.read().decode("utf-8")
    j = json.loads(raw)

    print("Raw response:")
    print(raw)
    print()

    # Helpfully print candidate chat ids from updates.
    result = j.get("result") if isinstance(j, dict) else None
    if not isinstance(result, list) or not result:
        print("No updates yet. Send your bot a message in Telegram, then run again.")
        return

    chat_ids: set[str] = set()
    for upd in result:
        msg = (upd or {}).get("message") or (upd or {}).get("channel_post") or (upd or {}).get("edited_message")
        chat = (msg or {}).get("chat") if isinstance(msg, dict) else None
        cid = chat.get("id") if isinstance(chat, dict) else None
        if cid is not None:
            chat_ids.add(str(cid))

    if chat_ids:
        print("Detected TELEGRAM_CHAT_ID candidates:")
        for cid in sorted(chat_ids):
            print("-", cid)


if __name__ == "__main__":
    main()

