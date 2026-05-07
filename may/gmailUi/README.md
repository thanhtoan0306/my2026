## gmailUi — send Gmail HTML weather (HCMC)

This sends an **HTML email** with **today’s weather in Ho Chi Minh City**.

## Option A (simple): Gmail SMTP + App Password

- **Use an App Password** (recommended). Your Google account must have 2‑Step Verification enabled, then create an “App password”.
- Do **not** use your normal Gmail password in scripts.

### Run

From repo root:

```bash
export GMAIL_USER="your@gmail.com"
export GMAIL_APP_PASSWORD="xxxx xxxx xxxx xxxx"   # Gmail App Password
export TO_EMAIL="to@example.com"                  # optional (defaults to GMAIL_USER)

python3 may/gmailUi/app.py
```

### Output

If successful, you’ll see:

```bash
Sent to to@example.com: Weather HCMC Today (YYYY-MM-DD): ...
```

## Option B (no App Password): Gmail API + OAuth2 (recommended)

This uses **Gmail API** with **OAuth2**. First run will open a browser for Google login and create `token.json`.

### 1) Install dependencies

```bash
python3 -m pip install -r may/gmailUi/requirements.txt
```

### 2) Create `credentials.json`

- Go to Google Cloud Console → create/select a project
- Enable **Gmail API**
- Configure **OAuth consent screen** (External is fine for personal use)
- Create **OAuth client ID** → choose **Desktop app**
- Download the JSON and save it here:

`may/gmailUi/credentials.json`

### 3) Run (OAuth flow)

From repo root:

```bash
export TO_EMAIL="rd_tony@bpo-it.net"
python3 may/gmailUi/send_gmail_api.py
```

Files created locally:
- `may/gmailUi/token.json` (OAuth token; keep private)


