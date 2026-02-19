-- Migration 000025 rollback: Remove Kaya performance optimization indexes

DROP INDEX IF EXISTS woulder.idx_kaya_ascents_slug_date;
DROP INDEX IF EXISTS woulder.idx_kaya_climbs_location_slug;
