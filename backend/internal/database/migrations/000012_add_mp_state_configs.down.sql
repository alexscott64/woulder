-- Migration 000012: Rollback Mountain Project state configurations table

DROP TRIGGER IF EXISTS update_mp_state_configs_updated_at ON woulder.mp_state_configs;
DROP INDEX IF EXISTS woulder.idx_mp_state_configs_region;
DROP INDEX IF EXISTS woulder.idx_mp_state_configs_active;
DROP TABLE IF EXISTS woulder.mp_state_configs;
