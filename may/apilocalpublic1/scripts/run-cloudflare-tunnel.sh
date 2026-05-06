#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CFG="${CLOUDFLARED_CONFIG:-${ROOT}/cloudflared/config.yml}"

[[ -f "${CFG}" ]] || {
	echo >&2 "Chưa có ${CFG}. Chạy trước:"
	echo >&2 "  CLOUDFLARE_HOSTNAME=api.example.com ./scripts/cloudflare-tunnel-bootstrap.sh"
	exit 1
}

exec cloudflared tunnel --config "${CFG}" run
