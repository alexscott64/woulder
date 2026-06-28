ALTER TABLE woulder.money_notes
  ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS deleted_by UUID REFERENCES woulder.users(id);

CREATE INDEX IF NOT EXISTS idx_money_notes_deleted_at
  ON woulder.money_notes(deleted_at)
  WHERE deleted_at IS NOT NULL;
