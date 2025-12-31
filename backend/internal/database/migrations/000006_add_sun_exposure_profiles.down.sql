-- Rollback sun exposure profiles migration

-- Remove seepage risk flag from locations
ALTER TABLE woulder.locations DROP COLUMN IF EXISTS has_seepage_risk;

-- Drop location_sun_exposure table (cascade will remove indexes)
DROP TABLE IF EXISTS woulder.location_sun_exposure;
