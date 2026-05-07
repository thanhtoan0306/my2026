# controlport (macOS)

Localhost web UI that lists ports currently in use (TCP LISTEN) and lets you kill the owning process.

## Run

```bash
python3 app.py
```

Then open `http://127.0.0.1:3011`.

## Notes / Safety

- Binds to `127.0.0.1` only (localhost).
- Uses `lsof` to enumerate listening sockets, and `kill` to terminate processes.
- You may only be able to kill processes owned by your user (unless you run with elevated privileges).

