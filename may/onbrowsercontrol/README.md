# onbrowsercontrol (macOS)

Localhost web UI to control your Mac system volume from a browser.

## Run

```bash
python3 app.py
```

Then open `http://127.0.0.1:3010`.

## Notes

- Uses `osascript` to set **system output volume** (0–100).
- This is meant for local network / localhost use. If you bind to `0.0.0.0`, anyone on your LAN could change your volume.

