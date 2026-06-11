export function getSiteUrl(requestUrl: URL): string {
	const fromEnv = import.meta.env.PUBLIC_SITE_URL;
	if (fromEnv) return fromEnv.replace(/\/$/, '');
	return `${requestUrl.protocol}//${requestUrl.host}`;
}

export function todayIsoDate(): string {
	return new Date().toISOString().slice(0, 10);
}

export function formatFullDate(isoDate: string): string {
	return new Intl.DateTimeFormat('en-US', {
		dateStyle: 'full',
		timeZone: 'UTC',
	}).format(new Date(isoDate));
}
