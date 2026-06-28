-- Migration 000040: Add private asset storage metadata for Money Creek uploads

ALTER TABLE woulder.money_uploads
  ADD COLUMN IF NOT EXISTS asset_kind TEXT NOT NULL DEFAULT 'original',
  ADD COLUMN IF NOT EXISTS storage_backend TEXT NOT NULL DEFAULT 'local',
  ADD COLUMN IF NOT EXISTS storage_bucket TEXT,
  ADD COLUMN IF NOT EXISTS storage_region TEXT,
  ADD COLUMN IF NOT EXISTS storage_etag TEXT,
  ADD COLUMN IF NOT EXISTS storage_version_id TEXT,
  ADD COLUMN IF NOT EXISTS visibility TEXT NOT NULL DEFAULT 'private',
  ADD COLUMN IF NOT EXISTS sync_status TEXT NOT NULL DEFAULT 'available',
  ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS deleted_by UUID REFERENCES woulder.users(id),
  ADD COLUMN IF NOT EXISTS delete_requested_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS physically_deleted_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

UPDATE woulder.money_uploads
SET updated_at = created_at
WHERE updated_at IS NULL;

ALTER TABLE woulder.money_uploads DROP CONSTRAINT IF EXISTS money_uploads_asset_kind_check;
ALTER TABLE woulder.money_uploads ADD CONSTRAINT money_uploads_asset_kind_check
  CHECK (asset_kind IN ('original','variant','map_pack','topo_pack','reference_pack'));

ALTER TABLE woulder.money_uploads DROP CONSTRAINT IF EXISTS money_uploads_storage_backend_check;
ALTER TABLE woulder.money_uploads ADD CONSTRAINT money_uploads_storage_backend_check
  CHECK (storage_backend IN ('local','r2'));

ALTER TABLE woulder.money_uploads DROP CONSTRAINT IF EXISTS money_uploads_visibility_check;
ALTER TABLE woulder.money_uploads ADD CONSTRAINT money_uploads_visibility_check
  CHECK (visibility IN ('private'));

ALTER TABLE woulder.money_uploads DROP CONSTRAINT IF EXISTS money_uploads_sync_status_check;
ALTER TABLE woulder.money_uploads ADD CONSTRAINT money_uploads_sync_status_check
  CHECK (sync_status IN ('available','pending_upload','deleted'));

CREATE INDEX IF NOT EXISTS idx_money_uploads_project_sync_updated
  ON woulder.money_uploads(project_id, sync_status, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_money_uploads_deleted_at
  ON woulder.money_uploads(deleted_at)
  WHERE deleted_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_money_uploads_storage_backend_bucket
  ON woulder.money_uploads(storage_backend, storage_bucket);
