export interface Article {
	id: string;
	slug: string;
	title: string;
	summary: string;
	body: string;
	category: string;
	author: string;
	publishedAt: string;
	imageUrl?: string;
	tags: string[];
}
