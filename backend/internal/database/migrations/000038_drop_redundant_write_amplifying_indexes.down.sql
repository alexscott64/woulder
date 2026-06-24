-- Recreate indexes dropped by 000038_drop_redundant_write_amplifying_indexes.
--
-- This rollback uses CONCURRENTLY. The custom migration runner detects
-- CONCURRENTLY and runs the rollback outside a transaction for `down`.
-- Do not apply this file via `migrate step -1`, because step migrations are
-- transaction-wrapped in backend/cmd/migrate/main.go.

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_kaya_users_kaya_user_id
    ON woulder.kaya_users(kaya_user_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_kaya_climbs_slug
    ON woulder.kaya_climbs(slug);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_kaya_ascents_kaya_ascent_id
    ON woulder.kaya_ascents(kaya_ascent_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_weather_data_location_timestamp
    ON woulder.weather_data(location_id, timestamp DESC);
