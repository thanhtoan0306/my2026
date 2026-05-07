#!/usr/bin/env python3
from __future__ import annotations

import base64
import os
from email.message import EmailMessage
from pathlib import Path

from app import build_weather_email_html, fetch_hcm_today_weather


def _env(name: str, *, required: bool = True) -> str | None:
    v = os.getenv(name)
    if required and (v is None or not v.strip()):
        raise SystemExit(f"Missing env var: {name}")
    return v.strip() if v is not None else None


def _load_gmail_credentials(*, creds_path: Path, token_path: Path):
    # Lazy imports so SMTP-only usage doesn't require these deps.
    from google.auth.transport.requests import Request
    from google.oauth2.credentials import Credentials
    from google_auth_oauthlib.flow import InstalledAppFlow

    scopes = ["https://www.googleapis.com/auth/gmail.send"]

    creds = None
    if token_path.exists():
        creds = Credentials.from_authorized_user_file(str(token_path), scopes=scopes)

    if creds and creds.expired and creds.refresh_token:
        creds.refresh(Request())
        token_path.write_text(creds.to_json(), encoding="utf-8")
        return creds

    if creds and creds.valid:
        return creds

    if not creds_path.exists():
        raise SystemExit(
            "\n".join(
                [
                    f"Missing OAuth client file: {creds_path}",
                    "",
                    "Create it in Google Cloud Console as an OAuth client ID (Desktop app),",
                    "download the JSON, and save it as credentials.json in this folder.",
                ]
            )
        )

    flow = InstalledAppFlow.from_client_secrets_file(str(creds_path), scopes=scopes)
    creds = flow.run_local_server(port=0)
    token_path.write_text(creds.to_json(), encoding="utf-8")
    return creds


def _build_raw_message(*, from_email: str, to_email: str, subject: str, html_body: str) -> str:
    msg = EmailMessage()
    msg["To"] = to_email
    msg["From"] = from_email
    msg["Subject"] = subject
    msg.set_content("This email contains HTML. If you see this, your client doesn't support HTML.")
    msg.add_alternative(html_body, subtype="html")

    raw_bytes = msg.as_bytes()
    return base64.urlsafe_b64encode(raw_bytes).decode("ascii")


def send_with_gmail_api(*, to_email: str, subject: str, html_body: str) -> None:
    # Lazy import to keep install optional.
    from googleapiclient.discovery import build

    here = Path(__file__).resolve().parent
    creds = _load_gmail_credentials(
        creds_path=here / "credentials.json",
        token_path=here / "token.json",
    )

    # "me" means the currently authenticated user.
    service = build("gmail", "v1", credentials=creds)
    # Avoid calling getProfile() which requires additional scopes.
    from_email = _env("GMAIL_FROM", required=False) or "me"

    raw = _build_raw_message(from_email=from_email, to_email=to_email, subject=subject, html_body=html_body)
    service.users().messages().send(userId="me", body={"raw": raw}).execute()
    print(f"Sent to {to_email}: {subject}")


def main() -> None:
    to_email = _env("TO_EMAIL")
    w = fetch_hcm_today_weather()
    subject, html_body = build_weather_email_html(w)
    send_with_gmail_api(to_email=to_email, subject=subject, html_body=html_body)


if __name__ == "__main__":
    main()

