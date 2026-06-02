# Google Drive public storage

Small Go app that creates a **public** Google Drive folder (anyone with the link can view) and uploads images/files there with shareable URLs.

## One-time Google setup

1. Open [Google Cloud Console](https://console.cloud.google.com/) → create or pick a project.
2. **APIs & Services** → **Library** → enable **Google Drive API**.
3. **APIs & Services** → **Credentials** → **Create credentials** → **OAuth client ID**.
   - Application type: **Desktop app** (or Web application with redirect `http://127.0.0.1:8093/oauth/callback`).
4. Download the JSON and save it as `credentials.json` in this directory.

## Run

```bash
cd june/dbstorageggdrive
go mod tidy
go run . init          # create folder + set public (anyone: reader)
go run .               # SSR web UI on http://127.0.0.1:8094
```

First run opens a browser for Google sign-in; token is saved to `token.json` (gitignored).

### CLI upload

```bash
go run . upload ./photo.png
```

### HTTP API

```bash
curl -F "file=@photo.png" -F "action=upload" "http://127.0.0.1:8094/?format=json" -H "Accept: application/json"
```

Response includes `viewUrl`, `directUrl` (images), and `downloadUrl`.

## Public URLs

| Use | URL pattern |
|-----|-------------|
| Folder | `https://drive.google.com/drive/folders/{folderId}` |
| File view | `https://drive.google.com/file/d/{fileId}/view` |
| Image embed | `https://drive.google.com/uc?export=view&id={fileId}` |
| Download | `https://drive.google.com/uc?export=download&id={fileId}` |

`folder_id.txt` stores the folder id after `init`. Override folder name with `DRIVE_FOLDER_NAME` (default `PublicStorage`).

## Env

| Variable | Default |
|----------|---------|
| `PORT` | `8094` (8092 is used by okxwebsocket) |

## SSR pages

| Route | Description |
|-------|-------------|
| `GET /` | List files in the public folder (thumbnails for images) |
| `POST /` | Upload (`action=upload`, field `file`) |
| `GET /file/{id}` | File detail, image preview, text preview for small text files |
| `DRIVE_FOLDER_NAME` | `PublicStorage` |

## Notes

- Files inherit the folder’s **anyone with the link → viewer** permission.
- Google may show a virus-scan warning for large files; direct hotlinking can be rate-limited.
- Do not commit `credentials.json` or `token.json`.
