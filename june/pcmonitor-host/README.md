# pcmonitor-host

Runs **on macOS host** (your real PC), not inside Docker/Colima.

## Run

Install Go if needed:

```bash
brew install go
```

Start:

```bash
cd june/pcmonitor-host
go mod tidy
go run .
```

Open: `http://127.0.0.1:8096`

## Public via Cloudflare Tunnel

```bash
brew install cloudflared
cloudflared tunnel --url http://127.0.0.1:8096
```

## Keep running after closing terminal (LaunchAgent)

This creates a per-user LaunchAgent (starts at login, restarts if it exits):

```bash
mkdir -p "$HOME/Library/LaunchAgents"
cat > "$HOME/Library/LaunchAgents/com.tony.pcmonitor-host.plist" <<'PLIST'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>Label</key><string>com.tony.pcmonitor-host</string>
    <key>RunAtLoad</key><true/>
    <key>KeepAlive</key><true/>
    <key>WorkingDirectory</key><string>__WORKDIR__</string>
    <key>EnvironmentVariables</key>
    <dict>
      <key>PORT</key><string>8096</string>
      <key>HOST</key><string>127.0.0.1</string>
      <!-- <key>MONITOR_TOKEN</key><string>change-me</string> -->
    </dict>
    <key>ProgramArguments</key>
    <array>
      <string>/opt/homebrew/bin/go</string>
      <string>run</string>
      <string>.</string>
    </array>
    <key>StandardOutPath</key><string>__WORKDIR__/pcmonitor-host.out.log</string>
    <key>StandardErrorPath</key><string>__WORKDIR__/pcmonitor-host.err.log</string>
  </dict>
</plist>
PLIST
perl -pi -e "s|__WORKDIR__|$PWD|g" "$HOME/Library/LaunchAgents/com.tony.pcmonitor-host.plist"
launchctl unload "$HOME/Library/LaunchAgents/com.tony.pcmonitor-host.plist" 2>/dev/null || true
launchctl load "$HOME/Library/LaunchAgents/com.tony.pcmonitor-host.plist"
```

Stop:

```bash
launchctl unload "$HOME/Library/LaunchAgents/com.tony.pcmonitor-host.plist"
```

