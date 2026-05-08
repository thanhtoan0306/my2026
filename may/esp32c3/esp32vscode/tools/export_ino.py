#!/usr/bin/env python3
from __future__ import annotations

import re
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
SRC = ROOT / "src"
DIST = ROOT / "dist"


def _read(p: Path) -> str:
    return p.read_text(encoding="utf-8").replace("\r\n", "\n")


def _strip_pragma_once(text: str) -> str:
    # Arduino builder doesn't care, but this avoids duplicating it in the export.
    return re.sub(r"^\s*#pragma\s+once\s*\n", "", text, flags=re.MULTILINE)


def _strip_local_includes(text: str) -> str:
    # The exported artifact must be a single .ino file. Strip local includes that
    # would otherwise require extra files next to the sketch.
    text = re.sub(r'^\s*#include\s+"config\.h"\s*\n', "", text, flags=re.MULTILINE)
    text = re.sub(r'^\s*#include\s+"web_ui\.h"\s*\n', "", text, flags=re.MULTILINE)
    return text


def _escape_raw_string_delimiter(html: str) -> str:
    # Our chosen raw-string delimiter is HTML ... HTML. If the page contains )HTML",
    # it would terminate early. Rare, but make it safe.
    return html.replace(')HTML"', ')H_T_M_L"')


def _emit_embedded_html(page_html: str) -> str:
    page_html = _escape_raw_string_delimiter(page_html.rstrip("\n"))
    return (
        "static const char kPageHtml[] PROGMEM = R\"HTML(\n"
        + page_html
        + "\n)HTML\";\n"
    )


def main() -> int:
    DIST.mkdir(parents=True, exist_ok=True)

    parts: list[tuple[str, str]] = []

    config_h = SRC / "config.h"
    page_html = SRC / "page.html"
    sketch_ino = SRC / "sketch.ino"

    for p in (config_h, page_html, sketch_ino):
        if not p.exists():
            raise SystemExit(f"Missing required file: {p}")

    parts.append(("config.h", _strip_pragma_once(_read(config_h)).strip() + "\n"))
    parts.append(("page.html (embedded)", _emit_embedded_html(_read(page_html)).strip() + "\n"))
    parts.append(("sketch.ino", _strip_local_includes(_read(sketch_ino)).strip() + "\n"))

    out = []
    out.append("// AUTO-GENERATED FILE. DO NOT EDIT.\n")
    out.append(f"// Source: {ROOT.as_posix()}\n")
    out.append("// Re-generate by running: python3 tools/export_ino.py\n\n")

    for name, body in parts:
        out.append(f"\n// ===== BEGIN {name} =====\n")
        out.append(body)
        out.append(f"// ===== END {name} =====\n")

    dest = DIST / "slideLEDandOLED.ino"
    dest.write_text("".join(out), encoding="utf-8")
    print(f"Wrote {dest}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

