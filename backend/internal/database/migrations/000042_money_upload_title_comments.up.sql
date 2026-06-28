-- Migration 000041: Add editable title and comments to Money Creek uploads

ALTER TABLE woulder.money_uploads
  ADD COLUMN IF NOT EXISTS title TEXT,
  ADD COLUMN IF NOT EXISTS comments TEXT;

CREATE INDEX IF NOT EXISTS idx_money_uploads_title_trgm
  ON woulder.money_uploads USING GIN (title gin_trgm_ops)
  WHERE title IS NOT NULL;
