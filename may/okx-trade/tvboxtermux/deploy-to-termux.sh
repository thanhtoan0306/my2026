#!/bin/bash
# Run on your Mac (new terminal, not inside SSH). Copies app to Termux.
set -e
HOST="${TERMUX_HOST:-192.168.1.153}"
PORT="${TERMUX_PORT:-8022}"
USER="${TERMUX_USER:-u0_a66}"
SRC="$(cd "$(dirname "$0")" && pwd)"

echo "Copying to ${USER}@${HOST}:${PORT} → ~/okx-ssr"
scp -P "$PORT" -r "$SRC" "${USER}@${HOST}:~/okx-ssr"
echo "Done. On Termux SSH session run:"
echo "  bash ~/okx-ssr/install-termux.sh"
echo "  BIND=0.0.0.0 ~/okx-ssr/okx-ssr"
