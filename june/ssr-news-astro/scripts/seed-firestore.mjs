#!/usr/bin/env node
/**
 * Seed Firestore from firestore/articles.seed.json
 *
 * Requires a Firebase service account JSON file:
 *   export FIREBASE_SERVICE_ACCOUNT=./path/to/serviceAccount.json
 *   npm run seed:firestore
 */
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { initializeApp, cert, getApps } from 'firebase-admin/app';
import { getFirestore } from 'firebase-admin/firestore';

const serviceAccountPath = process.env.FIREBASE_SERVICE_ACCOUNT;
if (!serviceAccountPath) {
	console.error('Set FIREBASE_SERVICE_ACCOUNT to your service account JSON path.');
	process.exit(1);
}

const serviceAccount = JSON.parse(readFileSync(resolve(serviceAccountPath), 'utf8'));
const seed = JSON.parse(readFileSync(resolve('firestore/articles.seed.json'), 'utf8'));

if (!getApps().length) {
	initializeApp({ credential: cert(serviceAccount) });
}

const db = getFirestore();
const collectionName = seed.collection ?? 'articles';

for (const document of seed.documents) {
	await db.collection(collectionName).doc(document.id).set(document.fields, { merge: true });
	console.log(`Seeded ${collectionName}/${document.id}`);
}

console.log(`Done — ${seed.documents.length} documents written.`);
