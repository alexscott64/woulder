-- Rollback migration 000040: Remove private asset storage metadata for Money Creek uploads

DROP INDEX IF EXISTS woulder.idx_money_uploads_storage_backend_bucket;
DROP INDEX IF EXISTS woulder.idx_money_uploads_deleted_at;
DROP INDEX IF EXISTS woulder.idx_money_uploads_project_sync_updated;

ALTER TABLE woulder.money_uploads DROP CONSTRAINT IF EXISTS money_uploads_sync_status_check;
ALTER TABLE woulder.money_uploads DROP CONSTRAINT IF EXISTS money_uploads_visibility_check;
ALTER TABLE woulder.money_uploads DROP CONSTRAINT IF EXISTS money_uploads_storage_backend_check;
ALTER TABLE woulder.money_uploads DROP CONSTRAINT IF EXISTS money_uploads_asset_kind_check;

ALTER TABLE woulder.money_uploads
  DROP COLUMN IF EXISTS updated_at,
  DROP COLUMN IF EXISTS physically_deleted_at,
  DROP COLUMN IF EXISTS delete_requested_at,
  DROP COLUMN IF EXISTS deleted_by,
  DROP COLUMN IF EXISTS deleted_at,
  DROP COLUMN IF EXISTS sync_status,
  DROP COLUMN IF EXISTS visibility,
  DROP COLUMN IF EXISTS storage_version_id,
  DROP COLUMN IF EXISTS storage_etag,
  DROP COLUMN IF EXISTS storage_region,
  DROP COLUMN IF EXISTS storage_bucket,
  DROP COLUMN IF EXISTS storage_backend,
  DROP COLUMN IF EXISTS asset_kind;
