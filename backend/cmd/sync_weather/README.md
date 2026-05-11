# sync_weather

Standalone CLI for manually refreshing weather data in `woulder.weather_data`.

## Purpose

This tool is the recommended dev companion to the `WEATHER_OFFLINE_MODE` config
flag. When the API server runs with `WEATHER_OFFLINE_MODE=true`, it stops
calling Open-Meteo on every page refresh — which is great for staying under the
free-tier rate limit while iterating on the UI, but it also means weather data
in the DB grows stale. Run `sync_weather` whenever you want to bring the cached
data up to date.

It is also useful in production for one-off backfills of a specific location
without waiting for the next background refresh tick.

The command bypasses the offline flag entirely — it always calls Open-Meteo
directly via the same `weather/client.OpenMeteoClient` the server uses, then
persists rows via the standard `weather.PostgresRepository.Save` method (same
upsert logic as the server's hourly refresh job).

## Prerequisites

- Network access to `api.open-meteo.com`
- DB env vars set (same as the API server — see [`backend/.env.example`](../../.env.example:1)):
  - `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`
- The tool will auto-load `.env` from the current directory or the parent
  directory.

## Usage

```bash
# Sync every location in woulder.locations
cd backend && go run ./cmd/sync_weather --all

# Sync just one location by ID
cd backend && go run ./cmd/sync_weather --location-id 12

# Dry run — log what would be fetched without touching Open-Meteo or the DB
cd backend && go run ./cmd/sync_weather --all --dry-run

# Custom rate limit (default 1100ms ≈ Open-Meteo free tier)
cd backend && go run ./cmd/sync_weather --all --rate-limit-ms 2000
```

Pass exactly one of `--all` or `--location-id` — the tool refuses to run if
both or neither are supplied.

## Flags

| Flag                | Default | Description                                                |
| ------------------- | ------- | ---------------------------------------------------------- |
| `--all`             | false   | Sync every row in `woulder.locations`                      |
| `--location-id INT` | 0       | Sync a single location by ID                               |
| `--rate-limit-ms N` | 1100    | Sleep between Open-Meteo calls (free tier ≈ 600 req/min)   |
| `--dry-run`         | false   | Log planned work without making API calls or DB writes     |

## Behavior

For each target location the tool:

1. Looks up `latitude`, `longitude` from `woulder.locations`.
2. Calls `OpenMeteoClient.GetCurrentAndForecast(lat, lon)`.
3. Purges stale future forecasts via `DeleteFutureForLocation`.
4. Saves the current row + every hourly forecast row via the repository's
   upsert `Save`.
5. Logs progress in the form `[3/47] location=12 (Gold Bar) rows_saved=192`.
6. Sleeps `--rate-limit-ms` before the next location.

The tool honors `SIGINT` / `SIGTERM`: on receipt it finishes the in-flight
location, writes its rows, then exits cleanly.

## Exit codes

- `0` — all locations synced successfully (or dry-run completed)
- `1` — at least one location failed (a non-zero failure count is reported in
  the summary; check the logs for per-location errors)

## Recommended dev workflow

```bash
# .env
WEATHER_OFFLINE_MODE=true
```

Then in your dev terminal:

```bash
# One-time DB warm-up after pulling fresh code
cd backend && go run ./cmd/sync_weather --all

# Re-run whenever you want fresh data (e.g. before a UI demo)
cd backend && go run ./cmd/sync_weather --all
```

The API server (running with `air` or `go run ./cmd/server`) will serve all
weather requests from the cached rows without ever touching Open-Meteo.

In production, leave `WEATHER_OFFLINE_MODE` unset (or set to `false`) and let
the server's built-in hourly background refresh keep the cache warm. You can
still use `sync_weather --location-id N` for one-off prod backfills.
