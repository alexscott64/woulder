-- Migration 000037 (down): drop the per-location timezone column.

ALTER TABLE woulder.locations
    DROP COLUMN timezone;
