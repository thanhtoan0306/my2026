#!/usr/bin/env bash
# API Go + Quick Tunnel cùng chạy nền (launchd). Đăng nhập lại vẫn tự chạy.
# PORT áp dụng cho cả API và tunnel (mặc định 58471).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

"${ROOT}/scripts/install-launchagent.sh"
"${ROOT}/scripts/install-quick-tunnel-launchagent.sh"

LOG_Q="${HOME}/Library/Logs/apilocalpublic1-quicktunnel/stderr.log"
echo ""
echo "--- Xong. Local API: curl -s http://127.0.0.1:${PORT:-58471}/"
echo "--- URL public (đọc log sau vài giây): grep trycloudflare \"${LOG_Q}\" | tail -3"
