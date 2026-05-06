#!/usr/bin/env bash
# Một lần: đăng nhập Cloudflare, tạo tunnel có tên, route DNS, ghi cloudflared/config.yml
# Yêu cầu: brew install cloudflared | Miền (hostname) phải thuộc zone trong Cloudflare của bạn
#
#   cloudflared tunnel login
#   CLOUDFLARE_HOSTNAME=api.example.com ./scripts/cloudflare-tunnel-bootstrap.sh
#
# Tuỳ chọn: CLOUDFLARE_TUNNEL_NAME=mytunnel PORT=58471
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CFG_DIR="${ROOT}/cloudflared"
CFG="${CFG_DIR}/config.yml"
TUNNEL_NAME="${CLOUDFLARE_TUNNEL_NAME:-apilocalpublic1}"
HOSTNAME="${CLOUDFLARE_HOSTNAME:?Đặt hostname đầy đủ, ví dụ: api.example.com}"
PORT="${PORT:-58471}"

command -v cloudflared >/dev/null || {
	echo >&2 "Cài cloudflared: brew install cloudflared"
	exit 1
}

CERT="${HOME}/.cloudflared/cert.pem"
[[ -f "${CERT}" ]] || {
	echo >&2 "Chạy một lần trước: cloudflared tunnel login"
	exit 1
}

tunnel_id_for_name() {
	cloudflared tunnel list -o json 2>/dev/null | python3 -c '
import json, sys
name = sys.argv[1]
raw = json.load(sys.stdin)

def rows(r):
    if isinstance(r, list):
        return r
    if isinstance(r, dict):
        for k in ("tunnels", "result", "data"):
            v = r.get(k)
            if isinstance(v, list):
                return v
    return []

for t in rows(raw):
    if not isinstance(t, dict) or t.get("name") != name:
        continue
    for k in ("id", "uuid", "tunnel_id"):
        v = t.get(k)
        if v:
            print(v)
            raise SystemExit(0)
raise SystemExit(1)
' "${1}" || return 1
}

UUID=""
UUID="$(tunnel_id_for_name "${TUNNEL_NAME}" 2>/dev/null)" || true

if [[ -z "${UUID}" ]]; then
	echo "Tạo tunnel '${TUNNEL_NAME}'..."
	CREATE_OUT="$(cloudflared tunnel create "${TUNNEL_NAME}" 2>&1)"
	echo "${CREATE_OUT}"
	UUID="$(tunnel_id_for_name "${TUNNEL_NAME}" 2>/dev/null)" || true
	[[ -n "${UUID}" ]] || UUID="$(echo "${CREATE_OUT}" | grep -oE '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}' | tail -1 || true)"
fi

if [[ -z "${UUID}" ]]; then
	echo >&2 "Không lấy được tunnel id. Chạy: cloudflared tunnel list"
	exit 1
fi

CREDS="${HOME}/.cloudflared/${UUID}.json"
[[ -f "${CREDS}" ]] || {
	echo >&2 "Không thấy file credentials: ${CREDS}"
	exit 1
}

echo "Route DNS: ${HOSTNAME} -> tunnel ${TUNNEL_NAME} ..."
cloudflared tunnel route dns "${TUNNEL_NAME}" "${HOSTNAME}" || {
	echo >&2 "(Gợi ý: bản ghi DNS có thể đã tồn tại — kiểm tra trên dashboard Cloudflare DNS.)"
}

mkdir -p "${CFG_DIR}"
cat >"${CFG}" <<EOF
tunnel: ${UUID}
credentials-file: ${CREDS}

ingress:
  - hostname: ${HOSTNAME}
    service: http://127.0.0.1:${PORT}
  - service: http_status:404
EOF

echo ""
echo "Đã ghi ${CFG}"
echo "Chạy tunnel (giữ API đang listen cổng ${PORT}):"
echo "  ./scripts/run-cloudflare-tunnel.sh"
echo "Hoặc chạy ngầm:"
echo "  ./scripts/install-cloudflare-tunnel-launchagent.sh"
