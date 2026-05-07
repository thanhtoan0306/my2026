# trackpconbrowser

Small localhost app to show what’s running on this Mac.

## Run

```bash
python3 may/trackpconbrowser/app.py
```

Then open `http://localhost:9123`.

## JSON APIs

- `http://localhost:9123/api/apps` – running GUI apps (from System Events)
- `http://localhost:9123/api/procs?limit=300` – process snapshot (from `ps`)

