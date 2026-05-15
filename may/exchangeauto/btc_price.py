#!/usr/bin/env python3
"""Fetch BTC-USDT spot ticker from OKX (public API, no keys)."""

import sys
from typing import Any

import okx.MarketData as MarketData

INST_ID = "BTC-USDT"
FLAG = "0"  # 0 = live, 1 = demo


def fetch_ticker(inst_id: str = INST_ID) -> dict[str, Any]:
    api = MarketData.MarketAPI(flag=FLAG)
    result = api.get_ticker(instId=inst_id)

    if result.get("code") != "0":
        raise RuntimeError(result.get("msg") or str(result))

    t = result["data"][0]
    return {
        "instId": t["instId"],
        "last": float(t["last"]),
        "bid": float(t["bidPx"]),
        "ask": float(t["askPx"]),
        "high24h": float(t["high24h"]),
        "low24h": float(t["low24h"]),
        "ts": t["ts"],
    }


def main() -> None:
    try:
        data = fetch_ticker()
    except RuntimeError as e:
        print(f"Error: {e}", file=sys.stderr)
        raise SystemExit(1) from e

    print(f"BTC/USDT  last:  ${data['last']:,.2f}")
    print(f"          bid:   ${data['bid']:,.2f}   ask: ${data['ask']:,.2f}")
    print(
        f"          24h:   ${data['low24h']:,.2f} – ${data['high24h']:,.2f}"
    )


if __name__ == "__main__":
    main()
