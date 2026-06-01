# curl → JSON (Rust SSR)

Rust port of `may/29mayjsoncurl`: paste a **curl** command, run it server-side, show **beautified JSON**, copy via the icon beside the output.

Stack: **axum** + **minijinja** + **tokio**.

## Run

```bash
cd may/29mayjsoncurl-rust
cargo run
```

Open http://127.0.0.1:3030

Release build:

```bash
cargo run --release
```

## Notes

- Same validation rules as the Python version (curl-only, no shell chaining).
- Non-JSON responses are shown as-is (still copyable).
- Listens on `127.0.0.1:3030` (Python version uses `3029`).
