# curl → JSON (Go SSR)

Go port of `may/29mayjsoncurl`: paste a **curl** command, run it server-side, show **beautified JSON**, copy via the icon beside the output.

Stack: **net/http** + **html/template** + embedded templates.

## Run

```bash
cd may/29mayjsoncurl-go
go run .
```

Open http://127.0.0.1:3031

## Notes

- Same validation rules as the Python and Rust versions.
- Non-JSON responses are shown as-is (still copyable).
- Listens on `127.0.0.1:3031` (Python `3029`, Rust `3030`).
