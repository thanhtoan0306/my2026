# 26maySuperBot

## Firebase sticky notes (`firebase-sticky-notes/`)

Realtime notes — save/read via Firestore in the browser.

```bash
cd firebase-sticky-notes
cp config.json.example config.json   # add your Firebase config
python3 -m http.server 8080
```

Save logic: `firebase-sticky-notes/index.html` → `saveNoteToFirestore()` / `addDoc()`.

Details: [firebase-sticky-notes/README.md](firebase-sticky-notes/README.md)

## Go Todo — SSR + HTMX + Firestore (`todo-htmx/`)

Server-rendered todo list; **saves to Firestore** via Go Admin SDK.

```bash
cd todo-htmx
export GOOGLE_APPLICATION_CREDENTIALS="$(pwd)/serviceAccount.json"
go run .
```

→ http://127.0.0.1:8080/dashboard

Details: [todo-htmx/README.md](todo-htmx/README.md)

**New backend?** Connect to the same Firestore: [FIRESTORE_BACKEND_GUIDE.md](FIRESTORE_BACKEND_GUIDE.md)
