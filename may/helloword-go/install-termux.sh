#!/data/data/com.termux/files/usr/bin/bash
# Run on Termux after files are in ~/helloword-go
set -e
cd ~/helloword-go
pkg install -y golang
go mod tidy
go build -o helloword .
echo ""
echo "Start (same Wi‑Fi, open in browser):"
echo "  ~/helloword-go/helloword"
echo "  → http://$(hostname -I 2>/dev/null | awk '{print $1}'):8080"
