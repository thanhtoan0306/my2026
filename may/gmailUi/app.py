#!/usr/bin/env python3
from __future__ import annotations

import os
import smtplib
import ssl
import urllib.error
import urllib.parse
import urllib.request
from dataclasses import dataclass
from datetime import date, datetime
from email.message import EmailMessage


@dataclass(frozen=True)
class WeatherToday:
    city: str
    tz: str
    today: date
    t_min_c: float | None
    t_max_c: float | None
    precip_mm: float | None
    wind_max_kph: float | None
    weather_code: int | None
    fetched_at: datetime


def _env(name: str, *, required: bool = True) -> str | None:
    v = os.getenv(name)
    if required and (v is None or not v.strip()):
        raise SystemExit(f"Missing env var: {name}")
    return v.strip() if v is not None else None


def _http_get_json(url: str, *, timeout_s: float = 15.0) -> dict:
    req = urllib.request.Request(
        url,
        headers={
            "User-Agent": "gmailUi-weather/1.0",
            "Accept": "application/json",
        },
        method="GET",
    )
    try:
        with urllib.request.urlopen(req, timeout=timeout_s) as resp:
            raw = resp.read()
    except urllib.error.HTTPError as e:
        body = ""
        try:
            body = e.read().decode("utf-8", "replace")
        except Exception:
            body = ""
        raise RuntimeError(f"HTTP {e.code} from weather API: {body[:300]}") from e
    except Exception as e:
        raise RuntimeError(f"Failed to fetch weather API: {e}") from e

    import json

    try:
        return json.loads(raw.decode("utf-8"))
    except Exception as e:
        raise RuntimeError(f"Invalid JSON from weather API: {e}") from e


def fetch_hcm_today_weather() -> WeatherToday:
    # Ho Chi Minh City center-ish coordinates
    lat = 10.8231
    lon = 106.6297
    tz = "Asia/Ho_Chi_Minh"

    params = {
        "latitude": f"{lat:.4f}",
        "longitude": f"{lon:.4f}",
        "timezone": tz,
        "daily": ",".join(
            [
                "temperature_2m_max",
                "temperature_2m_min",
                "precipitation_sum",
                "wind_speed_10m_max",
                "weather_code",
            ]
        ),
        "forecast_days": "1",
    }
    url = "https://api.open-meteo.com/v1/forecast?" + urllib.parse.urlencode(params)
    j = _http_get_json(url)

    daily = j.get("daily") or {}
    tmax = (daily.get("temperature_2m_max") or [None])[0]
    tmin = (daily.get("temperature_2m_min") or [None])[0]
    precip = (daily.get("precipitation_sum") or [None])[0]
    wind_max = (daily.get("wind_speed_10m_max") or [None])[0]
    wcode = (daily.get("weather_code") or [None])[0]

    def to_float(x) -> float | None:
        try:
            return None if x is None else float(x)
        except Exception:
            return None

    def to_int(x) -> int | None:
        try:
            return None if x is None else int(x)
        except Exception:
            return None

    return WeatherToday(
        city="Ho Chi Minh City",
        tz=tz,
        today=date.today(),
        t_min_c=to_float(tmin),
        t_max_c=to_float(tmax),
        precip_mm=to_float(precip),
        wind_max_kph=to_float(wind_max),
        weather_code=to_int(wcode),
        fetched_at=datetime.now(),
    )


def _wmo_label(code: int | None) -> str:
    # WMO weather interpretation codes (Open-Meteo uses these).
    # Keep labels short for email subject/body.
    if code is None:
        return "—"
    if code == 0:
        return "Clear"
    if code in (1, 2, 3):
        return "Partly cloudy"
    if code in (45, 48):
        return "Fog"
    if code in (51, 53, 55):
        return "Drizzle"
    if code in (61, 63, 65):
        return "Rain"
    if code in (66, 67):
        return "Freezing rain"
    if code in (71, 73, 75, 77):
        return "Snow"
    if code in (80, 81, 82):
        return "Showers"
    if code in (95, 96, 99):
        return "Thunderstorm"
    return f"Code {code}"


def _fmt(x: float | None, *, suffix: str = "", digits: int = 1) -> str:
    if x is None:
        return "—"
    return f"{x:.{digits}f}{suffix}"


