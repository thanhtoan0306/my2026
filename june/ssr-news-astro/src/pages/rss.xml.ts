import type { APIRoute } from 'astro';
import { CACHE_HOME, applyCacheHeader } from '../lib/cache';
import { getFeedArticles } from '../lib/news';
import { getSiteUrl, todayIsoDate } from '../lib/site';

export const prerender = false;

function escapeXml(value: string): string {
	return value
		.replaceAll('&', '&amp;')
		.replaceAll('<', '&lt;')
		.replaceAll('>', '&gt;')
		.replaceAll('"', '&quot;')
		.replaceAll("'", '&apos;');
}

export const GET: APIRoute = async ({ url }) => {
	const siteUrl = getSiteUrl(url);
	const displayDate = todayIsoDate();
	const { articles, mode } = await getFeedArticles(displayDate);

	const items = articles
		.map(
			(article) => `
    <item>
      <title>${escapeXml(article.title)}</title>
      <link>${siteUrl}/news/${article.slug}</link>
      <guid isPermaLink="true">${siteUrl}/news/${article.slug}</guid>
      <pubDate>${new Date(article.publishedAt).toUTCString()}</pubDate>
      <description>${escapeXml(article.summary)}</description>
    </item>`,
		)
		.join('');

	const body = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Crypto Today</title>
    <link>${siteUrl}/</link>
    <description>${mode === 'today' ? `Crypto news for ${displayDate}` : 'Latest crypto news'}</description>
    <language>en-us</language>
    <lastBuildDate>${new Date().toUTCString()}</lastBuildDate>${items}
  </channel>
</rss>`;

	const headers = new Headers({
		'Content-Type': 'application/rss+xml; charset=utf-8',
	});
	applyCacheHeader(headers, CACHE_HOME);

	return new Response(body, { headers });
};
