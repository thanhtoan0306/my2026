#!/usr/bin/env node
const base = (process.env.BASE_URL ?? 'http://localhost:4321').replace(/\/$/, '');

const routes = [
	{ path: '/', status: 200 },
	{ path: '/health', status: 200 },
	{ path: '/rss.xml', status: 200 },
	{ path: '/sitemap.xml', status: 200 },
	{ path: '/robots.txt', status: 200 },
	{ path: '/news/bitcoin-etf-inflows-record-june-2026', status: 200 },
	{ path: '/news/missing-story', status: 404 },
];

let failed = 0;

for (const { path, status } of routes) {
	const response = await fetch(`${base}${path}`);
	if (response.status !== status) {
		console.error(`FAIL ${path} expected ${status} got ${response.status}`);
		failed += 1;
		continue;
	}
	console.log(`OK   ${path} (${status})`);
}

if (failed > 0) {
	process.exit(1);
}

console.log(`Smoke test passed — ${routes.length} routes`);
