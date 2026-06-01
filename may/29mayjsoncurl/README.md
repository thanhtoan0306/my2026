# curl → JSON (Python SSR)

Small Flask app: paste a **curl** command, run it server-side, show **beautified JSON**, copy via the icon beside the output.

## Run

```bash
cd may/29mayjsoncurl
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
python app.py
```

Open http://127.0.0.1:3029

## Notes

- Only commands starting with `curl` are accepted; basic shell chaining is blocked.
- Non-JSON responses are shown as-is (still copyable).
- Runs locally on `127.0.0.1` — curl executes on your machine.
