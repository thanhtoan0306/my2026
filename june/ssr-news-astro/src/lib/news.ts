import { MOCK_ARTICLES } from '../data/mock-articles';
import type { Article } from '../types/news';
import {
	fetchAllArticlesFromFirebase,
	fetchArticleBySlug,
	fetchArticlesFromFirebase,
	isFirebaseConfigured,
} from './firebase';

function sortByPublishedAt(articles: Article[]): Article[] {
	return [...articles].sort(
		(a, b) => new Date(b.publishedAt).getTime() - new Date(a.publishedAt).getTime(),
	);
}

export async function getAllArticles(): Promise<Article[]> {
	if (isFirebaseConfigured()) {
		try {
			const articles = await fetchAllArticlesFromFirebase();
			if (articles.length > 0) return sortByPublishedAt(articles);
		} catch (error) {
			console.error('[news] Firebase fetch failed, using mock data:', error);
		}
	}

	return sortByPublishedAt(MOCK_ARTICLES);
}

export type FeedMode = 'today' | 'latest';

export async function getFeedArticles(
	forDate = new Date().toISOString().slice(0, 10),
): Promise<{ articles: Article[]; mode: FeedMode }> {
	const todayArticles = await getTodayArticles(forDate);
	if (todayArticles.length > 0) {
		return { articles: todayArticles, mode: 'today' };
	}

	const latest = (await getAllArticles()).slice(0, 10);
	return { articles: latest, mode: 'latest' };
}

export async function getTodayArticles(forDate = new Date().toISOString().slice(0, 10)): Promise<Article[]> {
	if (isFirebaseConfigured()) {
		try {
			const articles = await fetchArticlesFromFirebase(forDate);
			if (articles.length > 0) return articles;
		} catch (error) {
			console.error('[news] Firebase fetch failed, using mock data:', error);
		}
	}

	return sortByPublishedAt(
		MOCK_ARTICLES.filter((article) => article.publishedAt.startsWith(forDate)),
	);
}

export async function getArticleBySlug(slug: string): Promise<Article | undefined> {
	if (isFirebaseConfigured()) {
		try {
			const article = await fetchArticleBySlug(slug);
			if (article) return article;
		} catch (error) {
			console.error('[news] Firebase article lookup failed, using mock data:', error);
		}
	}

	return MOCK_ARTICLES.find((article) => article.slug === slug);
}

export function formatPublishedAt(iso: string): string {
	return new Intl.DateTimeFormat('en-US', {
		dateStyle: 'medium',
		timeStyle: 'short',
		timeZone: 'UTC',
	}).format(new Date(iso));
}
