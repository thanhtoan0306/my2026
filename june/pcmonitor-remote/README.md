# pcmonitor-remote

Small **Go SSR** page that shows this PC’s **CPU usage** and **RAM usage**, and can be published securely through **Cloudflare Tunnel** so you can view it from another machine’s browser.

## Run locally

```bash
cd june/pcmonitor-remote
go mod tidy
go run .
```

Open `http://127.0.0.1:8095`.

Optional:

```bash
PORT=3000 go run .
```

## Optional auth (recommended)

If you set `MONITOR_TOKEN`, the server requires either:

- Query token: `?token=...`
- Or header: `Authorization: Bearer ...`

Example:

```bash
export MONITOR_TOKEN="change-me"
go run .
```

Then open:

- `http://127.0.0.1:8095/?token=change-me`

## Publish via Cloudflare Tunnel

### Option A (quick, no DNS record)

Install `cloudflared`, then:

```bash
cloudflared tunnel --url http://127.0.0.1:8095
```

Cloudflare prints a public `https://...trycloudflare.com` URL. Open it in any browser.

### Option B (stable hostname)

Create a tunnel and map a hostname (high level):

```bash
cloudflared tunnel login
cloudflared tunnel create pcmonitor
cloudflared tunnel route dns pcmonitor pcmonitor.yourdomain.com
cloudflared tunnel run pcmonitor --url http://127.0.0.1:8095
```

For production, strongly consider adding **Cloudflare Access** in front of the hostname (instead of passing `?token=`).

## Endpoints

- `GET /` SSR page
- `GET /fragment/cpu` HTMX fragment (CPU card)
- `GET /fragment/ram` HTMX fragment (RAM card)

## Run with Docker / Colima (keeps running after terminal closes)

Build + start in background (auto-restarts unless you stop it):

```bash
cd june/pcmonitor-remote
docker compose up -d --build
```

Open `http://127.0.0.1:8095`.

Stop it:

```bash
docker compose down
```

See logs:

```bash
docker compose logs -f
```

### Docker + Cloudflare Tunnel

Start the container, then run the tunnel pointing to your local port:

```bash
cloudflared tunnel --url http://127.0.0.1:8095
```

## “htop-like” view

The UI is a lightweight “htop-ish” dashboard:

- CPU total + per-core bars
- Load average (1/5/15)
- Memory + Swap bars
- Top processes table (CPU/MEM sort)

