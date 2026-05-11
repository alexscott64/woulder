-- ════════════════════════════════════════════════════════════════════════════
-- Location Sun Exposure Seed (operator-editable TEMPLATE)
-- ════════════════════════════════════════════════════════════════════════════
--
-- ⚠️  IMPORTANT: This table does NOT have `aspect_degrees` or `face_dip_degrees`
--     columns (despite what some confidence.go reason strings claim — that is
--     a known misleading-message bug, follow-up to be filed separately).
--
--     The rock-temp calculator derives aspect/dip from the percentage
--     breakdown below, using these rules — see
--     backend/internal/weather/rock_temp/inputs.go::ResolveDominantFacet:
--
--       Aspect (60%-dominance + weighted fallback):
--         * If any of south/west/east/north_facing_percent > 60, that face
--           dominates:  S = 180°,  W = 270°,  E = 90°,  N = 0°.
--         * Otherwise the four percentages are vector-summed into a
--           weighted azimuth (WeightedAspect()), with a south-facing
--           fallback when the resultant is degenerate.
--
--       Dip (60%-dominance + weighted fallback):
--         * slab_percent     > 60  →  45°  (slab)
--         * overhang_percent > 60  → 110°  (overhang)
--         * else weighted blend (vertical = 100 - slab - overhang) using
--           WeightedDip() with weights 45 / 90 / 110.
--
--       Tree coverage:
--         * tree_coverage_percent → fraction in [0,1] via TreeFraction().
--
-- ─── Field guide ────────────────────────────────────────────────────────────
--   south/west/east/north_facing_percent  : 0–100, should sum to ~100
--   slab_percent                           : 0–100, low-angle (<70°)
--   overhang_percent                       : 0–100, past vertical (>90°)
--   tree_coverage_percent                  : 0–100, canopy shading the rock
--   description                            : free text
--
-- ─── Workflow ───────────────────────────────────────────────────────────────
--   1. cp seed_location_sun_exposure_TEMPLATE.sql seed_location_sun_exposure.sql
--      (the non-TEMPLATE filename is gitignored / treated as your working copy)
--   2. Edit seed_location_sun_exposure.sql, refining values from field
--      knowledge. Pre-populated values reflect the LIVE RDS state at audit
--      time; tweak any that look generic.
--   3. PGPASSWORD=… psql "host=… port=5432 user=woulder dbname=woulder \
--          sslmode=require" -f seed_location_sun_exposure.sql
--   4. Re-run audit_location_sun_exposure.sql to confirm no MISSING_ROW /
--      ZERO_FACETS / ZERO_DIP rows remain.
--
-- ⚠️  This is NOT a numbered migration. Do NOT rename it into a 000NNN_*.sql
--     slot — it is operator-driven seed data, run on demand.
--
-- ─── Audit snapshot (2026-05-11) ────────────────────────────────────────────
--   total_locations: 34   rows_present: 34   rows_missing: 0
--   rows_all_zero_facets: 0   rows_all_zero_dip: 0
--   All 34 active locations have non-zero facet/dip data. The blocks below
--   are pre-populated from the live values and may be tweaked freely.
-- ════════════════════════════════════════════════════════════════════════════

BEGIN;
SET search_path TO woulder, public;


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║  Northwest → Pacific Northwest                                           ║
-- ╚══════════════════════════════════════════════════════════════════════════╝

-- Bellingham  (Arkose / Sandstone — wet-sensitive, dense forest)
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 20.0, 30.0, 25.0, 25.0, 30.0, 30.0, 70.0,
       'Dense forest bouldering, mostly sandstone with heavy tree coverage'
FROM woulder.locations WHERE name = 'Bellingham'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;

-- Calendar Butte  (Graywacke / Argillite / Phyllite — slab-dominant)
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 45.0, 25.0, 20.0, 10.0, 55.0, 10.0, 40.0,
       'Excellent south/slab exposure with moderate tree coverage'
FROM woulder.locations WHERE name = 'Calendar Butte'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;

