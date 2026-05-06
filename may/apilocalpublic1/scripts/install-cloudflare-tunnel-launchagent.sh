#!/usr/bin/env bash
# Chạy cloudflared nền (LaunchAgent), dùng cloudflared/config.yml
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CFG="${ROOT}/cloudflared/config.yml"

[[ -f "${CFG}" ]] || {
	echo >&2 "Thiếu ${CFG}. Chạy: CLOUDFLARE_HOSTNAME=api.example.com ./scripts/cloudflare-tunnel-bootstrap.sh"
	exit 1
}

CLOUDFLARED_BIN="$(command -v cloudflared)"
LAUNCH_AGENTS="${HOME}/Library/LaunchAgents"
PLIST_NAME="local.apilocalpublic1.cloudflared.plist"
PLIST_DST="${LAUNCH_AGENTS}/${PLIST_NAME}"
LOG_DIR="${HOME}/Library/Logs/apilocalpublic1-cloudflared"
LABEL="local.apilocalpublic1.cloudflared"
DOMAIN="gui/$(id -u)"

mkdir -p "${LOG_DIR}" "${LAUNCH_AGENTS}"

launchctl bootout "${DOMAIN}/${LABEL}" 2>/dev/null || true

cat >"${PLIST_DST}" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>${LABEL}</string>
	<key>ProgramArguments</key>
	<array>
		<string>${CLOUDFLARED_BIN}</string>
		<string>tunnel</string>
		<string>--config</string>
		<string>${CFG}</string>
		<string>run</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
	<key>StandardOutPath</key>
	<string>${LOG_DIR}/stdout.log</string>
	<key>StandardErrorPath</key>
	<string>${LOG_DIR}/stderr.log</string>
</dict>
</plist>
EOF

launchctl bootstrap "${DOMAIN}" "${PLIST_DST}"
echo "Đã cài LaunchAgent tunnel: ${PLIST_DST}"
echo "Log: ${LOG_DIR}/stdout.log | ${LOG_DIR}/stderr.log"
