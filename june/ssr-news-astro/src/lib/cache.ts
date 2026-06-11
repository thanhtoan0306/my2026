/** CDN-friendly caching — SSR renders once, edge caches the HTML (ISR-style). */

export const CACHE_HOME = 'public, s-maxage=60, stale-while-revalidate=300';
export const CACHE_ARTICLE = 'public, s-maxage=300, stale-while-revalidate=3600';

export function applyCacheHeader(headers: Headers, value: string): void {
	headers.set('Cache-Control', value);
}
