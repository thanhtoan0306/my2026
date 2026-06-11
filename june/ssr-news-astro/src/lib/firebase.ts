import { initializeApp, getApps, type FirebaseApp } from 'firebase/app';
import {
	collection,
	doc,
	getDoc,
	getDocs,
	getFirestore,
	limit,
	orderBy,
	query,
	where,
} from 'firebase/firestore';
import type { Article } from '../types/news';

const firebaseConfig = {
	apiKey: import.meta.env.PUBLIC_FIREBASE_API_KEY,
	authDomain: import.meta.env.PUBLIC_FIREBASE_AUTH_DOMAIN,
	projectId: import.meta.env.PUBLIC_FIREBASE_PROJECT_ID,
	storageBucket: import.meta.env.PUBLIC_FIREBASE_STORAGE_BUCKET,
	messagingSenderId: import.meta.env.PUBLIC_FIREBASE_MESSAGING_SENDER_ID,
	appId: import.meta.env.PUBLIC_FIREBASE_APP_ID,
};

const ARTICLES_COLLECTION = 'articles';

let app: FirebaseApp | undefined;

function getApp(): FirebaseApp {
	if (!app) {
		app = getApps().length ? getApps()[0]! : initializeApp(firebaseConfig);
	}
	return app;
}

export const isFirebaseConfigured = (): boolean =>
	Boolean(firebaseConfig.projectId && firebaseConfig.apiKey);

function toIsoString(value: unknown): string {
	if (typeof value === 'string') return value;
	if (value && typeof value === 'object' && 'toDate' in value && typeof value.toDate === 'function') {
		return value.toDate().toISOString();
	}
	return new Date().toISOString();
}

function docToArticle(id: string, data: Record<string, unknown>): Article {
	return {
		id,
		slug: String(data.slug ?? id),
		title: String(data.title ?? ''),
		summary: String(data.summary ?? ''),
		body: String(data.body ?? ''),
		category: String(data.category ?? 'Crypto'),
		author: String(data.author ?? 'Editorial'),
		publishedAt: toIsoString(data.publishedAt),
		imageUrl: data.imageUrl ? String(data.imageUrl) : undefined,
		tags: Array.isArray(data.tags) ? data.tags.map(String) : [],
	};
}

export async function fetchAllArticlesFromFirebase(): Promise<Article[]> {
	const db = getFirestore(getApp());
	const snapshot = await getDocs(
		query(collection(db, ARTICLES_COLLECTION), orderBy('publishedAt', 'desc')),
	);

	return snapshot.docs.map((docSnap) => docToArticle(docSnap.id, docSnap.data()));
}

export async function fetchArticlesFromFirebase(forDate = new Date().toISOString().slice(0, 10)): Promise<Article[]> {
	const db = getFirestore(getApp());
	const snapshot = await getDocs(
		query(collection(db, ARTICLES_COLLECTION), orderBy('publishedAt', 'desc')),
	);

	return snapshot.docs
		.map((docSnap) => docToArticle(docSnap.id, docSnap.data()))
		.filter((article) => article.publishedAt.startsWith(forDate));
}

export async function fetchArticleBySlug(slug: string): Promise<Article | undefined> {
	const db = getFirestore(getApp());
	const byId = await getDoc(doc(db, ARTICLES_COLLECTION, slug));
	if (byId.exists()) return docToArticle(byId.id, byId.data());

	const snapshot = await getDocs(
		query(collection(db, ARTICLES_COLLECTION), where('slug', '==', slug), limit(1)),
	);
	const match = snapshot.docs[0];
	return match ? docToArticle(match.id, match.data()) : undefined;
}
