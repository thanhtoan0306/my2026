# simplewsokx

One HTML file. WebSocket only. Credentials inline — **no server, no fetch**.

## Setup

1. Open `index.html`
2. Edit the `CONFIG` block (~line 127):

```javascript
const CONFIG = {
  apiKey: "...",
  secretKey: "...",
  passphrase: "...",
  env: "real"   // or "demo"
};
```

3. Open the file in Termux browser / TV browser (`file://` or tap the file in Files)

Auto-connects WS on load.

## This device only

```bash
termux-open index.html
```

## Other devices on Wi‑Fi (Node)

Install Node on Termux once: `pkg install nodejs`

```bash
cd ~/simplewsokx
npm start
```

Or without `package.json`:

```bash
npx --yes serve -l 8088 .
```

Then on Mac/TV browser: **`http://PHONE_IP:8088`** (e.g. `http://192.168.1.153:8088`)

Get IP on Termux: `ip -4 addr show wlan0 | grep inet`

## Other devices (Python)

```bash
python -m http.server 8088 -b 0.0.0.0
```

Same URL: `http://PHONE_IP:8088`

**Security:** do not commit real keys. Keep your copy local only.
