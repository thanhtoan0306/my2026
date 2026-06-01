#!/usr/bin/env python3
from __future__ import annotations

import json
import re
import subprocess
from dataclasses import dataclass

from flask import Flask, render_template, request

app = Flask(__name__)

_CURL_START = re.compile(r"^\s*curl\b", re.IGNORECASE)
_UNSAFE = re.compile(
    r"(?:^|[\s;])(?:rm\s|sudo\s|curl\s+.*\|\s*sh|>\s*/|/dev/tcp)",
    re.IGNORECASE,
)
_SHELL_METAS = re.compile(r"(?:&&|\|\||`|\$\(|\$\{|<\(|>\(|;\s*(?!\s*$))")


@dataclass(frozen=True)
class RunResult:
    ok: bool
    status: str
    body: str
    error: str | None = None


def validate_curl(cmd: str) -> str | None:
    text = cmd.strip()
    if not text:
        return "Paste a curl command first."
    if not _CURL_START.match(text):
        return "Command must start with curl."
    if _UNSAFE.search(text):
        return "Command blocked for safety."
    if _SHELL_METAS.search(text):
        return "Shell chaining (&&, ||, ;, $(), etc.) is not allowed."
    return None


def beautify_body(raw: str) -> str:
    stripped = raw.strip()
    if not stripped:
        return ""
    try:
        data = json.loads(stripped)
    except json.JSONDecodeError:
        return raw
    return json.dumps(data, ensure_ascii=False, indent=2, sort_keys=True) + "\n"


def run_curl(cmd: str, *, timeout_s: float = 30.0) -> RunResult:
    try:
        proc = subprocess.run(
            ["/bin/bash", "-lc", cmd],
            capture_output=True,
            text=True,
            timeout=timeout_s,
        )
    except subprocess.TimeoutExpired:
        return RunResult(ok=False, status="timeout", body="", error=f"Timed out after {int(timeout_s)}s")
    except OSError as e:
        return RunResult(ok=False, status="error", body="", error=str(e))

    stdout = proc.stdout or ""
    stderr = proc.stderr or ""
    combined = stdout if stdout.strip() else stderr
    pretty = beautify_body(combined)

    if proc.returncode != 0 and not pretty.strip():
        return RunResult(
            ok=False,
            status=f"exit {proc.returncode}",
            body=pretty,
            error=stderr.strip() or f"curl exited with code {proc.returncode}",
        )

    return RunResult(
        ok=proc.returncode == 0,
        status=f"exit {proc.returncode}",
        body=pretty,
        error=stderr.strip() if proc.returncode != 0 and stderr.strip() else None,
    )


@app.get("/")
def index():
    return render_template(
        "index.html",
        curl_text="",
        result=None,
        validation_error=None,
    )


@app.post("/")
def run():
    curl_text = request.form.get("curl", "") or ""
    validation_error = validate_curl(curl_text)
    if validation_error:
        return render_template(
            "index.html",
            curl_text=curl_text,
            result=None,
            validation_error=validation_error,
        )

    result = run_curl(curl_text.strip())
    return render_template(
        "index.html",
        curl_text=curl_text,
        result=result,
        validation_error=None,
    )


def main() -> None:
    host = "127.0.0.1"
    port = 3029
    print(f"Listening on http://{host}:{port}")
    app.run(host=host, port=port, debug=False)


if __name__ == "__main__":
    main()
