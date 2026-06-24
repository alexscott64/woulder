-- Migration 000039: Extend Money Creek toolkit into crag hierarchy model

ALTER TABLE woulder.money_features
  ADD COLUMN IF NOT EXISTS parent_feature_id UUID REFERENCES woulder.money_features(id) ON DELETE CASCADE,
  ADD COLUMN IF NOT EXISTS sort_order INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS external_ref TEXT,
  ADD COLUMN IF NOT EXISTS import_source TEXT;

ALTER TABLE woulder.money_notes
  ADD COLUMN IF NOT EXISTS target_type TEXT NOT NULL DEFAULT 'feature',
  ADD COLUMN IF NOT EXISTS target_ref TEXT,
  ADD COLUMN IF NOT EXISTS tags TEXT[] NOT NULL DEFAULT '{}',
  ADD COLUMN IF NOT EXISTS blocks JSONB NOT NULL DEFAULT '[]',
  ADD COLUMN IF NOT EXISTS external_ref TEXT,
  ADD COLUMN IF NOT EXISTS import_source TEXT;

ALTER TABLE woulder.money_uploads
  ADD COLUMN IF NOT EXISTS block_kind TEXT NOT NULL DEFAULT 'photo',
  ADD COLUMN IF NOT EXISTS metadata JSONB NOT NULL DEFAULT '{}';

ALTER TABLE woulder.money_features DROP CONSTRAINT IF EXISTS money_features_feature_type_check;
ALTER TABLE woulder.money_features ADD CONSTRAINT money_features_feature_type_check
  CHECK (feature_type IN ('area','boulder','problem','trail','topo','poi','drawing'));

ALTER TABLE woulder.money_features DROP CONSTRAINT IF EXISTS money_features_status_check;
ALTER TABLE woulder.money_features ADD CONSTRAINT money_features_status_check
  CHECK (status IN ('draft','active','archived','scouted','needs-work','cleaning','established','project','sent'));

ALTER TABLE woulder.money_notes DROP CONSTRAINT IF EXISTS money_notes_target_type_check;
ALTER TABLE woulder.money_notes ADD CONSTRAINT money_notes_target_type_check
  CHECK (target_type IN ('project','feature','area','boulder','trail','point','none'));

ALTER TABLE woulder.money_uploads DROP CONSTRAINT IF EXISTS money_uploads_block_kind_check;
ALTER TABLE woulder.money_uploads ADD CONSTRAINT money_uploads_block_kind_check
  CHECK (block_kind IN ('photo','sketch','file','topo'));

CREATE UNIQUE INDEX IF NOT EXISTS idx_money_features_project_external_ref
  ON woulder.money_features(project_id, external_ref)
  WHERE external_ref IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_money_features_parent_sort
  ON woulder.money_features(project_id, parent_feature_id, sort_order, title);
CREATE INDEX IF NOT EXISTS idx_money_features_import_source
  ON woulder.money_features(project_id, import_source);
CREATE INDEX IF NOT EXISTS idx_money_notes_project_target_created
  ON woulder.money_notes(project_id, target_type, target_ref, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_money_notes_tags_gin
  ON woulder.money_notes USING GIN(tags);
CREATE INDEX IF NOT EXISTS idx_money_notes_blocks_gin
  ON woulder.money_notes USING GIN(blocks);
CREATE UNIQUE INDEX IF NOT EXISTS idx_money_notes_project_external_ref
  ON woulder.money_notes(project_id, external_ref)
  WHERE external_ref IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_money_uploads_project_kind_created
  ON woulder.money_uploads(project_id, block_kind, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_money_uploads_metadata_gin
  ON woulder.money_uploads USING GIN(metadata);
