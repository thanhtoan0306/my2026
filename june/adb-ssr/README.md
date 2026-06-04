# ADB SSR

Go **server-side rendered** web UI to input and control Android devices via [ADB](https://developer.android.com/tools/adb). Browser-based companion to [`adb-desktop`](../adb-desktop/README.md).

## Requirements

- Go 1.22+
- `adb` on your PATH, or Android SDK at `~/Library/Android/sdk/platform-tools` (macOS), or set `ADB_PATH` to the full `adb` binary
- Device: USB debugging or wireless debugging enabled

## Quick start

```bash
cd june/adb-ssr
go run .
```

Open **http://127.0.0.1:8092** (override with `PORT`).

## Features

| Area | Description |
|------|-------------|
| **HTMX** | Partial page updates — no full reload on buttons/actions |
| **Apps** | List third-party apps (`pm list packages -3`), stop one, or **Close all** (`am force-stop` each) |
| **Connection** | `adb connect` / `disconnect` for network hosts (`IP` → `IP:5555`) |
| **Devices** | Live `adb devices -l` list (works with USB-only, no host required) |
| **Remote** | D-pad, Home, Back, Menu, volume, mute, power, recent, Enter |
| **Input** | `input text` and `input tap X Y` |
| **Shell** | Device shell via `sh -c` (pipes supported) |
| **Terminal** | Local `adb` commands on the server machine |

## Environment

| Variable | Example |
|----------|---------|
| `PORT` | `8092` (default) |
| `ADB_HOST` | `192.168.1.50:5555` |
| `ADB_SERIAL` | serial from `adb devices` |

## Network ADB

```bash
adb tcpip 5555
adb connect 192.168.1.50:5555
```

In the UI: set **Network host**, **Save**, then **Connect**.

## Related

- Electron app: [`june/adb-desktop`](../adb-desktop/README.md)
- Earlier SSR variant: [`june/controladb`](../controladb/README.md)