-- Gold Bar  (Granodiorite / Granite)
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 40.0, 25.0, 20.0, 15.0, 45.0, 15.0, 55.0,
       'Forest bouldering with good south exposure'
FROM woulder.locations WHERE name = 'Gold Bar'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;

-- Icicle Creek (Leavenworth)  (Granite / Schist)
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 45.0, 25.0, 20.0, 10.0, 50.0, 10.0, 35.0,
       'Excellent south exposure, moderate tree coverage, many slabs'
FROM woulder.locations WHERE name = 'Icicle Creek (Leavenworth)'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;

-- Index  (Granite — tall walls, mixed aspects)
-- TODO (quick-win): seed value is generic; user has firsthand knowledge here.
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 30.0, 30.0, 20.0, 20.0, 35.0, 25.0, 50.0,
       'Wall climbing with good west/south exposure, moderate tree coverage'
FROM woulder.locations WHERE name = 'Index'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;

-- Skykomish - Money Creek  (mixed metavolcanics + granitics)
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 35.0, 25.0, 25.0, 15.0, 40.0, 20.0, 60.0,
       'Forest bouldering with mixed aspects and moderate tree shade'
FROM woulder.locations WHERE name = 'Skykomish - Money Creek'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;

-- Skykomish - Paradise  (Granodiorite / Granite — high elevation, snowmelt seepage)
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 35.0, 25.0, 25.0, 15.0, 45.0, 15.0, 65.0,
       'High elevation forest bouldering, good slab percentage but heavy tree shade'
FROM woulder.locations WHERE name = 'Skykomish - Paradise'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;

-- Squamish  (Granite — tall walls + boulders, well-known crag)
-- TODO (quick-win): pre-populated value is generic; tweak for The Boulders / The Apron.
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 40.0, 30.0, 20.0, 10.0, 40.0, 20.0, 45.0,
       'Mixed wall and boulder climbing with good sun exposure'
FROM woulder.locations WHERE name = 'Squamish'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;

-- Treasury  (Granite — high alpine, snowmelt seepage)
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 30.0, 30.0, 25.0, 15.0, 40.0, 20.0, 60.0,
       'High alpine bouldering with moderate tree coverage and snowmelt seepage risk'
FROM woulder.locations WHERE name = 'Treasury'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║  Southwest → Nevada (Red Rock — Aztec sandstone, wet-sensitive)          ║
-- ╚══════════════════════════════════════════════════════════════════════════╝

-- Ash Spring Boulders  (Calico Basin)
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 40.0, 30.0, 20.0, 10.0, 30.0, 30.0, 6.0,
       'Calico Basin near spring, good sun exposure, some vegetation'
FROM woulder.locations WHERE name = 'Ash Spring Boulders'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;

-- Black Velvet Canyon Boulders
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 35.0, 30.0, 20.0, 15.0, 25.0, 35.0, 3.0,
       'Desert canyon sandstone, excellent sun exposure, very sparse vegetation'
FROM woulder.locations WHERE name = 'Black Velvet Canyon Boulders'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;

-- First Creek Canyon Boulders
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 40.0, 30.0, 20.0, 10.0, 32.0, 28.0, 6.0,
       'Desert canyon sandstone, excellent sun exposure, sparse vegetation'
FROM woulder.locations WHERE name = 'First Creek Canyon Boulders'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;

-- First Pullout Boulders
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 45.0, 30.0, 15.0, 10.0, 35.0, 25.0, 2.0,
       'Roadside bouldering, excellent sun exposure, virtually no shade'
FROM woulder.locations WHERE name = 'First Pullout Boulders'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;

-- Gateway Canyon  (Calico Basin)
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 38.0, 32.0, 20.0, 10.0, 30.0, 30.0, 5.0,
       'Calico Basin gateway area, excellent sun exposure, sparse trees'
FROM woulder.locations WHERE name = 'Gateway Canyon'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;

-- Ice Box Canyon Boulders
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 35.0, 30.0, 22.0, 13.0, 28.0, 32.0, 10.0,
       'Shaded canyon with seasonal water, more vegetation than most Red Rock areas'
