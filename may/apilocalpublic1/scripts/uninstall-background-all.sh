#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

"${ROOT}/scripts/uninstall-launchagent.sh"
"${ROOT}/scripts/uninstall-quick-tunnel-launchagent.sh"
echo "Đã gỡ API + Quick Tunnel (nền)."
