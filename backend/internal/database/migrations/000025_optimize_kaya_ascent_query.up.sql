-- Migration 000025: Optimize Kaya ascent query performance
-- Adds composite index to improve performance of the GetAscentsWithDetailsForWoulderLocation query

-- The optimized query joins kaya_ascents -> kaya_climbs (on slug) -> filter by woulder_location_id
-- This composite index allows the database to efficiently:
-- 1. Filter climbs by woulder_location_id
-- 2. Look up ascents by kaya_climb_slug in the same index
-- Note: Order matters - woulder_location_id first because it's the WHERE clause filter
CREATE INDEX idx_kaya_climbs_location_slug ON woulder.kaya_climbs(woulder_location_id, slug);

-- Additional composite index for the kaya_ascents table to optimize the date-based ordering
-- after filtering by climb slug. This helps the query plan when joining and ordering.
-- Covers: WHERE c.woulder_location_id = X -> JOIN on slug -> ORDER BY date DESC
CREATE INDEX idx_kaya_ascents_slug_date ON woulder.kaya_ascents(kaya_climb_slug, date DESC);