FROM woulder.locations WHERE name = 'Ice Box Canyon Boulders'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;

-- Juniper Canyon Boulders
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 38.0, 30.0, 20.0, 12.0, 30.0, 30.0, 8.0,
       'Desert canyon with juniper, good sun exposure'
FROM woulder.locations WHERE name = 'Juniper Canyon Boulders'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;

-- Kraft Boulders  (Calico Basin — most popular RR bouldering)
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 42.0, 30.0, 18.0, 10.0, 30.0, 30.0, 4.0,
       'Calico Basin sandstone, excellent sun exposure, minimal shade'
FROM woulder.locations WHERE name = 'Kraft Boulders'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;

-- Little Springs Canyon Boulders  (Calico Basin)
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 38.0, 30.0, 20.0, 12.0, 28.0, 32.0, 7.0,
       'Calico Basin canyon with seasonal spring, some vegetation'
FROM woulder.locations WHERE name = 'Little Springs Canyon Boulders'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;

-- Mustang Canyon Boulders
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 38.0, 30.0, 20.0, 12.0, 30.0, 30.0, 5.0,
       'Desert canyon sandstone, excellent sun exposure, minimal trees'
FROM woulder.locations WHERE name = 'Mustang Canyon Boulders'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;

-- Oak Creek Canyon Boulders
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 40.0, 30.0, 20.0, 10.0, 30.0, 30.0, 5.0,
       'Desert canyon sandstone, excellent sun exposure, minimal trees'
FROM woulder.locations WHERE name = 'Oak Creek Canyon Boulders'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;

-- Pine Creek Canyon Boulders
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 38.0, 32.0, 20.0, 10.0, 30.0, 30.0, 7.0,
       'Desert canyon with some vegetation, good sun exposure'
FROM woulder.locations WHERE name = 'Pine Creek Canyon Boulders'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;

-- Red Spring Boulders  (Calico Basin)
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 40.0, 30.0, 20.0, 10.0, 28.0, 32.0, 6.0,
       'Calico Basin near spring, good sun exposure, some vegetation'
FROM woulder.locations WHERE name = 'Red Spring Boulders'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;

-- Sandstone Quarry Boulders
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 42.0, 30.0, 18.0, 10.0, 33.0, 27.0, 4.0,
       'Quarry area sandstone, excellent sun exposure, minimal vegetation'
FROM woulder.locations WHERE name = 'Sandstone Quarry Boulders'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;

-- Second Pullout Boulders
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 45.0, 30.0, 15.0, 10.0, 35.0, 25.0, 2.0,
       'Roadside bouldering, excellent sun exposure, virtually no shade'
FROM woulder.locations WHERE name = 'Second Pullout Boulders'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;

-- Southern Outcrops Boulders
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 40.0, 32.0, 18.0, 10.0, 32.0, 28.0, 3.0,
       'Desert outcrops, excellent sun exposure, very sparse vegetation'
FROM woulder.locations WHERE name = 'Southern Outcrops Boulders'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;

-- White Rock Spring Boulders
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 40.0, 30.0, 20.0, 10.0, 32.0, 28.0, 6.0,
       'Desert canyon sandstone, excellent sun exposure, sparse trees'
FROM woulder.locations WHERE name = 'White Rock Spring Boulders'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;

-- Willow Spring Boulders
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 38.0, 32.0, 20.0, 10.0, 28.0, 32.0, 8.0,
       'Desert canyon with some pinyon pine, good sun exposure'
FROM woulder.locations WHERE name = 'Willow Spring Boulders'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;

-- Windy Canyon Boulders
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 40.0, 35.0, 15.0, 10.0, 30.0, 30.0, 5.0,
       'Windy desert canyon, excellent sun exposure, minimal shade'
FROM woulder.locations WHERE name = 'Windy Canyon Boulders'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;


-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║  Southwest → Southern California                                         ║
-- ╚══════════════════════════════════════════════════════════════════════════╝

