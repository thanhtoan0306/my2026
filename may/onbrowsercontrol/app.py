#!/usr/bin/env python3
from __future__ import annotations

import json
import subprocess
from dataclasses import dataclass
from datetime import datetime
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import urlparse


def run_cmd(cmd: list[str]) -> str:
    return subprocess.check_output(cmd, text=True, stderr=subprocess.STDOUT).strip()


def get_volume() -> int:
    # macOS output volume is 0..100
    out = run_cmd(["osascript", "-e", "output volume of (get volume settings)"])
    try:
        v = int(out.strip())
    except Exception:
        v = 0
    return max(0, min(100, v))


def set_volume(volume: int) -> int:
    v = max(0, min(100, int(volume)))
    run_cmd(["osascript", "-e", f"set volume output volume {v}"])
    return get_volume()


@dataclass(frozen=True)
class ApiError(Exception):
    status: int
    message: str


HTML = """<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>PC Volume (localhost)</title>
    <style>
      :root {
        --bg: #0b0f14;
        --panel: rgba(17,24,36,0.78);
        --text: #e8eef9;
        --muted: #9fb0c7;
        --border: rgba(255,255,255,0.10);
        --accent: #7aa2ff;
        --mono: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
        --sans: ui-sans-serif, system-ui, -apple-system, Segoe UI, Roboto, Helvetica, Arial;
      }
      body {
        margin: 0;
        background: radial-gradient(1200px 600px at 15% 10%, rgba(122,162,255,0.18), transparent 60%),
                    radial-gradient(800px 500px at 90% 20%, rgba(255,107,107,0.10), transparent 55%),
                    var(--bg);
        color: var(--text);
        font-family: var(--sans);
      }
      .wrap {
        max-width: 860px;
        margin: 28px auto;
        padding: 0 18px 36px;
      }
      header {
        display: flex;
        gap: 12px;
        align-items: baseline;
        justify-content: space-between;
        flex-wrap: wrap;
        margin-bottom: 14px;
      }
      h1 { font-size: 18px; margin: 0; letter-spacing: 0.2px; }
      .meta { color: var(--muted); font-size: 12px; }
      .panel {
        background: var(--panel);
        border: 1px solid var(--border);
        border-radius: 14px;
        padding: 14px;
        box-shadow: 0 10px 30px rgba(0,0,0,0.25);
        backdrop-filter: blur(8px);
      }
      .row { display: grid; grid-template-columns: 1fr; gap: 12px; }
      .kpis { display: flex; gap: 10px; flex-wrap: wrap; align-items: center; justify-content: space-between; }
      .pill {
        font-family: var(--mono);
        font-size: 12px;
        padding: 2px 8px;
        border-radius: 999px;
        border: 1px solid rgba(255,255,255,0.15);
        color: var(--muted);
      }
      .big {
        font-family: var(--mono);
        font-size: 24px;
        font-weight: 800;
        letter-spacing: 0.4px;
      }
      input[type="range"] {
        width: 100%;
        accent-color: var(--accent);
      }
      .hint { margin-top: 8px; color: var(--muted); font-size: 12px; line-height: 1.35; }
      .err { margin-top: 10px; color: #ff6b6b; font-size: 12px; font-family: var(--mono); white-space: pre-wrap; }
      button {
        border: 1px solid rgba(255,255,255,0.14);
        background: rgba(255,255,255,0.06);
        color: var(--text);
        padding: 8px 10px;
        border-radius: 10px;
        font-size: 13px;
        cursor: pointer;
      }
      button:hover { background: rgba(255,255,255,0.09); }
    </style>
  </head>
  <body>
    <div class="wrap">
      <header>
        <h1>Volume control</h1>
        <div class="meta">Served from <span class="pill">localhost:3010</span> · <span id="updated">—</span></div>
      </header>

      <section class="panel">
        <div class="row">
          <div class="kpis">
            <div>
              <div class="pill">System output volume</div>
              <div class="big"><span id="volText">—</span><span style="color: var(--muted)">%</span></div>
            </div>
            <div style="display:flex; gap:10px; align-items:center;">
              <button id="refreshBtn" type="button">Refresh</button>
              <span class="pill" id="statusPill">idle</span>
            </div>
          </div>

          <input id="slider" type="range" min="0" max="100" step="1" value="0" />
          <div class="hint">
            Drag the slider to change system volume (0–100). This page talks to your Mac via <span class="pill">/api/volume</span>.
          </div>
          <div class="err" id="err" style="display:none;"></div>
        </div>
      </section>
    </div>

    <script>
      const slider = document.getElementById('slider');
      const volText = document.getElementById('volText');
      const updated = document.getElementById('updated');
      const err = document.getElementById('err');
      const statusPill = document.getElementById('statusPill');
      const refreshBtn = document.getElementById('refreshBtn');

      function setError(msg) {
        if (!msg) { err.style.display = 'none'; err.textContent = ''; return; }
        err.style.display = 'block'; err.textContent = msg;
      }

      function setStatus(s) { statusPill.textContent = s; }

      function applyState(state) {
        if (!state) return;
        if (typeof state.volume === 'number') {
          const v = Math.max(0, Math.min(100, Math.round(state.volume)));
          slider.value = String(v);
          volText.textContent = String(v);
        }
        if (state.updated) updated.textContent = state.updated;
      }

      async function apiGet() {
        setStatus('loading');
        setError('');
        const r = await fetch('/api/volume', {cache: 'no-store'});
        const j = await r.json().catch(() => ({}));
        if (!r.ok) throw new Error(j && j.error ? j.error : ('HTTP ' + r.status));
        applyState(j);
        setStatus('idle');
      }

      let lastSent = null;
      let timer = null;

      async function apiSet(vol) {
        setStatus('setting');
        setError('');
        const r = await fetch('/api/volume', {
          method: 'POST',
          headers: {'Content-Type': 'application/json'},
          body: JSON.stringify({volume: vol}),
        });
        const j = await r.json().catch(() => ({}));
        if (!r.ok) throw new Error(j && j.error ? j.error : ('HTTP ' + r.status));
        applyState(j);
        setStatus('idle');
      }

      function scheduleSend(vol) {
        lastSent = vol;
        if (timer) clearTimeout(timer);
        timer = setTimeout(async () => {
          const v = lastSent;
          try { await apiSet(v); }
          catch (e) { setStatus('error'); setError(String(e && e.message ? e.message : e)); }
        }, 60);
      }

      slider.addEventListener('input', () => {
        const v = Number(slider.value);
        volText.textContent = String(v);
        scheduleSend(v);
      });
      refreshBtn.addEventListener('click', () => apiGet().catch(e => { setStatus('error'); setError(String(e.message || e)); }));

      apiGet().catch(e => { setStatus('error'); setError(String(e.message || e)); });
    </script>
  </body>
</html>
"""


