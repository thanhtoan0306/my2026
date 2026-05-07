#!/usr/bin/env python3
from __future__ import annotations

import html
import json
import subprocess
from datetime import datetime
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import parse_qs, urlparse


def run_cmd(cmd: list[str]) -> str:
    return subprocess.check_output(cmd, text=True, stderr=subprocess.STDOUT).strip()


def get_running_gui_apps() -> list[str]:
    """
    Returns names of foreground (GUI) apps using macOS System Events.
    """
    script = r"""
set text item delimiters to "\n"
tell application "System Events"
  set appNames to name of (processes where background only is false)
end tell
return appNames as text
""".strip()
    out = run_cmd(["osascript", "-e", script])
    if not out:
        return []
    apps = [a.strip() for a in out.splitlines() if a.strip()]
    # Make stable, deduped
    return sorted(dict.fromkeys(apps), key=str.casefold)


def get_process_snapshot(limit: int = 300) -> list[dict]:
    """
    Returns a lightweight process snapshot.
    """
    ps = run_cmd(["ps", "-axo", "pid=,ppid=,user=,comm="])
    rows: list[dict] = []
    for line in ps.splitlines():
        line = line.strip()
        if not line:
            continue
        parts = line.split(None, 3)
        if len(parts) < 4:
            continue
        pid_s, ppid_s, user, comm = parts
        try:
            pid = int(pid_s)
            ppid = int(ppid_s)
        except ValueError:
            continue
        rows.append({"pid": pid, "ppid": ppid, "user": user, "comm": comm})
        if len(rows) >= max(1, limit):
            break
    return rows