-- Black Mountain  (Tonalite — high alpine)
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 50.0, 25.0, 15.0, 10.0, 50.0, 10.0, 20.0,
       'High alpine bouldering, excellent south/slab exposure, minimal trees'
FROM woulder.locations WHERE name = 'Black Mountain'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;

-- Buttermilks  (Granodiorite — Bishop, world-famous)
-- TODO (quick-win): pre-populated value is generic; user knows this crag personally.
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 45.0, 30.0, 15.0, 10.0, 45.0, 15.0, 10.0,
       'High desert bouldering, excellent sun exposure, sparse tree coverage'
FROM woulder.locations WHERE name = 'Buttermilks'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;

-- Happy / Sad Boulders  (Rhyolite — Volcanic Tablelands)
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 40.0, 30.0, 20.0, 10.0, 40.0, 20.0, 15.0,
       'High desert bouldering, good sun exposure, minimal tree coverage'
FROM woulder.locations WHERE name = 'Happy / Sad Boulders'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;

-- Joshua Tree  (Granodiorite — desert classic)
-- TODO (quick-win): pre-populated value is generic; user knows this crag personally.
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 40.0, 30.0, 20.0, 10.0, 35.0, 25.0, 5.0,
       'Desert granite bouldering, excellent sun exposure, minimal shade'
FROM woulder.locations WHERE name = 'Joshua Tree'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;

-- Tramway  (Tonalite — high alpine, San Jacinto)
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 45.0, 30.0, 15.0, 10.0, 50.0, 10.0, 25.0,
       'High alpine bouldering, excellent south exposure, sparse trees'
FROM woulder.locations WHERE name = 'Tramway'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;

-- Yosemite  (Granodiorite — Camp 4 boulders + walls)
-- TODO (quick-win): pre-populated value is generic; user knows this crag personally.
INSERT INTO woulder.location_sun_exposure
  (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
   north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 35.0, 30.0, 20.0, 15.0, 40.0, 20.0, 30.0,
       'Mixed alpine climbing, moderate tree coverage at lower elevations'
FROM woulder.locations WHERE name = 'Yosemite'
ON CONFLICT (location_id) DO UPDATE SET
  south_facing_percent  = EXCLUDED.south_facing_percent,
  west_facing_percent   = EXCLUDED.west_facing_percent,
  east_facing_percent   = EXCLUDED.east_facing_percent,
  north_facing_percent  = EXCLUDED.north_facing_percent,
  slab_percent          = EXCLUDED.slab_percent,
  overhang_percent      = EXCLUDED.overhang_percent,
  tree_coverage_percent = EXCLUDED.tree_coverage_percent,
  description           = EXCLUDED.description,
  updated_at            = CURRENT_TIMESTAMP;


-- ─── Reference: TODO block for any future MISSING_ROW location ──────────────
-- (Copy + uncomment + edit when adding a new location that the audit reports
--  as MISSING_ROW / ZERO_FACETS / ZERO_DIP.)
--
-- INSERT INTO woulder.location_sun_exposure
--   (location_id, south_facing_percent, west_facing_percent, east_facing_percent,
--    north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
-- SELECT id, NULL, NULL, NULL, NULL, NULL, NULL, NULL,
--        'TODO: fill in based on field knowledge'
-- FROM woulder.locations WHERE name = '<New Location Name>'
-- ON CONFLICT (location_id) DO UPDATE SET
--   south_facing_percent  = EXCLUDED.south_facing_percent,
--   west_facing_percent   = EXCLUDED.west_facing_percent,
--   east_facing_percent   = EXCLUDED.east_facing_percent,
--   north_facing_percent  = EXCLUDED.north_facing_percent,
--   slab_percent          = EXCLUDED.slab_percent,
--   overhang_percent      = EXCLUDED.overhang_percent,
--   tree_coverage_percent = EXCLUDED.tree_coverage_percent,
--   description           = EXCLUDED.description,
--   updated_at            = CURRENT_TIMESTAMP;

COMMIT;
