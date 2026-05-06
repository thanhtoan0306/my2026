#!/usr/bin/env bash
# Cài API chạy nền qua launchd (LaunchAgent). Đăng nhập lại vẫn tự chạy.
# Yêu cầu: Go để build. Mặc định PORT=58471 — đổi trước khi chạy: export PORT=...
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BINARY="${ROOT}/apilocalpublic1"
LAUNCH_AGENTS="${HOME}/Library/LaunchAgents"
PLIST_NAME="local.apilocalpublic1.api.plist"
PLIST_DST="${LAUNCH_AGENTS}/${PLIST_NAME}"
LOG_DIR="${HOME}/Library/Logs/apilocalpublic1"
LABEL="local.apilocalpublic1.api"
DOMAIN="gui/$(id -u)"
LISTEN_PORT="${PORT:-58471}"

mkdir -p "${LOG_DIR}" "${LAUNCH_AGENTS}"
(cd "${ROOT}" && go build -o "${BINARY}" .)

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
		<string>${BINARY}</string>
	</array>
	<key>WorkingDirectory</key>
	<string>${ROOT}</string>
	<key>EnvironmentVariables</key>
	<dict>
		<key>PORT</key>
		<string>${LISTEN_PORT}</string>
	</dict>
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
echo "Đã cài LaunchAgent: ${PLIST_DST}"
echo "Log: ${LOG_DIR}/stdout.log | ${LOG_DIR}/stderr.log"
echo "Kiểm tra: curl -s http://127.0.0.1:${LISTEN_PORT}/"
