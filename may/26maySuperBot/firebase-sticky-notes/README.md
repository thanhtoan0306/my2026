# Firebase sticky notes

Client-side Firestore realtime notes. **Not** in `.idea/`.

## Run

```bash
cp config.json.example config.json   # paste your Firebase web config
cd firebase-sticky-notes
python3 -m http.server 8080
```

Open http://localhost:8080/

Console: enable **Anonymous** auth + publish `firestore.rules`.

## Where save logic lives

| Action | File | Function |
|--------|------|----------|
| **Save note** | `index.html` | `saveNoteToFirestore()` → `addDoc(...)` |
| **Delete note** | `index.html` | `deleteDoc(...)` on delete button |
| **Read/sync** | `index.html` | `listenNotes()` → `onSnapshot(...)` |
| **Login** | `index.html` | `signInAnonymously(auth)` in `boot()` |
| **Config** | `config.json` | loaded by `loadConfig()` |

Firestore path:

```text
/artifacts/{projectId}/public/data/notes/{autoId}
```

`projectId` comes from `config.json`.

## `todo-htmx/`

Go todo app uses **in-memory** storage only — no Firebase.
