#!/bin/bash
# Run on your Mac — copies app to Termux
set -e
HOST="${TERMUX_HOST:-192.168.1.153}"
PORT="${TERMUX_PORT:-8022}"
USER="${TERMUX_USER:-u0_a66}"
SRC="$(cd "$(dirname "$0")" && pwd)"

echo "Copying to ${USER}@${HOST}:${PORT} → ~/helloword-go"
scp -P "$PORT" -r "$SRC" "${USER}@${HOST}:~/helloword-go"
echo ""
echo "On Termux SSH session:"
echo "  bash ~/helloword-go/install-termux.sh"
echo "  ~/helloword-go/helloword"
