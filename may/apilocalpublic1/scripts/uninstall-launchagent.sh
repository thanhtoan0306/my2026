#!/usr/bin/env bash
# Gỡ LaunchAgent (dừng chạy ngầm sau đăng nhập).
set -euo pipefail

PLIST_NAME="local.apilocalpublic1.api.plist"
PLIST_DST="${HOME}/Library/LaunchAgents/${PLIST_NAME}"
LABEL="local.apilocalpublic1.api"
DOMAIN="gui/$(id -u)"

launchctl bootout "${DOMAIN}/${LABEL}" 2>/dev/null || true
rm -f "${PLIST_DST}"
echo "Đã gỡ ${PLIST_NAME}"
