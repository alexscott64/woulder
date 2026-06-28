DROP INDEX IF EXISTS woulder.idx_money_notes_deleted_at;

ALTER TABLE woulder.money_notes
  DROP COLUMN IF EXISTS deleted_by,
  DROP COLUMN IF EXISTS deleted_at;
