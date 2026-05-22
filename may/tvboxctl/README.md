# tvboxctl

SSR dashboard (Go + `html/template`) to control an Android TV box over **ADB** and **SSH**.

## Prerequisites

- Go 1.22+
- [`adb`](https://developer.android.com/tools/adb) on your PATH
- TV box: enable **Developer options** → **USB debugging** or **Network debugging** (wireless ADB, usually port `5555`)
- SSH enabled on the box (many firmwares use `root` + password or key)

## Run

```bash
cd may/tvboxctl
go mod tidy
go run .
```

Open **http://127.0.0.1:8090**

## Environment (optional defaults)

| Variable | Example |
|----------|---------|
| `PORT` | `8090` |
| `ADB_HOST` | `192.168.1.50:5555` |
| `ADB_SERIAL` | device id from `adb devices` |
| `SSH_HOST` | `192.168.1.50` |
| `SSH_PORT` | `22` |
| `SSH_USER` | `root` |
| `SSH_KEY` | `/Users/you/.ssh/id_rsa` |
| `SSH_PASSWORD` | (if no key) |

## ADB tips

```bash
# First-time network connect
adb connect 192.168.1.50:5555
adb devices
```

On the dashboard: set **ADB host**, click **Check status**, then use remote buttons (Home, Back, D-pad, etc.).

## Security

Binds to **127.0.0.1** only. Do not expose this app to the internet without authentication.
