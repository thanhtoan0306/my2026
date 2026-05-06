#!/usr/bin/env bash
set -euo pipefail

# Thử nhanh (URL *.trycloudflare.com, không cần zone DNS riêng).
# Tunnel ổn định + tên miền của bạn: ./scripts/cloudflare-tunnel-bootstrap.sh rồi ./scripts/run-cloudflare-tunnel.sh
#
# Cần cloudflared (`brew install cloudflared`). API phải đang chạy (go run . hoặc LaunchAgent).
# Chạy nền (API + tunnel): ./scripts/install-background-all.sh

PORT="${PORT:-58471}"
exec cloudflared tunnel --url "http://127.0.0.1:${PORT}"
