#!/usr/bin/env python3
from __future__ import annotations

import html
import json
import re
import subprocess
from dataclasses import dataclass
from datetime import datetime
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import urlparse


def run_cmd(cmd: list[str]) -> str:
    return subprocess.check_output(cmd, text=True, stderr=subprocess.STDOUT).strip()


@dataclass(frozen=True)
class ApiError(Exception):
    status: int
    message: str


_NAME_RE = re.compile(
    # Examples:
    #   TCP 127.0.0.1:3010 (LISTEN)
    #   TCP *:5000 (LISTEN)
    r"^TCP\s+(?P<host>\S+):(?P<port>\d+)\s+\(LISTEN\)$"
)


def list_listening_ports() -> list[dict]:
    """
    Returns rows for TCP LISTEN sockets on macOS using lsof.
    """
    out = run_cmd(["lsof", "-nP", "-iTCP", "-sTCP:LISTEN"])
    lines = [ln.rstrip("\n") for ln in out.splitlines() if ln.strip()]
    if not lines:
        return []
    # First line is header:
    # COMMAND PID USER FD TYPE DEVICE SIZE/OFF NODE NAME
    rows: list[dict] = []
    for ln in lines[1:]:
        parts = ln.split()
        if len(parts) < 9:
            continue
        command = parts[0]
        try:
            pid = int(parts[1])
        except Exception:
            continue
        user = parts[2]
        name = " ".join(parts[8:])
        m = _NAME_RE.match(name)
        host = None
        port = None
        if m:
            host = m.group("host")
            try:
                port = int(m.group("port"))
            except Exception:
                port = None
        rows.append(
            {
                "command": command,
                "pid": pid,
                "user": user,
                "host": host,
                "port": port,
                "name": name,
            }
        )
    rows.sort(key=lambda r: (r["port"] if isinstance(r.get("port"), int) else 10**9, r["pid"]))
    return rows


def kill_pid(pid: int) -> None:
    if pid <= 1:
        raise ApiError(status=400, message="Refusing to kill pid <= 1")
    subprocess.check_call(["kill", str(pid)], stdout=subprocess.DEVNULL, stderr=subprocess.STDOUT)


