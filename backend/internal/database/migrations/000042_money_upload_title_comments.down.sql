-- Rollback migration 000041: Remove editable title and comments from Money Creek uploads

DROP INDEX IF EXISTS woulder.idx_money_uploads_title_trgm;

ALTER TABLE woulder.money_uploads
  DROP COLUMN IF EXISTS comments,
  DROP COLUMN IF EXISTS title;
