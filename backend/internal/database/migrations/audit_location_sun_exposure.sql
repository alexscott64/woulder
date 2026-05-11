-- Audit: which locations are missing or under-populated for sun exposure?
--
-- Read-only. Safe to run any time.
--
-- Run via psql:
--   psql "$DATABASE_URL" -f audit_location_sun_exposure.sql
-- Or with explicit connection params:
--   psql "host=$DB_HOST port=$DB_PORT user=$DB_USER dbname=$DB_NAME sslmode=$DB_SSLMODE" \
--        -f backend/internal/database/migrations/audit_location_sun_exposure.sql
--
-- Schema reference (000006_add_sun_exposure_profiles.up.sql):
--   woulder.location_sun_exposure has percentage columns only; there is NO
--   aspect_degrees and NO face_dip_degrees column. The rock-temp calculator
--   derives those values from the percentages via inputs.go::ResolveDominantFacet.
--
-- Notes about the locations table:
--   * No `state` column — geography is via area_id → areas.region.
--   * No `is_active` on locations — only on areas.

SET search_path TO woulder, public;

-- ─── Summary counts (active areas only) ──────────────────────────────
SELECT
  COUNT(*)                                                   AS total_locations,
  COUNT(lse.location_id)                                     AS rows_present,
  COUNT(*) FILTER (WHERE lse.location_id IS NULL)            AS rows_missing,
  COUNT(*) FILTER (
    WHERE lse.location_id IS NOT NULL
      AND COALESCE(lse.south_facing_percent,0)
        + COALESCE(lse.west_facing_percent,0)
        + COALESCE(lse.east_facing_percent,0)
        + COALESCE(lse.north_facing_percent,0) = 0
  )                                                          AS rows_all_zero_facets,
  COUNT(*) FILTER (
    WHERE lse.location_id IS NOT NULL
      AND COALESCE(lse.slab_percent,0) + COALESCE(lse.overhang_percent,0) = 0
  )                                                          AS rows_all_zero_dip
FROM woulder.locations l
LEFT JOIN woulder.location_sun_exposure lse ON lse.location_id = l.id
JOIN woulder.areas a ON a.id = l.area_id
WHERE a.is_active = true;

-- ─── Per-location detail (active areas only) ─────────────────────────
SELECT
  l.id,
  a.region,
  a.name                                          AS area_name,
  l.name                                          AS location_name,
  l.latitude,
  l.longitude,
  l.elevation_ft,
  lse.south_facing_percent  AS s_pct,
  lse.west_facing_percent   AS w_pct,
  lse.east_facing_percent   AS e_pct,
  lse.north_facing_percent  AS n_pct,
  lse.slab_percent          AS slab_pct,
  lse.overhang_percent      AS over_pct,
  lse.tree_coverage_percent AS tree_pct,
  CASE WHEN lse.location_id IS NULL THEN 'MISSING_ROW'
       WHEN COALESCE(lse.south_facing_percent,0)
          + COALESCE(lse.west_facing_percent,0)
          + COALESCE(lse.east_facing_percent,0)
          + COALESCE(lse.north_facing_percent,0) = 0 THEN 'ZERO_FACETS'
       ELSE 'ok' END                              AS aspect_status,
  CASE WHEN lse.location_id IS NULL THEN 'MISSING_ROW'
       WHEN COALESCE(lse.slab_percent,0) + COALESCE(lse.overhang_percent,0) = 0 THEN 'ZERO_DIP'
       ELSE 'ok' END                              AS dip_status,
  ARRAY(
    SELECT rt.name
    FROM woulder.location_rock_types lrt
    JOIN woulder.rock_types rt ON rt.id = lrt.rock_type_id
    WHERE lrt.location_id = l.id
    ORDER BY lrt.is_primary DESC, rt.name
  )                                               AS rock_types
FROM woulder.locations l
LEFT JOIN woulder.location_sun_exposure lse ON lse.location_id = l.id
JOIN woulder.areas a ON a.id = l.area_id
WHERE a.is_active = true
ORDER BY a.region, a.name, l.name;