HTML = """<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Port control (localhost)</title>
    <style>
      :root {
        --bg: #0b0f14;
        --panel: rgba(17,24,36,0.78);
        --text: #e8eef9;
        --muted: #9fb0c7;
        --border: rgba(255,255,255,0.10);
        --accent: #7aa2ff;
        --danger: #ff6b6b;
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
      .wrap { max-width: 1100px; margin: 28px auto; padding: 0 18px 36px; }
      header { display:flex; gap:12px; align-items: baseline; justify-content: space-between; flex-wrap: wrap; margin-bottom: 14px; }
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
      .toolbar { display:flex; gap:10px; align-items:center; justify-content: space-between; flex-wrap: wrap; margin-bottom: 10px; }
      .pill {
        font-family: var(--mono);
        font-size: 12px;
        padding: 2px 8px;
        border-radius: 999px;
        border: 1px solid rgba(255,255,255,0.15);
        color: var(--muted);
      }
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
      button.danger { border-color: rgba(255,107,107,0.35); background: rgba(255,107,107,0.08); }
      button.danger:hover { background: rgba(255,107,107,0.12); }
      table { width: 100%; border-collapse: collapse; overflow: hidden; border-radius: 12px; border: 1px solid var(--border); }
      thead th {
        text-align: left;
        font-size: 12px;
        color: var(--muted);
        font-weight: 700;
        padding: 10px;
        border-bottom: 1px solid var(--border);
        background: rgba(255,255,255,0.03);
      }
      tbody td { padding: 10px; border-bottom: 1px solid rgba(255,255,255,0.06); font-size: 13px; }
      tbody tr:hover { background: rgba(255,255,255,0.03); }
      td.mono { font-family: var(--mono); font-size: 12px; color: #dbe7ff; }
      .err { margin-top: 10px; color: var(--danger); font-size: 12px; font-family: var(--mono); white-space: pre-wrap; }
      .muted { color: var(--muted); }
      .right { text-align: right; }
    </style>
  </head>
  <body>
    <div class="wrap">
      <header>
        <h1>Port control</h1>
        <div class="meta">Served from <span class="pill">localhost:3011</span> · <span id="updated">—</span></div>
      </header>

      <section class="panel">
        <div class="toolbar">
          <div class="muted">Shows TCP ports in <span class="pill">LISTEN</span> state (via <span class="pill">lsof</span>).</div>
          <div style="display:flex; gap:10px; align-items:center;">
            <button id="refreshBtn" type="button">Refresh</button>
            <span class="pill" id="statusPill">idle</span>
          </div>
        </div>

        <div style="overflow:auto;">
          <table>
            <thead>
              <tr>
                <th style="width:90px;">Port</th>
                <th style="width:140px;">Host</th>
                <th style="width:90px;">PID</th>
                <th style="width:140px;">User</th>
                <th style="width:180px;">Command</th>
                <th>Name</th>
                <th class="right" style="width:120px;">Action</th>
              </tr>
            </thead>
            <tbody id="tbody"></tbody>
          </table>
        </div>

        <div class="err" id="err" style="display:none;"></div>
      </section>
    </div>

    <script>
      const tbody = document.getElementById('tbody');
      const err = document.getElementById('err');
      const statusPill = document.getElementById('statusPill');
      const updated = document.getElementById('updated');
      const refreshBtn = document.getElementById('refreshBtn');

      function setError(msg) {
        if (!msg) { err.style.display = 'none'; err.textContent = ''; return; }
        err.style.display = 'block'; err.textContent = msg;
      }
      function setStatus(s) { statusPill.textContent = s; }

      function esc(s) {
        return String(s ?? '').replace(/[&<>"']/g, (c) => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
      }

      function rowHtml(r) {
        const port = (typeof r.port === 'number') ? r.port : '';
        const host = r.host || '';
        const pid = r.pid || '';
        const user = r.user || '';
        const cmd = r.command || '';
        const name = r.name || '';
        return `
          <tr>
            <td class="mono">${esc(port)}</td>
            <td class="mono">${esc(host)}</td>
            <td class="mono">${esc(pid)}</td>
            <td>${esc(user)}</td>
            <td>${esc(cmd)}</td>
            <td class="mono">${esc(name)}</td>
            <td class="right">
              <button class="danger" data-pid="${esc(pid)}" type="button">Kill</button>
            </td>
          </tr>
        `;
      }

      function render(rows) {
        tbody.innerHTML = rows.map(rowHtml).join('') || `
          <tr><td colspan="7" class="muted">No listening ports found.</td></tr>
        `;
        for (const btn of tbody.querySelectorAll('button[data-pid]')) {
          btn.addEventListener('click', async () => {
            const pid = Number(btn.getAttribute('data-pid'));
            if (!pid) return;
            if (!confirm('Kill process PID ' + pid + '?')) return;
            try {
              setStatus('killing');
              setError('');
              const r = await fetch('/api/kill', {
                method: 'POST',
                headers: {'Content-Type':'application/json'},
                body: JSON.stringify({pid}),
              });
              const j = await r.json().catch(() => ({}));
              if (!r.ok) throw new Error(j && j.error ? j.error : ('HTTP ' + r.status));
              await refresh();
            } catch (e) {
              setStatus('error');
              setError(String(e && e.message ? e.message : e));
            } finally {
              if (statusPill.textContent !== 'error') setStatus('idle');
            }
          });
        }
      }

      async function refresh() {
        setStatus('loading');
        setError('');
        const r = await fetch('/api/listen', {cache: 'no-store'});
        const j = await r.json().catch(() => ({}));
        if (!r.ok) throw new Error(j && j.error ? j.error : ('HTTP ' + r.status));
        updated.textContent = j.updated || '—';
        render(j.rows || []);
        setStatus('idle');
      }

      refreshBtn.addEventListener('click', () => refresh().catch(e => { setStatus('error'); setError(String(e.message || e)); }));
      refresh().catch(e => { setStatus('error'); setError(String(e.message || e)); });
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

        if parsed.path == "/api/listen":
            try:
                rows = list_listening_ports()
                self._send_json(
                    200,
                    {"updated": datetime.now().isoformat(timespec="seconds"), "rows": rows},
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
        if parsed.path != "/api/kill":
            self._send(404, "text/plain; charset=utf-8", b"Not found")
            return

        try:
            data = self._read_json()
            if "pid" not in data:
                raise ApiError(status=400, message="Missing field: pid")
            try:
                pid = int(data["pid"])
            except Exception as e:
                raise ApiError(status=400, message="Field pid must be an integer") from e

            kill_pid(pid)
            self._send_json(200, {"ok": True})
        except ApiError as e:
            self._send_json(e.status, {"error": e.message})
        except subprocess.CalledProcessError as e:
            self._send_json(500, {"error": getattr(e, "output", str(e))})

    def log_message(self, fmt: str, *args) -> None:
        return


def main() -> None:
    host = "127.0.0.1"
    port = 3011
    httpd = ThreadingHTTPServer((host, port), Handler)
    print(f"Listening on http://{host}:{port}")
    httpd.serve_forever()


if __name__ == "__main__":
    main()

