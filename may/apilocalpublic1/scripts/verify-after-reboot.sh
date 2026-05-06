#!/usr/bin/env bash
# Kiểm tra sau khi reboot: LaunchAgent API + Quick Tunnel (nếu đã install-background-all.sh).
set -euo pipefail

PORT="${PORT:-58471}"
DOMAIN="gui/$(id -u)"
LABEL_API="local.apilocalpublic1.api"
LABEL_TUN="local.apilocalpublic1.quicktunnel"
LOG_TUN="${HOME}/Library/Logs/apilocalpublic1-quicktunnel/stderr.log"

echo "=== 1) LaunchAgent đã load? ==="
launchctl print "${DOMAIN}/${LABEL_API}" >/dev/null 2>&1 && echo "  OK: ${LABEL_API}" || echo "  FAIL: ${LABEL_API} (chạy ./scripts/install-background-all.sh)"
launchctl print "${DOMAIN}/${LABEL_TUN}" >/dev/null 2>&1 && echo "  OK: ${LABEL_TUN}" || echo "  SKIP/FAIL: ${LABEL_TUN} (chưa cài quick tunnel nền)"

echo ""
echo "=== 2) API local (127.0.0.1:${PORT}) ==="
if curl -sS --connect-timeout 3 "http://127.0.0.1:${PORT}/" | grep -q "Hello World"; then
	echo "  OK: Hello World"
else
	echo "  FAIL: không nhận Hello World"
	exit 1
fi

echo ""
echo "=== 3) GET /health ==="
code="$(curl -sS -o /dev/null -w "%{http_code}" --connect-timeout 3 "http://127.0.0.1:${PORT}/health")"
[[ "${code}" == "200" ]] && echo "  OK: HTTP ${code}" || { echo "  FAIL: HTTP ${code}"; exit 1; }

echo ""
echo "=== 4) URL trycloudflare (từ log) ==="
if [[ -f "${LOG_TUN}" ]]; then
	url="$(grep -oE 'https://[a-z0-9-]+\.trycloudflare\.com' "${LOG_TUN}" 2>/dev/null | tail -1 || true)"
	if [[ -n "${url}" ]]; then
		echo "  Dòng gần nhất: ${url}"
		if curl -sS --connect-timeout 15 "${url}/" | grep -q "Hello World"; then
			echo "  OK: public curl -> Hello World"
		else
			echo "  WARN: public curl không ra Hello World (tunnel có thể vừa restart, đợi vài giây rồi chạy lại script)"
		fi
	else
		echo "  WARN: chưa thấy URL trong ${LOG_TUN}"
	fi
else
	echo "  SKIP: không có ${LOG_TUN}"
fi

echo ""
echo "Xong."
