# Supabase Todo (Go + REST API)

Todo list using **Supabase API keys** — no database password or connection string.

## 1. Create table (once)

Supabase → **SQL Editor** → paste and run `schema.sql`.

## 2. Get API keys (no password)

1. **Connect** → **API Keys** (or **Project Settings** → **API**)
2. Copy:
   - **Project URL** → `SUPABASE_URL`
   - **service_role** key → `SUPABASE_KEY` (server app only; do not expose in browser)

```bash
cp .env.example .env
# fill SUPABASE_URL and SUPABASE_KEY
```

## 3. Run

```bash
set -a && source .env && set +a
go run .
```

Open [http://127.0.0.1:8080](http://127.0.0.1:8080).

## Environment variables

| Variable | Required | Description |
|----------|----------|-------------|
| `SUPABASE_URL` | yes | `https://xxxx.supabase.co` |
| `SUPABASE_KEY` | yes | `service_role` secret key |
| `PORT` | no | default `8080` |

## Data location

**Table Editor** → table `items`.

## Postgres password vs API key

| Method | Env | Password in URI? |
|--------|-----|------------------|
| Postgres (old) | `SUPABASE_DB_URL` | Yes — hard with `@` in password |
| **REST API (this app)** | `SUPABASE_URL` + `SUPABASE_KEY` | **No** |

Use `anon` key only if you add stricter RLS policies; `service_role` bypasses RLS for a simple local server.
