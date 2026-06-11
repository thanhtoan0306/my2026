import type { APIRoute } from 'astro';
import { getSiteUrl } from '../lib/site';

export const prerender = false;

export const GET: APIRoute = ({ url }) => {
	const siteUrl = getSiteUrl(url);
	const body = `User-agent: *
Allow: /

Sitemap: ${siteUrl}/sitemap.xml
`;

	return new Response(body, {
		headers: { 'Content-Type': 'text/plain; charset=utf-8' },
	});
};
