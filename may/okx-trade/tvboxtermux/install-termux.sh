#!/data/data/com.termux/files/usr/bin/bash
# Run this ON Termux after files are in ~/okx-ssr
set -e
cd ~/okx-ssr
pkg install -y golang
go mod tidy
go build -o okx-ssr .
echo ""
echo "Start (open from Mac/phone on same Wi‑Fi):"
echo "  BIND=0.0.0.0 ./okx-ssr"
echo "  → http://$(hostname -I 2>/dev/null | awk '{print $1}'):8091"
