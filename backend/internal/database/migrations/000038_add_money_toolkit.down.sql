-- Migration 000038 DOWN: Remove Money Creek toolkit and general app auth tables

DROP INDEX IF EXISTS woulder.idx_money_uploads_project_feature_created;
DROP INDEX IF EXISTS woulder.idx_money_notes_project_feature_created;
DROP INDEX IF EXISTS woulder.idx_money_features_properties_gin;
DROP INDEX IF EXISTS woulder.idx_money_features_geojson_gin;
DROP INDEX IF EXISTS woulder.idx_money_features_viewport;
DROP INDEX IF EXISTS woulder.idx_money_features_project_type_status_updated;
DROP INDEX IF EXISTS woulder.idx_auth_refresh_tokens_active;
DROP INDEX IF EXISTS woulder.idx_auth_refresh_tokens_user;

DROP TABLE IF EXISTS woulder.money_uploads;
DROP TABLE IF EXISTS woulder.money_notes;
DROP TABLE IF EXISTS woulder.money_features;
DROP TABLE IF EXISTS woulder.money_projects;
DROP TABLE IF EXISTS woulder.auth_refresh_tokens;
DROP TABLE IF EXISTS woulder.users;
