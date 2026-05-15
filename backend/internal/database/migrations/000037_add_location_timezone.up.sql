-- Migration 000037: Add per-location IANA timezone column
--
-- Adds a per-location IANA timezone string used by frontend display logic
-- (forecast "Now" detection, climbing-hour windows) and by backend
-- aggregation queries that currently hardcode 'America/Los_Angeles'.
--
-- Default = 'America/Los_Angeles' because all existing rows are PNW/Nevada,
-- which are in Pacific time. The cmd/backfill_location_timezone tool must
-- be run after this migration to correct any non-Pacific rows (e.g. Canada,
-- which should be 'America/Vancouver').

ALTER TABLE woulder.locations
    ADD COLUMN timezone TEXT NOT NULL DEFAULT 'America/Los_Angeles';

COMMENT ON COLUMN woulder.locations.timezone IS
    'IANA timezone name (e.g. America/Los_Angeles). Populated from (lat, lon) at insert time by the location service. Validated against time.LoadLocation in Go.';
