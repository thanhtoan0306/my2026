#!/usr/bin/env bash
# Quick Tunnel (*.trycloudflare.com) chạy nền — giống ./tunnel.sh nhưng qua launchd.
# URL public đổi khi service khởi động lại; xem log stderr để copy hostname.
# Gỡ named tunnel LaunchAgent trước nếu đã cài: ./scripts/uninstall-cloudflare-tunnel-launchagent.sh
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LISTEN_PORT="${PORT:-58471}"

command -v cloudflared >/dev/null || {
	echo >&2 "Cài: brew install cloudflared"
	exit 1
}

CLOUDFLARED_BIN="$(command -v cloudflared)"
LAUNCH_AGENTS="${HOME}/Library/LaunchAgents"
PLIST_NAME="local.apilocalpublic1.quicktunnel.plist"
PLIST_DST="${LAUNCH_AGENTS}/${PLIST_NAME}"
LOG_DIR="${HOME}/Library/Logs/apilocalpublic1-quicktunnel"
LABEL="local.apilocalpublic1.quicktunnel"
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
		<string>--url</string>
		<string>http://127.0.0.1:${LISTEN_PORT}</string>
	</array>
	<key>WorkingDirectory</key>
	<string>${ROOT}</string>
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
echo "Đã cài Quick Tunnel (nền): ${PLIST_DST}"
echo "Log: ${LOG_DIR}/stderr.log (tìm dòng https://....trycloudflare.com)"
echo "Ví dụ: grep trycloudflare \"${LOG_DIR}/stderr.log\" | tail -3"