class Handler(BaseHTTPRequestHandler):
    def _send(self, status: int, content_type: str, body: bytes) -> None:
        self.send_response(status)
        self.send_header("Content-Type", content_type)
        self.send_header("Content-Length", str(len(body)))
        self.send_header("Cache-Control", "no-store")
        self.end_headers()
        self.wfile.write(body)

    def _send_json(self, status: int, payload: dict) -> None:
        body = json.dumps(payload, ensure_ascii=False, indent=2).encode("utf-8")
        self._send(status, "application/json; charset=utf-8", body)

    def _read_json(self) -> dict:
        try:
            length = int(self.headers.get("Content-Length", "0") or "0")
        except Exception:
            length = 0
        raw = self.rfile.read(max(0, min(1_000_000, length))) if length else b""
        if not raw:
            return {}
        try:
            return json.loads(raw.decode("utf-8"))
        except Exception as e:
            raise ApiError(status=400, message=f"Invalid JSON: {e}") from e

    def do_GET(self) -> None:  # noqa: N802
        parsed = urlparse(self.path)

        if parsed.path == "/api/volume":
            try:
                v = get_volume()
                self._send_json(
                    200,
                    {"updated": datetime.now().isoformat(timespec="seconds"), "volume": v},
                )
            except subprocess.CalledProcessError as e:
                self._send_json(500, {"error": getattr(e, "output", str(e))})
            return

        if parsed.path != "/":
            self._send(404, "text/plain; charset=utf-8", b"Not found")
            return

        self._send(200, "text/html; charset=utf-8", HTML.encode("utf-8"))

    def do_POST(self) -> None:  # noqa: N802
        parsed = urlparse(self.path)
        if parsed.path != "/api/volume":
            self._send(404, "text/plain; charset=utf-8", b"Not found")
            return

        try:
            data = self._read_json()
            if "volume" not in data:
                raise ApiError(status=400, message="Missing field: volume")
            try:
                vol = int(data["volume"])
            except Exception as e:
                raise ApiError(status=400, message="Field volume must be an integer") from e

            v = set_volume(vol)
            self._send_json(
                200,
                {"updated": datetime.now().isoformat(timespec="seconds"), "volume": v},
            )
        except ApiError as e:
            self._send_json(e.status, {"error": e.message})
        except subprocess.CalledProcessError as e:
            self._send_json(500, {"error": getattr(e, "output", str(e))})

    def log_message(self, fmt: str, *args) -> None:
        return


def main() -> None:
    host = "127.0.0.1"
    port = 3010
    httpd = ThreadingHTTPServer((host, port), Handler)
    print(f"Listening on http://{host}:{port}")
    httpd.serve_forever()


if __name__ == "__main__":
    main()

