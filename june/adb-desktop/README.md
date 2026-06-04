# ADB Desktop

Electron desktop app to **input and control** Android devices over [ADB](https://developer.android.com/tools/adb). Works with USB devices and network debugging (TV boxes, phones, emulators).

## Requirements

- [Node.js](https://nodejs.org/) 18+
- `adb` on your PATH ([Android platform-tools](https://developer.android.com/tools/releases/platform-tools))
- Target device: **Developer options** → USB debugging or **Wireless debugging**

## Quick start

```bash
cd june/adb-desktop
npm install
npm start
```

## Features

| Area | What it does |
|------|----------------|
| **Connection** | `adb connect` / `disconnect` to `IP` or `IP:5555` (default port 5555) |
| **Devices** | Live list from `adb devices -l`; click to set serial |
| **Remote** | D-pad, Home, Back, Menu, volume, power, recent, Enter |
| **Input** | `adb shell input text` and `input tap X Y` |
| **Shell** | Arbitrary `adb shell` commands on the device |
| **Terminal** | Local `adb` commands on this Mac/PC (preset picker included) |

Settings (host, serial) are stored under the app user data directory.

## Network ADB (e.g. TV box)

On the device (USB connected once):

```bash
adb tcpip 5555
adb connect 192.168.1.50:5555
```

In the app: enter `192.168.1.50` or `192.168.1.50:5555`, click **Connect**, then use remote controls.

## Build installers

```bash
npm run dist
```

Output in `dist/` (DMG on macOS, NSIS on Windows, AppImage/deb on Linux).

## Related

- Go SSR (browser): [`june/adb-ssr`](../adb-ssr/README.md)
- Earlier SSR variant: [`june/controladb`](../controladb/README.md)
