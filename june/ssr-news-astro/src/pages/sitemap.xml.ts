import type { APIRoute } from 'astro';
import { CACHE_HOME, applyCacheHeader } from '../lib/cache';
import { getAllArticles } from '../lib/news';
import { getSiteUrl } from '../lib/site';

export const prerender = false;

export const GET: APIRoute = async ({ url }) => {
	const siteUrl = getSiteUrl(url);
	const articles = await getAllArticles();
	const lastmod = new Date().toISOString().slice(0, 10);

	const urls = [
		`<url><loc>${siteUrl}/</loc><lastmod>${lastmod}</lastmod><changefreq>hourly</changefreq><priority>1.0</priority></url>`,
		...articles.map(
			(article) =>
				`<url><loc>${siteUrl}/news/${article.slug}</loc><lastmod>${article.publishedAt.slice(0, 10)}</lastmod><changefreq>weekly</changefreq><priority>0.8</priority></url>`,
		),
	];

	const body = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${urls.join('\n')}
</urlset>`;

	const headers = new Headers({
		'Content-Type': 'application/xml; charset=utf-8',
	});
	applyCacheHeader(headers, CACHE_HOME);

	return new Response(body, { headers });
};
