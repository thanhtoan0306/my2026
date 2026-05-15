#!/usr/bin/env python3
"""Serve BTC price dashboard and /api/ticker JSON."""

import json
from http.server import ThreadingHTTPServer, SimpleHTTPRequestHandler
from pathlib import Path
from urllib.parse import urlparse

from btc_price import fetch_ticker

STATIC = Path(__file__).resolve().parent / "static"
PORT = 8080


class Handler(SimpleHTTPRequestHandler):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, directory=str(STATIC), **kwargs)

    def do_GET(self) -> None:
        path = urlparse(self.path).path

        if path == "/api/ticker":
            self._json_ticker()
            return

        if path in ("/", ""):
            self.path = "/index.html"

        super().do_GET()

    def _json_ticker(self) -> None:
        try:
            body = json.dumps(fetch_ticker()).encode()
            status = 200
        except RuntimeError as e:
            body = json.dumps({"error": str(e)}).encode()
            status = 502

        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, fmt: str, *args) -> None:
        if urlparse(self.path).path != "/api/ticker":
            super().log_message(fmt, *args)


def main() -> None:
    server = ThreadingHTTPServer(("", PORT), Handler)
    print(f"BTC dashboard: http://127.0.0.1:{PORT}/")
    print("Press Ctrl+C to stop.")
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\nStopped.")
        server.server_close()


if __name__ == "__main__":
    main()
