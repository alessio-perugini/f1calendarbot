# 🏎 F1CalendarBot

A simple telegram bot that will notify you 10 minutes before a new f1 sessions starts.

You can go on telegram and search for: [@f1CalendarAlert_bot](https://t.me/f1CalendarAlert_bot)

## ⌨️ Commands

`/subscribe` Use to receive notification

`/unsubscribe` Use to no longer receive notification

`/nextrace` Use to get the next GP info

## ⚙️ Configuration

| Variable | Required | Description |
|---|---|---|
| `F1CALENDAR__TELEGRAM_TOKEN` | yes | Telegram bot token |
| `F1CALENDAR__DATABASE_URL` | yes | Turso database name (`<name>.turso.io`) |
| `F1CALENDAR__TURSO_AUTH_TOKEN` | yes | Turso auth token |
| `F1CALENDAR__TIMEZONE` | no | Timezone for session times, default `Europe/Rome` |
| `F1CALENDAR__OTEL_ENDPOINT` | no | OTLP/HTTP endpoint; when empty logs go to stdout only |
| `F1CALENDAR__OTEL_ENDPOINT_AUTH` | no | Bearer token for the OTLP endpoint |
| `OTEL_SERVICE_NAME` | no | Service name reported to OTel |

Database migrations in `pkg/subscription/migrations` are applied automatically at startup.
