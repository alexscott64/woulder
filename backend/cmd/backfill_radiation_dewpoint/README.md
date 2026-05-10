# backfill_radiation_dewpoint

A one-off CLI tool that backfills historical solar radiation and dewpoint
values into [`woulder.weather_data`](../../internal/database/migrations/000032_add_radiation_dewpoint.up.sql:1)
for rows that pre-date migration `000032_add_radiation_dewpoint`.

## Purpose

Migration 000032 added four new columns to `woulder.weather_data`:

- `shortwave_radiation` (W/m²)
- `direct_radiation` (W/m²)
- `diffuse_radiation` (W/m²)
- `dewpoint_f` (°F)

Because the migration uses `NOT NULL DEFAULT 0`, any rows that already existed
when the migration ran were given `0` for all four columns. This tool finds
those rows and fills them in by querying the free
[Open-Meteo Historical Archive API](https://open-meteo.com/en/docs/historical-weather-api).

## Prerequisites

- Migration `000032_add_radiation_dewpoint` must be applied.
- The same DB env vars as the rest of the backend must be set (typically via
  [`backend/.env`](../../.env.example:1)):
  - `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`
- Locations referenced by `weather_data.location_id` must have valid `latitude`
  and `longitude` in `woulder.locations`.
- Outbound HTTPS access to `https://archive-api.open-meteo.com`.

No API key is required — the archive endpoint is part of Open-Meteo's free
tier (≈10,000 calls/day, ≈600/min).

## Build & run

From the repository root:

```bash
cd backend && go build ./cmd/backfill_radiation_dewpoint/...
```

Dry-run (recommended first):

```bash
cd backend && go run ./cmd/backfill_radiation_dewpoint --start-date 2024-01-01 --dry-run
```

Real run for one location:

```bash
cd backend && go run ./cmd/backfill_radiation_dewpoint --location-id 12
```

Real run for everything:

```bash
cd backend && go run ./cmd/backfill_radiation_dewpoint
```

## Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `--batch-size` | int | `100` | Informational batch grouping size. The tool currently issues one API call per location (covering its full missing date range), since Open-Meteo accepts arbitrary date windows. |
| `--start-date` | string (`YYYY-MM-DD`) | oldest row in DB | Earliest `weather_data.timestamp` (UTC) to consider. |
| `--end-date` | string (`YYYY-MM-DD`) | today | Latest `weather_data.timestamp` (UTC) to consider. |
| `--location-id` | int | `0` (all) | Restrict to a single `location_id`. |
| `--dry-run` | bool | `false` | Log how many rows would be updated without writing. |
| `--rate-limit-ms` | int | `1100` | Sleep between Open-Meteo calls. Default keeps the tool safely under the free-tier 600-req/min limit. |

## Behavior

1. Loads lat/lon for every (or just one) location from `woulder.locations`.
2. Finds every distinct `(location_id, date)` pair in `woulder.weather_data`
   where **all four** target columns are `NULL` or `0`, optionally filtered by
   the requested date range and `location_id`.
3. For each location, computes the `(min_date, max_date)` envelope of its
   missing dates and issues a single call to:
   ```
   https://archive-api.open-meteo.com/v1/archive
       ?latitude={lat}&longitude={lon}
       &start_date={YYYY-MM-DD}&end_date={YYYY-MM-DD}
       &hourly=shortwave_radiation,direct_radiation,diffuse_radiation,dew_point_2m
       &timezone=UTC
   ```
4. Parses each hourly sample (UTC) and runs:
   ```sql
   UPDATE woulder.weather_data
   SET shortwave_radiation = $1,
       direct_radiation    = $2,
       diffuse_radiation   = $3,
       dewpoint_f          = $4
   WHERE location_id = $5 AND timestamp = $6
   ```
   inside one transaction per location.
5. Converts dewpoint from °C → °F. Radiation values are stored as-is (W/m²).
6. Sleeps `--rate-limit-ms` between calls.
7. Logs progress like:
   ```
   [3/47] location=12 (47.61234, -120.66789) dates=2024-01-01..2024-03-15 (74 distinct dates)
       ✓ Fetched 1776 hourly samples
       ✓ rows_updated=1776
   ```
8. Honors `SIGINT` / `SIGTERM`: finishes the in-flight location batch, then
   exits cleanly.

## Notes

- **Idempotent.** Re-running is safe. Rows that already have non-zero values
  are skipped because the `WHERE` clause in step 2 only matches rows with
  all-zero/`NULL` columns. (A row touched by a previous run won't be re-found.)
- **Over-fetch is harmless.** The archive call covers the full envelope of
  missing dates for a location, which may include dates that already have
  data; those rows simply won't appear in the missing-rows query and won't
  be UPDATEd.
- **Open-Meteo rate limits (free tier):** ~600 requests/minute,
  ~5,000 requests/hour, ~10,000 requests/day. With one call per location and
  the default rate limit, the tool can comfortably handle thousands of
  locations per day.
- **Archive endpoint vs. forecast endpoint.** The existing
  [`OpenMeteoClient.GetHistoricalWeather`](../../internal/weather/client/openmeteo.go:465)
  uses the *forecast* endpoint with `past_days`, which is capped at ~92 days.
  This tool deliberately calls the dedicated archive endpoint
  (`archive-api.open-meteo.com`) directly so it can reach data going back
  several years.
