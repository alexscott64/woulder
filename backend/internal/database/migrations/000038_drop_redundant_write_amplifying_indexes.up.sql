-- Drop redundant write-amplifying indexes identified during PostgreSQL diagnostics.
--
-- These indexes duplicate unique constraints/indexes that already support the
-- same equality lookups and ON CONFLICT arbiters. Dropping them reduces WAL,
-- vacuum, checkpoint, and buffer churn during high-volume Kaya/weather syncs.
--
-- This migration uses CONCURRENTLY. The custom migration runner detects
-- CONCURRENTLY and runs the migration outside a transaction for `up`/`down`.
-- Do not apply this file via `migrate step`, because step migrations are still
-- transaction-wrapped in backend/cmd/migrate/main.go.

DROP INDEX CONCURRENTLY IF EXISTS woulder.idx_kaya_users_kaya_user_id;
DROP INDEX CONCURRENTLY IF EXISTS woulder.idx_kaya_climbs_slug;
DROP INDEX CONCURRENTLY IF EXISTS woulder.idx_kaya_ascents_kaya_ascent_id;

-- weather_data already has a UNIQUE(location_id, timestamp) index that supports
-- the upsert arbiter and equality/range access pattern. The extra non-unique
-- (location_id, timestamp DESC) index was amplifying writes on the hottest table.
DROP INDEX CONCURRENTLY IF EXISTS woulder.idx_weather_data_location_timestamp;
