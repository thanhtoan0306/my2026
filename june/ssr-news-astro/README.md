# ssr-news-astro

Server-rendered crypto news site built with [Astro](https://astro.build) and the Node adapter.

## Quick start

```bash
npm install
npm run dev
```

Open [http://localhost:4321](http://localhost:4321).

## SSR mode

- `output: 'server'` in `astro.config.mjs`
- `@astrojs/node` adapter (`standalone` mode)
- Pages use `export const prerender = false` so HTML is rendered on each request

## Content

Three mock crypto articles for **11 Jun 2026** live in `src/lib/news.ts` until Firebase is connected.

## Firebase

1. Create a Firebase project and enable **Firestore**.
2. Copy `.env.example` → `.env` and paste your web app config (`PUBLIC_FIREBASE_*`).
3. Deploy read-only rules from `firestore.rules` (public read on `articles`).
4. Seed Firestore (service account required):

```bash
export FIREBASE_SERVICE_ACCOUNT=./path/to/serviceAccount.json
npm run seed:firestore
```

Or import `firestore/articles.seed.json` manually in the Firebase console.

5. Restart dev server — SSR pages will read Firestore automatically.

If Firebase is missing or empty, the app falls back to mock data in `src/data/mock-articles.ts`.

### Firestore document shape

| Field         | Type     |
| ------------- | -------- |
| `slug`        | string   |
| `title`       | string   |
| `summary`     | string   |
| `body`        | string   |
| `category`    | string   |
| `author`      | string   |
| `publishedAt` | string (ISO) |
| `tags`        | string[] |
| `imageUrl`    | string (optional) |

## Caching (ISR-style)

Pages set `Cache-Control` so a CDN can cache rendered HTML:

- Home feed: 60s (`stale-while-revalidate=300`)
- Article pages: 5 min (`stale-while-revalidate=3600`)

Tune values in `src/lib/cache.ts`.

## Docker

```bash
docker compose up --build
```

Or manually:

```bash
npm run build
npm run start
```

Health check: `GET /health`  
RSS feed: `GET /rss.xml`  
Sitemap: `GET /sitemap.xml`  
Robots: `GET /robots.txt`

## Scripts

| Command           | Action                          |
| ----------------- | ------------------------------- |
| `npm run dev`     | Dev server at `localhost:4321`  |
| `npm run build`   | Production build to `dist/`     |
| `npm run start`   | Run standalone SSR server       |
| `npm run preview` | Preview production SSR server   |
| `npm run test:smoke` | Hit key routes (server must be running) |
| `npm run check`   | Astro type check                |

## Project layout

```
src/
  components/   Layout, NewsCard
  lib/          news.ts (data), firebase.ts (stub)
  pages/        index.astro, news/[slug].astro
  styles/       global.css
  types/        news.ts
```
