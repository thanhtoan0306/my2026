# Telegram group sync → JSON → Drive

SSR tool to sync **today’s messages** from a [Telegram Web](https://web.telegram.org/k/) group into JSON and upload to the **TelegramBackup** folder on Google Drive.

Default chat: `-5045031903` ([web link](https://web.telegram.org/k/#-5045031903))

Uses the **Telegram user API** (same account as Telegram Web), not browser scraping.

## Setup

1. Create an app at [my.telegram.org/apps](https://my.telegram.org/apps) → note **api_id** and **api_hash**.

2. Drive credentials must exist in `../dbstorageggdrive/` (`credentials.json`, `token.json`, `telegram_backup_folder_id.txt`).

```bash
cd june/telegramsync
export TELEGRAM_API_ID=12345678
export TELEGRAM_API_HASH=your_api_hash
go mod tidy
go run . login          # phone + SMS/Telegram code in terminal
go run .                # SSR UI → http://127.0.0.1:8095
```

3. Open the UI and click **Sync chat**.

## CLI

```bash
go run . sync           # sync today without UI
```

## Env

| Variable | Default |
|----------|---------|
| `TELEGRAM_API_ID` | required |
| `TELEGRAM_API_HASH` | required |
| `TELEGRAM_CHAT_ID` | `-5045031903` |
| `DRIVE_CONFIG_DIR` | `../dbstorageggdrive` |
| `DRIVE_TELEGRAM_FOLDER_ID` | from `telegram_backup_folder_id.txt` |
| `PORT` | `8095` |

## Output

`telegram-{chatId}-{YYYY-MM-DD}.json` locally and in Drive **TelegramBackup**, for example:

```json
{
  "chatId": -5045031903,
  "chatUrl": "https://web.telegram.org/k/#-5045031903",
  "date": "2026-06-02",
  "syncedAt": "2026-06-02T12:00:00+07:00",
  "count": 42,
  "messages": [ … ]
}
```
