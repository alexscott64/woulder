# backfill_location_timezone

A one-off CLI tool that populates / corrects the per-location IANA timezone
column on [`woulder.locations`](../../internal/database/migrations/000037_add_location_timezone.up.sql:1).

## Purpose

Migration `000037_add_location_timezone` added a `timezone TEXT NOT NULL
DEFAULT 'America/Los_Angeles'` column. Every existing row therefore has the
Pacific default — which is correct for PNW and Nevada but **wrong** for any
non-Pacific crag (notably Squamish/BC, where the right value is
`America/Vancouver`).

This tool re-derives each row's timezone from its `(latitude, longitude)`
using the offline `tzf` polygon dataset wrapped by
[`internal/geo.LookupTimezone`](../../internal/geo/timezone.go:1), and
`UPDATE`s rows whose derived value differs from what's currently stored.

## Prerequisites

- Migration `000037_add_location_timezone` must be applied.
- DB env vars (typically via [`backend/.env`](../../.env.example:1)):
  - `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`
- No network access required — the tzf dataset is embedded in the binary.

## Build & run

From the repository root:

```bash
cd backend && go build ./cmd/backfill_location_timezone/...
```

Dry-run (recommended first):

```bash
cd backend && go run ./cmd/backfill_location_timezone -dry-run
```

Apply for real:

```bash
cd backend && go run ./cmd/backfill_location_timezone
```

Re-derive a single location (e.g. after fixing its lat/lon):

```bash
cd backend && go run ./cmd/backfill_location_timezone -location-id 42
```

Force-overwrite even rows that have already been moved off the default
(use this if you know the previous backfill used a stale tzf dataset):

```bash
cd backend && go run ./cmd/backfill_location_timezone -force
```

## Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `-dry-run` | bool | `false` | Log planned UPDATEs without writing. |
| `-location-id` | int | `0` (all) | Restrict to a single `location_id`. |
| `-force` | bool | `false` | Re-derive every row regardless of its current timezone. Without `-force`, only rows whose current timezone equals the migration default `'America/Los_Angeles'` are eligible to be updated; this avoids stomping a hand-corrected value. |

## Behavior

1. `SELECT id, name, latitude, longitude, timezone FROM woulder.locations`
   (optionally filtered by `-location-id`). The query goes through bare
   `database/sql`, **not** through the locations repo, because the repo does
   not yet read the `timezone` column (that lands in §1c of the rollout).
2. For each row, calls `geo.LookupTimezone(lat, lon)` (offline tzf lookup,
   defensive `America/Los_Angeles` fallback if the lookup returns empty).
3. If the derived value equals the current value: skip (logged as
   `action=skip-equal`).
4. Otherwise:
   - If `-force` is set, **or** the current value is the migration default
     `'America/Los_Angeles'`: `UPDATE woulder.locations SET timezone = $1
     WHERE id = $2`.
   - Otherwise: skip (logged as `action=skip-not-default`).
5. Prints a per-row line and a final summary
   `(scanned=N, updated=N, skipped=N, dry-run=true|false)`.

Sample output:

```
=== Location Timezone Backfill Tool ===
DRY RUN MODE: no rows will be modified
Default mode: only updating rows whose current timezone = "America/Los_Angeles"

✓ Database connected
✓ Loaded 18 location(s)

  id=1  name="Icicle Creek"     (47.59000,-120.78000) current="America/Los_Angeles" derived="America/Los_Angeles" action=skip-equal
  id=2  name="Red Rocks"        (36.13000,-115.45000) current="America/Los_Angeles" derived="America/Los_Angeles" action=skip-equal
  id=12 name="Squamish"         (49.70160,-123.15580) current="America/Los_Angeles" -> derived="America/Vancouver" action=would-update
  ...

=== Backfill Complete ===
Summary: scanned=18, updated=1, skipped=17, dry-run=true
Elapsed: 42ms
```

## Verification SQL

After running (without `-dry-run`):

```sql
-- Distribution of timezones across all locations.
SELECT timezone, COUNT(*) FROM woulder.locations GROUP BY timezone ORDER BY 2 DESC;

-- Spot-check a known non-Pacific row, e.g. Squamish.
SELECT id, name, latitude, longitude, timezone
FROM woulder.locations
WHERE name ILIKE '%squamish%';

-- Pre-flight: any locations with bogus coordinates that would have produced
-- a fallback rather than a real lookup?
SELECT id, name, latitude, longitude, timezone
FROM woulder.locations
WHERE latitude IS NULL OR longitude IS NULL
   OR (latitude = 0 AND longitude = 0)
   OR latitude < -90 OR latitude > 90
   OR longitude < -180 OR longitude > 180;
```

## Notes

- **Idempotent.** Re-running with no source-data changes is a no-op
  (every row hits `skip-equal`).
- **Offline.** Unlike `backfill_radiation_dewpoint`, this tool makes no HTTP
  calls. The tzf polygon dataset is embedded via `tzf-rel-lite` (~3 MB).
- **Defensive fallback.** If `geo.LookupTimezone` cannot resolve a coordinate
  (e.g. a row with `(0, 0)` lat/lon), it returns `'America/Los_Angeles'` —
  which means the tool will treat such rows as `skip-equal` and leave them
  untouched. Surface bad coordinates with the pre-flight SQL above.
