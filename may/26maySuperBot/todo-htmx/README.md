# Go Todo — SSR + HTMX + Firestore

Server-rendered todo list with HTMX partial updates. **Todos are saved in Firebase Firestore** (Go Admin SDK on the server).

## Setup (one time)

1. **Service account** (not the web `apiKey` config):
   - Firebase Console → Project settings → **Service accounts**
   - **Generate new private key** → save as `serviceAccount.json` in this folder

2. **Project ID** — either:
   - `export FIREBASE_PROJECT_ID=your-project-id`, or
   - copy `config.json.example` → `config.json` with your `projectId`

3. **Firestore rules** — publish rules in Console (includes `todos` path):
   - see [../firebase-sticky-notes/firestore.rules](../firebase-sticky-notes/firestore.rules)

4. **Index** (first run may prompt): create composite index for `todos` collection on `createdAt` DESC, or follow the link in the error log.

## Run

```bash
cd may/26maySuperBot/todo-htmx
export GOOGLE_APPLICATION_CREDENTIALS="$(pwd)/serviceAccount.json"
export FIREBASE_PROJECT_ID=your-project-id   # or use config.json
go mod tidy
go run .
```

Open http://127.0.0.1:8080

## Where Firebase save happens

| Action | Code |
|--------|------|
| **Create todo** | `store_firestore.go` → `FirestoreStore.Create()` → `ref.Set()` |
| **Toggle** | `FirestoreStore.Toggle()` |
| **Delete** | `FirestoreStore.Delete()` |
| **List** | `FirestoreStore.List()` → `OrderBy createdAt` |

Path:

```text
/artifacts/{FIREBASE_APP_ID}/public/data/todos/{documentId}
```

Default `FIREBASE_APP_ID` = `FIREBASE_PROJECT_ID`.

## Stack

| Layer | Choice |
|--------|--------|
| Server | `net/http` + `html/template` |
| UI updates | HTMX 2 |
| Database | Cloud Firestore (Firebase) |

## Connect another backend to this database

See [../FIRESTORE_BACKEND_GUIDE.md](../FIRESTORE_BACKEND_GUIDE.md) — paths, schema, Go/Node/Python examples, rules, and Console links.

## Routes

| Method | Path | HTMX |
|--------|------|------|
| `GET /` | Full page | — |
| `POST /todos` | `#todo-list` | `todo_list.html` |
| `POST /todos/{id}/toggle` | `#todo-{id}` | `todo_item.html` |
| `DELETE /todos/{id}` | delete swap | 200 |