def build_weather_email_html(w: WeatherToday) -> tuple[str, str]:
    summary = _wmo_label(w.weather_code)
    subject = f"Weather HCMC Today ({w.today.isoformat()}): {summary}"

    html = f"""<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>{w.city} Weather Today</title>
  </head>
  <body style="margin:0; background:#0b0f14; color:#e8eef9; font-family: -apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;">
    <div style="max-width:720px; margin:0 auto; padding:24px 16px 40px;">
      <div style="background:rgba(17,24,36,0.92); border:1px solid rgba(255,255,255,0.10); border-radius:16px; padding:18px 18px 8px;">
        <div style="display:flex; align-items:baseline; justify-content:space-between; gap:12px; flex-wrap:wrap;">
          <div>
            <div style="font-size:14px; color:#9fb0c7;">Daily weather</div>
            <div style="font-size:20px; font-weight:700; margin-top:2px;">{w.city} · {w.today.isoformat()}</div>
          </div>
          <div style="font-size:12px; color:#9fb0c7;">
            Timezone: <span style="font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace; color:#dbe7ff;">{w.tz}</span>
          </div>
        </div>

        <div style="margin-top:14px; padding:12px 12px; background:rgba(255,255,255,0.04); border:1px solid rgba(255,255,255,0.08); border-radius:12px;">
          <div style="font-size:13px; color:#9fb0c7;">Summary</div>
          <div style="font-size:18px; font-weight:700; margin-top:4px;">{summary}</div>
        </div>

        <table role="presentation" cellspacing="0" cellpadding="0" style="width:100%; margin-top:14px; border-collapse:collapse;">
          <tr>
            <td style="padding:10px 0;">
              <div style="display:grid; grid-template-columns: 1fr 1fr; gap:10px;">
                <div style="background:rgba(255,255,255,0.03); border:1px solid rgba(255,255,255,0.08); border-radius:12px; padding:12px;">
                  <div style="font-size:12px; color:#9fb0c7;">Temp min</div>
                  <div style="font-size:18px; font-weight:700; margin-top:4px;">{_fmt(w.t_min_c, suffix="°C")}</div>
                </div>
                <div style="background:rgba(255,255,255,0.03); border:1px solid rgba(255,255,255,0.08); border-radius:12px; padding:12px;">
                  <div style="font-size:12px; color:#9fb0c7;">Temp max</div>
                  <div style="font-size:18px; font-weight:700; margin-top:4px;">{_fmt(w.t_max_c, suffix="°C")}</div>
                </div>
                <div style="background:rgba(255,255,255,0.03); border:1px solid rgba(255,255,255,0.08); border-radius:12px; padding:12px;">
                  <div style="font-size:12px; color:#9fb0c7;">Precipitation</div>
                  <div style="font-size:18px; font-weight:700; margin-top:4px;">{_fmt(w.precip_mm, suffix=" mm")}</div>
                </div>
                <div style="background:rgba(255,255,255,0.03); border:1px solid rgba(255,255,255,0.08); border-radius:12px; padding:12px;">
                  <div style="font-size:12px; color:#9fb0c7;">Max wind</div>
                  <div style="font-size:18px; font-weight:700; margin-top:4px;">{_fmt(w.wind_max_kph, suffix=" km/h")}</div>
                </div>
              </div>
            </td>
          </tr>
        </table>

        <div style="margin-top:14px; padding-top:10px; border-top:1px solid rgba(255,255,255,0.08); font-size:12px; color:#9fb0c7;">
          Fetched at: <span style="font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace; color:#dbe7ff;">{w.fetched_at.isoformat(timespec="seconds")}</span>
        </div>
      </div>

      <div style="margin-top:12px; font-size:12px; color:#9fb0c7;">
        Source: Open-Meteo (no API key).
      </div>
    </div>
  </body>
</html>
"""
    return subject, html


def send_gmail_html(*, gmail_user: str, app_password: str, to_email: str, subject: str, html_body: str) -> None:
    msg = EmailMessage()
    msg["From"] = gmail_user
    msg["To"] = to_email
    msg["Subject"] = subject
    msg.set_content("This email contains HTML. If you see this, your client doesn't support HTML.")
    msg.add_alternative(html_body, subtype="html")

    context = ssl.create_default_context()
    try:
        with smtplib.SMTP_SSL("smtp.gmail.com", 465, context=context) as s:
            s.login(gmail_user, app_password)
            s.send_message(msg)
    except smtplib.SMTPAuthenticationError as e:
        # e.smtp_code is int, e.smtp_error is bytes (usually). Never print password.
        code = getattr(e, "smtp_code", None)
        raw_err = getattr(e, "smtp_error", b"")
        if isinstance(raw_err, bytes):
            err_text = raw_err.decode("utf-8", "replace")
        else:
            err_text = str(raw_err)
        err_text = err_text.strip()
        raise SystemExit(
            "\n".join(
                [
                    "Gmail SMTP authentication failed.",
                    f"SMTP code: {code}",
                    f"Server message: {err_text or '(empty)'}",
                    "",
                    "Most common fixes:",
                    "- Make sure 2-Step Verification is enabled for this Google account.",
                    "- Use a Gmail App Password (16 characters), not your normal Gmail password.",
                    "- If you copied an App Password with spaces, try entering the 16 characters exactly.",
                    "- Check for Google security blocks / 'Verify it's you' prompts on web login.",
                    "",
                    "More info: https://support.google.com/mail/?p=BadCredentials",
                ]
            )
        ) from e
    except smtplib.SMTPException as e:
        raise SystemExit(f"Gmail SMTP error: {e.__class__.__name__}: {e}") from e


def main() -> None:
    gmail_user = _env("GMAIL_USER")
    app_password = _env("GMAIL_APP_PASSWORD")
    to_email = _env("TO_EMAIL", required=False) or gmail_user

    w = fetch_hcm_today_weather()
    subject, html_body = build_weather_email_html(w)
    send_gmail_html(
        gmail_user=gmail_user,
        app_password=app_password,
        to_email=to_email,
        subject=subject,
        html_body=html_body,
    )
    print(f"Sent to {to_email}: {subject}")


if __name__ == "__main__":
    main()

