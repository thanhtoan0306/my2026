const http = require("http");

const PORT = Number(process.env.PORT || 3000);

function escapeHtml(s) {
  return String(s)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

function renderHtml({ name, nowIso }) {
  const safeName = escapeHtml(name);
  return `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Node SSR Hello</title>
    <style>
      body { font-family: ui-sans-serif, system-ui, -apple-system, Segoe UI, Roboto, Helvetica, Arial; margin: 40px; line-height: 1.4; }
      .card { max-width: 720px; border: 1px solid #ddd; border-radius: 12px; padding: 20px 22px; }
      code { background: #f6f6f6; padding: 2px 6px; border-radius: 6px; }
    </style>
  </head>
  <body>
    <div class="card">
      <h1>Hello, SSR from Node.js</h1>
      <p>This HTML is rendered on the server (no framework).</p>
      <p><strong>Name:</strong> <code>${safeName}</code></p>
      <p><strong>Server time:</strong> <code>${nowIso}</code></p>
      <hr />
      <p>Try: <code>/?name=Tony</code> or <code>/healthz</code></p>
    </div>
  </body>
</html>`;
}

const server = http.createServer((req, res) => {
  const url = new URL(req.url || "/", `http://${req.headers.host || "localhost"}`);

  if (url.pathname === "/healthz") {
    res.writeHead(200, { "content-type": "text/plain; charset=utf-8" });
    res.end("ok");
    return;
  }

  if (url.pathname !== "/") {
    res.writeHead(404, { "content-type": "text/plain; charset=utf-8" });
    res.end("not found");
    return;
  }

  const name = url.searchParams.get("name") || "world";
  const html = renderHtml({ name, nowIso: new Date().toISOString() });

  res.writeHead(200, { "content-type": "text/html; charset=utf-8" });
  res.end(html);
});

server.listen(PORT, "0.0.0.0", () => {
  // Keep log stable and minimal for container usage.
  console.log(`listening on http://0.0.0.0:${PORT}`);
});

