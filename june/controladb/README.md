# Control ADB

Go SSR web UI to control Android devices over the network via `adb`.

## Requirements

- Go 1.22+
- [`adb`](https://developer.android.com/tools/adb) on your PATH (Android SDK `platform-tools`)
- Device with **Wireless debugging** or **Network debugging** enabled

## Run

```bash
cd june/controladb
go run .
```

Open **http://127.0.0.1:8091** (override with `PORT`).

## Usage

1. Enter the device **IP** (e.g. `192.168.1.50`). Port `:5555` is added automatically if omitted.
2. Click **Save IP**, then **Check status** or use remote buttons.
3. Optional: set `ADB_SERIAL` env or the serial field if you have multiple devices.

```bash
export ADB_HOST=192.168.1.50:5555
go run .
```

On the device (one-time or after reboot, depending on ROM):

```bash
adb tcpip 5555
adb connect 192.168.1.50:5555
```

## Features

- Connect / disconnect by IP
- D-pad and common key events (Home, Back, Power, volume, etc.)
- Send text via `input text`
- Run arbitrary `adb shell` commands
- Reboot device
