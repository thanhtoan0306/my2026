#!/usr/bin/env bash
set -euo pipefail

PLIST_NAME="local.apilocalpublic1.cloudflared.plist"
PLIST_DST="${HOME}/Library/LaunchAgents/${PLIST_NAME}"
LABEL="local.apilocalpublic1.cloudflared"
DOMAIN="gui/$(id -u)"

launchctl bootout "${DOMAIN}/${LABEL}" 2>/dev/null || true
rm -f "${PLIST_DST}"
echo "Đã gỡ ${PLIST_NAME}"