HTML_TEMPLATE = """<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Running apps (Mac)</title>
    <style>
      :root {{
        --bg: #0b0f14;
        --panel: #111824;
        --text: #e8eef9;
        --muted: #9fb0c7;
        --border: rgba(255,255,255,0.10);
        --accent: #7aa2ff;
        --danger: #ff6b6b;
        --mono: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
        --sans: ui-sans-serif, system-ui, -apple-system, Segoe UI, Roboto, Helvetica, Arial, "Apple Color Emoji", "Segoe UI Emoji";
      }}
      body {{
        margin: 0;
        background: radial-gradient(1200px 600px at 15% 10%, rgba(122,162,255,0.18), transparent 60%),
                    radial-gradient(800px 500px at 90% 20%, rgba(255,107,107,0.10), transparent 55%),
                    var(--bg);
        color: var(--text);
        font-family: var(--sans);
      }}
      .wrap {{
        max-width: 1100px;
        margin: 28px auto;
        padding: 0 18px 36px;
      }}
      header {{
        display: flex;
        gap: 12px;
        align-items: baseline;
        justify-content: space-between;
        flex-wrap: wrap;
        margin-bottom: 14px;
      }}
      h1 {{
        font-size: 20px;
        margin: 0;
        letter-spacing: 0.2px;
      }}
      .meta {{
        color: var(--muted);
        font-size: 12px;
      }}
      .panel {{
        background: rgba(17,24,36,0.78);
        border: 1px solid var(--border);
        border-radius: 14px;
        padding: 14px;
        box-shadow: 0 10px 30px rgba(0,0,0,0.25);
        backdrop-filter: blur(8px);
      }}
      .row {{
        display: grid;
        grid-template-columns: 1fr;
        gap: 14px;
      }}
      @media (min-width: 900px) {{
        .row {{
          grid-template-columns: 1fr 1fr;
        }}
      }}
      .toolbar {{
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 10px;
        margin-bottom: 10px;
        flex-wrap: wrap;
      }}
      .btn {{
        display: inline-flex;
        align-items: center;
        gap: 8px;
        padding: 8px 10px;
        border-radius: 10px;
        background: rgba(122,162,255,0.12);
        border: 1px solid rgba(122,162,255,0.25);
        color: var(--text);
        text-decoration: none;
        font-size: 13px;
      }}
      .btn:hover {{
        background: rgba(122,162,255,0.18);
      }}
      .btn.secondary {{
        background: rgba(255,255,255,0.06);
        border: 1px solid rgba(255,255,255,0.12);
      }}
      .title {{
        font-size: 14px;
        font-weight: 650;
        margin: 0;
      }}
      .sub {{
        font-size: 12px;
        color: var(--muted);
        margin-top: 4px;
      }}
      ul.apps {{
        list-style: none;
        padding: 0;
        margin: 10px 0 0;
        display: grid;
        grid-template-columns: 1fr;
        gap: 8px;
      }}
      @media (min-width: 520px) {{
        ul.apps {{
          grid-template-columns: 1fr 1fr;
        }}
      }}
      li.app {{
        padding: 10px 10px;
        border: 1px solid var(--border);
        background: rgba(255,255,255,0.04);
        border-radius: 12px;
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 10px;
      }}
      .appname {{
        font-size: 13px;
        font-weight: 600;
      }}
      .pill {{
        font-family: var(--mono);
        font-size: 12px;
        padding: 2px 8px;
        border-radius: 999px;
        border: 1px solid rgba(255,255,255,0.15);
        color: var(--muted);
      }}
      pre {{
        margin: 10px 0 0;
        padding: 10px;
        border-radius: 12px;
        background: rgba(0,0,0,0.25);
        border: 1px solid var(--border);
        overflow: auto;
        max-height: 520px;
        font-family: var(--mono);
        font-size: 12px;
        line-height: 1.35;
        color: #dbe7ff;
      }}
      .error {{
        color: var(--danger);
        font-size: 13px;
        white-space: pre-wrap;
        margin-top: 10px;
      }}
      footer {{
        margin-top: 14px;
        color: var(--muted);
        font-size: 12px;
      }}
      code.inline {{
        font-family: var(--mono);
        background: rgba(255,255,255,0.06);
        border: 1px solid rgba(255,255,255,0.12);
        padding: 1px 6px;
        border-radius: 8px;
        color: #dbe7ff;
      }}
    </style>
  </head>
  <body>
    <div class="wrap">
      <header>
        <h1>Running apps on this Mac</h1>
        <div class="meta">Served from <span class="pill">localhost:9123</span> · Updated {updated}</div>
      </header>

      <div class="row">
        <section class="panel">
          <div class="toolbar">
            <div>
              <p class="title">GUI apps</p>
              <div class="sub">From System Events (foreground apps)</div>
            </div>
            <a class="btn" href="/">Refresh</a>
          </div>

          {apps_section}
          {apps_error}
        </section>

        <section class="panel">
          <div class="toolbar">
            <div>
              <p class="title">Processes (optional)</p>
              <div class="sub">Raw process snapshot via <code class="inline">ps</code></div>
            </div>
            <a class="btn secondary" href="/?procs=1">Show</a>
          </div>

          {procs_section}
          {procs_error}
        </section>
      </div>

      <footer>
        Tip: open <code class="inline">/api/apps</code> for JSON apps, or <code class="inline">/api/procs</code> for JSON processes.
      </footer>
    </div>
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

    def do_GET(self) -> None:  # noqa: N802
        parsed = urlparse(self.path)
        qs = parse_qs(parsed.query or "")

        if parsed.path == "/api/apps":
            try:
                apps = get_running_gui_apps()
                body = json.dumps(
                    {"updated": datetime.now().isoformat(timespec="seconds"), "apps": apps},
                    ensure_ascii=False,
                    indent=2,
                ).encode("utf-8")
                self._send(200, "application/json; charset=utf-8", body)
            except subprocess.CalledProcessError as e:
                body = json.dumps(
                    {"error": str(e), "output": getattr(e, "output", "")},
                    ensure_ascii=False,
                    indent=2,
                ).encode("utf-8")
                self._send(500, "application/json; charset=utf-8", body)
            return

        if parsed.path == "/api/procs":
            limit = 300
            if "limit" in qs:
                try:
                    limit = max(1, min(2000, int(qs["limit"][0])))
                except Exception:
                    limit = 300
            try:
                procs = get_process_snapshot(limit=limit)
                body = json.dumps(
                    {"updated": datetime.now().isoformat(timespec="seconds"), "processes": procs},
                    ensure_ascii=False,
                    indent=2,
                ).encode("utf-8")
                self._send(200, "application/json; charset=utf-8", body)
            except subprocess.CalledProcessError as e:
                body = json.dumps(
                    {"error": str(e), "output": getattr(e, "output", "")},
                    ensure_ascii=False,
                    indent=2,
                ).encode("utf-8")
                self._send(500, "application/json; charset=utf-8", body)
            return

        if parsed.path != "/":
            self._send(404, "text/plain; charset=utf-8", b"Not found")
            return

        updated = datetime.now().strftime("%Y-%m-%d %H:%M:%S")

        apps_section = ""
        apps_error = ""
        try:
            apps = get_running_gui_apps()
            items = "\n".join(
                f'<li class="app"><span class="appname">{html.escape(name)}</span>'
                f'<span class="pill">GUI</span></li>'
                for name in apps
            )
            apps_section = f'<ul class="apps">{items}</ul>' if items else "<div class='sub'>No apps found.</div>"
        except subprocess.CalledProcessError as e:
            apps_error = f"<div class='error'>{html.escape(getattr(e, 'output', str(e)))}</div>"

        procs_section = "<div class='sub'>Click “Show” to load process snapshot.</div>"
        procs_error = ""
        if qs.get("procs", ["0"])[0] == "1":
            limit = 200
            if "limit" in qs:
                try:
                    limit = max(1, min(2000, int(qs["limit"][0])))
                except Exception:
                    limit = 200
            try:
                procs = get_process_snapshot(limit=limit)
                procs_section = "<pre>" + html.escape(json.dumps(procs, ensure_ascii=False, indent=2)) + "</pre>"
            except subprocess.CalledProcessError as e:
                procs_error = f"<div class='error'>{html.escape(getattr(e, 'output', str(e)))}</div>"

        page = HTML_TEMPLATE.format(
            updated=html.escape(updated),
            apps_section=apps_section,
            apps_error=apps_error,
            procs_section=procs_section,
            procs_error=procs_error,
        ).encode("utf-8")
        self._send(200, "text/html; charset=utf-8", page)

    def log_message(self, fmt: str, *args) -> None:
        # Keep terminal output clean.
        return


def main() -> None:
    host = "127.0.0.1"
    port = 9123
    httpd = ThreadingHTTPServer((host, port), Handler)
    print(f"Listening on http://{host}:{port}")
    httpd.serve_forever()


if __name__ == "__main__":
    main()

